package i18n

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// TranslationManager 管理翻译资源的加载和访问
type TranslationManager struct {
	// 翻译数据，格式为：locale -> key -> translation
	translations map[string]map[string]string

	// 复数形式翻译，格式为：locale -> key -> pluralForm -> translation
	pluralTranslations map[string]map[string]map[string]string

	// 当前语言
	defaultLocale string

	// 回退语言，当前语言找不到翻译时使用
	fallbackLocale string

	// 支持的语言列表
	availableLocales []string

	// 确保并发安全
	mu sync.RWMutex
}

// NewManager 创建一个新的翻译管理器
func NewManager(defaultLocale, fallbackLocale string) *TranslationManager {
	return &TranslationManager{
		translations:       make(map[string]map[string]string),
		pluralTranslations: make(map[string]map[string]map[string]string),
		defaultLocale:      defaultLocale,
		fallbackLocale:     fallbackLocale,
		availableLocales:   []string{defaultLocale, fallbackLocale},
		mu:                 sync.RWMutex{},
	}
}

// DefaultManager 创建一个使用英语作为默认和回退语言的管理器
func DefaultManager() *TranslationManager {
	return NewManager("en", "en")
}

// LoadTranslations 从指定目录加载所有翻译文件
func (m *TranslationManager) LoadTranslations(dir string) error {
	entries, err := ioutil.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("i18n: failed to read translations directory: %w", err)
	}

	// 清除现有翻译，以便重新加载
	m.mu.Lock()
	m.translations = make(map[string]map[string]string)
	m.pluralTranslations = make(map[string]map[string]map[string]string)
	m.availableLocales = []string{}
	m.mu.Unlock()

	// 遍历目录中的文件
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := filepath.Ext(entry.Name())
		if ext != ".json" {
			continue // 目前仅支持JSON格式
		}

		// 从文件名中提取locale，例如：en.json -> en
		locale := strings.TrimSuffix(entry.Name(), ext)
		if locale == "" {
			continue
		}

		// 加载文件
		if err := m.LoadFile(filepath.Join(dir, entry.Name()), locale); err != nil {
			return err
		}
	}

	return nil
}

// LoadFile 从文件加载特定语言的翻译
func (m *TranslationManager) LoadFile(file, locale string) error {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return fmt.Errorf("i18n: failed to read translation file %s: %w", file, err)
	}

	var translations map[string]interface{}
	if err := json.Unmarshal(data, &translations); err != nil {
		return fmt.Errorf("i18n: failed to parse translation file %s: %w", file, err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// 初始化该语言的翻译映射
	if _, exists := m.translations[locale]; !exists {
		m.translations[locale] = make(map[string]string)
		m.pluralTranslations[locale] = make(map[string]map[string]string)

		// 添加到可用语言列表
		m.availableLocales = append(m.availableLocales, locale)
	}

	// 解析扁平化的翻译对象
	m.parseTranslations(translations, "", locale)

	return nil
}

// parseTranslations 递归解析翻译对象，支持嵌套键
func (m *TranslationManager) parseTranslations(translations map[string]interface{}, prefix string, locale string) {
	for key, value := range translations {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		switch v := value.(type) {
		case string:
			// 简单字符串翻译
			m.translations[locale][fullKey] = v

		case map[string]interface{}:
			// 嵌套对象
			// 检查是否为复数形式
			if _, hasOne := v["one"]; hasOne {
				// 初始化复数形式映射
				if _, exists := m.pluralTranslations[locale][fullKey]; !exists {
					m.pluralTranslations[locale][fullKey] = make(map[string]string)
				}

				// 存储各种复数形式
				for form, text := range v {
					if textStr, ok := text.(string); ok {
						m.pluralTranslations[locale][fullKey][form] = textStr
					}
				}
			} else {
				// 普通嵌套对象，递归处理
				m.parseTranslations(v, fullKey, locale)
			}

		default:
			// 忽略其他类型
		}
	}
}

// Translate 翻译指定键
func (m *TranslationManager) Translate(ctx context.Context, key string, params map[string]interface{}) string {
	locale := m.GetLocale(ctx)
	return m.TranslateWithLocale(ctx, locale, key, params)
}

// TranslatePlural 翻译支持复数形式的键
func (m *TranslationManager) TranslatePlural(ctx context.Context, key string, count int, params map[string]interface{}) string {
	locale := m.GetLocale(ctx)
	return m.TranslatePluralWithLocale(ctx, locale, key, count, params)
}

// TranslateWithLocale 使用指定语言翻译
func (m *TranslationManager) TranslateWithLocale(ctx context.Context, locale, key string, params map[string]interface{}) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 尝试从指定语言获取翻译
	translations, ok := m.translations[locale]
	if ok {
		if translation, exists := translations[key]; exists {
			return m.interpolateParams(translation, params)
		}
	}

	// 如果指定语言没有找到，尝试回退语言
	if locale != m.fallbackLocale {
		fallbackTranslations, ok := m.translations[m.fallbackLocale]
		if ok {
			if translation, exists := fallbackTranslations[key]; exists {
				return m.interpolateParams(translation, params)
			}
		}
	}

	// 都没找到，返回键名
	return key
}

// TranslatePluralWithLocale 使用指定语言翻译复数形式
func (m *TranslationManager) TranslatePluralWithLocale(ctx context.Context, locale, key string, count int, params map[string]interface{}) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 确保params包含count
	if params == nil {
		params = make(map[string]interface{})
	}
	params["count"] = count

	// 确定复数形式
	pluralForm := getPluralForm(locale, count)

	// 尝试从指定语言获取复数形式翻译
	if plurals, ok := m.pluralTranslations[locale]; ok {
		if forms, exists := plurals[key]; exists {
			if translation, has := forms[pluralForm]; has {
				return m.interpolateParams(translation, params)
			}
		}
	}

	// 尝试回退语言
	if locale != m.fallbackLocale {
		if plurals, ok := m.pluralTranslations[m.fallbackLocale]; ok {
			if forms, exists := plurals[key]; exists {
				pluralForm = getPluralForm(m.fallbackLocale, count) // 使用回退语言的复数规则
				if translation, has := forms[pluralForm]; has {
					return m.interpolateParams(translation, params)
				}
			}
		}
	}

	// 尝试从非复数翻译中查找
	return m.TranslateWithLocale(ctx, locale, key, params)
}

// HasTranslation 检查是否存在指定键的翻译
func (m *TranslationManager) HasTranslation(ctx context.Context, key string) bool {
	locale := m.GetLocale(ctx)

	m.mu.RLock()
	defer m.mu.RUnlock()

	// 检查普通翻译
	if translations, ok := m.translations[locale]; ok {
		if _, exists := translations[key]; exists {
			return true
		}
	}

	// 检查复数形式翻译
	if plurals, ok := m.pluralTranslations[locale]; ok {
		if _, exists := plurals[key]; exists {
			return true
		}
	}

	// 检查回退语言
	if locale != m.fallbackLocale {
		if translations, ok := m.translations[m.fallbackLocale]; ok {
			if _, exists := translations[key]; exists {
				return true
			}
		}

		if plurals, ok := m.pluralTranslations[m.fallbackLocale]; ok {
			if _, exists := plurals[key]; exists {
				return true
			}
		}
	}

	return false
}

// HasLocale 检查是否支持指定的语言
func (m *TranslationManager) HasLocale(ctx context.Context, locale string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, l := range m.availableLocales {
		if l == locale {
			return true
		}
	}

	return false
}

// GetLocale 获取当前上下文的语言
func (m *TranslationManager) GetLocale(ctx context.Context) string {
	// 尝试从上下文中获取
	if locale, ok := ctx.Value(localeKey{}).(string); ok && locale != "" {
		return locale
	}

	return m.defaultLocale
}

// SetLocale 设置当前上下文的语言
func (m *TranslationManager) SetLocale(ctx context.Context, locale string) context.Context {
	return context.WithValue(ctx, localeKey{}, locale)
}

// GetFallbackLocale 获取回退语言
func (m *TranslationManager) GetFallbackLocale() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.fallbackLocale
}

// SetFallbackLocale 设置回退语言
func (m *TranslationManager) SetFallbackLocale(locale string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.fallbackLocale = locale

	// 确保回退语言在可用语言列表中
	exists := false
	for _, l := range m.availableLocales {
		if l == locale {
			exists = true
			break
		}
	}

	if !exists {
		m.availableLocales = append(m.availableLocales, locale)
	}
}

// GetAvailableLocales 获取所有可用的语言
func (m *TranslationManager) GetAvailableLocales() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 返回一个副本而不是原始切片
	result := make([]string, len(m.availableLocales))
	copy(result, m.availableLocales)

	return result
}

// SaveTranslations 将翻译保存到文件
func (m *TranslationManager) SaveTranslations(dir string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 确保目录存在
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("i18n: failed to create translations directory: %w", err)
	}

	// 遍历所有语言
	for _, locale := range m.availableLocales {
		translations := make(map[string]interface{})

		// 添加普通翻译
		if trans, ok := m.translations[locale]; ok {
			for key, value := range trans {
				// 将扁平键转换为嵌套对象
				m.setNestedValue(translations, key, value)
			}
		}

		// 添加复数形式翻译
		if plurals, ok := m.pluralTranslations[locale]; ok {
			for key, forms := range plurals {
				// 将复数形式转换为对象
				pluralObj := make(map[string]interface{})
				for form, text := range forms {
					pluralObj[form] = text
				}
				m.setNestedValue(translations, key, pluralObj)
			}
		}

		// 保存到文件
		data, err := json.MarshalIndent(translations, "", "  ")
		if err != nil {
			return fmt.Errorf("i18n: failed to marshal translations for locale %s: %w", locale, err)
		}

		filename := filepath.Join(dir, locale+".json")
		if err := ioutil.WriteFile(filename, data, 0644); err != nil {
			return fmt.Errorf("i18n: failed to write translations file %s: %w", filename, err)
		}
	}

	return nil
}

// setNestedValue 设置嵌套对象的值
func (m *TranslationManager) setNestedValue(obj map[string]interface{}, key string, value interface{}) {
	parts := strings.Split(key, ".")

	current := obj
	for i, part := range parts {
		if i == len(parts)-1 {
			// 最后一部分，设置值
			current[part] = value
		} else {
			// 中间部分，确保子对象存在
			if _, ok := current[part].(map[string]interface{}); !ok {
				current[part] = make(map[string]interface{})
			}
			current = current[part].(map[string]interface{})
		}
	}
}

// 替换翻译中的参数占位符
func (m *TranslationManager) interpolateParams(translation string, params map[string]interface{}) string {
	if params == nil || len(params) == 0 {
		return translation
	}

	result := translation
	for key, value := range params {
		placeholder := fmt.Sprintf(":%s", key)
		replacement := fmt.Sprintf("%v", value)
		result = strings.Replace(result, placeholder, replacement, -1)
	}

	return result
}

// localeKey 用于上下文中存储当前语言的键
type localeKey struct{}

// getPluralForm 获取指定语言和数量的复数形式
// 复数规则基于 CLDR (Unicode Common Locale Data Repository)
func getPluralForm(locale string, count int) string {
	switch locale {
	case "zh", "ja", "ko", "vi", "th":
		// 这些语言不区分复数形式
		return "other"

	case "en", "de", "nl", "sv", "da", "no", "nb", "nn", "et", "fi", "hu", "it", "es", "pt", "bg", "el", "tr":
		// 一个或其他
		if count == 1 {
			return "one"
		}
		return "other"

	case "fr", "pt_BR":
		// 零到一个，或其他
		if count == 0 || count == 1 {
			return "one"
		}
		return "other"

	case "ru", "uk", "hr", "sr", "bs", "sk", "cs", "pl":
		// 斯拉夫语系的复杂规则
		mod10 := count % 10
		mod100 := count % 100

		if mod10 == 1 && mod100 != 11 {
			return "one"
		}
		if mod10 >= 2 && mod10 <= 4 && (mod100 < 10 || mod100 >= 20) {
			return "few"
		}
		return "many"

	case "ar":
		// 阿拉伯语的复杂规则
		if count == 0 {
			return "zero"
		}
		if count == 1 {
			return "one"
		}
		if count == 2 {
			return "two"
		}
		mod100 := count % 100
		if mod100 >= 3 && mod100 <= 10 {
			return "few"
		}
		if mod100 >= 11 && mod100 <= 99 {
			return "many"
		}
		return "other"

	default:
		// 默认简单区分单复数
		if count == 1 {
			return "one"
		}
		return "other"
	}
}
