package cache

import (
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// 文件缓存的配置选项
type FileConfig struct {
	Directory  string        // 缓存文件存储目录
	Extension  string        // 缓存文件扩展名
	GCInterval time.Duration // 垃圾回收间隔
}

// FileDriver 文件缓存驱动
type FileDriver struct{}

// 确保实现了 Driver 接口
var _ Driver = (*FileDriver)(nil)

// 初始化函数，注册驱动
func init() {
	// 注册 gob 类型
	gob.Register(time.Time{})
	gob.Register(map[string]interface{}{})
	gob.Register([]interface{}{})

	// 注册驱动
	RegisterDriver("file", &FileDriver{})
}

// New 创建新的文件缓存存储
func (d *FileDriver) New(config map[string]interface{}) (Store, error) {
	// 解析配置
	fileConfig := FileConfig{
		Directory:  "storage/cache",
		Extension:  ".cache",
		GCInterval: 60 * time.Minute,
	}

	// 目录配置
	if dir, ok := config["directory"].(string); ok && dir != "" {
		fileConfig.Directory = dir
	}

	// 扩展名配置
	if ext, ok := config["extension"].(string); ok && ext != "" {
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		fileConfig.Extension = ext
	}

	// GC 间隔配置
	if gcInterval, ok := config["gc_interval"].(time.Duration); ok && gcInterval > 0 {
		fileConfig.GCInterval = gcInterval
	} else if gcIntervalStr, ok := config["gc_interval"].(string); ok && gcIntervalStr != "" {
		if duration, err := time.ParseDuration(gcIntervalStr); err == nil {
			fileConfig.GCInterval = duration
		}
	}

	// 确保缓存目录存在
	if err := os.MkdirAll(fileConfig.Directory, 0755); err != nil {
		return nil, fmt.Errorf("创建缓存目录失败: %w", err)
	}

	// 创建文件缓存存储
	store := &FileStore{
		directory:  fileConfig.Directory,
		extension:  fileConfig.Extension,
		gcInterval: fileConfig.GCInterval,
		mutex:      &sync.RWMutex{},
		tagManager: NewTagManager(nil), // 创建标签管理器
	}

	// 设置自身引用
	store.tagManager.(*StandardTagManager).store = store

	// 启动垃圾回收
	store.startGC()

	return store, nil
}

// fileCacheItem 文件缓存项
type fileCacheItem struct {
	Key        string
	Value      interface{}
	Expiration int64 // 存储为 Unix 纳秒时间戳
	Tags       []string
	CreatedAt  time.Time
}

// FileStore 文件缓存存储
type FileStore struct {
	directory  string
	extension  string
	gcInterval time.Duration
	mutex      *sync.RWMutex
	tagManager TagManager
	gcRunning  bool
}

// 确保实现了 Store 接口
var _ Store = (*FileStore)(nil)

// 获取项目缓存路径
func (s *FileStore) getCachePath(key string) string {
	// 对键进行文件名安全处理
	safeKey := strings.ReplaceAll(key, "/", "_")
	safeKey = strings.ReplaceAll(safeKey, "\\", "_")
	safeKey = strings.ReplaceAll(safeKey, ":", "_")
	safeKey = strings.ReplaceAll(safeKey, "*", "_")
	safeKey = strings.ReplaceAll(safeKey, "?", "_")
	safeKey = strings.ReplaceAll(safeKey, "\"", "_")
	safeKey = strings.ReplaceAll(safeKey, "<", "_")
	safeKey = strings.ReplaceAll(safeKey, ">", "_")
	safeKey = strings.ReplaceAll(safeKey, "|", "_")

	return filepath.Join(s.directory, safeKey+s.extension)
}

// 保存缓存项到文件
func (s *FileStore) saveItemToFile(item fileCacheItem) error {
	path := s.getCachePath(item.Key)

	// 创建临时文件
	tempFile, err := os.CreateTemp(s.directory, "temp-*"+s.extension)
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %w", err)
	}
	tempFilePath := tempFile.Name()

	// 确保临时文件会被删除
	defer func() {
		tempFile.Close()
		// 如果文件仍然存在（发生错误），则删除它
		os.Remove(tempFilePath)
	}()

	// 编码缓存项
	encoder := gob.NewEncoder(tempFile)
	if err := encoder.Encode(item); err != nil {
		return fmt.Errorf("编码缓存项失败: %w", err)
	}

	// 确保写入已完成
	if err := tempFile.Sync(); err != nil {
		return fmt.Errorf("文件同步失败: %w", err)
	}

	// 关闭临时文件
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("关闭临时文件失败: %w", err)
	}

	// 原子性地替换目标文件
	if err := os.Rename(tempFilePath, path); err != nil {
		return fmt.Errorf("重命名缓存文件失败: %w", err)
	}

	return nil
}

// 从文件加载缓存项
func (s *FileStore) loadItemFromFile(key string) (*fileCacheItem, error) {
	path := s.getCachePath(key)

	// 打开文件
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrCacheMiss
		}
		return nil, fmt.Errorf("打开缓存文件失败: %w", err)
	}
	defer file.Close()

	// 解码缓存项
	var item fileCacheItem
	decoder := gob.NewDecoder(file)
	if err := decoder.Decode(&item); err != nil {
		// 格式错误的文件，删除它
		file.Close()
		os.Remove(path)
		return nil, fmt.Errorf("解码缓存项失败: %w", err)
	}

	// 检查是否过期
	if item.Expiration > 0 && time.Now().UnixNano() > item.Expiration {
		// 已过期，删除文件
		file.Close()
		os.Remove(path)
		return nil, ErrCacheMiss
	}

	return &item, nil
}

// Get 获取缓存
func (s *FileStore) Get(ctx context.Context, key string) (interface{}, error) {
	if key == "" {
		return nil, ErrInvalidKey
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	item, err := s.loadItemFromFile(key)
	if err != nil {
		return nil, err
	}

	return item.Value, nil
}

// Set 设置缓存
func (s *FileStore) Set(ctx context.Context, key string, value interface{}, options ...Option) error {
	if key == "" {
		return ErrInvalidKey
	}

	opts := applyOptions(options...)

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 准备缓存项
	item := fileCacheItem{
		Key:       key,
		Value:     value,
		Tags:      opts.Tags,
		CreatedAt: time.Now(),
	}

	// 设置过期时间
	if opts.Expiration > 0 {
		item.Expiration = time.Now().Add(opts.Expiration).UnixNano()
	}

	// 保存缓存项
	if err := s.saveItemToFile(item); err != nil {
		return err
	}

	// 如果有标签，将标签与键关联
	if len(opts.Tags) > 0 {
		return s.tagManager.AddTagsToKey(ctx, key, opts.Tags)
	}

	return nil
}

// Delete 删除缓存
func (s *FileStore) Delete(ctx context.Context, key string) error {
	if key == "" {
		return ErrInvalidKey
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 删除缓存文件
	path := s.getCachePath(key)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除缓存文件失败: %w", err)
	}

	// 从所有标签中移除此键
	return s.tagManager.RemoveKeyFromAllTags(ctx, key)
}

// Has 检查缓存是否存在
func (s *FileStore) Has(ctx context.Context, key string) bool {
	if key == "" {
		return false
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	_, err := s.loadItemFromFile(key)
	return err == nil
}

// Clear 清空缓存
func (s *FileStore) Clear(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 遍历并删除所有缓存文件
	entries, err := os.ReadDir(s.directory)
	if err != nil {
		return fmt.Errorf("读取缓存目录失败: %w", err)
	}

	for _, entry := range entries {
		// 跳过目录
		if entry.IsDir() {
			continue
		}

		// 检查扩展名
		if !strings.HasSuffix(entry.Name(), s.extension) {
			continue
		}

		// 删除文件
		path := filepath.Join(s.directory, entry.Name())
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("删除缓存文件失败: %w", err)
		}
	}

	return nil
}

// GetMultiple 获取多个缓存项
func (s *FileStore) GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error) {
	if len(keys) == 0 {
		return make(map[string]interface{}), nil
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	result := make(map[string]interface{}, len(keys))

	for _, key := range keys {
		if key == "" {
			continue
		}

		item, err := s.loadItemFromFile(key)
		if err == nil {
			result[key] = item.Value
		}
	}

	return result, nil
}

// SetMultiple 设置多个缓存项
func (s *FileStore) SetMultiple(ctx context.Context, items map[string]interface{}, options ...Option) error {
	if len(items) == 0 {
		return nil
	}

	opts := applyOptions(options...)

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 处理每个缓存项
	for key, value := range items {
		if key == "" {
			continue
		}

		// 准备缓存项
		item := fileCacheItem{
			Key:       key,
			Value:     value,
			Tags:      opts.Tags,
			CreatedAt: time.Now(),
		}

		// 设置过期时间
		if opts.Expiration > 0 {
			item.Expiration = time.Now().Add(opts.Expiration).UnixNano()
		}

		// 保存缓存项
		if err := s.saveItemToFile(item); err != nil {
			return err
		}

		// 如果有标签，将标签与键关联
		if len(opts.Tags) > 0 {
			if err := s.tagManager.AddTagsToKey(ctx, key, opts.Tags); err != nil {
				return err
			}
		}
	}

	return nil
}

// DeleteMultiple 删除多个缓存项
func (s *FileStore) DeleteMultiple(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, key := range keys {
		if key == "" {
			continue
		}

		// 删除缓存文件
		path := s.getCachePath(key)
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("删除缓存文件失败: %w", err)
		}

		// 从所有标签中移除此键
		if err := s.tagManager.RemoveKeyFromAllTags(ctx, key); err != nil {
			return err
		}
	}

	return nil
}

// Increment 增加计数器值
func (s *FileStore) Increment(ctx context.Context, key string, value int64) (int64, error) {
	if key == "" {
		return 0, ErrInvalidKey
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 加载当前值
	item, err := s.loadItemFromFile(key)

	var currentValue int64 = 0

	if err == nil {
		// 根据当前值类型进行转换
		switch v := item.Value.(type) {
		case int:
			currentValue = int64(v)
		case int64:
			currentValue = v
		case int32:
			currentValue = int64(v)
		case int16:
			currentValue = int64(v)
		case int8:
			currentValue = int64(v)
		case uint:
			currentValue = int64(v)
		case uint64:
			currentValue = int64(v)
		case uint32:
			currentValue = int64(v)
		case uint16:
			currentValue = int64(v)
		case uint8:
			currentValue = int64(v)
		case float64:
			currentValue = int64(v)
		case float32:
			currentValue = int64(v)
		case string:
			// 尝试将字符串转换为整数
			var parseErr error
			currentValue, parseErr = parseInt64(v)
			if parseErr != nil {
				return 0, errors.New("无法增加非数值类型的缓存")
			}
		default:
			return 0, errors.New("无法增加非数值类型的缓存")
		}
	}

	// 增加值
	newValue := currentValue + value

	// 创建新的缓存项
	newItem := fileCacheItem{
		Key:       key,
		Value:     newValue,
		CreatedAt: time.Now(),
	}

	// 保留过期时间和标签
	if item != nil {
		newItem.Expiration = item.Expiration
		newItem.Tags = item.Tags
	}

	// 保存新值
	if err := s.saveItemToFile(newItem); err != nil {
		return 0, err
	}

	return newValue, nil
}

// parseInt64 将字符串转换为 int64
func parseInt64(s string) (int64, error) {
	var result int64
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}

// Decrement 减少计数器值
func (s *FileStore) Decrement(ctx context.Context, key string, value int64) (int64, error) {
	return s.Increment(ctx, key, -value)
}

// TaggedGet 获取带有标签的缓存项
func (s *FileStore) TaggedGet(ctx context.Context, tag string) (map[string]interface{}, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// 获取带有此标签的所有键
	keys, err := s.tagManager.GetKeysByTag(ctx, tag)
	if err != nil {
		return nil, err
	}

	if len(keys) == 0 {
		return make(map[string]interface{}), nil
	}

	// 获取所有键对应的值
	result := make(map[string]interface{}, len(keys))
	for _, key := range keys {
		item, err := s.loadItemFromFile(key)
		if err == nil {
			result[key] = item.Value
		}
	}

	return result, nil
}

// TaggedDelete 删除带有标签的缓存项
func (s *FileStore) TaggedDelete(ctx context.Context, tag string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 获取带有此标签的所有键
	keys, err := s.tagManager.GetKeysByTag(ctx, tag)
	if err != nil {
		return err
	}

	// 删除所有键
	for _, key := range keys {
		path := s.getCachePath(key)
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("删除缓存文件失败: %w", err)
		}
	}

	// 移除标签
	return s.tagManager.RemoveTag(ctx, tag)
}

// Count 计算缓存数量
func (s *FileStore) Count(ctx context.Context) int64 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var count int64 = 0

	// 读取目录
	entries, err := os.ReadDir(s.directory)
	if err != nil {
		return 0
	}

	// 计算有效缓存文件数量
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if strings.HasSuffix(entry.Name(), s.extension) {
			// 尝试读取文件，看是否是有效缓存
			key := strings.TrimSuffix(entry.Name(), s.extension)
			if _, err := s.loadItemFromFile(key); err == nil {
				count++
			}
		}
	}

	return count
}

// Flush 执行一次缓存清理
func (s *FileStore) Flush(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 读取目录
	entries, err := os.ReadDir(s.directory)
	if err != nil {
		return fmt.Errorf("读取缓存目录失败: %w", err)
	}

	// 检查每个文件
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if !strings.HasSuffix(entry.Name(), s.extension) {
			continue
		}

		path := filepath.Join(s.directory, entry.Name())

		// 打开文件
		file, err := os.Open(path)
		if err != nil {
			continue
		}

		// 解码缓存项
		var item fileCacheItem
		decoder := gob.NewDecoder(file)
		if err := decoder.Decode(&item); err != nil {
			// 格式错误的文件，删除它
			file.Close()
			os.Remove(path)
			continue
		}
		file.Close()

		// 检查是否过期
		if item.Expiration > 0 && time.Now().UnixNano() > item.Expiration {
			// 删除过期的缓存文件
			os.Remove(path)
			// 从所有标签中移除此键
			s.tagManager.RemoveKeyFromAllTags(ctx, item.Key)
		}
	}

	return nil
}

// 启动垃圾回收协程
func (s *FileStore) startGC() {
	if s.gcInterval <= 0 || s.gcRunning {
		return
	}

	s.gcRunning = true

	go func() {
		ticker := time.NewTicker(s.gcInterval)
		defer ticker.Stop()

		for range ticker.C {
			// 执行一次垃圾回收
			s.Flush(context.Background())
		}
	}()
}

// 检查文件修改时间是否超过过期时间
func isFileExpired(fileInfo fs.FileInfo, expiration int64) bool {
	if expiration <= 0 {
		return false
	}
	// 使用文件的修改时间与过期时间比较
	return fileInfo.ModTime().UnixNano()+expiration < time.Now().UnixNano()
}
