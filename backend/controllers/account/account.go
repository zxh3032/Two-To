package account

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zxh3032/two-to/backend/library/apperror"
	authlib "github.com/zxh3032/two-to/backend/library/auth"
	"github.com/zxh3032/two-to/backend/library/cache"
	"github.com/zxh3032/two-to/backend/library/config"
	"github.com/zxh3032/two-to/backend/library/response"
	pageAccount "github.com/zxh3032/two-to/backend/models/page/account"
	"github.com/zxh3032/two-to/backend/models/page/pagectx"
	"github.com/zxh3032/two-to/backend/proto"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Handler 是 account 模块的 HTTP 适配层。
type Handler struct {
	service *pageAccount.Service
}

// NewHandler 创建 account controller，并复用认证中间件解析出的登录态上下文。
func NewHandler(cfg config.Config, log *zap.Logger, db *gorm.DB, cacheStore cache.Store, tokenManager *authlib.TokenManager) *Handler {
	return &Handler{service: pageAccount.NewService(cfg, log, db, cacheStore, tokenManager)}
}

// Me 处理 GET /api/v1/account/me，返回当前用户、画像和联系方式。
func (h *Handler) Me(ctx *gin.Context) {
	data, err := h.service.Me(ctx.Request.Context(), authlib.GetGinUserID(ctx))
	write(ctx, data, err)
}

// UpdateProfile 处理 PATCH /api/v1/account/profile，更新昵称和基础画像资料。
func (h *Handler) UpdateProfile(ctx *gin.Context) {
	var req proto.UpdateProfileRequest
	if !bind(ctx, &req) {
		return
	}
	data, err := h.service.UpdateProfile(ctx.Request.Context(), authlib.GetGinUserID(ctx), &req, meta(ctx))
	write(ctx, data, err)
}

// UpdateEmail 处理 PUT /api/v1/account/security/email，绑定或更换当前账号邮箱。
func (h *Handler) UpdateEmail(ctx *gin.Context) {
	var req proto.UpdateEmailRequest
	if !bind(ctx, &req) {
		return
	}
	write(ctx, nil, h.service.UpdateEmail(ctx.Request.Context(), authlib.GetGinUserID(ctx), &req, meta(ctx)))
}

// UpdatePhone 处理 PUT /api/v1/account/security/phone，绑定或更换当前账号手机号。
func (h *Handler) UpdatePhone(ctx *gin.Context) {
	var req proto.UpdatePhoneRequest
	if !bind(ctx, &req) {
		return
	}
	write(ctx, nil, h.service.UpdatePhone(ctx.Request.Context(), authlib.GetGinUserID(ctx), &req, meta(ctx)))
}

// UnbindEmail 处理 DELETE /api/v1/account/security/email，解绑邮箱前需要校验当前密码。
func (h *Handler) UnbindEmail(ctx *gin.Context) {
	var req proto.CurrentPasswordRequest
	if !bind(ctx, &req) {
		return
	}
	write(ctx, nil, h.service.UnbindEmail(ctx.Request.Context(), authlib.GetGinUserID(ctx), req.GetCurrentPassword(), meta(ctx)))
}

// UnbindPhone 处理 DELETE /api/v1/account/security/phone，解绑手机号前需要校验当前密码。
func (h *Handler) UnbindPhone(ctx *gin.Context) {
	var req proto.CurrentPasswordRequest
	if !bind(ctx, &req) {
		return
	}
	write(ctx, nil, h.service.UnbindPhone(ctx.Request.Context(), authlib.GetGinUserID(ctx), req.GetCurrentPassword(), meta(ctx)))
}

// UpdatePassword 处理 PUT /api/v1/account/security/password，修改密码并撤销其他设备会话。
func (h *Handler) UpdatePassword(ctx *gin.Context) {
	var req proto.UpdatePasswordRequest
	if !bind(ctx, &req) {
		return
	}
	write(ctx, nil, h.service.UpdatePassword(ctx.Request.Context(), authlib.GetGinUserID(ctx), authlib.GetGinSessionID(ctx), &req, meta(ctx)))
}

// Sessions 处理 GET /api/v1/account/sessions，列出当前账号全部登录设备。
func (h *Handler) Sessions(ctx *gin.Context) {
	data, err := h.service.Sessions(ctx.Request.Context(), authlib.GetGinUserID(ctx), authlib.GetGinSessionID(ctx))
	write(ctx, data, err)
}

// RevokeSession 处理 DELETE /api/v1/account/sessions/:sessionId，踢出当前用户的指定非当前设备。
func (h *Handler) RevokeSession(ctx *gin.Context) {
	sessionID, err := strconv.ParseUint(ctx.Param("sessionId"), 10, 64)
	if err != nil || sessionID == 0 {
		response.Error(ctx, http.StatusBadRequest, response.CodeBadRequest, "设备 ID 不正确")
		return
	}
	write(ctx, nil, h.service.RevokeSession(ctx.Request.Context(), authlib.GetGinUserID(ctx), authlib.GetGinSessionID(ctx), sessionID, meta(ctx)))
}

// RevokeAllSessions 处理 POST /api/v1/account/sessions/revoke-all，校验密码后退出全部设备。
func (h *Handler) RevokeAllSessions(ctx *gin.Context) {
	var req proto.CurrentPasswordRequest
	if !bind(ctx, &req) {
		return
	}
	write(ctx, nil, h.service.RevokeAllSessions(ctx.Request.Context(), authlib.GetGinUserID(ctx), req.GetCurrentPassword(), meta(ctx)))
}

// bind 统一解析账号安全接口的 JSON 请求体，格式错误时直接中断请求。
func bind(ctx *gin.Context, target interface{}) bool {
	if err := ctx.ShouldBindJSON(target); err != nil {
		response.Error(ctx, http.StatusBadRequest, response.CodeBadRequest, "请求参数格式不正确")
		return false
	}
	return true
}

// write 统一输出成功和失败响应，controller 不关心具体业务错误分支。
func write(ctx *gin.Context, data interface{}, err error) {
	if err == nil {
		response.Success(ctx, data)
		return
	}
	var appErr *apperror.Error
	if errors.As(err, &appErr) {
		if appErr.Data != nil {
			response.ErrorData(ctx, appErr.HTTPStatus, appErr.Code, appErr.Message, appErr.Data)
			return
		}
		response.Error(ctx, appErr.HTTPStatus, appErr.Code, appErr.Message)
		return
	}
	response.Error(ctx, http.StatusInternalServerError, response.CodeInternalError, "服务暂时不可用")
}

// meta 提取安全日志需要的客户端信息，避免 page 层依赖 Gin。
func meta(ctx *gin.Context) pagectx.RequestMeta {
	return pagectx.RequestMeta{IP: ctx.ClientIP(), UserAgent: ctx.Request.UserAgent()}
}
