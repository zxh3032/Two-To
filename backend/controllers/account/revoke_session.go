package account

import (
	"github.com/gin-gonic/gin"
	"github.com/zxh3032/two-to/backend/library/servlet"
	pageAccount "github.com/zxh3032/two-to/backend/models/page/account"
	"github.com/zxh3032/two-to/backend/proto"
)

// RevokeSession 处理 DELETE /api/v1/account/sessions/:sessionId，踢出当前用户的指定非当前设备。
func RevokeSession(rt servlet.Runtime) gin.HandlerFunc {
	service := pageAccount.NewService(rt.Config, rt.Log, rt.DB, rt.Cache, rt.TokenManager)
	return servlet.Path[proto.RevokeSessionRequest, proto.RevokeSessionResponse](
		rt.Action("account", "revoke_session"),
		servlet.Uint64Path("sessionId", "设备 ID 不正确", func(req *proto.RevokeSessionRequest, value uint64) {
			req.SessionId = value
		}),
		service.RevokeSession,
	)
}
