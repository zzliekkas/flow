package dev

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

// DebugLogger 调试日志记录器
type DebugLogger struct {
	// 配置
	config *Config

	// 日志输出
	writer io.Writer

	// 日志级别
	level LogLevel

	// 日志前缀
	prefix string

	// 路由信息记录
	routeLog *RouteLog

	// SQL查询记录
	sqlLog *SQLLog

	// HTTP记录
	httpLog *HTTPLog

	// 互斥锁
	mu sync.Mutex

	// 文件输出
	file *os.File

	// 是否启用控制台颜色
	colorEnabled bool
}

// LogLevel 日志级别
type LogLevel int

const (
	// DebugLevel 调试级别
	DebugLevel LogLevel = iota
	// InfoLevel 信息级别
	InfoLevel
	// WarnLevel 警告级别
	WarnLevel
	// ErrorLevel 错误级别
	ErrorLevel
	// FatalLevel 致命级别
	FatalLevel
)

// RouteLog 路由日志
type RouteLog struct {
	// 路由列表
	Routes []RouteInfo
	// 互斥锁
	mu sync.Mutex
}

// RouteInfo 路由信息
type RouteInfo struct {
	// 方法
	Method string
	// 路径
	Path string
	// 处理器
	Handler string
	// 中间件
	Middlewares []string
	// 分组
	Group string
}

// SQLLog SQL日志
type SQLLog struct {
	// SQL查询列表
	Queries []SQLInfo
	// 互斥锁
	mu sync.Mutex
	// 最大记录数
	maxQueries int
}

// SQLInfo SQL信息
type SQLInfo struct {
	// 查询语句
	Query string
	// 参数
	Args []interface{}
	// 执行时间
	Duration time.Duration
	// 发生时间
	Time time.Time
	// 调用栈
	Stack string
	// 发生错误
	Error error
}

// HTTPLog HTTP日志
type HTTPLog struct {
	// 请求记录
	Requests []HTTPInfo
	// 互斥锁
	mu sync.Mutex
	// 最大记录数
	maxRequests int
}

// HTTPInfo HTTP信息
type HTTPInfo struct {
	// 请求ID
	ID string
	// 请求方法
	Method string
	// 请求路径
	Path string
	// 请求头
	Headers http.Header
	// 请求体
	Body []byte
	// 状态码
	StatusCode int
	// 响应头
	ResponseHeaders http.Header
	// 响应体
	ResponseBody []byte
	// 开始时间
	StartTime time.Time
	// 结束时间
	EndTime time.Time
	// 持续时间
	Duration time.Duration
	// IP地址
	IP string
}

// NewDebugLogger 创建调试日志记录器
func NewDebugLogger(config *Config) *DebugLogger {
	if config == nil {
		config = NewConfig()
	}

	// 创建路由日志
	routeLog := &RouteLog{
		Routes: make([]RouteInfo, 0),
	}

	// 创建SQL日志
	sqlLog := &SQLLog{
		Queries:    make([]SQLInfo, 0),
		maxQueries: 1000,
	}

	// 创建HTTP日志
	httpLog := &HTTPLog{
		Requests:    make([]HTTPInfo, 0),
		maxRequests: 1000,
	}

	return &DebugLogger{
		config:       config,
		writer:       os.Stdout,
		level:        InfoLevel,
		prefix:       "[DEBUG] ",
		routeLog:     routeLog,
		sqlLog:       sqlLog,
		httpLog:      httpLog,
		colorEnabled: true,
	}
}

// SetOutput 设置日志输出
func (l *DebugLogger) SetOutput(w io.Writer) *DebugLogger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.writer = w
	return l
}

// SetLevel 设置日志级别
func (l *DebugLogger) SetLevel(level LogLevel) *DebugLogger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
	return l
}

// SetPrefix 设置日志前缀
func (l *DebugLogger) SetPrefix(prefix string) *DebugLogger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.prefix = prefix
	return l
}

// SetColor 设置是否启用控制台颜色
func (l *DebugLogger) SetColor(enabled bool) *DebugLogger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.colorEnabled = enabled
	return l
}

// SetLogFile 设置日志文件
func (l *DebugLogger) SetLogFile(filePath string) error {
	// 创建日志目录
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %w", err)
	}

	// 打开日志文件
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %w", err)
	}

	// 关闭之前的文件
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		l.file.Close()
	}

	l.file = file
	l.writer = io.MultiWriter(os.Stdout, file)
	return nil
}

// Close 关闭日志记录器
func (l *DebugLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		err := l.file.Close()
		l.file = nil
		return err
	}

	return nil
}

// Debug 记录调试级别日志
func (l *DebugLogger) Debug(v ...interface{}) {
	if l.level <= DebugLevel {
		l.writeLog(DebugLevel, fmt.Sprint(v...))
	}
}

// Debugf 记录调试级别日志（格式化）
func (l *DebugLogger) Debugf(format string, v ...interface{}) {
	if l.level <= DebugLevel {
		l.writeLog(DebugLevel, fmt.Sprintf(format, v...))
	}
}

// Info 记录信息级别日志
func (l *DebugLogger) Info(v ...interface{}) {
	if l.level <= InfoLevel {
		l.writeLog(InfoLevel, fmt.Sprint(v...))
	}
}

// Infof 记录信息级别日志（格式化）
func (l *DebugLogger) Infof(format string, v ...interface{}) {
	if l.level <= InfoLevel {
		l.writeLog(InfoLevel, fmt.Sprintf(format, v...))
	}
}

// Warn 记录警告级别日志
func (l *DebugLogger) Warn(v ...interface{}) {
	if l.level <= WarnLevel {
		l.writeLog(WarnLevel, fmt.Sprint(v...))
	}
}

// Warnf 记录警告级别日志（格式化）
func (l *DebugLogger) Warnf(format string, v ...interface{}) {
	if l.level <= WarnLevel {
		l.writeLog(WarnLevel, fmt.Sprintf(format, v...))
	}
}

// Error 记录错误级别日志
func (l *DebugLogger) Error(v ...interface{}) {
	if l.level <= ErrorLevel {
		l.writeLog(ErrorLevel, fmt.Sprint(v...))
	}
}

// Errorf 记录错误级别日志（格式化）
func (l *DebugLogger) Errorf(format string, v ...interface{}) {
	if l.level <= ErrorLevel {
		l.writeLog(ErrorLevel, fmt.Sprintf(format, v...))
	}
}

// Fatal 记录致命级别日志并退出
func (l *DebugLogger) Fatal(v ...interface{}) {
	if l.level <= FatalLevel {
		l.writeLog(FatalLevel, fmt.Sprint(v...))
	}
	os.Exit(1)
}

// Fatalf 记录致命级别日志并退出（格式化）
func (l *DebugLogger) Fatalf(format string, v ...interface{}) {
	if l.level <= FatalLevel {
		l.writeLog(FatalLevel, fmt.Sprintf(format, v...))
	}
	os.Exit(1)
}

// writeLog 输出日志
func (l *DebugLogger) writeLog(level LogLevel, message string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 获取时间
	now := time.Now().Format("2006-01-02 15:04:05.000")

	// 获取调用信息
	_, file, line, ok := runtime.Caller(2)
	fileInfo := "???"
	if ok {
		fileInfo = filepath.Base(file)
	}

	// 构造日志前缀
	var prefix string
	if l.colorEnabled && l.writer == os.Stdout {
		// 使用颜色
		var colorCode string
		switch level {
		case DebugLevel:
			colorCode = "\033[36m" // 青色
		case InfoLevel:
			colorCode = "\033[32m" // 绿色
		case WarnLevel:
			colorCode = "\033[33m" // 黄色
		case ErrorLevel:
			colorCode = "\033[31m" // 红色
		case FatalLevel:
			colorCode = "\033[35m" // 紫色
		}
		prefix = fmt.Sprintf("%s%s %s [%s:%d] %s\033[0m", colorCode, now, getLevelString(level), fileInfo, line, l.prefix)
	} else {
		prefix = fmt.Sprintf("%s %s [%s:%d] %s", now, getLevelString(level), fileInfo, line, l.prefix)
	}

	// 写入日志
	fmt.Fprintf(l.writer, "%s %s\n", prefix, message)
}

// getLevelString 获取日志级别字符串
func getLevelString(level LogLevel) string {
	switch level {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// AddRoute 添加路由信息
func (l *DebugLogger) AddRoute(route RouteInfo) {
	if !l.config.Debug.ShowRoutes {
		return
	}

	l.routeLog.mu.Lock()
	defer l.routeLog.mu.Unlock()
	l.routeLog.Routes = append(l.routeLog.Routes, route)

	if l.config.Debug.VerboseLogging {
		middlewares := "无"
		if len(route.Middlewares) > 0 {
			middlewares = strings.Join(route.Middlewares, ", ")
		}
		l.Infof("注册路由: %s %s -> %s (中间件: %s)", route.Method, route.Path, route.Handler, middlewares)
	}
}

// GetRoutes 获取所有路由信息
func (l *DebugLogger) GetRoutes() []RouteInfo {
	l.routeLog.mu.Lock()
	defer l.routeLog.mu.Unlock()
	routes := make([]RouteInfo, len(l.routeLog.Routes))
	copy(routes, l.routeLog.Routes)
	return routes
}

// LogSQL 记录SQL查询
func (l *DebugLogger) LogSQL(query string, args []interface{}, duration time.Duration, err error) {
	if !l.config.Debug.ShowSQL {
		return
	}

	// 获取调用栈信息
	stackTrace := string(debug.Stack())

	// 创建SQL信息
	info := SQLInfo{
		Query:    query,
		Args:     args,
		Duration: duration,
		Time:     time.Now(),
		Stack:    stackTrace,
		Error:    err,
	}

	// 添加到日志
	l.sqlLog.mu.Lock()
	defer l.sqlLog.mu.Unlock()

	l.sqlLog.Queries = append(l.sqlLog.Queries, info)
	if len(l.sqlLog.Queries) > l.sqlLog.maxQueries {
		l.sqlLog.Queries = l.sqlLog.Queries[1:]
	}

	// 如果启用了详细日志，打印SQL信息
	if l.config.Debug.VerboseLogging {
		statusMsg := "成功"
		if err != nil {
			statusMsg = fmt.Sprintf("失败: %v", err)
		}

		// 格式化参数
		argsStr := "[]"
		if len(args) > 0 {
			argsBytes, _ := json.Marshal(args)
			argsStr = string(argsBytes)
		}

		l.Debugf("SQL查询: %s %s (耗时: %s, 状态: %s)", query, argsStr, duration, statusMsg)
	}
}

// GetSQLQueries 获取SQL查询记录
func (l *DebugLogger) GetSQLQueries() []SQLInfo {
	l.sqlLog.mu.Lock()
	defer l.sqlLog.mu.Unlock()
	queries := make([]SQLInfo, len(l.sqlLog.Queries))
	copy(queries, l.sqlLog.Queries)
	return queries
}

// LogHTTPRequest 记录HTTP请求
func (l *DebugLogger) LogHTTPRequest(id, method, path string, headers http.Header, body []byte, ip string) {
	if !l.config.Debug.ShowHTTP {
		return
	}

	// 创建HTTP信息
	info := HTTPInfo{
		ID:        id,
		Method:    method,
		Path:      path,
		Headers:   headers,
		Body:      body,
		StartTime: time.Now(),
		IP:        ip,
	}

	// 添加到日志
	l.httpLog.mu.Lock()
	defer l.httpLog.mu.Unlock()

	l.httpLog.Requests = append(l.httpLog.Requests, info)
	if len(l.httpLog.Requests) > l.httpLog.maxRequests {
		l.httpLog.Requests = l.httpLog.Requests[1:]
	}

	// 如果启用了详细日志，打印请求信息
	if l.config.Debug.VerboseLogging {
		l.Debugf("收到请求: %s %s (ID: %s, IP: %s)", method, path, id, ip)
	}
}

// LogHTTPResponse 记录HTTP响应
func (l *DebugLogger) LogHTTPResponse(id string, statusCode int, headers http.Header, body []byte) {
	if !l.config.Debug.ShowHTTP {
		return
	}

	// 查找匹配的请求
	l.httpLog.mu.Lock()
	defer l.httpLog.mu.Unlock()

	// 查找并更新响应信息
	for i, req := range l.httpLog.Requests {
		if req.ID == id {
			endTime := time.Now()
			l.httpLog.Requests[i].StatusCode = statusCode
			l.httpLog.Requests[i].ResponseHeaders = headers
			l.httpLog.Requests[i].ResponseBody = body
			l.httpLog.Requests[i].EndTime = endTime
			l.httpLog.Requests[i].Duration = endTime.Sub(req.StartTime)

			// 如果启用了详细日志，打印响应信息
			if l.config.Debug.VerboseLogging {
				l.Debugf("响应请求: %s %s (ID: %s, 状态码: %d, 耗时: %s)",
					req.Method, req.Path, id, statusCode, endTime.Sub(req.StartTime))
			}
			break
		}
	}
}

// GetHTTPRequests 获取HTTP请求记录
func (l *DebugLogger) GetHTTPRequests() []HTTPInfo {
	l.httpLog.mu.Lock()
	defer l.httpLog.mu.Unlock()
	requests := make([]HTTPInfo, len(l.httpLog.Requests))
	copy(requests, l.httpLog.Requests)
	return requests
}

// DumpDebugInfo 导出调试信息
func (l *DebugLogger) DumpDebugInfo(w io.Writer) error {
	// 导出基本信息
	fmt.Fprintf(w, "=== Flow框架调试信息 ===\n")
	fmt.Fprintf(w, "时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(w, "Go版本: %s\n", runtime.Version())
	fmt.Fprintf(w, "操作系统: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Fprintf(w, "CPU核心数: %d\n", runtime.NumCPU())
	fmt.Fprintf(w, "当前协程数: %d\n", runtime.NumGoroutine())

	// 导出内存信息
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	fmt.Fprintf(w, "\n=== 内存信息 ===\n")
	fmt.Fprintf(w, "系统内存: %s\n", formatBytes(memStats.Sys))
	fmt.Fprintf(w, "堆内存: %s (已使用: %s)\n", formatBytes(memStats.HeapAlloc), formatBytes(memStats.HeapInuse))
	fmt.Fprintf(w, "栈内存: %s\n", formatBytes(memStats.StackInuse))
	fmt.Fprintf(w, "对象数量: %d\n", memStats.HeapObjects)
	fmt.Fprintf(w, "GC次数: %d\n", memStats.NumGC)
	fmt.Fprintf(w, "GC暂停时间: %s\n", time.Duration(memStats.PauseTotalNs)*time.Nanosecond)

	// 导出路由信息
	routes := l.GetRoutes()
	fmt.Fprintf(w, "\n=== 路由信息 (%d条) ===\n", len(routes))
	for i, route := range routes {
		middlewares := "无"
		if len(route.Middlewares) > 0 {
			middlewares = strings.Join(route.Middlewares, ", ")
		}
		fmt.Fprintf(w, "%d. %s %s -> %s (中间件: %s)\n", i+1, route.Method, route.Path, route.Handler, middlewares)
	}

	// 导出SQL查询信息
	queries := l.GetSQLQueries()
	fmt.Fprintf(w, "\n=== SQL查询 (最近%d条) ===\n", len(queries))
	for i, query := range queries {
		statusMsg := "成功"
		if query.Error != nil {
			statusMsg = fmt.Sprintf("失败: %v", query.Error)
		}

		// 格式化参数
		argsStr := "[]"
		if len(query.Args) > 0 {
			argsBytes, _ := json.Marshal(query.Args)
			argsStr = string(argsBytes)
		}

		fmt.Fprintf(w, "%d. [%s] %s %s (耗时: %s, 状态: %s)\n",
			i+1, query.Time.Format("15:04:05"), query.Query, argsStr, query.Duration, statusMsg)
	}

	// 导出HTTP请求信息
	requests := l.GetHTTPRequests()
	fmt.Fprintf(w, "\n=== HTTP请求 (最近%d条) ===\n", len(requests))
	for i, req := range requests {
		statusMsg := "处理中"
		if req.StatusCode > 0 {
			statusMsg = fmt.Sprintf("%d", req.StatusCode)
		}
		fmt.Fprintf(w, "%d. [%s] %s %s (ID: %s, IP: %s, 状态: %s, 耗时: %s)\n",
			i+1, req.StartTime.Format("15:04:05"), req.Method, req.Path, req.ID, req.IP, statusMsg, req.Duration)
	}

	return nil
}

// formatBytes 格式化字节大小
func formatBytes(bytes uint64) string {
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
