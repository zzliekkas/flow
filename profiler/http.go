package profiler

import (
	"fmt"
	"html/template"
	"net/http"
	httppprof "net/http/pprof"
	"os"
	"runtime"
	rtpprof "runtime/pprof"
	"time"
)

// RegisterHTTPHandlers 注册HTTP处理器
func (p *Profiler) RegisterHTTPHandlers() {
	// 注册pprof处理器
	http.HandleFunc("/debug/pprof/", httppprof.Index)
	http.HandleFunc("/debug/pprof/cmdline", httppprof.Cmdline)
	http.HandleFunc("/debug/pprof/profile", httppprof.Profile)
	http.HandleFunc("/debug/pprof/symbol", httppprof.Symbol)
	http.HandleFunc("/debug/pprof/trace", httppprof.Trace)

	// 注册自定义处理器
	http.HandleFunc("/debug/profiler", p.dashboardHandler)
	http.HandleFunc("/debug/profiler/memory", p.memoryHandler)
	http.HandleFunc("/debug/profiler/goroutine", p.goroutineHandler)
	http.HandleFunc("/debug/profiler/snapshot", p.snapshotHandler)
}

// dashboardHandler 处理仪表盘页面
func (p *Profiler) dashboardHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>性能分析仪表盘</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif; line-height: 1.5; padding: 20px; max-width: 1200px; margin: 0 auto; }
        h1 { color: #333; }
        .card { background: #f9f9f9; border-radius: 5px; padding: 15px; margin-bottom: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .stats { display: flex; flex-wrap: wrap; gap: 20px; }
        .stat-item { flex: 1; min-width: 200px; }
        .value { font-size: 24px; font-weight: bold; margin: 10px 0; }
        .label { color: #666; }
        .actions { margin-top: 20px; }
        .btn { display: inline-block; background: #4a90e2; color: white; padding: 8px 16px; border-radius: 4px; text-decoration: none; margin-right: 10px; }
        .btn:hover { background: #3a80d2; }
        .links { margin-top: 30px; }
        .links a { display: block; margin-bottom: 10px; }
        pre { background: #f1f1f1; padding: 15px; border-radius: 5px; overflow-x: auto; }
    </style>
</head>
<body>
    <h1>性能分析仪表盘</h1>
    
    <div class="card">
        <h2>系统信息</h2>
        <div class="stats">
            <div class="stat-item">
                <div class="label">Go版本</div>
                <div class="value">{{.GoVersion}}</div>
            </div>
            <div class="stat-item">
                <div class="label">操作系统</div>
                <div class="value">{{.GOOS}}/{{.GOARCH}}</div>
            </div>
            <div class="stat-item">
                <div class="label">CPU核心数</div>
                <div class="value">{{.NumCPU}}</div>
            </div>
            <div class="stat-item">
                <div class="label">当前协程数</div>
                <div class="value">{{.NumGoroutine}}</div>
            </div>
        </div>
    </div>
    
    <div class="card">
        <h2>内存使用</h2>
        <div class="stats">
            <div class="stat-item">
                <div class="label">系统内存</div>
                <div class="value">{{.Sys}}</div>
            </div>
            <div class="stat-item">
                <div class="label">堆内存分配</div>
                <div class="value">{{.HeapAlloc}}</div>
            </div>
            <div class="stat-item">
                <div class="label">堆内存使用</div>
                <div class="value">{{.HeapInUse}}</div>
            </div>
            <div class="stat-item">
                <div class="label">栈内存使用</div>
                <div class="value">{{.StackInUse}}</div>
            </div>
        </div>
        <div class="stats">
            <div class="stat-item">
                <div class="label">对象数量</div>
                <div class="value">{{.Objects}}</div>
            </div>
            <div class="stat-item">
                <div class="label">GC次数</div>
                <div class="value">{{.NumGC}}</div>
            </div>
            <div class="stat-item">
                <div class="label">GC暂停时间</div>
                <div class="value">{{.PauseTotal}}</div>
            </div>
            <div class="stat-item">
                <div class="label">最后GC时间</div>
                <div class="value">{{.LastGC}}</div>
            </div>
        </div>
    </div>
    
    <div class="actions">
        <a href="/debug/profiler/snapshot?type=heap" class="btn">内存快照</a>
        <a href="/debug/profiler/snapshot?type=goroutine" class="btn">协程快照</a>
        <a href="/debug/profiler/snapshot?type=block" class="btn">阻塞快照</a>
        <a href="/debug/profiler/snapshot?type=mutex" class="btn">互斥锁快照</a>
    </div>
    
    <div class="links">
        <h2>详细分析</h2>
        <a href="/debug/pprof/">Pprof索引</a>
        <a href="/debug/pprof/heap?debug=1">堆内存分析</a>
        <a href="/debug/pprof/goroutine?debug=1">协程分析</a>
        <a href="/debug/pprof/block?debug=1">阻塞分析</a>
        <a href="/debug/pprof/threadcreate?debug=1">线程创建分析</a>
        <a href="/debug/pprof/mutex?debug=1">互斥锁分析</a>
    </div>
    
    <div class="card">
        <h2>内存详情</h2>
        <pre>{{.MemDetails}}</pre>
    </div>
</body>
</html>
`

	// 获取内存统计
	memStats := p.GetMemoryStats()

	// 准备数据
	data := map[string]interface{}{
		"GoVersion":    runtime.Version(),
		"GOOS":         runtime.GOOS,
		"GOARCH":       runtime.GOARCH,
		"NumCPU":       runtime.NumCPU(),
		"NumGoroutine": runtime.NumGoroutine(),
		"Sys":          fmt.Sprintf("%s", FormatBytes(memStats.Sys)),
		"HeapAlloc":    fmt.Sprintf("%s", FormatBytes(memStats.HeapAlloc)),
		"HeapInUse":    fmt.Sprintf("%s", FormatBytes(memStats.HeapInuse)),
		"StackInUse":   fmt.Sprintf("%s", FormatBytes(memStats.StackInuse)),
		"Objects":      memStats.HeapObjects,
		"NumGC":        memStats.NumGC,
		"PauseTotal":   time.Duration(memStats.PauseTotalNs) * time.Nanosecond,
		"LastGC":       time.Unix(0, int64(memStats.LastGC)).Format("2006-01-02 15:04:05"),
		"MemDetails":   p.FormatMemoryStats(memStats),
	}

	// 渲染模板
	t, err := template.New("dashboard").Parse(tmpl)
	if err != nil {
		http.Error(w, fmt.Sprintf("模板解析错误: %v", err), http.StatusInternalServerError)
		return
	}

	if err := t.Execute(w, data); err != nil {
		http.Error(w, fmt.Sprintf("模板渲染错误: %v", err), http.StatusInternalServerError)
		return
	}
}

// memoryHandler 处理内存统计页面
func (p *Profiler) memoryHandler(w http.ResponseWriter, r *http.Request) {
	memStats := p.GetMemoryStats()
	fmt.Fprintf(w, "%s", p.FormatMemoryStats(memStats))
}

// goroutineHandler 处理协程统计页面
func (p *Profiler) goroutineHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s", p.GetGoroutineStats())
}

// snapshotHandler 处理快照请求
func (p *Profiler) snapshotHandler(w http.ResponseWriter, r *http.Request) {
	snapshotType := r.URL.Query().Get("type")
	if snapshotType == "" {
		snapshotType = "heap"
	}

	var filePath string
	var err error

	switch snapshotType {
	case "heap":
		filePath, err = p.TakeMemorySnapshot("manual")
	case "goroutine":
		filePath = fmt.Sprintf("%s/goroutine-%s.pprof", p.config.OutputDir, time.Now().Format("20060102-150405"))
		f, err := os.Create(filePath)
		if err == nil {
			defer f.Close()
			err = rtpprof.Lookup("goroutine").WriteTo(f, 0)
		}
	case "block":
		filePath = fmt.Sprintf("%s/block-%s.pprof", p.config.OutputDir, time.Now().Format("20060102-150405"))
		f, err := os.Create(filePath)
		if err == nil {
			defer f.Close()
			err = rtpprof.Lookup("block").WriteTo(f, 0)
		}
	case "mutex":
		filePath = fmt.Sprintf("%s/mutex-%s.pprof", p.config.OutputDir, time.Now().Format("20060102-150405"))
		f, err := os.Create(filePath)
		if err == nil {
			defer f.Close()
			err = rtpprof.Lookup("mutex").WriteTo(f, 0)
		}
	default:
		http.Error(w, fmt.Sprintf("不支持的快照类型: %s", snapshotType), http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("创建快照失败: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "快照已创建: %s", filePath)
}

// FormatBytes 格式化字节大小为易读形式
func FormatBytes(bytes uint64) string {
	const (
		kb = 1024
		mb = 1024 * kb
		gb = 1024 * mb
		tb = 1024 * gb
	)

	switch {
	case bytes >= tb:
		return fmt.Sprintf("%.2f TB", float64(bytes)/tb)
	case bytes >= gb:
		return fmt.Sprintf("%.2f GB", float64(bytes)/gb)
	case bytes >= mb:
		return fmt.Sprintf("%.2f MB", float64(bytes)/mb)
	case bytes >= kb:
		return fmt.Sprintf("%.2f KB", float64(bytes)/kb)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
