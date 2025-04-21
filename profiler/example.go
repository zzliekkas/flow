package profiler

import (
	"fmt"
	"math/rand"
	"time"
)

// Example 性能分析模块使用示例
func Example() {
	// 创建一个性能分析器
	profiler := NewProfiler(&ProfilerConfig{
		OutputDir:       "./profiles",
		EnableHTTP:      true,
		HTTPAddr:        "localhost:6060",
		EnableMemory:    true,
		EnableCPU:       true,
		EnableBlock:     true,
		EnableGoroutine: true,
		Rate:            1,
	})

	// 注册HTTP处理器
	profiler.RegisterHTTPHandlers()
	fmt.Println("性能分析HTTP服务已启动: http://localhost:6060/debug/profiler")

	// 启动性能分析
	if err := profiler.Start(); err != nil {
		fmt.Printf("启动性能分析失败: %v\n", err)
		return
	}

	// 在程序结束时停止分析
	defer func() {
		result, err := profiler.Stop()
		if err != nil {
			fmt.Printf("停止性能分析失败: %v\n", err)
			return
		}
		fmt.Printf("性能分析已完成，结果保存在: %s\n", result.FilePath)
		if result.MemStats != nil {
			fmt.Println(profiler.FormatMemoryStats(result.MemStats))
		}
	}()

	// 模拟一些负载
	fmt.Println("开始执行一些工作...")

	// 1. CPU密集型操作
	fmt.Println("执行CPU密集型操作...")
	go cpuIntensiveTask(30 * time.Second)

	// 2. 内存分配
	fmt.Println("执行内存分配操作...")
	go memoryIntensiveTask(30 * time.Second)

	// 3. 创建大量协程
	fmt.Println("创建大量协程...")
	go createGoroutines(100, 30*time.Second)

	// 等待足够的时间来收集性能数据
	fmt.Println("正在收集性能数据，请稍候...")
	time.Sleep(60 * time.Second)

	fmt.Println("示例完成!")
}

// cpuIntensiveTask 模拟CPU密集型任务
func cpuIntensiveTask(duration time.Duration) {
	end := time.Now().Add(duration)
	for time.Now().Before(end) {
		// 计算密集型操作
		for i := 0; i < 1000000; i++ {
			_ = rand.Float64() * rand.Float64()
		}
		time.Sleep(10 * time.Millisecond) // 稍微休息一下，避免CPU占用过高
	}
}

// memoryIntensiveTask 模拟内存密集型任务
func memoryIntensiveTask(duration time.Duration) {
	end := time.Now().Add(duration)
	var memoryHogs [][]byte

	for time.Now().Before(end) {
		// 分配内存
		data := make([]byte, 1024*1024*10) // 10MB
		for i := range data {
			data[i] = byte(rand.Intn(256))
		}
		memoryHogs = append(memoryHogs, data)

		// 释放一些内存以防止OOM
		if len(memoryHogs) > 10 {
			memoryHogs = memoryHogs[1:]
		}

		time.Sleep(500 * time.Millisecond)
	}
}

// createGoroutines 创建大量协程
func createGoroutines(count int, duration time.Duration) {
	end := time.Now().Add(duration)
	for time.Now().Before(end) {
		for i := 0; i < count; i++ {
			go func(id int) {
				// 协程内部做一些工作
				time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
			}(i)
		}
		time.Sleep(1 * time.Second)
	}
}
