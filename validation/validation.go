// Package validation 提供数据验证功能和自定义验证规则
package validation

import (
	"reflect"
	"strings"
	"sync"

	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
)

var (
	validate      *validator.Validate     // 全局验证器实例
	trans         ut.Translator           // 全局翻译器实例
	translator    *ut.UniversalTranslator // 全局通用翻译器
	customRules   = make(map[string]Rule) // 自定义规则映射
	registerOnce  sync.Once               // 确保只注册一次
	initValidator sync.Once               // 确保只初始化一次
)

// Rule 定义自定义验证规则
type Rule struct {
	Validation    validator.Func               // 验证函数
	ErrorMessage  string                       // 默认错误消息
	TranslateFunc func(ut ut.Translator) error // 翻译器注册函数
}

// Initialize 初始化验证器并注册自定义规则
func Initialize() {
	initValidator.Do(func() {
		// 创建验证器实例
		validate = validator.New()

		// 使用结构体字段名称而非标签名
		validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return fld.Name
			}
			return name
		})

		// 设置中文翻译器
		zhTrans := zh.New()
		translator = ut.New(zhTrans, zhTrans)
		trans, _ = translator.GetTranslator("zh")
		zh_translations.RegisterDefaultTranslations(validate, trans)

		// 注册自定义规则
		registerCustomRules()
	})
}

// registerCustomRules 注册所有自定义验证规则
func registerCustomRules() {
	registerOnce.Do(func() {
		for tag, rule := range customRules {
			err := validate.RegisterValidation(tag, rule.Validation)
			if err != nil {
				panic("注册验证规则失败: " + err.Error())
			}

			if rule.TranslateFunc != nil {
				err = rule.TranslateFunc(trans)
				if err != nil {
					panic("注册验证翻译失败: " + err.Error())
				}
			} else if rule.ErrorMessage != "" {
				registerDefaultTranslation(tag, rule.ErrorMessage)
			}
		}
	})
}

// RegisterRule 注册自定义验证规则
func RegisterRule(tag string, rule Rule) {
	customRules[tag] = rule

	// 如果验证器已初始化，立即注册规则
	if validate != nil {
		err := validate.RegisterValidation(tag, rule.Validation)
		if err != nil {
			panic("注册验证规则失败: " + err.Error())
		}

		if rule.TranslateFunc != nil {
			err = rule.TranslateFunc(trans)
			if err != nil {
				panic("注册验证翻译失败: " + err.Error())
			}
		} else if rule.ErrorMessage != "" {
			registerDefaultTranslation(tag, rule.ErrorMessage)
		}
	}
}

// registerDefaultTranslation 注册简单翻译
func registerDefaultTranslation(tag string, message string) {
	_ = validate.RegisterTranslation(tag, trans, func(ut ut.Translator) error {
		return ut.Add(tag, message, true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T(tag, fe.Field())
		return t
	})
}

// Validate 执行结构体验证并返回错误信息
func Validate(s interface{}) error {
	if validate == nil {
		Initialize()
	}
	return validate.Struct(s)
}

// ValidateVar 验证单个变量
func ValidateVar(field interface{}, tag string) error {
	if validate == nil {
		Initialize()
	}
	return validate.Var(field, tag)
}

// TranslateError 翻译验证错误
func TranslateError(err error) []string {
	if err == nil {
		return nil
	}

	// 确保翻译器已初始化
	if validate == nil || trans == nil {
		Initialize()
	}

	errs, ok := err.(validator.ValidationErrors)
	if !ok {
		return []string{err.Error()}
	}

	var errMessages []string
	for _, e := range errs {
		translatedErr := e.Translate(trans)
		errMessages = append(errMessages, translatedErr)
	}

	return errMessages
}

// GetValidator 获取验证器实例
func GetValidator() *validator.Validate {
	if validate == nil {
		Initialize()
	}
	return validate
}

// GetTranslator 获取翻译器实例
func GetTranslator() ut.Translator {
	if trans == nil {
		Initialize()
	}
	return trans
}
