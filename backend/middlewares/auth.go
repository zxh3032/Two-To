package middlewares

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	authlib "github.com/zxh3032/two-to/backend/library/auth"
	"github.com/zxh3032/two-to/backend/library/cache"
	"github.com/zxh3032/two-to/backend/library/response"
	"github.com/zxh3032/two-to/backend/models/dao"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Auth 解析 Bearer access token，并校验 session 仍处于有效状态。
func Auth(tokenManager *authlib.TokenManager, cacheStore cache.Store, db *gorm.DB, log *zap.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		raw := ctx.GetHeader("Authorization")
		if !strings.HasPrefix(raw, "Bearer ") {
			response.Error(ctx, http.StatusUnauthorized, response.CodeUnauthorized, "请先登录")
			return
		}
		claims, err := tokenManager.ParseAccessToken(strings.TrimSpace(strings.TrimPrefix(raw, "Bearer ")))
		if err != nil {
			response.Error(ctx, http.StatusUnauthorized, response.CodeUnauthorized, "登录态已失效，请重新登录")
			return
		}
		if claims.UserID == 0 || claims.SessionID == 0 {
			response.Error(ctx, http.StatusUnauthorized, response.CodeUnauthorized, "登录态已失效，请重新登录")
			return
		}
		// access token 只证明签名有效，仍需要校验 session 没有被退出、改密或过期。
		if ok := validateSession(ctx, cacheStore, db, log, claims.UserID, claims.SessionID); !ok {
			response.Error(ctx, http.StatusUnauthorized, response.CodeUnauthorized, "登录态已失效，请重新登录")
			return
		}
		// 将鉴权结果写入 gin.Context，后续 controller 可直接读取当前用户和会话。
		authlib.SetGinAuth(ctx, claims.UserID, claims.SessionID)
		ctx.Next()
	}
}

// validateSession 优先读取 Redis 中的会话快照，缓存缺失时回源 MySQL 校验。
func validateSession(ctx *gin.Context, cacheStore cache.Store, db *gorm.DB, log *zap.Logger, userID uint64, sessionID uint64) bool {
	cached, err := authlib.GetCachedSession(ctx.Request.Context(), cacheStore, sessionID)
	if err == nil {
		return cached.UserID == userID && cached.ExpireTime > time.Now().Unix()
	}
	// Redis 不可用或缓存过期时，MySQL 是最终的会话状态来源。
	if db == nil {
		log.Error("session cache 未命中且 MySQL 未配置", zap.Uint64("sessionID", sessionID), zap.Error(err))
		return false
	}
	var session dao.UserSession
	dbErr := db.WithContext(ctx.Request.Context()).
		Where("id = ? AND user_id = ? AND status = ? AND expire_time > ?", sessionID, userID, dao.SessionStatusNormal, time.Now().Unix()).
		First(&session).Error
	if dbErr != nil {
		if dbErr != gorm.ErrRecordNotFound {
			log.Error("回源查询 session 失败", zap.Uint64("sessionID", sessionID), zap.Error(dbErr))
		}
		return false
	}
	// 回源命中后回填 Redis，降低后续请求对 MySQL 的压力。
	if err := authlib.SetCachedSession(ctx.Request.Context(), cacheStore, session.ID, authlib.CachedSession{
		UserID:           session.UserID,
		RefreshTokenHash: session.RefreshTokenHash,
		ExpireTime:       session.ExpireTime,
	}); err != nil {
		log.Error("回填 session cache 失败", zap.Uint64("sessionID", sessionID), zap.Error(err))
	}
	updateLastActive(ctx, db, sessionID)
	return true
}

// updateLastActive 轻量记录设备活跃时间，失败不影响本次鉴权结果。
func updateLastActive(ctx *gin.Context, db *gorm.DB, sessionID uint64) {
	now := time.Now().Unix()
	_ = db.WithContext(ctx.Request.Context()).Model(&dao.UserSession{}).
		Where("id = ?", sessionID).
		Updates(map[string]interface{}{"last_active_time": now, "update_time": now}).Error
}
