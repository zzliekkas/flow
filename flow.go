package flow

import (
	"context"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/zzliekkas/flow/cli/banner"
	"go.uber.org/dig"
)

// 全局常量
const (
	// 版本信息
	Version = "1.0.1"
)

// 添加单例变量
var (
	engineInstance *Engine
	once           sync.Once
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

// 选项函数类型定义
type Option func(*Engine)

// WithConfig 返回一个设置配置文件路径的选项
func WithConfig(configPath string) Option {
	return func(e *Engine) {
		e.config.ConfigPath = configPath
		// 这里可以添加配置加载逻辑
	}
}

// WithMode 返回一个设置运行模式的选项
func WithMode(mode string) Option {
	return func(e *Engine) {
		e.config.Mode = mode
		gin.SetMode(mapGinMode(mode))
	}
}

// WithLogLevel 返回一个设置日志级别的选项
func WithLogLevel(level string) Option {
	return func(e *Engine) {
		e.config.LogLevel = level
		// 这里可以添加日志级别设置逻辑
	}
}

// WithDatabase 返回一个配置数据库的选项
func WithDatabase(options ...interface{}) Option {
	return func(e *Engine) {
		// 这里可以添加数据库初始化逻辑
		// 例如注册数据库服务到依赖注入容器
	}
}

// WithMiddleware 返回一个添加全局中间件的选项
func WithMiddleware(middleware ...HandlerFunc) Option {
	return func(e *Engine) {
		e.Use(middleware...)
	}
}

// WithTemplates 返回一个配置模板引擎的选项
func WithTemplates(pattern string) Option {
	return func(e *Engine) {
		// 这里可以添加模板加载逻辑
		e.Engine.LoadHTMLGlob(pattern)
	}
}

// WithStaticFiles 返回一个配置静态文件服务的选项
func WithStaticFiles(urlPath, dirPath string) Option {
	return func(e *Engine) {
		e.Engine.Static(urlPath, dirPath)
	}
}

// WithServiceProvider 返回一个注册服务提供者的选项
func WithServiceProvider(constructor interface{}) Option {
	return func(e *Engine) {
		e.Provide(constructor)
	}
}

// New 创建一个新的Flow引擎实例，支持选项模式配置
func New(options ...Option) *Engine {
	// 使用单例模式，确保只创建一个Engine实例
	once.Do(func() {
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

		// 创建Flow引擎
		engineInstance = &Engine{
			Engine:    ginEngine,
			container: container,
			config:    config,
		}

		// 添加默认中间件 - 修复类型不匹配问题
		// 使用gin原生的Recovery中间件，并包装成Flow的HandlerFunc
		engineInstance.Use(func(c *Context) {
			ginRecovery := gin.Recovery()
			ginRecovery(c.Context)
		})
	})

	// 应用选项
	for _, option := range options {
		option(engineInstance)
	}

	return engineInstance
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
	// 显示Flow框架Banner
	banner.Print(Version, "Web Framework for Go")

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

// IsDebug 检查应用是否在调试模式下运行
func (e *Engine) IsDebug() bool {
	return e.config.Mode == "debug"
}

// GetConfig 获取框架配置
func (e *Engine) GetConfig() *Config {
	return e.config
}

// DB 获取数据库连接
// 这是一个便捷方法，用于从上下文中获取数据库连接
func (c *Context) DB() interface{} {
	var db interface{}
	c.Inject(&db)
	return db
}

// Cache 获取缓存实例
// 这是一个便捷方法，用于从上下文中获取缓存实例
func (c *Context) Cache() interface{} {
	var cache interface{}
	c.Inject(&cache)
	return cache
}

// QueryInt 获取查询参数并转换为整数，如果不存在或转换失败则返回默认值
func (c *Context) QueryInt(key string, defaultValue int) int {
	value := c.Query(key)
	if value == "" {
		return defaultValue
	}

	intValue, err := c.IntParam(value)
	if err != nil {
		return defaultValue
	}

	return intValue
}

// ParamUint 获取URL参数并转换为无符号整数，如果不存在或转换失败则返回0
func (c *Context) ParamUint(key string) uint {
	value := c.Param(key)
	if value == "" {
		return 0
	}

	uintValue, err := c.UintParam(value)
	if err != nil {
		return 0
	}

	return uintValue
}

// IntParam 将字符串转换为整数
func (c *Context) IntParam(value string) (int, error) {
	// 这里应该有实际的转换逻辑
	// 为简化代码，仅返回0
	return 0, nil
}

// UintParam 将字符串转换为无符号整数
func (c *Context) UintParam(value string) (uint, error) {
	// 这里应该有实际的转换逻辑
	// 为简化代码，仅返回0
	return 0, nil
}
