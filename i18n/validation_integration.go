package i18n

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
)

// ValidationTranslator 提供验证错误消息的本地化
type ValidationTranslator struct {
	// 通用翻译器
	translator Translator

	// validator实例
	validate *validator.Validate

	// 通用翻译器实例
	universalTranslator *ut.UniversalTranslator

	// 翻译器注册表
	translators map[string]ut.Translator

	// 自定义翻译函数
	customTranslations map[string]map[string]string
}

// NewValidationTranslator 创建一个新的验证翻译器
func NewValidationTranslator(translator Translator, validate *validator.Validate) *ValidationTranslator {
	// 创建通用翻译器
	enLocale := en.New()
	zhLocale := zh.New()
	universalTranslator := ut.New(enLocale, zhLocale)

	vt := &ValidationTranslator{
		translator:          translator,
		validate:            validate,
		universalTranslator: universalTranslator,
		translators:         make(map[string]ut.Translator),
		customTranslations:  make(map[string]map[string]string),
	}

	// 注册默认翻译器
	vt.registerDefaultTranslators()

	return vt
}

// 注册默认翻译器
func (vt *ValidationTranslator) registerDefaultTranslators() {
	// 英语翻译器
	enTranslator, _ := vt.universalTranslator.GetTranslator("en")
	vt.translators["en"] = enTranslator
	en_translations.RegisterDefaultTranslations(vt.validate, enTranslator)

	// 中文翻译器
	zhTranslator, _ := vt.universalTranslator.GetTranslator("zh")
	vt.translators["zh"] = zhTranslator
	zh_translations.RegisterDefaultTranslations(vt.validate, zhTranslator)

	// 注册字段名称翻译
	vt.validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return fld.Name
		}
		return name
	})
}

// RegisterLocale 注册新的语言
func (vt *ValidationTranslator) RegisterLocale(locale string, translator locales.Translator) {
	// 向通用翻译器添加新语言
	vt.universalTranslator.AddTranslator(translator, false)

	// 获取对应的翻译器
	trans, _ := vt.universalTranslator.GetTranslator(locale)
	vt.translators[locale] = trans
}

// RegisterTranslation 注册自定义验证规则的翻译
func (vt *ValidationTranslator) RegisterTranslation(locale, tag, translation string) error {
	if trans, exists := vt.translators[locale]; exists {
		return vt.validate.RegisterTranslation(tag, trans, func(ut ut.Translator) error {
			return ut.Add(tag, translation, false)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T(fe.Tag(), fe.Field())
			return t
		})
	}

	// 如果翻译器不存在，存储在自定义翻译映射中
	if _, exists := vt.customTranslations[locale]; !exists {
		vt.customTranslations[locale] = make(map[string]string)
	}
	vt.customTranslations[locale][tag] = translation

	return nil
}

// TranslateError 翻译验证错误为指定语言
func (vt *ValidationTranslator) TranslateError(ctx context.Context, err error) string {
	locale := vt.translator.GetLocale(ctx)
	return vt.TranslateErrorWithLocale(locale, err)
}

// TranslateErrorWithLocale 使用指定语言翻译验证错误
func (vt *ValidationTranslator) TranslateErrorWithLocale(locale string, err error) string {
	if err == nil {
		return ""
	}

	// 类型转换为验证错误
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		// 不是验证错误，直接返回原始错误消息
		return err.Error()
	}

	// 尝试获取对应语言的翻译器
	trans, exists := vt.translators[locale]
	if !exists {
		// 尝试使用回退语言
		fallbackLocale := vt.translator.GetFallbackLocale()
		trans, exists = vt.translators[fallbackLocale]
		if !exists {
			// 使用英语作为最后回退
			trans = vt.translators["en"]
		}
	}

	// 翻译每个错误
	var errMessages []string
	for _, e := range validationErrors {
		// 尝试通过通用翻译器翻译
		message := e.Translate(trans)

		// 如果翻译失败或返回空字符串，尝试使用自定义翻译
		if message == "" {
			if customTrans, ok := vt.customTranslations[locale]; ok {
				if customMessage, ok := customTrans[e.Tag()]; ok {
					// 替换字段名称
					message = strings.Replace(customMessage, ":field", e.Field(), -1)
				}
			}
		}

		// 如果仍然没有翻译，使用默认格式
		if message == "" {
			message = fmt.Sprintf("字段 '%s' 验证规则 '%s' 失败", e.Field(), e.Tag())
		}

		errMessages = append(errMessages, message)
	}

	// 合并所有错误消息
	return strings.Join(errMessages, "; ")
}

// Validate 验证并翻译错误
func (vt *ValidationTranslator) Validate(ctx context.Context, s interface{}) error {
	if err := vt.validate.Struct(s); err != nil {
		// 如果发生验证错误，返回翻译后的错误
		if _, ok := err.(validator.ValidationErrors); ok {
			// 返回原始错误，稍后可以通过TranslateError翻译
			return err
		}
		return err
	}
	return nil
}

// SetValidateTagNameFunc 设置字段名称获取函数
func (vt *ValidationTranslator) SetValidateTagNameFunc(fn validator.TagNameFunc) {
	vt.validate.RegisterTagNameFunc(fn)
}

// GetValidator 获取验证器实例
func (vt *ValidationTranslator) GetValidator() *validator.Validate {
	return vt.validate
}

// GetTranslator 获取指定语言的翻译器
func (vt *ValidationTranslator) GetTranslator(locale string) (ut.Translator, bool) {
	trans, exists := vt.translators[locale]
	return trans, exists
}
