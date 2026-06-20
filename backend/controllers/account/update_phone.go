package account

import (
	"github.com/gin-gonic/gin"
	"github.com/zxh3032/two-to/backend/library/servlet"
	pageAccount "github.com/zxh3032/two-to/backend/models/page/account"
	"github.com/zxh3032/two-to/backend/proto"
)

// UpdatePhone 处理 PUT /api/v1/account/security/phone，绑定或更换当前账号手机号。
func UpdatePhone(rt servlet.Runtime) gin.HandlerFunc {
	service := pageAccount.NewService(rt.Config, rt.Log, rt.DB, rt.Cache, rt.TokenManager)
	return servlet.JSON[proto.UpdatePhoneRequest, proto.UpdatePhoneResponse](rt.Action("account", "update_phone"), service.UpdatePhone)
}
