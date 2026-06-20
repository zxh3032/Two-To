package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/zxh3032/two-to/backend/library/servlet"
	pageAuth "github.com/zxh3032/two-to/backend/models/page/auth"
	"github.com/zxh3032/two-to/backend/proto"
)

// EmailRegisterVerify 处理 POST /api/v1/auth/email-register/verify，只校验邮箱注册条件并签发资料完善 token。
func EmailRegisterVerify(rt servlet.Runtime) gin.HandlerFunc {
	service := pageAuth.NewService(rt.Config, rt.Log, rt.DB, rt.Cache, rt.TokenManager)
	return servlet.JSON[proto.EmailRegisterVerifyRequest, proto.EmailRegisterVerifyResponse](rt.Action("auth", "email_register_verify"), service.EmailRegisterVerify)
}
