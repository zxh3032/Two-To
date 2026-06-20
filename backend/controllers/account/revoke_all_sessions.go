package account

import (
	"github.com/gin-gonic/gin"
	authlib "github.com/zxh3032/two-to/backend/library/auth"
	"github.com/zxh3032/two-to/backend/proto"
)

// RevokeAllSessions 处理 POST /api/v1/account/sessions/revoke-all，校验密码后退出全部设备。
func (h *Handler) RevokeAllSessions(ctx *gin.Context) {
	var req proto.CurrentPasswordRequest
	if !h.bind(ctx, "revoke_all_sessions", &req) {
		return
	}
	h.write(ctx, "revoke_all_sessions", nil, h.service.RevokeAllSessions(ctx.Request.Context(), authlib.GetGinUserID(ctx), req.GetCurrentPassword(), meta(ctx)))
}
