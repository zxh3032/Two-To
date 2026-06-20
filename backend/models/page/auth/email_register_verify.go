package auth

import (
	"github.com/zxh3032/two-to/backend/library/apperror"
	authlib "github.com/zxh3032/two-to/backend/library/auth"
	"github.com/zxh3032/two-to/backend/library/response"
	"github.com/zxh3032/two-to/backend/library/security"
	"github.com/zxh3032/two-to/backend/library/servlet"
	"github.com/zxh3032/two-to/backend/library/verification"
	"github.com/zxh3032/two-to/backend/models/dao"
	"github.com/zxh3032/two-to/backend/proto"
	"go.uber.org/zap"
)

// EmailRegisterVerify 校验邮箱注册验证码和密码规则，成功后签发资料完善 token。
func (s *Service) EmailRegisterVerify(ctx *servlet.Context, req *proto.EmailRegisterVerifyRequest, resp *proto.EmailRegisterVerifyResponse) error {
	if err := s.requireDB(); err != nil {
		return err
	}
	emailValue, err := security.NormalizeEmail(req.GetEmail())
	if err != nil {
		return apperror.BadRequest(response.CodeBadRequest, err.Error())
	}
	if err := security.ValidatePassword(req.GetPassword(), emailValue); err != nil {
		return apperror.BadRequest(response.CodeBadRequest, err.Error())
	}
	// 邮箱注册分两步：这里仅做验证和密码哈希，不创建用户，避免缺失画像数据的账号落库。
	identity, err := s.store.FindIdentity(ctx, dao.IdentityTypeEmail, emailValue)
	if err != nil {
		return internalError(err)
	}
	if identity != nil {
		return apperror.Conflict(response.CodeConflict, "该邮箱已注册，请直接登录或找回密码")
	}
	if err := s.verifyCode(ctx, verification.CodeTypeEmail, verification.SceneRegister, emailValue, req.GetEmailCode()); err != nil {
		return err
	}
	passwordHash, err := authlib.HashPassword(req.GetPassword())
	if err != nil {
		return internalError(err)
	}
	token, err := s.createProfileSetupToken(ctx, verification.SetupPayload{Kind: identityTypeEmailText, Email: emailValue, PasswordHash: passwordHash})
	if err != nil {
		return err
	}
	s.log.Debug("邮箱注册验证码校验完成，进入资料完善", zap.String("email", security.MaskEmail(emailValue)), zap.String("ipHash", security.ClientIPHash(ctx.IP)))
	resp.ProfileSetupToken = token
	return nil
}
