package profiler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/zzliekkas/flow/utils"
)

// Profiler 性能分析器
type Profiler struct {
	// 配置项
	config *ProfilerConfig

	// CPU profile文件
	cpuFile *os.File

	// 是否正在进行CPU分析
	isCPUProfiling bool

	// 分析开始时间
	startTime time.Time
}

// ProfilerConfig 性能分析器配置
type ProfilerConfig struct {
	// 输出目录
	OutputDir string

	// 是否启用HTTP分析服务
	EnableHTTP bool

	// HTTP服务地址
	HTTPAddr string

	// 是否启用内存分析
	EnableMemory bool

	// 是否启用CPU分析
	EnableCPU bool

	// 是否启用阻塞分析
	EnableBlock bool

	// 是否启用协程分析
	EnableGoroutine bool

	// 采样率
	Rate int
}

// ProfileResult 分析结果
type ProfileResult struct {
	// 分析类型
	Type string

	// 文件路径
	FilePath string

	// 分析时长
	Duration time.Duration

	// 分析开始时间
	StartTime time.Time

	// 分析结束时间
	EndTime time.Time

	// 内存统计
	MemStats *runtime.MemStats
}

// NewProfiler 创建新的性能分析器
func NewProfiler(config *ProfilerConfig) *Profiler {
	// 使用默认配置
	if config == nil {
		config = &ProfilerConfig{
			OutputDir:       "./profiles",
			EnableHTTP:      true,
			HTTPAddr:        "localhost:6060",
			EnableMemory:    true,
			EnableCPU:       true,
			EnableBlock:     true,
			EnableGoroutine: true,
			Rate:            1,
		}
	}

	// 确保输出目录存在
	if config.OutputDir != "" {
		os.MkdirAll(config.OutputDir, 0755)
	}

	return &Profiler{
		config: config,
	}
}

// Start 启动性能分析
func (p *Profiler) Start() error {
	p.startTime = time.Now()

	// 启用阻塞分析
	if p.config.EnableBlock {
		runtime.SetBlockProfileRate(p.config.Rate)
	}

	// 启动CPU分析
	if p.config.EnableCPU {
		if err := p.StartCPUProfile(); err != nil {
			return err
		}
	}

	// 启动HTTP服务
	if p.config.EnableHTTP {
		go func() {
			http.ListenAndServe(p.config.HTTPAddr, nil)
		}()
	}

	return nil
}

// Stop 停止性能分析
func (p *Profiler) Stop() (*ProfileResult, error) {
	endTime := time.Now()
	duration := endTime.Sub(p.startTime)

	// 停止CPU分析
	if p.config.EnableCPU && p.isCPUProfiling {
		pprof.StopCPUProfile()
		if p.cpuFile != nil {
			p.cpuFile.Close()
			p.cpuFile = nil
		}
		p.isCPUProfiling = false
	}

	// 进行内存分析
	var memStats runtime.MemStats
	var memProfilePath string
	if p.config.EnableMemory {
		runtime.ReadMemStats(&memStats)
		memProfilePath = filepath.Join(p.config.OutputDir, fmt.Sprintf("memory-%s.pprof", endTime.Format("20060102-150405")))
		f, err := os.Create(memProfilePath)
		if err != nil {
			return nil, fmt.Errorf("创建内存分析文件失败: %w", err)
		}
		defer f.Close()
		if err := pprof.WriteHeapProfile(f); err != nil {
			return nil, fmt.Errorf("写入内存分析数据失败: %w", err)
		}
	}

	// 进行协程分析
	var goroutineProfilePath string
	if p.config.EnableGoroutine {
		goroutineProfilePath = filepath.Join(p.config.OutputDir, fmt.Sprintf("goroutine-%s.pprof", endTime.Format("20060102-150405")))
		f, err := os.Create(goroutineProfilePath)
		if err != nil {
			return nil, fmt.Errorf("创建协程分析文件失败: %w", err)
		}
		defer f.Close()
		if err := pprof.Lookup("goroutine").WriteTo(f, 0); err != nil {
			return nil, fmt.Errorf("写入协程分析数据失败: %w", err)
		}
	}

	// 创建分析结果
	result := &ProfileResult{
		Type:      "full",
		FilePath:  p.config.OutputDir,
		Duration:  duration,
		StartTime: p.startTime,
		EndTime:   endTime,
	}

	if p.config.EnableMemory {
		result.MemStats = &memStats
	}

	return result, nil
}

// StartCPUProfile 启动CPU分析
func (p *Profiler) StartCPUProfile() error {
	if p.isCPUProfiling {
		return nil
	}

	cpuProfilePath := filepath.Join(p.config.OutputDir, fmt.Sprintf("cpu-%s.pprof", time.Now().Format("20060102-150405")))
	f, err := os.Create(cpuProfilePath)
	if err != nil {
		return fmt.Errorf("创建CPU分析文件失败: %w", err)
	}

	if err := pprof.StartCPUProfile(f); err != nil {
		f.Close()
		return fmt.Errorf("启动CPU分析失败: %w", err)
	}

	p.cpuFile = f
	p.isCPUProfiling = true
	return nil
}

// StopCPUProfile 停止CPU分析
func (p *Profiler) StopCPUProfile() {
	if p.isCPUProfiling {
		pprof.StopCPUProfile()
		if p.cpuFile != nil {
			p.cpuFile.Close()
			p.cpuFile = nil
		}
		p.isCPUProfiling = false
	}
}

// TakeMemorySnapshot 获取内存快照
func (p *Profiler) TakeMemorySnapshot(label string) (string, error) {
	snapshotPath := filepath.Join(p.config.OutputDir, fmt.Sprintf("memory-%s-%s.pprof", label, time.Now().Format("20060102-150405")))
	f, err := os.Create(snapshotPath)
	if err != nil {
		return "", fmt.Errorf("创建内存快照文件失败: %w", err)
	}
	defer f.Close()

	if err := pprof.WriteHeapProfile(f); err != nil {
		return "", fmt.Errorf("写入内存快照数据失败: %w", err)
	}

	return snapshotPath, nil
}

// GetMemoryStats 获取内存统计信息
func (p *Profiler) GetMemoryStats() *runtime.MemStats {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	return &memStats
}

// FormatMemoryStats 格式化内存统计信息
func (p *Profiler) FormatMemoryStats(stats *runtime.MemStats) string {
	if stats == nil {
		stats = p.GetMemoryStats()
	}

	return fmt.Sprintf(`内存统计:
  系统内存: %s
  堆内存: %s (已使用: %s)
  栈内存: %s
  对象数量: %d
  GC次数: %d
  GC暂停时间: %s
`,
		utils.FormatBytes(stats.Sys),
		utils.FormatBytes(stats.HeapAlloc),
		utils.FormatBytes(stats.HeapInuse),
		utils.FormatBytes(stats.StackInuse),
		stats.HeapObjects,
		stats.NumGC,
		time.Duration(stats.PauseTotalNs)*time.Nanosecond,
	)
}

// GetGoroutineStats 获取协程统计信息
func (p *Profiler) GetGoroutineStats() string {
	return fmt.Sprintf("协程数量: %d", runtime.NumGoroutine())
}
