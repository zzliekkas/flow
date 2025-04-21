package security

import (
	"net/http"
)

// Manager 是安全框架的主要管理器
type Manager struct {
	// 安全配置
	config Config
	// 头部安全管理器
	headers HeadersManager
	// XSS防护管理器
	xss *XSSManager
	// CSP管理器
	csp CSPBuilder
	// 密码策略管理器
	password PasswordPolicy
	// 审计日志管理器
	audit AuditLogger
	// CSRF防护管理器
	csrf interface{} // 暂时用接口占位
	// 速率限制管理器
	rateLimit interface{} // 暂时用接口占位
}

// NewManager 创建并返回一个新的安全管理器实例
func NewManager(config Config) *Manager {
	m := &Manager{
		config: config,
	}

	// 初始化各组件
	if config.Headers.Enabled {
		m.headers = NewHeadersManager(config.Headers)
	}

	if config.XSS.Enabled {
		m.xss = NewXSSManager(config.XSS)
	}

	if config.CSP.Enabled {
		m.csp = NewCSPBuilder(config.CSP)
	}

	if config.Password.MinLength > 0 {
		m.password = NewPasswordPolicy(PasswordConfig(config.Password))
	}

	// 初始化审计日志
	if config.Audit.Enabled {
		m.audit = NewBasicAuditLogger(config.Audit)
	}

	// 其他组件的初始化暂不实现

	return m
}

// Middleware 返回安全中间件
func (m *Manager) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 应用安全头部
			if m.config.Headers.Enabled && m.headers != nil {
				m.headers.ApplyHeadersWithRequest(w, r)
			}

			// 应用CSP
			if m.config.CSP.Enabled && m.csp != nil {
				// 生成并应用nonce
				var nonce string
				if m.config.CSP.EnableNonce {
					nonce = GenerateNonce()
					if builder, ok := m.csp.(*CSPBuilderImpl); ok {
						builder.AddDirective("script-src", "'nonce-"+nonce+"'")
					}
					m.csp.ApplyCSP(w)
				} else {
					m.csp.ApplyCSP(w)
				}

				// 存储nonce以供模板使用
				if nonce != "" {
					ctx := r.Context()
					ctx = WithCSPNonce(ctx, nonce)
					r = r.WithContext(ctx)
				}
			}

			// 记录请求审计日志
			if m.config.Audit.Enabled && m.audit != nil {
				m.audit.LogRequest(r)
			}

			// 继续处理请求
			next.ServeHTTP(w, r)
		})
	}
}

// HTMLPolicy 返回HTML策略
func (m *Manager) HTMLPolicy() *HTMLPolicy {
	if m.xss != nil {
		return m.xss.HTMLPolicy()
	}
	return nil
}

// SanitizeHTML 清理HTML内容
func (m *Manager) SanitizeHTML(input string) string {
	if m.xss != nil {
		return m.xss.Sanitize(input)
	}
	return input
}

// ValidatePassword 验证密码是否符合策略
func (m *Manager) ValidatePassword(password string) error {
	if m.password != nil {
		valid, reasons := m.password.Validate(password)
		if !valid {
			return &PasswordValidationError{Reasons: reasons}
		}
	}
	return nil
}

// GetCSPNonce 获取当前请求的CSP nonce值
func (m *Manager) GetCSPNonce(r *http.Request) string {
	return GetCSPNonce(r.Context())
}

// GetAuditLogger 获取审计日志记录器
func (m *Manager) GetAuditLogger() AuditLogger {
	return m.audit
}

// LogAuditEvent 记录安全审计事件
func (m *Manager) LogAuditEvent(eventType, userID, action, resource string, success bool, details map[string]interface{}) {
	if m.audit != nil {
		m.audit.LogEvent(eventType, userID, action, resource, success, details)
	}
}

// PasswordValidationError 密码验证错误
type PasswordValidationError struct {
	Reasons []string
}

// Error 实现错误接口
func (e *PasswordValidationError) Error() string {
	if len(e.Reasons) == 0 {
		return "密码验证失败"
	}
	return e.Reasons[0]
}

// WithConfig 使用新配置创建克隆的Manager
func (m *Manager) WithConfig(config Config) *Manager {
	newManager := NewManager(config)
	return newManager
}

// GetConfig 获取当前配置
func (m *Manager) GetConfig() Config {
	return m.config
}

// UpdateConfig 更新配置
func (m *Manager) UpdateConfig(config Config) {
	m.config = config

	// 重新初始化各组件
	if config.Headers.Enabled {
		m.headers = NewHeadersManager(config.Headers)
	}

	if config.XSS.Enabled {
		m.xss = NewXSSManager(config.XSS)
	}

	if config.CSP.Enabled {
		m.csp = NewCSPBuilder(config.CSP)
	}

	if config.Password.MinLength > 0 {
		m.password = NewPasswordPolicy(PasswordConfig(config.Password))
	}

	// 更新审计日志
	if config.Audit.Enabled {
		m.audit = NewBasicAuditLogger(config.Audit)
	} else {
		m.audit = nil
	}
}
