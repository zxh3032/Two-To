package auth

import (
	"context"

	"github.com/zxh3032/two-to/backend/library/apperror"
	"github.com/zxh3032/two-to/backend/library/response"
	"github.com/zxh3032/two-to/backend/library/security"
	"github.com/zxh3032/two-to/backend/library/verification"
	"github.com/zxh3032/two-to/backend/models/dao"
	"github.com/zxh3032/two-to/backend/models/page/pagectx"
	"github.com/zxh3032/two-to/backend/proto"
)

// PhoneLogin 使用手机号验证码登录注册一体。已绑定手机号直接登录，新手机号返回资料完善 token。
func (s *Service) PhoneLogin(ctx context.Context, req *proto.PhoneLoginRequest, meta pagectx.RequestMeta) (*proto.PhoneLoginResponse, error) {
	if err := s.requireDB(); err != nil {
		return nil, err
	}
	phone, err := security.NormalizeMainlandPhone(req.GetPhone())
	if err != nil {
		return nil, apperror.BadRequest(response.CodeBadRequest, err.Error())
	}
	if err := s.verifyCode(ctx, verification.CodeTypeSMS, verification.SceneLogin, phone, req.GetSmsCode()); err != nil {
		return nil, err
	}
	// 手机号验证码已经证明联系方式可用；若身份不存在，就先进入资料完善而不是直接创建空资料账号。
	identity, err := s.store.FindIdentity(ctx, dao.IdentityTypePhone, phone)
	if err != nil {
		return nil, internalError(err)
	}
	if identity == nil {
		token, err := s.createProfileSetupToken(ctx, verification.SetupPayload{Kind: identityTypePhoneText, Phone: phone})
		if err != nil {
			return nil, err
		}
		return &proto.PhoneLoginResponse{RequiresProfileSetup: true, ProfileSetupToken: token}, nil
	}
	user, err := s.store.FindUser(ctx, identity.UserID)
	if err != nil {
		return nil, internalError(err)
	}
	if user == nil || user.Status != dao.UserStatusNormal {
		return nil, apperror.Unauthorized(response.CodeUnauthorized, "账号状态不可用")
	}
	authResp, err := s.issueLogin(ctx, user, meta)
	if err != nil {
		return nil, err
	}
	s.logSecurity(ctx, user.ID, "phone_login", map[string]string{"phone": security.MaskPhone(phone)}, meta)
	return &proto.PhoneLoginResponse{Auth: authResp}, nil
}
