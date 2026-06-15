package auth

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/zxh3032/two-to/backend/library/cache"
)

// CachedSession 是 Redis/session cache 中保存的最小会话状态。
type CachedSession struct {
	UserID           uint64 `json:"userId"`
	RefreshTokenHash string `json:"refreshTokenHash"`
	ExpireTime       int64  `json:"expireTime"`
}

// SessionCacheKey 生成 session cache key，统一 Redis 和内存 cache 的命名。
func SessionCacheKey(store cache.Store, sessionID uint64) string {
	return store.Key("session", strconv.FormatUint(sessionID, 10))
}

// SetCachedSession 写入 session 快照，TTL 跟随会话过期时间。
func SetCachedSession(ctx context.Context, store cache.Store, sessionID uint64, session CachedSession) error {
	raw, err := json.Marshal(session)
	if err != nil {
		return err
	}
	ttl := time.Until(time.Unix(session.ExpireTime, 0))
	if ttl <= 0 {
		// 已过期会话仍给一个极短 TTL，避免底层 cache 拒绝非正数过期时间。
		ttl = time.Second
	}
	return store.Set(ctx, SessionCacheKey(store, sessionID), string(raw), ttl)
}

// GetCachedSession 读取 session 快照，调用方负责在未命中时回源 MySQL。
func GetCachedSession(ctx context.Context, store cache.Store, sessionID uint64) (*CachedSession, error) {
	raw, err := store.Get(ctx, SessionCacheKey(store, sessionID))
	if err != nil {
		return nil, err
	}
	var session CachedSession
	if err := json.Unmarshal([]byte(raw), &session); err != nil {
		return nil, err
	}
	return &session, nil
}

// DeleteCachedSession 删除 session 快照，退出登录和吊销设备时需要同步调用。
func DeleteCachedSession(ctx context.Context, store cache.Store, sessionID uint64) error {
	return store.Del(ctx, SessionCacheKey(store, sessionID))
}
