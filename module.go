package flow

// Module 定义模块接口
// 模块是自包含的功能单元，负责向DI容器注册自己的服务和路由
type Module interface {
	// Name 返回模块名称，用于日志和调试
	Name() string
	// Init 初始化模块：注册服务到DI容器
	Init(e *Engine) error
}

// RoutableModule 可选接口，模块可以同时实现路由注册
type RoutableModule interface {
	Module
	// RegisterRoutes 注册模块的HTTP路由
	RegisterRoutes(e *Engine)
}

// RegisterModule 注册单个模块到引擎
func (e *Engine) RegisterModule(m Module) error {
	flog.Debugf("注册模块: %s", m.Name())
	if err := m.Init(e); err != nil {
		flog.Errorf("模块 %s 初始化失败: %v", m.Name(), err)
		return err
	}

	// 如果模块实现了 RoutableModule，自动注册路由
	if rm, ok := m.(RoutableModule); ok {
		rm.RegisterRoutes(e)
	}

	return nil
}

// RegisterModules 批量注册模块
func (e *Engine) RegisterModules(modules ...Module) error {
	for _, m := range modules {
		if err := e.RegisterModule(m); err != nil {
			return err
		}
	}
	return nil
}
