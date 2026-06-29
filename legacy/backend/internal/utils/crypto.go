package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"

	"modelmarket/pkg/logger"
)

// Crypto 提供与 Fernet 等价的对称加密：AES-256-GCM
// 密钥来源：配置中的 base64 字符串，或 32 字节原始字符串；都不满足时用 SHA-256 派生。
type Crypto struct {
	gcm cipher.AEAD
}

// NewCrypto 用任意长度字符串构建 32 字节 key（SHA-256 派生）
func NewCrypto(secret string) (*Crypto, error) {
	if secret == "" {
		return nil, errors.New("encryption key is empty")
	}

	// 先尝试 base64 解码，失败再走 SHA-256 派生
	var key []byte
	if decoded, err := base64.StdEncoding.DecodeString(secret); err == nil && len(decoded) == 32 {
		key = decoded
		logger.L().Debug("Encryption: using base64-decoded 32-byte key")
	} else {
		h := sha256.Sum256([]byte(secret))
		key = h[:]
		logger.L().Debug("Encryption: derived 32-byte key via SHA-256")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &Crypto{gcm: gcm}, nil
}

// Encrypt 加密明文 -> base64(nonce|ciphertext)
func (c *Crypto) Encrypt(plain string) (string, error) {
	if plain == "" {
		return "", nil
	}
	nonce := make([]byte, c.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	cipherText := c.gcm.Seal(nonce, nonce, []byte(plain), nil)
	return base64.StdEncoding.EncodeToString(cipherText), nil
}

// Decrypt 反解 base64(nonce|ciphertext) -> 明文
func (c *Crypto) Decrypt(b64 string) (string, error) {
	if b64 == "" {
		return "", nil
	}
	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return "", err
	}
	if len(data) < c.gcm.NonceSize() {
		return "", errors.New("cipher too short")
	}
	nonce, ct := data[:c.gcm.NonceSize()], data[c.gcm.NonceSize():]
	plain, err := c.gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

// GenerateKey 生成一个全新的 32 字节 base64 密钥（首次启动可用）
func GenerateKey() string {
	b := make([]byte, 32)
	_, _ = io.ReadFull(rand.Reader, b)
	return base64.StdEncoding.EncodeToString(b)
}
