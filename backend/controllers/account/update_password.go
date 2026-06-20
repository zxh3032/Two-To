package account

import (
	"github.com/gin-gonic/gin"
	"github.com/zxh3032/two-to/backend/library/servlet"
	pageAccount "github.com/zxh3032/two-to/backend/models/page/account"
	"github.com/zxh3032/two-to/backend/proto"
)

// UpdatePassword 处理 PUT /api/v1/account/security/password，修改密码并撤销其他设备会话。
func UpdatePassword(rt servlet.Runtime) gin.HandlerFunc {
	service := pageAccount.NewService(rt.Config, rt.Log, rt.DB, rt.Cache, rt.TokenManager)
	return servlet.JSON[proto.UpdatePasswordRequest, proto.UpdatePasswordResponse](rt.Action("account", "update_password"), service.UpdatePassword)
}
