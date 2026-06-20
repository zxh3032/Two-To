package account

import (
	"github.com/gin-gonic/gin"
	"github.com/zxh3032/two-to/backend/library/servlet"
	pageAccount "github.com/zxh3032/two-to/backend/models/page/account"
	"github.com/zxh3032/two-to/backend/proto"
)

// UnbindPhone 处理 DELETE /api/v1/account/security/phone，解绑手机号前需要校验当前密码。
func UnbindPhone(rt servlet.Runtime) gin.HandlerFunc {
	service := pageAccount.NewService(rt.Config, rt.Log, rt.DB, rt.Cache, rt.TokenManager)
	return servlet.JSON[proto.UnbindPhoneRequest, proto.UnbindPhoneResponse](rt.Action("account", "unbind_phone"), service.UnbindPhone)
}
