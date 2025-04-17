package event

import (
	"time"
)

// Event 事件接口
type Event interface {
	// GetName 获取事件名称
	GetName() string
	// GetTimestamp 获取事件时间戳
	GetTimestamp() time.Time
	// GetPayload 获取事件负载数据
	GetPayload() map[string]interface{}
	// SetPayload 设置事件负载数据
	SetPayload(payload map[string]interface{})
	// GetPayloadValue 获取特定的负载值
	GetPayloadValue(key string) (interface{}, bool)
	// SetPayloadValue 设置特定的负载值
	SetPayloadValue(key string, value interface{})
	// GetContext 获取事件上下文
	GetContext() interface{}
	// SetContext 设置事件上下文
	SetContext(ctx interface{})
	// IsPropagationStopped 事件传播是否已停止
	IsPropagationStopped() bool
	// StopPropagation 停止事件传播
	StopPropagation()
}

// BaseEvent 基础事件实现
type BaseEvent struct {
	// 事件名称
	name string
	// 事件时间戳
	timestamp time.Time
	// 事件负载
	payload map[string]interface{}
	// 事件上下文
	context interface{}
	// 是否停止传播
	propagationStopped bool
}

// NewBaseEvent 创建基础事件
func NewBaseEvent(name string) *BaseEvent {
	return &BaseEvent{
		name:      name,
		timestamp: time.Now(),
		payload:   make(map[string]interface{}),
	}
}

// GetName 获取事件名称
func (e *BaseEvent) GetName() string {
	return e.name
}

// GetTimestamp 获取事件时间戳
func (e *BaseEvent) GetTimestamp() time.Time {
	return e.timestamp
}

// GetPayload 获取事件负载
func (e *BaseEvent) GetPayload() map[string]interface{} {
	return e.payload
}

// SetPayload 设置事件负载
func (e *BaseEvent) SetPayload(payload map[string]interface{}) {
	e.payload = payload
}

// GetPayloadValue 获取特定的负载值
func (e *BaseEvent) GetPayloadValue(key string) (interface{}, bool) {
	value, exists := e.payload[key]
	return value, exists
}

// SetPayloadValue 设置特定的负载值
func (e *BaseEvent) SetPayloadValue(key string, value interface{}) {
	if e.payload == nil {
		e.payload = make(map[string]interface{})
	}
	e.payload[key] = value
}

// GetContext 获取事件上下文
func (e *BaseEvent) GetContext() interface{} {
	return e.context
}

// SetContext 设置事件上下文
func (e *BaseEvent) SetContext(ctx interface{}) {
	e.context = ctx
}

// IsPropagationStopped 事件传播是否已停止
func (e *BaseEvent) IsPropagationStopped() bool {
	return e.propagationStopped
}

// StopPropagation 停止事件传播
func (e *BaseEvent) StopPropagation() {
	e.propagationStopped = true
}

// Listener 事件监听器接口
type Listener interface {
	// Handle 处理事件
	Handle(event Event) error

	// ShouldHandle 检查是否应该处理此事件
	ShouldHandle(event Event) bool
}

// AsyncListener 异步事件监听器接口
type AsyncListener interface {
	Listener

	// ShouldProcess 检查是否应该异步处理
	ShouldProcess() bool
}

// ListenerFunc 函数类型的监听器
type ListenerFunc func(event Event) error

// Handle 实现 Listener 接口
func (f ListenerFunc) Handle(event Event) error {
	return f(event)
}

// ShouldHandle 实现 Listener 接口
func (f ListenerFunc) ShouldHandle(event Event) bool {
	return true
}

// AsyncListenerFunc 异步函数类型的监听器
type AsyncListenerFunc struct {
	ListenerFunc
	async bool
}

// NewAsyncListenerFunc 创建异步函数监听器
func NewAsyncListenerFunc(handler func(event Event) error, async bool) AsyncListener {
	return &AsyncListenerFunc{
		ListenerFunc: handler,
		async:        async,
	}
}

// ShouldProcess 实现 AsyncListener 接口
func (f *AsyncListenerFunc) ShouldProcess() bool {
	return f.async
}
