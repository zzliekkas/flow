package examples

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/zzliekkas/flow/event"
	"github.com/zzliekkas/flow/queue"
	"github.com/zzliekkas/flow/queue/memory"
)

// QueueExample 展示队列基本用法的示例
func QueueExample() {
	// 创建一个内存队列，设置最大重试次数为3
	memQueue := memory.New(3)

	// 创建一个队列管理器
	manager := queue.NewQueueManager()

	// 添加内存队列
	err := manager.AddQueue("default", memQueue)
	if err != nil {
		log.Fatalf("添加队列失败: %v", err)
	}

	// 注册任务处理器
	manager.Register("send_email", func(ctx context.Context, job *queue.Job) error {
		// 从任务负载中获取数据
		to, _ := job.GetPayloadValue("to")
		subject, _ := job.GetPayloadValue("subject")
		body, _ := job.GetPayloadValue("body")

		log.Printf("发送邮件: 收件人=%v, 主题=%v", to, subject)
		log.Printf("邮件内容: %v", body)

		// 模拟发送邮件
		time.Sleep(500 * time.Millisecond)

		return nil
	})

	// 创建一个任务负载
	payload := map[string]interface{}{
		"to":      "user@example.com",
		"subject": "队列示例",
		"body":    "这是一个任务队列示例。",
	}

	// 推送任务到队列
	jobID, err := manager.Push(context.Background(), "send_email", payload)
	if err != nil {
		log.Fatalf("推送任务失败: %v", err)
	}
	log.Printf("任务已推送, ID: %s", jobID)

	// 延迟5秒执行任务
	jobID, err = manager.PushWithDelay(context.Background(), "send_email", payload, 5*time.Second)
	if err != nil {
		log.Fatalf("推送延迟任务失败: %v", err)
	}
	log.Printf("延迟任务已推送, ID: %s", jobID)

	// 计划在特定时间执行任务
	scheduledTime := time.Now().Add(10 * time.Second)
	jobID, err = manager.Schedule(context.Background(), "send_email", payload, scheduledTime)
	if err != nil {
		log.Fatalf("计划任务失败: %v", err)
	}
	log.Printf("计划任务已推送, ID: %s, 执行时间: %v", jobID, scheduledTime)

	// 启动工作进程处理任务，允许3个并发工作进程
	queue, _ := manager.GetQueue("default")
	err = queue.StartWorker(context.Background(), "default", 3)
	if err != nil {
		log.Fatalf("启动工作进程失败: %v", err)
	}
	log.Println("工作进程已启动")

	// 保持主程序运行一段时间，以便处理任务
	time.Sleep(15 * time.Second)

	// 停止工作进程
	err = queue.StopWorker(context.Background(), "default")
	if err != nil {
		log.Fatalf("停止工作进程失败: %v", err)
	}
	log.Println("工作进程已停止")
}

// QueueMiddlewareExample 展示队列中间件的使用
func QueueMiddlewareExample() {
	// 创建一个内存队列，设置最大重试次数为3
	memQueue := memory.New(3)

	// 创建一个队列管理器
	manager := queue.NewQueueManager()

	// 添加内存队列
	err := manager.AddQueue("default", memQueue)
	if err != nil {
		log.Fatalf("添加队列失败: %v", err)
	}

	// 创建一个任务处理器
	sendEmailHandler := func(ctx context.Context, job *queue.Job) error {
		// 从任务负载中获取数据
		to, _ := job.GetPayloadValue("to")
		subject, _ := job.GetPayloadValue("subject")

		log.Printf("发送邮件: 收件人=%v, 主题=%v", to, subject)

		// 模拟发送邮件
		time.Sleep(500 * time.Millisecond)

		// 模拟随机失败
		if job.Attempts < 2 {
			return fmt.Errorf("模拟邮件发送失败")
		}

		return nil
	}

	// 使用中间件注册任务处理器
	manager.RegisterWithMiddleware("send_email",
		sendEmailHandler,
		queue.LoggingMiddleware(), // 添加日志中间件
		queue.RetryMiddleware(3),  // 添加重试中间件
	)

	// 创建一个任务负载
	payload := map[string]interface{}{
		"to":      "user@example.com",
		"subject": "中间件示例",
		"body":    "这是一个展示队列中间件的示例。",
	}

	// 推送任务到队列
	jobID, err := manager.Push(context.Background(), "send_email", payload)
	if err != nil {
		log.Fatalf("推送任务失败: %v", err)
	}
	log.Printf("任务已推送, ID: %s", jobID)

	// 启动工作进程处理任务
	queue, _ := manager.GetQueue("default")
	err = queue.StartWorker(context.Background(), "default", 1)
	if err != nil {
		log.Fatalf("启动工作进程失败: %v", err)
	}
	log.Println("工作进程已启动")

	// 保持主程序运行一段时间，以便处理任务和重试
	time.Sleep(10 * time.Second)

	// 停止工作进程
	err = queue.StopWorker(context.Background(), "default")
	if err != nil {
		log.Fatalf("停止工作进程失败: %v", err)
	}
	log.Println("工作进程已停止")
}

// QueueEventIntegrationExample 展示队列与事件系统集成
func QueueEventIntegrationExample() {
	// 创建事件分发器
	dispatcher := event.NewDispatcher(100)

	// 创建一个内存队列
	memQueue := memory.New(3)

	// 创建一个队列管理器
	manager := queue.NewQueueManager()

	// 添加内存队列
	err := manager.AddQueue("default", memQueue)
	if err != nil {
		log.Fatalf("添加队列失败: %v", err)
	}

	// 创建队列事件监听器
	queueListener := queue.NewQueueEventListener(manager, "default")

	// 注册事件到任务的映射
	queueListener.RegisterEventJob("user.registered", "send_welcome_email")
	queueListener.RegisterEventJob("order.created", "process_order")

	// 将监听器添加到事件分发器
	err = dispatcher.AddListener("user.registered", queueListener)
	if err != nil {
		log.Fatalf("添加监听器失败: %v", err)
	}
	err = dispatcher.AddListener("order.created", queueListener)
	if err != nil {
		log.Fatalf("添加监听器失败: %v", err)
	}

	// 注册任务处理器
	manager.RegisterWithMiddleware("send_welcome_email",
		func(ctx context.Context, job *queue.Job) error {
			email, _ := job.GetPayloadValue("email")
			name, _ := job.GetPayloadValue("name")
			log.Printf("发送欢迎邮件给 %s (%s)", name, email)
			return nil
		},
		queue.LoggingMiddleware(),
	)

	manager.RegisterWithMiddleware("process_order",
		func(ctx context.Context, job *queue.Job) error {
			orderID, _ := job.GetPayloadValue("order_id")
			amount, _ := job.GetPayloadValue("amount")
			log.Printf("处理订单 #%v, 金额: %v", orderID, amount)
			return nil
		},
		queue.LoggingMiddleware(),
	)

	// 启动工作进程处理任务
	queue, _ := manager.GetQueue("default")
	err = queue.StartWorker(context.Background(), "default", 2)
	if err != nil {
		log.Fatalf("启动工作进程失败: %v", err)
	}
	log.Println("工作进程已启动")

	// 创建并分发用户注册事件
	userEvent := event.NewBaseEvent("user.registered")
	userEvent.SetPayload(map[string]interface{}{
		"user_id": "12345",
		"name":    "张三",
		"email":   "zhang@example.com",
	})

	err = dispatcher.Dispatch(userEvent)
	if err != nil {
		log.Fatalf("分发事件失败: %v", err)
	}
	log.Println("用户注册事件已分发")

	// 创建并分发订单创建事件
	orderEvent := event.NewBaseEvent("order.created")
	orderEvent.SetPayload(map[string]interface{}{
		"order_id": "ORD-9876",
		"user_id":  "12345",
		"amount":   99.99,
		"items":    []string{"产品1", "产品2"},
	})

	err = dispatcher.Dispatch(orderEvent)
	if err != nil {
		log.Fatalf("分发事件失败: %v", err)
	}
	log.Println("订单创建事件已分发")

	// 保持主程序运行一段时间，以便处理任务
	time.Sleep(5 * time.Second)

	// 停止工作进程
	err = queue.StopWorker(context.Background(), "default")
	if err != nil {
		log.Fatalf("停止工作进程失败: %v", err)
	}
	log.Println("工作进程已停止")
}
