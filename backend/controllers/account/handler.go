package account

import (
	"net/http"

	"github.com/gin-gonic/gin"
	authlib "github.com/zxh3032/two-to/backend/library/auth"
	"github.com/zxh3032/two-to/backend/library/cache"
	"github.com/zxh3032/two-to/backend/library/config"
	"github.com/zxh3032/two-to/backend/library/response"
	pageAccount "github.com/zxh3032/two-to/backend/models/page/account"
	"github.com/zxh3032/two-to/backend/models/page/pagectx"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Handler 是 account 模块的 HTTP 适配层。
type Handler struct {
	service *pageAccount.Service
	log     *zap.Logger
}

// NewHandler 创建 account controller，并复用认证中间件解析出的登录态上下文。
func NewHandler(cfg config.Config, log *zap.Logger, db *gorm.DB, cacheStore cache.Store, tokenManager *authlib.TokenManager) *Handler {
	return &Handler{service: pageAccount.NewService(cfg, log, db, cacheStore, tokenManager), log: log}
}

// bind 统一解析账号安全接口的 JSON 请求体，格式错误时直接中断请求。
func (h *Handler) bind(ctx *gin.Context, action string, target interface{}) bool {
	if err := ctx.ShouldBindJSON(target); err != nil {
		response.WriteError(ctx, h.log, "account", action, http.StatusBadRequest, response.CodeBadRequest, "请求参数格式不正确", err)
		return false
	}
	return true
}

// write 统一输出成功和失败响应，controller 不关心具体业务错误分支。
func (h *Handler) write(ctx *gin.Context, action string, data interface{}, err error) {
	response.WriteResult(ctx, h.log, "account", action, data, err)
}

// meta 提取安全日志需要的客户端信息，避免 page 层依赖 Gin。
func meta(ctx *gin.Context) pagectx.RequestMeta {
	return pagectx.RequestMeta{IP: ctx.ClientIP(), UserAgent: ctx.Request.UserAgent()}
}
