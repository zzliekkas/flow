package app

import (
	"sync"
)

var (
	// 全局应用实例
	instance *Application
	// 确保线程安全
	once sync.Once
)

// SetApplication 设置全局应用实例
func SetApplication(app *Application) {
	once.Do(func() {
		instance = app
	})
}

// GetApplication 获取全局应用实例
func GetApplication() *Application {
	return instance
}

// HasApplication 检查是否已设置全局应用实例
func HasApplication() bool {
	return instance != nil
}
