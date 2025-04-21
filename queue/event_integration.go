package queue

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/zzliekkas/flow/event"
)

// QueueEventListener 队列事件监听器，将事件转换为队列任务
type QueueEventListener struct {
	manager      *QueueManager
	queueName    string
	eventMapping map[string]string // 事件名称 -> 任务名称映射
}

// NewQueueEventListener 创建一个新的队列事件监听器
func NewQueueEventListener(manager *QueueManager, queueName string) *QueueEventListener {
	return &QueueEventListener{
		manager:      manager,
		queueName:    queueName,
		eventMapping: make(map[string]string),
	}
}

// RegisterEventJob 注册事件到任务的映射
func (l *QueueEventListener) RegisterEventJob(eventName, jobName string) {
	l.eventMapping[eventName] = jobName
}

// Handle 处理事件，将其转换为队列任务
func (l *QueueEventListener) Handle(evt event.Event) error {
	eventName := evt.GetName()
	jobName, exists := l.eventMapping[eventName]

	if !exists {
		// 如果没有找到映射，可以使用事件名称作为任务名称
		jobName = eventName
	}

	// 使用事件的负载作为任务负载
	payload := evt.GetPayload()

	// 为了跟踪，添加一些元数据
	payload["event_timestamp"] = evt.GetTimestamp()
	payload["event_name"] = eventName

	// 将任务推送到队列
	queue, err := l.manager.GetQueue(l.queueName)
	if err != nil {
		return fmt.Errorf("获取队列失败: %w", err)
	}

	jobID, err := queue.Push(context.Background(), l.queueName, jobName, payload)
	if err != nil {
		return fmt.Errorf("推送任务到队列失败: %w", err)
	}

	log.Printf("事件 %s 已转换为任务 %s，任务ID: %s", eventName, jobName, jobID)
	return nil
}

// ShouldHandle 判断是否应该处理此事件
func (l *QueueEventListener) ShouldHandle(evt event.Event) bool {
	_, exists := l.eventMapping[evt.GetName()]
	return exists
}

// QueueEventProvider 将队列任务状态变化转换为事件
type QueueEventProvider struct {
	dispatcher event.Dispatcher
}

// NewQueueEventProvider 创建一个新的队列事件提供者
func NewQueueEventProvider(dispatcher event.Dispatcher) *QueueEventProvider {
	return &QueueEventProvider{
		dispatcher: dispatcher,
	}
}

// JobStatusChanged 当任务状态改变时触发事件
func (p *QueueEventProvider) JobStatusChanged(job *Job, oldStatus JobStatus) {
	// 创建事件名称，例如 "queue.job.completed"
	eventName := fmt.Sprintf("queue.job.%s", job.Status)

	// 创建事件
	evt := event.NewBaseEvent(eventName)

	// 设置事件负载
	evt.SetPayload(map[string]interface{}{
		"job_id":     job.ID,
		"queue":      job.Queue,
		"name":       job.Name,
		"status":     job.Status,
		"old_status": oldStatus,
		"attempts":   job.Attempts,
		"created_at": job.CreatedAt,
		"updated_at": job.UpdatedAt,
		"error":      job.Error,
	})

	// 分发事件
	err := p.dispatcher.Dispatch(evt)
	if err != nil {
		log.Printf("分发任务状态变化事件失败: %v", err)
	}
}

// QueueMiddleware 队列中间件，为任务处理添加额外功能
type QueueMiddleware func(next Handler) Handler

// LoggingMiddleware 创建一个记录日志的中间件
func LoggingMiddleware() QueueMiddleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, job *Job) error {
			start := time.Now()
			log.Printf("开始处理任务 %s (ID: %s)", job.Name, job.ID)

			err := next(ctx, job)

			duration := time.Since(start)
			if err != nil {
				log.Printf("任务 %s (ID: %s) 处理失败: %v，耗时: %v", job.Name, job.ID, err, duration)
			} else {
				log.Printf("任务 %s (ID: %s) 处理成功，耗时: %v", job.Name, job.ID, duration)
			}

			return err
		}
	}
}

// RetryMiddleware 创建一个处理重试的中间件
func RetryMiddleware(maxRetries int) QueueMiddleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, job *Job) error {
			var err error

			// 设置最大重试次数
			if job.MaxRetries <= 0 {
				job.MaxRetries = maxRetries
			}

			// 尝试执行任务
			err = next(ctx, job)

			// 如果任务执行失败且未超过最大重试次数，返回错误以便重试
			if err != nil && job.Attempts < job.MaxRetries {
				return fmt.Errorf("任务执行失败，将进行重试: %w", err)
			}

			return err
		}
	}
}

// ApplyMiddleware 应用中间件到任务处理器
func ApplyMiddleware(handler Handler, middlewares ...QueueMiddleware) Handler {
	// 从后向前应用中间件，确保第一个中间件位于最外层
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

// RegisterWithMiddleware 使用中间件注册任务处理器
func (m *QueueManager) RegisterWithMiddleware(jobName string, handler Handler, middlewares ...QueueMiddleware) {
	// 应用中间件
	wrappedHandler := ApplyMiddleware(handler, middlewares...)

	// 注册包装后的处理器
	m.Register(jobName, wrappedHandler)
}
