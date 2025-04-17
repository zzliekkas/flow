package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zzliekkas/flow"
)

// 存储器错误定义
var (
	ErrRateLimitExceeded = errors.New("请求频率超出限制")
)

// RateLimiterStore 速率限制器存储接口
type RateLimiterStore interface {
	// Increment 增加计数器并返回当前计数
	Increment(ctx context.Context, key string, expiry time.Duration) (int, error)
	// Reset 重置计数器
	Reset(ctx context.Context, key string) error
	// Get 获取当前计数
	Get(ctx context.Context, key string) (int, error)
}

// MemoryStore 内存存储实现
type MemoryStore struct {
	// 计数器映射
	counters map[string]*Counter
	// 互斥锁
	mutex sync.RWMutex
}

// Counter 计数器结构
type Counter struct {
	// 计数值
	Value int
	// 过期时间
	Expiry time.Time
}

// NewMemoryStore 创建内存存储
func NewMemoryStore() *MemoryStore {
	store := &MemoryStore{
		counters: make(map[string]*Counter),
	}

	// 启动清理过期计数器的goroutine
	go store.cleanExpired()

	return store
}

// cleanExpired 清理过期的计数器
func (s *MemoryStore) cleanExpired() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mutex.Lock()
		now := time.Now()
		for key, counter := range s.counters {
			if counter.Expiry.Before(now) {
				delete(s.counters, key)
			}
		}
		s.mutex.Unlock()
	}
}

// Increment 增加计数器
func (s *MemoryStore) Increment(ctx context.Context, key string, expiry time.Duration) (int, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	now := time.Now()
	counter, exists := s.counters[key]

	// 如果计数器不存在或已过期，创建新的计数器
	if !exists || counter.Expiry.Before(now) {
		s.counters[key] = &Counter{
			Value:  1,
			Expiry: now.Add(expiry),
		}
		return 1, nil
	}

	// 增加计数器
	counter.Value++
	return counter.Value, nil
}

// Reset 重置计数器
func (s *MemoryStore) Reset(ctx context.Context, key string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.counters, key)
	return nil
}

// Get 获取当前计数
func (s *MemoryStore) Get(ctx context.Context, key string) (int, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	counter, exists := s.counters[key]
	if !exists || counter.Expiry.Before(time.Now()) {
		return 0, nil
	}

	return counter.Value, nil
}

// RedisStore Redis存储实现
type RedisStore struct {
	// Redis客户端
	client *redis.Client
	// 键前缀
	prefix string
}

// NewRedisStore 创建Redis存储
func NewRedisStore(client *redis.Client, prefix string) *RedisStore {
	return &RedisStore{
		client: client,
		prefix: prefix,
	}
}

// Increment 增加计数器
func (s *RedisStore) Increment(ctx context.Context, key string, expiry time.Duration) (int, error) {
	key = s.prefix + key

	// 使用Redis的INCR和EXPIRE命令
	incr := s.client.Incr(ctx, key)
	if incr.Err() != nil {
		return 0, incr.Err()
	}

	value := int(incr.Val())

	// 如果是第一次设置，设置过期时间
	if value == 1 {
		expire := s.client.Expire(ctx, key, expiry)
		if expire.Err() != nil {
			return value, expire.Err()
		}
	}

	return value, nil
}

// Reset 重置计数器
func (s *RedisStore) Reset(ctx context.Context, key string) error {
	key = s.prefix + key
	return s.client.Del(ctx, key).Err()
}

// Get 获取当前计数
func (s *RedisStore) Get(ctx context.Context, key string) (int, error) {
	key = s.prefix + key
	val, err := s.client.Get(ctx, key).Int()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return val, nil
}

// RateLimiterConfig 速率限制器配置
type RateLimiterConfig struct {
	// 存储器
	Store RateLimiterStore
	// 最大请求数
	Max int
	// 时间窗口
	Duration time.Duration
	// 键生成函数
	KeyGenerator func(*flow.Context) string
	// 忽略规则
	Skipper func(*flow.Context) bool
	// 错误处理函数
	ErrorHandler func(*flow.Context, error)
	// 头部名称
	HeaderEnabled bool
	// 头部名称前缀
	HeaderPrefix string
}

// DefaultRateLimiterConfig 返回默认的速率限制器配置
func DefaultRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		Store:         NewMemoryStore(),
		Max:           100,
		Duration:      time.Hour,
		HeaderPrefix:  "X-RateLimit",
		HeaderEnabled: true,
		KeyGenerator: func(c *flow.Context) string {
			return c.ClientIP()
		},
		Skipper: func(c *flow.Context) bool {
			return false
		},
		ErrorHandler: func(c *flow.Context, err error) {
			c.String(http.StatusTooManyRequests, "请求频率超出限制")
			c.Abort()
		},
	}
}

// RateLimit 创建速率限制中间件，使用默认配置
func RateLimit() flow.HandlerFunc {
	return RateLimitWithConfig(DefaultRateLimiterConfig())
}

// RateLimitWithMax 创建具有指定最大请求数和持续时间的速率限制中间件
func RateLimitWithMax(max int, duration time.Duration) flow.HandlerFunc {
	config := DefaultRateLimiterConfig()
	config.Max = max
	config.Duration = duration
	return RateLimitWithConfig(config)
}

// RateLimitWithConfig 使用自定义配置创建速率限制中间件
func RateLimitWithConfig(config RateLimiterConfig) flow.HandlerFunc {
	// 创建默认存储
	if config.Store == nil {
		config.Store = NewMemoryStore()
	}

	// 默认最大请求数
	if config.Max <= 0 {
		config.Max = 100
	}

	// 默认持续时间
	if config.Duration <= 0 {
		config.Duration = time.Hour
	}

	// 默认键生成函数
	if config.KeyGenerator == nil {
		config.KeyGenerator = func(c *flow.Context) string {
			return c.ClientIP()
		}
	}

	// 默认跳过函数
	if config.Skipper == nil {
		config.Skipper = func(c *flow.Context) bool {
			return false
		}
	}

	// 默认错误处理函数
	if config.ErrorHandler == nil {
		config.ErrorHandler = func(c *flow.Context, err error) {
			c.String(http.StatusTooManyRequests, "请求频率超出限制")
			c.Abort()
		}
	}

	// 默认头部前缀
	if config.HeaderPrefix == "" {
		config.HeaderPrefix = "X-RateLimit"
	}

	return func(c *flow.Context) {
		// 检查是否跳过
		if config.Skipper(c) {
			c.Next()
			return
		}

		// 生成键
		key := config.KeyGenerator(c)

		// 增加计数
		current, err := config.Store.Increment(c.Request.Context(), key, config.Duration)
		if err != nil {
			config.ErrorHandler(c, err)
			return
		}

		// 设置头部
		if config.HeaderEnabled {
			c.Header(fmt.Sprintf("%s-Limit", config.HeaderPrefix), fmt.Sprintf("%d", config.Max))
			c.Header(fmt.Sprintf("%s-Remaining", config.HeaderPrefix), fmt.Sprintf("%d", config.Max-current))
			c.Header(fmt.Sprintf("%s-Reset", config.HeaderPrefix), fmt.Sprintf("%d", time.Now().Add(config.Duration).Unix()))
		}

		// 检查是否超出限制
		if current > config.Max {
			config.ErrorHandler(c, ErrRateLimitExceeded)
			return
		}

		c.Next()
	}
}

// UserRateLimit 创建用户级别的速率限制中间件
func UserRateLimit(max int, duration time.Duration, userExtractor func(*flow.Context) string) flow.HandlerFunc {
	config := DefaultRateLimiterConfig()
	config.Max = max
	config.Duration = duration
	config.KeyGenerator = func(c *flow.Context) string {
		user := userExtractor(c)
		if user == "" {
			// 如果无法提取用户，回退到IP
			return "ip:" + c.ClientIP()
		}
		return "user:" + user
	}

	return RateLimitWithConfig(config)
}

// IPRateLimit 创建IP级别的速率限制中间件
func IPRateLimit(max int, duration time.Duration) flow.HandlerFunc {
	config := DefaultRateLimiterConfig()
	config.Max = max
	config.Duration = duration
	config.KeyGenerator = func(c *flow.Context) string {
		return "ip:" + c.ClientIP()
	}

	return RateLimitWithConfig(config)
}

// RouteRateLimit 创建路由级别的速率限制中间件
func RouteRateLimit(routes map[string]int, duration time.Duration) flow.HandlerFunc {
	config := DefaultRateLimiterConfig()
	config.Duration = duration
	config.KeyGenerator = func(c *flow.Context) string {
		return "route:" + c.Request.URL.Path
	}

	return func(c *flow.Context) {
		path := c.Request.URL.Path

		// 检查路由是否配置了限制
		max, exists := routes[path]
		if !exists {
			// 如果未配置限制，则跳过
			c.Next()
			return
		}

		// 生成键
		key := config.KeyGenerator(c)

		// 增加计数
		current, err := config.Store.Increment(c.Request.Context(), key, config.Duration)
		if err != nil {
			config.ErrorHandler(c, err)
			return
		}

		// 设置头部
		if config.HeaderEnabled {
			c.Header(fmt.Sprintf("%s-Limit", config.HeaderPrefix), fmt.Sprintf("%d", max))
			c.Header(fmt.Sprintf("%s-Remaining", config.HeaderPrefix), fmt.Sprintf("%d", max-current))
			c.Header(fmt.Sprintf("%s-Reset", config.HeaderPrefix), fmt.Sprintf("%d", time.Now().Add(config.Duration).Unix()))
		}

		// 检查是否超出限制
		if current > max {
			config.ErrorHandler(c, ErrRateLimitExceeded)
			return
		}

		c.Next()
	}
}
