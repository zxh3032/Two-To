package account

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	authlib "github.com/zxh3032/two-to/backend/library/auth"
	"github.com/zxh3032/two-to/backend/library/response"
)

// RevokeSession 处理 DELETE /api/v1/account/sessions/:sessionId，踢出当前用户的指定非当前设备。
func (h *Handler) RevokeSession(ctx *gin.Context) {
	sessionID, err := strconv.ParseUint(ctx.Param("sessionId"), 10, 64)
	if err != nil || sessionID == 0 {
		if err == nil {
			err = errors.New("session id is zero")
		}
		response.WriteError(ctx, h.log, "account", "revoke_session", http.StatusBadRequest, response.CodeBadRequest, "设备 ID 不正确", err)
		return
	}
	h.write(ctx, "revoke_session", nil, h.service.RevokeSession(ctx.Request.Context(), authlib.GetGinUserID(ctx), authlib.GetGinSessionID(ctx), sessionID, meta(ctx)))
}
