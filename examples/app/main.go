package main

import (
	"log"
	"net/http"
	"time"

	"github.com/zzliekkas/flow"
	"github.com/zzliekkas/flow/app"
	"github.com/zzliekkas/flow/middleware"
)

// 自定义服务提供者
type ExampleProvider struct {
	*app.BaseProvider
}

// 创建新的示例提供者
func NewExampleProvider() *ExampleProvider {
	return &ExampleProvider{
		BaseProvider: app.NewBaseProvider("example_provider", 100),
	}
}

// Register 注册服务
func (p *ExampleProvider) Register(application *app.Application) error {
	application.Logger().Info("注册示例服务...")

	// 向DI容器注册服务
	return application.Engine().Provide(func() *ExampleService {
		return &ExampleService{
			Message: "Hello from Example Service!",
		}
	})
}

// Boot 启动服务
func (p *ExampleProvider) Boot(application *app.Application) error {
	application.Logger().Info("启动示例服务...")
	return nil
}

// 示例服务
type ExampleService struct {
	Message string
}

// 获取服务消息
func (s *ExampleService) GetMessage() string {
	return s.Message
}

func main() {
	// 创建Flow引擎
	flowEngine := flow.New()

	// 添加中间件
	flowEngine.Use(middleware.Logger())
	flowEngine.Use(middleware.Recovery())

	// 创建应用容器
	application := app.New(flowEngine)

	// 注册服务提供者
	application.RegisterProvider(NewExampleProvider())

	// 添加启动钩子
	application.OnAfterStart("setup_routes", func() {
		// 设置路由
		setupRoutes(flowEngine)
	}, 100)

	// 添加关闭钩子
	application.OnBeforeShutdown("cleanup", func() {
		// 在应用关闭前执行清理工作
		application.Logger().Info("执行清理工作...")
		time.Sleep(1 * time.Second) // 模拟清理工作
	}, 100)

	// 启动应用
	log.Println("启动Flow示例应用...")
	if err := application.Run(":8080"); err != nil {
		log.Fatalf("应用启动失败: %v", err)
	}
}

// 设置路由
func setupRoutes(e *flow.Engine) {
	e.GET("/", func(c *flow.Context) {
		c.JSON(http.StatusOK, flow.H{
			"message": "欢迎使用Flow应用容器!",
			"version": flow.Version,
		})
	})

	e.GET("/service", func(c *flow.Context) {
		var service *ExampleService

		// 从DI容器获取服务
		err := c.Inject(&service)
		if err != nil {
			c.JSON(http.StatusInternalServerError, flow.H{
				"error": "服务注入失败: " + err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, flow.H{
			"message": service.GetMessage(),
		})
	})

	e.GET("/shutdown", func(c *flow.Context) {
		c.JSON(http.StatusOK, flow.H{
			"message": "正在关闭应用...",
		})

		// 异步关闭应用
		go func() {
			time.Sleep(2 * time.Second) // 等待响应发送
			app.GetApplication().Shutdown(10 * time.Second)
		}()
	})
}
