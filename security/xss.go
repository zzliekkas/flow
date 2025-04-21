package security

import (
	"html"
	"regexp"
	"strings"
)

// HTMLPolicy 定义HTML内容清理策略
type HTMLPolicy struct {
	config         XSSConfig
	allowedTags    map[string]bool
	allowedAttrs   map[string]bool
	allowedProtos  map[string]bool
	tagPattern     *regexp.Regexp
	attrPattern    *regexp.Regexp
	linkifyPattern *regexp.Regexp
}

// XSSManager XSS防护管理器
type XSSManager struct {
	config XSSConfig
	policy *HTMLPolicy
}

// NewXSSManager 创建XSS防护管理器
func NewXSSManager(config XSSConfig) *XSSManager {
	manager := &XSSManager{
		config: config,
		policy: newHTMLPolicy(config),
	}
	return manager
}

// newHTMLPolicy 创建HTML策略
func newHTMLPolicy(config XSSConfig) *HTMLPolicy {
	p := &HTMLPolicy{
		config:        config,
		allowedTags:   make(map[string]bool),
		allowedAttrs:  make(map[string]bool),
		allowedProtos: make(map[string]bool),
		tagPattern:    regexp.MustCompile(`<([a-z][a-z0-9]*)\b([^>]*)>(?:(.*?)</\1>)?`),
		attrPattern:   regexp.MustCompile(`([a-z0-9\-_]+)(?:\s*=\s*(?:(?:"([^"]*)")|(?:'([^']*)')|([^>\s]+)))?`),
	}

	// 初始化允许的标签
	for _, tag := range config.AllowedTags {
		p.allowedTags[strings.ToLower(tag)] = true
	}

	// 初始化允许的属性
	for _, attr := range config.AllowedAttributes {
		p.allowedAttrs[strings.ToLower(attr)] = true
	}

	// 初始化允许的协议
	for _, proto := range config.AllowedProtocols {
		p.allowedProtos[strings.ToLower(proto)] = true
	}

	// 配置链接识别正则表达式
	p.linkifyPattern = regexp.MustCompile(`(https?|ftp)://([^\s<>"']+)`)

	return p
}

// HTMLPolicy 返回HTML策略
func (m *XSSManager) HTMLPolicy() *HTMLPolicy {
	return m.policy
}

// Sanitize 清理HTML内容
func (m *XSSManager) Sanitize(input string) string {
	if input == "" {
		return ""
	}

	if !m.config.AllowHTML {
		return html.EscapeString(input)
	}

	return m.policy.Sanitize(input)
}

// Escape 转义HTML内容
func (m *XSSManager) Escape(input string) string {
	if input == "" {
		return ""
	}
	return html.EscapeString(input)
}

// Sanitize 按照策略清理HTML内容
func (p *HTMLPolicy) Sanitize(input string) string {
	if !p.config.AllowHTML {
		return html.EscapeString(input)
	}

	// 这里实现简化版的HTML清理逻辑
	// 实际实现应该使用更复杂的HTML解析和重建
	result := p.tagPattern.ReplaceAllStringFunc(input, func(match string) string {
		// 解析标签
		submatch := p.tagPattern.FindStringSubmatch(match)
		if len(submatch) < 2 {
			return html.EscapeString(match)
		}

		tagName := strings.ToLower(submatch[1])
		if !p.allowedTags[tagName] {
			return html.EscapeString(match)
		}

		// 处理属性
		attrs := submatch[2]
		cleanAttrs := p.sanitizeAttributes(attrs, tagName)

		// 处理内容
		content := ""
		if len(submatch) > 3 && submatch[3] != "" {
			content = p.Sanitize(submatch[3]) // 递归处理内容
		}

		return "<" + tagName + cleanAttrs + ">" + content + "</" + tagName + ">"
	})

	// 如果启用了链接处理
	if p.config.EnableLinkify {
		result = p.linkifyPattern.ReplaceAllStringFunc(result, func(url string) string {
			// 检查URL是否已经在标签内
			if strings.HasPrefix(url, "<a") {
				return url
			}
			return `<a href="` + url + `" target="_blank">` + url + `</a>`
		})
	}

	return result
}

// sanitizeAttributes 清理属性
func (p *HTMLPolicy) sanitizeAttributes(attrs string, tagName string) string {
	if attrs == "" {
		return ""
	}

	result := " "
	matches := p.attrPattern.FindAllStringSubmatch(attrs, -1)
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		attrName := strings.ToLower(match[1])
		if !p.allowedAttrs[attrName] && !strings.HasPrefix(attrName, "data-") {
			continue
		}

		// 特殊处理URL属性
		attrValue := ""
		if len(match) > 2 {
			for i := 2; i < len(match); i++ {
				if match[i] != "" {
					attrValue = match[i]
					break
				}
			}
		}

		if isURLAttribute(attrName) && attrValue != "" {
			if !p.isSafeURL(attrValue) {
				continue
			}
		}

		// 特殊处理样式属性
		if attrName == "style" && attrValue != "" {
			attrValue = p.sanitizeCSS(attrValue)
		}

		result += attrName + `="` + html.EscapeString(attrValue) + `" `
	}

	return strings.TrimSpace(result)
}

// isSafeURL 检查URL是否安全
func (p *HTMLPolicy) isSafeURL(url string) bool {
	url = strings.TrimSpace(strings.ToLower(url))
	if url == "" {
		return false
	}

	// 检查JavaScript URL
	if strings.HasPrefix(url, "javascript:") {
		return false
	}

	// 检查数据URL
	if strings.HasPrefix(url, "data:") && !strings.HasPrefix(url, "data:image/") {
		return false
	}

	// 检查协议
	for _, proto := range p.config.AllowedProtocols {
		if strings.HasPrefix(url, proto+":") {
			return true
		}
	}

	// 默认允许相对URL
	return !strings.Contains(url, ":")
}

// sanitizeCSS 清理CSS
func (p *HTMLPolicy) sanitizeCSS(css string) string {
	// 简化实现：移除可能包含JavaScript的属性
	css = strings.ReplaceAll(css, "expression", "")
	css = strings.ReplaceAll(css, "javascript", "")
	css = strings.ReplaceAll(css, "eval", "")
	return css
}

// isURLAttribute 检查属性是否为URL类型
func isURLAttribute(attrName string) bool {
	switch attrName {
	case "href", "src", "cite", "action", "formaction", "data", "poster", "background":
		return true
	default:
		return false
	}
}

// DefaultXSSConfig 返回默认的XSS配置
func DefaultXSSConfig() XSSConfig {
	return XSSConfig{
		Enabled:   true,
		AllowHTML: true,
		AllowedTags: []string{
			"a", "b", "blockquote", "br", "code", "dd", "div", "dl", "dt",
			"em", "h1", "h2", "h3", "h4", "h5", "h6", "hr", "i", "img",
			"li", "ol", "p", "pre", "s", "span", "strong", "sub", "sup",
			"table", "tbody", "td", "tfoot", "th", "thead", "tr", "u", "ul",
		},
		AllowedAttributes: []string{
			"alt", "class", "colspan", "data-*", "href", "id", "rowspan",
			"src", "style", "target", "title",
		},
		AllowedProtocols: []string{
			"http", "https", "mailto", "tel",
		},
		EnableLinkify: true,
	}
}
