package security

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"regexp"
	"strings"
	"unicode/utf8"
)

var (
	emailRegexp       = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	mainlandPhoneExpr = regexp.MustCompile(`^1[3-9]\d{9}$`)
)

// RandomToken 生成 URL 安全的随机 token，用于 refresh token 和短期流程凭证。
func RandomToken(byteSize int) (string, error) {
	buf := make([]byte, byteSize)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// HMACSHA256 使用密钥生成稳定摘要，避免明文 token、验证码、IP 入库。
func HMACSHA256(secret string, value string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(value))
	return hex.EncodeToString(mac.Sum(nil))
}

// HashPlain 对非密钥场景生成不可逆摘要，主要用于 IP 和 User-Agent。
func HashPlain(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

// NormalizeEmail 标准化邮箱并做基础格式校验，系统内部统一保存小写邮箱。
func NormalizeEmail(email string) (string, error) {
	value := strings.ToLower(strings.TrimSpace(email))
	if !emailRegexp.MatchString(value) || len(value) > 128 {
		return "", errors.New("邮箱格式不正确")
	}
	return value, nil
}

// NormalizeMainlandPhone 标准化中国大陆手机号，系统内部统一保存为 +86 开头。
func NormalizeMainlandPhone(phone string) (string, error) {
	value := strings.TrimSpace(phone)
	value = strings.TrimPrefix(value, "+86")
	value = strings.TrimPrefix(value, "86")
	value = strings.ReplaceAll(value, " ", "")
	value = strings.ReplaceAll(value, "-", "")
	if !mainlandPhoneExpr.MatchString(value) {
		return "", errors.New("手机号格式不正确")
	}
	return "+86" + value, nil
}

// MaskEmail 对邮箱做脱敏展示，安全日志和前端展示不能暴露完整邮箱。
func MaskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "****"
	}
	name := []rune(parts[0])
	if len(name) <= 2 {
		return fmt.Sprintf("%s***@%s", string(name[:1]), parts[1])
	}
	return fmt.Sprintf("%s***%s@%s", string(name[:1]), string(name[len(name)-1:]), parts[1])
}

// MaskPhone 对手机号做脱敏展示，只保留前三后四。
func MaskPhone(phone string) string {
	value := strings.TrimPrefix(phone, "+86")
	if len(value) != 11 {
		return "***********"
	}
	return value[:3] + "****" + value[7:]
}

// MaskIdentity 按身份类型选择邮箱或手机号脱敏规则。
func MaskIdentity(identityType int32, value string) string {
	if identityType == 1 {
		return MaskEmail(value)
	}
	return MaskPhone(value)
}

// ClientIPHash 标准化 IP 后生成摘要，用于安全日志和风控计数。
func ClientIPHash(ip string) string {
	parsed := net.ParseIP(strings.TrimSpace(ip))
	if parsed == nil {
		return HashPlain(ip)
	}
	return HashPlain(parsed.String())
}

// ValidatePassword 校验首期密码规则，避免明显弱密码进入账号体系。
func ValidatePassword(password string, identityValues ...string) error {
	if utf8.RuneCountInString(password) < 8 || utf8.RuneCountInString(password) > 64 {
		return errors.New("密码需为 8-64 位")
	}
	if strings.TrimSpace(password) == "" {
		return errors.New("密码不能全部为空格")
	}
	lower := strings.ToLower(password)
	weakPasswords := map[string]struct{}{
		"12345678":  {},
		"password":  {},
		"qwerty123": {},
		"11111111":  {},
		"00000000":  {},
	}
	if _, ok := weakPasswords[lower]; ok {
		return errors.New("密码过于简单")
	}
	for _, identity := range identityValues {
		identity = strings.ToLower(strings.TrimSpace(identity))
		if identity != "" && strings.Contains(lower, identity) {
			return errors.New("密码不能包含邮箱或手机号")
		}
	}
	return nil
}
