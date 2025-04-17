package middleware

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/zzliekkas/flow"
)

// CSRF错误常量
var (
	ErrCSRFTokenInvalid   = errors.New("CSRF令牌无效")
	ErrCSRFTokenNotFound  = errors.New("未找到CSRF令牌")
	ErrCSRFTokenUsed      = errors.New("CSRF令牌已使用")
	ErrCSRFTokenExpired   = errors.New("CSRF令牌已过期")
	ErrCSRFHeaderNotSet   = errors.New("缺少CSRF头")
	ErrCSRFCookieNotFound = errors.New("CSRF Cookie不存在")
)

// 默认设置
const (
	DefaultCSRFCookieName     = "csrf_token"
	DefaultCSRFHeaderName     = "X-CSRF-TOKEN"
	DefaultCSRFParameterName  = "_csrf"
	DefaultCSRFContextKey     = "csrf"
	DefaultCSRFTokenLength    = 32
	DefaultCSRFTokenExpiry    = time.Hour * 24
	DefaultCSRFSingleUseToken = false
)

// TokenStore 定义了CSRF令牌存储接口
type TokenStore interface {
	// GenerateToken 生成一个新的CSRF令牌
	GenerateToken(ctx *flow.Context) (string, error)

	// ValidateToken 验证CSRF令牌
	ValidateToken(ctx *flow.Context, token string) bool

	// DeleteToken 删除CSRF令牌
	DeleteToken(ctx *flow.Context, token string) error
}

// CSRFConfig 定义CSRF中间件配置
type CSRFConfig struct {
	// TokenLength 指定CSRF令牌的长度
	TokenLength int

	// TokenExpiry 指定CSRF令牌的过期时间
	TokenExpiry time.Duration

	// CookieName 指定存储CSRF令牌的Cookie名称
	CookieName string

	// CookiePath 指定Cookie的路径
	CookiePath string

	// CookieDomain 指定Cookie的域
	CookieDomain string

	// CookieSecure 指定Cookie是否只通过HTTPS发送
	CookieSecure bool

	// CookieHTTPOnly 指定Cookie是否只能通过HTTP访问
	CookieHTTPOnly bool

	// CookieSameSite 指定Cookie的SameSite策略
	CookieSameSite http.SameSite

	// HeaderName 指定CSRF头名称
	HeaderName string

	// ParameterName 指定CSRF表单参数名称
	ParameterName string

	// ContextKey 指定存储在上下文中的键名
	ContextKey string

	// Secret 用于签名CSRF令牌的密钥
	Secret string

	// ErrorFunc 自定义错误处理函数
	ErrorFunc func(*flow.Context, error)

	// TokenStore 令牌存储接口
	TokenStore TokenStore

	// ExcludeURLs 排除的URL列表
	ExcludeURLs []string

	// AllowedOrigins 允许的源列表
	AllowedOrigins []string

	// AllowedMethods 允许的HTTP方法列表，不检查CSRF
	AllowedMethods []string

	// SingleUseToken 是否为单次使用令牌
	SingleUseToken bool
}

// DefaultCSRFConfig 返回默认CSRF配置
func DefaultCSRFConfig() CSRFConfig {
	return CSRFConfig{
		TokenLength:    DefaultCSRFTokenLength,
		TokenExpiry:    DefaultCSRFTokenExpiry,
		CookieName:     DefaultCSRFCookieName,
		CookiePath:     "/",
		CookieHTTPOnly: true,
		CookieSameSite: http.SameSiteLaxMode,
		HeaderName:     DefaultCSRFHeaderName,
		ParameterName:  DefaultCSRFParameterName,
		ContextKey:     DefaultCSRFContextKey,
		TokenStore:     NewMemoryTokenStore(),
		AllowedMethods: []string{"GET", "HEAD", "OPTIONS"},
		SingleUseToken: DefaultCSRFSingleUseToken,
		ErrorFunc: func(c *flow.Context, err error) {
			c.JSON(http.StatusForbidden, flow.H{
				"error": err.Error(),
			})
		},
	}
}

// MemoryTokenStore 内存令牌存储实现
type MemoryTokenStore struct {
	// tokens 存储有效的CSRF令牌
	tokens map[string]time.Time
	// usedTokens 存储已使用的CSRF令牌
	usedTokens map[string]bool
	// 互斥锁
	mutex sync.RWMutex
}

// NewMemoryTokenStore 创建一个新的内存令牌存储
func NewMemoryTokenStore() *MemoryTokenStore {
	store := &MemoryTokenStore{
		tokens:     make(map[string]time.Time),
		usedTokens: make(map[string]bool),
	}

	// 启动后台清理例程
	go store.cleanup()

	return store
}

// cleanup 定期清理过期的令牌
func (s *MemoryTokenStore) cleanup() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		s.mutex.Lock()
		now := time.Now()

		// 清理过期令牌
		for token, expiry := range s.tokens {
			if now.After(expiry) {
				delete(s.tokens, token)
				delete(s.usedTokens, token)
			}
		}

		s.mutex.Unlock()
	}
}

// GenerateToken 生成一个新的CSRF令牌
func (s *MemoryTokenStore) GenerateToken(ctx *flow.Context) (string, error) {
	// 生成随机字节
	b := make([]byte, DefaultCSRFTokenLength)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	// 编码为Base64字符串
	token := base64.StdEncoding.EncodeToString(b)

	// 存储令牌
	s.mutex.Lock()
	s.tokens[token] = time.Now().Add(DefaultCSRFTokenExpiry)
	s.mutex.Unlock()

	return token, nil
}

// ValidateToken 验证CSRF令牌
func (s *MemoryTokenStore) ValidateToken(ctx *flow.Context, token string) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// 检查令牌是否存在
	expiry, exists := s.tokens[token]
	if !exists {
		return false
	}

	// 检查令牌是否过期
	if time.Now().After(expiry) {
		return false
	}

	// 检查令牌是否已被使用（如果启用单次使用）
	if DefaultCSRFSingleUseToken {
		if s.usedTokens[token] {
			return false
		}
		s.usedTokens[token] = true
	}

	return true
}

// DeleteToken 删除CSRF令牌
func (s *MemoryTokenStore) DeleteToken(ctx *flow.Context, token string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.tokens, token)
	delete(s.usedTokens, token)

	return nil
}

// RedisTokenStore Redis令牌存储实现
type RedisTokenStore struct {
	// TODO: 实现Redis存储
}

// CSRF 使用默认配置创建CSRF中间件
func CSRF() flow.HandlerFunc {
	return CSRFWithConfig(DefaultCSRFConfig())
}

// CSRFWithConfig 使用自定义配置创建CSRF中间件
func CSRFWithConfig(config CSRFConfig) flow.HandlerFunc {
	// 确保配置有效
	if config.TokenLength == 0 {
		config.TokenLength = DefaultCSRFTokenLength
	}
	if config.TokenExpiry == 0 {
		config.TokenExpiry = DefaultCSRFTokenExpiry
	}
	if config.CookieName == "" {
		config.CookieName = DefaultCSRFCookieName
	}
	if config.HeaderName == "" {
		config.HeaderName = DefaultCSRFHeaderName
	}
	if config.ParameterName == "" {
		config.ParameterName = DefaultCSRFParameterName
	}
	if config.ContextKey == "" {
		config.ContextKey = DefaultCSRFContextKey
	}
	if config.TokenStore == nil {
		config.TokenStore = NewMemoryTokenStore()
	}
	if config.ErrorFunc == nil {
		config.ErrorFunc = func(c *flow.Context, err error) {
			c.JSON(http.StatusForbidden, flow.H{
				"error": err.Error(),
			})
		}
	}
	if len(config.AllowedMethods) == 0 {
		config.AllowedMethods = []string{"GET", "HEAD", "OPTIONS"}
	}

	return func(c *flow.Context) {
		// 检查是否跳过此URL
		for _, url := range config.ExcludeURLs {
			if strings.HasPrefix(c.Request.URL.Path, url) {
				c.Next()
				return
			}
		}

		// 检查是否为安全方法（不需要CSRF保护）
		method := c.Request.Method
		for _, m := range config.AllowedMethods {
			if method == m {
				// 安全方法，只需要设置令牌
				setCSRFToken(c, config)
				c.Next()
				return
			}
		}

		// 不安全方法，需要验证CSRF令牌
		token := extractCSRFToken(c, config)
		if token == "" {
			config.ErrorFunc(c, ErrCSRFTokenNotFound)
			c.Abort()
			return
		}

		// 验证令牌
		valid := config.TokenStore.ValidateToken(c, token)
		if !valid {
			config.ErrorFunc(c, ErrCSRFTokenInvalid)
			c.Abort()
			return
		}

		// 如果是单次使用令牌，在验证后删除
		if config.SingleUseToken {
			config.TokenStore.DeleteToken(c, token)

			// 生成新令牌
			setCSRFToken(c, config)
		}

		c.Next()
	}
}

// setCSRFToken 设置CSRF令牌到Cookie和上下文
func setCSRFToken(c *flow.Context, config CSRFConfig) {
	// 获取现有令牌
	var token string
	cookie, err := c.Cookie(config.CookieName)

	if err == nil && cookie != "" {
		// 验证现有令牌
		if config.TokenStore.ValidateToken(c, cookie) {
			token = cookie
		}
	}

	// 如果没有有效令牌，生成新令牌
	if token == "" {
		var err error
		token, err = config.TokenStore.GenerateToken(c)
		if err != nil {
			return
		}

		// 设置Cookie
		c.SetCookie(
			config.CookieName,
			token,
			int(config.TokenExpiry.Seconds()),
			config.CookiePath,
			config.CookieDomain,
			config.CookieSecure,
			config.CookieHTTPOnly,
		)
	}

	// 将令牌存储在上下文中
	c.Set(config.ContextKey, token)
}

// extractCSRFToken 从请求中提取CSRF令牌
func extractCSRFToken(c *flow.Context, config CSRFConfig) string {
	// 优先从头部获取
	token := c.GetHeader(config.HeaderName)
	if token != "" {
		return token
	}

	// 尝试从表单参数获取
	token = c.Request.FormValue(config.ParameterName)
	if token != "" {
		return token
	}

	// 最后从Cookie获取
	token, _ = c.Cookie(config.CookieName)
	return token
}

// GenerateCSRFHTML 生成包含CSRF令牌的HTML表单字段
func GenerateCSRFHTML(c *flow.Context) template.HTML {
	token, exists := c.Get(DefaultCSRFContextKey)
	if !exists {
		return ""
	}

	return template.HTML(fmt.Sprintf(`<input type="hidden" name="%s" value="%s">`,
		DefaultCSRFParameterName, token))
}

// GetCSRFToken 获取当前的CSRF令牌
func GetCSRFToken(c *flow.Context) string {
	token, exists := c.Get(DefaultCSRFContextKey)
	if !exists {
		return ""
	}

	return token.(string)
}

// CSRFRequired 提供更细粒度的验证，可在特定路由上启用
func CSRFRequired(config CSRFConfig) flow.HandlerFunc {
	return func(c *flow.Context) {
		token := extractCSRFToken(c, config)
		if token == "" {
			config.ErrorFunc(c, ErrCSRFTokenNotFound)
			c.Abort()
			return
		}

		// 验证令牌
		valid := config.TokenStore.ValidateToken(c, token)
		if !valid {
			config.ErrorFunc(c, ErrCSRFTokenInvalid)
			c.Abort()
			return
		}

		// 如果是单次使用令牌，在验证后删除
		if config.SingleUseToken {
			config.TokenStore.DeleteToken(c, token)

			// 生成新令牌
			setCSRFToken(c, config)
		}

		c.Next()
	}
}

// ProtectForm 保护表单免受CSRF攻击的辅助函数
func ProtectForm(formHTML string, c *flow.Context) string {
	csrfField := GenerateCSRFHTML(c)
	return strings.Replace(
		formHTML,
		"</form>",
		fmt.Sprintf("%s</form>", csrfField),
		-1,
	)
}

// GenerateCSRFToken 生成CSRF令牌并返回
func GenerateCSRFToken(config CSRFConfig) flow.HandlerFunc {
	return func(c *flow.Context) {
		setCSRFToken(c, config)
		token := GetCSRFToken(c)

		c.JSON(http.StatusOK, flow.H{
			"token": token,
		})
	}
}

// GenerateCSRFTokenHash 根据密钥生成令牌哈希
func GenerateCSRFTokenHash(token, secret string) string {
	h := sha256.New()
	h.Write([]byte(token + secret))
	return fmt.Sprintf("%x", h.Sum(nil))
}
