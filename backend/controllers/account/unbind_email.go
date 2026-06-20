package account

import (
	"github.com/gin-gonic/gin"
	"github.com/zxh3032/two-to/backend/library/servlet"
	pageAccount "github.com/zxh3032/two-to/backend/models/page/account"
	"github.com/zxh3032/two-to/backend/proto"
)

// UnbindEmail 处理 DELETE /api/v1/account/security/email，解绑邮箱前需要校验当前密码。
func UnbindEmail(rt servlet.Runtime) gin.HandlerFunc {
	service := pageAccount.NewService(rt.Config, rt.Log, rt.DB, rt.Cache, rt.TokenManager)
	return servlet.JSON[proto.UnbindEmailRequest, proto.UnbindEmailResponse](rt.Action("account", "unbind_email"), service.UnbindEmail)
}
