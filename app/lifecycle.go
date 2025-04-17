package app

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/zzliekkas/flow"
)

// 应用状态常量
const (
	// StatusInit 初始化状态
	StatusInit = iota
	// StatusStarting 正在启动
	StatusStarting
	// StatusRunning 运行中
	StatusRunning
	// StatusStopping 正在停止
	StatusStopping
	// StatusStopped 已停止
	StatusStopped
)

// 状态名称映射
var statusNames = map[int]string{
	StatusInit:     "初始化",
	StatusStarting: "正在启动",
	StatusRunning:  "运行中",
	StatusStopping: "正在停止",
	StatusStopped:  "已停止",
}

// 错误定义
var (
	// ErrAppAlreadyRunning 应用已经在运行
	ErrAppAlreadyRunning = errors.New("应用已经在运行中")
	// ErrAppNotRunning 应用未运行
	ErrAppNotRunning = errors.New("应用尚未运行")
	// ErrShutdownTimeout 关闭超时错误
	ErrShutdownTimeout = errors.New("应用关闭超时")
)

// LifecycleManager 应用生命周期管理器
type LifecycleManager struct {
	engine        *flow.Engine   // Flow 引擎
	status        int            // 当前状态
	statusMutex   sync.RWMutex   // 状态锁
	shutdownCh    chan struct{}  // 关闭信号通道
	shutdownOnce  sync.Once      // 确保只执行一次关闭
	shutdownHooks []func()       // 关闭钩子函数
	logger        *logrus.Logger // 日志记录器
}

// NewLifecycleManager 创建新的生命周期管理器
func NewLifecycleManager(engine *flow.Engine) *LifecycleManager {
	return &LifecycleManager{
		engine:        engine,
		status:        StatusInit,
		shutdownCh:    make(chan struct{}),
		shutdownHooks: make([]func(), 0),
		logger:        logrus.New(),
	}
}

// Status 获取当前状态
func (lm *LifecycleManager) Status() int {
	lm.statusMutex.RLock()
	defer lm.statusMutex.RUnlock()
	return lm.status
}

// StatusName 获取当前状态名称
func (lm *LifecycleManager) StatusName() string {
	return statusNames[lm.Status()]
}

// setStatus 设置状态
func (lm *LifecycleManager) setStatus(status int) {
	lm.statusMutex.Lock()
	defer lm.statusMutex.Unlock()
	lm.status = status
}

// Start 启动应用
func (lm *LifecycleManager) Start(addr string) error {
	// 检查应用是否已经在运行
	if lm.Status() != StatusInit && lm.Status() != StatusStopped {
		return ErrAppAlreadyRunning
	}

	// 设置为正在启动状态
	lm.setStatus(StatusStarting)
	lm.logger.Info("应用正在启动...")

	// 设置信号处理
	lm.setupSignalHandling()

	// 设置为运行状态
	lm.setStatus(StatusRunning)
	lm.logger.Info("应用已启动，监听地址: ", addr)

	// 启动HTTP服务器（非阻塞）
	go func() {
		if err := lm.engine.Run(addr); err != nil {
			lm.logger.Errorf("HTTP服务器错误: %v", err)
		}
	}()

	// 等待关闭信号
	<-lm.shutdownCh
	return nil
}

// Shutdown 优雅关闭应用
func (lm *LifecycleManager) Shutdown(timeout time.Duration) error {
	var err error
	lm.shutdownOnce.Do(func() {
		// 检查应用是否正在运行
		if lm.Status() != StatusRunning {
			err = ErrAppNotRunning
			return
		}

		// 设置为正在停止状态
		lm.setStatus(StatusStopping)
		lm.logger.Info("应用正在关闭...")

		// 创建带超时的上下文
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		// 执行关闭钩子
		for _, hook := range lm.shutdownHooks {
			hook()
		}

		// 关闭HTTP服务器
		if shutdownErr := lm.engine.Shutdown(ctx); shutdownErr != nil {
			err = shutdownErr
			lm.logger.Errorf("应用关闭错误: %v", shutdownErr)
		}

		// 设置为已停止状态
		lm.setStatus(StatusStopped)
		lm.logger.Info("应用已关闭")

		// 发送关闭信号
		close(lm.shutdownCh)
	})

	return err
}

// RegisterShutdownHook 注册关闭钩子函数
func (lm *LifecycleManager) RegisterShutdownHook(hook func()) {
	lm.shutdownHooks = append(lm.shutdownHooks, hook)
}

// setupSignalHandling 设置信号处理
func (lm *LifecycleManager) setupSignalHandling() {
	// 创建信号通道
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	go func() {
		sig := <-sigCh
		lm.logger.Infof("收到系统信号: %s, 开始关闭应用...", sig)
		// 默认超时时间为30秒
		if err := lm.Shutdown(30 * time.Second); err != nil {
			lm.logger.Errorf("应用关闭错误: %v", err)
		}
	}()
}
