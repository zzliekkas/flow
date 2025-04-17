package cache

import (
	"context"
	"errors"
	"sync"
	"time"
)

// 错误定义
var (
	ErrInvalidValue = errors.New("无效的缓存值类型")
)

// MemoryDriver 内存缓存驱动
type MemoryDriver struct{}

// New 创建新的内存缓存实例
func (d *MemoryDriver) New(config map[string]interface{}) (Store, error) {
	store := NewMemoryStore()
	return store, nil
}

func init() {
	RegisterDriver("memory", &MemoryDriver{})
}

// MemoryStore 内存缓存存储实现
type MemoryStore struct {
	items      map[string]Item
	mutex      sync.RWMutex
	tagManager TagManager
}

// NewMemoryStore 创建新的内存缓存存储
func NewMemoryStore() *MemoryStore {
	store := &MemoryStore{
		items: make(map[string]Item),
	}
	store.tagManager = NewTagManager(store)
	return store
}

// Get 获取缓存项
func (s *MemoryStore) Get(ctx context.Context, key string) (interface{}, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	item, found := s.items[key]
	if !found {
		return nil, ErrCacheMiss
	}

	// 检查过期时间
	if !item.CreatedAt.IsZero() && item.Expiration > 0 &&
		time.Now().After(item.CreatedAt.Add(item.Expiration)) {
		return nil, ErrCacheMiss
	}

	return item.Value, nil
}

// GetMultiple 获取多个缓存项
func (s *MemoryStore) GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, key := range keys {
		item, found := s.items[key]
		if !found {
			continue
		}

		// 检查过期时间
		if !item.CreatedAt.IsZero() && item.Expiration > 0 &&
			time.Now().After(item.CreatedAt.Add(item.Expiration)) {
			continue
		}

		result[key] = item.Value
	}

	return result, nil
}

// Set 设置缓存项
func (s *MemoryStore) Set(ctx context.Context, key string, value interface{}, opts ...Option) error {
	options := &Options{}
	for _, opt := range opts {
		opt(options)
	}

	item := Item{
		Key:        key,
		Value:      value,
		Expiration: options.Expiration,
		CreatedAt:  time.Now(),
		Tags:       options.Tags,
	}

	s.mutex.Lock()
	s.items[key] = item
	s.mutex.Unlock()

	// 处理标签
	if len(options.Tags) > 0 {
		err := s.tagManager.AddTagsToKey(ctx, key, options.Tags)
		if err != nil {
			return err
		}
	}

	return nil
}

// SetMultiple 设置多个缓存项
func (s *MemoryStore) SetMultiple(ctx context.Context, items map[string]interface{}, opts ...Option) error {
	options := &Options{}
	for _, opt := range opts {
		opt(options)
	}

	s.mutex.Lock()
	for key, value := range items {
		item := Item{
			Key:        key,
			Value:      value,
			Expiration: options.Expiration,
			CreatedAt:  time.Now(),
			Tags:       options.Tags,
		}
		s.items[key] = item
	}
	s.mutex.Unlock()

	// 处理标签
	if len(options.Tags) > 0 {
		for key := range items {
			err := s.tagManager.AddTagsToKey(ctx, key, options.Tags)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Delete 删除缓存项
func (s *MemoryStore) Delete(ctx context.Context, key string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, exists := s.items[key]
	if !exists {
		return nil
	}

	delete(s.items, key)

	// 从所有标签中移除该键
	return s.tagManager.RemoveKeyFromAllTags(ctx, key)
}

// DeleteMultiple 删除多个缓存项
func (s *MemoryStore) DeleteMultiple(ctx context.Context, keys []string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, key := range keys {
		delete(s.items, key)

		// 从所有标签中移除该键
		err := s.tagManager.RemoveKeyFromAllTags(ctx, key)
		if err != nil {
			return err
		}
	}

	return nil
}

// Clear 清空所有缓存
func (s *MemoryStore) Clear(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.items = make(map[string]Item)
	// 重置标签管理器
	s.tagManager = NewTagManager(s)

	return nil
}

// Has 检查缓存项是否存在
func (s *MemoryStore) Has(ctx context.Context, key string) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	item, exists := s.items[key]
	if !exists {
		return false
	}

	// 检查是否已过期
	if !item.CreatedAt.IsZero() && item.Expiration > 0 &&
		time.Now().After(item.CreatedAt.Add(item.Expiration)) {
		return false
	}

	return true
}

// Increment 增加计数器值
func (s *MemoryStore) Increment(ctx context.Context, key string, value int64) (int64, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 获取当前值
	item, exists := s.items[key]
	var current int64
	if exists {
		// 检查是否已过期
		if !item.CreatedAt.IsZero() && item.Expiration > 0 &&
			time.Now().After(item.CreatedAt.Add(item.Expiration)) {
			exists = false
		} else {
			// 尝试转换为 int64
			switch v := item.Value.(type) {
			case int64:
				current = v
			case int:
				current = int64(v)
			case int32:
				current = int64(v)
			case float64:
				current = int64(v)
			case float32:
				current = int64(v)
			default:
				return 0, ErrInvalidValue
			}
		}
	}

	// 计算新值
	newValue := current + value

	// 如果不存在，创建新项；如果存在，更新现有项
	expiration := time.Duration(0)
	tags := []string{}
	if exists {
		expiration = item.Expiration
		tags = item.Tags
	}

	// 更新缓存
	s.items[key] = Item{
		Key:        key,
		Value:      newValue,
		Expiration: expiration,
		CreatedAt:  time.Now(),
		Tags:       tags,
	}

	return newValue, nil
}

// Decrement 减少计数器值
func (s *MemoryStore) Decrement(ctx context.Context, key string, value int64) (int64, error) {
	return s.Increment(ctx, key, -value)
}

// TaggedGet 获取带有指定标签的所有项
func (s *MemoryStore) TaggedGet(ctx context.Context, tag string) (map[string]interface{}, error) {
	keys, err := s.tagManager.GetKeysByTag(ctx, tag)
	if err != nil {
		return nil, err
	}

	if len(keys) == 0 {
		return map[string]interface{}{}, nil
	}

	return s.GetMultiple(ctx, keys)
}

// TaggedDelete 删除带有指定标签的所有项
func (s *MemoryStore) TaggedDelete(ctx context.Context, tag string) error {
	keys, err := s.tagManager.GetKeysByTag(ctx, tag)
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		return nil
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 删除所有带有该标签的项
	for _, key := range keys {
		delete(s.items, key)
	}

	// 移除标签
	return s.tagManager.RemoveTag(ctx, tag)
}

// Count 返回缓存项数量
func (s *MemoryStore) Count(ctx context.Context) int64 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var count int64
	now := time.Now()

	for _, item := range s.items {
		// 只计算未过期的项
		if !item.CreatedAt.IsZero() && item.Expiration > 0 &&
			now.After(item.CreatedAt.Add(item.Expiration)) {
			continue
		}
		count++
	}

	return count
}

// Flush 清理并停止服务
func (s *MemoryStore) Flush(ctx context.Context) error {
	return s.Clear(ctx)
}

// GC 垃圾回收，清理过期的缓存项
func (s *MemoryStore) GC(ctx context.Context) error {
	now := time.Now()
	expiredKeys := make([]string, 0)

	s.mutex.RLock()
	for key, item := range s.items {
		if !item.CreatedAt.IsZero() && item.Expiration > 0 &&
			now.After(item.CreatedAt.Add(item.Expiration)) {
			expiredKeys = append(expiredKeys, key)
		}
	}
	s.mutex.RUnlock()

	if len(expiredKeys) > 0 {
		return s.DeleteMultiple(ctx, expiredKeys)
	}

	return nil
}
