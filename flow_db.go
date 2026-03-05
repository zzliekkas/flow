package flow

import (
	"fmt"
	"sync"

	"github.com/zzliekkas/flow/v2/db"
)

// 数据库选项存储 - 使用互斥锁保护
var (
	databaseOptions []interface{}
	dbOptionsMutex  sync.Mutex
)

// WithDatabase 配置数据库连接
// 传入各种配置选项，如直接的Config、ConnectionOption函数或配置结构
func (e *Engine) WithDatabase(options ...interface{}) *Engine {
	// 线程安全地存储选项
	dbOptionsMutex.Lock()
	databaseOptions = make([]interface{}, len(options))
	copy(databaseOptions, options)
	dbOptionsMutex.Unlock()

	// 直接调用 db.InitializeDatabase 初始化数据库
	e.Provide(func() (*db.DbProvider, error) {
		dbOptionsMutex.Lock()
		opts := make([]interface{}, len(databaseOptions))
		copy(opts, databaseOptions)
		dbOptionsMutex.Unlock()

		result, err := db.InitializeDatabase(opts)
		if err != nil {
			return nil, fmt.Errorf("数据库初始化失败: %w", err)
		}

		provider, ok := result.(*db.DbProvider)
		if !ok {
			return nil, fmt.Errorf("数据库初始化返回类型错误")
		}

		return provider, nil
	})

	// 注册关闭钩子
	e.OnShutdown(func() {
		e.Invoke(func(provider *db.DbProvider) {
			if err := provider.Close(); err != nil {
				flog.Errorf("关闭数据库连接时出错: %v", err)
			} else {
				flog.Info("数据库连接已安全关闭")
			}
		})
	})

	return e
}
