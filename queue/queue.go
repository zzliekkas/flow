package queue

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

// 常见错误定义
var (
	ErrQueueNotFound      = errors.New("queue: 队列不存在")
	ErrQueueAlreadyExists = errors.New("queue: 队列已存在")
	ErrJobNotFound        = errors.New("queue: 任务不存在")
	ErrInvalidJobID       = errors.New("queue: 无效的任务ID")
	ErrQueueFull          = errors.New("queue: 队列已满")
	ErrInvalidPayload     = errors.New("queue: 无效的任务负载")
)

// Job 表示队列中的一个任务
type Job struct {
	ID          string                 `json:"id"`                     // 任务唯一标识
	Queue       string                 `json:"queue"`                  // 所属队列
	Name        string                 `json:"name"`                   // 任务名称
	Payload     map[string]interface{} `json:"payload"`                // 任务负载数据
	Attempts    int                    `json:"attempts"`               // 尝试次数
	MaxRetries  int                    `json:"max_retries"`            // 最大重试次数
	Status      JobStatus              `json:"status"`                 // 任务状态
	CreatedAt   time.Time              `json:"created_at"`             // 创建时间
	UpdatedAt   time.Time              `json:"updated_at"`             // 更新时间
	ScheduledAt *time.Time             `json:"scheduled_at,omitempty"` // 计划执行时间
	StartedAt   *time.Time             `json:"started_at,omitempty"`   // 开始执行时间
	FinishedAt  *time.Time             `json:"finished_at,omitempty"`  // 完成时间
	Error       string                 `json:"error,omitempty"`        // 错误信息
}

// JobStatus 表示任务的状态
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"   // 等待执行
	JobStatusScheduled JobStatus = "scheduled" // 已计划执行
	JobStatusRunning   JobStatus = "running"   // 执行中
	JobStatusCompleted JobStatus = "completed" // 已完成
	JobStatusFailed    JobStatus = "failed"    // 执行失败
	JobStatusRetrying  JobStatus = "retrying"  // 等待重试
	JobStatusCancelled JobStatus = "cancelled" // 已取消
)

// Handler 表示任务处理器
type Handler func(ctx context.Context, job *Job) error

// Queue 表示一个任务队列的抽象接口
type Queue interface {
	// Push 将任务推送到队列
	Push(ctx context.Context, queueName string, jobName string, payload map[string]interface{}) (string, error)

	// PushWithDelay 将任务推送到队列，延迟指定时间后执行
	PushWithDelay(ctx context.Context, queueName string, jobName string, payload map[string]interface{}, delay time.Duration) (string, error)

	// Schedule 计划一个任务在指定时间执行
	Schedule(ctx context.Context, queueName string, jobName string, payload map[string]interface{}, scheduledAt time.Time) (string, error)

	// Get 获取任务信息
	Get(ctx context.Context, queueName string, jobID string) (*Job, error)

	// Delete 删除任务
	Delete(ctx context.Context, queueName string, jobID string) error

	// Clear 清空队列
	Clear(ctx context.Context, queueName string) error

	// Size 获取队列大小
	Size(ctx context.Context, queueName string) (int, error)

	// Register 注册任务处理器
	Register(jobName string, handler Handler)

	// ProcessNext 处理队列中的下一个任务
	ProcessNext(ctx context.Context, queueName string) error

	// StartWorker 启动工作进程处理任务
	StartWorker(ctx context.Context, queueName string, concurrency int) error

	// StopWorker 停止工作进程
	StopWorker(ctx context.Context, queueName string) error

	// Retry 重试失败的任务
	Retry(ctx context.Context, queueName string, jobID string) error
}

// GetPayload 将任务负载解析为指定类型
func (j *Job) GetPayload(v interface{}) error {
	data, err := json.Marshal(j.Payload)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// GetPayloadValue 从任务负载中获取指定键的值，返回值和是否存在
func (j *Job) GetPayloadValue(key string) (interface{}, bool) {
	value, exists := j.Payload[key]
	return value, exists
}
