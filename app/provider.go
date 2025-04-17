package app

import (
	"fmt"
	"sort"
	"sync"
)

// ServiceProvider 服务提供者接口
type ServiceProvider interface {
	// Register 向DI容器注册服务
	Register(app *Application) error

	// Boot 在所有服务都注册后启动服务
	Boot(app *Application) error

	// Name 获取提供者名称
	Name() string

	// Priority 获取提供者优先级，数值越小优先级越高
	Priority() int
}

// ProviderManager 提供者管理器
type ProviderManager struct {
	providers       []ServiceProvider // 注册的服务提供者
	bootedProviders map[string]bool   // 已启动的提供者
	mutex           sync.RWMutex      // 互斥锁
}

// NewProviderManager 创建提供者管理器
func NewProviderManager() *ProviderManager {
	return &ProviderManager{
		providers:       make([]ServiceProvider, 0),
		bootedProviders: make(map[string]bool),
	}
}

// Register 注册服务提供者
func (pm *ProviderManager) Register(provider ServiceProvider) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// 检查提供者是否已注册
	for _, p := range pm.providers {
		if p.Name() == provider.Name() {
			// 提供者已注册，跳过
			return
		}
	}

	// 添加提供者
	pm.providers = append(pm.providers, provider)

	// 按优先级排序
	pm.sortProviders()
}

// RegisterAll 注册多个服务提供者
func (pm *ProviderManager) RegisterAll(providers []ServiceProvider) {
	for _, provider := range providers {
		pm.Register(provider)
	}
}

// RegisterAndBoot 注册并启动服务提供者
func (pm *ProviderManager) RegisterAndBoot(provider ServiceProvider, app *Application) error {
	pm.Register(provider)
	return pm.BootProvider(provider, app)
}

// BootProvider 启动单个服务提供者
func (pm *ProviderManager) BootProvider(provider ServiceProvider, app *Application) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// 检查提供者是否已启动
	if pm.bootedProviders[provider.Name()] {
		return nil
	}

	// 先注册服务
	if err := provider.Register(app); err != nil {
		return fmt.Errorf("注册服务提供者 %s 失败: %w", provider.Name(), err)
	}

	// 再启动服务
	if err := provider.Boot(app); err != nil {
		return fmt.Errorf("启动服务提供者 %s 失败: %w", provider.Name(), err)
	}

	// 标记为已启动
	pm.bootedProviders[provider.Name()] = true

	return nil
}

// BootAll 启动所有注册的服务提供者
func (pm *ProviderManager) BootAll(app *Application) error {
	pm.mutex.RLock()
	providers := make([]ServiceProvider, len(pm.providers))
	copy(providers, pm.providers)
	pm.mutex.RUnlock()

	// 按优先级顺序启动所有提供者
	for _, provider := range providers {
		if err := pm.BootProvider(provider, app); err != nil {
			return err
		}
	}

	return nil
}

// sortProviders 按优先级排序提供者
func (pm *ProviderManager) sortProviders() {
	sort.Slice(pm.providers, func(i, j int) bool {
		return pm.providers[i].Priority() < pm.providers[j].Priority()
	})
}

// GetProviders 获取所有注册的服务提供者
func (pm *ProviderManager) GetProviders() []ServiceProvider {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	providers := make([]ServiceProvider, len(pm.providers))
	copy(providers, pm.providers)
	return providers
}

// IsBooted 检查服务提供者是否已启动
func (pm *ProviderManager) IsBooted(name string) bool {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	return pm.bootedProviders[name]
}

// BaseProvider 基础服务提供者结构体，可作为自定义提供者的基类
type BaseProvider struct {
	name     string
	priority int
}

// NewBaseProvider 创建基础服务提供者
func NewBaseProvider(name string, priority int) *BaseProvider {
	return &BaseProvider{
		name:     name,
		priority: priority,
	}
}

// Name 获取提供者名称
func (bp *BaseProvider) Name() string {
	return bp.name
}

// Priority 获取提供者优先级
func (bp *BaseProvider) Priority() int {
	return bp.priority
}

// Register 注册服务（需要子类重写）
func (bp *BaseProvider) Register(app *Application) error {
	return nil
}

// Boot 启动服务（需要子类重写）
func (bp *BaseProvider) Boot(app *Application) error {
	return nil
}
