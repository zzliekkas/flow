package event

import (
	"github.com/zzliekkas/flow"
	"github.com/zzliekkas/flow/di"
)

const (
	// ManagerService 事件管理器服务名称
	ManagerService = "event.manager"
	// DispatcherService 事件分发器服务名称
	DispatcherService = "event.dispatcher"
)

// EventProvider 事件系统服务提供者
type EventProvider struct {
	container *di.Container
}

// NewEventProvider 创建事件服务提供者
func NewEventProvider(container *di.Container) *EventProvider {
	return &EventProvider{
		container: container,
	}
}

// Register 注册事件系统服务
func (p *EventProvider) Register() error {
	// 注册事件管理器
	err := p.container.Provide(func() *Manager {
		return NewManager()
	})
	if err != nil {
		return err
	}

	// 注册默认事件分发器
	err = p.container.Provide(func(manager *Manager) Dispatcher {
		return manager.GetDefaultDispatcher()
	})
	if err != nil {
		return err
	}

	return nil
}

// Boot 引导事件系统
func (p *EventProvider) Boot() error {
	// 注册应用启动事件
	var manager *Manager
	err := p.container.Extract(&manager)
	if err != nil {
		return err
	}

	// 监听应用启动事件
	err = manager.AddListenerFunc("app.started", func(event Event) error {
		// 这里可以添加应用启动时的事件处理逻辑
		return nil
	})
	if err != nil {
		return err
	}

	// 提取 Flow 引擎
	var engine *flow.Engine
	err = p.container.Extract(&engine)
	if err != nil {
		return err
	}

	// 注册中间件，用于处理请求开始和结束的事件
	engine.Use(func(c *flow.Context) {
		// 派发请求开始事件
		requestStartedEvent := NewBaseEvent("http.request.started")
		requestStartedEvent.SetPayload(map[string]interface{}{
			"path":   c.Request.URL.Path,
			"method": c.Request.Method,
		})
		requestStartedEvent.SetContext(c)

		_ = manager.DispatchAsync(requestStartedEvent)

		// 继续处理请求
		c.Next()

		// 派发请求结束事件
		requestEndedEvent := NewBaseEvent("http.request.ended")
		requestEndedEvent.SetPayload(map[string]interface{}{
			"path":   c.Request.URL.Path,
			"method": c.Request.Method,
			"status": c.Writer.Status(),
		})
		requestEndedEvent.SetContext(c)

		_ = manager.DispatchAsync(requestEndedEvent)
	})

	return nil
}

// Name 获取提供者名称
func (p *EventProvider) Name() string {
	return "event"
}

// Priority 获取提供者优先级
func (p *EventProvider) Priority() int {
	return 0
}
