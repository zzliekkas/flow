package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/zzliekkas/flow"
)

// CORSConfig 是CORS中间件的配置选项
type CORSConfig struct {
	// AllowOrigins 是允许的源列表，例如 ["https://example.com"]
	// 特殊的 "*" 表示允许所有源
	AllowOrigins []string

	// AllowMethods 是允许的HTTP方法列表
	// 默认是 ["GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"]
	AllowMethods []string

	// AllowHeaders 是允许的HTTP头部列表
	// 默认是 ["Origin", "Content-Type", "Content-Length", "Accept", "Authorization"]
	AllowHeaders []string

	// ExposeHeaders 是客户端可以访问的自定义头部列表
	ExposeHeaders []string

	// AllowCredentials 表示请求中是否可以包含用户凭证
	AllowCredentials bool

	// MaxAge 表示预检请求的结果可以缓存多长时间（秒）
	MaxAge int

	// AllowPrivateNetwork 指示是否允许私有网络请求
	AllowPrivateNetwork bool
}

// DefaultCORSConfig 返回CORS中间件的默认配置
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins:        []string{"*"},
		AllowMethods:        []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:        []string{"Origin", "Content-Type", "Content-Length", "Accept", "Authorization"},
		ExposeHeaders:       []string{},
		AllowCredentials:    false,
		MaxAge:              86400,
		AllowPrivateNetwork: false,
	}
}

// CORS 返回一个CORS中间件，允许所有源访问
func CORS() flow.HandlerFunc {
	return CORSWithConfig(DefaultCORSConfig())
}

// CORSWithConfig 返回一个使用指定配置的CORS中间件
func CORSWithConfig(config CORSConfig) flow.HandlerFunc {
	// 如果没有配置来源，使用默认的所有来源
	if len(config.AllowOrigins) == 0 {
		config.AllowOrigins = []string{"*"}
	}

	// 如果没有配置方法，使用默认方法
	if len(config.AllowMethods) == 0 {
		config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"}
	}

	// 如果没有配置头部，使用默认头部
	if len(config.AllowHeaders) == 0 {
		config.AllowHeaders = []string{"Origin", "Content-Type", "Content-Length", "Accept", "Authorization"}
	}

	// 方法和头部转换为大写
	allowMethods := normalizeHeaders(config.AllowMethods)
	allowHeaders := normalizeHeaders(config.AllowHeaders)
	exposeHeaders := normalizeHeaders(config.ExposeHeaders)

	// 处理所有源的情况
	allowAllOrigins := false
	allowOrigins := []string{}
	for _, origin := range config.AllowOrigins {
		if origin == "*" {
			allowAllOrigins = true
			break
		}
		allowOrigins = append(allowOrigins, strings.ToLower(origin))
	}

	return func(c *flow.Context) {
		// 如果这不是CORS请求，跳过
		origin := c.Request.Header.Get("Origin")
		if origin == "" {
			c.Next()
			return
		}

		// 始终设置Vary头部
		c.Header("Vary", "Origin")

		// 如果这是一个预检请求，处理预检
		if c.Request.Method == "OPTIONS" {
			// 检查请求方法
			reqMethod := c.Request.Header.Get("Access-Control-Request-Method")
			if reqMethod == "" {
				c.Next()
				return
			}

			// 处理预检请求
			handlePreflightRequest(c, config, allowAllOrigins, allowOrigins, allowMethods, allowHeaders)
			c.Status(http.StatusNoContent)
			return
		}

		// 简单请求或实际请求
		handleActualRequest(c, config, allowAllOrigins, allowOrigins, exposeHeaders)

		c.Next()
	}
}

// handlePreflightRequest 处理预检请求
func handlePreflightRequest(c *flow.Context, config CORSConfig, allowAllOrigins bool, allowOrigins, allowMethods, allowHeaders []string) {
	origin := c.Request.Header.Get("Origin")

	// 设置允许的来源
	if allowAllOrigins {
		c.Header("Access-Control-Allow-Origin", "*")
	} else {
		originLower := strings.ToLower(origin)
		allowed := false
		for _, allowOrigin := range allowOrigins {
			if allowOrigin == originLower {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		} else {
			c.Header("Access-Control-Allow-Origin", "")
		}
	}

	// 设置允许的方法
	c.Header("Access-Control-Allow-Methods", strings.Join(allowMethods, ", "))

	// 设置允许的头部
	reqHeaders := c.Request.Header.Get("Access-Control-Request-Headers")
	if reqHeaders != "" {
		c.Header("Access-Control-Allow-Headers", reqHeaders)
	} else {
		c.Header("Access-Control-Allow-Headers", strings.Join(allowHeaders, ", "))
	}

	// 设置允许凭证
	if config.AllowCredentials {
		c.Header("Access-Control-Allow-Credentials", "true")
	}

	// 设置缓存时间
	if config.MaxAge > 0 {
		c.Header("Access-Control-Max-Age", strconv.Itoa(config.MaxAge))
	}

	// 设置允许私有网络
	if config.AllowPrivateNetwork {
		c.Header("Access-Control-Allow-Private-Network", "true")
	}
}

// handleActualRequest 处理实际请求
func handleActualRequest(c *flow.Context, config CORSConfig, allowAllOrigins bool, allowOrigins []string, exposeHeaders []string) {
	origin := c.Request.Header.Get("Origin")

	// 设置允许的来源
	if allowAllOrigins {
		c.Header("Access-Control-Allow-Origin", "*")
	} else {
		originLower := strings.ToLower(origin)
		allowed := false
		for _, allowOrigin := range allowOrigins {
			if allowOrigin == originLower {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		} else {
			c.Header("Access-Control-Allow-Origin", "")
		}
	}

	// 设置允许凭证
	if config.AllowCredentials {
		c.Header("Access-Control-Allow-Credentials", "true")
	}

	// 设置允许访问的头部
	if len(exposeHeaders) > 0 {
		c.Header("Access-Control-Expose-Headers", strings.Join(exposeHeaders, ", "))
	}
}

// normalizeHeaders 将头部名称转换为大写，并移除重复的
func normalizeHeaders(headers []string) []string {
	if len(headers) == 0 {
		return headers
	}

	normalized := make([]string, 0, len(headers))
	seen := make(map[string]bool)

	for _, h := range headers {
		h = strings.ToUpper(h)
		if seen[h] {
			continue
		}

		normalized = append(normalized, h)
		seen[h] = true
	}

	return normalized
}
