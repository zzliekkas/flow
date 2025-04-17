package cache

import (
	"context"
	"errors"
	"sync"
	"time"
)

// 定义错误常量
var (
	ErrCacheMiss  = errors.New("缓存不存在")
	ErrInvalidKey = errors.New("无效的缓存键")
)

// Item 缓存项结构
type Item struct {
	Key        string        // 缓存键
	Value      interface{}   // 缓存值
	Tags       []string      // 关联标签
	Expiration time.Duration // 过期时间
	CreatedAt  time.Time     // 创建时间
}

// Store 缓存存储接口
type Store interface {
	// 基本操作
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}, options ...Option) error
	Delete(ctx context.Context, key string) error
	Has(ctx context.Context, key string) bool
	Clear(ctx context.Context) error

	// 批量操作
	GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error)
	SetMultiple(ctx context.Context, items map[string]interface{}, options ...Option) error
	DeleteMultiple(ctx context.Context, keys []string) error

	// 计数器操作
	Increment(ctx context.Context, key string, value int64) (int64, error)
	Decrement(ctx context.Context, key string, value int64) (int64, error)

	// 标签操作
	TaggedGet(ctx context.Context, tag string) (map[string]interface{}, error)
	TaggedDelete(ctx context.Context, tag string) error

	// 统计操作
	Count(ctx context.Context) int64

	// 维护操作
	Flush(ctx context.Context) error
}

// Options 缓存选项
type Options struct {
	Expiration time.Duration // 过期时间
	Tags       []string      // 标签
}

// Option 缓存配置函数
type Option func(*Options)

// WithExpiration 设置过期时间
func WithExpiration(expiration time.Duration) Option {
	return func(o *Options) {
		o.Expiration = expiration
	}
}

// WithTags 设置标签
func WithTags(tags ...string) Option {
	return func(o *Options) {
		o.Tags = tags
	}
}

// 应用配置选项
func applyOptions(options ...Option) Options {
	opts := Options{}
	for _, option := range options {
		option(&opts)
	}
	return opts
}

// Driver 缓存驱动接口
type Driver interface {
	New(config map[string]interface{}) (Store, error)
}

// 缓存驱动管理
var (
	drivers = make(map[string]Driver)
	mu      sync.RWMutex
)

// RegisterDriver 注册缓存驱动
func RegisterDriver(name string, driver Driver) {
	mu.Lock()
	defer mu.Unlock()
	drivers[name] = driver
}

// GetDriver 获取缓存驱动
func GetDriver(name string) (Driver, bool) {
	mu.RLock()
	defer mu.RUnlock()
	driver, ok := drivers[name]
	return driver, ok
}

// New 创建新的缓存实例
func New(driver string, config map[string]interface{}) (Store, error) {
	d, ok := GetDriver(driver)
	if !ok {
		return nil, errors.New("未知的缓存驱动: " + driver)
	}
	return d.New(config)
}
