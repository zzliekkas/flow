package flow

// 导入优先级：
// 1. 标准库
// 2. 内部ginmode包 (确保在gin之前初始化)
// 3. 外部依赖
// 4. 内部包
import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	// 确保在导入gin之前设置好gin模式
	_ "github.com/zzliekkas/flow/v2/ginmode"

	"github.com/gin-gonic/gin"
	"github.com/zzliekkas/flow/v2/config"
	"github.com/zzliekkas/flow/v2/db"
	"github.com/zzliekkas/flow/v2/di"
	"go.uber.org/dig"
)

// 全局常量
const (
	// 版本信息
	Version = "2.0.0"

	FlowBanner = `
[%s] [INFO] 

    ┌─┐┬  ┌─┐┬ ┬
    ├┤ │  │ ││││
    └  ┴─┘└─┘└┴┘

`
)

// defaultEngine 全局默认引擎实例（可选，用于向后兼容）
var defaultEngine *Engine

// H is a shortcut for map[string]interface{}
type H map[string]interface{}

// Engine 是Flow框架的主结构体，封装了Gin引擎和依赖注入容器
type Engine struct {
	*gin.Engine
	container     *di.Container
	config        *Config
	server        *http.Server // HTTP服务器实例，用于优雅关闭
	dbInitialized bool         // 数据库是否已初始化

	// 生命周期钩子
	startHooks    []hook // 启动钩子（Run之前执行）
	shutdownHooks []hook // 关闭钩子（Shutdown时执行）
}

// hook 带优先级的钩子函数
type hook struct {
	fn       func()
	priority int // 数值越小优先级越高
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
			flog.Warnf("加载配置文件失败: %v", err)
			// 创建一个空的配置管理器继续使用
			configManager = config.NewConfigManager()
			configManager.Set("app.name", "flow")
			configManager.Set("app.version", Version)
			configManager.Set("app.mode", e.config.Mode)
			configManager.Set("app.log_level", e.config.LogLevel)
		}

		// 注册到依赖注入容器
		e.Provide(func() *config.ConfigManager {
			return configManager
		})

		// 为兼容性提供Manager类型别名
		e.Provide(func(cfg *config.ConfigManager) *config.ConfigManager {
			return cfg
		})

		// 应用配置到框架
		applyConfigToEngine(e, configManager)
	}
}

// loadConfig 加载配置文件
func loadConfig(configPath string) (*config.ConfigManager, error) {
	// 判断是文件路径还是目录
	fi, err := os.Stat(configPath)

	var dirPath, configName string
	if err == nil && fi.IsDir() {
		// 如果是目录，使用目录和默认文件名"app"
		dirPath = configPath
		configName = "app"
	} else {
		// 如果是文件路径或不存在，拆分为目录和文件名
		dirPath = filepath.Dir(configPath)
		baseName := filepath.Base(configPath)
		configName = strings.TrimSuffix(baseName, filepath.Ext(baseName))
	}

	// 创建配置实例
	cfg := config.NewConfigManager(
		config.WithConfigPath(dirPath),
		config.WithConfigName(configName),
	)

	// 尝试加载配置
	err = cfg.Load()
	if err != nil {
		// 如果加载失败，使用空配置
		flog.Debugf("加载配置文件失败: %v，将使用默认配置继续运行", err)

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
func applyConfigToEngine(e *Engine, cfg *config.ConfigManager) {
	if cfg == nil {
		// 如果配置为nil，不做任何操作
		flog.Warn("配置为空，跳过应用配置到引擎")
		return
	}

	// 应用模式设置
	if mode := cfg.GetString("app.mode"); mode != "" {
		// 使用WithMode选项设置模式，确保一致性
		WithMode(mode)(e)
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
		// 设置flow模式
		e.config.Mode = mode

		// 同步设置gin模式，保持一致性
		var ginMode string
		switch strings.ToLower(mode) {
		case "release", "production":
			ginMode = "release"
		case "test":
			ginMode = "test"
		case "debug", "development":
			ginMode = "debug"
		default:
			// 未知模式默认为debug
			ginMode = "debug"
		}

		// 设置GIN_MODE环境变量和gin模式
		os.Setenv("GIN_MODE", ginMode)
		gin.SetMode(ginMode)

		// 同时设置FLOW_MODE环境变量，保持一致性
		os.Setenv("FLOW_MODE", mode)
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
		flog.Debug("日志级别设置为: DEBUG")
	case "info":
		flog.Debug("日志级别设置为: INFO")
	case "warn", "warning":
		flog.Debug("日志级别设置为: WARN")
	case "error":
		flog.Debug("日志级别设置为: ERROR")
	default:
		flog.Debug("未知日志级别，使用默认: INFO")
	}
}

// WithDatabase 返回一个配置数据库的选项
func WithDatabase(options ...interface{}) Option {
	return func(e *Engine) {
		e.WithDatabase(options...)
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
		e.Invoke(func(cfg *config.ConfigManager) {
			// 注册配置变更回调
			cfg.OnChange(callback)
		})
	}
}

// New 创建一个新的Flow引擎实例，支持选项模式配置
func New(options ...Option) *Engine {
	// 创建依赖注入容器
	container := di.New()

	// 默认配置
	defaultMode := "debug"
	// 检查环境变量中的配置
	if flowMode := os.Getenv("FLOW_MODE"); flowMode != "" {
		defaultMode = flowMode
	}
	// 保持与GIN_MODE一致性
	if ginMode := os.Getenv("GIN_MODE"); ginMode != "" {
		switch ginMode {
		case "release":
			defaultMode = "release"
		case "test":
			defaultMode = "test"
		case "debug":
			defaultMode = "debug"
		}
	}

	cfg := &Config{
		Mode:       defaultMode,
		JSONLib:    "default",
		LogLevel:   "info",
		ConfigPath: "./config",
	}

	// 创建gin引擎
	ginEngine := gin.New()

	// 创建Flow引擎
	e := &Engine{
		Engine:    ginEngine,
		container: container,
		config:    cfg,
	}

	// 添加默认中间件
	e.Use(func(c *Context) {
		ginRecovery := gin.Recovery()
		ginRecovery(c.Context)
	})

	// 应用选项
	for _, option := range options {
		option(e)
	}

	// 设置为默认引擎（首次创建的实例）
	if defaultEngine == nil {
		defaultEngine = e
	}

	return e
}

// Default 返回默认的全局引擎实例
// 如果尚未创建，会自动创建一个默认配置的实例
func Default() *Engine {
	if defaultEngine == nil {
		defaultEngine = New()
	}
	return defaultEngine
}

// Provide 向依赖注入容器注册服务
func (e *Engine) Provide(constructor interface{}) error {
	return e.container.Provide(constructor)
}

// Invoke 从依赖注入容器获取服务
func (e *Engine) Invoke(function interface{}) error {
	return e.container.Invoke(function)
}

// DI 返回依赖注入容器，提供完整的DI操作API（ProvideNamed, ProvideValue, Extract等）
func (e *Engine) DI() *di.Container {
	return e.container
}

// IsDebug 检查应用是否在调试模式下运行
func (e *Engine) IsDebug() bool {
	return e.config.Mode == "debug"
}

// GetConfig 获取框架配置
func (e *Engine) GetConfig() *Config {
	return e.config
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

// Container 返回底层的dig依赖注入容器（向后兼容）
// 推荐使用 DI() 方法获取增强的容器API
func (e *Engine) Container() *dig.Container {
	return e.container.Dig()
}
