package middleware

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/zzliekkas/flow"
)

// JWT错误定义
var (
	ErrJWTTokenMissing      = errors.New("令牌缺失")
	ErrJWTTokenInvalid      = errors.New("令牌无效")
	ErrJWTTokenExpired      = errors.New("令牌已过期")
	ErrJWTTokenNotActive    = errors.New("令牌尚未生效")
	ErrJWTUnexpectedSigning = errors.New("意外的签名方法")
)

// JWTConfig JWT中间件的配置
type JWTConfig struct {
	// 从请求中提取token的函数
	TokenExtractor func(*flow.Context) (string, error)

	// 用于签名的密钥
	SigningKey interface{}

	// 用于验证的密钥
	ValidationKey interface{}

	// 签名方法
	SigningMethod jwt.SigningMethod

	// 令牌在上下文中的键
	ContextKey string

	// 处理令牌错误的回调函数
	ErrorHandler func(*flow.Context, error)

	// 路径白名单，这些路径不需要检查令牌
	Skipper func(*flow.Context) bool

	// 令牌查找顺序：1. 查询参数，2. Cookie，3. 请求头
	TokenLookup string

	// 请求头中包含令牌的头部名称
	AuthScheme string

	// 是否使用令牌黑名单
	UseBlacklist bool

	// 黑名单检查函数
	BlacklistCheck func(string) bool
}

// TokenExtractorFunc 函数类型用于从请求中提取令牌
type TokenExtractorFunc func(*flow.Context) (string, error)

// DefaultJWTConfig 返回JWT中间件的默认配置
func DefaultJWTConfig() JWTConfig {
	return JWTConfig{
		SigningMethod: jwt.SigningMethodHS256,
		ContextKey:    "user",
		TokenLookup:   "header:Authorization",
		AuthScheme:    "Bearer",
		ErrorHandler: func(c *flow.Context, err error) {
			c.JSON(http.StatusUnauthorized, flow.H{
				"error": err.Error(),
			})
		},
		Skipper: func(c *flow.Context) bool {
			return false
		},
		UseBlacklist: false,
		BlacklistCheck: func(token string) bool {
			return false
		},
	}
}

// JWT 返回使用默认配置的JWT中间件
func JWT(key interface{}) flow.HandlerFunc {
	config := DefaultJWTConfig()
	config.SigningKey = key
	config.ValidationKey = key

	return JWTWithConfig(config)
}

// JWTWithConfig 返回使用自定义配置的JWT中间件
func JWTWithConfig(config JWTConfig) flow.HandlerFunc {
	// 验证配置
	if config.SigningKey == nil && config.ValidationKey == nil {
		panic("jwt: 验证密钥是必需的")
	}
	if config.SigningMethod == nil {
		config.SigningMethod = jwt.SigningMethodHS256
	}
	if config.ContextKey == "" {
		config.ContextKey = "user"
	}
	if config.ErrorHandler == nil {
		config.ErrorHandler = func(c *flow.Context, err error) {
			c.JSON(http.StatusUnauthorized, flow.H{
				"error": err.Error(),
			})
		}
	}
	if config.Skipper == nil {
		config.Skipper = func(c *flow.Context) bool {
			return false
		}
	}

	// 如果没有提供令牌提取器，使用默认提取器
	if config.TokenExtractor == nil {
		config.TokenExtractor = fromHeaderOrQueryOrCookie(config.TokenLookup, config.AuthScheme)
	}

	return func(c *flow.Context) {
		// 如果当前请求在跳过列表中，则跳过验证
		if config.Skipper(c) {
			c.Next()
			return
		}

		// 从请求中提取令牌
		tokenString, err := config.TokenExtractor(c)
		if err != nil {
			config.ErrorHandler(c, ErrJWTTokenMissing)
			return
		}

		// 解析并验证令牌
		token, err := jwt.Parse(tokenString, keyFuncForConfig(config))
		if err != nil {
			// 根据错误信息判断错误类型
			if err.Error() == "token is expired" {
				config.ErrorHandler(c, ErrJWTTokenExpired)
				return
			}

			if err.Error() == "token is not valid yet" {
				config.ErrorHandler(c, ErrJWTTokenNotActive)
				return
			}

			config.ErrorHandler(c, ErrJWTTokenInvalid)
			return
		}

		// 检查令牌是否有效
		if !token.Valid {
			config.ErrorHandler(c, ErrJWTTokenInvalid)
			return
		}

		// 检查令牌是否在黑名单中
		if config.UseBlacklist && config.BlacklistCheck(tokenString) {
			config.ErrorHandler(c, ErrJWTTokenInvalid)
			return
		}

		// 将令牌存储在上下文中
		c.Set(config.ContextKey, token)

		c.Next()
	}
}

// keyFuncForConfig 基于配置创建用于令牌验证的密钥函数
func keyFuncForConfig(config JWTConfig) jwt.Keyfunc {
	return func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if token.Method.Alg() != config.SigningMethod.Alg() {
			return nil, ErrJWTUnexpectedSigning
		}

		// 使用验证密钥(如果提供)，否则使用签名密钥
		if config.ValidationKey != nil {
			return config.ValidationKey, nil
		}

		return config.SigningKey, nil
	}
}

// fromHeaderOrQueryOrCookie 创建从请求头、查询参数或Cookie中提取令牌的函数
func fromHeaderOrQueryOrCookie(tokenLookup string, authScheme string) TokenExtractorFunc {
	parts := strings.Split(tokenLookup, ":")
	extractor := tokenFromHeader(authScheme)

	if len(parts) == 2 {
		switch parts[0] {
		case "query":
			extractor = tokenFromQuery(parts[1])
		case "cookie":
			extractor = tokenFromCookie(parts[1])
		case "header":
			extractor = tokenFromHeader(authScheme)
		}
	}

	return extractor
}

// tokenFromHeader 从请求头中提取令牌
func tokenFromHeader(authScheme string) TokenExtractorFunc {
	return func(c *flow.Context) (string, error) {
		auth := c.Request.Header.Get("Authorization")
		if auth == "" {
			return "", ErrJWTTokenMissing
		}

		l := len(authScheme)
		if len(auth) > l+1 && auth[:l] == authScheme {
			return auth[l+1:], nil
		}

		return "", ErrJWTTokenMissing
	}
}

// tokenFromQuery 从查询参数中提取令牌
func tokenFromQuery(param string) TokenExtractorFunc {
	return func(c *flow.Context) (string, error) {
		token := c.Query(param)
		if token == "" {
			return "", ErrJWTTokenMissing
		}
		return token, nil
	}
}

// tokenFromCookie 从Cookie中提取令牌
func tokenFromCookie(name string) TokenExtractorFunc {
	return func(c *flow.Context) (string, error) {
		cookie, err := c.Request.Cookie(name)
		if err != nil {
			return "", ErrJWTTokenMissing
		}
		return cookie.Value, nil
	}
}

// CreateToken 创建一个新的JWT令牌
func CreateToken(claims jwt.Claims, key interface{}, method jwt.SigningMethod) (string, error) {
	// 创建令牌
	token := jwt.NewWithClaims(method, claims)

	// 签名令牌
	return token.SignedString(key)
}

// CreateTokenWithExp 创建一个带有过期时间的令牌
func CreateTokenWithExp(id string, issuer string, audience string, subject string, expiry time.Duration, key interface{}, method jwt.SigningMethod) (string, error) {
	now := time.Now()

	// 创建标准声明
	claims := jwt.RegisteredClaims{
		ID:        id,
		Issuer:    issuer,
		Subject:   subject,
		Audience:  jwt.ClaimStrings{audience},
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
	}

	// 创建并返回令牌
	return CreateToken(claims, key, method)
}
