package event

import (
	"errors"
	"sync"
)

var (
	// ErrListenerAlreadyRegistered 监听器已注册错误
	ErrListenerAlreadyRegistered = errors.New("监听器已经被注册")

	// ErrEventNotSupported 事件不支持错误
	ErrEventNotSupported = errors.New("事件类型不受支持")
)

// Dispatcher 事件分发器接口
type Dispatcher interface {
	// Dispatch 分发事件
	Dispatch(event Event) error
	// DispatchAsync 异步分发事件
	DispatchAsync(event Event) error
	// AddListener 添加事件监听器
	AddListener(eventName string, listener Listener) error
	// AddListenerFunc 添加函数形式的事件监听器
	AddListenerFunc(eventName string, listenerFunc func(event Event) error) error
	// RemoveListener 移除事件监听器
	RemoveListener(eventName string, listener Listener) error
	// GetListeners 获取指定事件的所有监听器
	GetListeners(eventName string) []Listener
	// HasListeners 检查事件是否有监听器
	HasListeners(eventName string) bool
}

// EventDispatcher 事件分发器接口
type EventDispatcher interface {
	// Dispatch 分发事件
	Dispatch(event Event) error
	// Listen 添加事件监听器
	Listen(eventName string, listener EventListener)
	// HasListeners 检查事件是否有监听器
	HasListeners(eventName string) bool
	// GetListeners 获取指定事件的所有监听器
	GetListeners(eventName string) []EventListener
	// RemoveListener 移除监听器
	RemoveListener(listener EventListener)
}

// StandardEventDispatcher 标准事件分发器实现
type StandardEventDispatcher struct {
	// 监听器管理器
	listenerManager ListenerManager
	// 用于异步事件处理的WaitGroup
	wg sync.WaitGroup
}

// NewEventDispatcher 创建新的事件分发器
func NewEventDispatcher() *StandardEventDispatcher {
	return &StandardEventDispatcher{
		listenerManager: NewListenerManager(),
	}
}

// Dispatch 分发事件
func (d *StandardEventDispatcher) Dispatch(event Event) error {
	eventName := event.GetName()
	if !d.HasListeners(eventName) {
		return nil
	}

	listeners := d.GetListeners(eventName)
	for _, listener := range listeners {
		// 检查事件传播是否已停止
		if event.IsPropagationStopped() {
			break
		}

		// 异步处理判断方式一：检查AsyncEventListener接口
		if asyncListener, ok := listener.(*AsyncEventListener); ok && asyncListener.IsAsync() {
			d.wg.Add(1)
			go func(l EventListener, e Event) {
				defer d.wg.Done()
				_ = l.Handle(e)
			}(asyncListener, event)
			continue
		}

		// 异步处理判断方式二：检查AsyncListener接口
		if asyncListener, ok := listener.(AsyncListener); ok && asyncListener.ShouldProcess() {
			d.wg.Add(1)
			go func(l EventListener, e Event) {
				defer d.wg.Done()
				_ = l.Handle(e)
			}(listener, event)
			continue
		}

		// 同步执行
		err := listener.Handle(event)
		if err != nil {
			return err
		}
	}

	return nil
}

// DispatchSync 同步分发事件，等待所有异步监听器完成
func (d *StandardEventDispatcher) DispatchSync(event Event) error {
	err := d.Dispatch(event)
	d.wg.Wait()
	return err
}

// Listen 添加事件监听器
func (d *StandardEventDispatcher) Listen(eventName string, listener EventListener) {
	// 创建单事件监听器包装
	singleEventListener := &BaseEventListener{
		handler:  listener.Handle,
		events:   []string{eventName},
		priority: listener.GetPriority(),
	}

	d.listenerManager.AddListener(singleEventListener)
}

// HasListeners 检查事件是否有监听器
func (d *StandardEventDispatcher) HasListeners(eventName string) bool {
	return d.listenerManager.HasListeners(eventName)
}

// GetListeners 获取指定事件的所有监听器
func (d *StandardEventDispatcher) GetListeners(eventName string) []EventListener {
	return d.listenerManager.GetListeners(eventName)
}

// RemoveListener 移除监听器
func (d *StandardEventDispatcher) RemoveListener(listener EventListener) {
	d.listenerManager.RemoveListener(listener)
}

// AsyncEventListener 异步事件监听器
type AsyncEventListener struct {
	BaseEventListener
	async bool
}

// NewAsyncListener 创建新的异步事件监听器
func NewAsyncListener(handler func(event Event) error, events []string, priority int, async bool) *AsyncEventListener {
	return &AsyncEventListener{
		BaseEventListener: BaseEventListener{
			handler:  handler,
			events:   events,
			priority: priority,
		},
		async: async,
	}
}

// IsAsync 检查监听器是否为异步
func (l *AsyncEventListener) IsAsync() bool {
	return l.async
}

// StandardDispatcher 标准分发器实现，实现 Dispatcher 接口
type StandardDispatcher struct {
	eventDispatcher EventDispatcher
	maxQueue        int
	asyncEvents     chan Event
	stopChan        chan struct{}
	wg              sync.WaitGroup
}

// NewDispatcher 创建新的标准分发器
func NewDispatcher(maxQueue int) *StandardDispatcher {
	if maxQueue <= 0 {
		maxQueue = 100
	}

	dispatcher := &StandardDispatcher{
		eventDispatcher: NewEventDispatcher(),
		maxQueue:        maxQueue,
		asyncEvents:     make(chan Event, maxQueue),
		stopChan:        make(chan struct{}),
	}

	// 启动异步事件处理
	dispatcher.startAsyncProcessing()

	return dispatcher
}

// startAsyncProcessing 启动异步事件处理
func (d *StandardDispatcher) startAsyncProcessing() {
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		for {
			select {
			case event := <-d.asyncEvents:
				// 处理异步事件
				_ = d.eventDispatcher.Dispatch(event)
			case <-d.stopChan:
				// 关闭信号
				return
			}
		}
	}()
}

// Dispatch 分发事件
func (d *StandardDispatcher) Dispatch(event Event) error {
	return d.eventDispatcher.Dispatch(event)
}

// DispatchAsync 异步分发事件
func (d *StandardDispatcher) DispatchAsync(event Event) error {
	select {
	case d.asyncEvents <- event:
		return nil
	default:
		// 队列已满，同步处理
		return d.eventDispatcher.Dispatch(event)
	}
}

// AddListener 添加事件监听器
func (d *StandardDispatcher) AddListener(eventName string, listener Listener) error {
	// 转换为 EventListener 接口
	eventListener := &BaseEventListener{
		handler:  listener.Handle,
		events:   []string{eventName},
		priority: 0, // 默认优先级
	}
	d.eventDispatcher.Listen(eventName, eventListener)
	return nil
}

// AddListenerFunc 添加函数形式的事件监听器
func (d *StandardDispatcher) AddListenerFunc(eventName string, listenerFunc func(event Event) error) error {
	return d.AddListener(eventName, ListenerFunc(listenerFunc))
}

// RemoveListener 移除事件监听器
func (d *StandardDispatcher) RemoveListener(eventName string, listener Listener) error {
	// 这里简化实现，实际上需要在 EventDispatcher 中找到对应的 EventListener
	return nil
}

// GetListeners 获取指定事件的所有监听器
func (d *StandardDispatcher) GetListeners(eventName string) []Listener {
	// 从 EventDispatcher 获取 EventListener 并转换为 Listener
	eventListeners := d.eventDispatcher.GetListeners(eventName)
	listeners := make([]Listener, 0, len(eventListeners))

	// 转换接口类型
	for _, el := range eventListeners {
		listeners = append(listeners, ListenerFunc(el.Handle))
	}

	return listeners
}

// HasListeners 检查事件是否有监听器
func (d *StandardDispatcher) HasListeners(eventName string) bool {
	return d.eventDispatcher.HasListeners(eventName)
}

// Stop 停止异步事件处理
func (d *StandardDispatcher) Stop() {
	close(d.stopChan)
	d.wg.Wait()
}
