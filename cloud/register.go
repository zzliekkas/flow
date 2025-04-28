package cloud

import (
	"go.uber.org/dig"
)

// Register 注册云服务到依赖注入容器
func Register(container *dig.Container) error {
	// 注册存储服务工厂
	if err := container.Provide(NewStorageProvider); err != nil {
		return err
	}

	return nil
}
