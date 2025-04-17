// Package validation 提供数据验证功能和自定义验证规则
package validation

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// InitializeDomainValidation 初始化领域验证器
func InitializeDomainValidation() {
	if validate == nil {
		Initialize() // 使用主验证模块的Initialize函数
	}
}

// DomainValidator 领域模型验证器接口
type DomainValidator interface {
	Validate(domain interface{}) error
}

// DomainValidatorImpl 领域验证器实现
type DomainValidatorImpl struct{}

// NewDomainValidator 创建领域验证器
func NewDomainValidator() DomainValidator {
	return &DomainValidatorImpl{}
}

// Validate 验证领域模型
func (v *DomainValidatorImpl) Validate(domain interface{}) error {
	// 确保验证器已初始化
	if validate == nil {
		Initialize()
	}

	// 设置自定义标签名称
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	// 执行验证
	err := validate.Struct(domain)
	if err == nil {
		return nil
	}

	// 处理验证错误
	validationErrors := err.(validator.ValidationErrors)
	fieldErrors := make([]FieldError, 0)

	for _, e := range validationErrors {
		fieldErrors = append(fieldErrors, FieldError{
			Field:   e.Field(),
			Message: e.Error(),
			Tag:     e.Tag(),
			Value:   e.Value(),
		})
	}

	return ValidationError{
		Errors: fieldErrors,
	}
}

// ValidationError 验证错误
type ValidationError struct {
	Errors []FieldError
}

// Error 实现error接口
func (e ValidationError) Error() string {
	var errStr strings.Builder
	for i, fe := range e.Errors {
		if i > 0 {
			errStr.WriteString("; ")
		}
		errStr.WriteString(fe.Error())
	}
	return errStr.String()
}

// FieldError 字段错误
type FieldError struct {
	Field   string
	Message string
	Tag     string
	Value   interface{}
}

// Error 返回错误信息
func (e FieldError) Error() string {
	return "Field '" + e.Field + "' validation failed: " + e.Message
}

// StructValidator 结构体验证器
type StructValidator struct {
	// 需要验证的领域模型
	Model interface{}
	// 验证标签名称，默认为"validate"
	TagName string
	// 自定义错误消息映射 map[字段名]错误消息
	ErrorMessages map[string]string
}

// NewStructValidator 创建结构体验证器
func NewStructValidator(model interface{}) *StructValidator {
	return &StructValidator{
		Model:         model,
		TagName:       "validate",
		ErrorMessages: make(map[string]string),
	}
}

// WithTagName 设置验证标签名称
func (v *StructValidator) WithTagName(tagName string) *StructValidator {
	v.TagName = tagName
	return v
}

// WithErrorMessage 添加自定义错误消息
func (v *StructValidator) WithErrorMessage(field, message string) *StructValidator {
	v.ErrorMessages[field] = message
	return v
}

// WithErrorMessages 批量添加自定义错误消息
func (v *StructValidator) WithErrorMessages(messages map[string]string) *StructValidator {
	for field, message := range messages {
		v.ErrorMessages[field] = message
	}
	return v
}

// Validate 执行验证
func (v *StructValidator) Validate() error {
	// 确保验证器已初始化
	if validate == nil {
		Initialize()
	}

	// 如果需要使用自定义标签名称
	if v.TagName != "validate" {
		// 保存原始标签名称
		originalTagName := "validate" // 默认标签名
		// 设置新的标签名称
		validate.SetTagName(v.TagName)
		// 确保在函数返回时恢复原始标签名称
		defer validate.SetTagName(originalTagName)
	}

	err := validate.Struct(v.Model)
	if err == nil {
		return nil
	}

	// 处理验证错误
	errs, ok := err.(validator.ValidationErrors)
	if !ok {
		return err
	}

	// 应用自定义错误消息
	fieldErrors := make(validator.ValidationErrors, 0, len(errs))
	if len(v.ErrorMessages) > 0 {
		for _, e := range errs {
			fieldName := e.Field()
			// 查找字段的JSON标签
			if t := reflect.TypeOf(v.Model); t.Kind() == reflect.Ptr {
				if field, ok := t.Elem().FieldByName(fieldName); ok {
					jsonTag := field.Tag.Get("json")
					if jsonTag != "" {
						fieldName = strings.SplitN(jsonTag, ",", 2)[0]
					}
				}
			}

			// 检查是否有自定义错误消息
			if msg, ok := v.ErrorMessages[fieldName]; ok {
				fieldErrors = append(fieldErrors, NewValidationErrorWithCustomMessage(e, msg))
			} else {
				fieldErrors = append(fieldErrors, e)
			}
		}
		return fieldErrors
	}

	return errs
}

// ValidationErrorWithCustomMessage 带有自定义消息的验证错误
type ValidationErrorWithCustomMessage struct {
	validator.FieldError
	customMsg string
}

// NewValidationErrorWithCustomMessage 创建带有自定义消息的验证错误
func NewValidationErrorWithCustomMessage(err validator.FieldError, message string) ValidationErrorWithCustomMessage {
	return ValidationErrorWithCustomMessage{
		FieldError: err,
		customMsg:  message,
	}
}

// Error 返回错误消息
func (e ValidationErrorWithCustomMessage) Error() string {
	if e.customMsg != "" {
		return e.customMsg
	}
	return e.FieldError.Error()
}

// DomainValidatorRegistry 领域验证器注册表
type DomainValidatorRegistry struct {
	validators map[string]func() DomainValidator
}

// NewDomainValidatorRegistry 创建领域验证器注册表
func NewDomainValidatorRegistry() *DomainValidatorRegistry {
	return &DomainValidatorRegistry{
		validators: make(map[string]func() DomainValidator),
	}
}

// Register 注册领域验证器
func (r *DomainValidatorRegistry) Register(name string, factory func() DomainValidator) {
	r.validators[name] = factory
}

// Get 获取领域验证器
func (r *DomainValidatorRegistry) Get(name string) (DomainValidator, bool) {
	factory, ok := r.validators[name]
	if !ok {
		return nil, false
	}
	return factory(), true
}
