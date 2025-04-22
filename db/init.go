package db

import (
	"log"

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
	} else {
		log.Println("警告: 数据库初始化器注册函数未设置")
	}
}

// RegisterDatabaseInitializer 注册数据库初始化器到flow框架
// 此函数由flow包调用，用于建立flow包与db包的连接
func RegisterDatabaseInitializer(registerFunc func(DbInitializer)) {
	flowRegisterFunc = registerFunc
	// 立即注册初始化器
	if flowRegisterFunc != nil {
		flowRegisterFunc(InitializeDatabase)
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
