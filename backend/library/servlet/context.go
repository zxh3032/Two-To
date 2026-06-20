package servlet

import (
	"context"

	"github.com/gin-gonic/gin"
	authlib "github.com/zxh3032/two-to/backend/library/auth"
	"github.com/zxh3032/two-to/backend/library/requestctx"
)

// Context 是 page 层可见的请求上下文，屏蔽 Gin 细节，只保留业务需要的元信息。
type Context struct {
	context.Context

	RequestID string
	IP        string
	UserAgent string
	UserID    uint64
	SessionID uint64
}

func newContext(ctx *gin.Context) *Context {
	return &Context{
		Context:   ctx.Request.Context(),
		RequestID: requestctx.GetGinRequestID(ctx),
		IP:        ctx.ClientIP(),
		UserAgent: ctx.Request.UserAgent(),
		UserID:    authlib.GetGinUserID(ctx),
		SessionID: authlib.GetGinSessionID(ctx),
	}
}
