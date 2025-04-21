package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zzliekkas/flow/i18n"
)

// 定义上下文中保存语言的键
const localeContextKey = "app.locale"

// LocaleOptions 本地化中间件的选项
type LocaleOptions struct {
	// 默认语言
	DefaultLocale string

	// 翻译器
	Translator i18n.Translator

	// 支持的语言列表
	SupportedLocales []string

	// 从查询参数获取语言的键名
	QueryParameterName string

	// 从Cookie获取语言的键名
	CookieName string

	// 从Header获取语言的键名
	HeaderName string

	// 是否自动保存语言到Cookie
	SaveToCookie bool

	// Cookie有效期（秒）
	CookieMaxAge int

	// Cookie路径
	CookiePath string

	// 是否创建子域名Cookie
	CookieSecure bool

	// 是否只允许HTTP访问Cookie
	CookieHTTPOnly bool

	// 是否使用会话Cookie（会话结束就过期）
	CookieSession bool

	// 是否自动检测浏览器语言
	DetectBrowserLocale bool
}

// DefaultLocaleOptions 返回默认的本地化选项
func DefaultLocaleOptions() *LocaleOptions {
	return &LocaleOptions{
		DefaultLocale:       "en",
		QueryParameterName:  "locale",
		CookieName:          "locale",
		HeaderName:          "Accept-Language",
		SaveToCookie:        true,
		CookieMaxAge:        60 * 60 * 24 * 30, // 30天
		CookiePath:          "/",
		CookieSecure:        false,
		CookieHTTPOnly:      false,
		CookieSession:       false,
		DetectBrowserLocale: true,
		SupportedLocales:    []string{"en", "zh"},
	}
}

// Locale 创建一个本地化中间件
func Locale(translator i18n.Translator, opts ...*LocaleOptions) gin.HandlerFunc {
	// 使用默认选项
	options := DefaultLocaleOptions()

	// 应用自定义选项
	if len(opts) > 0 && opts[0] != nil {
		options = opts[0]
	}

	// 确保有翻译器
	options.Translator = translator

	return func(c *gin.Context) {
		// 尝试获取语言，按优先级：查询参数 > Cookie > Header > 默认值
		locale := options.DefaultLocale

		// 尝试从查询参数获取
		if options.QueryParameterName != "" {
			if qlocale := c.Query(options.QueryParameterName); qlocale != "" {
				if isSupported(qlocale, options.SupportedLocales) {
					locale = qlocale
				}
			}
		}

		// 如果没有从查询参数获取，尝试从Cookie获取
		if options.CookieName != "" && locale == options.DefaultLocale {
			if clocale, err := c.Cookie(options.CookieName); err == nil && clocale != "" {
				if isSupported(clocale, options.SupportedLocales) {
					locale = clocale
				}
			}
		}

		// 如果仍未找到，尝试从Header获取
		if options.HeaderName != "" && locale == options.DefaultLocale && options.DetectBrowserLocale {
			headerLocale := extractLocaleFromHeader(c.GetHeader(options.HeaderName), options.SupportedLocales)
			if headerLocale != "" {
				locale = headerLocale
			}
		}

		// 如果选项设置为保存到Cookie，且语言不是从Cookie获取的
		if options.SaveToCookie &&
			options.CookieName != "" &&
			c.Query(options.QueryParameterName) != "" {
			cookieMaxAge := options.CookieMaxAge
			if options.CookieSession {
				cookieMaxAge = 0 // 会话Cookie
			}

			c.SetCookie(
				options.CookieName,
				locale,
				cookieMaxAge,
				options.CookiePath,
				"", // 域名，空表示当前域名
				options.CookieSecure,
				options.CookieHTTPOnly,
			)
		}

		// 将语言保存到上下文
		c.Set(localeContextKey, locale)

		// 将语言保存到翻译器上下文
		transCtx := c.Request.Context()
		transCtx = translator.SetLocale(transCtx, locale)

		// 更新请求上下文
		c.Request = c.Request.WithContext(transCtx)

		c.Next()
	}
}

// GetLocale 从Gin上下文获取当前语言
func GetLocale(c *gin.Context) string {
	if locale, exists := c.Get(localeContextKey); exists {
		if strLocale, ok := locale.(string); ok {
			return strLocale
		}
	}
	return "en" // 默认返回英语
}

// 检查语言是否在支持列表中
func isSupported(locale string, supportedLocales []string) bool {
	for _, supported := range supportedLocales {
		if locale == supported {
			return true
		}
	}
	return false
}

// 从Accept-Language头解析语言
func extractLocaleFromHeader(header string, supportedLocales []string) string {
	if header == "" {
		return ""
	}

	// 解析Accept-Language头，格式如：zh-CN,zh;q=0.9,en;q=0.8
	parts := strings.Split(header, ",")
	localeMap := make(map[string]float64)

	for _, part := range parts {
		// 分离语言和权重
		subParts := strings.Split(strings.TrimSpace(part), ";")
		locale := strings.Split(subParts[0], "-")[0] // 只取主要部分，如zh-CN -> zh

		q := 1.0 // 默认权重
		if len(subParts) > 1 {
			qPart := strings.TrimSpace(subParts[1])
			if strings.HasPrefix(qPart, "q=") {
				if qf, err := parseFloat(qPart[2:]); err == nil {
					q = qf
				}
			}
		}

		// 保存权重最高的语言
		if currentQ, exists := localeMap[locale]; !exists || q > currentQ {
			localeMap[locale] = q
		}
	}

	// 按权重排序
	type LocaleWeight struct {
		locale string
		weight float64
	}

	var localeWeights []LocaleWeight
	for locale, weight := range localeMap {
		localeWeights = append(localeWeights, LocaleWeight{locale, weight})
	}

	// 简单的冒泡排序按权重降序排列
	for i := 0; i < len(localeWeights); i++ {
		for j := i + 1; j < len(localeWeights); j++ {
			if localeWeights[i].weight < localeWeights[j].weight {
				localeWeights[i], localeWeights[j] = localeWeights[j], localeWeights[i]
			}
		}
	}

	// 返回支持的最高权重语言
	for _, lw := range localeWeights {
		if isSupported(lw.locale, supportedLocales) {
			return lw.locale
		}
	}

	return ""
}

// 解析浮点数，简化版
func parseFloat(s string) (float64, error) {
	switch s {
	case "0":
		return 0, nil
	case "0.1":
		return 0.1, nil
	case "0.2":
		return 0.2, nil
	case "0.3":
		return 0.3, nil
	case "0.4":
		return 0.4, nil
	case "0.5":
		return 0.5, nil
	case "0.6":
		return 0.6, nil
	case "0.7":
		return 0.7, nil
	case "0.8":
		return 0.8, nil
	case "0.9":
		return 0.9, nil
	case "1":
		return 1.0, nil
	case "1.0":
		return 1.0, nil
	default:
		return 0, http.ErrNotSupported
	}
}
