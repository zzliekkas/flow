package flow

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"
)

// resolveAddr 解析监听地址，与gin.resolveAddress逻辑保持一致
func resolveAddr(addr []string) string {
	switch len(addr) {
	case 0:
		if port := os.Getenv("PORT"); port != "" {
			return ":" + port
		}
		return ":8080"
	case 1:
		return addr[0]
	default:
		panic("too many parameters")
	}
}

// executeHooks 按优先级排序并执行钩子
func executeHooks(hooks []hook) {
	sort.Slice(hooks, func(i, j int) bool {
		return hooks[i].priority < hooks[j].priority
	})
	for _, h := range hooks {
		h.fn()
	}
}

// Run 启动HTTP服务器
func (e *Engine) Run(addr ...string) error {
	// 显示Flow框架Banner
	if os.Getenv("FLOW_HIDE_BANNER") != "true" {
		fmt.Printf(FlowBanner, Version)
	}

	// 执行启动钩子
	executeHooks(e.startHooks)

	address := resolveAddr(addr)

	// 创建并持有http.Server引用，支持优雅关闭
	e.server = &http.Server{
		Addr:    address,
		Handler: e.Engine,
	}

	// 使用自定义listener以便在启动前打印地址
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	flog.Infof("Flow 服务器监听地址: %s", address)
	return e.server.Serve(listener)
}

// OnStart 注册启动钩子函数，priority 越小越先执行
func (e *Engine) OnStart(fn func(), priority ...int) {
	p := 100
	if len(priority) > 0 {
		p = priority[0]
	}
	e.startHooks = append(e.startHooks, hook{fn: fn, priority: p})
}

// OnShutdown 注册关闭钩子函数，priority 越小越先执行
func (e *Engine) OnShutdown(fn func(), priority ...int) {
	p := 100
	if len(priority) > 0 {
		p = priority[0]
	}
	e.shutdownHooks = append(e.shutdownHooks, hook{fn: fn, priority: p})
}

// Shutdown 优雅关闭HTTP服务器
func (e *Engine) Shutdown(ctx context.Context) error {
	// 执行关闭钩子
	executeHooks(e.shutdownHooks)

	// 关闭HTTP服务器
	if e.server != nil {
		return e.server.Shutdown(ctx)
	}
	return nil
}

// WaitForTermination 等待终止信号
func (e *Engine) WaitForTermination() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		flog.Errorf("服务器关闭失败: %v", err)
	}
}
