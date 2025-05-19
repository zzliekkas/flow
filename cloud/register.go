package cloud

import (
	"github.com/zzliekkas/flow/cloud/providers"
	"go.uber.org/dig"
)

// Register 注册云服务到依赖注入容器
func Register(container *dig.Container) error {
	// 注册存储服务工厂
	if err := container.Provide(NewStorageProvider); err != nil {
		return err
	}

	// 注册快递100服务Provider
	if err := container.Provide(func() *providers.Kd100Provider {
		return &providers.Kd100Provider{}
	}); err != nil {
		return err
	}

	return nil
}
