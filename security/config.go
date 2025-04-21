package security

// Config 表示安全框架的配置
type Config struct {
	// 安全头部配置
	Headers HeadersConfig
	// XSS防护配置
	XSS XSSConfig
	// CSP配置
	CSP CSPConfig
	// 密码策略配置
	Password PasswordPolicyConfig
	// 审计日志配置
	Audit AuditConfig
	// CSRF防护配置
	CSRF CSRFConfig
	// 速率限制配置
	RateLimit RateLimitConfig
	// 加密配置
	Encryption EncryptionConfig
}

// HeadersConfig 定义安全头部的配置
type HeadersConfig struct {
	// 是否启用安全头部
	Enabled bool
	// X-Frame-Options 头部值，如 DENY, SAMEORIGIN
	FrameOptions string
	// X-Content-Type-Options 头部值，通常为 nosniff
	ContentTypeOptions string
	// X-XSS-Protection 头部值
	XSSProtection string
	// 是否启用 Referrer-Policy
	EnableReferrerPolicy bool
	// Referrer-Policy 头部值
	ReferrerPolicy string
	// HSTS 配置
	HSTS HSTSConfig
	// 跨域资源共享策略
	CORS CORSConfig
	// 功能策略 (Permissions-Policy)
	PermissionsPolicy map[string]string
	// 自定义头部
	CustomHeaders map[string]string
}

// CORSConfig 定义跨域资源共享配置
type CORSConfig struct {
	// 是否启用CORS
	Enabled bool
	// 允许的源
	AllowOrigins []string
	// 允许的方法
	AllowMethods []string
	// 允许的头部
	AllowHeaders []string
	// 暴露的头部
	ExposeHeaders []string
	// 是否允许凭证
	AllowCredentials bool
	// 预检请求缓存时间（秒）
	MaxAge int
}

// HSTSConfig 定义 HTTP Strict Transport Security 配置
type HSTSConfig struct {
	// 是否启用 HSTS
	Enabled bool
	// max-age 指令值（单位：秒）
	MaxAge int
	// 是否包含子域名
	IncludeSubDomains bool
	// 是否预加载
	Preload bool
}

// XSSConfig 定义 XSS 防护配置
type XSSConfig struct {
	// 是否启用 XSS 防护
	Enabled bool
	// 是否自动清理用户输入
	AutoSanitize bool
	// 是否在响应中转义 HTML
	EscapeHTML bool
	// 是否允许HTML (如果为false，将完全转义所有HTML)
	AllowHTML bool
	// 允许的HTML标签 (仅当AllowHTML为true时有效)
	AllowedTags []string
	// 允许的HTML属性 (仅当AllowHTML为true时有效)
	AllowedAttributes []string
	// 允许的URL协议 (用于href、src等属性)
	AllowedProtocols []string
	// 是否启用自动链接处理
	EnableLinkify bool
}

// CSPConfig 定义内容安全策略配置
type CSPConfig struct {
	// 是否启用 CSP
	Enabled bool
	// 默认来源策略
	DefaultSrc []string
	// 脚本来源策略
	ScriptSrc []string
	// 样式来源策略
	StyleSrc []string
	// 图片来源策略
	ImgSrc []string
	// 字体来源策略
	FontSrc []string
	// 对象来源策略
	ObjectSrc []string
	// 媒体来源策略
	MediaSrc []string
	// 框架来源策略
	FrameSrc []string
	// 连接来源策略
	ConnectSrc []string
	// 是否为仅报告模式
	ReportOnly bool
	// 报告 URI
	ReportURI string
	// 是否为每个请求启用 nonce
	EnableNonce bool
	// 沙盒指令
	Sandbox []string
	// 子来源策略
	ChildSrc []string
	// 表单提交目标
	FormAction []string
	// 框架祖先
	FrameAncestors []string
	// 插件类型
	PluginTypes []string
	// 基本URI
	BaseURI []string
	// Worker脚本源
	WorkerSrc []string
	// 清单源
	ManifestSrc []string
	// 预加载源
	PrefetchSrc []string
	// 需要SRI的元素
	RequireSriFor []string
	// 是否升级不安全请求
	UpgradeInsecureRequests bool
	// 是否阻止所有混合内容
	BlockAllMixedContent bool
}

// PasswordPolicyConfig 定义密码策略配置
type PasswordPolicyConfig struct {
	// 最小长度
	MinLength int
	// 是否需要大写字母
	RequireUpper bool
	// 是否需要小写字母
	RequireLower bool
	// 是否需要数字
	RequireNumber bool
	// 是否需要特殊字符
	RequireSpecial bool
	// 密码最大有效期（天）
	MaxAge int
	// 禁止使用的常见密码列表文件路径
	CommonPasswordsFilePath string
	// 历史密码检查数量（防止重用旧密码）
	HistoryCount int
}

// PasswordConfig 与 PasswordPolicyConfig 相同，用于兼容现有代码
type PasswordConfig PasswordPolicyConfig

// AuditConfig 定义审计日志配置
type AuditConfig struct {
	// 是否启用审计
	Enabled bool
	// 是否记录到文件
	LogToFile bool
	// 日志文件路径
	FilePath string
	// 是否记录到数据库
	LogToDatabase bool
	// 数据库连接字符串
	DatabaseDSN string
	// 审计日志级别: 0-错误，1-警告，2-信息，3-调试
	LogLevel int
	// 日志目标: file, database, console, webhook
	Destination string
	// 是否记录认证事件
	LogAuthenticationEvents bool
	// 是否记录访问控制事件
	LogAccessControl bool
	// 是否记录数据访问事件
	LogDataAccess bool
	// 是否记录敏感操作事件
	LogSensitiveActions bool
}

// CSRFConfig 定义 CSRF 防护配置
type CSRFConfig struct {
	// 是否启用 CSRF 防护
	Enabled bool
	// 令牌有效期（秒）
	TokenExpiry int
	// 是否在 cookie 中存储令牌
	UseCookie bool
	// Cookie 名称
	CookieName string
	// Cookie 路径
	CookiePath string
	// Cookie 域
	CookieDomain string
	// 是否启用安全 cookie（仅 HTTPS）
	SecureCookie bool
	// 是否启用 HTTP-only cookie
	HttpOnlyCookie bool
	// 表单字段名称
	FormFieldName string
	// 头部名称
	HeaderName string
}

// RateLimitConfig 定义速率限制配置
type RateLimitConfig struct {
	// 是否启用速率限制
	Enabled bool
	// 时间窗口（秒）
	Window int
	// 窗口内最大请求数
	MaxRequests int
	// 是否使用客户端 IP 作为限制依据
	UseClientIP bool
	// 是否使用 JWT 标识符作为限制依据
	UseJWTIdentifier bool
	// 存储类型：memory, redis
	StorageType string
	// Redis 连接字符串（如果使用）
	RedisDSN string
}

// EncryptionConfig 定义加密配置
type EncryptionConfig struct {
	// 是否启用加密
	Enabled bool
	// 加密算法：AES, ChaCha20
	Algorithm string
	// 密钥（建议通过环境变量设置）
	Key string
	// 向量（建议通过环境变量设置）
	IV string
}

// DefaultConfig 返回默认安全配置
func DefaultConfig() Config {
	return Config{
		Headers: HeadersConfig{
			Enabled:              true,
			FrameOptions:         "SAMEORIGIN",
			ContentTypeOptions:   "nosniff",
			XSSProtection:        "1; mode=block",
			EnableReferrerPolicy: true,
			ReferrerPolicy:       "strict-origin-when-cross-origin",
			HSTS: HSTSConfig{
				Enabled:           true,
				MaxAge:            31536000, // 1年
				IncludeSubDomains: true,
				Preload:           false,
			},
		},
		XSS: XSSConfig{
			Enabled:      true,
			AutoSanitize: true,
			EscapeHTML:   true,
			AllowHTML:    true,
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
		},
		CSP: CSPConfig{
			Enabled:     true,
			DefaultSrc:  []string{"'self'"},
			ScriptSrc:   []string{"'self'"},
			StyleSrc:    []string{"'self'"},
			ImgSrc:      []string{"'self'", "data:"},
			FontSrc:     []string{"'self'"},
			ObjectSrc:   []string{"'none'"},
			MediaSrc:    []string{"'self'"},
			FrameSrc:    []string{"'self'"},
			ConnectSrc:  []string{"'self'"},
			ReportOnly:  false,
			EnableNonce: true,
		},
		Password: PasswordPolicyConfig{
			MinLength:      8,
			RequireUpper:   true,
			RequireLower:   true,
			RequireNumber:  true,
			RequireSpecial: true,
			MaxAge:         90, // 90天
			HistoryCount:   5,
		},
		Audit: AuditConfig{
			Enabled:                 true,
			LogToFile:               true,
			FilePath:                "./logs/security_audit.log",
			LogLevel:                2, // 信息级别
			Destination:             "file",
			LogAuthenticationEvents: true,
			LogAccessControl:        true,
			LogDataAccess:           true,
			LogSensitiveActions:     true,
		},
		CSRF: CSRFConfig{
			Enabled:        true,
			TokenExpiry:    3600, // 1小时
			UseCookie:      true,
			CookieName:     "_csrf",
			CookiePath:     "/",
			SecureCookie:   true,
			HttpOnlyCookie: true,
			FormFieldName:  "_csrf",
			HeaderName:     "X-CSRF-Token",
		},
		RateLimit: RateLimitConfig{
			Enabled:     true,
			Window:      60, // 1分钟
			MaxRequests: 100,
			UseClientIP: true,
			StorageType: "memory",
		},
		Encryption: EncryptionConfig{
			Enabled:   true,
			Algorithm: "AES",
		},
	}
}

// Merge 将另一个配置合并到当前配置
func (c *Config) Merge(other Config) {
	// 实现合并逻辑
	// ...
}
