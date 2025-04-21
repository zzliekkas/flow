package dev

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// FileChange 文件变更信息
type FileChange struct {
	// 文件路径
	Path string

	// 变更类型
	Op fsnotify.Op

	// 变更时间
	Time time.Time
}

// HotReloader 热重载器
type HotReloader struct {
	// 配置
	config *Config

	// 文件监视器
	watcher *fsnotify.Watcher

	// 文件变更通道
	changes chan FileChange

	// 重载通道
	reload chan struct{}

	// 退出通道
	done chan struct{}

	// 是否正在运行
	running bool

	// 互斥锁
	mu sync.Mutex

	// 最后一次触发时间
	lastReload time.Time

	// 监听的文件类型
	fileExtensions []string

	// 监听的文件后缀
	fileMatchers []func(string) bool
}

// NewHotReloader 创建热重载器
func NewHotReloader(config *Config) (*HotReloader, error) {
	if config == nil {
		config = NewConfig()
	}

	// 创建文件监视器
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("创建文件监视器失败: %w", err)
	}

	return &HotReloader{
		config:         config,
		watcher:        watcher,
		changes:        make(chan FileChange, 100),
		reload:         make(chan struct{}, 1),
		done:           make(chan struct{}),
		running:        false,
		fileExtensions: []string{".go", ".html", ".tpl", ".tmpl", ".css", ".js", ".json", ".yaml", ".yml"},
		fileMatchers:   []func(string) bool{},
	}, nil
}

// AddFileExtension 添加监听的文件类型
func (r *HotReloader) AddFileExtension(ext ...string) *HotReloader {
	r.fileExtensions = append(r.fileExtensions, ext...)
	return r
}

// AddFileMatcher 添加文件匹配器
func (r *HotReloader) AddFileMatcher(matcher func(string) bool) *HotReloader {
	r.fileMatchers = append(r.fileMatchers, matcher)
	return r
}

// Start 启动热重载器
func (r *HotReloader) Start() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.running {
		return nil
	}

	// 获取要监视的目录
	dirs := r.config.GetWatchDirsAbsolute()
	if len(dirs) == 0 {
		return fmt.Errorf("没有可监视的目录")
	}

	// 添加要监视的目录
	for _, dir := range dirs {
		if err := r.watchDir(dir); err != nil {
			return fmt.Errorf("添加监视目录 %s 失败: %w", dir, err)
		}
	}

	// 启动变更处理
	r.running = true
	go r.processEvents()
	go r.processChanges()

	log.Printf("热重载已启动，监视目录: %v\n", dirs)
	log.Printf("忽略模式: %v\n", r.config.Reload.IgnorePatterns)
	log.Printf("监视文件类型: %v\n", r.fileExtensions)

	return nil
}

// Stop 停止热重载器
func (r *HotReloader) Stop() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.running {
		return nil
	}

	// 关闭通道
	close(r.done)
	r.running = false

	// 关闭文件监视器
	if err := r.watcher.Close(); err != nil {
		return fmt.Errorf("关闭文件监视器失败: %w", err)
	}

	log.Println("热重载已停止")
	return nil
}

// OnReload 设置重载回调
func (r *HotReloader) OnReload(callback func()) *HotReloader {
	r.config.Reload.OnReload = callback
	return r
}

// IsRunning 检查热重载器是否正在运行
func (r *HotReloader) IsRunning() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.running
}

// watchDir 监视目录及其子目录
func (r *HotReloader) watchDir(dir string) error {
	// 检查目录是否存在
	info, err := os.Stat(dir)
	if err != nil {
		return err
	}

	// 如果不是目录，直接返回
	if !info.IsDir() {
		return nil
	}

	// 判断是否应该忽略该目录
	if r.config.ShouldIgnoreFile(dir) {
		return nil
	}

	// 监视当前目录
	if err := r.watcher.Add(dir); err != nil {
		return err
	}

	// 递归监视子目录
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 只处理目录
		if !info.IsDir() {
			return nil
		}

		// 判断是否应该忽略
		if r.config.ShouldIgnoreFile(path) {
			return filepath.SkipDir
		}

		// 监视该目录
		return r.watcher.Add(path)
	})
}

// processEvents 处理监视事件
func (r *HotReloader) processEvents() {
	for {
		select {
		case <-r.done:
			return
		case event, ok := <-r.watcher.Events:
			if !ok {
				return
			}

			// 处理文件变更事件
			if r.shouldHandleEvent(event) {
				r.changes <- FileChange{
					Path: event.Name,
					Op:   event.Op,
					Time: time.Now(),
				}
			}

			// 如果是新建目录，增加监视
			if event.Op&fsnotify.Create == fsnotify.Create {
				info, err := os.Stat(event.Name)
				if err == nil && info.IsDir() && !r.config.ShouldIgnoreFile(event.Name) {
					r.watchDir(event.Name)
				}
			}
		case err, ok := <-r.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("文件监视错误: %v\n", err)
		}
	}
}

// shouldHandleEvent 判断是否应该处理事件
func (r *HotReloader) shouldHandleEvent(event fsnotify.Event) bool {
	// 忽略特定操作
	if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) == 0 {
		return false
	}

	// 检查是否应该忽略该文件
	if r.config.ShouldIgnoreFile(event.Name) {
		return false
	}

	// 检查文件扩展名
	ext := filepath.Ext(event.Name)
	if len(r.fileExtensions) > 0 {
		matched := false
		for _, e := range r.fileExtensions {
			if strings.EqualFold(ext, e) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// 检查自定义匹配器
	if len(r.fileMatchers) > 0 {
		for _, matcher := range r.fileMatchers {
			if matcher(event.Name) {
				return true
			}
		}
		return false
	}

	return true
}

// processChanges 处理文件变更
func (r *HotReloader) processChanges() {
	var debounceTimer *time.Timer
	var pendingChanges []FileChange

	for {
		select {
		case <-r.done:
			return
		case change := <-r.changes:
			// 添加到待处理变更
			pendingChanges = append(pendingChanges, change)

			// 设置防抖计时器
			if debounceTimer == nil {
				debounceTimer = time.AfterFunc(time.Duration(r.config.Reload.DelayMilliseconds)*time.Millisecond, func() {
					r.handleChanges(pendingChanges)
					pendingChanges = nil
					debounceTimer = nil
				})
			}
		}
	}
}

// handleChanges 处理一批变更
func (r *HotReloader) handleChanges(changes []FileChange) {
	// 检查频率限制
	now := time.Now()
	minInterval := time.Duration(r.config.Reload.MaxFrequency) * time.Second
	if !r.lastReload.IsZero() && now.Sub(r.lastReload) < minInterval {
		// 太频繁，稍后再触发
		time.AfterFunc(minInterval, func() {
			r.reload <- struct{}{}
		})
		return
	}

	// 记录并打印变更
	if len(changes) > 0 {
		log.Printf("检测到 %d 个文件变更:\n", len(changes))
		for i, change := range changes {
			if i < 5 { // 只显示前5个
				log.Printf("  %s: %s\n", change.Op, change.Path)
			}
		}
		if len(changes) > 5 {
			log.Printf("  ... 还有 %d 个变更\n", len(changes)-5)
		}

		// 更新最后触发时间
		r.lastReload = now

		// 触发重载
		select {
		case r.reload <- struct{}{}:
			// 执行重载回调
			if r.config.Reload.OnReload != nil {
				go r.config.Reload.OnReload()
			}
		default:
			// 重载通道已满，跳过
		}
	}
}

// GetReloadChannel 获取重载通道
func (r *HotReloader) GetReloadChannel() <-chan struct{} {
	return r.reload
}

// GetChangesChannel 获取变更通道
func (r *HotReloader) GetChangesChannel() <-chan FileChange {
	return r.changes
}

// ForceReload 强制触发重载
func (r *HotReloader) ForceReload() {
	r.reload <- struct{}{}
	if r.config.Reload.OnReload != nil {
		go r.config.Reload.OnReload()
	}
}

// WatchStats 监视器统计
type WatchStats struct {
	// 监视的目录数量
	DirCount int

	// 最近的变更
	RecentChanges []FileChange

	// 最后重载时间
	LastReload time.Time
}

// GetStats 获取监视器统计
func (r *HotReloader) GetStats() WatchStats {
	r.mu.Lock()
	defer r.mu.Unlock()

	watchedDirs := 0
	// 计算监视的目录数量
	for _, dir := range r.config.GetWatchDirsAbsolute() {
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err == nil && info.IsDir() && !r.config.ShouldIgnoreFile(path) {
				watchedDirs++
			}
			return nil
		})
	}

	return WatchStats{
		DirCount:   watchedDirs,
		LastReload: r.lastReload,
	}
}
