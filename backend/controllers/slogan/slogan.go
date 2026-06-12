package slogan

import (
	"net/http"

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
		if err != nil {
			log.Error("项目 slogan 请求处理失败",
				zap.String("requestId", request.RequestId),
				zap.Error(err),
			)
			response.Error(ctx, http.StatusInternalServerError, response.CodeInternalError, "slogan 处理失败")
			return
		}

		response.Success(ctx, result)
	}
}
