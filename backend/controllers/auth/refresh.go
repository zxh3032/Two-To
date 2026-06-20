package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/zxh3032/two-to/backend/library/servlet"
	pageAuth "github.com/zxh3032/two-to/backend/models/page/auth"
	"github.com/zxh3032/two-to/backend/proto"
)

// Refresh 处理 POST /api/v1/auth/refresh，完成 refresh token 轮换和 access token 续签。
func Refresh(rt servlet.Runtime) gin.HandlerFunc {
	service := pageAuth.NewService(rt.Config, rt.Log, rt.DB, rt.Cache, rt.TokenManager)
	return servlet.JSON[proto.RefreshRequest, proto.RefreshResponse](rt.Action("auth", "refresh"), service.Refresh)
}
