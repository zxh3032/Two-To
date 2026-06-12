package ping

import (
	"net/http"

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
		if err != nil {
			log.Error("ping 请求处理失败",
				zap.String("requestId", request.RequestId),
				zap.Error(err),
			)
			response.Error(ctx, http.StatusInternalServerError, response.CodeInternalError, "ping 处理失败")
			return
		}

		response.Success(ctx, result)
	}
}
