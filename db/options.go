package db

import "time"

// ConnectionOption 用于配置数据库连接管理器的函数选项
type ConnectionOption func(*Manager)

// WithConnection 创建一个自定义连接的选项
func WithConnection(name string, config Config) ConnectionOption {
	return func(m *Manager) {
		m.Register(name, config)
	}
}

// WithDefaultConnection 设置默认连接
func WithDefaultConnection(name string) ConnectionOption {
	return func(m *Manager) {
		m.SetDefaultConnection(name)
	}
}

// WithMaxOpenConns 设置最大打开连接数
func WithMaxOpenConns(name string, maxOpen int) ConnectionOption {
	return func(m *Manager) {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		if config, exists := m.configs[name]; exists {
			config.MaxOpenConns = maxOpen
			m.configs[name] = config
		}
	}
}

// WithMaxIdleConns 设置最大空闲连接数
func WithMaxIdleConns(name string, maxIdle int) ConnectionOption {
	return func(m *Manager) {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		if config, exists := m.configs[name]; exists {
			config.MaxIdleConns = maxIdle
			m.configs[name] = config
		}
	}
}

// WithConnMaxLifetime 设置连接最大生存时间
func WithConnMaxLifetime(name string, lifetime time.Duration) ConnectionOption {
	return func(m *Manager) {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		if config, exists := m.configs[name]; exists {
			config.ConnMaxLifetime = lifetime
			m.configs[name] = config
		}
	}
}

// WithAutoMigrate 自动迁移模型
func WithAutoMigrate(models ...interface{}) ConnectionOption {
	return func(m *Manager) {
		db, err := m.Default()
		if err != nil {
			return
		}

		db.AutoMigrate(models...)
	}
}

// WithDebug 启用GORM调试模式
func WithDebug(enable bool) ConnectionOption {
	return func(m *Manager) {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		for name, config := range m.configs {
			if enable {
				config.LogLevel = 4 // gorm.Config.Logger.LogLevel
			} else {
				config.LogLevel = 1 // gorm.Config.Logger.Silent
			}
			m.configs[name] = config
		}
	}
}
