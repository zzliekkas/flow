package cache

import (
	"context"
	"sync"
	"time"
)

// TagManager 标签管理器接口
type TagManager interface {
	// AddTagsToKey 为缓存键添加标签
	AddTagsToKey(ctx context.Context, key string, tags []string) error

	// RemoveTagsFromKey 从缓存键中移除标签
	RemoveTagsFromKey(ctx context.Context, key string, tags []string) error

	// GetKeysByTag 根据标签获取所有关联的键
	GetKeysByTag(ctx context.Context, tag string) ([]string, error)

	// RemoveTag 移除标签及其所有关联
	RemoveTag(ctx context.Context, tag string) error

	// RemoveKeyFromAllTags 从所有标签中移除指定的键
	RemoveKeyFromAllTags(ctx context.Context, key string) error

	// GetKeyTags 获取键关联的所有标签
	GetKeyTags(ctx context.Context, key string) ([]string, error)
}

// StandardTagManager 标准标签管理器实现
type StandardTagManager struct {
	// 标签到键的映射
	tagToKeys map[string]map[string]struct{}
	// 键到标签的映射
	keyToTags map[string]map[string]struct{}
	// 用于同步的锁
	mutex sync.RWMutex
	// 标签到过期时间的映射
	expirations map[string]time.Time
	// 标签处理接口
	store Store
}

// NewTagManager 创建新的标签管理器
func NewTagManager(store Store) TagManager {
	return &StandardTagManager{
		tagToKeys:   make(map[string]map[string]struct{}),
		keyToTags:   make(map[string]map[string]struct{}),
		expirations: make(map[string]time.Time),
		store:       store,
	}
}

// AddTagsToKey 为缓存键添加标签
func (m *StandardTagManager) AddTagsToKey(ctx context.Context, key string, tags []string) error {
	if len(tags) == 0 {
		return nil
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 为每个标签添加键关联
	for _, tag := range tags {
		if _, exists := m.tagToKeys[tag]; !exists {
			m.tagToKeys[tag] = make(map[string]struct{})
		}
		m.tagToKeys[tag][key] = struct{}{}
	}

	// 为键添加标签关联
	if _, exists := m.keyToTags[key]; !exists {
		m.keyToTags[key] = make(map[string]struct{})
	}
	for _, tag := range tags {
		m.keyToTags[key][tag] = struct{}{}
	}

	return nil
}

// RemoveTagsFromKey 从缓存键中移除标签
func (m *StandardTagManager) RemoveTagsFromKey(ctx context.Context, key string, tags []string) error {
	if len(tags) == 0 {
		return nil
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.keyToTags[key]; !exists {
		return nil
	}

	// 从键的标签列表中移除标签
	for _, tag := range tags {
		delete(m.keyToTags[key], tag)

		// 从标签的键列表中移除键
		if keyMap, exists := m.tagToKeys[tag]; exists {
			delete(keyMap, key)

			// 如果标签没有关联的键，则删除该标签
			if len(keyMap) == 0 {
				delete(m.tagToKeys, tag)
				delete(m.expirations, tag)
			}
		}
	}

	// 如果键没有关联的标签，则从映射中删除
	if len(m.keyToTags[key]) == 0 {
		delete(m.keyToTags, key)
	}

	return nil
}

// GetKeysByTag 根据标签获取所有关联的键
func (m *StandardTagManager) GetKeysByTag(ctx context.Context, tag string) ([]string, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	keyMap, exists := m.tagToKeys[tag]
	if !exists {
		return []string{}, nil
	}

	keys := make([]string, 0, len(keyMap))
	for key := range keyMap {
		keys = append(keys, key)
	}

	return keys, nil
}

// RemoveTag 移除标签及其所有关联
func (m *StandardTagManager) RemoveTag(ctx context.Context, tag string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 获取与标签关联的所有键
	keyMap, exists := m.tagToKeys[tag]
	if !exists {
		return nil
	}

	// 从每个键的标签列表中移除该标签
	for key := range keyMap {
		if tagMap, ok := m.keyToTags[key]; ok {
			delete(tagMap, tag)

			// 如果键没有关联的标签，则删除该键
			if len(tagMap) == 0 {
				delete(m.keyToTags, key)
			}
		}
	}

	// 删除标签
	delete(m.tagToKeys, tag)
	delete(m.expirations, tag)

	return nil
}

// RemoveKeyFromAllTags 从所有标签中移除指定的键
func (m *StandardTagManager) RemoveKeyFromAllTags(ctx context.Context, key string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 获取与键关联的所有标签
	tagMap, exists := m.keyToTags[key]
	if !exists {
		return nil
	}

	// 从每个标签的键列表中移除该键
	for tag := range tagMap {
		if keyMap, ok := m.tagToKeys[tag]; ok {
			delete(keyMap, key)

			// 如果标签没有关联的键，则删除该标签
			if len(keyMap) == 0 {
				delete(m.tagToKeys, tag)
				delete(m.expirations, tag)
			}
		}
	}

	// 删除键
	delete(m.keyToTags, key)

	return nil
}

// GetKeyTags 获取键关联的所有标签
func (m *StandardTagManager) GetKeyTags(ctx context.Context, key string) ([]string, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	tagMap, exists := m.keyToTags[key]
	if !exists {
		return []string{}, nil
	}

	tags := make([]string, 0, len(tagMap))
	for tag := range tagMap {
		tags = append(tags, tag)
	}

	return tags, nil
}
