package account

import (
	"github.com/gin-gonic/gin"
	"github.com/zxh3032/two-to/backend/library/servlet"
	pageAccount "github.com/zxh3032/two-to/backend/models/page/account"
	"github.com/zxh3032/two-to/backend/proto"
)

// UpdateProfile 处理 PATCH /api/v1/account/profile，更新昵称和基础画像资料。
func UpdateProfile(rt servlet.Runtime) gin.HandlerFunc {
	service := pageAccount.NewService(rt.Config, rt.Log, rt.DB, rt.Cache, rt.TokenManager)
	return servlet.JSON[proto.UpdateProfileRequest, proto.UpdateProfileResponse](rt.Action("account", "update_profile"), service.UpdateProfile)
}
