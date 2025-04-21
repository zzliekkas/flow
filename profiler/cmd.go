package profiler

import (
	"fmt"
	"os"
	"time"
)

// ProfileCmd 性能分析命令
type ProfileCmd struct {
	// 分析器实例
	profiler *Profiler

	// 命令行参数
	outputDir string
	duration  time.Duration
	cpuOnly   bool
	memOnly   bool
	http      bool
	httpAddr  string
}

// NewProfileCmd 创建新的性能分析命令
func NewProfileCmd() *ProfileCmd {
	return &ProfileCmd{
		outputDir: "./profiles",
		duration:  30 * time.Second,
		cpuOnly:   false,
		memOnly:   false,
		http:      true,
		httpAddr:  "localhost:6060",
	}
}

// SetOutputDir 设置输出目录
func (c *ProfileCmd) SetOutputDir(dir string) *ProfileCmd {
	c.outputDir = dir
	return c
}

// SetDuration 设置分析持续时间
func (c *ProfileCmd) SetDuration(d time.Duration) *ProfileCmd {
	c.duration = d
	return c
}

// SetCPUOnly 设置仅进行CPU分析
func (c *ProfileCmd) SetCPUOnly(cpuOnly bool) *ProfileCmd {
	c.cpuOnly = cpuOnly
	return c
}

// SetMemOnly 设置仅进行内存分析
func (c *ProfileCmd) SetMemOnly(memOnly bool) *ProfileCmd {
	c.memOnly = memOnly
	return c
}

// SetHTTP 设置是否启用HTTP服务
func (c *ProfileCmd) SetHTTP(http bool) *ProfileCmd {
	c.http = http
	return c
}

// SetHTTPAddr 设置HTTP服务地址
func (c *ProfileCmd) SetHTTPAddr(addr string) *ProfileCmd {
	c.httpAddr = addr
	return c
}

// Run 运行性能分析命令
func (c *ProfileCmd) Run() error {
	// 创建配置
	config := &ProfilerConfig{
		OutputDir:       c.outputDir,
		EnableHTTP:      c.http,
		HTTPAddr:        c.httpAddr,
		EnableMemory:    !c.cpuOnly,
		EnableCPU:       !c.memOnly,
		EnableBlock:     !c.cpuOnly && !c.memOnly,
		EnableGoroutine: !c.cpuOnly && !c.memOnly,
		Rate:            1,
	}

	// 创建分析器
	c.profiler = NewProfiler(config)

	// 注册HTTP处理器
	if c.http {
		c.profiler.RegisterHTTPHandlers()
		fmt.Printf("HTTP性能分析服务已启动: http://%s/debug/profiler\n", c.httpAddr)
	}

	// 创建输出目录
	if err := os.MkdirAll(c.outputDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	// 启动分析
	if err := c.profiler.Start(); err != nil {
		return fmt.Errorf("启动性能分析失败: %w", err)
	}

	fmt.Printf("性能分析已启动，持续 %s...\n", c.duration)

	// 如果设置了持续时间，则在指定时间后停止分析
	if c.duration > 0 {
		time.Sleep(c.duration)
		result, err := c.profiler.Stop()
		if err != nil {
			return fmt.Errorf("停止性能分析失败: %w", err)
		}

		fmt.Printf("性能分析已完成，结果保存在: %s\n", result.FilePath)
		if result.MemStats != nil {
			fmt.Println(c.profiler.FormatMemoryStats(result.MemStats))
		}
	} else {
		fmt.Println("性能分析将持续运行，直到程序退出...")
		// 在程序退出时停止分析
		defer func() {
			_, _ = c.profiler.Stop()
		}()
	}

	return nil
}

// GetProfiler 获取分析器实例
func (c *ProfileCmd) GetProfiler() *Profiler {
	return c.profiler
}
