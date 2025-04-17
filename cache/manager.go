package cache

import (
	"context"
	"errors"
	"sync"
	"time"
)

// Manager 缓存管理器
type Manager struct {
	stores   map[string]Store  // 存储的缓存实例
	configs  map[string]Config // 缓存配置
	mutex    sync.RWMutex      // 并发锁
	default_ string            // 默认存储
}

// Config 缓存配置
type Config struct {
	Driver string                 // 驱动类型
	Prefix string                 // 键前缀
	TTL    time.Duration          // 默认过期时间
	Config map[string]interface{} // 驱动特定配置
}

// NewManager 创建缓存管理器
func NewManager() *Manager {
	return &Manager{
		stores:   make(map[string]Store),
		configs:  make(map[string]Config),
		default_: "memory",
	}
}

// SetDefault 设置默认存储
func (m *Manager) SetDefault(name string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.stores[name]; exists {
		m.default_ = name
	}
}

// Register 注册缓存配置
func (m *Manager) Register(name string, config Config) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 保存配置
	m.configs[name] = config

	// 如果需要即时创建存储，可以在这里调用 m.createStore(name, config)
	return nil
}

// GetStore 获取指定名称的缓存存储
func (m *Manager) GetStore(name string) (Store, error) {
	m.mutex.RLock()

	// 检查存储是否已存在
	if store, exists := m.stores[name]; exists {
		m.mutex.RUnlock()
		return store, nil
	}

	// 获取配置
	config, exists := m.configs[name]
	if !exists {
		m.mutex.RUnlock()
		return nil, errors.New("缓存配置不存在: " + name)
	}

	m.mutex.RUnlock()

	// 创建存储
	return m.createStore(name, config)
}

// Store 获取默认缓存存储
func (m *Manager) Store() (Store, error) {
	return m.GetStore(m.default_)
}

// 创建缓存存储
func (m *Manager) createStore(name string, config Config) (Store, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 如果在加锁期间已经有其他协程创建了，直接返回
	if store, exists := m.stores[name]; exists {
		return store, nil
	}

	// 获取驱动
	driver, exists := GetDriver(config.Driver)
	if !exists {
		return nil, errors.New("缓存驱动不存在: " + config.Driver)
	}

	// 创建存储
	store, err := driver.New(config.Config)
	if err != nil {
		return nil, err
	}

	// 保存到缓存
	m.stores[name] = store

	return store, nil
}

// 以下方法是对默认存储的操作的便捷封装

// Get 从默认存储获取缓存
func (m *Manager) Get(ctx context.Context, key string) (interface{}, error) {
	store, err := m.Store()
	if err != nil {
		return nil, err
	}
	return store.Get(ctx, key)
}

// Set 向默认存储设置缓存
func (m *Manager) Set(ctx context.Context, key string, value interface{}, opts ...Option) error {
	store, err := m.Store()
	if err != nil {
		return err
	}
	return store.Set(ctx, key, value, opts...)
}

// Delete 从默认存储删除缓存
func (m *Manager) Delete(ctx context.Context, key string) error {
	store, err := m.Store()
	if err != nil {
		return err
	}
	return store.Delete(ctx, key)
}

// Has 检查默认存储中是否存在缓存
func (m *Manager) Has(ctx context.Context, key string) bool {
	store, err := m.Store()
	if err != nil {
		return false
	}
	return store.Has(ctx, key)
}

// Clear 清空默认存储
func (m *Manager) Clear(ctx context.Context) error {
	store, err := m.Store()
	if err != nil {
		return err
	}
	return store.Clear(ctx)
}

// Increment 增加计数器值
func (m *Manager) Increment(ctx context.Context, key string, value int64) (int64, error) {
	store, err := m.Store()
	if err != nil {
		return 0, err
	}
	return store.Increment(ctx, key, value)
}

// Decrement 减少计数器值
func (m *Manager) Decrement(ctx context.Context, key string, value int64) (int64, error) {
	store, err := m.Store()
	if err != nil {
		return 0, err
	}
	return store.Decrement(ctx, key, value)
}

// GetMultiple 获取多个缓存项
func (m *Manager) GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error) {
	store, err := m.Store()
	if err != nil {
		return nil, err
	}
	return store.GetMultiple(ctx, keys)
}

// SetMultiple 设置多个缓存项
func (m *Manager) SetMultiple(ctx context.Context, items map[string]interface{}, opts ...Option) error {
	store, err := m.Store()
	if err != nil {
		return err
	}
	return store.SetMultiple(ctx, items, opts...)
}

// DeleteMultiple 删除多个缓存项
func (m *Manager) DeleteMultiple(ctx context.Context, keys []string) error {
	store, err := m.Store()
	if err != nil {
		return err
	}
	return store.DeleteMultiple(ctx, keys)
}

// TaggedGet 获取带有标签的缓存项
func (m *Manager) TaggedGet(ctx context.Context, tag string) (map[string]interface{}, error) {
	store, err := m.Store()
	if err != nil {
		return nil, err
	}
	return store.TaggedGet(ctx, tag)
}

// TaggedDelete 删除带有标签的缓存项
func (m *Manager) TaggedDelete(ctx context.Context, tag string) error {
	store, err := m.Store()
	if err != nil {
		return err
	}
	return store.TaggedDelete(ctx, tag)
}

// WithPrefix 创建带有前缀的缓存管理器
func (m *Manager) WithPrefix(prefix string) *PrefixedManager {
	return &PrefixedManager{
		manager: m,
		prefix:  prefix,
	}
}

// PrefixedManager 带前缀的缓存管理器
type PrefixedManager struct {
	manager *Manager
	prefix  string
}

// 生成带前缀的键
func (p *PrefixedManager) prefixKey(key string) string {
	return p.prefix + ":" + key
}

// 从带前缀的键数组中提取原始键和带前缀的键的映射
func (p *PrefixedManager) prefixKeys(keys []string) ([]string, map[string]string) {
	prefixedKeys := make([]string, len(keys))
	mapping := make(map[string]string, len(keys))

	for i, key := range keys {
		prefixedKey := p.prefixKey(key)
		prefixedKeys[i] = prefixedKey
		mapping[prefixedKey] = key
	}

	return prefixedKeys, mapping
}

// Get 获取缓存
func (p *PrefixedManager) Get(ctx context.Context, key string) (interface{}, error) {
	return p.manager.Get(ctx, p.prefixKey(key))
}

// Set 设置缓存
func (p *PrefixedManager) Set(ctx context.Context, key string, value interface{}, opts ...Option) error {
	return p.manager.Set(ctx, p.prefixKey(key), value, opts...)
}

// Delete 删除缓存
func (p *PrefixedManager) Delete(ctx context.Context, key string) error {
	return p.manager.Delete(ctx, p.prefixKey(key))
}

// Has 检查缓存是否存在
func (p *PrefixedManager) Has(ctx context.Context, key string) bool {
	return p.manager.Has(ctx, p.prefixKey(key))
}

// GetMultiple 获取多个缓存项
func (p *PrefixedManager) GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error) {
	prefixedKeys, mapping := p.prefixKeys(keys)

	prefixedResult, err := p.manager.GetMultiple(ctx, prefixedKeys)
	if err != nil {
		return nil, err
	}

	// 转换回原始键
	result := make(map[string]interface{}, len(prefixedResult))
	for prefixedKey, value := range prefixedResult {
		if originalKey, exists := mapping[prefixedKey]; exists {
			result[originalKey] = value
		}
	}

	return result, nil
}

// SetMultiple 设置多个缓存项
func (p *PrefixedManager) SetMultiple(ctx context.Context, items map[string]interface{}, opts ...Option) error {
	prefixedItems := make(map[string]interface{}, len(items))

	for key, value := range items {
		prefixedItems[p.prefixKey(key)] = value
	}

	return p.manager.SetMultiple(ctx, prefixedItems, opts...)
}

// DeleteMultiple 删除多个缓存项
func (p *PrefixedManager) DeleteMultiple(ctx context.Context, keys []string) error {
	prefixedKeys, _ := p.prefixKeys(keys)
	return p.manager.DeleteMultiple(ctx, prefixedKeys)
}
