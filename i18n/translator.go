package i18n

import (
	"context"
	"errors"
)

var (
	// ErrNotFound 表示翻译键未找到
	ErrNotFound = errors.New("i18n: translation key not found")

	// ErrInvalidLocale 表示无效的语言环境
	ErrInvalidLocale = errors.New("i18n: invalid locale")

	// ErrInvalidPluralCount 表示无效的复数计数
	ErrInvalidPluralCount = errors.New("i18n: invalid plural count")
)

// Translator 定义了翻译功能的接口
type Translator interface {
	// Translate 翻译指定的键到当前语言
	Translate(ctx context.Context, key string, params map[string]interface{}) string

	// TranslatePlural 翻译支持复数形式的键
	TranslatePlural(ctx context.Context, key string, count int, params map[string]interface{}) string

	// TranslateWithLocale 翻译到指定的语言
	TranslateWithLocale(ctx context.Context, locale, key string, params map[string]interface{}) string

	// TranslatePluralWithLocale 翻译支持复数形式的键到指定的语言
	TranslatePluralWithLocale(ctx context.Context, locale, key string, count int, params map[string]interface{}) string

	// HasTranslation 检查是否存在指定键的翻译
	HasTranslation(ctx context.Context, key string) bool

	// HasLocale 检查是否支持指定的语言
	HasLocale(ctx context.Context, locale string) bool

	// GetLocale 获取当前上下文的语言
	GetLocale(ctx context.Context) string

	// SetLocale 设置当前上下文的语言
	SetLocale(ctx context.Context, locale string) context.Context

	// GetFallbackLocale 获取回退语言
	GetFallbackLocale() string

	// SetFallbackLocale 设置回退语言
	SetFallbackLocale(locale string)

	// GetAvailableLocales 获取所有可用的语言
	GetAvailableLocales() []string
}

// 简化翻译函数的别名
// 使用方式: T("welcome") 或 T("hello", map[string]interface{}{"name": "John"})
func T(key string, params ...map[string]interface{}) string {
	var p map[string]interface{}
	if len(params) > 0 {
		p = params[0]
	}

	// 使用默认翻译器
	translator := GetTranslator()
	return translator.Translate(context.Background(), key, p)
}

// 处理复数形式的翻译函数别名
// 使用方式: TP("messages.count", 5)
func TP(key string, count int, params ...map[string]interface{}) string {
	var p map[string]interface{}
	if len(params) > 0 {
		p = params[0]
	}

	// 使用默认翻译器
	translator := GetTranslator()
	return translator.TranslatePlural(context.Background(), key, count, p)
}

// 全局翻译器实例
var globalTranslator Translator

// SetTranslator 设置全局翻译器实例
func SetTranslator(translator Translator) {
	globalTranslator = translator
}

// GetTranslator 获取全局翻译器实例
func GetTranslator() Translator {
	if globalTranslator == nil {
		// 如果未设置，返回一个空实现
		return &nullTranslator{}
	}
	return globalTranslator
}

// nullTranslator 空翻译器实现，用作默认回退
type nullTranslator struct{}

func (n *nullTranslator) Translate(ctx context.Context, key string, params map[string]interface{}) string {
	return key
}

func (n *nullTranslator) TranslatePlural(ctx context.Context, key string, count int, params map[string]interface{}) string {
	return key
}

func (n *nullTranslator) TranslateWithLocale(ctx context.Context, locale, key string, params map[string]interface{}) string {
	return key
}

func (n *nullTranslator) TranslatePluralWithLocale(ctx context.Context, locale, key string, count int, params map[string]interface{}) string {
	return key
}

func (n *nullTranslator) HasTranslation(ctx context.Context, key string) bool {
	return false
}

func (n *nullTranslator) HasLocale(ctx context.Context, locale string) bool {
	return false
}

func (n *nullTranslator) GetLocale(ctx context.Context) string {
	return "en"
}

func (n *nullTranslator) SetLocale(ctx context.Context, locale string) context.Context {
	return ctx
}

func (n *nullTranslator) GetFallbackLocale() string {
	return "en"
}

func (n *nullTranslator) SetFallbackLocale(locale string) {
	// 空实现
}

func (n *nullTranslator) GetAvailableLocales() []string {
	return []string{"en"}
}
