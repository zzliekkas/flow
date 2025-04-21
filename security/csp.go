package security

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strings"
)

// CSPBuilder 内容安全策略构建器接口
type CSPBuilder interface {
	// ApplyCSP 将CSP头部应用到响应中
	ApplyCSP(w http.ResponseWriter)

	// GetPolicyString 获取CSP策略字符串
	GetPolicyString() string

	// AddDirective 添加CSP指令
	AddDirective(directive string, values ...string)

	// RemoveDirective 移除CSP指令
	RemoveDirective(directive string)

	// SetReportOnly 设置是否为仅报告模式
	SetReportOnly(reportOnly bool)
}

// CSPBuilderImpl 内容安全策略构建器实现
type CSPBuilderImpl struct {
	// 配置
	config CSPConfig

	// 策略指令映射，键为指令名称，值为指令值列表
	directives map[string][]string

	// 是否为仅报告模式
	reportOnly bool
}

// NewCSPBuilder 创建CSP构建器
func NewCSPBuilder(config CSPConfig) CSPBuilder {
	directives := make(map[string][]string)

	if len(config.DefaultSrc) > 0 {
		directives["default-src"] = config.DefaultSrc
	}

	if len(config.ScriptSrc) > 0 {
		directives["script-src"] = config.ScriptSrc
	}

	if len(config.StyleSrc) > 0 {
		directives["style-src"] = config.StyleSrc
	}

	if len(config.ImgSrc) > 0 {
		directives["img-src"] = config.ImgSrc
	}

	if len(config.ConnectSrc) > 0 {
		directives["connect-src"] = config.ConnectSrc
	}

	if len(config.FontSrc) > 0 {
		directives["font-src"] = config.FontSrc
	}

	if len(config.ObjectSrc) > 0 {
		directives["object-src"] = config.ObjectSrc
	}

	if len(config.MediaSrc) > 0 {
		directives["media-src"] = config.MediaSrc
	}

	if len(config.FrameSrc) > 0 {
		directives["frame-src"] = config.FrameSrc
	}

	if len(config.Sandbox) > 0 {
		directives["sandbox"] = config.Sandbox
	}

	if len(config.ChildSrc) > 0 {
		directives["child-src"] = config.ChildSrc
	}

	if len(config.FormAction) > 0 {
		directives["form-action"] = config.FormAction
	}

	if len(config.FrameAncestors) > 0 {
		directives["frame-ancestors"] = config.FrameAncestors
	}

	if len(config.PluginTypes) > 0 {
		directives["plugin-types"] = config.PluginTypes
	}

	if len(config.BaseURI) > 0 {
		directives["base-uri"] = config.BaseURI
	}

	if len(config.WorkerSrc) > 0 {
		directives["worker-src"] = config.WorkerSrc
	}

	if len(config.ManifestSrc) > 0 {
		directives["manifest-src"] = config.ManifestSrc
	}

	if len(config.PrefetchSrc) > 0 {
		directives["prefetch-src"] = config.PrefetchSrc
	}

	if len(config.RequireSriFor) > 0 {
		directives["require-sri-for"] = config.RequireSriFor
	}

	// 添加其他特殊指令
	if config.ReportURI != "" {
		directives["report-uri"] = []string{config.ReportURI}
	}

	if config.UpgradeInsecureRequests {
		directives["upgrade-insecure-requests"] = []string{}
	}

	if config.BlockAllMixedContent {
		directives["block-all-mixed-content"] = []string{}
	}

	return &CSPBuilderImpl{
		config:     config,
		directives: directives,
		reportOnly: config.ReportOnly,
	}
}

// ApplyCSP 将CSP头部应用到响应中
func (b *CSPBuilderImpl) ApplyCSP(w http.ResponseWriter) {
	policy := b.GetPolicyString()
	if policy == "" {
		return
	}

	headerName := "Content-Security-Policy"
	if b.reportOnly {
		headerName = "Content-Security-Policy-Report-Only"
	}

	w.Header().Set(headerName, policy)
}

// GetPolicyString 获取CSP策略字符串
func (b *CSPBuilderImpl) GetPolicyString() string {
	if len(b.directives) == 0 {
		return ""
	}

	parts := make([]string, 0, len(b.directives))
	for directive, values := range b.directives {
		if len(values) == 0 {
			parts = append(parts, directive)
		} else {
			parts = append(parts, directive+" "+strings.Join(values, " "))
		}
	}

	return strings.Join(parts, "; ")
}

// AddDirective 添加CSP指令
func (b *CSPBuilderImpl) AddDirective(directive string, values ...string) {
	directive = strings.ToLower(directive)

	// 如果已存在，合并值
	if existing, ok := b.directives[directive]; ok {
		valueSet := make(map[string]bool)

		for _, v := range existing {
			valueSet[v] = true
		}

		for _, v := range values {
			if !valueSet[v] {
				existing = append(existing, v)
				valueSet[v] = true
			}
		}

		b.directives[directive] = existing
	} else {
		b.directives[directive] = values
	}
}

// RemoveDirective 移除CSP指令
func (b *CSPBuilderImpl) RemoveDirective(directive string) {
	delete(b.directives, strings.ToLower(directive))
}

// SetReportOnly 设置是否为仅报告模式
func (b *CSPBuilderImpl) SetReportOnly(reportOnly bool) {
	b.reportOnly = reportOnly
}

// GenerateNonce 生成CSP nonce
func GenerateNonce() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

// CSPContextKey 为CSP nonce提供的上下文键类型
type CSPContextKey string

// CSPNonceKey CSP nonce的上下文键
const CSPNonceKey = CSPContextKey("csp-nonce")

// WithCSPNonce 将CSP nonce添加到上下文
func WithCSPNonce(ctx context.Context, nonce string) context.Context {
	return context.WithValue(ctx, CSPNonceKey, nonce)
}

// GetCSPNonce 从上下文中获取CSP nonce
func GetCSPNonce(ctx context.Context) string {
	if nonce, ok := ctx.Value(CSPNonceKey).(string); ok {
		return nonce
	}
	return ""
}
