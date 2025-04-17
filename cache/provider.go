package cache

import (
	"context"
	"time"

	"github.com/zzliekkas/flow/app"
	"github.com/zzliekkas/flow/config"
)

// CacheProvider 缓存服务提供者
type CacheProvider struct {
	*app.BaseProvider
}

// NewCacheProvider 创建缓存服务提供者
func NewCacheProvider() *CacheProvider {
	return &CacheProvider{
		BaseProvider: app.NewBaseProvider("cache", 50), // 优先级50
	}
}

// Register 注册缓存服务
func (p *CacheProvider) Register(application *app.Application) error {
	application.Logger().Info("注册缓存服务...")

	// 创建缓存管理器
	manager := NewManager()

	// 从配置加载缓存设置
	p.loadCacheConfig(application, manager)

	// 向DI容器注册缓存管理器
	return application.Engine().Provide(func() *Manager {
		return manager
	})
}

// Boot 启动缓存服务
func (p *CacheProvider) Boot(application *app.Application) error {
	application.Logger().Info("启动缓存服务...")

	// 注册缓存管理器的关闭钩子
	application.OnBeforeShutdown("flush_cache", func() {
		var manager *Manager
		if err := application.Engine().Invoke(func(m *Manager) {
			manager = m
		}); err == nil && manager != nil {
			// 刷新所有缓存存储
			for _, storeName := range manager.getStoreNames() {
				if store, err := manager.GetStore(storeName); err == nil {
					_ = store.Flush(context.Background())
				}
			}
			application.Logger().Info("缓存已刷新")
		}
	}, 100)

	return nil
}

// 加载缓存配置
func (p *CacheProvider) loadCacheConfig(application *app.Application, manager *Manager) {
	var configManager *config.Manager
	if err := application.Engine().Invoke(func(cm *config.Manager) {
		configManager = cm
	}); err != nil || configManager == nil {
		// 配置管理器不可用，使用默认配置
		p.registerDefaultConfig(manager)
		return
	}

	// 获取缓存配置
	cacheConfig := configManager.GetStringMap("cache")
	if len(cacheConfig) == 0 {
		// 没有缓存配置，使用默认配置
		p.registerDefaultConfig(manager)
		return
	}

	// 获取默认缓存驱动
	defaultStore := "memory"
	if def, ok := cacheConfig["default"].(string); ok && def != "" {
		defaultStore = def
	}

	// 获取存储配置
	stores := make(map[string]interface{})
	if storesConfig, ok := cacheConfig["stores"].(map[string]interface{}); ok {
		stores = storesConfig
	}

	// 注册所有缓存配置
	for storeName, storeConfig := range stores {
		if config, ok := storeConfig.(map[string]interface{}); ok {
			// 获取驱动类型
			driver := "memory"
			if d, ok := config["driver"].(string); ok && d != "" {
				driver = d
			}

			// 获取键前缀
			prefix := ""
			if p, ok := config["prefix"].(string); ok {
				prefix = p
			}

			// 获取过期时间
			var ttl time.Duration
			if t, ok := config["ttl"].(int); ok && t > 0 {
				ttl = time.Duration(t) * time.Second
			} else if t, ok := config["ttl"].(string); ok && t != "" {
				if parsedTTL, err := time.ParseDuration(t); err == nil {
					ttl = parsedTTL
				}
			}

			// 注册配置
			manager.Register(storeName, Config{
				Driver: driver,
				Prefix: prefix,
				TTL:    ttl,
				Config: config,
			})

			application.Logger().Infof("已注册缓存存储: %s (驱动: %s)", storeName, driver)
		}
	}

	// 设置默认存储
	manager.SetDefault(defaultStore)
	application.Logger().Infof("默认缓存存储: %s", defaultStore)
}

// 注册默认配置
func (p *CacheProvider) registerDefaultConfig(manager *Manager) {
	// 注册内存缓存
	manager.Register("memory", Config{
		Driver: "memory",
		Prefix: "",
		TTL:    5 * time.Minute,
		Config: map[string]interface{}{},
	})

	// 设置默认存储
	manager.SetDefault("memory")
}

// 获取所有存储名称
func (m *Manager) getStoreNames() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	names := make([]string, 0, len(m.stores))
	for name := range m.stores {
		names = append(names, name)
	}
	return names
}

// GetInstance 获取缓存管理器实例（便捷方法）
func GetInstance(application *app.Application) (*Manager, error) {
	var manager *Manager
	err := application.Engine().Invoke(func(m *Manager) {
		manager = m
	})
	return manager, err
}
