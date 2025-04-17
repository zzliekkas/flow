package event

import "sync"

// 定义监听器优先级常量
const (
	PriorityHigh   int = 100
	PriorityNormal int = 0
	PriorityLow    int = -100
)

// BaseListener 基础监听器实现
type BaseListener struct {
	// 处理器函数
	handler func(event Event) error
	// 过滤器函数，用于判断是否处理事件
	filter func(event Event) bool
	// 是否异步处理
	async bool
}

// NewBaseListener 创建一个新的基础监听器
func NewBaseListener(handler func(event Event) error) *BaseListener {
	return &BaseListener{
		handler: handler,
		filter:  func(event Event) bool { return true },
		async:   false,
	}
}

// NewAsyncBaseListener 创建一个新的异步基础监听器
func NewAsyncBaseListener(handler func(event Event) error) *BaseListener {
	listener := NewBaseListener(handler)
	listener.async = true
	return listener
}

// Handle 处理事件
func (l *BaseListener) Handle(event Event) error {
	return l.handler(event)
}

// ShouldHandle 检查是否应该处理此事件
func (l *BaseListener) ShouldHandle(event Event) bool {
	return l.filter(event)
}

// ShouldProcess 检查是否应异步处理
func (l *BaseListener) ShouldProcess() bool {
	return l.async
}

// WithFilter 设置过滤器
func (l *BaseListener) WithFilter(filter func(event Event) bool) *BaseListener {
	l.filter = filter
	return l
}

// WithAsync 设置是否异步处理
func (l *BaseListener) WithAsync(async bool) *BaseListener {
	l.async = async
	return l
}

// EventMap 事件到监听器的映射类型
type EventMap map[string][]Listener

// EventSubscriber 事件订阅者接口
type EventSubscriber interface {
	// Subscribe 返回事件到监听器的映射
	Subscribe() EventMap
}

// BaseSubscriber 基础事件订阅者实现
type BaseSubscriber struct {
	events EventMap
}

// NewSubscriber 创建一个新的基础订阅者
func NewSubscriber() *BaseSubscriber {
	return &BaseSubscriber{
		events: make(EventMap),
	}
}

// Subscribe 实现 EventSubscriber 接口
func (s *BaseSubscriber) Subscribe() EventMap {
	return s.events
}

// Listen 添加事件和对应的监听器
func (s *BaseSubscriber) Listen(eventName string, listener Listener) *BaseSubscriber {
	s.events[eventName] = append(s.events[eventName], listener)
	return s
}

// ListenFunc 添加事件和对应的函数处理器
func (s *BaseSubscriber) ListenFunc(eventName string, handler func(event Event) error) *BaseSubscriber {
	return s.Listen(eventName, ListenerFunc(handler))
}

// ListenAsync 添加事件和对应的异步处理器
func (s *BaseSubscriber) ListenAsync(eventName string, handler func(event Event) error) *BaseSubscriber {
	// 创建一个实现 Listener 和 AsyncListener 接口的对象
	asyncListener := NewAsyncListenerFunc(handler, true)
	return s.Listen(eventName, asyncListener)
}

// EventListener 事件监听器接口
type EventListener interface {
	// Handle 处理事件
	Handle(event Event) error
	// GetEvents 获取该监听器关注的事件列表
	GetEvents() []string
	// GetPriority 获取监听器优先级
	GetPriority() int
}

// BaseEventListener 基础事件监听器实现
type BaseEventListener struct {
	// 处理函数
	handler func(event Event) error
	// 关注的事件列表
	events []string
	// 优先级
	priority int
}

// NewEventListener 创建新的事件监听器
func NewEventListener(handler func(event Event) error, events []string, priority int) *BaseEventListener {
	return &BaseEventListener{
		handler:  handler,
		events:   events,
		priority: priority,
	}
}

// Handle 实现事件处理方法
func (l *BaseEventListener) Handle(event Event) error {
	return l.handler(event)
}

// GetEvents 获取监听器关注的事件列表
func (l *BaseEventListener) GetEvents() []string {
	return l.events
}

// GetPriority 获取监听器优先级
func (l *BaseEventListener) GetPriority() int {
	return l.priority
}

// ListenerManager 监听器管理器接口
type ListenerManager interface {
	// AddListener 添加监听器
	AddListener(listener EventListener)
	// RemoveListener 移除监听器
	RemoveListener(listener EventListener)
	// GetListeners 获取指定事件的所有监听器
	GetListeners(eventName string) []EventListener
	// HasListeners 检查是否有监听器关注指定事件
	HasListeners(eventName string) bool
}

// StandardListenerManager 标准监听器管理器实现
type StandardListenerManager struct {
	// 事件到监听器的映射
	listeners map[string][]EventListener
	// 互斥锁
	mutex sync.RWMutex
}

// NewListenerManager 创建新的监听器管理器
func NewListenerManager() *StandardListenerManager {
	return &StandardListenerManager{
		listeners: make(map[string][]EventListener),
	}
}

// AddListener 添加监听器
func (m *StandardListenerManager) AddListener(listener EventListener) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	events := listener.GetEvents()
	for _, eventName := range events {
		listeners := m.listeners[eventName]

		// 按优先级插入监听器
		inserted := false
		for i, l := range listeners {
			if listener.GetPriority() > l.GetPriority() {
				// 在i位置插入
				listeners = append(listeners[:i+1], listeners[i:]...)
				listeners[i] = listener
				inserted = true
				break
			}
		}

		// 如果没有插入，则添加到末尾
		if !inserted {
			listeners = append(listeners, listener)
		}

		m.listeners[eventName] = listeners
	}
}

// RemoveListener 移除监听器
func (m *StandardListenerManager) RemoveListener(listener EventListener) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	events := listener.GetEvents()
	for _, eventName := range events {
		listeners := m.listeners[eventName]
		for i, l := range listeners {
			if l == listener {
				// 移除该监听器
				m.listeners[eventName] = append(listeners[:i], listeners[i+1:]...)
				break
			}
		}
	}
}

// GetListeners 获取指定事件的所有监听器
func (m *StandardListenerManager) GetListeners(eventName string) []EventListener {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.listeners[eventName]
}

// HasListeners 检查是否有监听器关注指定事件
func (m *StandardListenerManager) HasListeners(eventName string) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	listeners, exists := m.listeners[eventName]
	return exists && len(listeners) > 0
}
