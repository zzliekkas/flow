package db

import (
	"errors"

	"github.com/zzliekkas/flow/config"
	"github.com/zzliekkas/flow/di"
	"gorm.io/gorm"
)

// DatabaseProvider 数据库服务提供者
type DatabaseProvider struct {
	Name      string
	Priority  int
	container *di.Container
}

// NewDatabaseProvider 创建数据库服务提供者
func NewDatabaseProvider(container *di.Container) *DatabaseProvider {
	return &DatabaseProvider{
		Name:      "database",
		Priority:  50, // 设置优先级为50
		container: container,
	}
}

// Register 注册数据库服务
func (p *DatabaseProvider) Register(app interface{}) error {
	// 注册数据库连接管理器
	err := p.container.Provide(func(configManager *config.ConfigManager) (*Manager, error) {
		manager := NewManager()

		// 从配置中加载数据库配置
		if err := manager.FromConfig(configManager); err != nil {
			return nil, err
		}

		return manager, nil
	})

	if err != nil {
		return err
	}

	// 注册默认数据库连接
	err = p.container.Provide(func(manager *Manager) (*gorm.DB, error) {
		return manager.Default()
	})

	if err != nil {
		return err
	}

	// 注册种子数据管理器
	err = p.container.Provide(func(dbManager *Manager) (*SeederManager, error) {
		// 获取默认数据库连接
		db, err := dbManager.Default()
		if err != nil {
			return nil, err
		}
		return NewSeederManager(db), nil
	})
	if err != nil {
		return err
	}

	return nil
}

// Boot 启动数据库服务
func (p *DatabaseProvider) Boot(app interface{}) error {
	// 由于移除了直接依赖，这里暂时保留但不再直接注册关闭钩子
	// 关闭逻辑移至Container的销毁过程中

	return nil
}

// GetProviderName 返回提供者名称
func (p *DatabaseProvider) GetProviderName() string {
	return p.Name
}

// GetPriority 返回提供者优先级
func (p *DatabaseProvider) GetPriority() int {
	return p.Priority
}

// GetInstance 获取数据库管理器实例
func GetInstance(engine interface{}) (*Manager, error) {
	// 检查engine是否实现了Invoke方法
	invoker, ok := engine.(interface {
		Invoke(interface{}) error
	})

	if !ok {
		return nil, errors.New("引擎不支持依赖注入")
	}

	var manager *Manager
	if err := invoker.Invoke(func(m *Manager) {
		manager = m
	}); err != nil {
		return nil, err
	}

	return manager, nil
}

// GetConnection 获取指定名称的数据库连接
func GetConnection(engine interface{}, name string) (*gorm.DB, error) {
	manager, err := GetInstance(engine)
	if err != nil {
		return nil, err
	}

	return manager.Connection(name)
}

// GetDefaultConnection 获取默认数据库连接
func GetDefaultConnection(engine interface{}) (*gorm.DB, error) {
	manager, err := GetInstance(engine)
	if err != nil {
		return nil, err
	}

	return manager.Default()
}

// GetSeederManager 获取种子数据管理器
func GetSeederManager(engine interface{}) (*SeederManager, error) {
	if engine == nil {
		return nil, errors.New("引擎实例不能为空")
	}

	// 检查engine是否实现了Invoke方法
	invoker, ok := engine.(interface {
		Invoke(interface{}) error
	})

	if !ok {
		return nil, errors.New("引擎不支持依赖注入")
	}

	var seederManager *SeederManager
	err := invoker.Invoke(func(sm *SeederManager) {
		seederManager = sm
	})
	if err != nil {
		return nil, err
	}

	return seederManager, nil
}
