package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/zxh3032/two-to/backend/library/servlet"
	pageAuth "github.com/zxh3032/two-to/backend/models/page/auth"
	"github.com/zxh3032/two-to/backend/proto"
)

// PhoneLogin 处理 POST /api/v1/auth/phone-login，手机号验证码登录和注册入口共用该接口。
func PhoneLogin(rt servlet.Runtime) gin.HandlerFunc {
	service := pageAuth.NewService(rt.Config, rt.Log, rt.DB, rt.Cache, rt.TokenManager)
	return servlet.JSON[proto.PhoneLoginRequest, proto.PhoneLoginResponse](rt.Action("auth", "phone_login"), service.PhoneLogin)
}
