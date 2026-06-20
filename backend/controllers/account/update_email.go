package account

import (
	"github.com/gin-gonic/gin"
	authlib "github.com/zxh3032/two-to/backend/library/auth"
	"github.com/zxh3032/two-to/backend/proto"
)

// UpdateEmail 处理 PUT /api/v1/account/security/email，绑定或更换当前账号邮箱。
func (h *Handler) UpdateEmail(ctx *gin.Context) {
	var req proto.UpdateEmailRequest
	if !h.bind(ctx, "update_email", &req) {
		return
	}
	h.write(ctx, "update_email", nil, h.service.UpdateEmail(ctx.Request.Context(), authlib.GetGinUserID(ctx), &req, meta(ctx)))
}
