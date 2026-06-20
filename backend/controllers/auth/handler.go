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
