package queue

import (
	"context"
	"errors"
	"sync"
	"time"
)

// 队列管理器特有的错误定义
var (
	ErrDefaultQueueNotSet = errors.New("未设置默认队列")
)

// QueueManager 队列管理器实现
type QueueManager struct {
	mu           sync.RWMutex
	queues       map[string]Queue
	defaultQueue string
}

// NewQueueManager 创建一个新的队列管理器
func NewQueueManager() *QueueManager {
	return &QueueManager{
		queues: make(map[string]Queue),
	}
}

// AddQueue 添加队列到管理器
func (m *QueueManager) AddQueue(name string, queue Queue) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.queues[name]; exists {
		return ErrQueueAlreadyExists
	}

	m.queues[name] = queue

	// 如果这是第一个添加的队列，将其设为默认队列
	if m.defaultQueue == "" {
		m.defaultQueue = name
	}

	return nil
}

// RemoveQueue 从管理器删除队列
func (m *QueueManager) RemoveQueue(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.queues[name]; !exists {
		return ErrQueueNotFound
	}

	delete(m.queues, name)

	// 如果删除的是默认队列，则清空默认队列设置
	if m.defaultQueue == name {
		m.defaultQueue = ""
	}

	return nil
}

// GetQueue 获取指定队列
func (m *QueueManager) GetQueue(name string) (Queue, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	queue, exists := m.queues[name]
	if !exists {
		return nil, ErrQueueNotFound
	}

	return queue, nil
}

// ListQueues 列出所有队列名称
func (m *QueueManager) ListQueues() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var names []string
	for name := range m.queues {
		names = append(names, name)
	}

	return names
}

// HasQueue 检查队列是否存在
func (m *QueueManager) HasQueue(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, exists := m.queues[name]
	return exists
}

// DefaultQueue 获取默认队列
func (m *QueueManager) DefaultQueue() Queue {
	queue, _ := m.GetDefaultQueue()
	return queue
}

// SetDefaultQueue 设置默认队列
func (m *QueueManager) SetDefaultQueue(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.queues[name]; !exists {
		return ErrQueueNotFound
	}

	m.defaultQueue = name
	return nil
}

// GetDefaultQueue 获取默认队列
func (m *QueueManager) GetDefaultQueue() (Queue, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.defaultQueue == "" {
		return nil, ErrDefaultQueueNotSet
	}

	queue, exists := m.queues[m.defaultQueue]
	if !exists {
		return nil, ErrQueueNotFound
	}

	return queue, nil
}

// GetDefaultQueueName 获取默认队列名称
func (m *QueueManager) GetDefaultQueueName() (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.defaultQueue == "" {
		return "", ErrDefaultQueueNotSet
	}

	return m.defaultQueue, nil
}

// Push 使用默认队列推送任务
func (m *QueueManager) Push(ctx context.Context, jobName string, payload map[string]interface{}) (string, error) {
	queue, err := m.GetDefaultQueue()
	if err != nil {
		return "", err
	}

	return queue.Push(ctx, m.defaultQueue, jobName, payload)
}

// PushWithDelay 使用默认队列延迟推送任务
func (m *QueueManager) PushWithDelay(ctx context.Context, jobName string, payload map[string]interface{}, delay time.Duration) (string, error) {
	queue, err := m.GetDefaultQueue()
	if err != nil {
		return "", err
	}

	return queue.PushWithDelay(ctx, m.defaultQueue, jobName, payload, delay)
}

// Schedule 使用默认队列计划任务
func (m *QueueManager) Schedule(ctx context.Context, jobName string, payload map[string]interface{}, scheduledAt time.Time) (string, error) {
	queue, err := m.GetDefaultQueue()
	if err != nil {
		return "", err
	}

	return queue.Schedule(ctx, m.defaultQueue, jobName, payload, scheduledAt)
}

// Register 为所有队列注册同一个处理器
func (m *QueueManager) Register(jobName string, handler Handler) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, queue := range m.queues {
		queue.Register(jobName, handler)
	}
}
