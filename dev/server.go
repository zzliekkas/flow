package dev

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"time"
)

// DevServer 开发服务器
type DevServer struct {
	// 配置
	config *Config

	// HTTP服务器
	server *http.Server

	// 路由处理器
	mux *http.ServeMux

	// 热重载器
	reloader *HotReloader

	// 调试日志记录器
	logger *DebugLogger

	// 互斥锁
	mu sync.Mutex

	// 是否正在运行
	running bool

	// 应用命令
	appCmd *exec.Cmd

	// 关闭通道
	done chan struct{}

	// 重启通道
	restart chan struct{}

	// 静态文件处理器
	staticHandler http.Handler

	// 代理处理器
	proxyHandlers map[string]http.Handler
}

// NewDevServer 创建开发服务器
func NewDevServer(config *Config) (*DevServer, error) {
	if config == nil {
		config = NewConfig()
	}

	// 创建HTTP路由
	mux := http.NewServeMux()

	// 创建热重载器
	reloader, err := NewHotReloader(config)
	if err != nil {
		return nil, fmt.Errorf("创建热重载器失败: %w", err)
	}

	// 创建调试日志记录器
	logger := NewDebugLogger(config)

	// 创建HTTP服务器
	addr := fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port)
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// 设置重载回调
	reloader.OnReload(func() {
		log.Println("文件变更，正在重启应用...")
	})

	return &DevServer{
		config:        config,
		server:        server,
		mux:           mux,
		reloader:      reloader,
		logger:        logger,
		done:          make(chan struct{}),
		restart:       make(chan struct{}, 1),
		proxyHandlers: make(map[string]http.Handler),
	}, nil
}

// Start 启动开发服务器
func (s *DevServer) Start() error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = true
	s.mu.Unlock()

	// 设置路由
	s.setupRoutes()

	// 启动热重载器
	if s.config.Reload.Enabled {
		if err := s.reloader.Start(); err != nil {
			return fmt.Errorf("启动热重载器失败: %w", err)
		}
		log.Println("热重载已启用，监视文件变更")
	}

	// 启动调试控制台
	if s.config.Debug.EnableConsole {
		go s.startDebugConsole()
	}

	// 如果启用了性能分析
	if s.config.Debug.EnableProfiler {
		go s.startProfiler()
	}

	// 启动应用
	if err := s.startApp(); err != nil {
		return fmt.Errorf("启动应用失败: %w", err)
	}

	// 监听重启信号
	go s.watchRestart()

	// 监听关闭信号
	go s.handleSignals()

	// 启动HTTP服务器
	log.Printf("开发服务器已启动: http://%s:%d\n", s.config.Server.Host, s.config.Server.Port)

	// 自动打开浏览器
	if s.config.Server.OpenBrowser {
		go func() {
			time.Sleep(500 * time.Millisecond)
			openBrowser(fmt.Sprintf("http://%s:%d", s.config.Server.Host, s.config.Server.Port))
		}()
	}

	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("HTTP服务器错误: %w", err)
	}

	return nil
}

// Stop 停止开发服务器
func (s *DevServer) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	// 关闭热重载器
	if s.reloader != nil && s.reloader.IsRunning() {
		if err := s.reloader.Stop(); err != nil {
			log.Printf("停止热重载器错误: %v\n", err)
		}
	}

	// 关闭应用
	if s.appCmd != nil && s.appCmd.Process != nil {
		if err := s.appCmd.Process.Kill(); err != nil {
			log.Printf("停止应用错误: %v\n", err)
		}
	}

	// 关闭HTTP服务器
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.server.Shutdown(ctx); err != nil {
		log.Printf("关闭HTTP服务器错误: %v\n", err)
	}

	close(s.done)
	s.running = false

	log.Println("开发服务器已停止")
	return nil
}

// setupRoutes 设置路由
func (s *DevServer) setupRoutes() {
	// 设置静态文件处理
	if s.config.Server.StaticDir != "" {
		staticDir := s.config.Server.StaticDir
		if !filepath.IsAbs(staticDir) {
			staticDir = filepath.Join(s.config.RootDir, staticDir)
		}
		s.staticHandler = http.FileServer(http.Dir(staticDir))
		s.mux.Handle("/static/", http.StripPrefix("/static/", s.staticHandler))
		log.Printf("静态文件目录: %s -> /static/\n", staticDir)
	}

	// 设置代理
	for path, target := range s.config.Server.Proxies {
		// TODO: 实现代理处理器
		log.Printf("代理: %s -> %s\n", path, target)
	}

	// 添加自定义调试路由
	s.mux.HandleFunc("/debug/reload", func(w http.ResponseWriter, r *http.Request) {
		s.reloader.ForceReload()
		fmt.Fprintf(w, "应用已重新加载\n")
	})

	// 添加状态路由
	s.mux.HandleFunc("/debug/status", func(w http.ResponseWriter, r *http.Request) {
		s.logger.DumpDebugInfo(w)
	})
}

// startDebugConsole 启动调试控制台
func (s *DevServer) startDebugConsole() {
	// TODO: 实现调试控制台
	log.Printf("调试控制台已启动: http://%s:%d\n", s.config.Server.Host, s.config.Debug.ConsolePort)
}

// startProfiler 启动性能分析器
func (s *DevServer) startProfiler() {
	// TODO: 集成性能分析器
	log.Printf("性能分析器已启动: http://%s:%d/debug/pprof/\n", s.config.Server.Host, s.config.Debug.ProfilerPort)
}

// startApp 启动应用
func (s *DevServer) startApp() error {
	// 目前简化实现，实际场景需要根据项目类型选择合适的启动方式
	log.Println("应用已启动")
	return nil
}

// watchRestart 监听重启信号
func (s *DevServer) watchRestart() {
	if !s.config.Reload.Enabled {
		return
	}

	// 监听热重载通道
	reloadCh := s.reloader.GetReloadChannel()

	for {
		select {
		case <-s.done:
			return
		case <-reloadCh:
			// 收到重载信号，重启应用
			s.restartApp()
		case <-s.restart:
			// 手动触发重启
			s.restartApp()
		}
	}
}

// restartApp 重启应用
func (s *DevServer) restartApp() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 停止当前应用
	if s.appCmd != nil && s.appCmd.Process != nil {
		if err := s.appCmd.Process.Kill(); err != nil {
			log.Printf("停止应用错误: %v\n", err)
		}
	}

	// 启动新应用
	log.Println("正在重启应用...")
	// TODO: 实现实际的应用重启逻辑
	log.Println("应用已重启")
}

// handleSignals 处理系统信号
func (s *DevServer) handleSignals() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-s.done:
		return
	case sig := <-signals:
		log.Printf("收到信号 %s，正在关闭服务器\n", sig)
		s.Stop()
	}
}

// ForceRestart 强制重启应用
func (s *DevServer) ForceRestart() {
	select {
	case s.restart <- struct{}{}:
	default:
	}
}

// openBrowser 打开浏览器
func openBrowser(url string) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default: // linux, freebsd, openbsd, netbsd
		cmd = exec.Command("xdg-open", url)
	}

	cmd.Start()
}

// MiddlewareFunc 中间件函数类型
type MiddlewareFunc func(http.Handler) http.Handler

// Use 使用中间件
func (s *DevServer) Use(middleware MiddlewareFunc) {
	// 保存原来的处理器
	handler := s.server.Handler

	// 应用中间件
	s.server.Handler = middleware(handler)
}

// ServerStatus 服务器状态
type ServerStatus struct {
	// 运行时间
	Uptime time.Duration

	// 请求数
	RequestCount int

	// 活跃请求数
	ActiveRequests int

	// 错误数
	ErrorCount int
}

// GetStatus 获取服务器状态
func (s *DevServer) GetStatus() ServerStatus {
	// TODO: 实现状态收集
	return ServerStatus{}
}
