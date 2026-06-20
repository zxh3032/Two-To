package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	authlib "github.com/zxh3032/two-to/backend/library/auth"
	"github.com/zxh3032/two-to/backend/library/cache"
	"github.com/zxh3032/two-to/backend/library/config"
	"github.com/zxh3032/two-to/backend/library/response"
	pageAuth "github.com/zxh3032/two-to/backend/models/page/auth"
	"github.com/zxh3032/two-to/backend/models/page/pagectx"
	"github.com/zxh3032/two-to/backend/proto"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Handler 是 auth 模块的 HTTP 适配层。
type Handler struct {
	service *pageAuth.Service
	log     *zap.Logger
}

// NewHandler 创建 auth controller，并在这里完成 page service 的依赖装配。
func NewHandler(cfg config.Config, log *zap.Logger, db *gorm.DB, cacheStore cache.Store, tokenManager *authlib.TokenManager) *Handler {
	return &Handler{service: pageAuth.NewService(cfg, log, db, cacheStore, tokenManager), log: log}
}

// Captcha 处理 GET /api/v1/auth/captcha，返回图片验证码 ID、base64 图片和有效期。
func (h *Handler) Captcha(ctx *gin.Context) {
	data, err := h.service.Captcha(ctx.Request.Context())
	h.write(ctx, "captcha", data, err)
}

// SendCode 处理 POST /api/v1/auth/send-code，发送短信或邮箱验证码。
func (h *Handler) SendCode(ctx *gin.Context) {
	var req proto.SendCodeRequest
	if !h.bind(ctx, "send_code", &req) {
		return
	}
	data, err := h.service.SendCode(ctx.Request.Context(), &req, meta(ctx))
	h.write(ctx, "send_code", data, err)
}

// EmailLogin 处理 POST /api/v1/auth/email-login，使用邮箱和密码换取登录态。
func (h *Handler) EmailLogin(ctx *gin.Context) {
	var req proto.EmailLoginRequest
	if !h.bind(ctx, "email_login", &req) {
		return
	}
	data, err := h.service.EmailLogin(ctx.Request.Context(), &req, meta(ctx))
	h.write(ctx, "email_login", data, err)
}

// PhoneLogin 处理 POST /api/v1/auth/phone-login，手机号验证码登录和注册入口共用该接口。
func (h *Handler) PhoneLogin(ctx *gin.Context) {
	var req proto.PhoneLoginRequest
	if !h.bind(ctx, "phone_login", &req) {
		return
	}
	data, err := h.service.PhoneLogin(ctx.Request.Context(), &req, meta(ctx))
	h.write(ctx, "phone_login", data, err)
}

// EmailRegisterVerify 处理 POST /api/v1/auth/email-register/verify，只校验邮箱注册条件并签发资料完善 token。
func (h *Handler) EmailRegisterVerify(ctx *gin.Context) {
	var req proto.EmailRegisterVerifyRequest
	if !h.bind(ctx, "email_register_verify", &req) {
		return
	}
	data, err := h.service.EmailRegisterVerify(ctx.Request.Context(), &req, meta(ctx))
	h.write(ctx, "email_register_verify", data, err)
}

// CompleteProfile 处理 POST /api/v1/auth/complete-profile，提交基础资料后正式创建账号并返回 token。
func (h *Handler) CompleteProfile(ctx *gin.Context) {
	var req proto.CompleteProfileRequest
	if !h.bind(ctx, "complete_profile", &req) {
		return
	}
	data, err := h.service.CompleteProfile(ctx.Request.Context(), &req, meta(ctx))
	h.write(ctx, "complete_profile", data, err)
}

// Refresh 处理 POST /api/v1/auth/refresh，完成 refresh token 轮换和 access token 续签。
func (h *Handler) Refresh(ctx *gin.Context) {
	var req proto.RefreshRequest
	if !h.bind(ctx, "refresh", &req) {
		return
	}
	data, err := h.service.Refresh(ctx.Request.Context(), &req, meta(ctx))
	h.write(ctx, "refresh", data, err)
}

// Logout 处理 POST /api/v1/auth/logout，撤销当前 refresh token 对应的会话。
func (h *Handler) Logout(ctx *gin.Context) {
	var req proto.LogoutRequest
	if !h.bind(ctx, "logout", &req) {
		return
	}
	h.write(ctx, "logout", nil, h.service.Logout(ctx.Request.Context(), &req))
}

// ForgotPasswordVerifyCode 处理 POST /api/v1/auth/forgot-password/verify-code，验证码通过后返回重置密码 token。
func (h *Handler) ForgotPasswordVerifyCode(ctx *gin.Context) {
	var req proto.ForgotPasswordVerifyCodeRequest
	if !h.bind(ctx, "forgot_password_verify_code", &req) {
		return
	}
	data, err := h.service.ForgotPasswordVerifyCode(ctx.Request.Context(), &req)
	h.write(ctx, "forgot_password_verify_code", data, err)
}

// ForgotPasswordReset 处理 POST /api/v1/auth/forgot-password/reset，使用重置 token 设置新密码。
func (h *Handler) ForgotPasswordReset(ctx *gin.Context) {
	var req proto.ForgotPasswordResetRequest
	if !h.bind(ctx, "forgot_password_reset", &req) {
		return
	}
	h.write(ctx, "forgot_password_reset", nil, h.service.ForgotPasswordReset(ctx.Request.Context(), &req, meta(ctx)))
}

// bind 统一解析 JSON 请求体，参数格式错误时直接输出标准错误响应，避免各接口重复处理。
func (h *Handler) bind(ctx *gin.Context, action string, target interface{}) bool {
	if err := ctx.ShouldBindJSON(target); err != nil {
		response.WriteError(ctx, h.log, "auth", action, http.StatusBadRequest, response.CodeBadRequest, "请求参数格式不正确", err)
		return false
	}
	return true
}

// write 统一把 page 层返回值转换为 HTTP 响应，保证业务错误码、requestId 和 data 结构一致。
func (h *Handler) write(ctx *gin.Context, action string, data interface{}, err error) {
	response.WriteResult(ctx, h.log, "auth", action, data, err)
}

// meta 从 Gin 请求中提取风控和审计需要的 IP、User-Agent，不把 HTTP 对象传入 page 层。
func meta(ctx *gin.Context) pagectx.RequestMeta {
	return pagectx.RequestMeta{IP: ctx.ClientIP(), UserAgent: ctx.Request.UserAgent()}
}
