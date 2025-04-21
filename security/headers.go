package security

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// HeadersManager 安全头部管理器接口
type HeadersManager interface {
	// ApplyHeaders 将安全头部应用到响应中
	ApplyHeaders(w http.ResponseWriter)

	// ApplyHeadersWithRequest 将安全头部应用到响应中，考虑请求的属性（如TLS）
	ApplyHeadersWithRequest(w http.ResponseWriter, r *http.Request)

	// SetHeader 设置自定义安全头部
	SetHeader(name, value string)

	// RemoveHeader 移除安全头部
	RemoveHeader(name string)

	// GetHeader 获取安全头部值
	GetHeader(name string) string
}

// HeadersManagerImpl 安全头部管理器实现
type HeadersManagerImpl struct {
	// 配置
	config HeadersConfig

	// 头部映射
	headers map[string]string
}

// NewHeadersManager 创建安全头部管理器
func NewHeadersManager(config HeadersConfig) HeadersManager {
	// 初始化头部映射
	headers := make(map[string]string)

	// 添加X-Frame-Options头部
	if config.FrameOptions != "" {
		headers["X-Frame-Options"] = config.FrameOptions
	}

	// 添加X-Content-Type-Options头部
	if config.ContentTypeOptions != "" {
		headers["X-Content-Type-Options"] = config.ContentTypeOptions
	}

	// 添加X-XSS-Protection头部
	if config.XSSProtection != "" {
		headers["X-XSS-Protection"] = config.XSSProtection
	}

	// 添加Referrer-Policy头部
	if config.EnableReferrerPolicy && config.ReferrerPolicy != "" {
		headers["Referrer-Policy"] = config.ReferrerPolicy
	}

	// HSTS头部将在ApplyHeadersWithRequest中设置，以便检查是否是HTTPS连接

	// 添加Permissions-Policy头部
	if config.PermissionsPolicy != nil && len(config.PermissionsPolicy) > 0 {
		policies := make([]string, 0, len(config.PermissionsPolicy))
		for feature, value := range config.PermissionsPolicy {
			policies = append(policies, feature+"="+value)
		}
		headers["Permissions-Policy"] = strings.Join(policies, ", ")
	}

	// 添加自定义头部
	if config.CustomHeaders != nil {
		for name, value := range config.CustomHeaders {
			headers[name] = value
		}
	}

	return &HeadersManagerImpl{
		config:  config,
		headers: headers,
	}
}

// ApplyHeaders 将安全头部应用到响应中
// 注意：此方法不设置HSTS头部，因为它需要检查请求是否通过HTTPS
func (m *HeadersManagerImpl) ApplyHeaders(w http.ResponseWriter) {
	for name, value := range m.headers {
		w.Header().Set(name, value)
	}
}

// ApplyHeadersWithRequest 将安全头部应用到响应中，考虑请求的属性
func (m *HeadersManagerImpl) ApplyHeadersWithRequest(w http.ResponseWriter, r *http.Request) {
	// 先应用基本头部
	m.ApplyHeaders(w)

	// 检查是否应该应用HSTS
	if m.config.HSTS.Enabled && isHTTPS(r) {
		value := fmt.Sprintf("max-age=%d", m.config.HSTS.MaxAge)

		if m.config.HSTS.IncludeSubDomains {
			value += "; includeSubDomains"
		}

		if m.config.HSTS.Preload {
			value += "; preload"
		}

		w.Header().Set("Strict-Transport-Security", value)
	}
}

// isHTTPS 检查请求是否通过HTTPS
func isHTTPS(r *http.Request) bool {
	// 检查TLS
	if r.TLS != nil {
		return true
	}

	// 检查X-Forwarded-Proto头部（对于代理场景）
	if r.Header.Get("X-Forwarded-Proto") == "https" {
		return true
	}

	// 如果无法确定，默认假设不是HTTPS
	return false
}

// SetHeader 设置自定义安全头部
func (m *HeadersManagerImpl) SetHeader(name, value string) {
	m.headers[name] = value
}

// RemoveHeader 移除安全头部
func (m *HeadersManagerImpl) RemoveHeader(name string) {
	delete(m.headers, name)
}

// GetHeader 获取安全头部值
func (m *HeadersManagerImpl) GetHeader(name string) string {
	return m.headers[name]
}

// Common security header values
const (
	// X-Frame-Options values
	XFrameOptionsDeny       = "DENY"
	XFrameOptionsSameOrigin = "SAMEORIGIN"
	XFrameOptionsAllowFrom  = "ALLOW-FROM"

	// X-Content-Type-Options values
	XContentTypeOptionsNoSniff = "nosniff"

	// X-XSS-Protection values
	XXSSProtectionDisabled = "0"
	XXSSProtectionEnabled  = "1"
	XXSSProtectionBlock    = "1; mode=block"
	XXSSProtectionReport   = "1; report="

	// Referrer-Policy values
	ReferrerPolicyNoReferrer                  = "no-referrer"
	ReferrerPolicyNoReferrerWhenDowngrade     = "no-referrer-when-downgrade"
	ReferrerPolicySameOrigin                  = "same-origin"
	ReferrerPolicyOrigin                      = "origin"
	ReferrerPolicyStrictOrigin                = "strict-origin"
	ReferrerPolicyOriginWhenCrossOrigin       = "origin-when-cross-origin"
	ReferrerPolicyStrictOriginWhenCrossOrigin = "strict-origin-when-cross-origin"
	ReferrerPolicyUnsafeUrl                   = "unsafe-url"
)

// BuildPermissionsPolicyHeader 构建Permissions-Policy头部值
func BuildPermissionsPolicyHeader(permissions map[string][]string) string {
	if len(permissions) == 0 {
		return ""
	}

	parts := make([]string, 0, len(permissions))

	for feature, origins := range permissions {
		// 禁用特性
		if len(origins) == 0 {
			parts = append(parts, feature+"=()")
			continue
		}

		// 允许全部
		if origins[0] == "*" {
			parts = append(parts, feature+"=*")
			continue
		}

		// 允许特定源
		originList := make([]string, len(origins))
		for i, origin := range origins {
			originList[i] = "\"" + origin + "\""
		}

		parts = append(parts, feature+"=("+strings.Join(originList, " ")+")")
	}

	return strings.Join(parts, ", ")
}

// BuildFeaturePolicyHeader 构建Feature-Policy头部值（已废弃，但某些浏览器仍支持）
// 使用此函数前请考虑使用BuildPermissionsPolicyHeader
func BuildFeaturePolicyHeader(features map[string][]string) string {
	if len(features) == 0 {
		return ""
	}

	parts := make([]string, 0, len(features))

	for feature, origins := range features {
		if len(origins) == 0 {
			parts = append(parts, feature+" 'none'")
		} else if origins[0] == "*" {
			parts = append(parts, feature+" *")
		} else {
			origins := strings.Join(origins, " ")
			parts = append(parts, feature+" "+origins)
		}
	}

	return strings.Join(parts, "; ")
}

// ParseStrictTransportSecurity 解析Strict-Transport-Security头部值
func ParseStrictTransportSecurity(header string) (maxAge int, includeSubDomains, preload bool) {
	parts := strings.Split(header, ";")

	for _, part := range parts {
		part = strings.TrimSpace(part)

		if strings.HasPrefix(part, "max-age=") {
			ageStr := strings.TrimPrefix(part, "max-age=")
			age, err := strconv.Atoi(ageStr)
			if err == nil && age > 0 {
				maxAge = age
			}
		} else if part == "includeSubDomains" {
			includeSubDomains = true
		} else if part == "preload" {
			preload = true
		}
	}

	return
}
