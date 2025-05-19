package providers

import (
	"errors"

	"github.com/zzliekkas/flow/config"
)

type Kd100Provider struct{}

func (p *Kd100Provider) Name() string  { return "kd100" }
func (p *Kd100Provider) Priority() int { return 100 }

func (p *Kd100Provider) Register(application interface{}) error {
	// 通过类型断言获取Engine方法
	app, ok := application.(interface{ Engine() interface{} })
	if !ok {
		return errors.New("invalid application type: missing Engine method")
	}
	engine := app.Engine()

	// 读取配置
	var cfg config.Kd100Config
	if err := engine.(interface {
		Invoke(func(*config.Config)) error
	}).Invoke(func(c *config.Config) {
		_ = c.Unmarshal("kd100", &cfg)
	}); err != nil {
		return err
	}
	// 注册服务到容器
	return engine.(interface{ Provide(interface{}) error }).Provide(func() *Kd100Service {
		return NewKd100Service(cfg)
	})
}

func (p *Kd100Provider) Boot(application interface{}) error {
	return nil
}
