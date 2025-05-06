package flow

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/zzliekkas/flow/db"
)

// 初始化时执行
func init() {
	// 向db包注册初始化函数，用于建立双向通信
	db.SetRegisterFunction(registerDatabaseInitializer)
}

// 数据库选项存储 - 使用互斥锁保护
var (
	databaseOptions []interface{}
	dbOptionsMutex  sync.Mutex
)

// 数据库初始化器，由db包设置
var databaseInitializer db.DbInitializer

// registerDatabaseInitializer 注册数据库初始化器函数
func registerDatabaseInitializer(initializer db.DbInitializer) {
	databaseInitializer = initializer
	// 只在调试模式下输出日志
	if os.Getenv("FLOW_DB_DEBUG") == "true" {
		log.Println("已注册数据库初始化器")
	}
}

// 包装函数，传递给WithDatabase选项
func initDatabaseProvider() (interface{}, error) {
	if databaseInitializer == nil {
		return nil, fmt.Errorf("数据库初始化器未注册")
	}

	// 线程安全地获取选项
	dbOptionsMutex.Lock()
	options := make([]interface{}, len(databaseOptions))
	copy(options, databaseOptions)
	dbOptionsMutex.Unlock()

	// 调用db包中的初始化器函数
	provider, err := databaseInitializer(options)
	if err != nil {
		return nil, fmt.Errorf("数据库初始化失败: %w", err)
	}

	return provider, nil
}

// WithDatabase 配置数据库连接
// 传入各种配置选项，如直接的Config、ConnectionOption函数或配置结构
func (e *Engine) WithDatabase(options ...interface{}) *Engine {
	// 线程安全地更新选项
	dbOptionsMutex.Lock()
	// 创建新切片并正确复制
	databaseOptions = make([]interface{}, len(options))
	copy(databaseOptions, options) // 正确的复制顺序
	dbOptionsMutex.Unlock()

	// 添加数据库初始化提供者
	e.Provide(initDatabaseProvider)

	// 支持链式调用
	return e
}
