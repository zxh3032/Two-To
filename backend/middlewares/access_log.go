package middlewares

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zxh3032/two-to/backend/library/requestctx"
	"go.uber.org/zap"
)

// AccessLog 记录请求入口和响应结果，是排查线上接口问题的第一层日志。
func AccessLog(log *zap.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		ctx.Next()

		fields := []zap.Field{
			zap.String("requestId", requestctx.GetGinRequestID(ctx)),
			zap.String("method", ctx.Request.Method),
			zap.String("path", ctx.Request.URL.Path),
			zap.Int("status", ctx.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
			zap.String("clientIP", ctx.ClientIP()),
			zap.String("userAgent", ctx.Request.UserAgent()),
		}
		if len(ctx.Errors) > 0 {
			fields = append(fields, zap.String("errors", ctx.Errors.String()))
		}

		if ctx.Writer.Status() >= 500 {
			log.Error("HTTP 请求处理失败", fields...)
			return
		}
		log.Info("HTTP 请求处理完成", fields...)
	}
}
