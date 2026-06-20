package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/zxh3032/two-to/backend/library/servlet"
	pageAuth "github.com/zxh3032/two-to/backend/models/page/auth"
	"github.com/zxh3032/two-to/backend/proto"
)

// CompleteProfile 处理 POST /api/v1/auth/complete-profile，提交基础资料后正式创建账号并返回 token。
func CompleteProfile(rt servlet.Runtime) gin.HandlerFunc {
	service := pageAuth.NewService(rt.Config, rt.Log, rt.DB, rt.Cache, rt.TokenManager)
	return servlet.JSON[proto.CompleteProfileRequest, proto.CompleteProfileResponse](rt.Action("auth", "complete_profile"), service.CompleteProfile)
}
