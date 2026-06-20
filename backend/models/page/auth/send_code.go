package auth

import (
	"context"
	"strings"
	"time"

	"github.com/zxh3032/two-to/backend/library/apperror"
	"github.com/zxh3032/two-to/backend/library/response"
	"github.com/zxh3032/two-to/backend/library/security"
	"github.com/zxh3032/two-to/backend/library/verification"
	"github.com/zxh3032/two-to/backend/models/dao"
	"github.com/zxh3032/two-to/backend/models/page/pagectx"
	"github.com/zxh3032/two-to/backend/proto"
)

// SendCode 发送短信或邮箱验证码。流程包含目标标准化、图片验证码校验、频控、发送和审计。
func (s *Service) SendCode(ctx context.Context, req *proto.SendCodeRequest, meta pagectx.RequestMeta) (*proto.SendCodeResponse, error) {
	// 先把邮箱/手机号标准化，保证 cache key、唯一索引和审计记录使用同一种格式。
	codeType, target, err := s.normalizeCodeTarget(req.GetCodeType(), req.GetTarget())
	if err != nil {
		return nil, apperror.BadRequest(response.CodeBadRequest, err.Error())
	}
	scene := strings.TrimSpace(req.GetScene())
	if !validScene(scene) {
		return nil, apperror.BadRequest(response.CodeBadRequest, "验证码场景不正确")
	}
	if err := s.verifyCaptcha(ctx, req.GetCaptchaId(), req.GetCaptchaCode()); err != nil {
		return nil, err
	}
	// 发送验证码前做按目标、按 IP 的多维频控，避免短信/邮件接口被刷。
	if err := s.checkSendRateLimit(ctx, scene, target, security.ClientIPHash(meta.IP)); err != nil {
		return nil, err
	}
	// 邮箱注册场景需要提前判断是否已注册，避免用户走完整流程后才失败。
	if codeType == verification.CodeTypeEmail && scene == verification.SceneRegister && s.store.DBReady() {
		identity, err := s.store.FindIdentity(ctx, dao.IdentityTypeEmail, target)
		if err != nil {
			return nil, internalError(err)
		}
		if identity != nil {
			return nil, apperror.Conflict(response.CodeConflict, "该邮箱已注册，请直接登录或找回密码")
		}
	}

	// 明文验证码只写入 cache；MySQL 审计表只存 hash，避免验证码泄露风险。
	code := verification.NewDigitCode(6)
	codeKey := s.cacheStore.Key("verify-code", scene, target)
	if err := s.cacheStore.Set(ctx, codeKey, code, time.Duration(s.cfg.Verification.CodeTTLSeconds)*time.Second); err != nil {
		return nil, internalError(err)
	}

	provider := ""
	// 邮件和短信走不同 sender，但对 page 层统一成“发送 6 位验证码”的业务动作。
	if codeType == verification.CodeTypeEmail {
		provider = s.emailSender.Name()
		err = s.emailSender.SendCode(ctx, target, code, scene)
	} else {
		provider, err = s.smsSender.SendCode(ctx, target, code, scene)
	}
	if err != nil {
		// 发送失败时删除 cache 中的验证码，避免用户收到失败提示却仍能用该验证码通过校验。
		_ = s.cacheStore.Del(ctx, codeKey)
		return nil, internalError(err)
	}

	s.auditVerificationCode(ctx, codeType, scene, target, code, provider, meta)
	return &proto.SendCodeResponse{
		CooldownSeconds: s.cfg.Verification.CodeResendCooldownSeconds,
		ExpiresIn:       s.cfg.Verification.CodeTTLSeconds,
	}, nil
}
