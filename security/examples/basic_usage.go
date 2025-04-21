package examples

import (
	"net/http"

	"github.com/zzliekkas/flow"
	"github.com/zzliekkas/flow/security"
)

func BasicSecurityExample() {
	// 初始化Flow引擎
	app := flow.New()

	// 创建安全配置
	config := security.Config{
		// 启用安全头部
		Headers: security.HeadersConfig{
			Enabled: true,
			// 设置框架选项
			FrameOptions: "SAMEORIGIN",
			// 启用严格传输安全
			HSTS: security.HSTSConfig{
				Enabled:           true,
				MaxAge:            31536000, // 1年
				IncludeSubDomains: true,
				Preload:           true,
			},
		},
		// 启用XSS防护
		XSS: security.XSSConfig{
			Enabled: true,
		},
		// 启用内容安全策略
		CSP: security.CSPConfig{
			Enabled:    true,
			DefaultSrc: []string{"'self'"},
			ScriptSrc:  []string{"'self'"},
			StyleSrc:   []string{"'self'"},
			ImgSrc:     []string{"'self'", "data:"},
			FontSrc:    []string{"'self'"},
			ObjectSrc:  []string{"'none'"},
			ReportOnly: false,
		},
		// 配置密码策略
		Password: security.PasswordPolicyConfig{
			MinLength:      8,
			RequireUpper:   true,
			RequireLower:   true,
			RequireNumber:  true,
			RequireSpecial: true,
			MaxAge:         90, // 密码有效期90天
		},
		// 启用审计日志
		Audit: security.AuditConfig{
			Enabled:   true,
			LogToFile: true,
			FilePath:  "./logs/security_audit.log",
		},
	}

	// 创建安全管理器
	securityManager := security.NewManager(config)

	// 将安全中间件集成到应用程序
	securityManager.IntegrateWithExistingMiddleware(app)

	// 使用CSP nonce中间件
	app.Use(securityManager.CSPNonceMiddleware())

	// 定义一个受保护的路由
	app.GET("/protected", func(c *flow.Context) {
		// 获取CSP nonce
		nonce := security.GetCSPNonce(c)

		// 使用CSP nonce渲染内联脚本
		html := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>安全框架示例</title>
		</head>
		<body>
			<h1>安全框架示例</h1>
			<p>此页面已应用了安全头部和CSP策略。</p>
			<script nonce="` + nonce + `">
				console.log("这个脚本使用nonce属性，被CSP允许执行");
			</script>
		</body>
		</html>`

		// 返回HTML响应
		c.HTML(http.StatusOK, html, nil)
	})

	// 定义一个用于演示表单净化的路由
	app.POST("/submit", func(c *flow.Context) {
		// 获取并净化用户输入
		userInput := c.PostForm("user_input")
		cleanInput := securityManager.SanitizeHTML(userInput)

		// 记录敏感操作
		securityManager.GetAuditLogger().LogSensitiveAction(
			c.ClientIP(),
			"form_submission",
			"user_input",
			true,
			map[string]interface{}{
				"input_length":   len(userInput),
				"cleaned_length": len(cleanInput),
			},
		)

		// 返回净化后的输入
		c.String(http.StatusOK, "净化后的输入: %s", cleanInput)
	})

	// 密码验证示例
	app.POST("/validate-password", func(c *flow.Context) {
		password := c.PostForm("password")

		// 验证密码是否符合策略
		err := securityManager.ValidatePassword(password)

		if err == nil {
			c.String(http.StatusOK, "密码符合安全策略")
		} else {
			if validErr, ok := err.(*security.PasswordValidationError); ok {
				c.JSON(http.StatusBadRequest, map[string]interface{}{
					"valid":   false,
					"reasons": validErr.Reasons,
				})
			} else {
				c.JSON(http.StatusBadRequest, map[string]interface{}{
					"valid":   false,
					"reasons": []string{err.Error()},
				})
			}
		}
	})

	// 启动服务器
	app.Run(":8080")
}

// CSPNonceExample 展示如何在模板中使用CSP nonce
func CSPNonceExample() {
	app := flow.New()

	// 创建安全配置
	config := security.Config{
		CSP: security.CSPConfig{
			Enabled:    true,
			DefaultSrc: []string{"'self'"},
			// 不在这里添加nonce，而是在中间件中动态添加
			ScriptSrc: []string{"'self'"},
		},
	}

	securityManager := security.NewManager(config)

	// 使用CSP nonce中间件
	app.Use(securityManager.CSPNonceMiddleware())

	// 定义路由
	app.GET("/", func(c *flow.Context) {
		// 获取为此请求生成的nonce
		nonce := security.GetCSPNonce(c)

		// 使用nonce渲染HTML
		html := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>CSP Nonce示例</title>
		</head>
		<body>
			<h1>CSP Nonce示例</h1>
			
			<!-- 内联脚本使用nonce -->
			<script nonce="` + nonce + `">
				document.body.style.backgroundColor = '#f0f0f0';
			</script>
			
			<!-- 没有nonce的内联脚本将被CSP阻止 -->
			<script>
				alert('此脚本将被CSP阻止');
			</script>
		</body>
		</html>`

		c.HTML(http.StatusOK, html, nil)
	})

	app.Run(":8080")
}

// AuditLogExample 展示如何使用审计日志功能
func AuditLogExample() {
	app := flow.New()

	// 创建配置
	config := security.Config{
		Audit: security.AuditConfig{
			Enabled:   true,
			LogToFile: true,
			FilePath:  "./logs/audit.log",
		},
	}

	securityManager := security.NewManager(config)

	// 在中间件中使用审计日志
	app.Use(func(c *flow.Context) {
		// 在请求开始时记录访问
		securityManager.GetAuditLogger().LogAccessControl(
			c.ClientIP(),
			"page_access",
			c.Request.URL.Path,
			true,
			map[string]interface{}{
				"path":   c.Request.URL.Path,
				"method": c.Request.Method,
			},
		)

		// 继续处理请求
		c.Next()

		// 在请求结束时记录状态码
		securityManager.GetAuditLogger().LogEvent(
			"response",
			c.ClientIP(),
			"http_response",
			c.Request.URL.Path,
			true,
			map[string]interface{}{
				"status_code": c.Writer.Status(),
				"path":        c.Request.URL.Path,
			},
		)
	})

	// 定义登录路由
	app.POST("/login", func(c *flow.Context) {
		username := c.PostForm("username")

		// 记录认证事件
		securityManager.GetAuditLogger().LogAuthentication(
			username,
			true,
			c.ClientIP(),
			map[string]interface{}{
				"username": username,
				"success":  true, // 假设登录成功
			},
		)

		c.String(http.StatusOK, "登录成功")
	})

	app.Run(":8080")
}
