package app

import (
	"time"

	"github.com/sirupsen/logrus"
	"github.com/zzliekkas/flow"
)

// Application 是Flow应用容器
type Application struct {
	engine          *flow.Engine      // Flow引擎
	lifecycle       *LifecycleManager // 生命周期管理器
	hooks           *HooksManager     // 钩子管理器
	environment     *Environment      // 环境信息
	providerManager *ProviderManager  // 服务提供者管理器
	logger          *logrus.Logger    // 日志记录器
	bootStartTime   time.Time         // 启动开始时间
}

// New 创建一个新的应用容器
func New(engine *flow.Engine) *Application {
	app := &Application{
		engine:          engine,
		lifecycle:       NewLifecycleManager(engine),
		hooks:           NewHooksManager(),
		environment:     NewEnvironment(),
		providerManager: NewProviderManager(),
		logger:          logrus.New(),
		bootStartTime:   time.Now(),
	}

	// 初始化应用
	app.initialize()

	// 设置为全局应用实例
	SetApplication(app)

	return app
}

// initialize 初始化应用
func (a *Application) initialize() {
	// 设置日志格式
	a.logger.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	})

	// 根据环境设置日志级别
	if a.environment.Debug {
		a.logger.SetLevel(logrus.DebugLevel)
	} else {
		a.logger.SetLevel(logrus.InfoLevel)
	}

	// 注册默认钩子
	a.registerDefaultHooks()
}

// registerDefaultHooks 注册默认钩子
func (a *Application) registerDefaultHooks() {
	// 启动前钩子 - 打印环境信息
	a.hooks.RegisterBeforeStart("print_environment", func() {
		a.logger.Info("应用环境信息:\n", a.environment.Summary())
	}, 10)

	// 启动后钩子 - 打印启动时间
	a.hooks.RegisterAfterStart("print_boot_time", func() {
		bootTime := time.Since(a.bootStartTime)
		a.logger.Infof("应用启动完成，耗时: %s", bootTime)
	}, 10)

	// 关闭前钩子 - 打印关闭提示
	a.hooks.RegisterBeforeShutdown("print_shutdown", func() {
		a.logger.Info("应用正在关闭...")
	}, 10)

	// 关闭后钩子 - 打印运行时间
	a.hooks.RegisterAfterShutdown("print_runtime", func() {
		uptime := a.environment.Uptime()
		a.logger.Infof("应用已关闭，总运行时间: %s", uptime)
	}, 10)
}

// Engine 获取Flow引擎
func (a *Application) Engine() *flow.Engine {
	return a.engine
}

// Environment 获取环境信息
func (a *Application) Environment() *Environment {
	return a.environment
}

// Logger 获取日志记录器
func (a *Application) Logger() *logrus.Logger {
	return a.logger
}

// RegisterProvider 注册服务提供者
func (a *Application) RegisterProvider(provider ServiceProvider) {
	a.providerManager.Register(provider)
}

// RegisterProviders 注册多个服务提供者
func (a *Application) RegisterProviders(providers []ServiceProvider) {
	a.providerManager.RegisterAll(providers)
}

// RegisterHook 注册应用钩子
func (a *Application) RegisterHook(hookType HookType, name string, function func(), priority int) {
	a.hooks.Register(Hook{
		Name:     name,
		Function: function,
		Type:     hookType,
		Priority: priority,
	})
}

// Boot 启动应用
func (a *Application) Boot() error {
	// 执行启动前钩子
	a.hooks.Execute(HookBeforeStart)

	// 启动所有服务提供者
	if err := a.providerManager.BootAll(a); err != nil {
		return err
	}

	// 执行启动后钩子
	a.hooks.Execute(HookAfterStart)

	return nil
}

// Run 运行应用
func (a *Application) Run(addr string) error {
	// 先启动应用
	if err := a.Boot(); err != nil {
		return err
	}

	// 启动HTTP服务器
	return a.lifecycle.Start(addr)
}

// Shutdown 关闭应用
func (a *Application) Shutdown(timeout time.Duration) error {
	// 执行关闭前钩子
	a.hooks.Execute(HookBeforeShutdown)

	// 关闭应用
	err := a.lifecycle.Shutdown(timeout)

	// 执行关闭后钩子
	a.hooks.Execute(HookAfterShutdown)

	return err
}

// OnBeforeStart 注册启动前钩子
func (a *Application) OnBeforeStart(name string, function func(), priority int) {
	a.hooks.RegisterBeforeStart(name, function, priority)
}

// OnAfterStart 注册启动后钩子
func (a *Application) OnAfterStart(name string, function func(), priority int) {
	a.hooks.RegisterAfterStart(name, function, priority)
}

// OnBeforeShutdown 注册关闭前钩子
func (a *Application) OnBeforeShutdown(name string, function func(), priority int) {
	a.hooks.RegisterBeforeShutdown(name, function, priority)
}

// OnAfterShutdown 注册关闭后钩子
func (a *Application) OnAfterShutdown(name string, function func(), priority int) {
	a.hooks.RegisterAfterShutdown(name, function, priority)
}
