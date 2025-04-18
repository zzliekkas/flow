package flow

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/dig"
)

// 全局常量
const (
	// 版本信息
	Version = "0.1.0"
)

// H is a shortcut for map[string]interface{}
type H map[string]interface{}

// Engine 是Flow框架的主结构体，封装了Gin引擎和依赖注入容器
type Engine struct {
	*gin.Engine
	container *dig.Container
	config    *Config
}

// Context 是Flow框架的上下文结构体，扩展了Gin的Context
type Context struct {
	*gin.Context
	engine *Engine
}

// Config 包含框架配置选项
type Config struct {
	Mode       string // 运行模式: debug, release, test
	JSONLib    string // JSON库: default, gojson
	LogLevel   string // 日志级别: debug, info, warn, error
	ConfigPath string // 配置文件路径
}

// HandlerFunc 定义Flow处理函数
type HandlerFunc func(*Context)

// New 创建一个新的Flow引擎实例
func New() *Engine {
	// 创建依赖注入容器
	container := dig.New()

	// 默认配置
	config := &Config{
		Mode:       "debug",
		JSONLib:    "default",
		LogLevel:   "info",
		ConfigPath: "./config",
	}

	// 检查环境变量中的配置
	if mode := os.Getenv("FLOW_MODE"); mode != "" {
		config.Mode = mode
	}

	// 设置gin模式
	gin.SetMode(mapGinMode(config.Mode))

	// 创建gin引擎
	ginEngine := gin.New()

	// 创建并返回Flow引擎
	engine := &Engine{
		Engine:    ginEngine,
		container: container,
		config:    config,
	}

	// 添加默认中间件 - 修复类型不匹配问题
	// 使用gin原生的Recovery中间件，并包装成Flow的HandlerFunc
	engine.Use(func(c *Context) {
		ginRecovery := gin.Recovery()
		ginRecovery(c.Context)
	})

	return engine
}

// mapGinMode 将Flow模式映射到Gin模式
func mapGinMode(mode string) string {
	switch strings.ToLower(mode) {
	case "release":
		return gin.ReleaseMode
	case "test":
		return gin.TestMode
	default:
		return gin.DebugMode
	}
}

// NewContext 创建Flow上下文
func (e *Engine) NewContext(c *gin.Context) *Context {
	return &Context{
		Context: c,
		engine:  e,
	}
}

// Handle 注册处理函数到给定的HTTP方法和路径
func (e *Engine) Handle(httpMethod, relativePath string, handlers ...HandlerFunc) {
	ginHandlers := make([]gin.HandlerFunc, len(handlers))
	for i, handler := range handlers {
		handler := handler // 创建局部变量避免闭包问题
		ginHandlers[i] = func(c *gin.Context) {
			flowContext := e.NewContext(c)
			handler(flowContext)
		}
	}
	e.Engine.Handle(httpMethod, relativePath, ginHandlers...)
}

// GET 是对Handle("GET", path, handlers)的简便方法
func (e *Engine) GET(relativePath string, handlers ...HandlerFunc) {
	e.Handle(http.MethodGet, relativePath, handlers...)
}

// POST 是对Handle("POST", path, handlers)的简便方法
func (e *Engine) POST(relativePath string, handlers ...HandlerFunc) {
	e.Handle(http.MethodPost, relativePath, handlers...)
}

// PUT 是对Handle("PUT", path, handlers)的简便方法
func (e *Engine) PUT(relativePath string, handlers ...HandlerFunc) {
	e.Handle(http.MethodPut, relativePath, handlers...)
}

// DELETE 是对Handle("DELETE", path, handlers)的简便方法
func (e *Engine) DELETE(relativePath string, handlers ...HandlerFunc) {
	e.Handle(http.MethodDelete, relativePath, handlers...)
}

// PATCH 是对Handle("PATCH", path, handlers)的简便方法
func (e *Engine) PATCH(relativePath string, handlers ...HandlerFunc) {
	e.Handle(http.MethodPatch, relativePath, handlers...)
}

// Group 创建一个新的路由组
func (e *Engine) Group(relativePath string, handlers ...HandlerFunc) *RouterGroup {
	ginHandlers := make([]gin.HandlerFunc, len(handlers))
	for i, handler := range handlers {
		handler := handler // 创建局部变量避免闭包问题
		ginHandlers[i] = func(c *gin.Context) {
			flowContext := e.NewContext(c)
			handler(flowContext)
		}
	}

	// 修复无效字段名问题
	ginGroup := e.Engine.Group(relativePath, ginHandlers...)
	return &RouterGroup{
		RouterGroup: *ginGroup,
		engine:      e,
	}
}

// Use 添加全局中间件
func (e *Engine) Use(middleware ...HandlerFunc) *Engine {
	ginMiddlewares := make([]gin.HandlerFunc, len(middleware))
	for i, m := range middleware {
		m := m // 创建局部变量避免闭包问题
		ginMiddlewares[i] = func(c *gin.Context) {
			flowContext := e.NewContext(c)
			m(flowContext)
		}
	}
	e.Engine.Use(ginMiddlewares...)
	return e
}

// Provide 向依赖注入容器注册服务
func (e *Engine) Provide(constructor interface{}) error {
	return e.container.Provide(constructor)
}

// Invoke 从依赖注入容器获取服务
func (e *Engine) Invoke(function interface{}) error {
	return e.container.Invoke(function)
}

// Run 启动HTTP服务器
func (e *Engine) Run(addr ...string) error {
	return e.Engine.Run(addr...)
}

// Shutdown 优雅关闭HTTP服务器
func (e *Engine) Shutdown(ctx context.Context) error {
	// 获取 Gin 底层的 HTTP 服务器
	// 由于 Gin 实际上不暴露内部 HTTP 服务器的引用，我们只能模拟关闭
	// 在真实应用中，需要持有 http.Server 的引用才能调用其 Shutdown 方法
	// 这里做一个简单的占位实现，后续可以扩展
	return nil
}

// Inject 向上下文注入依赖
func (c *Context) Inject(target interface{}) error {
	return c.engine.container.Invoke(func(injected interface{}) {
		*target.(*interface{}) = injected
	})
}

// RouterGroup 是Flow的路由组结构
type RouterGroup struct {
	RouterGroup gin.RouterGroup
	engine      *Engine
}

// Handle 在路由组中注册处理函数
func (g *RouterGroup) Handle(httpMethod, relativePath string, handlers ...HandlerFunc) {
	ginHandlers := make([]gin.HandlerFunc, len(handlers))
	for i, handler := range handlers {
		handler := handler // 创建局部变量避免闭包问题
		ginHandlers[i] = func(c *gin.Context) {
			flowContext := g.engine.NewContext(c)
			handler(flowContext)
		}
	}
	g.RouterGroup.Handle(httpMethod, relativePath, ginHandlers...)
}

// GET 是对Handle("GET", path, handlers)的简便方法
func (g *RouterGroup) GET(relativePath string, handlers ...HandlerFunc) {
	g.Handle(http.MethodGet, relativePath, handlers...)
}

// POST 是对Handle("POST", path, handlers)的简便方法
func (g *RouterGroup) POST(relativePath string, handlers ...HandlerFunc) {
	g.Handle(http.MethodPost, relativePath, handlers...)
}

// PUT 是对Handle("PUT", path, handlers)的简便方法
func (g *RouterGroup) PUT(relativePath string, handlers ...HandlerFunc) {
	g.Handle(http.MethodPut, relativePath, handlers...)
}

// DELETE 是对Handle("DELETE", path, handlers)的简便方法
func (g *RouterGroup) DELETE(relativePath string, handlers ...HandlerFunc) {
	g.Handle(http.MethodDelete, relativePath, handlers...)
}

// PATCH 是对Handle("PATCH", path, handlers)的简便方法
func (g *RouterGroup) PATCH(relativePath string, handlers ...HandlerFunc) {
	g.Handle(http.MethodPatch, relativePath, handlers...)
}

// Group 创建一个子路由组
func (g *RouterGroup) Group(relativePath string, handlers ...HandlerFunc) *RouterGroup {
	ginHandlers := make([]gin.HandlerFunc, len(handlers))
	for i, handler := range handlers {
		handler := handler // 创建局部变量避免闭包问题
		ginHandlers[i] = func(c *gin.Context) {
			flowContext := g.engine.NewContext(c)
			handler(flowContext)
		}
	}

	// 修复无效字段名问题
	ginGroup := g.RouterGroup.Group(relativePath, ginHandlers...)
	return &RouterGroup{
		RouterGroup: *ginGroup,
		engine:      g.engine,
	}
}

// Use 添加路由组中间件
func (g *RouterGroup) Use(middleware ...HandlerFunc) *RouterGroup {
	ginMiddlewares := make([]gin.HandlerFunc, len(middleware))
	for i, m := range middleware {
		m := m // 创建局部变量避免闭包问题
		ginMiddlewares[i] = func(c *gin.Context) {
			flowContext := g.engine.NewContext(c)
			m(flowContext)
		}
	}
	g.RouterGroup.Use(ginMiddlewares...)
	return g
}
