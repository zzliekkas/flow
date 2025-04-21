package examples

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/zzliekkas/flow/queue"
	qredis "github.com/zzliekkas/flow/queue/redis"
)

// RedisQueueExample 演示Redis队列的基本用法
func RedisQueueExample() {
	// 创建Redis队列
	options := qredis.DefaultOptions()
	options.Addr = "localhost:6379"
	redisQueue, err := qredis.New(options)
	if err != nil {
		log.Fatalf("创建Redis队列失败: %v", err)
	}
	defer redisQueue.Close()

	// 创建队列管理器
	manager := queue.NewQueueManager()
	manager.AddQueue("emails", redisQueue)

	// 注册任务处理器
	redisQueue.Register("send_email", func(ctx context.Context, job *queue.Job) error {
		to, _ := job.GetPayloadValue("to")
		subject, _ := job.GetPayloadValue("subject")
		body, _ := job.GetPayloadValue("body")

		log.Printf("发送邮件: 收件人=%v, 主题=%v", to, subject)
		log.Printf("邮件内容: %v", body)

		// 模拟发送邮件的延迟
		time.Sleep(500 * time.Millisecond)
		return nil
	})

	// 注册处理延迟任务的处理器
	redisQueue.Register("send_delayed_email", func(ctx context.Context, job *queue.Job) error {
		to, _ := job.GetPayloadValue("to")
		subject, _ := job.GetPayloadValue("subject")
		scheduledTime, _ := job.GetPayloadValue("scheduled_time")

		log.Printf("发送延迟邮件: 收件人=%v, 主题=%v, 计划时间=%v", to, subject, scheduledTime)
		return nil
	})

	// 启动工作进程
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		if err := redisQueue.StartWorker(ctx, "emails", 3); err != nil {
			log.Printf("启动工作进程失败: %v", err)
		}
	}()

	// 推送任务到队列
	for i := 0; i < 5; i++ {
		payload := map[string]interface{}{
			"to":   fmt.Sprintf("user%d@example.com", i+1),
			"body": fmt.Sprintf("这是第 %d 封测试邮件", i+1),
		}

		payload["subject"] = fmt.Sprintf("Redis队列示例 #%d", i+1)
		var jobID string
		var err error
		jobID, err = redisQueue.Push(ctx, "emails", "send_email", payload)
		if err != nil {
			log.Fatalf("推送任务失败: %v", err)
		}
		log.Printf("任务 %s 已创建", jobID)
	}

	// 推送延迟任务
	delayedPayload := map[string]interface{}{
		"to":             "delayed@example.com",
		"subject":        "这是一封延迟邮件",
		"body":           "这封邮件将在10秒后发送",
		"scheduled_time": time.Now().Add(10 * time.Second).Format(time.RFC3339),
	}

	jobID, err := redisQueue.PushWithDelay(ctx, "emails", "send_delayed_email", delayedPayload, 10*time.Second)
	if err != nil {
		log.Fatalf("推送延迟任务失败: %v", err)
	}
	log.Printf("延迟任务 %s 已创建", jobID)

	// 等待任务完成
	time.Sleep(15 * time.Second)
}

// DistributedWorkerExample 演示分布式工作进程
func DistributedWorkerExample() {
	// 创建Redis队列
	options := qredis.DefaultOptions()
	options.Addr = "localhost:6379"
	redisQueue, err := qredis.New(options)
	if err != nil {
		log.Fatalf("创建Redis队列失败: %v", err)
	}
	defer redisQueue.Close()

	// 注册订单处理器，模拟失败
	redisQueue.Register("process_order", func(ctx context.Context, job *queue.Job) error {
		orderID, _ := job.GetPayloadValue("order_id")
		amount, _ := job.GetPayloadValue("amount")
		log.Printf("处理订单 #%v, 金额: %v, 尝试次数: %d", orderID, amount, job.Attempts)

		// 随机模拟失败 (50% 概率)
		if rand.Float32() < 0.5 && job.Attempts < 3 {
			return fmt.Errorf("模拟订单处理失败")
		}

		log.Printf("订单 #%v 处理成功!", orderID)
		return nil
	})

	// 启动工作进程
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err = redisQueue.StartWorker(ctx, "orders", 2)
	if err != nil {
		log.Fatalf("启动工作进程失败: %v", err)
	}

	// 保持工作进程运行
	log.Println("工作进程已启动，等待处理任务...")
	select {}
}

// OrderProducerExample 演示订单生产者
func OrderProducerExample() {
	// 创建Redis队列
	options := qredis.DefaultOptions()
	options.Addr = "localhost:6379"
	redisQueue, err := qredis.New(options)
	if err != nil {
		log.Fatalf("创建Redis队列失败: %v", err)
	}
	defer redisQueue.Close()

	// 初始化随机数生成器
	rand.Seed(time.Now().UnixNano())

	// 生成随机订单并推送到队列
	for i := 1; i <= 20; i++ {
		orderID := fmt.Sprintf("ORDER-%d", i)
		amount := 10.0 + rand.Float64()*990.0

		payload := map[string]interface{}{
			"order_id": orderID,
			"amount":   fmt.Sprintf("%.2f", amount),
			"customer": fmt.Sprintf("customer-%d", rand.Intn(100)),
			"items":    rand.Intn(10) + 1,
		}

		ctx := context.Background()
		jobID, err := redisQueue.Push(ctx, "orders", "process_order", payload)
		if err != nil {
			log.Printf("推送订单失败: %v", err)
			continue
		}

		log.Printf("订单 #%s 已提交, 任务 ID: %s", orderID, jobID)

		// 随机延迟
		time.Sleep(time.Duration(500+rand.Intn(2000)) * time.Millisecond)
	}
}

// MonitorQueueExample 监控队列状态
func MonitorQueueExample() {
	// 创建Redis队列
	options := qredis.DefaultOptions()
	options.Addr = "localhost:6379"
	redisQueue, err := qredis.New(options)
	if err != nil {
		log.Fatalf("创建Redis队列失败: %v", err)
	}
	defer redisQueue.Close()

	ctx := context.Background()

	// 每5秒监控一次队列状态
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 获取队列大小
			ordersSize, err := redisQueue.Size(ctx, "orders")
			if err != nil {
				log.Printf("获取orders队列大小失败: %v", err)
			}

			emailsSize, err := redisQueue.Size(ctx, "emails")
			if err != nil {
				log.Printf("获取emails队列大小失败: %v", err)
			}

			log.Printf("队列状态 - 订单队列: %d, 邮件队列: %d", ordersSize, emailsSize)
		}
	}
}
