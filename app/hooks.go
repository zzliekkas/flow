package app

import (
	"sync"
)

// HookType 定义钩子类型
type HookType int

const (
	// HookBeforeStart 启动前钩子
	HookBeforeStart HookType = iota
	// HookAfterStart 启动后钩子
	HookAfterStart
	// HookBeforeShutdown 关闭前钩子
	HookBeforeShutdown
	// HookAfterShutdown 关闭后钩子
	HookAfterShutdown
)

// Hook 表示应用钩子函数
type Hook struct {
	Name     string   // 钩子名称
	Function func()   // 钩子函数
	Type     HookType // 钩子类型
	Priority int      // 优先级，数值越小优先级越高
}

// HooksManager 钩子管理器
type HooksManager struct {
	hooks     map[HookType][]Hook // 按类型存储的钩子
	hookMutex sync.RWMutex        // 钩子锁
}

// NewHooksManager 创建新的钩子管理器
func NewHooksManager() *HooksManager {
	return &HooksManager{
		hooks: make(map[HookType][]Hook),
	}
}

// Register 注册钩子
func (hm *HooksManager) Register(hook Hook) {
	hm.hookMutex.Lock()
	defer hm.hookMutex.Unlock()

	// 如果该类型的钩子列表不存在，则创建
	if _, exists := hm.hooks[hook.Type]; !exists {
		hm.hooks[hook.Type] = make([]Hook, 0)
	}

	// 添加钩子
	hm.hooks[hook.Type] = append(hm.hooks[hook.Type], hook)

	// 按优先级排序
	hm.sortHooks(hook.Type)
}

// sortHooks 按优先级排序钩子
func (hm *HooksManager) sortHooks(hookType HookType) {
	hooks := hm.hooks[hookType]
	// 简单插入排序
	for i := 1; i < len(hooks); i++ {
		key := hooks[i]
		j := i - 1
		for j >= 0 && hooks[j].Priority > key.Priority {
			hooks[j+1] = hooks[j]
			j--
		}
		hooks[j+1] = key
	}
	hm.hooks[hookType] = hooks
}

// Execute 执行指定类型的所有钩子
func (hm *HooksManager) Execute(hookType HookType) {
	hm.hookMutex.RLock()
	defer hm.hookMutex.RUnlock()

	// 按顺序执行所有钩子
	if hooks, exists := hm.hooks[hookType]; exists {
		for _, hook := range hooks {
			hook.Function()
		}
	}
}

// RegisterBeforeStart 注册启动前钩子
func (hm *HooksManager) RegisterBeforeStart(name string, function func(), priority int) {
	hm.Register(Hook{
		Name:     name,
		Function: function,
		Type:     HookBeforeStart,
		Priority: priority,
	})
}

// RegisterAfterStart 注册启动后钩子
func (hm *HooksManager) RegisterAfterStart(name string, function func(), priority int) {
	hm.Register(Hook{
		Name:     name,
		Function: function,
		Type:     HookAfterStart,
		Priority: priority,
	})
}

// RegisterBeforeShutdown 注册关闭前钩子
func (hm *HooksManager) RegisterBeforeShutdown(name string, function func(), priority int) {
	hm.Register(Hook{
		Name:     name,
		Function: function,
		Type:     HookBeforeShutdown,
		Priority: priority,
	})
}

// RegisterAfterShutdown 注册关闭后钩子
func (hm *HooksManager) RegisterAfterShutdown(name string, function func(), priority int) {
	hm.Register(Hook{
		Name:     name,
		Function: function,
		Type:     HookAfterShutdown,
		Priority: priority,
	})
}
