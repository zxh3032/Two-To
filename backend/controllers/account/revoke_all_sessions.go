package account

import (
	"github.com/gin-gonic/gin"
	"github.com/zxh3032/two-to/backend/library/servlet"
	pageAccount "github.com/zxh3032/two-to/backend/models/page/account"
	"github.com/zxh3032/two-to/backend/proto"
)

// RevokeAllSessions 处理 POST /api/v1/account/sessions/revoke-all，校验密码后退出全部设备。
func RevokeAllSessions(rt servlet.Runtime) gin.HandlerFunc {
	service := pageAccount.NewService(rt.Config, rt.Log, rt.DB, rt.Cache, rt.TokenManager)
	return servlet.JSON[proto.RevokeAllSessionsRequest, proto.RevokeAllSessionsResponse](rt.Action("account", "revoke_all_sessions"), service.RevokeAllSessions)
}
