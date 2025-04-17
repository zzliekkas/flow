package event

import (
	"errors"
	"sync"
)

var (
	// ErrDispatcherNotFound 分发器未找到错误
	ErrDispatcherNotFound = errors.New("事件分发器未找到")
)

// Manager 事件管理器
type Manager struct {
	// 默认分发器名称
	defaultName string
	// 分发器映射
	dispatchers map[string]Dispatcher
	// 互斥锁
	mutex sync.RWMutex
}

// NewManager 创建新的事件管理器
func NewManager() *Manager {
	manager := &Manager{
		defaultName: "default",
		dispatchers: make(map[string]Dispatcher),
		mutex:       sync.RWMutex{},
	}

	// 创建默认分发器
	manager.Register(manager.defaultName, NewDispatcher(10))

	return manager
}

// Register 注册事件分发器
func (m *Manager) Register(name string, dispatcher Dispatcher) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.dispatchers[name] = dispatcher
}

// GetDispatcher 获取指定名称的分发器
func (m *Manager) GetDispatcher(name string) (Dispatcher, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	dispatcher, exists := m.dispatchers[name]
	if !exists {
		return nil, ErrDispatcherNotFound
	}

	return dispatcher, nil
}

// GetDefaultDispatcher 获取默认分发器
func (m *Manager) GetDefaultDispatcher() Dispatcher {
	dispatcher, _ := m.GetDispatcher(m.defaultName)
	return dispatcher
}

// SetDefaultDispatcher 设置默认分发器
func (m *Manager) SetDefaultDispatcher(name string) error {
	if _, err := m.GetDispatcher(name); err != nil {
		return err
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.defaultName = name

	return nil
}

// AddListener 添加事件监听器到默认分发器
func (m *Manager) AddListener(eventName string, listener Listener) error {
	return m.GetDefaultDispatcher().AddListener(eventName, listener)
}

// AddListenerFunc 添加函数形式的事件监听器到默认分发器
func (m *Manager) AddListenerFunc(eventName string, listenerFunc func(event Event) error) error {
	return m.GetDefaultDispatcher().AddListenerFunc(eventName, listenerFunc)
}

// RemoveListener 从默认分发器移除事件监听器
func (m *Manager) RemoveListener(eventName string, listener Listener) error {
	return m.GetDefaultDispatcher().RemoveListener(eventName, listener)
}

// GetListeners 获取默认分发器中指定事件的所有监听器
func (m *Manager) GetListeners(eventName string) []Listener {
	return m.GetDefaultDispatcher().GetListeners(eventName)
}

// HasListeners 检查默认分发器中事件是否有监听器
func (m *Manager) HasListeners(eventName string) bool {
	return m.GetDefaultDispatcher().HasListeners(eventName)
}

// Dispatch 通过默认分发器派发事件
func (m *Manager) Dispatch(event Event) error {
	return m.GetDefaultDispatcher().Dispatch(event)
}

// DispatchAsync 通过默认分发器异步派发事件
func (m *Manager) DispatchAsync(event Event) error {
	return m.GetDefaultDispatcher().DispatchAsync(event)
}

// DispatchNamed 通过指定名称的分发器派发事件
func (m *Manager) DispatchNamed(name string, event Event) error {
	dispatcher, err := m.GetDispatcher(name)
	if err != nil {
		return err
	}
	return dispatcher.Dispatch(event)
}

// RegisterSubscriber 注册事件订阅者
func (m *Manager) RegisterSubscriber(subscriber EventSubscriber) error {
	return m.RegisterSubscriberToDispatcher(m.defaultName, subscriber)
}

// RegisterSubscriberToDispatcher 注册事件订阅者到指定分发器
func (m *Manager) RegisterSubscriberToDispatcher(dispatcherName string, subscriber EventSubscriber) error {
	dispatcher, err := m.GetDispatcher(dispatcherName)
	if err != nil {
		return err
	}

	eventMap := subscriber.Subscribe()
	for eventName, listeners := range eventMap {
		for _, listener := range listeners {
			if err := dispatcher.AddListener(eventName, listener); err != nil {
				return err
			}
		}
	}

	return nil
}
