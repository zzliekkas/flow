package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// 连接状态常量
const (
	ConnStatusUnknown = iota
	ConnStatusConnected
	ConnStatusDisconnected
	ConnStatusError
)

// RedisStore 实现了Store接口，使用Redis作为存储后端
type RedisStore struct {
	client        *redis.Client
	prefix        string
	defaultExpiry time.Duration
	healthStatus  int
	healthMutex   sync.RWMutex
	healthTicker  *time.Ticker
	stopChan      chan struct{}
	tagManager    TagManager
}

// RedisOptions 用于配置Redis缓存
type RedisOptions struct {
	Prefix              string
	DefaultExpiry       time.Duration
	TagManager          TagManager
	HealthCheck         bool
	HealthCheckInterval time.Duration
	MaxRetries          int
	PoolSize            int
	MinIdleConns        int
}

// WithRedisPrefix 设置缓存键前缀
func WithRedisPrefix(prefix string) func(*RedisOptions) {
	return func(o *RedisOptions) {
		o.Prefix = prefix
	}
}

// WithRedisExpiry 设置默认过期时间
func WithRedisExpiry(expiry time.Duration) func(*RedisOptions) {
	return func(o *RedisOptions) {
		o.DefaultExpiry = expiry
	}
}

// WithRedisTagManager 设置标签管理器
func WithRedisTagManager(tagManager TagManager) func(*RedisOptions) {
	return func(o *RedisOptions) {
		o.TagManager = tagManager
	}
}

// WithRedisHealthCheck 设置健康检查选项
func WithRedisHealthCheck(enabled bool, interval time.Duration) func(*RedisOptions) {
	return func(o *RedisOptions) {
		o.HealthCheck = enabled
		if interval > 0 {
			o.HealthCheckInterval = interval
		}
	}
}

// WithRedisPool 设置连接池选项
func WithRedisPool(maxRetries, poolSize, minIdleConns int) func(*RedisOptions) {
	return func(o *RedisOptions) {
		o.MaxRetries = maxRetries
		o.PoolSize = poolSize
		o.MinIdleConns = minIdleConns
	}
}

// NewRedisStore 创建一个新的Redis缓存存储
func NewRedisStore(client *redis.Client, opts ...func(*RedisOptions)) *RedisStore {
	options := &RedisOptions{
		Prefix:              "flow:", // 默认前缀
		DefaultExpiry:       5 * time.Minute,
		HealthCheck:         true,
		HealthCheckInterval: 30 * time.Second,
		MaxRetries:          3,
		PoolSize:            10,
		MinIdleConns:        2,
	}

	// 应用选项
	for _, opt := range opts {
		opt(options)
	}

	store := &RedisStore{
		client:        client,
		prefix:        options.Prefix,
		defaultExpiry: options.DefaultExpiry,
		healthStatus:  ConnStatusUnknown,
		stopChan:      make(chan struct{}),
	}

	// 初始化标签管理器
	if options.TagManager != nil {
		store.tagManager = options.TagManager
	} else {
		store.tagManager = NewRedisTagManager(client, options.Prefix)
	}

	// 启动健康检查
	if options.HealthCheck {
		store.healthTicker = time.NewTicker(options.HealthCheckInterval)
		go store.runHealthCheck()
	}

	return store
}

// runHealthCheck 运行定期健康检查
func (r *RedisStore) runHealthCheck() {
	for {
		select {
		case <-r.healthTicker.C:
			r.checkHealth()
		case <-r.stopChan:
			r.healthTicker.Stop()
			return
		}
	}
}

// checkHealth 检查Redis连接健康状态
func (r *RedisStore) checkHealth() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := r.client.Ping(ctx).Err()

	r.healthMutex.Lock()
	defer r.healthMutex.Unlock()

	if err != nil {
		r.healthStatus = ConnStatusError
	} else {
		r.healthStatus = ConnStatusConnected
	}
}

// GetHealthStatus 获取Redis连接健康状态
func (r *RedisStore) GetHealthStatus() int {
	r.healthMutex.RLock()
	defer r.healthMutex.RUnlock()
	return r.healthStatus
}

// Close 关闭Redis存储及相关资源
func (r *RedisStore) Close() error {
	// 停止健康检查
	if r.healthTicker != nil {
		close(r.stopChan)
	}

	// 关闭Redis连接
	return r.client.Close()
}

// prefixKey 为键添加前缀
func (r *RedisStore) prefixKey(key string) string {
	return r.prefix + key
}

// tagKey 为标签键添加前缀
func (r *RedisStore) tagKey(tag string) string {
	return r.prefix + "tag:" + tag
}

// Get 从缓存中获取一个项目
func (r *RedisStore) Get(ctx context.Context, key string) (interface{}, error) {
	prefixedKey := r.prefixKey(key)
	val, err := r.client.Get(ctx, prefixedKey).Result()
	if err == redis.Nil {
		return nil, ErrCacheMiss
	}
	if err != nil {
		return nil, err
	}

	var item Item
	if err := json.Unmarshal([]byte(val), &item); err != nil {
		return nil, err
	}

	return item.Value, nil
}

// Set 将一个项目放入缓存
func (r *RedisStore) Set(ctx context.Context, key string, value interface{}, options ...Option) error {
	opts := applyOptions(options...)
	expiration := r.defaultExpiry
	if opts.Expiration > 0 {
		expiration = opts.Expiration
	}

	item := Item{
		Key:        key,
		Value:      value,
		Tags:       opts.Tags,
		Expiration: expiration,
		CreatedAt:  time.Now(),
	}

	// 序列化缓存项
	jsonData, err := json.Marshal(item)
	if err != nil {
		return err
	}

	prefixedKey := r.prefixKey(key)

	// 使用管道执行多个命令
	pipe := r.client.Pipeline()

	// 设置主键值
	pipe.Set(ctx, prefixedKey, jsonData, expiration)

	// 处理标签
	if len(opts.Tags) > 0 {
		// 使用标签管理器关联标签和键
		if err := r.tagManager.AddTagsToKey(ctx, key, opts.Tags); err != nil {
			return err
		}
	}

	_, err = pipe.Exec(ctx)
	return err
}

// Delete 从缓存中删除一个项目
func (r *RedisStore) Delete(ctx context.Context, key string) error {
	prefixedKey := r.prefixKey(key)

	// 先获取项以找到关联的标签
	val, err := r.client.Get(ctx, prefixedKey).Result()
	if err == redis.Nil {
		// 键不存在，不需要操作
		return nil
	} else if err != nil {
		return err
	}

	var item Item
	if err := json.Unmarshal([]byte(val), &item); err != nil {
		// 如果解析失败，仍删除主键
		return r.client.Del(ctx, prefixedKey).Err()
	}

	// 使用管道执行多个命令
	pipe := r.client.Pipeline()

	// 删除主键
	pipe.Del(ctx, prefixedKey)

	// 从所有标签中移除此键
	if len(item.Tags) > 0 {
		r.tagManager.RemoveKeyFromAllTags(ctx, key)
	}

	_, err = pipe.Exec(ctx)
	return err
}

// Has 检查缓存中是否存在一个项目
func (r *RedisStore) Has(ctx context.Context, key string) bool {
	prefixedKey := r.prefixKey(key)
	val, err := r.client.Exists(ctx, prefixedKey).Result()
	return err == nil && val > 0
}

// Clear 清空缓存
func (r *RedisStore) Clear(ctx context.Context) error {
	keys, err := r.client.Keys(ctx, r.prefix+"*").Result()
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		return nil
	}

	return r.client.Del(ctx, keys...).Err()
}

// GetMultiple 批量获取多个缓存项
func (r *RedisStore) GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error) {
	if len(keys) == 0 {
		return make(map[string]interface{}), nil
	}

	prefixedKeys := make([]string, len(keys))
	for i, key := range keys {
		prefixedKeys[i] = r.prefixKey(key)
	}

	pipeline := r.client.Pipeline()
	cmds := make([]*redis.StringCmd, len(prefixedKeys))

	for i, key := range prefixedKeys {
		cmds[i] = pipeline.Get(ctx, key)
	}

	_, err := pipeline.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, err
	}

	result := make(map[string]interface{}, len(keys))
	for i, cmd := range cmds {
		val, err := cmd.Result()
		if err == redis.Nil {
			continue
		}
		if err != nil {
			return nil, err
		}

		var item Item
		if err := json.Unmarshal([]byte(val), &item); err != nil {
			return nil, err
		}

		// 去掉前缀，返回原始键
		result[keys[i]] = item.Value
	}

	return result, nil
}

// SetMultiple 批量设置多个缓存项
func (r *RedisStore) SetMultiple(ctx context.Context, items map[string]interface{}, options ...Option) error {
	if len(items) == 0 {
		return nil
	}

	opts := applyOptions(options...)
	expiration := r.defaultExpiry
	if opts.Expiration > 0 {
		expiration = opts.Expiration
	}

	now := time.Now()

	pipeline := r.client.Pipeline()
	var allKeys []string

	for key, value := range items {
		item := Item{
			Key:        key,
			Value:      value,
			Tags:       opts.Tags,
			Expiration: expiration,
			CreatedAt:  now,
		}

		// 序列化缓存项
		jsonData, err := json.Marshal(item)
		if err != nil {
			return err
		}

		prefixedKey := r.prefixKey(key)
		allKeys = append(allKeys, key)

		// 设置主键值
		pipeline.Set(ctx, prefixedKey, jsonData, expiration)
	}

	// 处理标签
	if len(opts.Tags) > 0 {
		for _, key := range allKeys {
			// 使用标签管理器关联标签和键
			if err := r.tagManager.AddTagsToKey(ctx, key, opts.Tags); err != nil {
				return err
			}
		}
	}

	_, err := pipeline.Exec(ctx)
	return err
}

// DeleteMultiple 批量删除多个缓存项
func (r *RedisStore) DeleteMultiple(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	for _, key := range keys {
		// 从所有标签中移除键
		if err := r.tagManager.RemoveKeyFromAllTags(ctx, key); err != nil {
			// 记录错误但继续执行
			fmt.Printf("Error removing key %s from tags: %v\n", key, err)
		}
	}

	prefixedKeys := make([]string, len(keys))
	for i, key := range keys {
		prefixedKeys[i] = r.prefixKey(key)
	}

	// 删除主键
	return r.client.Del(ctx, prefixedKeys...).Err()
}

// Increment 增加缓存项的整数值
func (r *RedisStore) Increment(ctx context.Context, key string, value int64) (int64, error) {
	prefixedKey := r.prefixKey(key)

	// 检查键是否存在
	exists, err := r.client.Exists(ctx, prefixedKey).Result()
	if err != nil {
		return 0, err
	}

	if exists == 0 {
		// 键不存在，创建一个新的缓存项
		item := Item{
			Key:        key,
			Value:      value,
			Tags:       []string{},
			Expiration: r.defaultExpiry,
			CreatedAt:  time.Now(),
		}

		jsonData, err := json.Marshal(item)
		if err != nil {
			return 0, err
		}

		// 设置初始值
		if err := r.client.Set(ctx, prefixedKey, jsonData, r.defaultExpiry).Err(); err != nil {
			return 0, err
		}

		return value, nil
	}

	// 键存在，获取当前值
	val, err := r.client.Get(ctx, prefixedKey).Result()
	if err != nil {
		return 0, err
	}

	var item Item
	if err := json.Unmarshal([]byte(val), &item); err != nil {
		return 0, err
	}

	// 将当前值转换为int64
	var currentVal int64 = 0
	switch v := item.Value.(type) {
	case float64:
		currentVal = int64(v)
	case int:
		currentVal = int64(v)
	case int64:
		currentVal = v
	case string:
		// 尝试从字符串解析
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			currentVal = i
		}
	}

	// 增加值
	newVal := currentVal + value
	item.Value = newVal

	// 保存回缓存
	jsonData, err := json.Marshal(item)
	if err != nil {
		return 0, err
	}

	// 使用原始过期时间
	ttl := item.Expiration
	if ttl <= 0 {
		ttl = r.defaultExpiry
	}

	if err := r.client.Set(ctx, prefixedKey, jsonData, ttl).Err(); err != nil {
		return 0, err
	}

	return newVal, nil
}

// IncrementFloat 增加缓存项的浮点值
func (r *RedisStore) IncrementFloat(ctx context.Context, key string, value float64) (float64, error) {
	prefixedKey := r.prefixKey(key)

	// 检查键是否存在
	exists, err := r.client.Exists(ctx, prefixedKey).Result()
	if err != nil {
		return 0, err
	}

	if exists == 0 {
		// 键不存在，创建一个新的缓存项
		item := Item{
			Key:        key,
			Value:      value,
			Tags:       []string{},
			Expiration: r.defaultExpiry,
			CreatedAt:  time.Now(),
		}

		jsonData, err := json.Marshal(item)
		if err != nil {
			return 0, err
		}

		// 设置初始值
		if err := r.client.Set(ctx, prefixedKey, jsonData, r.defaultExpiry).Err(); err != nil {
			return 0, err
		}

		return value, nil
	}

	// 键存在，获取当前值
	val, err := r.client.Get(ctx, prefixedKey).Result()
	if err != nil {
		return 0, err
	}

	var item Item
	if err := json.Unmarshal([]byte(val), &item); err != nil {
		return 0, err
	}

	// 将当前值转换为float64
	var currentVal float64 = 0
	switch v := item.Value.(type) {
	case float64:
		currentVal = v
	case int:
		currentVal = float64(v)
	case int64:
		currentVal = float64(v)
	case string:
		// 尝试从字符串解析
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			currentVal = f
		}
	}

	// 增加值
	newVal := currentVal + value
	item.Value = newVal

	// 保存回缓存
	jsonData, err := json.Marshal(item)
	if err != nil {
		return 0, err
	}

	// 使用原始过期时间
	ttl := item.Expiration
	if ttl <= 0 {
		ttl = r.defaultExpiry
	}

	if err := r.client.Set(ctx, prefixedKey, jsonData, ttl).Err(); err != nil {
		return 0, err
	}

	return newVal, nil
}

// Decrement 减少缓存项的整数值
func (r *RedisStore) Decrement(ctx context.Context, key string, value int64) (int64, error) {
	return r.Increment(ctx, key, -value)
}

// DecrementFloat 减少缓存项的浮点值
func (r *RedisStore) DecrementFloat(ctx context.Context, key string, value float64) (float64, error) {
	return r.IncrementFloat(ctx, key, -value)
}

// TaggedGet 获取带标签的缓存项
func (r *RedisStore) TaggedGet(ctx context.Context, tag string) (map[string]interface{}, error) {
	// 使用标签管理器获取标签关联的所有键
	keys, err := r.tagManager.GetKeysByTag(ctx, tag)
	if err != nil {
		return nil, err
	}

	if len(keys) == 0 {
		return make(map[string]interface{}), nil
	}

	return r.GetMultiple(ctx, keys)
}

// TaggedDelete 删除带标签的所有缓存项
func (r *RedisStore) TaggedDelete(ctx context.Context, tag string) error {
	// 使用标签管理器获取标签关联的所有键
	keys, err := r.tagManager.GetKeysByTag(ctx, tag)
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		return nil
	}

	// 删除所有相关的键
	if err := r.DeleteMultiple(ctx, keys); err != nil {
		return err
	}

	// 删除标签本身
	return r.tagManager.RemoveTag(ctx, tag)
}

// Count 返回缓存中的项目数
func (r *RedisStore) Count(ctx context.Context) int64 {
	keys, err := r.client.Keys(ctx, r.prefix+"*").Result()
	if err != nil {
		return 0
	}

	// 排除标签键
	var count int64 = 0
	tagPrefix := r.prefix + "tag:"
	for _, key := range keys {
		if len(key) < len(tagPrefix) || key[:len(tagPrefix)] != tagPrefix {
			count++
		}
	}

	return count
}

// Flush 刷新缓存，删除所有项目
func (r *RedisStore) Flush(ctx context.Context) error {
	return r.Clear(ctx)
}

// GetClient 获取Redis客户端
func (r *RedisStore) GetClient() *redis.Client {
	return r.client
}

// GetPrefix 获取键前缀
func (r *RedisStore) GetPrefix() string {
	return r.prefix
}

// GetDefaultExpiry 获取默认过期时间
func (r *RedisStore) GetDefaultExpiry() time.Duration {
	return r.defaultExpiry
}

// GetTagManager 获取标签管理器
func (r *RedisStore) GetTagManager() TagManager {
	return r.tagManager
}

// RedisTagManager 实现TagManager接口，使用Redis作为存储后端
type RedisTagManager struct {
	client *redis.Client
	prefix string
}

// NewRedisTagManager 创建一个新的Redis标签管理器
func NewRedisTagManager(client *redis.Client, prefix string) *RedisTagManager {
	return &RedisTagManager{
		client: client,
		prefix: prefix,
	}
}

// tagKey 生成标签键名
func (m *RedisTagManager) tagKey(tag string) string {
	return m.prefix + "tag:" + tag
}

// keyTagsKey 生成键到标签映射的键名
func (m *RedisTagManager) keyTagsKey(key string) string {
	return m.prefix + "key_tags:" + key
}

// prefixKey 为键添加前缀
func (m *RedisTagManager) prefixKey(key string) string {
	return m.prefix + key
}

// AddTagsToKey 为缓存键添加标签
func (m *RedisTagManager) AddTagsToKey(ctx context.Context, key string, tags []string) error {
	if len(tags) == 0 {
		return nil
	}

	prefixedKey := m.prefixKey(key)
	keyTagsKey := m.keyTagsKey(key)
	pipe := m.client.Pipeline()

	// 将键添加到每个标签的集合中
	for _, tag := range tags {
		tagKey := m.tagKey(tag)
		pipe.SAdd(ctx, tagKey, prefixedKey)
	}

	// 存储键关联的所有标签
	pipe.SAdd(ctx, keyTagsKey, tags)

	_, err := pipe.Exec(ctx)
	return err
}

// RemoveTagsFromKey 从缓存键中移除标签
func (m *RedisTagManager) RemoveTagsFromKey(ctx context.Context, key string, tags []string) error {
	if len(tags) == 0 {
		return nil
	}

	prefixedKey := m.prefixKey(key)
	keyTagsKey := m.keyTagsKey(key)
	pipe := m.client.Pipeline()

	// 从每个标签的集合中移除键
	for _, tag := range tags {
		tagKey := m.tagKey(tag)
		pipe.SRem(ctx, tagKey, prefixedKey)
	}

	// 从键的标签集合中移除这些标签
	pipe.SRem(ctx, keyTagsKey, tags)

	_, err := pipe.Exec(ctx)
	return err
}

// GetKeysByTag 根据标签获取所有关联的键
func (m *RedisTagManager) GetKeysByTag(ctx context.Context, tag string) ([]string, error) {
	tagKey := m.tagKey(tag)
	prefixedKeys, err := m.client.SMembers(ctx, tagKey).Result()
	if err != nil {
		return nil, err
	}

	// 移除前缀
	prefixLen := len(m.prefix)
	keys := make([]string, len(prefixedKeys))
	for i, prefixedKey := range prefixedKeys {
		if len(prefixedKey) > prefixLen {
			keys[i] = prefixedKey[prefixLen:]
		} else {
			keys[i] = prefixedKey
		}
	}

	return keys, nil
}

// RemoveTag 移除标签及其所有关联
func (m *RedisTagManager) RemoveTag(ctx context.Context, tag string) error {
	tagKey := m.tagKey(tag)

	// 获取标签关联的所有键
	prefixedKeys, err := m.client.SMembers(ctx, tagKey).Result()
	if err != nil {
		return err
	}

	pipe := m.client.Pipeline()

	// 从每个键的标签集合中移除此标签
	for _, prefixedKey := range prefixedKeys {
		// 从前缀键中提取原始键
		originalKey := ""
		if len(prefixedKey) > len(m.prefix) {
			originalKey = prefixedKey[len(m.prefix):]
		} else {
			originalKey = prefixedKey
		}

		keyTagsKey := m.keyTagsKey(originalKey)
		pipe.SRem(ctx, keyTagsKey, tag)
	}

	// 删除标签集合
	pipe.Del(ctx, tagKey)

	_, err = pipe.Exec(ctx)
	return err
}

// RemoveKeyFromAllTags 从所有标签中移除指定的键
func (m *RedisTagManager) RemoveKeyFromAllTags(ctx context.Context, key string) error {
	keyTagsKey := m.keyTagsKey(key)
	prefixedKey := m.prefixKey(key)

	// 获取键关联的所有标签
	tags, err := m.client.SMembers(ctx, keyTagsKey).Result()
	if err != nil {
		return err
	}

	pipe := m.client.Pipeline()

	// 从每个标签集合中移除键
	for _, tag := range tags {
		tagKey := m.tagKey(tag)
		pipe.SRem(ctx, tagKey, prefixedKey)
	}

	// 删除键的标签集合
	pipe.Del(ctx, keyTagsKey)

	_, err = pipe.Exec(ctx)
	return err
}

// GetKeyTags 获取键关联的所有标签
func (m *RedisTagManager) GetKeyTags(ctx context.Context, key string) ([]string, error) {
	keyTagsKey := m.keyTagsKey(key)
	return m.client.SMembers(ctx, keyTagsKey).Result()
}

// RedisDriver Redis缓存驱动
type RedisDriver struct{}

// New 创建新的Redis缓存实例
func (d *RedisDriver) New(config map[string]interface{}) (Store, error) {
	var client *redis.Client
	ctx := context.Background()

	// 检查客户端是否已通过DI传入
	if c, ok := config["client"].(*redis.Client); ok {
		client = c
	} else {
		// 创建Redis客户端
		addr := "127.0.0.1:6379"
		db := 0
		var password string
		var username string
		maxRetries := 3
		poolSize := 10
		minIdleConns := 2

		// 解析配置
		if a, ok := config["addr"].(string); ok {
			addr = a
		} else if host, ok := config["host"].(string); ok {
			port := "6379"
			if p, ok := config["port"].(string); ok {
				port = p
			} else if p, ok := config["port"].(int); ok {
				port = fmt.Sprintf("%d", p)
			}
			addr = fmt.Sprintf("%s:%s", host, port)
		}

		if d, ok := config["db"].(int); ok {
			db = d
		}
		if p, ok := config["password"].(string); ok {
			password = p
		}
		if u, ok := config["username"].(string); ok {
			username = u
		}
		if m, ok := config["max_retries"].(int); ok {
			maxRetries = m
		}
		if p, ok := config["pool_size"].(int); ok {
			poolSize = p
		}
		if m, ok := config["min_idle_conns"].(int); ok {
			minIdleConns = m
		}

		// 创建标准的Redis客户端
		client = redis.NewClient(&redis.Options{
			Addr:         addr,
			DB:           db,
			Password:     password,
			Username:     username,
			MaxRetries:   maxRetries,
			PoolSize:     poolSize,
			MinIdleConns: minIdleConns,
		})
	}

	// 测试连接
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %s", err)
	}

	// 解析Redis选项
	var prefix string = "flow:"
	var expiry time.Duration = 5 * time.Minute
	healthCheck := true
	healthCheckInterval := 30 * time.Second

	if p, ok := config["prefix"].(string); ok {
		prefix = p
	}
	if e, ok := config["expiry"].(time.Duration); ok {
		expiry = e
	} else if e, ok := config["expiry"].(string); ok {
		if dur, err := time.ParseDuration(e); err == nil {
			expiry = dur
		}
	} else if e, ok := config["ttl"].(string); ok {
		if dur, err := time.ParseDuration(e); err == nil {
			expiry = dur
		}
	}

	if h, ok := config["health_check"].(bool); ok {
		healthCheck = h
	}
	if h, ok := config["health_check_interval"].(time.Duration); ok {
		healthCheckInterval = h
	} else if h, ok := config["health_check_interval"].(string); ok {
		if dur, err := time.ParseDuration(h); err == nil {
			healthCheckInterval = dur
		}
	}

	// 创建Redis标签管理器
	tagManager := NewRedisTagManager(client, prefix)

	// 创建Redis存储
	store := NewRedisStore(client,
		WithRedisPrefix(prefix),
		WithRedisExpiry(expiry),
		WithRedisTagManager(tagManager),
		WithRedisHealthCheck(healthCheck, healthCheckInterval),
	)

	return store, nil
}

// 注册Redis缓存驱动
func init() {
	RegisterDriver("redis", &RedisDriver{})
}
