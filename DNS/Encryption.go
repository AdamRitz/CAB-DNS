package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
)

// 加密算法实现
import (
	"crypto/rand"
	"crypto/sha256"
	"errors"
)

const privateKeyHex = "af44a005ad9e6d4c0873d609de9df16a9f8c8e490597087f4286b8103e0e7149"

// parsePrivateKey 解析 P-256 曲线下的私钥（16 进制字符串），生成 *ecdsa.PrivateKey
func parsePrivateKey() (*ecdsa.PrivateKey, error) {
	dBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return nil, err
	}
	d := new(big.Int).SetBytes(dBytes)
	curve := elliptic.P256()
	X, Y := curve.ScalarBaseMult(dBytes)
	pub := ecdsa.PublicKey{Curve: curve, X: X, Y: Y}
	priv := &ecdsa.PrivateKey{PublicKey: pub, D: d}
	return priv, nil
}

// compressPublicKey 将 *ecdsa.PublicKey 压缩为 33 字节（前缀 0x02 或 0x03，加上 32 字节 X 坐标）
func compressPublicKey(pub *ecdsa.PublicKey) []byte {
	prefix := byte(0x02)
	if pub.Y.Bit(0) == 1 {
		prefix = 0x03
	}
	xBytes := pub.X.Bytes()
	padded := make([]byte, 32)
	copy(padded[32-len(xBytes):], xBytes)
	return append([]byte{prefix}, padded...)
}

// ECCEncrypt 使用 ECIES 加密（基于 P-256 和 AES-GCM）
// 使用常量私钥推导出接收者公钥，并生成临时密钥对进行 ECDH 得到共享密钥。
func ECCEncrypt(msg string) (string, error) {
	priv, err := parsePrivateKey()
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %v", err)
	}
	recipientPub := &priv.PublicKey
	curve := elliptic.P256()

	// 生成临时密钥对
	ephemeralPriv, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return "", fmt.Errorf("failed to generate ephemeral key: %v", err)
	}
	ephemeralPub := &ephemeralPriv.PublicKey

	// 计算共享密钥：临时私钥与接收者公钥做 ECDH
	sx, _ := curve.ScalarMult(recipientPub.X, recipientPub.Y, ephemeralPriv.D.Bytes())
	aesKey := sha256.Sum256(sx.Bytes())

	// 使用 AES-GCM 加密
	block, err := aes.NewCipher(aesKey[:])
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %v", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %v", err)
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := aead.Seal(nil, nonce, []byte(msg), nil)

	// 最终输出：临时公钥（压缩，33 字节） || nonce || 密文，然后转为 16 进制字符串
	ephemeralCompressed := compressPublicKey(ephemeralPub)
	final := append(ephemeralCompressed, nonce...)
	final = append(final, ciphertext...)
	return hex.EncodeToString(final), nil
}

// ECCDecrypt 对上面 ECCEncrypt 输出的密文进行解密
func ECCDecrypt(cipherHex string) (string, error) {
	data, err := hex.DecodeString(cipherHex)
	if err != nil {
		return "", err
	}
	curve := elliptic.P256()
	if len(data) < 33 {
		return "", errors.New("ciphertext too short")
	}
	// 提取临时公钥（压缩格式 33 字节）
	ephemeralCompressed := data[:33]
	ephemeralPub, err := decompressPublicKey(ephemeralCompressed, curve)
	if err != nil {
		return "", fmt.Errorf("failed to decompress ephemeral public key: %v", err)
	}
	priv, err := parsePrivateKey()
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %v", err)
	}
	// 计算共享密钥：接收者私钥与临时公钥做 ECDH
	sx, _ := curve.ScalarMult(ephemeralPub.X, ephemeralPub.Y, priv.D.Bytes())
	aesKey := sha256.Sum256(sx.Bytes())

	block, err := aes.NewCipher(aesKey[:])
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %v", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %v", err)
	}
	nonceSize := aead.NonceSize()
	if len(data) < 33+nonceSize {
		return "", errors.New("ciphertext too short for nonce")
	}
	nonce := data[33 : 33+nonceSize]
	ciphertext := data[33+nonceSize:]
	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decryption failed: %v", err)
	}
	return string(plaintext), nil
}

// decompressPublicKey 将压缩格式的公钥还原为 *ecdsa.PublicKey（适用于 P-256）
// 压缩格式为：1 字节前缀（0x02 或 0x03）+ 32 字节 X 坐标，需根据前缀恢复 Y 坐标。
func decompressPublicKey(compressed []byte, curve elliptic.Curve) (*ecdsa.PublicKey, error) {
	if len(compressed) != 33 {
		return nil, errors.New("invalid compressed public key length")
	}
	prefix := compressed[0]
	x := new(big.Int).SetBytes(compressed[1:])
	params := curve.Params()
	// P-256 的曲线方程为： y^2 = x^3 - 3x + b (mod p)
	x3 := new(big.Int).Exp(x, big.NewInt(3), params.P)
	threeX := new(big.Int).Mul(x, big.NewInt(3))
	threeX.Mod(threeX, params.P)
	y2 := new(big.Int).Sub(x3, threeX)
	y2.Add(y2, params.B)
	y2.Mod(y2, params.P)
	y := new(big.Int).ModSqrt(y2, params.P)
	if y == nil {
		return nil, errors.New("failed to compute square root")
	}
	// 根据前缀判断奇偶性：若前缀 0x02 表示 y 应为偶数，0x03 表示 y 应为奇数
	if (prefix == 0x02 && y.Bit(0) == 1) || (prefix == 0x03 && y.Bit(0) == 0) {
		y.Sub(params.P, y)
		y.Mod(y, params.P)
	}
	return &ecdsa.PublicKey{Curve: curve, X: x, Y: y}, nil
}
