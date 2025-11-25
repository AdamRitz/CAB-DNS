import { PrivateKey, decrypt, encrypt } from "eciesjs";
import { secp256k1 } from "ethereum-cryptography/secp256k1";

const privateKey = 'af44a005ad9e6d4c0873d609de9df16a9f8c8e490597087f4286b8103e0e7149';
const publicKey=secp256k1.getPublicKey(BigInt('0x' + privateKey), true)
const sk=PrivateKey.fromHex(privateKey)

export function ECCEncrypt(msg){
    return encrypt(publicKey, msg).toString('hex')
}
export function ECCDecrpt(msg){
    msg=Buffer.from(msg,'hex')
    return decrypt(sk.secret,msg)
}
