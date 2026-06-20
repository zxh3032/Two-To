package account

import (
	"github.com/gin-gonic/gin"
	authlib "github.com/zxh3032/two-to/backend/library/auth"
	"github.com/zxh3032/two-to/backend/proto"
)

// UpdatePassword 处理 PUT /api/v1/account/security/password，修改密码并撤销其他设备会话。
func (h *Handler) UpdatePassword(ctx *gin.Context) {
	var req proto.UpdatePasswordRequest
	if !h.bind(ctx, "update_password", &req) {
		return
	}
	h.write(ctx, "update_password", nil, h.service.UpdatePassword(ctx.Request.Context(), authlib.GetGinUserID(ctx), authlib.GetGinSessionID(ctx), &req, meta(ctx)))
}
