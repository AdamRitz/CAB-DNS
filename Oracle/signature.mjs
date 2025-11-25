import * as crypto from 'crypto'
import { ethers } from 'ethers';

// --- ELLIPTIC-CURVE IMPORTS ---
import { secp256k1 } from "ethereum-cryptography/secp256k1";
import { keccak256 } from "ethereum-cryptography/keccak";
import { randomBytes } from "crypto";
import { bytesToHex, hexToBytes, concatBytes } from "ethereum-cryptography/utils";
const p = 115792089237316195423570985008687907852837564279074904382605163141518161494337n;


// ============== 工具函数 ==============
function bigIntTo32Bytes(x) {
  const hex = x.toString(16).padStart(64, '0');
  return hexToBytes(hex);
}

function getPublicKeyUncompressed(skHex) {
  const skBytes = hexToBytes(skHex);
  return secp256k1.getPublicKey(skBytes, false);
}

// ============== Schnorr 签名与验证 ==============
function schnorrSign(privateKeyHex, msgHash) {
  const n = secp256k1.CURVE.n;
  const d = BigInt('0x' + privateKeyHex); // 私钥转成 BigInt

  // 1. 生成随机 nonce k
  const k = BigInt('0x' + bytesToHex(randomBytes(32))) % n;

  // 2. 计算 R = kG
  const R = secp256k1.ProjectivePoint.BASE.multiply(k);
  const Rx = R.toAffine().x; 

  // 3. 计算 e = H(Rx || msgHash)
  const RxBytes = bigIntTo32Bytes(Rx);
  const eBytes = keccak256(concatBytes(RxBytes, msgHash));
  const e = BigInt('0x' + bytesToHex(eBytes)) % n;

  // 4. 计算 s = k + e*d (mod n)
  const s = (k + e * d) % n;

  return { R: Rx, s };
}

function schnorrVerify(pubKey, msgHash, sig, isPoint) {
  let P;
  if (isPoint) {
    P = pubKey; // pubKey 已经是 ProjectivePoint
  } else {
    P = secp256k1.ProjectivePoint.fromHex(pubKey);
  }

  // 计算 e = H(Rx || msgHash)
  const RxBytes = bigIntTo32Bytes(sig.R);
  const eBytes = keccak256(concatBytes(RxBytes, msgHash));
  const e = BigInt('0x' + bytesToHex(eBytes)) % secp256k1.CURVE.n;

  // 验证: R' = sG - eP，检查 R'.x === R
  const sG = secp256k1.ProjectivePoint.BASE.multiply(sig.s);
  const eP = P.multiply(e);
  const Rprime = sG.subtract(eP).toAffine();

  return Rprime.x === sig.R;
}

// ============== 代理签名相关函数 ==============
export function Delegate(sk, mw) {
  const n = secp256k1.CURVE.n;
  // 1. 生成随机 k
  const k = BigInt('0x' + crypto.randomBytes(32).toString('hex')) % n; 
  // 2. 计算 K = kG
  const K = secp256k1.ProjectivePoint.BASE.multiply(k);

  // 3. 计算 e = sha256( mwBytes || KBytes )
  const mwBytes = new TextEncoder().encode(mw);
  const KBytes = K.toRawBytes(true); // 压缩格式，或者用 false 得到非压缩格式
  const eHex = crypto.createHash("sha256").update(Buffer.concat([mwBytes, KBytes])).digest("hex");
  const eBigInt = BigInt("0x" + eHex) % n;

  // 4. 计算 proxysk = sk * e + k (mod n)
  const proxysk = (sk * eBigInt + k) % n;

  return { mw, proxysk, K };
}

export function ProxySign(mw, proxysk, K, mp) {
  // 目前只是把 mpHash 用 proxysk 做 schnorrSign
  const mpHash = keccak256(new TextEncoder().encode(mp));
  const sig = schnorrSign(proxysk.toString(16), mpHash); // proxysk 转成 16 进制字符串

  return { mp, sig, K, mw };
}

function ProxyVerify(sigObj, mw, K, pk, mp) {
  const n = secp256k1.CURVE.n;
  // 原始公钥
  const P = secp256k1.ProjectivePoint.fromHex(pk);

  // 重新计算 e = sha256( mwBytes || KBytes )
  const mwBytes = new TextEncoder().encode(mw);
  const KBytes = K.toRawBytes(true);
  const eHex = crypto.createHash("sha256").update(Buffer.concat([mwBytes, KBytes])).digest("hex");
  const eBigInt = BigInt("0x" + eHex) % n;

  // proxy 公钥 = P*e + K
  const proxypk = P.multiply(eBigInt).add(K);

  // 再对 mp 做哈希
  const mpHash = keccak256(new TextEncoder().encode(mp));

  // 验证 schnorr
  return schnorrVerify(proxypk, mpHash, sigObj.sig, true);
}



