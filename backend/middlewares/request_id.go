package middlewares

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zxh3032/two-to/backend/library/requestctx"
)

const requestIDHeader = "X-Request-Id"

// RequestID 为每个请求生成或透传链路追踪 ID，方便前后端联调和线上日志串联。
func RequestID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestID := ctx.GetHeader(requestIDHeader)
		if requestID == "" {
			requestID = newRequestID()
		}

		ctx.Set(requestctx.RequestIDKey, requestID)
		ctx.Writer.Header().Set(requestIDHeader, requestID)
		ctx.Next()
	}
}

// newRequestID 优先使用随机数生成追踪 ID，随机源异常时用时间戳兜底。
func newRequestID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err == nil {
		return hex.EncodeToString(bytes)
	}
	return time.Now().Format("20060102150405.000000000")
}
