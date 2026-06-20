package slogan

import (
	"github.com/gin-gonic/gin"
	"github.com/zxh3032/two-to/backend/library/servlet"
	sloganpage "github.com/zxh3032/two-to/backend/models/page/slogan"
	"github.com/zxh3032/two-to/backend/proto"
	"go.uber.org/zap"
)

// Get 返回项目 slogan，作为前后端正式业务联调的第一个入口。
func Get(log *zap.Logger) gin.HandlerFunc {
	page := sloganpage.New(log)
	return servlet.Empty[proto.SloganRequest, proto.SloganResponse](servlet.Action{Module: "slogan", Name: "get", Log: log}, page.Handle)
}
