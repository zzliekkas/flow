package i18n

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/go-playground/validator/v10"
	"github.com/zzliekkas/flow/di"
)

// ProviderConfig 国际化服务提供者配置
type ProviderConfig struct {
	// 默认语言
	DefaultLocale string `mapstructure:"default_locale"`

	// 回退语言
	FallbackLocale string `mapstructure:"fallback_locale"`

	// 翻译文件存放目录
	TranslationsDir string `mapstructure:"translations_dir"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() ProviderConfig {
	return ProviderConfig{
		DefaultLocale:   "en",
		FallbackLocale:  "en",
		TranslationsDir: "./resources/translations",
	}
}

// Provider 国际化服务提供者
type Provider struct {
	// 配置
	config ProviderConfig
}

// NewProvider 创建国际化服务提供者
func NewProvider(config ProviderConfig) *Provider {
	return &Provider{
		config: config,
	}
}

// Register 注册国际化服务
func (p *Provider) Register(container *di.Container) error {
	// 创建翻译管理器
	manager := NewManager(p.config.DefaultLocale, p.config.FallbackLocale)

	// 加载翻译文件
	if p.config.TranslationsDir != "" {
		if _, err := os.Stat(p.config.TranslationsDir); !os.IsNotExist(err) {
			if err := manager.LoadTranslations(p.config.TranslationsDir); err != nil {
				return fmt.Errorf("国际化服务: 加载翻译文件失败: %w", err)
			}
		} else {
			// 如果翻译目录不存在，尝试创建并添加默认翻译
			if err := os.MkdirAll(p.config.TranslationsDir, 0755); err != nil {
				return fmt.Errorf("国际化服务: 创建翻译目录失败: %w", err)
			}

			// 创建基本的英文翻译文件
			if err := p.createDefaultTranslations(p.config.TranslationsDir); err != nil {
				return fmt.Errorf("国际化服务: 创建默认翻译文件失败: %w", err)
			}

			// 加载刚创建的翻译
			if err := manager.LoadTranslations(p.config.TranslationsDir); err != nil {
				return fmt.Errorf("国际化服务: 加载默认翻译文件失败: %w", err)
			}
		}
	}

	// 设置全局翻译器
	SetTranslator(manager)

	// 将翻译器注册到容器
	container.Provide(func() Translator {
		return manager
	})

	// 注册格式化器
	formatter := NewFormatter(manager)
	container.Provide(func() *Formatter {
		return formatter
	})

	// 注册验证翻译器
	// 创建新的验证器实例，而不是尝试从容器获取
	validate := validator.New()
	validationTranslator := NewValidationTranslator(manager, validate)
	container.Provide(func() *ValidationTranslator {
		return validationTranslator
	})

	return nil
}

// Boot 启动国际化服务
func (p *Provider) Boot(container *di.Container) error {
	return nil
}

// 创建默认翻译文件
func (p *Provider) createDefaultTranslations(dir string) error {
	// 英文翻译
	enTranslations := `{
		"messages": {
			"welcome": "Welcome to our application",
			"goodbye": "Goodbye",
			"error": {
				"not_found": "Resource not found",
				"server_error": "An error occurred on the server"
			},
			"validation": {
				"required": "The :field field is required",
				"min": "The :field must be at least :min characters",
				"max": "The :field may not be greater than :max characters",
				"email": "The :field must be a valid email address"
			},
			"auth": {
				"login": "Login",
				"register": "Register",
				"logout": "Logout",
				"forgot_password": "Forgot Password?",
				"reset_password": "Reset Password"
			},
			"pagination": {
				"previous": "Previous",
				"next": "Next",
				"showing": "Showing :from to :to of :total results"
			},
			"count": {
				"one": "You have :count item",
				"other": "You have :count items"
			}
		}
	}`

	// 中文翻译
	zhTranslations := `{
		"messages": {
			"welcome": "欢迎使用我们的应用",
			"goodbye": "再见",
			"error": {
				"not_found": "未找到资源",
				"server_error": "服务器发生错误"
			},
			"validation": {
				"required": ":field 字段是必填的",
				"min": ":field 至少需要 :min 个字符",
				"max": ":field 不能超过 :max 个字符",
				"email": ":field 必须是有效的电子邮件地址"
			},
			"auth": {
				"login": "登录",
				"register": "注册",
				"logout": "退出",
				"forgot_password": "忘记密码？",
				"reset_password": "重置密码"
			},
			"pagination": {
				"previous": "上一页",
				"next": "下一页",
				"showing": "显示第 :from 到 :to 条，共 :total 条"
			},
			"count": {
				"other": "您有 :count 个项目"
			}
		}
	}`

	// 写入英文翻译
	if err := ioutil.WriteFile(filepath.Join(dir, "en.json"), []byte(enTranslations), 0644); err != nil {
		return err
	}

	// 写入中文翻译
	if err := ioutil.WriteFile(filepath.Join(dir, "zh.json"), []byte(zhTranslations), 0644); err != nil {
		return err
	}

	return nil
}
