package requestctx

import "github.com/gin-gonic/gin"

const RequestIDKey = "requestId"

// GetGinRequestID 从 Gin 上下文读取链路追踪 ID，便于响应和业务日志保持一致。
func GetGinRequestID(ctx *gin.Context) string {
	if requestID, ok := ctx.Get(RequestIDKey); ok {
		if value, ok := requestID.(string); ok {
			return value
		}
	}
	return ""
}
