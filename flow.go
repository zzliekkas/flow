package flow

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zzliekkas/flow/config"
	"github.com/zzliekkas/flow/db"
	"go.uber.org/dig"
	"gorm.io/gorm"
)

// 全局常量
const (
	// 版本信息
	Version = "1.0.4"
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
	container     *dig.Container
	config        *Config
	dbInitialized bool // 数据库是否已初始化

	// 生命周期钩子
	shutdownHooks []func()
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

		// 初始化和加载配置
		configManager, err := loadConfig(configPath)
		if err != nil {
			// 记录错误但不中断
			fmt.Printf("加载配置文件失败: %v\n", err)
			// 创建一个空的配置管理器继续使用
			configManager = config.NewConfig()
			configManager.Set("app.name", "flow")
			configManager.Set("app.version", Version)
			configManager.Set("app.mode", e.config.Mode)
			configManager.Set("app.log_level", e.config.LogLevel)
		}

		// 注册到依赖注入容器
		e.Provide(func() *config.Config {
			return configManager
		})

		// 为兼容性提供Manager类型别名
		e.Provide(func(cfg *config.Config) *config.Manager {
			return cfg
		})

		// 应用配置到框架
		applyConfigToEngine(e, configManager)
	}
}

// loadConfig 加载配置文件
func loadConfig(configPath string) (*config.Config, error) {
	// 判断是文件路径还是目录
	fi, err := os.Stat(configPath)

	var dirPath, configName string
	if err == nil && fi.IsDir() {
		// 如果是目录，使用目录和默认文件名"app"
		dirPath = configPath
		configName = "app"
		fmt.Printf("配置目录: %s, 配置文件: %s\n", dirPath, configName)
	} else {
		// 如果是文件路径或不存在，拆分为目录和文件名
		dirPath = filepath.Dir(configPath)
		baseName := filepath.Base(configPath)
		configName = strings.TrimSuffix(baseName, filepath.Ext(baseName))
		fmt.Printf("配置目录: %s, 配置文件: %s\n", dirPath, configName)
	}

	// 创建配置实例
	cfg := config.NewConfig(
		config.WithConfigPath(dirPath),
		config.WithConfigName(configName),
	)

	// 尝试加载配置
	err = cfg.Load()
	if err != nil {
		// 如果加载失败，记录警告但继续使用空配置
		fmt.Printf("警告: 加载配置文件失败: %v\n", err)
		fmt.Println("将使用默认配置继续运行")

		// 初始化一些基本配置值
		cfg.Set("app.name", "flow")
		cfg.Set("app.version", Version)
		cfg.Set("app.mode", "debug")
		cfg.Set("app.log_level", "info")

		// 不返回错误，让应用继续运行
		return cfg, nil
	}

	return cfg, nil
}

// applyConfigToEngine 将配置应用到引擎
func applyConfigToEngine(e *Engine, cfg *config.Config) {
	if cfg == nil {
		// 如果配置为nil，不做任何操作
		fmt.Println("警告: 配置为空，跳过应用配置到引擎")
		return
	}

	// 应用模式设置
	if mode := cfg.GetString("app.mode"); mode != "" {
		e.config.Mode = mode
		gin.SetMode(mapGinMode(mode))
	}

	// 应用日志级别设置
	if logLevel := cfg.GetString("app.log_level"); logLevel != "" {
		e.config.LogLevel = logLevel
	}

	// 应用其它配置
	if templates := cfg.GetString("app.templates"); templates != "" {
		e.LoadHTMLGlob(templates)
	}

	// 应用静态文件配置
	if staticPath := cfg.GetString("app.static.path"); staticPath != "" {
		urlPath := cfg.GetString("app.static.url")
		if urlPath == "" {
			urlPath = "/static"
		}
		e.Static(urlPath, staticPath)
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

		// 配置日志级别
		configureLogLevel(level)
	}
}

// configureLogLevel 配置日志级别
func configureLogLevel(level string) {
	// 依据级别配置日志
	switch strings.ToLower(level) {
	case "debug":
		// 设置为详细日志模式
		fmt.Println("日志级别设置为: DEBUG")
	case "info":
		// 设置为信息日志模式
		fmt.Println("日志级别设置为: INFO")
	case "warn", "warning":
		// 设置为警告日志模式
		fmt.Println("日志级别设置为: WARN")
	case "error":
		// 设置为错误日志模式
		fmt.Println("日志级别设置为: ERROR")
	default:
		// 默认为INFO级别
		fmt.Println("未知日志级别，使用默认: INFO")
	}
}

// WithDatabase 返回一个配置数据库的选项
func WithDatabase(options ...interface{}) Option {
	return func(e *Engine) {
		// 线程安全地存储数据库选项
		dbOptionsMutex.Lock()
		if len(options) > 0 {
			databaseOptions = make([]interface{}, len(options))
			copy(options, databaseOptions)
		}
		// 确保db包能获取选项
		db.SetDatabaseOptions(databaseOptions)
		dbOptionsMutex.Unlock()

		// 注册数据库初始化提供者
		e.Provide(initDatabaseProvider)

		// 注册关闭钩子
		e.OnShutdown(func() {
			// 尝试关闭数据库连接
			e.Invoke(func(provider interface{}) {
				if dbProvider, ok := provider.(*db.DbProvider); ok {
					if err := dbProvider.Close(); err != nil {
						fmt.Printf("关闭数据库连接时出错: %v\n", err)
					} else {
						log.Println("数据库连接已安全关闭")
					}
				}
			})
		})
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

// WithConfigWatcher 返回一个监听配置变更的选项
func WithConfigWatcher(callback func()) Option {
	return func(e *Engine) {
		// 检查是否有配置管理器
		e.Invoke(func(cfg *config.Config) {
			// 注册配置变更回调
			cfg.OnChange(callback)
		})
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
	fmt.Printf("Flow Framework for Go %s\n", Version)

	return e.Engine.Run(addr...)
}

// OnShutdown 注册应用关闭时的钩子函数
func (e *Engine) OnShutdown(fn func()) {
	e.shutdownHooks = append(e.shutdownHooks, fn)
}

// Shutdown 优雅关闭HTTP服务器
func (e *Engine) Shutdown(ctx context.Context) error {
	// 调用所有关闭钩子
	for _, hook := range e.shutdownHooks {
		hook()
	}

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
func (c *Context) DB() *gorm.DB {
	var dbProvider *db.DbProvider
	err := c.engine.Invoke(func(p *db.DbProvider) {
		dbProvider = p
	})

	if err != nil || dbProvider == nil {
		return nil
	}

	return dbProvider.DB
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
	return strconv.Atoi(value)
}

// UintParam 将字符串转换为无符号整数
func (c *Context) UintParam(value string) (uint, error) {
	val, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(val), nil
}

// Config 获取配置实例
func (c *Context) Config() *config.Config {
	var cfg *config.Config
	err := c.engine.Invoke(func(c *config.Config) {
		cfg = c
	})

	if err != nil || cfg == nil {
		// 如果没有注册配置，返回一个具备安全默认值的空配置
		cfg = config.NewConfig()
		// 手动初始化 viper，确保不会发生空指针异常
		cfg.Set("app.name", "flow")
		cfg.Set("app.version", Version)
		cfg.Set("app.mode", c.engine.config.Mode)
		cfg.Set("app.log_level", c.engine.config.LogLevel)
	}

	return cfg
}

// ConfigValue 获取指定键的配置值
func (c *Context) ConfigValue(key string) interface{} {
	cfg := c.Config()
	if cfg == nil {
		return nil
	}
	return cfg.Get(key)
}

// ConfigString 获取字符串配置值
func (c *Context) ConfigString(key string, defaultValue string) string {
	cfg := c.Config()
	if cfg == nil {
		return defaultValue
	}

	value := cfg.GetString(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// ConfigInt 获取整数配置值
func (c *Context) ConfigInt(key string, defaultValue int) int {
	cfg := c.Config()
	if cfg == nil {
		return defaultValue
	}

	return cfg.GetInt(key)
}

// ConfigBool 获取布尔配置值
func (c *Context) ConfigBool(key string, defaultValue bool) bool {
	cfg := c.Config()
	if cfg == nil {
		return defaultValue
	}

	return cfg.GetBool(key)
}

// RegisterDatabaseInitializer 注册数据库初始化器
func (e *Engine) RegisterDatabaseInitializer() {
	if !e.dbInitialized {
		db.RegisterDatabaseInitializer(func(initializer db.DbInitializer) {
			// 初始化数据库
			dbProvider, err := initializer(nil)
			if err != nil {
				log.Printf("数据库初始化失败: %v", err)
				return
			}

			// 将数据库实例注入容器
			err = e.container.Provide(func() interface{} {
				return dbProvider
			})
			if err != nil {
				log.Printf("注入数据库实例失败: %v", err)
				return
			}
			e.dbInitialized = true
		})
	}
}

// DB 返回默认数据库连接
func (e *Engine) DB() (*db.DbProvider, bool) {
	var provider *db.DbProvider
	err := e.container.Invoke(func(p *db.DbProvider) {
		provider = p
	})
	if err != nil || provider == nil {
		return nil, false
	}
	return provider, true
}

// Container 返回依赖注入容器
func (e *Engine) Container() *dig.Container {
	return e.container
}

// WaitForTermination 等待终止信号
func (e *Engine) WaitForTermination() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Printf("服务器关闭失败: %v", err)
	}
}
