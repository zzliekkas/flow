package commands

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/zzliekkas/flow/cli"
)

// 初始化随机数种子
var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

// 仅用于示例的随机数生成器
var randomCounter int = 0

// 伪随机生成器结构，用于模拟功能
var randomGen struct {
	Intn    func(n int) int
	Float64 func() float64
}

func init() {
	// 在init函数中初始化函数，避免循环引用
	randomGen.Intn = func(n int) int {
		randomCounter++
		return (randomCounter * 17) % n
	}
	randomGen.Float64 = func() float64 {
		randomCounter++
		return float64(randomCounter%100) / 100.0
	}
}

// NewQueueCommand 创建队列管理命令
func NewQueueCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "queue",
		Aliases: []string{"queues", "q"},
		Short:   "管理后台队列和任务",
		Long:    `管理后台队列和异步任务，包括查看、监控、清理和重试任务。`,
	}

	// 添加子命令
	cmd.AddCommand(newQueueWorkCommand())
	cmd.AddCommand(newQueueListCommand())
	cmd.AddCommand(newQueueFailedCommand())
	cmd.AddCommand(newQueueRetryCommand())
	cmd.AddCommand(newQueueClearCommand())
	cmd.AddCommand(newQueueStatsCommand())

	return cmd
}

// newQueueWorkCommand 创建队列工作命令
func newQueueWorkCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "work",
		Aliases: []string{"worker", "listen"},
		Short:   "启动队列工作进程",
		Long:    `启动一个队列工作进程，处理队列中的任务。`,
		Run:     runQueueWorker,
	}

	cmd.Flags().StringP("connection", "c", "default", "队列连接名称")
	cmd.Flags().StringP("queue", "q", "", "要处理的队列名称")
	cmd.Flags().IntP("tries", "t", 3, "任务最大尝试次数")
	cmd.Flags().IntP("memory", "m", 128, "内存限制(MB)，超过此值将重启工作进程")
	cmd.Flags().IntP("timeout", "", 60, "任务执行超时时间(秒)")
	cmd.Flags().IntP("sleep", "s", 3, "队列为空时的睡眠时间(秒)")
	cmd.Flags().BoolP("daemon", "d", false, "作为守护进程运行")
	cmd.Flags().BoolP("force", "f", false, "即使队列中有任务在执行中也强制启动")

	return cmd
}

// newQueueListCommand 创建队列列表命令
func newQueueListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "jobs"},
		Short:   "查看队列中的任务",
		Long:    `查看队列中等待执行的任务列表。`,
		Run:     listQueueJobs,
	}

	cmd.Flags().StringP("connection", "c", "default", "队列连接名称")
	cmd.Flags().StringP("queue", "q", "", "要查看的队列名称")
	cmd.Flags().StringP("status", "s", "", "筛选任务状态 (waiting, reserved, failed, done)")
	cmd.Flags().IntP("limit", "l", 25, "显示的最大任务数量")
	cmd.Flags().BoolP("full", "f", false, "显示完整的任务信息")

	return cmd
}

// newQueueFailedCommand 创建失败任务命令
func newQueueFailedCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "failed",
		Aliases: []string{"failures", "f"},
		Short:   "查看失败的任务",
		Long:    `查看队列中执行失败的任务列表。`,
		Run:     listFailedJobs,
	}

	cmd.Flags().StringP("connection", "c", "default", "队列连接名称")
	cmd.Flags().StringP("queue", "q", "", "要查看的队列名称")
	cmd.Flags().StringP("error", "e", "", "按错误信息筛选")
	cmd.Flags().IntP("limit", "l", 25, "显示的最大任务数量")
	cmd.Flags().BoolP("full", "f", false, "显示完整的任务信息")

	return cmd
}

// newQueueRetryCommand 创建重试任务命令
func newQueueRetryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "retry [id]",
		Short: "重试失败的任务",
		Long:  `重试一个或所有失败的队列任务。`,
		Run:   retryFailedJobs,
	}

	cmd.Flags().StringP("connection", "c", "default", "队列连接名称")
	cmd.Flags().StringP("queue", "q", "", "队列名称")
	cmd.Flags().BoolP("all", "a", false, "重试所有失败的任务")

	return cmd
}

// newQueueClearCommand 创建清理队列命令
func newQueueClearCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "clear",
		Aliases: []string{"flush", "purge"},
		Short:   "清理队列任务",
		Long:    `清理队列中的任务，可清理全部、失败或等待中的任务。`,
		Run:     clearQueueJobs,
	}

	cmd.Flags().StringP("connection", "c", "default", "队列连接名称")
	cmd.Flags().StringP("queue", "q", "", "队列名称")
	cmd.Flags().StringP("status", "s", "failed", "要清理的任务状态 (all, failed, waiting, reserved)")
	cmd.Flags().BoolP("force", "f", false, "不提示确认直接清理")

	return cmd
}

// newQueueStatsCommand 创建队列统计命令
func newQueueStatsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "stats",
		Aliases: []string{"statistics", "info"},
		Short:   "显示队列统计信息",
		Long:    `显示队列的统计信息，包括任务数量、处理速率和错误率等。`,
		Run:     showQueueStats,
	}

	cmd.Flags().StringP("connection", "c", "default", "队列连接名称")
	cmd.Flags().StringP("queue", "q", "", "队列名称")
	cmd.Flags().StringP("period", "p", "hour", "统计周期 (hour, day, week, month)")
	cmd.Flags().BoolP("live", "l", false, "实时更新统计信息")
	cmd.Flags().StringP("format", "", "table", "输出格式 (table, json)")

	return cmd
}

// runQueueWorker 启动队列工作进程
func runQueueWorker(cmd *cobra.Command, args []string) {
	connection, _ := cmd.Flags().GetString("connection")
	queue, _ := cmd.Flags().GetString("queue")
	tries, _ := cmd.Flags().GetInt("tries")
	memoryLimit, _ := cmd.Flags().GetInt("memory")
	timeout, _ := cmd.Flags().GetInt("timeout")
	sleep, _ := cmd.Flags().GetInt("sleep")
	daemon, _ := cmd.Flags().GetBool("daemon")
	force, _ := cmd.Flags().GetBool("force")

	// 将force变量用于日志记录
	if force {
		cli.PrintInfo("强制模式已启用")
	}

	// 构建队列工作进程的描述
	description := fmt.Sprintf("队列工作进程 (连接: %s", connection)
	if queue != "" {
		description += fmt.Sprintf(", 队列: %s", queue)
	}
	description += ")"

	// 显示启动信息
	cli.PrintInfo("启动%s", description)
	cli.PrintInfo("最大尝试次数: %d, 超时: %d秒, 休眠间隔: %d秒", tries, timeout, sleep)
	cli.PrintInfo("内存限制: %dMB", memoryLimit)

	if daemon {
		cli.PrintInfo("作为守护进程运行")
		// 在实际应用中，这里应该将进程分离到后台
		// 对于示例，我们只是打印一条消息
		cli.PrintSuccess("守护进程已启动，使用 'ps' 命令查看进程状态")
		return
	}

	// 在实际应用中，此处应该连接到实际的队列系统并开始处理任务
	// 以下是一个示例实现，模拟处理任务

	// 监听中断信号以便优雅退出
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// 模拟任务处理
	go func() {
		jobCounter := 0
		for {
			// 模拟从队列获取任务
			time.Sleep(time.Duration(2+jobCounter%3) * time.Second)
			jobCounter++

			// 打印任务处理信息
			jobType := []string{"EmailJob", "ImageProcessJob", "ReportGenerationJob"}[jobCounter%3]
			jobId := fmt.Sprintf("job-%d", jobCounter)

			cli.PrintInfo("处理任务: %s [%s]", jobId, jobType)

			// 模拟任务处理
			processingTime := time.Duration((500+jobCounter*200)%2000) * time.Millisecond
			time.Sleep(processingTime)

			// 根据任务ID决定任务是否成功
			if jobCounter%7 == 0 {
				cli.PrintError("任务失败: %s (尝试: %d/%d) - 数据库连接错误", jobId, 1, tries)
			} else {
				cli.PrintSuccess("任务完成: %s (耗时: %v)", jobId, processingTime)
			}

			// 每5个任务显示一次队列统计
			if jobCounter%5 == 0 {
				waiting := 8 - jobCounter%7
				if waiting < 0 {
					waiting = 0
				}
				fmt.Println()
				cli.PrintInfo("队列状态: %d 个等待中, %d 个处理中, %d 个已完成, %d 个失败",
					waiting, 1, jobCounter-jobCounter/7, jobCounter/7)
				fmt.Println()
			}

			// 模拟队列空
			if jobCounter%10 == 9 {
				cli.PrintInfo("队列为空，等待 %d 秒...", sleep)
				time.Sleep(time.Duration(sleep) * time.Second)
			}
		}
	}()

	cli.PrintSuccess("工作进程已启动，按Ctrl+C停止")

	// 等待终止信号
	<-quit
	cli.PrintInfo("正在关闭工作进程...")
	time.Sleep(500 * time.Millisecond) // 给处理中的任务一点时间完成
	cli.PrintSuccess("工作进程已停止")
}

// listQueueJobs 列出队列任务
func listQueueJobs(cmd *cobra.Command, args []string) {
	connection, _ := cmd.Flags().GetString("connection")
	queue, _ := cmd.Flags().GetString("queue")
	status, _ := cmd.Flags().GetString("status")
	limit, _ := cmd.Flags().GetInt("limit")
	full, _ := cmd.Flags().GetBool("full")

	// 构建查询描述
	description := "任务"
	if status != "" {
		description = fmt.Sprintf("%s任务", status)
	}

	// 显示查询信息
	var queueStr string
	if queue != "" {
		queueStr = fmt.Sprintf("'%s'", queue)
	} else {
		queueStr = ""
	}
	cli.PrintInfo("查看%s队列中的%s (连接: %s, 限制: %d)",
		queueStr, description, connection, limit)

	// 在实际应用中，此处应该查询实际的队列系统
	// 以下是一个示例实现，返回模拟的任务列表
	jobs := generateSampleJobs(limit, status, queue)

	if len(jobs) == 0 {
		cli.PrintInfo("没有找到符合条件的任务")
		return
	}

	cli.PrintSuccess("找到 %d 个任务", len(jobs))
	fmt.Println()

	// 打印任务表头
	fmt.Println("ID\t类型\t\t队列\t状态\t\t尝试\t提交时间\t\t下次尝试")
	fmt.Println("--\t----\t\t----\t----\t\t----\t--------\t\t--------")

	// 打印任务列表
	for _, job := range jobs {
		nextAttempt := "-"
		if job.Status == "failed" && job.Attempts < 3 {
			nextAttempt = job.FailedAt.Add(time.Duration(job.Attempts*5) * time.Minute).Format("15:04:05")
		}

		fmt.Printf("%s\t%-20s\t%s\t%-10s\t%d/3\t%s\t%s\n",
			job.ID,
			job.Type,
			job.Queue,
			job.Status,
			job.Attempts,
			job.CreatedAt.Format("2006-01-02 15:04:05"),
			nextAttempt,
		)

		// 如果启用了完整模式，显示任务详情
		if full {
			fmt.Printf("  Payload: %s\n", job.Payload)
			if job.Status == "failed" {
				fmt.Printf("  Error: %s\n", job.Error)
			}
			fmt.Println()
		}
	}
}

// listFailedJobs 列出失败的任务
func listFailedJobs(cmd *cobra.Command, args []string) {
	connection, _ := cmd.Flags().GetString("connection")
	queue, _ := cmd.Flags().GetString("queue")
	errorFilter, _ := cmd.Flags().GetString("error")
	limit, _ := cmd.Flags().GetInt("limit")
	full, _ := cmd.Flags().GetBool("full")

	// 显示查询信息
	queueInfo := ""
	if queue != "" {
		queueInfo = fmt.Sprintf("队列: '%s', ", queue)
	}

	cli.PrintInfo("查看失败的任务 (%s连接: %s, 限制: %d)", queueInfo, connection, limit)
	if errorFilter != "" {
		cli.PrintInfo("按错误信息筛选: %s", errorFilter)
	}

	// 在实际应用中，此处应该查询实际的队列系统
	// 以下是一个示例实现，返回模拟的失败任务列表
	jobs := generateSampleJobs(limit, "failed", queue)

	// 按错误信息筛选
	if errorFilter != "" {
		var filteredJobs []queueJob
		for _, job := range jobs {
			if strings.Contains(strings.ToLower(job.Error), strings.ToLower(errorFilter)) {
				filteredJobs = append(filteredJobs, job)
			}
		}
		jobs = filteredJobs
	}

	if len(jobs) == 0 {
		cli.PrintInfo("没有找到失败的任务")
		return
	}

	cli.PrintSuccess("找到 %d 个失败的任务", len(jobs))
	fmt.Println()

	// 打印失败任务表头
	fmt.Println("ID\t类型\t\t队列\t尝试\t失败时间\t\t错误")
	fmt.Println("--\t----\t\t----\t----\t--------\t\t----")

	// 打印失败任务列表
	for _, job := range jobs {
		shortError := job.Error
		if len(shortError) > 40 && !full {
			shortError = shortError[:37] + "..."
		}

		fmt.Printf("%s\t%-20s\t%s\t%d/3\t%s\t%s\n",
			job.ID,
			job.Type,
			job.Queue,
			job.Attempts,
			job.FailedAt.Format("2006-01-02 15:04:05"),
			shortError,
		)

		// 如果启用了完整模式，显示任务详情
		if full {
			fmt.Printf("  完整错误: %s\n", job.Error)
			fmt.Printf("  Payload: %s\n", job.Payload)
			fmt.Println()
		}
	}

	// 显示重试提示
	fmt.Println()
	cli.PrintInfo("使用 'flow queue retry <id>' 重试特定任务或 'flow queue retry --all' 重试所有失败的任务")
}

// retryFailedJobs 重试失败的任务
func retryFailedJobs(cmd *cobra.Command, args []string) {
	connection, _ := cmd.Flags().GetString("connection")
	queue, _ := cmd.Flags().GetString("queue")
	all, _ := cmd.Flags().GetBool("all")

	// 检查参数
	if !all && len(args) == 0 {
		cli.PrintError("请指定任务ID或使用 --all 标志重试所有失败的任务")
		os.Exit(1)
	}

	// 显示操作信息
	if all {
		queueInfo := ""
		if queue != "" {
			queueInfo = fmt.Sprintf("队列 '%s' 中", queue)
		}
		cli.PrintInfo("正在重试%s所有失败的任务 (连接: %s)", queueInfo, connection)
	} else {
		cli.PrintInfo("正在重试任务 ID: %s (连接: %s)", args[0], connection)
	}

	// 在实际应用中，此处应该连接到实际的队列系统并重试任务
	// 以下是一个示例实现，模拟重试任务
	time.Sleep(500 * time.Millisecond)

	// 模拟重试结果
	if all {
		// 假设我们找到了5个失败的任务
		failedCount := 5
		cli.PrintSuccess("已将 %d 个失败的任务重新加入队列", failedCount)
	} else {
		jobId := args[0]
		// 随机决定是否找到了任务
		if jobId == "not-found" {
			cli.PrintError("找不到ID为 '%s' 的失败任务", jobId)
			os.Exit(1)
		}
		cli.PrintSuccess("任务 '%s' 已重新加入队列", jobId)
	}
}

// clearQueueJobs 清理队列任务
func clearQueueJobs(cmd *cobra.Command, args []string) {
	connection, _ := cmd.Flags().GetString("connection")
	queue, _ := cmd.Flags().GetString("queue")
	status, _ := cmd.Flags().GetString("status")
	force, _ := cmd.Flags().GetBool("force")

	// 显示确认提示
	description := "所有任务"
	if status == "failed" {
		description = "失败的任务"
	} else if status == "waiting" {
		description = "等待中的任务"
	}

	// 如果不是强制模式，显示确认提示
	if !force {
		// 临时替代cli.ConfirmAction
		fmt.Printf("确定要清除%s队列中的%s吗? (y/n): ", queue, description)
		var response string
		fmt.Scanln(&response)
		confirmed := strings.ToLower(response) == "y" || strings.ToLower(response) == "yes"

		if !confirmed {
			cli.PrintInfo("操作已取消")
			return
		}
	}

	// 构建清理描述
	statusDesc := "失败的"
	if status == "all" {
		statusDesc = "所有"
	} else if status == "waiting" {
		statusDesc = "等待中的"
	} else if status == "reserved" {
		statusDesc = "处理中的"
	}

	queueDesc := "所有队列"
	if queue != "" {
		queueDesc = fmt.Sprintf("队列 '%s'", queue)
	}

	// 显示操作信息
	cli.PrintInfo("正在清理%s%s任务 (连接: %s)", queueDesc, statusDesc, connection)

	// 在实际应用中，此处应该连接到实际的队列系统并清理任务
	// 以下是一个示例实现，模拟清理任务
	time.Sleep(500 * time.Millisecond)

	// 模拟清理结果
	clearedCount := 0
	if status == "all" {
		clearedCount = 25
	} else if status == "failed" {
		clearedCount = 8
	} else if status == "waiting" {
		clearedCount = 12
	} else if status == "reserved" {
		clearedCount = 5
	}

	cli.PrintSuccess("已清理 %d 个任务", clearedCount)
}

// showQueueStats 显示队列统计信息
func showQueueStats(cmd *cobra.Command, args []string) {
	connection, _ := cmd.Flags().GetString("connection")
	queue, _ := cmd.Flags().GetString("queue")
	period, _ := cmd.Flags().GetString("period")
	live, _ := cmd.Flags().GetBool("live")
	format, _ := cmd.Flags().GetString("format")

	// 构建描述
	queueDesc := "所有队列"
	if queue != "" {
		queueDesc = fmt.Sprintf("队列 '%s'", queue)
	}

	// 显示查询信息
	cli.PrintInfo("获取%s的统计信息 (连接: %s, 周期: %s)", queueDesc, connection, period)

	// 在实际应用中，此处应该查询实际的队列系统
	// 以下是一个示例实现，返回模拟的统计信息

	printQueueStats(queue, connection, format)

	// 如果是实时模式，每3秒更新一次统计信息
	if live {
		cli.PrintInfo("实时统计模式已启动，按Ctrl+C退出")

		// 监听中断信号以便优雅退出
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// 清除屏幕
				fmt.Print("\033[H\033[2J")

				// 重新显示标题
				cli.PrintInfo("获取%s的实时统计信息 (连接: %s, 周期: %s)", queueDesc, connection, period)
				cli.PrintInfo("每3秒自动更新，按Ctrl+C退出")
				fmt.Println()

				// 显示最新统计
				printQueueStats(queue, connection, format)

				// 显示更新时间
				fmt.Printf("\n最后更新: %s\n", time.Now().Format("15:04:05"))

			case <-quit:
				fmt.Println()
				cli.PrintInfo("实时统计已停止")
				return
			}
		}
	}
}

// 用于测试的队列任务结构
type queueJob struct {
	ID        string
	Type      string
	Queue     string
	Status    string
	Payload   string
	Attempts  int
	CreatedAt time.Time
	FailedAt  time.Time
	Error     string
}

// 生成样本任务用于展示
func generateSampleJobs(count int, status string, queue string) []queueJob {
	jobs := []queueJob{}

	jobTypes := []string{
		"App\\Jobs\\SendEmailJob",
		"App\\Jobs\\ProcessPaymentJob",
		"App\\Jobs\\GenerateReportJob",
		"App\\Jobs\\ImageProcessingJob",
		"App\\Jobs\\NotificationJob",
	}

	queueNames := []string{"default", "emails", "processing", "reports"}
	if queue != "" {
		queueNames = []string{queue}
	}

	errorMessages := []string{
		"连接到数据库时超时",
		"无法发送电子邮件: SMTP连接失败",
		"外部API返回错误: 503 Service Unavailable",
		"数据验证失败: 缺少必需字段",
		"处理图像时内存不足",
	}

	statuses := []string{"waiting", "reserved", "failed", "done"}
	if status != "" {
		if status == "all" {
			// 保持所有状态
		} else {
			statuses = []string{status}
		}
	}

	now := time.Now()

	for i := 0; i < count; i++ {
		jobId := fmt.Sprintf("job-%d", i+1)
		jobType := jobTypes[i%len(jobTypes)]
		jobQueue := queueNames[i%len(queueNames)]
		jobStatus := statuses[i%len(statuses)]

		createdAt := now.Add(-time.Duration(i*10+30) * time.Minute)
		failedAt := time.Time{}
		var attempts int
		var errorMsg string

		if jobStatus == "failed" {
			failedAt = now.Add(-time.Duration(i*5+10) * time.Minute)
			attempts = (i % 3) + 1
			errorMsg = errorMessages[i%len(errorMessages)]
		} else if jobStatus == "done" {
			attempts = 1
		} else if jobStatus == "reserved" {
			attempts = 1
		}

		payload := fmt.Sprintf(`{"id":%d,"type":"%s","data":{"param1":"value%d","param2":%d}}`,
			i+1, strings.Replace(jobType, "App\\Jobs\\", "", 1), i+1, (i+1)*10)

		job := queueJob{
			ID:        jobId,
			Type:      jobType,
			Queue:     jobQueue,
			Status:    jobStatus,
			Payload:   payload,
			Attempts:  attempts,
			CreatedAt: createdAt,
			FailedAt:  failedAt,
			Error:     errorMsg,
		}

		jobs = append(jobs, job)
	}

	return jobs
}

// 打印队列统计信息
func printQueueStats(queue, connection, format string) {
	// 模拟统计数据
	stats := struct {
		Waiting    int
		Reserved   int
		Failed     int
		Completed  int
		Total      int
		AvgTime    float64
		ErrorRate  float64
		Throughput int
	}{
		Waiting:    12 + rand.Intn(5),
		Reserved:   2 + rand.Intn(3),
		Failed:     8 + rand.Intn(4),
		Completed:  120 + rand.Intn(20),
		AvgTime:    234.5 + float64(rand.Intn(100))/10.0,
		ErrorRate:  5.8 + float64(rand.Intn(30))/10.0,
		Throughput: 42 + rand.Intn(8),
	}
	stats.Total = stats.Waiting + stats.Reserved + stats.Failed + stats.Completed

	// 按队列划分的细节
	queueDetails := map[string]struct {
		Waiting   int
		Reserved  int
		Failed    int
		Completed int
		AvgTime   float64
	}{
		"default": {
			Waiting:   5 + rand.Intn(3),
			Reserved:  1,
			Failed:    3 + rand.Intn(2),
			Completed: 60 + rand.Intn(10),
			AvgTime:   215.3 + float64(rand.Intn(50))/10.0,
		},
		"emails": {
			Waiting:   4 + rand.Intn(2),
			Reserved:  0,
			Failed:    2 + rand.Intn(2),
			Completed: 35 + rand.Intn(8),
			AvgTime:   189.7 + float64(rand.Intn(40))/10.0,
		},
		"processing": {
			Waiting:   3 + rand.Intn(3),
			Reserved:  1,
			Failed:    3 + rand.Intn(1),
			Completed: 25 + rand.Intn(6),
			AvgTime:   312.5 + float64(rand.Intn(80))/10.0,
		},
	}

	// 打印总体统计
	fmt.Println("队列统计概览:")
	fmt.Println("---------------------")
	fmt.Printf("总任务数: %d\n", stats.Total)
	fmt.Printf("等待中: %d\n", stats.Waiting)
	fmt.Printf("处理中: %d\n", stats.Reserved)
	fmt.Printf("已完成: %d\n", stats.Completed)
	fmt.Printf("失败数: %d\n", stats.Failed)
	fmt.Printf("错误率: %.1f%%\n", stats.ErrorRate)
	fmt.Printf("平均处理时间: %.1f ms\n", stats.AvgTime)
	fmt.Printf("每分钟处理量: %d 任务/分钟\n", stats.Throughput)

	// 如果没有指定特定队列，显示队列明细
	if queue == "" {
		fmt.Println("\n按队列统计:")
		fmt.Println("队列名称\t等待中\t处理中\t已完成\t失败\t平均时间(ms)")
		fmt.Println("--------\t------\t------\t------\t----\t-----------")

		for qName, qStat := range queueDetails {
			fmt.Printf("%s\t%d\t%d\t%d\t%d\t%.1f\n",
				qName,
				qStat.Waiting,
				qStat.Reserved,
				qStat.Completed,
				qStat.Failed,
				qStat.AvgTime,
			)
		}
	}

	// 按任务类型统计
	fmt.Println("\n按任务类型统计:")
	fmt.Println("任务类型\t\t\t总数\t成功率\t平均时间(ms)")
	fmt.Println("--------\t\t\t----\t------\t-----------")
	fmt.Printf("SendEmailJob\t\t\t%d\t%.1f%%\t%.1f\n",
		42+rand.Intn(10), 94.5+float64(rand.Intn(50))/10.0, 186.2+float64(rand.Intn(40)))
	fmt.Printf("ProcessPaymentJob\t\t%d\t%.1f%%\t%.1f\n",
		28+rand.Intn(8), 92.3+float64(rand.Intn(70))/10.0, 234.5+float64(rand.Intn(50)))
	fmt.Printf("GenerateReportJob\t\t%d\t%.1f%%\t%.1f\n",
		18+rand.Intn(5), 98.1+float64(rand.Intn(20))/10.0, 345.7+float64(rand.Intn(100)))
	fmt.Printf("ImageProcessingJob\t\t%d\t%.1f%%\t%.1f\n",
		35+rand.Intn(10), 91.2+float64(rand.Intn(50))/10.0, 512.8+float64(rand.Intn(150)))
	fmt.Printf("NotificationJob\t\t\t%d\t%.1f%%\t%.1f\n",
		52+rand.Intn(15), 99.2+float64(rand.Intn(10))/10.0, 156.3+float64(rand.Intn(30)))
}
