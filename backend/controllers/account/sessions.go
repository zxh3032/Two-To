package account

import (
	"github.com/gin-gonic/gin"
	"github.com/zxh3032/two-to/backend/library/servlet"
	pageAccount "github.com/zxh3032/two-to/backend/models/page/account"
	"github.com/zxh3032/two-to/backend/proto"
)

// Sessions 处理 GET /api/v1/account/sessions，列出当前账号全部登录设备。
func Sessions(rt servlet.Runtime) gin.HandlerFunc {
	service := pageAccount.NewService(rt.Config, rt.Log, rt.DB, rt.Cache, rt.TokenManager)
	return servlet.Empty[proto.SessionsRequest, proto.SessionsResponse](rt.Action("account", "sessions"), service.Sessions)
}
