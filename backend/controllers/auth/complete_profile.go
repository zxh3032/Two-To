package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/zxh3032/two-to/backend/proto"
)

// CompleteProfile 处理 POST /api/v1/auth/complete-profile，提交基础资料后正式创建账号并返回 token。
func (h *Handler) CompleteProfile(ctx *gin.Context) {
	var req proto.CompleteProfileRequest
	if !h.bind(ctx, "complete_profile", &req) {
		return
	}
	data, err := h.service.CompleteProfile(ctx.Request.Context(), &req, meta(ctx))
	h.write(ctx, "complete_profile", data, err)
}
