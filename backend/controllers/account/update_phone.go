package account

import (
	"github.com/gin-gonic/gin"
	authlib "github.com/zxh3032/two-to/backend/library/auth"
	"github.com/zxh3032/two-to/backend/proto"
)

// UpdatePhone 处理 PUT /api/v1/account/security/phone，绑定或更换当前账号手机号。
func (h *Handler) UpdatePhone(ctx *gin.Context) {
	var req proto.UpdatePhoneRequest
	if !h.bind(ctx, "update_phone", &req) {
		return
	}
	h.write(ctx, "update_phone", nil, h.service.UpdatePhone(ctx.Request.Context(), authlib.GetGinUserID(ctx), &req, meta(ctx)))
}
