package db

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"regexp"
	"time"

	"gorm.io/gorm"
)

// DbInitializer 数据库初始化器类型
type DbInitializer func([]interface{}) (interface{}, error)

// 存储数据库配置选项
var databaseOptions []interface{}

// 全局变量，用于存储flow包导出的注册函数
var flowRegisterFunc func(DbInitializer)

// 为测试启用的标志
var skipDatabaseConnection = false

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
				if connections, ok := dbConfig["connections"].(map[string]interface{}); ok {
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

	// 当skipDatabaseConnection为true时，跳过实际连接数据库
	// 这主要用于测试
	if skipDatabaseConnection {
		// 只进行初始化，不连接数据库
		return &DbProvider{
			Manager: manager,
			DB:      nil,
		}, nil
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

// 处理嵌套配置结构
// 返回数据库配置部分和是否找到配置的标志
func processNestedConfig(config map[string]interface{}) (map[string]interface{}, bool) {
	// 直接检查是否为数据库配置
	if _, hasConn := config["connections"]; hasConn {
		log.Println("[DB] 使用嵌套格式的数据库配置")
		return config, true
	}

	// 查找database配置部分
	if db, ok := config["database"].(map[string]interface{}); ok {
		log.Println("[DB] 找到database配置部分")
		return db, true
	}

	log.Println("[DB] 使用平面配置格式")
	return config, false
}

// 从映射创建配置对象
func createConfigFromMap(configMap interface{}) (Config, bool) {
	m, ok := configMap.(map[string]interface{})
	if !ok {
		log.Println("[DB] 配置格式错误: 期待map[string]interface{}, 得到", reflect.TypeOf(configMap))
		return Config{}, false
	}

	driver, ok := m["driver"].(string)
	if !ok || driver == "" {
		log.Println("[DB] 配置错误: 缺少driver字段或非字符串类型")
		return Config{}, false
	}

	// 创建基本配置
	config := Config{
		Driver: driver,
	}

	// 设置主要连接参数
	if host, ok := m["host"].(string); ok {
		config.Host = host
	}

	if port, ok := m["port"].(int); ok {
		config.Port = port
	} else if portFloat, ok := m["port"].(float64); ok {
		config.Port = int(portFloat)
	}

	if database, ok := m["database"].(string); ok {
		config.Database = database
	}

	if username, ok := m["username"].(string); ok {
		config.Username = username
	}

	if password, ok := m["password"].(string); ok {
		config.Password = password
	}

	// 设置其他连接参数
	if charset, ok := m["charset"].(string); ok {
		config.Charset = charset
	}

	if sslmode, ok := m["sslmode"].(string); ok {
		config.SSLMode = sslmode
	}

	if timezone, ok := m["timezone"].(string); ok {
		config.TimeZone = timezone
	}

	// 设置连接池参数
	if maxIdle, ok := m["max_idle_conns"].(int); ok {
		config.MaxIdleConns = maxIdle
	} else if maxIdleFloat, ok := m["max_idle_conns"].(float64); ok {
		config.MaxIdleConns = int(maxIdleFloat)
	}

	if maxOpen, ok := m["max_open_conns"].(int); ok {
		config.MaxOpenConns = maxOpen
	} else if maxOpenFloat, ok := m["max_open_conns"].(float64); ok {
		config.MaxOpenConns = int(maxOpenFloat)
	}

	if lifetime, ok := m["conn_max_lifetime"].(int); ok {
		config.ConnMaxLifetime = time.Duration(lifetime) * time.Second
	} else if lifetimeFloat, ok := m["conn_max_lifetime"].(float64); ok {
		config.ConnMaxLifetime = time.Duration(lifetimeFloat) * time.Second
	}

	// 兼容DSN参数
	if dsn, ok := m["dsn"].(string); ok && dsn != "" {
		// 如果提供了DSN，使用它代替拆分的连接参数
		// 这处理直接传递完整DSN的情况
		switch driver {
		case "mysql", "postgres", "sqlite3", "sqlserver":
			log.Printf("[DB] 使用提供的DSN连接%s数据库\n", driver)
		default:
			log.Printf("[DB] 警告: 未知的数据库驱动 %s，可能不支持\n", driver)
		}
	} else {
		// 如果没有提供DSN，则根据驱动类型构建DSN
		switch driver {
		case "mysql":
			if config.Host != "" && config.Database != "" {
				protocol := "tcp"
				port := config.Port
				if port == 0 {
					port = 3306
				}
				charset := config.Charset
				if charset == "" {
					charset = "utf8mb4"
				}
				// 构建MySQL DSN
				dsn := fmt.Sprintf("%s:%s@%s(%s:%d)/%s",
					config.Username, config.Password, protocol, config.Host, port, config.Database)
				if charset != "" {
					dsn += "?charset=" + charset
				}
				log.Printf("[DB] 生成MySQL DSN: %s\n", maskDSN(dsn))
			}
		case "postgres":
			if config.Host != "" {
				// 构建PostgreSQL连接字符串
				sslMode := config.SSLMode
				if sslMode == "" {
					sslMode = "disable"
				}
				port := config.Port
				if port == 0 {
					port = 5432
				}
				// 构建PostgreSQL DSN
				dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
					config.Host, port, config.Username, config.Password, config.Database, sslMode)
				log.Printf("[DB] 生成PostgreSQL DSN: %s\n", maskDSN(dsn))
			}
		case "sqlite3":
			if config.Database != "" {
				// 构建SQLite连接字符串
				dsn := config.Database
				log.Printf("[DB] 生成SQLite DSN: %s\n", dsn)
			}
		}
	}

	log.Printf("[DB] 成功创建数据库配置: driver=%s\n", config.Driver)
	return config, true
}

// maskDSN 掩盖DSN中的敏感信息（如密码）
func maskDSN(dsn string) string {
	if dsn == "" {
		return ""
	}

	// 处理MySQL格式的DSN (username:password@tcp(...))
	mysqlRegex := regexp.MustCompile(`([^:@]+):([^@]+)@`)
	maskedDsn := mysqlRegex.ReplaceAllString(dsn, "$1:********@")

	// 处理键值对格式的DSN
	passwordRegex := regexp.MustCompile(`(password)=([^;& ]+)`)
	maskedDsn = passwordRegex.ReplaceAllString(maskedDsn, "$1=********")

	return maskedDsn
}

// 从对象中提取数据库配置
func extractDatabaseConfig(obj interface{}) (map[string]Config, bool) {
	val := reflect.ValueOf(obj)

	// 检查对象类型
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil, false
	}

	typ := val.Type()
	configs := make(map[string]Config)
	found := false

	// 查找所有标记了db标签的字段
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// 查找db标签
		tag := fieldType.Tag.Get("db")
		if tag == "" {
			continue
		}

		// 检查字段是否为Config类型
		if config, ok := field.Interface().(Config); ok {
			configs[tag] = config
			found = true
			log.Printf("[DB] 从%s类型中提取到数据库配置: %s\n", typ.Name(), tag)
		}
	}

	return configs, found
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
