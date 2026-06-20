package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/zxh3032/two-to/backend/library/servlet"
	pageAuth "github.com/zxh3032/two-to/backend/models/page/auth"
	"github.com/zxh3032/two-to/backend/proto"
)

// SendCode 处理 POST /api/v1/auth/send-code，发送短信或邮箱验证码。
func SendCode(rt servlet.Runtime) gin.HandlerFunc {
	service := pageAuth.NewService(rt.Config, rt.Log, rt.DB, rt.Cache, rt.TokenManager)
	return servlet.JSON[proto.SendCodeRequest, proto.SendCodeResponse](rt.Action("auth", "send_code"), service.SendCode)
}
