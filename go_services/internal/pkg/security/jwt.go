package security

import (
	"errors"
	"fmt"
	"strconv" // [新增] 内部处理类型转换
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gulugulu3399/bifrost/internal/pkg/xerr"
)

// JWTConfig (保持不变)
type JWTConfig struct {
	SecretKey     string        `yaml:"secret_key"`
	Expiration    time.Duration `yaml:"expiration"`
	Issuer        string        `yaml:"issuer"`
	RefreshExp    time.Duration `yaml:"refresh_expiration"`
	SigningMethod string        `yaml:"signing_method"`
}

func DefaultJWTConfig() *JWTConfig {
	return &JWTConfig{
		Expiration:    24 * time.Hour,
		RefreshExp:    7 * 24 * time.Hour,
		Issuer:        "bifrost",
		SigningMethod: "HS256",
	}
}

// Claims JWT 声明
// [定制化修改]
type Claims struct {
	UserID   string `json:"user_id"` // JSON 中依然保持 string 以防前端精度丢失
	Username string `json:"username"`
	IsAdmin  bool   `json:"is_admin"` // [修改] 直接存 bool，代替原来的 Roles []string
	// TenantID 删掉，暂时用不上
	jwt.RegisteredClaims
}

type JWTManager struct {
	config *JWTConfig
}

func NewJWTManager(config *JWTConfig) (*JWTManager, error) {
	if config == nil {
		config = DefaultJWTConfig()
	}
	if config.SecretKey == "" {
		return nil, xerr.New(xerr.CodeBadRequest, "JWT secret key cannot be empty")
	}
	return &JWTManager{config: config}, nil
}

// GenerateToken 生成单 Token
// [修改] 参数直接接收 int64 和 bool，内部转 string
func (j *JWTManager) GenerateToken(userID int64, username string, isAdmin bool) (string, error) {
	if userID == 0 {
		return "", xerr.New(xerr.CodeBadRequest, "user ID cannot be zero")
	}

	now := time.Now()
	// int64 -> string
	userIDStr := strconv.FormatInt(userID, 10)

	claims := Claims{
		UserID:   userIDStr,
		Username: username,
		IsAdmin:  isAdmin, // 直接赋值
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.config.Issuer,
			Subject:   userIDStr, // Subject 也是 ID
			ExpiresAt: jwt.NewNumericDate(now.Add(j.config.Expiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.GetSigningMethod(j.config.SigningMethod), claims)
	tokenString, err := token.SignedString([]byte(j.config.SecretKey))
	if err != nil {
		return "", xerr.Wrap(err, xerr.CodeInternal, "failed to generate JWT token")
	}

	return tokenString, nil
}

// GenerateTokenPair 生成双 Token
// [修改] 参数极简
func (j *JWTManager) GenerateTokenPair(userID int64, username string, isAdmin bool) (accessToken, refreshToken string, err error) {
	// 1. 生成 Access Token
	accessToken, err = j.GenerateToken(userID, username, isAdmin)
	if err != nil {
		return "", "", err
	}

	// 2. 生成 Refresh Token
	now := time.Now()
	userIDStr := strconv.FormatInt(userID, 10)

	refreshClaims := Claims{
		UserID:   userIDStr,
		Username: username,
		IsAdmin:  isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.config.Issuer,
			Subject:   userIDStr,
			ExpiresAt: jwt.NewNumericDate(now.Add(j.config.RefreshExp)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	refreshTokenObj := jwt.NewWithClaims(jwt.GetSigningMethod(j.config.SigningMethod), refreshClaims)
	refreshToken, err = refreshTokenObj.SignedString([]byte(j.config.SecretKey))
	if err != nil {
		return "", "", xerr.Wrap(err, xerr.CodeInternal, "failed to generate refresh token")
	}

	return accessToken, refreshToken, nil
}

// ValidateToken (大体不变，只需适配新的 Claims)
func (j *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	if tokenString == "" {
		return nil, xerr.New(xerr.CodeUnauthorized, "token cannot be empty")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, xerr.New(xerr.CodeUnauthorized, fmt.Sprintf("unexpected signing method: %v", token.Header["alg"]))
		}
		return []byte(j.config.SecretKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, xerr.New(xerr.CodeUnauthorized, "token has expired")
		}
		return nil, xerr.Wrap(err, xerr.CodeUnauthorized, "failed to parse token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, xerr.New(xerr.CodeUnauthorized, "invalid token")
	}

	if claims.Issuer != j.config.Issuer {
		return nil, xerr.New(xerr.CodeUnauthorized, "invalid token issuer")
	}

	return claims, nil
}

// RefreshToken 刷新逻辑
func (j *JWTManager) RefreshToken(refreshTokenString string) (string, error) {
	claims, err := j.ValidateToken(refreshTokenString)
	if err != nil {
		return "", xerr.Wrap(err, xerr.CodeUnauthorized, "invalid refresh token")
	}

	// 解析 string ID 回 int64 (为了保持接口一致性)
	uid, err := strconv.ParseInt(claims.UserID, 10, 64)
	if err != nil {
		return "", xerr.New(xerr.CodeUnauthorized, "invalid user id in token")
	}

	return j.GenerateToken(uid, claims.Username, claims.IsAdmin)
}
