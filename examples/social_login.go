package examples

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/zzliekkas/flow/auth/drivers"
)

// User 代表应用程序中的用户模型
type User struct {
	ID       int64
	Name     string
	Email    string
	Avatar   string
	Provider string
	SocialID string
}

// UserRepository 实现了UserRepository接口
type UserRepository struct {
	// 在实际应用中，这里应该有数据库连接
	users  map[string]*User // 使用SocialID作为键
	lastID int64
}

// FindUserBySocialID 通过社交平台ID查找用户
func (r *UserRepository) FindUserBySocialID(ctx context.Context, provider, socialID string) (interface{}, error) {
	key := fmt.Sprintf("%s:%s", provider, socialID)
	if user, ok := r.users[key]; ok {
		return user, nil
	}
	return nil, nil
}

// CreateUser 创建新用户
func (r *UserRepository) CreateUser(ctx context.Context, user interface{}) error {
	u, ok := user.(*User)
	if !ok {
		return fmt.Errorf("无效的用户类型")
	}

	r.lastID++
	u.ID = r.lastID

	key := fmt.Sprintf("%s:%s", u.Provider, u.SocialID)
	r.users[key] = u

	return nil
}

// GetUsers 返回所有用户
func (r *UserRepository) GetUsers() []*User {
	result := make([]*User, 0, len(r.users))
	for _, user := range r.users {
		result = append(result, user)
	}
	return result
}

// 应用程序配置
type Config struct {
	GithubClientID     string
	GithubClientSecret string
	GoogleClientID     string
	GoogleClientSecret string
	WeChatAppID        string
	WeChatAppSecret    string
	BaseURL            string
}

// 加载配置（实际应用中可以从环境变量或配置文件加载）
func loadConfig() Config {
	return Config{
		GithubClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		GithubClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		WeChatAppID:        os.Getenv("WECHAT_APP_ID"),
		WeChatAppSecret:    os.Getenv("WECHAT_APP_SECRET"),
		BaseURL:            "http://localhost:8080",
	}
}

func main() {
	// 加载配置
	config := loadConfig()

	// 创建用户存储库
	userRepo := &UserRepository{
		users:  make(map[string]*User),
		lastID: 0,
	}

	// 创建社交登录管理器
	socialManager := drivers.NewSocialManager(userRepo)

	// 设置用户创建回调
	socialManager.SetCreateUserCallback(func(ctx context.Context, socialUser *drivers.SocialUser) (interface{}, error) {
		user := &User{
			Name:     socialUser.Name,
			Email:    socialUser.Email,
			Avatar:   socialUser.Avatar,
			Provider: socialUser.Provider,
			SocialID: socialUser.ID,
		}
		return user, nil
	})

	// 注册GitHub提供商（如果配置了）
	if config.GithubClientID != "" && config.GithubClientSecret != "" {
		githubConfig := map[string]interface{}{
			"client_id":     config.GithubClientID,
			"client_secret": config.GithubClientSecret,
			"redirect_url":  config.BaseURL + "/auth/github/callback",
			"scopes":        []string{"user:email"},
		}
		socialManager.RegisterProvider(drivers.NewGitHubProvider(githubConfig))
		log.Println("已注册GitHub登录提供商")
	}

	// 注册Google提供商（如果配置了）
	if config.GoogleClientID != "" && config.GoogleClientSecret != "" {
		googleConfig := map[string]interface{}{
			"client_id":     config.GoogleClientID,
			"client_secret": config.GoogleClientSecret,
			"redirect_url":  config.BaseURL + "/auth/google/callback",
			"scopes":        []string{"profile", "email"},
		}
		socialManager.RegisterProvider(drivers.NewGoogleProvider(googleConfig))
		log.Println("已注册Google登录提供商")
	}

	// 注册微信提供商（如果配置了）
	if config.WeChatAppID != "" && config.WeChatAppSecret != "" {
		wechatConfig := map[string]interface{}{
			"client_id":     config.WeChatAppID,
			"client_secret": config.WeChatAppSecret,
			"redirect_url":  config.BaseURL + "/auth/wechat/callback",
			"scopes":        []string{"snsapi_login"},
		}
		socialManager.RegisterProvider(drivers.NewWeChatProvider(wechatConfig))
		log.Println("已注册微信登录提供商")
	}

	// 创建路由处理器
	mux := http.NewServeMux()

	// 首页 - 显示登录页面
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		// 显示登录页面
		tmpl := template.Must(template.New("login").Parse(`
		<!DOCTYPE html>
		<html>
		<head>
			<title>社交登录示例</title>
			<style>
				body {
					font-family: Arial, sans-serif;
					max-width: 800px;
					margin: 0 auto;
					padding: 20px;
				}
				h1 {
					color: #333;
				}
				.login-buttons {
					display: flex;
					flex-direction: column;
					gap: 10px;
					margin-top: 20px;
				}
				.login-button {
					display: inline-block;
					padding: 10px 15px;
					background-color: #f5f5f5;
					border: 1px solid #ddd;
					border-radius: 4px;
					text-decoration: none;
					color: #333;
					font-weight: bold;
				}
				.github-login {
					background-color: #24292e;
					color: white;
				}
				.google-login {
					background-color: #4285F4;
					color: white;
				}
				.wechat-login {
					background-color: #07C160;
					color: white;
				}
			</style>
		</head>
		<body>
			<h1>社交登录示例</h1>
			<div class="login-buttons">
				{{if .GitHub}}<a href="/auth/github" class="login-button github-login">使用GitHub账号登录</a>{{end}}
				{{if .Google}}<a href="/auth/google" class="login-button google-login">使用Google账号登录</a>{{end}}
				{{if .WeChat}}<a href="/auth/wechat" class="login-button wechat-login">使用微信账号登录</a>{{end}}
			</div>
		</body>
		</html>
		`))

		data := struct {
			GitHub bool
			Google bool
			WeChat bool
		}{
			GitHub: config.GithubClientID != "",
			Google: config.GoogleClientID != "",
			WeChat: config.WeChatAppID != "",
		}

		tmpl.Execute(w, data)
	})

	// GitHub登录路由
	mux.HandleFunc("/auth/github", socialManager.HandleLogin(drivers.ProviderGitHub))
	mux.HandleFunc("/auth/github/callback", func(w http.ResponseWriter, r *http.Request) {
		handleCallback(w, r, socialManager, userRepo, drivers.ProviderGitHub)
	})

	// Google登录路由
	mux.HandleFunc("/auth/google", socialManager.HandleLogin(drivers.ProviderGoogle))
	mux.HandleFunc("/auth/google/callback", func(w http.ResponseWriter, r *http.Request) {
		handleCallback(w, r, socialManager, userRepo, drivers.ProviderGoogle)
	})

	// 微信登录路由
	mux.HandleFunc("/auth/wechat", socialManager.HandleLogin(drivers.ProviderWeChat))
	mux.HandleFunc("/auth/wechat/callback", func(w http.ResponseWriter, r *http.Request) {
		handleCallback(w, r, socialManager, userRepo, drivers.ProviderWeChat)
	})

	// 用户个人资料页面
	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		// 获取查询参数中的用户ID
		userID := r.URL.Query().Get("id")
		if userID == "" {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		// 查找用户（在实际应用中应该从数据库查询）
		var foundUser *User
		for _, user := range userRepo.users {
			if fmt.Sprintf("%d", user.ID) == userID {
				foundUser = user
				break
			}
		}

		if foundUser == nil {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		// 显示用户资料
		tmpl := template.Must(template.New("profile").Parse(`
		<!DOCTYPE html>
		<html>
		<head>
			<title>用户资料</title>
			<style>
				body {
					font-family: Arial, sans-serif;
					max-width: 800px;
					margin: 0 auto;
					padding: 20px;
				}
				h1 {
					color: #333;
				}
				.profile {
					background-color: #f9f9f9;
					border: 1px solid #ddd;
					border-radius: 4px;
					padding: 20px;
					margin-top: 20px;
				}
				.avatar {
					width: 100px;
					height: 100px;
					border-radius: 50%;
				}
				.back-link {
					display: inline-block;
					margin-top: 20px;
					color: #666;
				}
			</style>
		</head>
		<body>
			<h1>用户资料</h1>
			<div class="profile">
				{{if .Avatar}}<img src="{{.Avatar}}" alt="用户头像" class="avatar">{{end}}
				<h2>{{.Name}}</h2>
				<p><strong>ID:</strong> {{.ID}}</p>
				<p><strong>邮箱:</strong> {{.Email}}</p>
				<p><strong>登录提供商:</strong> {{.Provider}}</p>
				<p><strong>社交ID:</strong> {{.SocialID}}</p>
			</div>
			<a href="/" class="back-link">返回登录页面</a>
		</body>
		</html>
		`))

		tmpl.Execute(w, foundUser)
	})

	// 启动服务器
	log.Printf("社交登录示例应用已启动，访问 http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

// 处理回调的辅助函数
func handleCallback(w http.ResponseWriter, r *http.Request, manager *drivers.SocialManager, repo *UserRepository, provider string) {
	// 使用社交登录处理回调
	callbackHandler := manager.HandleCallback(provider)

	// 捕获原始响应
	responseRecorder := newResponseRecorder(w)

	// 调用标准处理器
	callbackHandler(responseRecorder, r)

	// 如果是重定向到/user，修改URL添加用户ID
	if responseRecorder.statusCode == http.StatusFound && responseRecorder.redirectURL == "/user" {
		// 查找最新创建的用户
		users := repo.GetUsers()
		var lastUser *User
		var maxID int64

		for _, user := range users {
			if user.ID > maxID {
				maxID = user.ID
				lastUser = user
			}
		}

		if lastUser != nil {
			// 使用最新用户ID重定向
			http.Redirect(w, r, fmt.Sprintf("/user?id=%d", lastUser.ID), http.StatusFound)
			return
		}
	}

	// 否则复制原始响应
	for key, values := range responseRecorder.header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	if responseRecorder.statusCode != 0 {
		w.WriteHeader(responseRecorder.statusCode)
	}

	w.Write(responseRecorder.body)
}

// responseRecorder 是自定义的响应记录器
type responseRecorder struct {
	header      http.Header
	body        []byte
	statusCode  int
	redirectURL string
	writer      http.ResponseWriter
}

// newResponseRecorder 创建新的响应记录器
func newResponseRecorder(w http.ResponseWriter) *responseRecorder {
	return &responseRecorder{
		header: make(http.Header),
		writer: w,
	}
}

// Header 实现http.ResponseWriter接口
func (r *responseRecorder) Header() http.Header {
	return r.header
}

// Write 实现http.ResponseWriter接口
func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body = append(r.body, b...)
	return len(b), nil
}

// WriteHeader 实现http.ResponseWriter接口
func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode

	// 捕获重定向URL
	if statusCode == http.StatusFound || statusCode == http.StatusMovedPermanently {
		r.redirectURL = r.header.Get("Location")
	}
}
