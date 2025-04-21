package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/zzliekkas/flow/queue"
)

// 定义Redis队列的键前缀
const (
	// 等待执行的任务队列
	queuePrefix = "flow:queue:"
	// 计划执行的任务集合
	scheduledSetPrefix = "flow:scheduled:"
	// 任务数据哈希表
	jobDataPrefix = "flow:job:"
	// 处理中的任务集合
	processingSetPrefix = "flow:processing:"
	// 已完成的任务集合
	completedSetPrefix = "flow:completed:"
	// 失败的任务集合
	failedSetPrefix = "flow:failed:"
)

// RedisQueue 是基于Redis的队列实现
type RedisQueue struct {
	// Redis客户端
	client *redis.Client
	// 任务处理器映射
	handlers map[string]queue.Handler
	// 工作进程上下文和取消函数
	workerContexts map[string]context.CancelFunc
	// 最大重试次数
	maxRetries int
	// 互斥锁，保证并发安全
	mu sync.RWMutex
}

// Options Redis队列配置选项
type Options struct {
	// Redis连接地址
	Addr string
	// Redis密码
	Password string
	// Redis数据库
	DB int
	// 最大重试次数
	MaxRetries int
	// 连接池大小
	PoolSize int
}

// DefaultOptions 返回默认配置选项
func DefaultOptions() Options {
	return Options{
		Addr:       "localhost:6379",
		Password:   "",
		DB:         0,
		MaxRetries: 3,
		PoolSize:   10,
	}
}

// New 创建一个新的Redis队列
func New(options Options) (*RedisQueue, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     options.Addr,
		Password: options.Password,
		DB:       options.DB,
		PoolSize: options.PoolSize,
	})

	// 测试连接
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, fmt.Errorf("连接Redis失败: %w", err)
	}

	return &RedisQueue{
		client:         client,
		handlers:       make(map[string]queue.Handler),
		workerContexts: make(map[string]context.CancelFunc),
		maxRetries:     options.MaxRetries,
	}, nil
}

// Push 将任务推送到队列
func (r *RedisQueue) Push(ctx context.Context, queueName string, jobName string, payload map[string]interface{}) (string, error) {
	jobID := uuid.New().String()

	// 创建任务
	job := &queue.Job{
		ID:         jobID,
		Queue:      queueName,
		Name:       jobName,
		Payload:    payload,
		Attempts:   0,
		MaxRetries: r.maxRetries,
		Status:     queue.JobStatusPending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// 序列化任务数据
	jobData, err := json.Marshal(job)
	if err != nil {
		return "", fmt.Errorf("序列化任务失败: %w", err)
	}

	// 使用Redis管道执行多个操作
	pipe := r.client.Pipeline()

	// 存储任务数据
	pipe.Set(ctx, jobDataKey(jobID), jobData, 7*24*time.Hour) // 保存7天

	// 将任务添加到队列
	pipe.LPush(ctx, queueKey(queueName), jobID)

	// 执行管道
	_, err = pipe.Exec(ctx)
	if err != nil {
		return "", fmt.Errorf("推送任务到Redis失败: %w", err)
	}

	return jobID, nil
}

// PushWithDelay 延迟执行任务
func (r *RedisQueue) PushWithDelay(ctx context.Context, queueName string, jobName string, payload map[string]interface{}, delay time.Duration) (string, error) {
	scheduledTime := time.Now().Add(delay)
	return r.Schedule(ctx, queueName, jobName, payload, scheduledTime)
}

// Schedule 计划在特定时间执行任务
func (r *RedisQueue) Schedule(ctx context.Context, queueName string, jobName string, payload map[string]interface{}, scheduledAt time.Time) (string, error) {
	jobID := uuid.New().String()

	// 创建任务
	job := &queue.Job{
		ID:          jobID,
		Queue:       queueName,
		Name:        jobName,
		Payload:     payload,
		Attempts:    0,
		MaxRetries:  r.maxRetries,
		Status:      queue.JobStatusScheduled,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		ScheduledAt: &scheduledAt,
	}

	// 序列化任务数据
	jobData, err := json.Marshal(job)
	if err != nil {
		return "", fmt.Errorf("序列化任务失败: %w", err)
	}

	// 使用Redis管道执行多个操作
	pipe := r.client.Pipeline()

	// 存储任务数据
	pipe.Set(ctx, jobDataKey(jobID), jobData, 7*24*time.Hour) // 保存7天

	// 将任务添加到计划集合，使用时间戳作为分数
	score := float64(scheduledAt.Unix())
	pipe.ZAdd(ctx, scheduledSetKey(queueName), redis.Z{
		Score:  score,
		Member: jobID,
	})

	// 执行管道
	_, err = pipe.Exec(ctx)
	if err != nil {
		return "", fmt.Errorf("计划任务到Redis失败: %w", err)
	}

	return jobID, nil
}

// Get 获取任务信息
func (r *RedisQueue) Get(ctx context.Context, queueName string, jobID string) (*queue.Job, error) {
	// 从Redis获取任务数据
	jobData, err := r.client.Get(ctx, jobDataKey(jobID)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, queue.ErrJobNotFound
		}
		return nil, fmt.Errorf("从Redis获取任务失败: %w", err)
	}

	// 反序列化任务
	var job queue.Job
	err = json.Unmarshal(jobData, &job)
	if err != nil {
		return nil, fmt.Errorf("解析任务数据失败: %w", err)
	}

	return &job, nil
}

// Delete 删除任务
func (r *RedisQueue) Delete(ctx context.Context, queueName string, jobID string) error {
	// 检查任务是否存在
	exists, err := r.client.Exists(ctx, jobDataKey(jobID)).Result()
	if err != nil {
		return fmt.Errorf("检查任务是否存在失败: %w", err)
	}
	if exists == 0 {
		return queue.ErrJobNotFound
	}

	// 使用Redis管道执行多个操作
	pipe := r.client.Pipeline()

	// 从所有可能的位置移除任务ID
	pipe.LRem(ctx, queueKey(queueName), 0, jobID)
	pipe.ZRem(ctx, scheduledSetKey(queueName), jobID)
	pipe.ZRem(ctx, processingSetKey(queueName), jobID)
	pipe.ZRem(ctx, completedSetKey(queueName), jobID)
	pipe.ZRem(ctx, failedSetKey(queueName), jobID)

	// 删除任务数据
	pipe.Del(ctx, jobDataKey(jobID))

	// 执行管道
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("从Redis删除任务失败: %w", err)
	}

	return nil
}

// Clear 清空队列
func (r *RedisQueue) Clear(ctx context.Context, queueName string) error {
	// 获取所有相关的键
	keys := []string{
		queueKey(queueName),
		scheduledSetKey(queueName),
		processingSetKey(queueName),
		completedSetKey(queueName),
		failedSetKey(queueName),
	}

	// 使用Redis管道执行多个操作
	pipe := r.client.Pipeline()

	// 获取队列中的所有任务ID
	pipe.LRange(ctx, queueKey(queueName), 0, -1)

	// 获取计划任务中的所有任务ID
	pipe.ZRange(ctx, scheduledSetKey(queueName), 0, -1)

	// 获取处理中任务的所有任务ID
	pipe.ZRange(ctx, processingSetKey(queueName), 0, -1)

	// 获取已完成任务的所有任务ID
	pipe.ZRange(ctx, completedSetKey(queueName), 0, -1)

	// 获取失败任务的所有任务ID
	pipe.ZRange(ctx, failedSetKey(queueName), 0, -1)

	// 执行管道
	results, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("获取任务ID失败: %w", err)
	}

	// 收集所有任务ID
	var jobIDs []string

	for i := 0; i < 5; i++ {
		if ids, ok := results[i].(*redis.StringSliceCmd); ok {
			jobIDs = append(jobIDs, ids.Val()...)
		}
	}

	// 删除所有队列相关的键
	if len(keys) > 0 {
		err = r.client.Del(ctx, keys...).Err()
		if err != nil {
			return fmt.Errorf("删除队列键失败: %w", err)
		}
	}

	// 删除所有任务数据
	if len(jobIDs) > 0 {
		// 构建所有任务数据键
		jobDataKeys := make([]string, len(jobIDs))
		for i, id := range jobIDs {
			jobDataKeys[i] = jobDataKey(id)
		}

		// 批量删除
		err = r.client.Del(ctx, jobDataKeys...).Err()
		if err != nil {
			return fmt.Errorf("删除任务数据失败: %w", err)
		}
	}

	return nil
}

// Size 获取队列大小
func (r *RedisQueue) Size(ctx context.Context, queueName string) (int, error) {
	// 使用Redis管道执行多个操作
	pipe := r.client.Pipeline()

	// 获取等待中的任务数量
	pipe.LLen(ctx, queueKey(queueName))

	// 获取计划中的任务数量
	pipe.ZCard(ctx, scheduledSetKey(queueName))

	// 获取处理中的任务数量
	pipe.ZCard(ctx, processingSetKey(queueName))

	// 执行管道
	results, err := pipe.Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("获取队列大小失败: %w", err)
	}

	// 计算总数
	var total int64
	for _, result := range results {
		switch cmd := result.(type) {
		case *redis.IntCmd:
			total += cmd.Val()
		}
	}

	return int(total), nil
}

// Register 注册任务处理器
func (r *RedisQueue) Register(jobName string, handler queue.Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.handlers[jobName] = handler
}

// ProcessNext 处理队列中的下一个任务
func (r *RedisQueue) ProcessNext(ctx context.Context, queueName string) error {
	// 1. 将到期的计划任务移动到主队列
	now := float64(time.Now().Unix())

	// 查找所有到期的计划任务
	jobIDs, err := r.client.ZRangeByScore(ctx, scheduledSetKey(queueName), &redis.ZRangeBy{
		Min: "0",
		Max: fmt.Sprintf("%f", now),
	}).Result()

	if err != nil && err != redis.Nil {
		return fmt.Errorf("获取到期计划任务失败: %w", err)
	}

	// 将到期任务移动到主队列
	if len(jobIDs) > 0 {
		pipe := r.client.Pipeline()

		// 从计划任务集合中删除
		pipe.ZRemRangeByScore(ctx, scheduledSetKey(queueName), "0", fmt.Sprintf("%f", now))

		// 更新任务状态并添加到主队列
		for _, jobID := range jobIDs {
			// 获取任务数据
			job, err := r.Get(ctx, queueName, jobID)
			if err != nil {
				continue // 跳过无法获取的任务
			}

			// 更新任务状态
			job.Status = queue.JobStatusPending
			job.UpdatedAt = time.Now()

			// 序列化任务数据
			jobData, err := json.Marshal(job)
			if err != nil {
				continue
			}

			// 更新任务数据并添加到主队列
			pipe.Set(ctx, jobDataKey(jobID), jobData, 7*24*time.Hour)
			pipe.LPush(ctx, queueKey(queueName), jobID)
		}

		// 执行管道
		_, err = pipe.Exec(ctx)
		if err != nil {
			return fmt.Errorf("处理到期计划任务失败: %w", err)
		}
	}

	// 2. 从主队列中获取下一个任务
	jobID, err := r.client.RPop(ctx, queueKey(queueName)).Result()
	if err != nil {
		if err == redis.Nil {
			// 队列为空
			return nil
		}
		return fmt.Errorf("从队列获取任务失败: %w", err)
	}

	// 获取任务数据
	job, err := r.Get(ctx, queueName, jobID)
	if err != nil {
		return fmt.Errorf("获取任务数据失败: %w", err)
	}

	// 查找任务处理器
	r.mu.RLock()
	handler, exists := r.handlers[job.Name]
	r.mu.RUnlock()

	if !exists {
		// 任务处理器不存在，将任务标记为失败
		job.Status = queue.JobStatusFailed
		job.Error = "任务处理器不存在"
		job.UpdatedAt = time.Now()

		// 序列化并保存任务数据
		jobData, _ := json.Marshal(job)

		// 添加到失败集合
		pipe := r.client.Pipeline()
		pipe.Set(ctx, jobDataKey(jobID), jobData, 7*24*time.Hour)
		pipe.ZAdd(ctx, failedSetKey(queueName), redis.Z{
			Score:  float64(time.Now().Unix()),
			Member: jobID,
		})
		_, err = pipe.Exec(ctx)

		return errors.New("任务处理器不存在")
	}

	// 更新任务状态为处理中
	job.Status = queue.JobStatusRunning
	job.Attempts++
	nowTime := time.Now()
	job.StartedAt = &nowTime
	job.UpdatedAt = nowTime

	// 序列化并保存任务数据
	jobData, _ := json.Marshal(job)

	pipe := r.client.Pipeline()
	pipe.Set(ctx, jobDataKey(jobID), jobData, 7*24*time.Hour)
	pipe.ZAdd(ctx, processingSetKey(queueName), redis.Z{
		Score:  float64(nowTime.Unix()),
		Member: jobID,
	})
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("更新任务状态失败: %w", err)
	}

	// 执行任务
	err = handler(ctx, job)

	// 更新任务状态
	finishTime := time.Now()
	job.UpdatedAt = finishTime
	job.FinishedAt = &finishTime

	if err != nil {
		// 任务执行失败
		job.Status = queue.JobStatusFailed
		job.Error = err.Error()

		if job.Attempts < job.MaxRetries {
			// 还可以重试，将任务状态改为等待重试
			job.Status = queue.JobStatusRetrying

			// 稍后重试，添加到主队列
			pipe := r.client.Pipeline()

			// 更新任务数据
			jobData, _ = json.Marshal(job)
			pipe.Set(ctx, jobDataKey(jobID), jobData, 7*24*time.Hour)

			// 从处理中集合移除
			pipe.ZRem(ctx, processingSetKey(queueName), jobID)

			// 添加回主队列，延迟重试时间随尝试次数增加
			delayTime := time.Duration(job.Attempts*5) * time.Second
			scheduleAt := time.Now().Add(delayTime)

			// 添加到计划任务集合
			pipe.ZAdd(ctx, scheduledSetKey(queueName), redis.Z{
				Score:  float64(scheduleAt.Unix()),
				Member: jobID,
			})

			_, err = pipe.Exec(ctx)
			if err != nil {
				return fmt.Errorf("安排任务重试失败: %w", err)
			}
		} else {
			// 不再重试，将任务标记为失败
			pipe := r.client.Pipeline()

			// 更新任务数据
			jobData, _ = json.Marshal(job)
			pipe.Set(ctx, jobDataKey(jobID), jobData, 7*24*time.Hour)

			// 从处理中集合移除
			pipe.ZRem(ctx, processingSetKey(queueName), jobID)

			// 添加到失败集合
			pipe.ZAdd(ctx, failedSetKey(queueName), redis.Z{
				Score:  float64(finishTime.Unix()),
				Member: jobID,
			})

			_, err = pipe.Exec(ctx)
			if err != nil {
				return fmt.Errorf("更新失败任务状态失败: %w", err)
			}
		}

		// 添加到已完成集合
		pipe.ZAdd(ctx, completedSetKey(queueName), redis.Z{
			Score:  float64(finishTime.Unix()),
			Member: jobID,
		})

		_, err = pipe.Exec(ctx)
		if err != nil {
			return fmt.Errorf("更新已完成任务状态失败: %w", err)
		}
	} else {
		// 任务执行成功
		job.Status = queue.JobStatusCompleted

		pipe := r.client.Pipeline()

		// 更新任务数据
		jobData, _ = json.Marshal(job)
		pipe.Set(ctx, jobDataKey(jobID), jobData, 7*24*time.Hour)

		// 从处理中集合移除
		pipe.ZRem(ctx, processingSetKey(queueName), jobID)

		// 添加到已完成集合
		pipe.ZAdd(ctx, completedSetKey(queueName), redis.Z{
			Score:  float64(finishTime.Unix()),
			Member: jobID,
		})

		_, err = pipe.Exec(ctx)
		if err != nil {
			return fmt.Errorf("更新已完成任务状态失败: %w", err)
		}
	}

	return nil
}

// StartWorker 启动工作进程
func (r *RedisQueue) StartWorker(ctx context.Context, queueName string, concurrency int) error {
	r.mu.Lock()

	// 如果已有工作进程在运行，先停止
	if cancel, exists := r.workerContexts[queueName]; exists {
		cancel()
	}

	// 创建新的上下文和取消函数
	workerCtx, cancel := context.WithCancel(ctx)
	r.workerContexts[queueName] = cancel

	r.mu.Unlock()

	// 启动工作进程
	for i := 0; i < concurrency; i++ {
		go func(workerID int) {
			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-workerCtx.Done():
					return
				case <-ticker.C:
					err := r.ProcessNext(workerCtx, queueName)
					if err != nil {
						log.Printf("工作进程 %d 处理任务失败: %v", workerID, err)
					}
				}
			}
		}(i)
	}

	return nil
}

// StopWorker 停止工作进程
func (r *RedisQueue) StopWorker(ctx context.Context, queueName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if cancel, exists := r.workerContexts[queueName]; exists {
		cancel()
		delete(r.workerContexts, queueName)
		return nil
	}

	return queue.ErrQueueNotFound
}

// Retry 重试失败的任务
func (r *RedisQueue) Retry(ctx context.Context, queueName string, jobID string) error {
	// 获取任务数据
	job, err := r.Get(ctx, queueName, jobID)
	if err != nil {
		return err
	}

	// 只能重试失败的任务
	if job.Status != queue.JobStatusFailed {
		return errors.New("只能重试失败的任务")
	}

	// 更新任务状态
	job.Status = queue.JobStatusPending
	job.Error = ""
	job.UpdatedAt = time.Now()

	// 序列化任务数据
	jobData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("序列化任务失败: %w", err)
	}

	// 使用Redis管道执行多个操作
	pipe := r.client.Pipeline()

	// 更新任务数据
	pipe.Set(ctx, jobDataKey(jobID), jobData, 7*24*time.Hour)

	// 从失败集合中移除
	pipe.ZRem(ctx, failedSetKey(queueName), jobID)

	// 添加到主队列
	pipe.LPush(ctx, queueKey(queueName), jobID)

	// 执行管道
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("重试任务失败: %w", err)
	}

	return nil
}

// Close 关闭Redis连接
func (r *RedisQueue) Close() error {
	// 停止所有工作进程
	r.mu.Lock()
	for _, cancel := range r.workerContexts {
		cancel()
	}
	r.workerContexts = make(map[string]context.CancelFunc)
	r.mu.Unlock()

	// 关闭Redis连接
	return r.client.Close()
}

// 辅助函数：构建队列键
func queueKey(queueName string) string {
	return queuePrefix + queueName
}

// 辅助函数：构建计划任务集合键
func scheduledSetKey(queueName string) string {
	return scheduledSetPrefix + queueName
}

// 辅助函数：构建任务数据键
func jobDataKey(jobID string) string {
	return jobDataPrefix + jobID
}

// 辅助函数：构建处理中任务集合键
func processingSetKey(queueName string) string {
	return processingSetPrefix + queueName
}

// 辅助函数：构建已完成任务集合键
func completedSetKey(queueName string) string {
	return completedSetPrefix + queueName
}

// 辅助函数：构建失败任务集合键
func failedSetKey(queueName string) string {
	return failedSetPrefix + queueName
}
