package middlewares

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/zxh3032/two-to/backend/library/requestctx"
	"github.com/zxh3032/two-to/backend/library/response"
	"go.uber.org/zap"
)

// Recovery 捕获未处理 panic，记录堆栈后返回统一错误响应。
func Recovery(log *zap.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Error("HTTP 请求发生 panic",
					zap.String("requestId", requestctx.GetGinRequestID(ctx)),
					zap.Any("panic", recovered),
					zap.ByteString("stack", debug.Stack()),
				)
				response.Error(ctx, http.StatusInternalServerError, response.CodeInternalError, "服务器内部错误")
			}
		}()

		ctx.Next()
	}
}
