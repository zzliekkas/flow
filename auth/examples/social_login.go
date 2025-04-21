package examples

import (
	"context"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/zzliekkas/flow"
	"github.com/zzliekkas/flow/auth"
	"github.com/zzliekkas/flow/auth/drivers"
)

// User 是一个用户模型示例
type User struct {
	ID          string
	Username    string
	Email       string
	Avatar      string
	Password    string
	Provider    string
	ProviderID  string
	Roles       []string
	Permissions []string
	IsActive    bool
}

// GetAuthIdentifier 实现 Authenticatable 接口
func (u *User) GetAuthIdentifier() string {
	return u.ID
}

// GetAuthUsername 实现 Authenticatable 接口
func (u *User) GetAuthUsername() string {
	return u.Username
}

// GetPermissions 实现 Authenticatable 接口
func (u *User) GetPermissions() []string {
	return u.Permissions
}

// GetRoles 实现 Authenticatable 接口
func (u *User) GetRoles() []string {
	return u.Roles
}

// SocialUserRepository 是社交用户存储库示例
type SocialUserRepository struct {
	users map[string]*User
}

// NewSocialUserRepository 创建新的社交用户存储库
func NewSocialUserRepository() *SocialUserRepository {
	return &SocialUserRepository{
		users: make(map[string]*User),
	}
}

// FindUserBySocialID 通过社交ID查找用户
func (r *SocialUserRepository) FindUserBySocialID(ctx context.Context, provider, socialID string) (interface{}, error) {
	for _, user := range r.users {
		if user.Provider == provider && user.ProviderID == socialID {
			return user, nil
		}
	}
	return nil, nil
}

// CreateUser 创建新用户
func (r *SocialUserRepository) CreateUser(ctx context.Context, user interface{}) error {
	u, ok := user.(*User)
	if !ok {
		return auth.ErrUserNotFound
	}
	r.users[u.ID] = u
	return nil
}

// SaveUser 保存用户
func (r *SocialUserRepository) SaveUser(user *User) {
	r.users[user.ID] = user
}

// UserRepository 是用户存储库示例
type UserRepository struct {
	users map[string]*User
}

// NewUserRepository 创建新的用户存储库
func NewUserRepository() *UserRepository {
	return &UserRepository{
		users: make(map[string]*User),
	}
}

// FindByID 通过ID查找用户
func (r *UserRepository) FindByID(ctx context.Context, id string) (auth.Authenticatable, error) {
	user, exists := r.users[id]
	if !exists {
		return nil, auth.ErrUserNotFound
	}
	return user, nil
}

// FindByCredentials 通过凭证查找用户
func (r *UserRepository) FindByCredentials(ctx context.Context, credentials map[string]string) (auth.Authenticatable, error) {
	// 通过电子邮件查找
	if email, ok := credentials["email"]; ok {
		for _, user := range r.users {
			if user.Email == email {
				return user, nil
			}
		}
	}

	// 通过社交ID查找
	if providerID, ok := credentials["provider_id"]; ok {
		for _, user := range r.users {
			if user.Provider+"_"+user.ProviderID == providerID {
				return user, nil
			}
		}
	}

	return nil, auth.ErrUserNotFound
}

// ValidateCredentials 验证用户凭证
func (r *UserRepository) ValidateCredentials(ctx context.Context, user auth.Authenticatable, credentials map[string]string) (bool, error) {
	// 在实际应用中，这里应该验证密码
	return true, nil
}

// SaveUser 保存用户
func (r *UserRepository) SaveUser(user *User) {
	r.users[user.ID] = user
}

// SetupSocialLogin 设置社交登录
func SetupSocialLogin(app *flow.Engine) {
	// 创建用户存储库
	userRepo := NewUserRepository()

	// 创建社交用户存储库
	socialRepo := NewSocialUserRepository()

	// 创建社交登录管理器
	socialManager := drivers.NewSocialManager(socialRepo)

	// 配置GitHub提供商
	githubConfig := map[string]interface{}{
		"client_id":     "your-github-client-id",
		"client_secret": "your-github-client-secret",
		"redirect_url":  "http://localhost:8080/auth/github/callback",
		"scopes":        []string{"user", "user:email"},
	}
	githubProvider := drivers.NewGitHubProvider(githubConfig)
	socialManager.RegisterProvider(githubProvider)

	// 配置Google提供商
	googleConfig := map[string]interface{}{
		"client_id":     "your-google-client-id",
		"client_secret": "your-google-client-secret",
		"redirect_url":  "http://localhost:8080/auth/google/callback",
		"scopes":        []string{"profile", "email"},
	}
	googleProvider := drivers.NewGoogleProvider(googleConfig)
	socialManager.RegisterProvider(googleProvider)

	// 配置微信提供商
	wechatConfig := map[string]interface{}{
		"client_id":     "your-wechat-appid",
		"client_secret": "your-wechat-secret",
		"redirect_url":  "http://localhost:8080/auth/wechat/callback",
		"scopes":        []string{"snsapi_login"},
	}
	wechatProvider := drivers.NewWeChatProvider(wechatConfig)
	socialManager.RegisterProvider(wechatProvider)

	// 设置用户创建回调
	socialManager.SetCreateUserCallback(func(ctx context.Context, socialUser *drivers.SocialUser) (interface{}, error) {
		log.Printf("创建用户：%s，来自：%s", socialUser.Name, socialUser.Provider)

		// 创建新用户
		newUser := &User{
			ID:          generateUniqueID(),
			Username:    socialUser.Name,
			Email:       socialUser.Email,
			Avatar:      socialUser.Avatar,
			Provider:    socialUser.Provider,
			ProviderID:  socialUser.ID,
			Roles:       []string{"user"},
			Permissions: []string{"profile:read"},
			IsActive:    true,
		}

		// 保存用户
		userRepo.SaveUser(newUser)

		return newUser, nil
	})

	// 注册路由
	registerSocialRoutes(app, socialManager)
}

// registerSocialRoutes 注册社交登录路由
func registerSocialRoutes(app *flow.Engine, manager *drivers.SocialManager) {
	// 创建适配器，将http.HandlerFunc转换为flow.HandlerFunc
	adapter := func(handler http.HandlerFunc) flow.HandlerFunc {
		return func(c *flow.Context) {
			handler(c.Writer, c.Request)
		}
	}

	// GitHub登录路由
	app.GET("/auth/github", adapter(manager.HandleLogin(drivers.ProviderGitHub)))
	app.GET("/auth/github/callback", adapter(manager.HandleCallback(drivers.ProviderGitHub)))

	// Google登录路由
	app.GET("/auth/google", adapter(manager.HandleLogin(drivers.ProviderGoogle)))
	app.GET("/auth/google/callback", adapter(manager.HandleCallback(drivers.ProviderGoogle)))

	// 微信登录路由
	app.GET("/auth/wechat", adapter(manager.HandleLogin(drivers.ProviderWeChat)))
	app.GET("/auth/wechat/callback", adapter(manager.HandleCallback(drivers.ProviderWeChat)))

	// 登录页面路由
	app.GET("/login", func(c *flow.Context) {
		html := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>社交登录演示</title>
			<style>
				body {
					font-family: Arial, sans-serif;
					max-width: 500px;
					margin: 0 auto;
					padding: 20px;
				}
				.social-buttons {
					display: flex;
					flex-direction: column;
					gap: 10px;
					margin-top: 20px;
				}
				.social-button {
					display: flex;
					align-items: center;
					padding: 10px 15px;
					border-radius: 4px;
					text-decoration: none;
					color: white;
					font-weight: bold;
				}
				.github {
					background-color: #24292e;
				}
				.google {
					background-color: #4285f4;
				}
				.wechat {
					background-color: #09b83e;
				}
			</style>
		</head>
		<body>
			<h1>选择登录方式</h1>
			<div class="social-buttons">
				<a href="/auth/github" class="social-button github">使用GitHub登录</a>
				<a href="/auth/google" class="social-button google">使用Google登录</a>
				<a href="/auth/wechat" class="social-button wechat">使用微信登录</a>
			</div>
		</body>
		</html>
		`
		c.HTML(200, html, nil) // 添加第三个参数
	})

	// 受保护的个人资料页面
	app.GET("/profile", func(c *flow.Context) {
		// 获取认证用户
		user, exists := auth.UserFromContext(c.Request.Context())
		if !exists {
			c.Redirect(302, "/login")
			return
		}

		// 显示用户信息
		c.String(200, "欢迎回来，%s！", user.GetAuthUsername())
	})
}

// GenerateRandomString 生成指定长度的随机字符串
func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano())

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// generateUniqueID 生成唯一ID
func generateUniqueID() string {
	// 实际应用中应使用UUID或其他方式生成唯一ID
	return "user_" + GenerateRandomString(10)
}
