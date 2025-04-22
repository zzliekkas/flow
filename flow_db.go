package flow

import (
	"fmt"
	"log"

	"github.com/zzliekkas/flow/db"
)

// 初始化时执行
func init() {
	// 向db包注册初始化函数，用于建立双向通信
	db.SetRegisterFunction(registerDatabaseInitializer)
}

// 数据库选项存储
var databaseOptions []interface{}

// 数据库初始化器，由db包设置
var databaseInitializer db.DbInitializer

// registerDatabaseInitializer 注册数据库初始化器函数
func registerDatabaseInitializer(initializer db.DbInitializer) {
	databaseInitializer = initializer
	log.Println("已注册数据库初始化器")
}

// 包装函数，传递给WithDatabase选项
func initDatabaseProvider() (interface{}, error) {
	if databaseInitializer == nil {
		return nil, fmt.Errorf("数据库初始化器未注册")
	}

	// 调用db包中的初始化器函数
	return databaseInitializer(databaseOptions)
}

// 现在，我们将删除重复声明的方法，并从flow.go文件更新它们的实现
// 在flow.go中更新以下方法的实现：
// 1. Context.DB() - 使用db.DbProvider类型
// 2. Context.IntParam() - 使用strconv.Atoi
// 3. Context.UintParam() - 使用strconv.ParseUint
