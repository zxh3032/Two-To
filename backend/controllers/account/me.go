package account

import (
	"github.com/gin-gonic/gin"
	"github.com/zxh3032/two-to/backend/library/servlet"
	pageAccount "github.com/zxh3032/two-to/backend/models/page/account"
	"github.com/zxh3032/two-to/backend/proto"
)

// Me 处理 GET /api/v1/account/me，返回当前用户、画像和联系方式。
func Me(rt servlet.Runtime) gin.HandlerFunc {
	service := pageAccount.NewService(rt.Config, rt.Log, rt.DB, rt.Cache, rt.TokenManager)
	return servlet.Empty[proto.AccountMeRequest, proto.AccountMeResponse](rt.Action("account", "me"), service.Me)
}
