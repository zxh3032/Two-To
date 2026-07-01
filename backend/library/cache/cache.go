package cache

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zxh3032/two-to/backend/library/config"
	"go.uber.org/zap"
)

// ErrMiss 表示 cache 中不存在对应 key，业务层可以按需回源 MySQL。
var ErrMiss = errors.New("cache miss")

// Store 抽象 Redis 和本地内存 cache，避免验证码和 session 逻辑依赖具体实现。
type Store interface {
	Key(parts ...string) string
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	SetNX(ctx context.Context, key string, value string, ttl time.Duration) (bool, error)
	Del(ctx context.Context, keys ...string) error
	IncrWithTTL(ctx context.Context, key string, ttl time.Duration) (int64, error)
	TTL(ctx context.Context, key string) (time.Duration, error)
	Close() error
}

// Open 按配置初始化 cache。生产默认使用 Redis，本地可以显式使用 memory。
func Open(ctx context.Context, cfg config.CacheConfig, log *zap.Logger) (Store, error) {
	switch strings.ToLower(cfg.Driver) {
	case "redis":
		client := redis.NewClient(&redis.Options{
			Addr:     cfg.RedisAddr,
			Username: cfg.Username,
			Password: cfg.Password,
			DB:       cfg.DB,
		})
		if err := client.Ping(ctx).Err(); err != nil {
			return nil, err
		}
		log.Info("Redis cache 初始化完成", zap.String("addr", cfg.RedisAddr), zap.Int("db", cfg.DB))
		return &RedisStore{client: client, prefix: strings.Trim(cfg.KeyPrefix, ":")}, nil
	case "memory", "":
		log.Warn("当前使用内存 cache，仅适合本地开发或测试")
		return NewMemoryStore(cfg.KeyPrefix), nil
	default:
		return nil, fmt.Errorf("不支持的 cache driver: %s", cfg.Driver)
	}
}

// RedisStore 是 go-redis 的轻量封装。
type RedisStore struct {
	client *redis.Client
	prefix string
}

// Key 生成带统一前缀的 Redis key，避免不同环境或业务 key 冲突。
func (s *RedisStore) Key(parts ...string) string {
	return joinKey(s.prefix, parts...)
}

// Get 读取 Redis 字符串值，key 不存在时统一返回 ErrMiss。
func (s *RedisStore) Get(ctx context.Context, key string) (string, error) {
	value, err := s.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrMiss
	}
	return value, err
}

// Set 写入带 TTL 的 Redis 字符串值，验证码和会话都必须设置过期时间。
func (s *RedisStore) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return s.client.Set(ctx, key, value, ttl).Err()
}

// SetNX 在 key 不存在时写入值，适合实现短期锁和节流标记。
func (s *RedisStore) SetNX(ctx context.Context, key string, value string, ttl time.Duration) (bool, error) {
	return s.client.SetNX(ctx, key, value, ttl).Result()
}

// Del 删除一个或多个 Redis key，调用方可安全传入空列表。
func (s *RedisStore) Del(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return s.client.Del(ctx, keys...).Err()
}

// IncrWithTTL 原子递增计数并刷新 TTL，用于验证码频控和登录失败计数。
func (s *RedisStore) IncrWithTTL(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	pipe := s.client.TxPipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, ttl)
	if _, err := pipe.Exec(ctx); err != nil {
		return 0, err
	}
	return incr.Val(), nil
}

// TTL 查询 key 剩余有效期，key 不存在时返回 ErrMiss。
func (s *RedisStore) TTL(ctx context.Context, key string) (time.Duration, error) {
	duration, err := s.client.TTL(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return 0, ErrMiss
	}
	return duration, err
}

// Close 关闭 Redis 客户端连接。
func (s *RedisStore) Close() error {
	return s.client.Close()
}

// MemoryStore 是本地开发兜底实现，语义尽量贴近 Redis 的 TTL 行为。
type MemoryStore struct {
	mu     sync.Mutex
	prefix string
	items  map[string]memoryItem
}

type memoryItem struct {
	value      string
	expireTime time.Time
}

// NewMemoryStore 创建本地内存 cache，主要用于不依赖 Redis 的本地开发和单测。
func NewMemoryStore(prefix string) *MemoryStore {
	return &MemoryStore{
		prefix: strings.Trim(prefix, ":"),
		items:  make(map[string]memoryItem),
	}
}

// Key 生成带统一前缀的内存 cache key。
func (s *MemoryStore) Key(parts ...string) string {
	return joinKey(s.prefix, parts...)
}

// Get 读取内存 cache，过期或不存在时统一返回 ErrMiss。
func (s *MemoryStore) Get(_ context.Context, key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.items[key]
	if !ok || item.expired(time.Now()) {
		delete(s.items, key)
		return "", ErrMiss
	}
	return item.value, nil
}

// Set 写入内存 cache，并记录过期时间。
func (s *MemoryStore) Set(_ context.Context, key string, value string, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[key] = memoryItem{value: value, expireTime: time.Now().Add(ttl)}
	return nil
}

// SetNX 在内存 cache 中模拟 Redis SET NX 语义。
func (s *MemoryStore) SetNX(_ context.Context, key string, value string, ttl time.Duration) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	item, ok := s.items[key]
	if ok && !item.expired(now) {
		return false, nil
	}
	s.items[key] = memoryItem{value: value, expireTime: now.Add(ttl)}
	return true, nil
}

// Del 删除内存 cache key，支持批量删除。
func (s *MemoryStore) Del(_ context.Context, keys ...string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, key := range keys {
		delete(s.items, key)
	}
	return nil
}

// IncrWithTTL 递增内存计数并刷新 TTL，模拟 Redis 计数器语义。
func (s *MemoryStore) IncrWithTTL(_ context.Context, key string, ttl time.Duration) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	item, ok := s.items[key]
	if !ok || item.expired(now) {
		s.items[key] = memoryItem{value: "1", expireTime: now.Add(ttl)}
		return 1, nil
	}
	current, _ := strconv.ParseInt(item.value, 10, 64)
	current++
	item.value = strconv.FormatInt(current, 10)
	item.expireTime = now.Add(ttl)
	s.items[key] = item
	return current, nil
}

// TTL 查询内存 key 剩余有效期。
func (s *MemoryStore) TTL(_ context.Context, key string) (time.Duration, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.items[key]
	if !ok || item.expired(time.Now()) {
		delete(s.items, key)
		return 0, ErrMiss
	}
	return time.Until(item.expireTime), nil
}

// Close 清空内存 cache，满足 Store 接口的资源释放语义。
func (s *MemoryStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	clear(s.items)
	return nil
}

// expired 判断内存项是否已经超过过期时间。
func (i memoryItem) expired(now time.Time) bool {
	return !i.expireTime.IsZero() && now.After(i.expireTime)
}

// joinKey 统一拼接 cache key，并清理调用方传入的多余冒号。
func joinKey(prefix string, parts ...string) string {
	clean := make([]string, 0, len(parts)+1)
	if prefix != "" {
		clean = append(clean, prefix)
	}
	for _, part := range parts {
		if p := strings.Trim(part, ":"); p != "" {
			clean = append(clean, p)
		}
	}
	return strings.Join(clean, ":")
}
