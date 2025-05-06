package db

import (
	"log"
	"os"
	"time"

	"gorm.io/gorm"
)

// DbInitializer 数据库初始化器类型
type DbInitializer func([]interface{}) (interface{}, error)

// 存储数据库配置选项
var databaseOptions []interface{}

// 全局变量，用于存储flow包导出的注册函数
var flowRegisterFunc func(DbInitializer)

// 用于初始化数据库的包级函数
func init() {
	// 如果flow包已注册了初始化函数，则调用它
	if flowRegisterFunc != nil {
		flowRegisterFunc(InitializeDatabase)
	}
	// 移除警告日志，减少启动时的噪音
}

// RegisterDatabaseInitializer 注册数据库初始化器到flow框架
// 此函数由flow包调用，用于建立flow包与db包的连接
func RegisterDatabaseInitializer(registerFunc func(DbInitializer)) {
	flowRegisterFunc = registerFunc
	// 立即注册初始化器
	if flowRegisterFunc != nil {
		flowRegisterFunc(InitializeDatabase)
		// 这里添加一个调试级别的日志，可以通过环境变量控制是否显示
		if os.Getenv("FLOW_DB_DEBUG") == "true" {
			log.Println("已注册数据库初始化器")
		}
	}
}

// SetDatabaseOptions 设置数据库配置选项
// 在WithDatabase选项中使用
func SetDatabaseOptions(options []interface{}) {
	databaseOptions = options
}

// SetRegisterFunction 设置注册函数，由flow包调用
// 已被新的RegisterDatabaseInitializer替代，保留以兼容旧代码
func SetRegisterFunction(registerFunc func(DbInitializer)) {
	RegisterDatabaseInitializer(registerFunc)
}

// InitializeDatabase 初始化数据库
func InitializeDatabase(options []interface{}) (interface{}, error) {
	// 如果options为空但databaseOptions不为空，使用databaseOptions
	if len(options) == 0 && len(databaseOptions) > 0 {
		options = databaseOptions
	}

	// 创建数据库管理器
	manager := NewManager()

	// 处理选项
	for _, opt := range options {
		switch o := opt.(type) {
		case Config:
			// 注册默认数据库配置
			manager.Register("default", o)
			manager.SetDefaultConnection("default")
		case ConnectionOption:
			// 应用连接配置选项
			o(manager)
		case func(*Manager):
			// 直接应用函数选项
			o(manager)
		case map[string]interface{}:
			// 处理嵌套的配置结构
			if dbConfig, found := processNestedConfig(o); found {
				if name, ok := dbConfig["default"].(string); ok && name != "" {
					// 使用指定的默认连接名称
					manager.SetDefaultConnection(name)
				}

				// 处理配置中的connections部分
				if connections, ok := o["connections"].(map[string]interface{}); ok {
					for connName, connConfig := range connections {
						if config, ok := createConfigFromMap(connConfig); ok {
							manager.Register(connName, config)
						}
					}
				}
			}
		default:
			// 尝试处理其他类型的配置对象
			if configObj, ok := extractDatabaseConfig(o); ok {
				// 处理从对象中提取的配置
				for name, config := range configObj {
					manager.Register(name, config)
				}
			}
		}
	}

	// 返回默认数据库连接
	db, err := manager.Default()
	if err != nil {
		return nil, err
	}

	// 返回数据库管理器和默认连接
	return &DbProvider{
		Manager: manager,
		DB:      db,
	}, nil
}

// processNestedConfig 处理嵌套的配置结构
func processNestedConfig(config map[string]interface{}) (map[string]interface{}, bool) {
	// 检查是否有database键
	if dbConfig, ok := config["database"].(map[string]interface{}); ok {
		return dbConfig, true
	}

	// 检查是否直接包含数据库配置
	if _, ok := config["default"]; ok {
		if _, ok := config["connections"]; ok {
			return config, true
		}
	}

	return nil, false
}

// createConfigFromMap 从映射创建数据库配置
func createConfigFromMap(configMap interface{}) (Config, bool) {
	config := Config{}

	if cm, ok := configMap.(map[string]interface{}); ok {
		// 设置基本属性
		if driver, ok := cm["driver"].(string); ok {
			config.Driver = driver
		} else {
			return config, false
		}

		// 设置连接信息
		if host, ok := cm["host"].(string); ok {
			config.Host = host
		}

		if port, ok := cm["port"].(int); ok {
			config.Port = port
		}

		if database, ok := cm["database"].(string); ok {
			config.Database = database
		}

		if username, ok := cm["username"].(string); ok {
			config.Username = username
		}

		if password, ok := cm["password"].(string); ok {
			config.Password = password
		}

		// 设置其他连接参数
		if charset, ok := cm["charset"].(string); ok {
			config.Charset = charset
		}

		if sslmode, ok := cm["sslmode"].(string); ok {
			config.SSLMode = sslmode
		}

		if timezone, ok := cm["timezone"].(string); ok {
			config.TimeZone = timezone
		}

		// 设置默认值
		config.MaxIdleConns = 10
		config.MaxOpenConns = 100
		config.ConnMaxLifetime = time.Hour

		// 从配置中获取连接池设置
		if maxIdle, ok := cm["max_idle_conns"].(int); ok {
			config.MaxIdleConns = maxIdle
		}

		if maxOpen, ok := cm["max_open_conns"].(int); ok {
			config.MaxOpenConns = maxOpen
		}

		return config, true
	}

	return config, false
}

// extractDatabaseConfig 从各种类型的对象中提取数据库配置
func extractDatabaseConfig(obj interface{}) (map[string]Config, bool) {
	// 使用反射来尝试获取对象的数据库配置字段
	// 这是一个简化版实现，实际可能需要更复杂的反射逻辑

	// 尝试检查常见的配置对象方法
	if configProvider, ok := obj.(interface{ GetDatabaseConfig() map[string]Config }); ok {
		return configProvider.GetDatabaseConfig(), true
	}

	// 尝试通过Get方法获取数据库配置
	if getter, ok := obj.(interface{ Get(string) interface{} }); ok {
		if dbConfig := getter.Get("database"); dbConfig != nil {
			if configMap, ok := dbConfig.(map[string]Config); ok {
				return configMap, true
			}
		}
	}

	return nil, false
}

// DbProvider 数据库服务提供者(重命名以避免冲突)
type DbProvider struct {
	Manager *Manager
	DB      *gorm.DB
}

// Close 关闭数据库连接
func (p *DbProvider) Close() error {
	if p.Manager != nil {
		return p.Manager.Close()
	}
	return nil
}
