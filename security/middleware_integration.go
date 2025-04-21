package security

import (
	"strings"

	"github.com/zzliekkas/flow"
)

// SecurityMiddleware 返回完整的安全中间件链
func (m *Manager) SecurityMiddleware() flow.HandlerFunc {
	return func(c *flow.Context) {
		// 应用安全头部，包括根据HTTPS状态设置HSTS
		if m.config.Headers.Enabled && m.headers != nil {
			m.headers.ApplyHeadersWithRequest(c.Writer, c.Request)
		}

		// 应用CSP头部
		if m.config.CSP.Enabled && m.csp != nil {
			m.csp.ApplyCSP(c.Writer)
		}

		// 继续处理请求
		c.Next()

		// 可以在这里添加响应处理逻辑
	}
}

// CSRFMiddleware 创建CSRF保护中间件
func (m *Manager) CSRFMiddleware() flow.HandlerFunc {
	return func(c *flow.Context) {
		// 先应用安全头部
		if m.config.Headers.Enabled {
			// 设置特定的CSRF相关头部
			c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
		}

		// 在实际项目中，这里应该调用真实的CSRF中间件
		c.Next()
	}
}

// XSSProtectionMiddleware 创建XSS保护中间件
func (m *Manager) XSSProtectionMiddleware() flow.HandlerFunc {
	return func(c *flow.Context) {
		// 为所有响应应用XSS保护头部
		c.Writer.Header().Set("X-XSS-Protection", "1; mode=block")

		// 内容类型嗅探保护
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")

		// 继续处理请求
		c.Next()

		// 如果启用了XSS过滤
		if m.config.XSS.Enabled && m.xss != nil {
			// 对HTML响应进行处理
			contentType := c.Writer.Header().Get("Content-Type")
			if strings.Contains(contentType, "text/html") {
				// 此处仅做示例，实际应该使用ResponseWriter包装器
				// 来捕获并处理响应体
			}
		}
	}
}

// CSPNonceMiddleware 创建带有CSP nonce的中间件
func (m *Manager) CSPNonceMiddleware() flow.HandlerFunc {
	return func(c *flow.Context) {
		// 生成随机nonce
		nonce := GenerateNonce()

		// 将nonce存储在上下文中
		c.Set("csp-nonce", nonce)

		// 如果CSP启用，将nonce添加到script-src指令
		if m.config.CSP.Enabled && m.csp != nil {
			// 添加script-src 'nonce-{nonce}'
			if builder, ok := m.csp.(*CSPBuilderImpl); ok {
				builder.AddDirective("script-src", "'nonce-"+nonce+"'")
				builder.ApplyCSP(c.Writer)
			}
		}

		c.Next()
	}
}

// RateLimitMiddleware 创建速率限制中间件
func (m *Manager) RateLimitMiddleware() flow.HandlerFunc {
	return func(c *flow.Context) {
		// 在实际项目中，这里应该实现真实的速率限制逻辑
		c.Next()
	}
}

// IntegrateWithExistingMiddleware 将安全管理器集成到现有中间件
func (m *Manager) IntegrateWithExistingMiddleware(app *flow.Engine) {
	// 添加安全头部和CSP
	app.Use(m.SecurityMiddleware())

	// 根据配置添加其他中间件
	if m.config.XSS.Enabled {
		app.Use(m.XSSProtectionMiddleware())
	}
}

// SanitizeInput 对用户输入进行净化
func (m *Manager) SanitizeInput(input string) string {
	if m.config.XSS.Enabled && m.xss != nil {
		return m.xss.Sanitize(input)
	}
	return input
}

// SecureHeaders 中间件函数简单版本
func SecureHeaders(c *flow.Context) {
	// 设置安全相关的HTTP头部
	headers := c.Writer.Header()

	// X-Frame-Options 防止点击劫持
	headers.Set("X-Frame-Options", "SAMEORIGIN")

	// X-Content-Type-Options 防止MIME类型嗅探
	headers.Set("X-Content-Type-Options", "nosniff")

	// X-XSS-Protection 启用浏览器XSS过滤器
	headers.Set("X-XSS-Protection", "1; mode=block")

	// Referrer-Policy 控制Referrer头部
	headers.Set("Referrer-Policy", "strict-origin-when-cross-origin")

	// 严格传输安全 (只在HTTPS环境使用)
	if c.Request.TLS != nil {
		headers.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	}

	c.Next()
}

// ContentSecurityPolicy 中间件函数简单版本
func ContentSecurityPolicy(c *flow.Context) {
	// 设置内容安全策略
	policy := "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; font-src 'self'; object-src 'none'; frame-ancestors 'self'; form-action 'self'; upgrade-insecure-requests;"

	c.Writer.Header().Set("Content-Security-Policy", policy)
	c.Next()
}
