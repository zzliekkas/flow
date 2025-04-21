package memory

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/zzliekkas/flow/queue"
)

// MemoryQueue 是基于内存的队列实现
type MemoryQueue struct {
	mu             sync.RWMutex
	queues         map[string][]*queue.Job       // 队列名称 -> 任务列表
	scheduled      map[string][]*queue.Job       // 计划任务队列
	handlers       map[string]queue.Handler      // 任务名称 -> 处理函数
	workerContexts map[string]context.CancelFunc // 队列名称 -> 停止函数
	maxRetries     int                           // 最大重试次数
}

// New 创建一个新的内存队列
func New(maxRetries int) *MemoryQueue {
	return &MemoryQueue{
		queues:         make(map[string][]*queue.Job),
		scheduled:      make(map[string][]*queue.Job),
		handlers:       make(map[string]queue.Handler),
		workerContexts: make(map[string]context.CancelFunc),
		maxRetries:     maxRetries,
	}
}

// Push 将任务推送到队列
func (m *MemoryQueue) Push(ctx context.Context, queueName string, jobName string, payload map[string]interface{}) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 创建新任务
	jobID := uuid.New().String()
	job := &queue.Job{
		ID:         jobID,
		Queue:      queueName,
		Name:       jobName,
		Payload:    payload,
		Attempts:   0,
		MaxRetries: m.maxRetries,
		Status:     queue.JobStatusPending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// 确保队列存在
	if _, exists := m.queues[queueName]; !exists {
		m.queues[queueName] = []*queue.Job{}
	}

	// 添加任务到队列
	m.queues[queueName] = append(m.queues[queueName], job)

	return jobID, nil
}

// PushWithDelay 将任务推送到队列，延迟指定时间后执行
func (m *MemoryQueue) PushWithDelay(ctx context.Context, queueName string, jobName string, payload map[string]interface{}, delay time.Duration) (string, error) {
	scheduledTime := time.Now().Add(delay)
	return m.Schedule(ctx, queueName, jobName, payload, scheduledTime)
}

// Schedule 计划一个任务在指定时间执行
func (m *MemoryQueue) Schedule(ctx context.Context, queueName string, jobName string, payload map[string]interface{}, scheduledAt time.Time) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 创建新任务
	jobID := uuid.New().String()
	job := &queue.Job{
		ID:          jobID,
		Queue:       queueName,
		Name:        jobName,
		Payload:     payload,
		Attempts:    0,
		MaxRetries:  m.maxRetries,
		Status:      queue.JobStatusScheduled,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		ScheduledAt: &scheduledAt,
	}

	// 确保计划队列存在
	if _, exists := m.scheduled[queueName]; !exists {
		m.scheduled[queueName] = []*queue.Job{}
	}

	// 添加任务到计划队列
	m.scheduled[queueName] = append(m.scheduled[queueName], job)

	return jobID, nil
}

// Get 获取任务信息
func (m *MemoryQueue) Get(ctx context.Context, queueName string, jobID string) (*queue.Job, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 在主队列中查找
	for _, job := range m.queues[queueName] {
		if job.ID == jobID {
			return job, nil
		}
	}

	// 在计划队列中查找
	for _, job := range m.scheduled[queueName] {
		if job.ID == jobID {
			return job, nil
		}
	}

	return nil, queue.ErrJobNotFound
}

// Delete 删除任务
func (m *MemoryQueue) Delete(ctx context.Context, queueName string, jobID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 从主队列中删除
	if jobs, exists := m.queues[queueName]; exists {
		for i, job := range jobs {
			if job.ID == jobID {
				m.queues[queueName] = append(jobs[:i], jobs[i+1:]...)
				return nil
			}
		}
	}

	// 从计划队列中删除
	if jobs, exists := m.scheduled[queueName]; exists {
		for i, job := range jobs {
			if job.ID == jobID {
				m.scheduled[queueName] = append(jobs[:i], jobs[i+1:]...)
				return nil
			}
		}
	}

	return queue.ErrJobNotFound
}

// Clear 清空队列
func (m *MemoryQueue) Clear(ctx context.Context, queueName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.queues[queueName] = []*queue.Job{}
	m.scheduled[queueName] = []*queue.Job{}
	return nil
}

// Size 获取队列大小
func (m *MemoryQueue) Size(ctx context.Context, queueName string) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.queues[queueName]) + len(m.scheduled[queueName]), nil
}

// Register 注册任务处理器
func (m *MemoryQueue) Register(jobName string, handler queue.Handler) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.handlers[jobName] = handler
}

// ProcessNext 处理队列中的下一个任务
func (m *MemoryQueue) ProcessNext(ctx context.Context, queueName string) error {
	m.mu.Lock()
	// 首先检查是否有到期的计划任务
	now := time.Now()
	if scheduledJobs, exists := m.scheduled[queueName]; exists && len(scheduledJobs) > 0 {
		var remainingJobs []*queue.Job
		var dueJobs []*queue.Job

		for _, job := range scheduledJobs {
			if job.ScheduledAt != nil && job.ScheduledAt.Before(now) {
				job.Status = queue.JobStatusPending
				dueJobs = append(dueJobs, job)
			} else {
				remainingJobs = append(remainingJobs, job)
			}
		}

		// 更新计划队列
		m.scheduled[queueName] = remainingJobs

		// 将到期任务移动到主队列
		if _, exists := m.queues[queueName]; !exists {
			m.queues[queueName] = []*queue.Job{}
		}
		m.queues[queueName] = append(m.queues[queueName], dueJobs...)
	}

	// 从主队列取出一个任务处理
	if jobs, exists := m.queues[queueName]; exists && len(jobs) > 0 {
		job := jobs[0]
		m.queues[queueName] = jobs[1:]

		// 查找处理器
		handler, exists := m.handlers[job.Name]
		if !exists {
			job.Status = queue.JobStatusFailed
			job.Error = "没有注册对应的任务处理器"
			job.UpdatedAt = time.Now()
			return errors.New("没有注册对应的任务处理器")
		}

		// 更新任务状态
		job.Status = queue.JobStatusRunning
		job.Attempts++
		now := time.Now()
		job.StartedAt = &now
		job.UpdatedAt = now

		// 解锁以避免处理任务时长时间持有锁
		m.mu.Unlock()

		// 执行任务
		err := handler(ctx, job)

		// 重新获取锁更新任务状态
		m.mu.Lock()
		defer m.mu.Unlock()

		if err != nil {
			job.Status = queue.JobStatusFailed
			job.Error = err.Error()

			// 如果未达到最大重试次数，将任务重新加入队列
			if job.Attempts < job.MaxRetries {
				job.Status = queue.JobStatusRetrying
				// 放回队列末尾
				if _, exists := m.queues[queueName]; !exists {
					m.queues[queueName] = []*queue.Job{}
				}
				m.queues[queueName] = append(m.queues[queueName], job)
			}
		} else {
			job.Status = queue.JobStatusCompleted
			finishTime := time.Now()
			job.FinishedAt = &finishTime
		}

		job.UpdatedAt = time.Now()
		return err
	}

	m.mu.Unlock()
	return nil
}

// StartWorker 启动工作进程处理任务
func (m *MemoryQueue) StartWorker(ctx context.Context, queueName string, concurrency int) error {
	m.mu.Lock()

	// 如果已有工作进程在运行，先停止
	if cancel, exists := m.workerContexts[queueName]; exists {
		cancel()
	}

	// 创建新的上下文
	workerCtx, cancel := context.WithCancel(ctx)
	m.workerContexts[queueName] = cancel
	m.mu.Unlock()

	// 启动工作进程
	for i := 0; i < concurrency; i++ {
		go func() {
			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-workerCtx.Done():
					return
				case <-ticker.C:
					// 处理下一个任务，忽略错误
					_ = m.ProcessNext(workerCtx, queueName)
				}
			}
		}()
	}

	return nil
}

// StopWorker 停止工作进程
func (m *MemoryQueue) StopWorker(ctx context.Context, queueName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if cancel, exists := m.workerContexts[queueName]; exists {
		cancel()
		delete(m.workerContexts, queueName)
		return nil
	}

	return queue.ErrQueueNotFound
}

// Retry 重试失败的任务
func (m *MemoryQueue) Retry(ctx context.Context, queueName string, jobID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 在主队列中查找
	for _, job := range m.queues[queueName] {
		if job.ID == jobID {
			if job.Status == queue.JobStatusFailed {
				job.Status = queue.JobStatusPending
				job.Error = ""
				job.UpdatedAt = time.Now()
				return nil
			}
			return errors.New("只能重试失败的任务")
		}
	}

	return queue.ErrJobNotFound
}
