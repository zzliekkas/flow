package security

import (
	"net/http"
)

// Provider 安全提供者接口
type Provider interface {
	// Register 注册安全服务到容器
	Register()

	// Boot 启动安全服务
	Boot()
}

// SecurityProvider 安全提供者实现
type SecurityProvider struct {
	config  Config
	manager *Manager
}

// NewSecurityProvider 创建安全提供者
func NewSecurityProvider(config Config) *SecurityProvider {
	return &SecurityProvider{
		config: config,
	}
}

// Register 注册安全服务
func (p *SecurityProvider) Register() {
	p.manager = NewManager(p.config)
}

// Boot 启动安全服务
func (p *SecurityProvider) Boot() {
	// 可以在这里添加启动时需要执行的操作
}

// Manager 获取安全管理器
func (p *SecurityProvider) Manager() *Manager {
	return p.manager
}

// Middleware 获取安全中间件
func (p *SecurityProvider) Middleware() func(http.Handler) http.Handler {
	return p.manager.Middleware()
}
