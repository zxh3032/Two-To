package auth

import (
	"context"
	"time"

	"github.com/zxh3032/two-to/backend/library/apperror"
	"github.com/zxh3032/two-to/backend/library/response"
	"github.com/zxh3032/two-to/backend/library/security"
	"github.com/zxh3032/two-to/backend/library/verification"
	"github.com/zxh3032/two-to/backend/models/dao"
	"github.com/zxh3032/two-to/backend/proto"
)

// ForgotPasswordVerifyCode 校验找回密码验证码，账号存在时签发短期 reset token。
func (s *Service) ForgotPasswordVerifyCode(ctx context.Context, req *proto.ForgotPasswordVerifyCodeRequest) (*proto.ForgotPasswordVerifyCodeResponse, error) {
	if err := s.requireDB(); err != nil {
		return nil, err
	}
	identityType, target, err := s.normalizeIdentity(req.GetIdentityType(), req.GetIdentityValue())
	if err != nil {
		return nil, apperror.BadRequest(response.CodeBadRequest, err.Error())
	}
	codeType := verification.CodeTypeEmail
	if identityType == dao.IdentityTypePhone {
		codeType = verification.CodeTypeSMS
	}
	if err := s.verifyCode(ctx, codeType, verification.SceneForgotPassword, target, req.GetCode()); err != nil {
		return nil, err
	}
	identity, err := s.store.FindIdentity(ctx, identityType, target)
	if err != nil {
		return nil, internalError(err)
	}
	if identity == nil {
		// 找回密码不暴露账号是否存在，避免邮箱/手机号被枚举。
		return &proto.ForgotPasswordVerifyCodeResponse{}, nil
	}
	token, err := security.RandomToken(32)
	if err != nil {
		return nil, internalError(err)
	}
	key := s.cacheStore.Key("password-reset", token)
	if err := verification.StoreJSON(ctx, s.cacheStore, key, verification.ResetPayload{UserID: identity.UserID}, time.Duration(s.cfg.Verification.PasswordResetTTLSeconds)*time.Second); err != nil {
		return nil, internalError(err)
	}
	return &proto.ForgotPasswordVerifyCodeResponse{ResetToken: token}, nil
}
