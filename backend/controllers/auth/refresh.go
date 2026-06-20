package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/zxh3032/two-to/backend/proto"
)

// Refresh 处理 POST /api/v1/auth/refresh，完成 refresh token 轮换和 access token 续签。
func (h *Handler) Refresh(ctx *gin.Context) {
	var req proto.RefreshRequest
	if !h.bind(ctx, "refresh", &req) {
		return
	}
	data, err := h.service.Refresh(ctx.Request.Context(), &req, meta(ctx))
	h.write(ctx, "refresh", data, err)
}
