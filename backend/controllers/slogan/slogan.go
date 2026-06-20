package slogan

import (
	"github.com/gin-gonic/gin"
	"github.com/zxh3032/two-to/backend/library/requestctx"
	"github.com/zxh3032/two-to/backend/library/response"
	sloganpage "github.com/zxh3032/two-to/backend/models/page/slogan"
	"github.com/zxh3032/two-to/backend/proto"
	"go.uber.org/zap"
)

// Get 返回项目 slogan，作为前后端正式业务联调的第一个入口。
func Get(log *zap.Logger) gin.HandlerFunc {
	page := sloganpage.New(log)
	return func(ctx *gin.Context) {
		request := &proto.SloganRequest{
			RequestId: requestctx.GetGinRequestID(ctx),
		}

		result, err := page.Handle(ctx.Request.Context(), request)
		response.WriteResult(ctx, log, "slogan", "get", result, err)
	}
}
