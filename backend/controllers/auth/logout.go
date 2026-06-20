package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/zxh3032/two-to/backend/proto"
)

// Logout 处理 POST /api/v1/auth/logout，撤销当前 refresh token 对应的会话。
func (h *Handler) Logout(ctx *gin.Context) {
	var req proto.LogoutRequest
	if !h.bind(ctx, "logout", &req) {
		return
	}
	h.write(ctx, "logout", nil, h.service.Logout(ctx.Request.Context(), &req))
}
