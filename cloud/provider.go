package cloud

import (
	"github.com/zzliekkas/flow/app"
)

// CloudServiceProvider 云服务提供者
type CloudServiceProvider struct {
	*app.BaseProvider
}

// NewCloudServiceProvider 创建云服务提供者
func NewCloudServiceProvider() *CloudServiceProvider {
	return &CloudServiceProvider{
		BaseProvider: app.NewBaseProvider("cloud", 50), // 优先级为50
	}
}

// Register 注册云服务到DI容器
func (p *CloudServiceProvider) Register(app *app.Application) error {
	container := app.Engine().Container()

	// 注册云存储服务
	if err := Register(container); err != nil {
		return err
	}

	app.Logger().Info("Cloud services registered")
	return nil
}

// Boot 启动云服务
func (p *CloudServiceProvider) Boot(app *app.Application) error {
	app.Logger().Info("Cloud services booted")
	return nil
}
