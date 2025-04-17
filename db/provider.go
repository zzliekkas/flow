package db

import (
	"errors"

	"github.com/zzliekkas/flow/app"
	"github.com/zzliekkas/flow/config"
	"github.com/zzliekkas/flow/di"
	"gorm.io/gorm"
)

// DatabaseProvider 数据库服务提供者
type DatabaseProvider struct {
	*app.BaseProvider
	container *di.Container
}

// NewDatabaseProvider 创建数据库服务提供者
func NewDatabaseProvider(container *di.Container) *DatabaseProvider {
	return &DatabaseProvider{
		BaseProvider: app.NewBaseProvider("database", 50), // 设置优先级为50
		container:    container,
	}
}

// Register 注册数据库服务
func (p *DatabaseProvider) Register(application *app.Application) error {
	// 注册数据库连接管理器
	err := p.container.Provide(func(configManager *config.Manager) (*Manager, error) {
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
func (p *DatabaseProvider) Boot(application *app.Application) error {
	// 注册应用关闭钩子，确保在应用关闭时关闭所有数据库连接
	application.OnBeforeShutdown("database.close", func() {
		var manager *Manager
		if err := p.container.Extract(&manager); err == nil && manager != nil {
			_ = manager.Close()
		}
	}, 10)

	return nil
}

// GetInstance 获取数据库管理器实例
func GetInstance(application *app.Application) (*Manager, error) {
	var manager *Manager
	if err := application.Engine().Invoke(func(m *Manager) {
		manager = m
	}); err != nil {
		return nil, err
	}

	return manager, nil
}

// GetConnection 获取指定名称的数据库连接
func GetConnection(application *app.Application, name string) (*gorm.DB, error) {
	manager, err := GetInstance(application)
	if err != nil {
		return nil, err
	}

	return manager.Connection(name)
}

// GetDefaultConnection 获取默认数据库连接
func GetDefaultConnection(application *app.Application) (*gorm.DB, error) {
	manager, err := GetInstance(application)
	if err != nil {
		return nil, err
	}

	return manager.Default()
}

// GetSeederManager 获取种子数据管理器
func GetSeederManager(application *app.Application) (*SeederManager, error) {
	if application == nil {
		return nil, errors.New("应用实例不能为空")
	}

	var seederManager *SeederManager
	err := application.Engine().Invoke(func(sm *SeederManager) {
		seederManager = sm
	})
	if err != nil {
		return nil, err
	}

	return seederManager, nil
}
