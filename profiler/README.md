# Flow框架性能分析模块

这是Flow框架的性能分析模块，提供了全面的性能监控和分析功能，帮助开发者优化应用性能、排查性能问题。

## 功能特点

- **CPU分析**：监控和分析CPU使用情况，找出热点函数
- **内存分析**：跟踪内存分配和使用情况，找出内存泄漏
- **协程分析**：监控协程的创建和运行情况
- **阻塞分析**：发现可能的死锁和阻塞点
- **互斥锁分析**：检测互斥锁争用情况
- **实时监控**：通过Web界面实时查看应用性能指标
- **性能分析报告**：生成详细的性能分析报告

## 快速开始

### 在应用中集成

```go
package main

import (
    "time"
    
    "github.com/zzliekkas/flow/profiler"
)

func main() {
    // 创建性能分析器（使用默认配置）
    prof := profiler.NewProfiler(nil)
    
    // 启动性能分析
    prof.Start()
    
    // 在程序结束时停止分析
    defer func() {
        result, _ := prof.Stop()
        println("性能分析结果:", result.FilePath)
    }()
    
    // 您的应用代码...
    
    // 可以随时获取内存统计
    memStats := prof.GetMemoryStats()
    println(prof.FormatMemoryStats(memStats))
}
```

### 使用命令行工具

```go
package main

import (
    "time"
    
    "github.com/zzliekkas/flow/profiler"
)

func main() {
    // 创建性能分析命令
    cmd := profiler.NewProfileCmd()
    
    // 配置分析参数
    cmd.SetOutputDir("./profiles")
       .SetDuration(60 * time.Second)
       .SetHTTP(true)
       .SetHTTPAddr("localhost:6060")
    
    // 运行分析
    if err := cmd.Run(); err != nil {
        panic(err)
    }
    
    // 您的应用代码...
}
```

### 通过HTTP接口访问

启用HTTP服务后，可以通过以下URL访问性能分析功能：

- 仪表盘: http://localhost:6060/debug/profiler
- 内存统计: http://localhost:6060/debug/profiler/memory
- 协程统计: http://localhost:6060/debug/profiler/goroutine
- 获取快照: http://localhost:6060/debug/profiler/snapshot?type=heap

## 详细配置

### ProfilerConfig 配置项

```go
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
```

## 使用pprof工具分析结果

生成的profile文件可以使用Go自带的pprof工具进行分析：

```bash
# 分析CPU profile
go tool pprof -http=:8080 ./profiles/cpu-20230601-120000.pprof

# 分析内存profile
go tool pprof -http=:8080 ./profiles/memory-20230601-120000.pprof

# 分析协程profile
go tool pprof -http=:8080 ./profiles/goroutine-20230601-120000.pprof
```

## 最佳实践

1. **生产环境谨慎使用**：性能分析可能会对应用性能产生一定影响，生产环境应谨慎使用
2. **设置合理的采样率**：高采样率会提供更精确的数据，但也会带来更大的性能开销
3. **分析单一问题**：每次只关注一个性能问题，避免同时启用过多分析功能
4. **定期分析**：定期对应用进行性能分析，及早发现潜在问题
5. **与监控系统结合**：将性能分析与监控系统结合，在发现异常时自动触发分析

## 常见问题

**Q: 性能分析会影响应用性能吗？**

A: 会有一定影响，特别是CPU分析。建议在非关键环境测试，或者在生产环境短时间开启。

**Q: 如何分析分布式应用？**

A: 可以在每个节点单独启用分析器，然后汇总结果，或使用分布式跟踪工具如Jaeger结合使用。

**Q: 内存泄漏如何排查？**

A: 可以定期使用`TakeMemorySnapshot`获取内存快照，然后使用pprof比较不同时间点的内存状态。

## 后续规划

- [ ] 分布式性能分析支持
- [ ] 性能瓶颈自动识别
- [ ] 与APM系统集成
- [ ] 更友好的Web界面
- [ ] 自定义监控指标 