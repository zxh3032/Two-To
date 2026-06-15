package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/zxh3032/two-to/backend/library/config"
	"github.com/zxh3032/two-to/backend/library/security"
	"golang.org/x/crypto/bcrypt"
)

// Claims 是 access token 中最小必要的登录态信息。
type Claims struct {
	UserID    uint64 `json:"userId"`
	SessionID uint64 `json:"sessionId"`
	jwt.RegisteredClaims
}

// TokenManager 负责 access token 签发解析和 refresh token 摘要。
type TokenManager struct {
	cfg config.TokenConfig
}

// NewTokenManager 创建 token 管理器，集中持有 access/refresh token 密钥和 TTL。
func NewTokenManager(cfg config.TokenConfig) *TokenManager {
	return &TokenManager{cfg: cfg}
}

// AccessTTL 返回 access token 有效期。
func (m *TokenManager) AccessTTL() time.Duration {
	return time.Duration(m.cfg.AccessTTLSeconds) * time.Second
}

// RefreshTTL 返回 refresh token 有效期。
func (m *TokenManager) RefreshTTL() time.Duration {
	return time.Duration(m.cfg.RefreshTTLSeconds) * time.Second
}

// GenerateAccessToken 签发短期 access token，payload 只包含用户 ID 和会话 ID。
func (m *TokenManager) GenerateAccessToken(userID uint64, sessionID uint64) (string, int64, error) {
	now := time.Now()
	expireTime := now.Add(m.AccessTTL())
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID:    userID,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	})
	signed, err := token.SignedString([]byte(m.cfg.AccessSecret))
	if err != nil {
		return "", 0, err
	}
	return signed, int64(m.AccessTTL().Seconds()), nil
}

// ParseAccessToken 校验 access token 签名和过期时间，并返回登录态 claims。
func (m *TokenManager) ParseAccessToken(raw string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("不支持的 token 签名算法")
		}
		return []byte(m.cfg.AccessSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("登录态已失效")
	}
	return claims, nil
}

// GenerateRefreshToken 生成 refresh token 明文和摘要；明文只返回前端，服务端只保存摘要。
func (m *TokenManager) GenerateRefreshToken() (string, string, error) {
	token, err := security.RandomToken(48)
	if err != nil {
		return "", "", err
	}
	return token, m.HashRefreshToken(token), nil
}

// HashRefreshToken 对 refresh token 做 HMAC 摘要，避免数据库保存可直接使用的 token。
func (m *TokenManager) HashRefreshToken(token string) string {
	return security.HMACSHA256(m.cfg.RefreshSecret, token)
}

// HashCode 对验证码做 HMAC 摘要，用于验证码审计记录。
func (m *TokenManager) HashCode(value string) string {
	return security.HMACSHA256(m.cfg.AccessSecret, value)
}

// HashPassword 使用 bcrypt 生成密码哈希，数据库永远不保存密码明文。
func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

// CheckPassword 校验用户输入密码是否匹配 bcrypt 哈希。
func CheckPassword(passwordHash string, password string) bool {
	if passwordHash == "" {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)) == nil
}
