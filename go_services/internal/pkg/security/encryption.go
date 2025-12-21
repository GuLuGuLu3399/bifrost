package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"io"

	"github.com/gulugulu3399/bifrost/internal/pkg/xerr"
)

// EncryptionConfig 加密配置
type EncryptionConfig struct {
	Key       string `yaml:"key"`
	Algorithm string `yaml:"algorithm"`
}

// DefaultEncryptionConfig 返回默认加密配置
func DefaultEncryptionConfig() *EncryptionConfig {
	return &EncryptionConfig{
		Algorithm: "AES-256-GCM",
	}
}

// Encryptor 加密器接口
type Encryptor interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

// AESEncryptor AES 加密器
type AESEncryptor struct {
	key []byte
}

// NewAESEncryptor 创建 AES 加密器
func NewAESEncryptor(key string) (*AESEncryptor, error) {
	if key == "" {
		return nil, xerr.New(xerr.CodeBadRequest, "encryption key cannot be empty")
	}

	// 使用 SHA256 生成 32 字节的密钥
	keyHash := sha256.Sum256([]byte(key))

	return &AESEncryptor{
		key: keyHash[:],
	}, nil
}

// Encrypt 加密文本
func (e *AESEncryptor) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", xerr.New(xerr.CodeBadRequest, "plaintext cannot be empty")
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", xerr.Wrap(err, xerr.CodeInternal, "failed to create cipher block")
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", xerr.Wrap(err, xerr.CodeInternal, "failed to create GCM")
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", xerr.Wrap(err, xerr.CodeInternal, "failed to generate nonce")
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)

	// 使用 base64 编码
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 解密文本
func (e *AESEncryptor) Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", xerr.New(xerr.CodeBadRequest, "ciphertext cannot be empty")
	}

	// 解码 base64
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", xerr.Wrap(err, xerr.CodeBadRequest, "failed to decode base64 ciphertext")
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", xerr.Wrap(err, xerr.CodeInternal, "failed to create cipher block")
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", xerr.Wrap(err, xerr.CodeInternal, "failed to create GCM")
	}

	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return "", xerr.New(xerr.CodeBadRequest, "ciphertext too short")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", xerr.Wrap(err, xerr.CodeBadRequest, "failed to decrypt ciphertext")
	}

	return string(plaintext), nil
}

// Hash 使用 SHA256 哈希文本
func Hash(text string) string {
	if text == "" {
		return ""
	}

	hash := sha256.Sum256([]byte(text))
	return base64.StdEncoding.EncodeToString(hash[:])
}

// HashWithSalt 使用盐值哈希文本
func HashWithSalt(text, salt string) string {
	if text == "" {
		return ""
	}

	saltedText := text + salt
	hash := sha256.Sum256([]byte(saltedText))
	return base64.StdEncoding.EncodeToString(hash[:])
}

// GenerateRandomBytes 生成随机字节
func GenerateRandomBytes(length int) ([]byte, error) {
	if length <= 0 {
		return nil, xerr.New(xerr.CodeBadRequest, "length must be positive")
	}

	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, xerr.Wrap(err, xerr.CodeInternal, "failed to generate random bytes")
	}

	return bytes, nil
}

// GenerateRandomString 生成随机字符串
func GenerateRandomString(length int) (string, error) {
	if length <= 0 {
		return "", xerr.New(xerr.CodeBadRequest, "length must be positive")
	}

	bytes, err := GenerateRandomBytes(length)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

// XOR XOR 加密/解密
func XOR(text, key string) string {
	if text == "" || key == "" {
		return text
	}

	result := make([]byte, len(text))
	keyLen := len(key)

	for i, b := range []byte(text) {
		result[i] = b ^ []byte(key)[i%keyLen]
	}

	return string(result)
}

// Caesar 凯撒密码（仅用于演示，不安全）
func Caesar(text string, shift int) string {
	if text == "" {
		return text
	}

	result := make([]byte, len(text))

	for i, b := range []byte(text) {
		if b >= 'a' && b <= 'z' {
			result[i] = 'a' + (b-'a'+byte(shift))%26
		} else if b >= 'A' && b <= 'Z' {
			result[i] = 'A' + (b-'A'+byte(shift))%26
		} else {
			result[i] = b
		}
	}

	return string(result)
}

// ==========================================
// 密钥管理
// ==========================================

// KeyDerivation 密钥派生
func KeyDerivation(password, salt string, iterations, keyLen int) ([]byte, error) {
	if password == "" {
		return nil, xerr.New(xerr.CodeBadRequest, "password cannot be empty")
	}

	if salt == "" {
		return nil, xerr.New(xerr.CodeBadRequest, "salt cannot be empty")
	}

	if iterations <= 0 {
		return nil, xerr.New(xerr.CodeBadRequest, "iterations must be positive")
	}

	if keyLen <= 0 {
		return nil, xerr.New(xerr.CodeBadRequest, "key length must be positive")
	}

	// 简化的 PBKDF2 实现（实际项目中应使用 crypto/pbkdf2）
	hash := sha256.Sum256([]byte(password + salt))
	result := hash[:]

	for i := 1; i < iterations; i++ {
		hash = sha256.Sum256(result)
		result = hash[:]
	}

	if len(result) > keyLen {
		result = result[:keyLen]
	}

	return result, nil
}

// GenerateKeyPair 生成密钥对
func GenerateKeyPair() (string, string, error) {
	privateKey, err := GenerateRandomString(32)
	if err != nil {
		return "", "", err
	}

	publicKey, err := GenerateRandomString(32)
	if err != nil {
		return "", "", err
	}

	return privateKey, publicKey, nil
}

// ValidateKey 验证密钥强度
func ValidateKey(key string) error {
	if key == "" {
		return xerr.New(xerr.CodeBadRequest, "key cannot be empty")
	}

	if len(key) < 16 {
		return xerr.New(xerr.CodeValidation, "key must be at least 16 characters")
	}

	// 检查是否包含多种字符类型
	hasUpper := false
	hasLower := false
	hasDigit := false

	for _, c := range key {
		switch {
		case c >= 'A' && c <= 'Z':
			hasUpper = true
		case c >= 'a' && c <= 'z':
			hasLower = true
		case c >= '0' && c <= '9':
			hasDigit = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit {
		return xerr.New(xerr.CodeValidation, "key must contain uppercase, lowercase, and digits")
	}

	return nil
}
