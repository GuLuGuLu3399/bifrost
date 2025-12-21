package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/gulugulu3399/bifrost/internal/pkg/xerr"
)

// PasswordConfig 密码配置
type PasswordConfig struct {
	MinLength      int           `yaml:"min_length"`
	RequireUpper   bool          `yaml:"require_upper"`
	RequireLower   bool          `yaml:"require_lower"`
	RequireDigit   bool          `yaml:"require_digit"`
	RequireSpecial bool          `yaml:"require_special"`
	SpecialChars   string        `yaml:"special_chars"`
	MaxAge         time.Duration `yaml:"max_age"`
}

// DefaultPasswordConfig 返回默认密码配置
func DefaultPasswordConfig() *PasswordConfig {
	return &PasswordConfig{
		MinLength:      8,
		RequireUpper:   true,
		RequireLower:   true,
		RequireDigit:   true,
		RequireSpecial: true,
		SpecialChars:   "!@#$%^&*()_+-=[]{}|;:,.<>?",
		MaxAge:         90 * 24 * time.Hour, // 90天
	}
}

// HashPassword 对密码进行加密
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", xerr.New(xerr.CodeBadRequest, "password cannot be empty")
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", xerr.Wrap(err, xerr.CodeInternal, "failed to hash password")
	}
	return string(bytes), nil
}

// HashPasswordWithCost 使用指定成本对密码进行加密
func HashPasswordWithCost(password string, cost int) (string, error) {
	if password == "" {
		return "", xerr.New(xerr.CodeBadRequest, "password cannot be empty")
	}

	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		return "", xerr.New(xerr.CodeBadRequest, fmt.Sprintf("invalid bcrypt cost: %d", cost))
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", xerr.Wrap(err, xerr.CodeInternal, "failed to hash password")
	}
	return string(bytes), nil
}

// CheckPassword 验证密码是否正确
func CheckPassword(password, hash string) bool {
	if password == "" || hash == "" {
		return false
	}

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ValidatePassword 验证密码强度
func ValidatePassword(password string, config *PasswordConfig) error {
	if config == nil {
		config = DefaultPasswordConfig()
	}

	if len(password) < config.MinLength {
		return xerr.New(xerr.CodeValidation, fmt.Sprintf("password must be at least %d characters", config.MinLength))
	}

	if config.RequireUpper && !hasUpper(password) {
		return xerr.New(xerr.CodeValidation, "password must contain at least one uppercase letter")
	}

	if config.RequireLower && !hasLower(password) {
		return xerr.New(xerr.CodeValidation, "password must contain at least one lowercase letter")
	}

	if config.RequireDigit && !hasDigit(password) {
		return xerr.New(xerr.CodeValidation, "password must contain at least one digit")
	}

	if config.RequireSpecial && !hasSpecial(password, config.SpecialChars) {
		return xerr.New(xerr.CodeValidation, fmt.Sprintf("password must contain at least one special character from: %s", config.SpecialChars))
	}

	return nil
}

// GenerateRandomPassword 生成随机密码
func GenerateRandomPassword(length int, config *PasswordConfig) (string, error) {
	if length <= 0 {
		return "", xerr.New(xerr.CodeBadRequest, "password length must be positive")
	}

	if config == nil {
		config = DefaultPasswordConfig()
	}

	// 定义字符集
	const (
		uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		lowercase = "abcdefghijklmnopqrstuvwxyz"
		digits    = "0123456789"
	)

	var charset string
	if config.RequireUpper {
		charset += uppercase
	}
	if config.RequireLower {
		charset += lowercase
	}
	if config.RequireDigit {
		charset += digits
	}
	if config.RequireSpecial {
		charset += config.SpecialChars
	}

	if charset == "" {
		charset = uppercase + lowercase + digits + config.SpecialChars
	}

	// 生成随机密码
	password := make([]byte, length)
	for i := range password {
		randomIndex, err := secureRandomInt(len(charset))
		if err != nil {
			return "", xerr.Wrap(err, xerr.CodeInternal, "failed to generate random password")
		}
		password[i] = charset[randomIndex]
	}

	// 验证生成的密码是否符合要求
	if err := ValidatePassword(string(password), config); err != nil {
		// 如果不符合，递归重新生成（避免无限循环）
		return GenerateRandomPassword(length, config)
	}

	return string(password), nil
}

// GenerateSecureToken 生成安全令牌
func GenerateSecureToken(length int) (string, error) {
	if length <= 0 {
		return "", xerr.New(xerr.CodeBadRequest, "token length must be positive")
	}

	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", xerr.Wrap(err, xerr.CodeInternal, "failed to generate secure token")
	}

	return base64.URLEncoding.EncodeToString(bytes), nil
}

// GenerateAPIKey 生成 API Key
func GenerateAPIKey() (string, error) {
	// 生成 32 字节的随机数据
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", xerr.Wrap(err, xerr.CodeInternal, "failed to generate API key")
	}

	// 使用 base64 编码，并移除可能产生歧义的字符
	key := base64.URLEncoding.EncodeToString(bytes)
	return key, nil
}

// CheckPasswordStrength 检查密码强度（返回 1-5 的评分）
func CheckPasswordStrength(password string) int {
	score := 0

	// 长度评分
	if len(password) >= 8 {
		score++
	}
	if len(password) >= 12 {
		score++
	}

	// 字符类型评分
	if hasUpper(password) {
		score++
	}
	if hasLower(password) {
		score++
	}
	if hasDigit(password) {
		score++
	}
	if hasSpecial(password, DefaultPasswordConfig().SpecialChars) {
		score++
	}

	// 确保评分在 1-5 范围内
	if score > 5 {
		score = 5
	}
	if score < 1 {
		score = 1
	}

	return score
}

// SecureCompare 安全比较两个字符串，防止时序攻击
func SecureCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}

	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}

	return subtle.ConstantTimeByteEq(result, 0) == 1
}

// ==========================================
// 辅助函数
// ==========================================

// hasUpper 检查是否包含大写字母
func hasUpper(s string) bool {
	for _, c := range s {
		if c >= 'A' && c <= 'Z' {
			return true
		}
	}
	return false
}

// hasLower 检查是否包含小写字母
func hasLower(s string) bool {
	for _, c := range s {
		if c >= 'a' && c <= 'z' {
			return true
		}
	}
	return false
}

// hasDigit 检查是否包含数字
func hasDigit(s string) bool {
	for _, c := range s {
		if c >= '0' && c <= '9' {
			return true
		}
	}
	return false
}

// hasSpecial 检查是否包含特殊字符
func hasSpecial(s, specialChars string) bool {
	for _, c := range s {
		for _, sc := range specialChars {
			if c == sc {
				return true
			}
		}
	}
	return false
}

// secureRandomInt 生成安全的随机整数
func secureRandomInt(max int) (int, error) {
	if max <= 0 {
		return 0, xerr.New(xerr.CodeBadRequest, "max must be positive")
	}

	// 计算需要的字节数
	byteLen := 1
	for (1 << (8 * byteLen)) < max {
		byteLen++
	}

	bytes := make([]byte, byteLen)
	_, err := rand.Read(bytes)
	if err != nil {
		return 0, err
	}

	// 将字节转换为整数
	result := 0
	for i, b := range bytes {
		result += int(b) << (8 * i)
	}

	return result % max, nil
}
