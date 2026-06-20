package auth

import (
	"context"
	"net/http"

	"github.com/zxh3032/two-to/backend/library/apperror"
	authlib "github.com/zxh3032/two-to/backend/library/auth"
	"github.com/zxh3032/two-to/backend/library/response"
	"github.com/zxh3032/two-to/backend/library/security"
	"github.com/zxh3032/two-to/backend/models/dao"
	"github.com/zxh3032/two-to/backend/models/page/pagectx"
	"github.com/zxh3032/two-to/backend/proto"
)

// EmailLogin 使用邮箱密码登录。连续失败超过阈值后要求图片验证码，并统一返回模糊错误。
func (s *Service) EmailLogin(ctx context.Context, req *proto.EmailLoginRequest, meta pagectx.RequestMeta) (*proto.LoginResponse, error) {
	if err := s.requireDB(); err != nil {
		return nil, err
	}
	emailValue, err := security.NormalizeEmail(req.GetEmail())
	if err != nil {
		return nil, apperror.BadRequest(response.CodeBadRequest, err.Error())
	}
	failKey := s.loginFailKey(emailValue, meta.IP)
	failCount := s.currentFailCount(ctx, failKey)
	// 失败次数达到阈值后启用图片验证码；验证码缺失时通过 data 告诉前端展示验证码。
	if failCount >= s.cfg.Verification.LoginCaptchaFailThreshold {
		if err := s.verifyCaptcha(ctx, req.GetCaptchaId(), req.GetCaptchaCode()); err != nil {
			return nil, apperror.WithData(http.StatusBadRequest, response.CodeBadRequest, "需要完成图片验证码后再登录", map[string]bool{"captchaRequired": true})
		}
	}

	// 先查身份表再查用户主表，邮箱不直接耦合在 users 表上。
	identity, err := s.store.FindIdentity(ctx, dao.IdentityTypeEmail, emailValue)
	if err != nil {
		return nil, internalError(err)
	}
	if identity == nil {
		// 不暴露邮箱是否存在，避免账号枚举；同时记录失败计数。
		s.recordLoginFailure(ctx, failKey, dao.IdentityTypeEmail, emailValue, meta)
		return nil, apperror.Unauthorized(response.CodeUnauthorized, "账号或密码不正确")
	}
	user, err := s.store.FindUser(ctx, identity.UserID)
	if err != nil {
		return nil, internalError(err)
	}
	if user == nil || user.Status != dao.UserStatusNormal || !authlib.CheckPassword(user.PasswordHash, req.GetPassword()) {
		s.recordLoginFailure(ctx, failKey, dao.IdentityTypeEmail, emailValue, meta)
		return nil, apperror.Unauthorized(response.CodeUnauthorized, "账号或密码不正确")
	}
	// 登录成功后清理失败计数，并签发新的设备会话。
	_ = s.cacheStore.Del(ctx, failKey)
	s.logSecurity(ctx, user.ID, "email_login", map[string]string{"email": security.MaskEmail(emailValue)}, meta)
	return s.issueLogin(ctx, user, meta)
}
