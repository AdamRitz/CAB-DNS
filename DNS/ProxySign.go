package main

// 签名算法实现

import (
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"

	"crypto/ecdsa"

	"golang.org/x/crypto/sha3"
)

var curve = elliptic.P256()

//	P-256 曲线阶
//
// n = FFFFFFFF00000000FFFFFFFFFFFFFFFFBCE6FAADA7179E84F3B9CAC2FC632551
var p256Order, _ = new(big.Int).SetString("ffffffff00000000ffffffffffffffffbce6faada7179e84f3b9cac2fc632551", 16)

// bigIntTo32Bytes 将一个 big.Int 转为 32 字节的切片（不足前补零）
func bigIntTo32Bytes(x *big.Int) []byte {
	hexStr := fmt.Sprintf("%064x", x)
	b, _ := hex.DecodeString(hexStr)
	return b
}

// SchnorrSignature 保存签名中的 R（仅取 x 坐标）和 s
type SchnorrSignature struct {
	R *big.Int
	S *big.Int
}

// schnorrSign 对消息哈希进行 Schnorr 签名，privateKeyHex 为私钥 16 进制字符串
func schnorrSign(privateKeyHex string, msgHash []byte) (SchnorrSignature, error) {
	// 将私钥转换为大整数
	d, ok := new(big.Int).SetString(privateKeyHex, 16)
	if !ok {
		return SchnorrSignature{}, fmt.Errorf("invalid private key hex")
	}
	// 生成随机 nonce k
	kBytes := make([]byte, 32)
	_, err := rand.Read(kBytes)
	if err != nil {
		return SchnorrSignature{}, err
	}
	k := new(big.Int).SetBytes(kBytes)
	k.Mod(k, p256Order)

	// 计算 R = k*G
	Rx, _ := curve.ScalarBaseMult(k.Bytes())
	// 用 Rx（32 字节）计算哈希： e = keccak256(Rx || msgHash)
	RxBytes := bigIntTo32Bytes(Rx)
	data := append(RxBytes, msgHash...)
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(data)
	eBytes := hasher.Sum(nil)
	e := new(big.Int).SetBytes(eBytes)
	e.Mod(e, p256Order)

	// 计算 s = k + e*d (mod n)
	s := new(big.Int).Mul(e, d)
	s.Add(s, k)
	s.Mod(s, p256Order)

	return SchnorrSignature{R: Rx, S: s}, nil
}

// schnorrVerify 验证 Schnorr 签名，pubKey 以其 X、Y 坐标给出（使用 P-256）
func schnorrVerify(pubKeyX, pubKeyY *big.Int, msgHash []byte, sig SchnorrSignature) bool {
	// 计算 e = keccak256(bigIntTo32Bytes(sig.R) || msgHash)
	RxBytes := bigIntTo32Bytes(sig.R)
	data := append(RxBytes, msgHash...)
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(data)
	eBytes := hasher.Sum(nil)
	e := new(big.Int).SetBytes(eBytes)
	e.Mod(e, p256Order)

	// 计算 sG = s*G
	sGx, sGy := curve.ScalarBaseMult(sig.S.Bytes())
	// 计算 eP = e*(pubKey)
	ePx, ePy := curve.ScalarMult(pubKeyX, pubKeyY, e.Bytes())
	// 计算 R' = sG - eP，即 sG + (-eP)
	neg_ePy := new(big.Int).Neg(ePy)
	neg_ePy.Mod(neg_ePy, curve.Params().P)
	RprimeX, _ := curve.Add(sGx, sGy, ePx, neg_ePy)

	// 验证 R'.x 是否等于签名中的 R
	return RprimeX.Cmp(sig.R) == 0
}

// Delegate 根据原始私钥 sk 和中间信息 mw 计算代理私钥 proxysk 和中间点 K
func Delegate(sk *big.Int, mw string) (mwOut string, proxysk *big.Int, Kx, Ky *big.Int, err error) {
	// 生成随机 k
	kBytes := make([]byte, 32)
	_, err = rand.Read(kBytes)
	if err != nil {
		return "", nil, nil, nil, err
	}
	k := new(big.Int).SetBytes(kBytes)
	k.Mod(k, p256Order)

	// 计算 K = k*G
	Kx, Ky = curve.ScalarBaseMult(k.Bytes())

	// 对 K 做压缩编码：如果 Ky 偶数，前缀 0x02，否则 0x03；后接 X 坐标（32 字节）
	prefix := byte(0x02)
	if Ky.Bit(0) == 1 {
		prefix = 0x03
	}
	KxBytes := bigIntTo32Bytes(Kx)
	KCompressed := append([]byte{prefix}, KxBytes...)

	// 计算 e = sha256( mw || compressed(K) )
	hash := sha256.New()
	hash.Write([]byte(mw))
	hash.Write(KCompressed)
	eBytes := hash.Sum(nil)
	e := new(big.Int).SetBytes(eBytes)
	e.Mod(e, p256Order)

	// 计算 proxysk = sk*e + k (mod n)
	proxysk = new(big.Int).Mul(sk, e)
	proxysk.Add(proxysk, k)
	proxysk.Mod(proxysk, p256Order)
	return mw, proxysk, Kx, Ky, nil
}

// ProxySign 用代理私钥 proxysk 对消息 mp 进行签名（先对 mp 用 keccak256 哈希）
func ProxySign(mw string, proxysk *big.Int, Kx, Ky *big.Int, mp string) (sig SchnorrSignature, err error) {
	// 对 mp 计算 keccak256 哈希
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write([]byte(mp))
	mpHash := hasher.Sum(nil)
	// 将 proxysk 转成 16 进制字符串后参与签名
	proxyskHex := fmt.Sprintf("%x", proxysk)
	sig, err = schnorrSign(proxyskHex, mpHash)
	return sig, nil
}

//	func ProxySign(mw string, proxysk *big.Int, Kx, Ky *big.Int, mp string) (mpOut string, sig SchnorrSignature, KxOut, KyOut *big.Int, err error) {
//		// 对 mp 计算 keccak256 哈希
//		hasher := sha3.NewLegacyKeccak256()
//		hasher.Write([]byte(mp))
//		mpHash := hasher.Sum(nil)
//		// 将 proxysk 转成 16 进制字符串后参与签名
//		proxyskHex := fmt.Sprintf("%x", proxysk)
//		sig, err = schnorrSign(proxyskHex, mpHash)
//		if err != nil {
//			return "", SchnorrSignature{}, nil, nil, err
//		}
//		return mp, sig, Kx, Ky, nil
//	}
//
// ProxyVerify 验证代理签名，pk 为原始公钥（类型 *ecdsa.PublicKey，基于 P-256）
func ProxyVerify(mp string, sig SchnorrSignature, mw string, Kx, Ky *big.Int, pk *ecdsa.PublicKey) (bool, error) {
	// 重新计算 e = sha256( mw || compressed(K) )
	prefix := byte(0x02)
	if Ky.Bit(0) == 1 {
		prefix = 0x03
	}
	KxBytes := bigIntTo32Bytes(Kx)
	KCompressed := append([]byte{prefix}, KxBytes...)
	hash := sha256.New()
	hash.Write([]byte(mw))
	hash.Write(KCompressed)
	eBytes := hash.Sum(nil)
	e := new(big.Int).SetBytes(eBytes)
	e.Mod(e, p256Order)

	// 计算代理公钥 proxypk = P*e + K
	ePx, ePy := curve.ScalarMult(pk.X, pk.Y, e.Bytes())
	proxypkX, proxypkY := curve.Add(ePx, ePy, Kx, Ky)

	// 对 mp 计算 keccak256 哈希
	hasher2 := sha3.NewLegacyKeccak256()
	hasher2.Write([]byte(mp))
	mpHash := hasher2.Sum(nil)

	valid := schnorrVerify(proxypkX, proxypkY, mpHash, sig)
	return valid, nil
}
