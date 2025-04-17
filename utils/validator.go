package utils

import (
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/zzliekkas/flow"
)

// Validator 是验证器结构体
type Validator struct {
	validator *validator.Validate
}

// NewValidator 创建一个新的验证器
func NewValidator() *Validator {
	v := validator.New()

	// 注册自定义标签处理函数
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &Validator{
		validator: v,
	}
}

// Validate 验证对象
func (v *Validator) Validate(obj interface{}) error {
	if err := v.validator.Struct(obj); err != nil {
		return err
	}
	return nil
}

// HandleValidationErrors 处理验证错误
func HandleValidationErrors(err error, c *flow.Context) {
	if err == nil {
		return
	}

	var details []map[string]interface{}

	// 检查是否为验证错误
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			details = append(details, map[string]interface{}{
				"field":   e.Field(),
				"tag":     e.Tag(),
				"param":   e.Param(),
				"message": getErrorMessage(e),
			})
		}

		// 返回验证错误响应
		c.JSON(http.StatusBadRequest, flow.H{
			"error":   "验证错误",
			"details": details,
		})
		return
	}

	// 其他错误类型
	c.JSON(http.StatusBadRequest, flow.H{
		"error": err.Error(),
	})
}

// getErrorMessage 获取错误信息
func getErrorMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "此字段是必需的"
	case "email":
		return "必须是有效的电子邮件地址"
	case "min":
		if e.Type().Kind() == reflect.String {
			return "长度必须至少为 " + e.Param() + " 个字符"
		}
		return "必须大于或等于 " + e.Param()
	case "max":
		if e.Type().Kind() == reflect.String {
			return "长度必须小于或等于 " + e.Param() + " 个字符"
		}
		return "必须小于或等于 " + e.Param()
	case "oneof":
		return "必须是以下值之一: " + e.Param()
	default:
		return "验证失败于 '" + e.Tag() + "' 标签"
	}
}
