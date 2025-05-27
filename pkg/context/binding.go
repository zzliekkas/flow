package context

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Rules 定义验证规则
type Rules map[string][]string

// Bind 绑定请求数据到结构体
func (c *Context) Bind(obj interface{}) error {
	if obj == nil {
		return fmt.Errorf("binding object cannot be nil")
	}

	contentType := c.Request.Header.Get("Content-Type")

	switch {
	case strings.Contains(contentType, "application/json"):
		return c.bindJSON(obj)
	case strings.Contains(contentType, "application/x-www-form-urlencoded"):
		return c.bindForm(obj)
	default:
		return c.bindForm(obj)
	}
}

// bindJSON 绑定JSON数据
func (c *Context) bindJSON(obj interface{}) error {
	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(obj); err != nil {
		return err
	}
	return nil
}

// bindForm 绑定表单数据
func (c *Context) bindForm(obj interface{}) error {
	if err := c.Request.ParseForm(); err != nil {
		return err
	}

	val := reflect.ValueOf(obj)
	if val.Kind() != reflect.Ptr {
		return fmt.Errorf("binding object must be a pointer")
	}

	val = val.Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		typeField := typ.Field(i)

		// 获取form标签
		formKey := typeField.Tag.Get("form")
		if formKey == "" {
			formKey = strings.ToLower(typeField.Name)
		}

		// 获取表单值
		formValue := c.Request.Form.Get(formKey)
		if formValue == "" {
			continue
		}

		// 设置字段值
		switch field.Kind() {
		case reflect.String:
			field.SetString(formValue)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			// 处理整数类型
			if intVal, err := parseInt(formValue, field.Type().Bits()); err == nil {
				field.SetInt(intVal)
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			// 处理无符号整数类型
			if uintVal, err := parseUint(formValue, field.Type().Bits()); err == nil {
				field.SetUint(uintVal)
			}
		case reflect.Float32, reflect.Float64:
			// 处理浮点数类型
			if floatVal, err := parseFloat(formValue, field.Type().Bits()); err == nil {
				field.SetFloat(floatVal)
			}
		case reflect.Bool:
			// 处理布尔类型
			if boolVal, err := parseBool(formValue); err == nil {
				field.SetBool(boolVal)
			}
		}
	}

	return nil
}

// Valid 验证结构体
func (c *Context) Valid(obj interface{}) error {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		typeField := typ.Field(i)

		// 获取validate标签
		validateTag := typeField.Tag.Get("validate")
		if validateTag == "" {
			continue
		}

		// 解析验证规则
		rules := parseValidateTag(validateTag)

		// 验证字段
		if err := validateField(field.Interface(), rules); err != nil {
			return fmt.Errorf("field %s validation failed: %v", typeField.Name, err)
		}
	}

	return nil
}

// ValidField 验证单个字段
func (c *Context) ValidField(rules Rules) error {
	for field, fieldRules := range rules {
		value := c.FindString(field)
		if err := validateValue(value, fieldRules); err != nil {
			return fmt.Errorf("field %s validation failed: %v", field, err)
		}
	}
	return nil
}

// 辅助函数用于类型转换
func parseInt(val string, bitSize int) (int64, error) {
	return strconv.ParseInt(val, 10, bitSize)
}

func parseUint(val string, bitSize int) (uint64, error) {
	return strconv.ParseUint(val, 10, bitSize)
}

func parseFloat(val string, bitSize int) (float64, error) {
	return strconv.ParseFloat(val, bitSize)
}

func parseBool(val string) (bool, error) {
	return strconv.ParseBool(val)
}

// 解析validate标签
func parseValidateTag(tag string) []string {
	return strings.Split(tag, ",")
}

// 验证字段值
func validateField(value interface{}, rules []string) error {
	for _, rule := range rules {
		if err := validateRule(value, rule); err != nil {
			return err
		}
	}
	return nil
}

// 验证单个值
func validateValue(value string, rules []string) error {
	for _, rule := range rules {
		if err := validateRule(value, rule); err != nil {
			return err
		}
	}
	return nil
}

// 验证规则
func validateRule(value interface{}, rule string) error {
	switch {
	case rule == "required":
		return validateRequired(value)
	case strings.HasPrefix(rule, "min="):
		min := strings.TrimPrefix(rule, "min=")
		return validateMin(value, min)
	case strings.HasPrefix(rule, "max="):
		max := strings.TrimPrefix(rule, "max=")
		return validateMax(value, max)
	case rule == "email":
		return validateEmail(value)
	case strings.HasPrefix(rule, "len="):
		length := strings.TrimPrefix(rule, "len=")
		return validateLength(value, length)
	}
	return nil
}

// 验证必填
func validateRequired(value interface{}) error {
	v := reflect.ValueOf(value)

	switch v.Kind() {
	case reflect.String:
		if v.String() == "" {
			return fmt.Errorf("field is required")
		}
	case reflect.Slice, reflect.Map, reflect.Array:
		if v.Len() == 0 {
			return fmt.Errorf("field is required")
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v.Int() == 0 {
			return fmt.Errorf("field is required")
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if v.Uint() == 0 {
			return fmt.Errorf("field is required")
		}
	case reflect.Float32, reflect.Float64:
		if v.Float() == 0 {
			return fmt.Errorf("field is required")
		}
	case reflect.Bool:
		if !v.Bool() {
			return fmt.Errorf("field is required")
		}
	case reflect.Ptr, reflect.Interface:
		if v.IsNil() {
			return fmt.Errorf("field is required")
		}
	}
	return nil
}

// 验证最小值
func validateMin(value interface{}, min string) error {
	minVal, err := strconv.ParseFloat(min, 64)
	if err != nil {
		return fmt.Errorf("invalid min value")
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		if float64(len(v.String())) < minVal {
			return fmt.Errorf("length must be greater than %v", min)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if float64(v.Int()) < minVal {
			return fmt.Errorf("value must be greater than %v", min)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if float64(v.Uint()) < minVal {
			return fmt.Errorf("value must be greater than %v", min)
		}
	case reflect.Float32, reflect.Float64:
		if v.Float() < minVal {
			return fmt.Errorf("value must be greater than %v", min)
		}
	}
	return nil
}

// 验证最大值
func validateMax(value interface{}, max string) error {
	maxVal, err := strconv.ParseFloat(max, 64)
	if err != nil {
		return fmt.Errorf("invalid max value")
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		if float64(len(v.String())) > maxVal {
			return fmt.Errorf("length must be less than %v", max)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if float64(v.Int()) > maxVal {
			return fmt.Errorf("value must be less than %v", max)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if float64(v.Uint()) > maxVal {
			return fmt.Errorf("value must be less than %v", max)
		}
	case reflect.Float32, reflect.Float64:
		if v.Float() > maxVal {
			return fmt.Errorf("value must be less than %v", max)
		}
	}
	return nil
}

// 验证邮箱格式
func validateEmail(value interface{}) error {
	v := reflect.ValueOf(value)
	if v.Kind() != reflect.String {
		return fmt.Errorf("email must be string")
	}

	email := v.String()
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

// 验证长度
func validateLength(value interface{}, length string) error {
	reqLen, err := strconv.Atoi(length)
	if err != nil {
		return fmt.Errorf("invalid length value")
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		if len(v.String()) != reqLen {
			return fmt.Errorf("length must be %v", length)
		}
	case reflect.Slice, reflect.Map, reflect.Array:
		if v.Len() != reqLen {
			return fmt.Errorf("length must be %v", length)
		}
	}
	return nil
}
