package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/zxh3032/two-to/backend/library/servlet"
	pageAuth "github.com/zxh3032/two-to/backend/models/page/auth"
	"github.com/zxh3032/two-to/backend/proto"
)

// Logout 处理 POST /api/v1/auth/logout，撤销当前 refresh token 对应的会话。
func Logout(rt servlet.Runtime) gin.HandlerFunc {
	service := pageAuth.NewService(rt.Config, rt.Log, rt.DB, rt.Cache, rt.TokenManager)
	return servlet.JSON[proto.LogoutRequest, proto.LogoutResponse](rt.Action("auth", "logout"), service.Logout)
}
