package account

import (
	"github.com/gin-gonic/gin"
	authlib "github.com/zxh3032/two-to/backend/library/auth"
	"github.com/zxh3032/two-to/backend/proto"
)

// UpdateProfile 处理 PATCH /api/v1/account/profile，更新昵称和基础画像资料。
func (h *Handler) UpdateProfile(ctx *gin.Context) {
	var req proto.UpdateProfileRequest
	if !h.bind(ctx, "update_profile", &req) {
		return
	}
	data, err := h.service.UpdateProfile(ctx.Request.Context(), authlib.GetGinUserID(ctx), &req, meta(ctx))
	h.write(ctx, "update_profile", data, err)
}
