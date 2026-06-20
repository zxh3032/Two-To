package ping

import (
	"github.com/gin-gonic/gin"
	"github.com/zxh3032/two-to/backend/library/config"
	"github.com/zxh3032/two-to/backend/library/servlet"
	pingpage "github.com/zxh3032/two-to/backend/models/page/ping"
	"github.com/zxh3032/two-to/backend/proto"
	"go.uber.org/zap"
)

// Ping 验证 controller -> page -> proto 的最小调用链路。
func Ping(cfg config.Config, log *zap.Logger) gin.HandlerFunc {
	page := pingpage.New(cfg, log)
	return servlet.Empty[proto.PingRequest, proto.PingResponse](servlet.Action{Module: "ping", Name: "handle", Log: log}, page.Handle)
}
