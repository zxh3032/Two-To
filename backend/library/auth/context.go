package auth

import "github.com/gin-gonic/gin"

const (
	ginUserIDKey    = "authUserID"
	ginSessionIDKey = "authSessionID"
)

// SetGinAuth 将鉴权中间件解析出的用户 ID 和会话 ID 写入请求上下文。
func SetGinAuth(ctx *gin.Context, userID uint64, sessionID uint64) {
	ctx.Set(ginUserIDKey, userID)
	ctx.Set(ginSessionIDKey, sessionID)
}

// GetGinUserID 从请求上下文读取当前用户 ID，未登录或类型异常时返回 0。
func GetGinUserID(ctx *gin.Context) uint64 {
	value, ok := ctx.Get(ginUserIDKey)
	if !ok {
		return 0
	}
	userID, _ := value.(uint64)
	return userID
}

// GetGinSessionID 从请求上下文读取当前会话 ID，供设备管理和退出登录使用。
func GetGinSessionID(ctx *gin.Context) uint64 {
	value, ok := ctx.Get(ginSessionIDKey)
	if !ok {
		return 0
	}
	sessionID, _ := value.(uint64)
	return sessionID
}
