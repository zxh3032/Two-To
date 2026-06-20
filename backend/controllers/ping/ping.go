package ping

import (
	"github.com/gin-gonic/gin"
	"github.com/zxh3032/two-to/backend/library/config"
	"github.com/zxh3032/two-to/backend/library/requestctx"
	"github.com/zxh3032/two-to/backend/library/response"
	pingpage "github.com/zxh3032/two-to/backend/models/page/ping"
	"github.com/zxh3032/two-to/backend/proto"
	"go.uber.org/zap"
)

// Ping 验证 controller -> page -> proto 的最小调用链路。
func Ping(cfg config.Config, log *zap.Logger) gin.HandlerFunc {
	page := pingpage.New(cfg, log)
	return func(ctx *gin.Context) {
		request := &proto.PingRequest{
			RequestId: requestctx.GetGinRequestID(ctx),
		}

		result, err := page.Handle(ctx.Request.Context(), request)
		response.WriteResult(ctx, log, "ping", "handle", result, err)
	}
}
