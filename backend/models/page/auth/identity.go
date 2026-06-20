package auth

import (
	"errors"
	"strings"

	"github.com/zxh3032/two-to/backend/library/security"
	"github.com/zxh3032/two-to/backend/library/verification"
	"github.com/zxh3032/two-to/backend/models/dao"
)

// normalizeCodeTarget 将验证码发送目标统一转换为系统内部格式，手机号当前只支持中国大陆 +86。
func (s *Service) normalizeCodeTarget(codeTypeText string, target string) (int32, string, error) {
	switch strings.ToLower(strings.TrimSpace(codeTypeText)) {
	case "sms":
		value, err := security.NormalizeMainlandPhone(target)
		return verification.CodeTypeSMS, value, err
	case "email":
		value, err := security.NormalizeEmail(target)
		return verification.CodeTypeEmail, value, err
	default:
		return 0, "", errors.New("验证码类型不正确")
	}
}

// normalizeIdentity 将找回密码入参中的账号类型和值转换成身份表枚举和值。
func (s *Service) normalizeIdentity(identityTypeText string, value string) (int32, string, error) {
	switch strings.ToLower(strings.TrimSpace(identityTypeText)) {
	case identityTypeEmailText:
		normalized, err := security.NormalizeEmail(value)
		return dao.IdentityTypeEmail, normalized, err
	case identityTypePhoneText:
		normalized, err := security.NormalizeMainlandPhone(value)
		return dao.IdentityTypePhone, normalized, err
	default:
		return 0, "", errors.New("账号类型不正确")
	}
}
