// Package validation 提供数据验证功能和自定义验证规则
package validation

import (
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

// 预定义正则表达式
var (
	// 中国大陆手机号码正则表达式
	mobileRegex = regexp.MustCompile(`^1[3-9]\d{9}$`)

	// 中国身份证号码正则表达式（支持15位和18位）
	idCardRegex = regexp.MustCompile(`(^\d{15}$)|(^\d{17}(\d|X|x)$)`)

	// 统一社会信用代码正则表达式
	creditCodeRegex = regexp.MustCompile(`^[0-9A-HJ-NPQRTUWXY]{2}\d{6}[0-9A-HJ-NPQRTUWXY]{10}$`)

	// IPv4正则表达式
	ipv4Regex = regexp.MustCompile(`^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`)

	// 日期格式正则表达式 (YYYY-MM-DD)
	dateRegex = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

	// 中文字符正则表达式
	chineseRegex = regexp.MustCompile(`^[\p{Han}]+$`)
)

// InitCustomRules 初始化所有自定义验证规则
func InitCustomRules() {
	// 注册中国大陆手机号验证规则
	RegisterRule("mobile", Rule{
		Validation:   validateMobile,
		ErrorMessage: "{0}必须是有效的中国大陆手机号",
	})

	// 注册身份证号验证规则
	RegisterRule("idcard", Rule{
		Validation:   validateIDCard,
		ErrorMessage: "{0}必须是有效的身份证号码",
	})

	// 注册统一社会信用代码验证规则
	RegisterRule("creditcode", Rule{
		Validation:   validateCreditCode,
		ErrorMessage: "{0}必须是有效的统一社会信用代码",
	})

	// 注册中文字符验证规则
	RegisterRule("chinese", Rule{
		Validation:   validateChinese,
		ErrorMessage: "{0}必须只包含中文字符",
	})

	// 注册密码强度验证规则
	RegisterRule("password", Rule{
		Validation:   validatePassword,
		ErrorMessage: "{0}必须包含大小写字母、数字和特殊字符，长度至少为8位",
	})

	// 注册日期字符串验证规则
	RegisterRule("datestr", Rule{
		Validation:   validateDateStr,
		ErrorMessage: "{0}必须是有效的日期字符串(YYYY-MM-DD)",
	})

	// 注册字符串不包含特殊字符验证规则
	RegisterRule("nospecial", Rule{
		Validation:   validateNoSpecialChars,
		ErrorMessage: "{0}不能包含特殊字符",
	})

	// 注册整数范围验证规则 (参数化)
	RegisterRule("intrange", Rule{
		Validation: validateIntRange,
		TranslateFunc: func(ut ut.Translator) error {
			return ut.Add("intrange", "{0}必须在{1}和{2}之间", true)
		},
	})
}

// validateMobile 验证中国大陆手机号
func validateMobile(fl validator.FieldLevel) bool {
	if fl.Field().String() == "" {
		return true // 允许空值，使用required标签来要求非空
	}
	return mobileRegex.MatchString(fl.Field().String())
}

// validateIDCard 验证中国身份证号
func validateIDCard(fl validator.FieldLevel) bool {
	if fl.Field().String() == "" {
		return true
	}

	idcard := fl.Field().String()

	// 基本格式检查
	if !idCardRegex.MatchString(idcard) {
		return false
	}

	// 对于18位身份证，验证最后一位校验码
	if len(idcard) == 18 {
		// 加权因子
		weight := []int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}
		// 校验码对应值
		vals := []byte{'1', '0', 'X', '9', '8', '7', '6', '5', '4', '3', '2'}

		sum := 0
		for i := 0; i < 17; i++ {
			n, _ := strconv.Atoi(string(idcard[i]))
			sum += n * weight[i]
		}

		checksum := vals[sum%11]
		return strings.ToUpper(string(idcard[17])) == strings.ToUpper(string(checksum))
	}

	return true
}

// validateCreditCode 验证统一社会信用代码
func validateCreditCode(fl validator.FieldLevel) bool {
	if fl.Field().String() == "" {
		return true
	}
	return creditCodeRegex.MatchString(fl.Field().String())
}

// validateChinese 验证字符串只包含中文字符
func validateChinese(fl validator.FieldLevel) bool {
	if fl.Field().String() == "" {
		return true
	}
	return chineseRegex.MatchString(fl.Field().String())
}

// validatePassword 验证密码强度
func validatePassword(fl validator.FieldLevel) bool {
	if fl.Field().String() == "" {
		return true
	}

	password := fl.Field().String()

	// 密码长度至少为8位
	if len(password) < 8 {
		return false
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	// 要求密码包含大小写字母、数字和特殊字符
	return hasUpper && hasLower && hasNumber && hasSpecial
}

// validateDateStr 验证日期字符串格式
func validateDateStr(fl validator.FieldLevel) bool {
	dateStr := fl.Field().String()
	if dateStr == "" {
		return true
	}

	// 检查格式
	if !dateRegex.MatchString(dateStr) {
		return false
	}

	// 解析日期
	_, err := time.Parse("2006-01-02", dateStr)
	return err == nil
}

// validateNoSpecialChars 验证字符串不包含特殊字符
func validateNoSpecialChars(fl validator.FieldLevel) bool {
	if fl.Field().String() == "" {
		return true
	}

	str := fl.Field().String()
	for _, r := range str {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != ' ' && r != '-' && r != '_' {
			return false
		}
	}

	return true
}

// validateIntRange 验证整数在指定范围内
func validateIntRange(fl validator.FieldLevel) bool {
	param := fl.Param()
	if param == "" {
		return true
	}

	// 从参数中获取范围（格式：min,max）
	parts := strings.Split(param, ",")
	if len(parts) != 2 {
		return false
	}

	min, err1 := strconv.ParseInt(parts[0], 10, 64)
	max, err2 := strconv.ParseInt(parts[1], 10, 64)

	if err1 != nil || err2 != nil {
		return false
	}

	val := fl.Field().Int()
	return val >= min && val <= max
}

// IsValidIPv4 验证IPv4地址
func IsValidIPv4(ip string) bool {
	if ip == "" {
		return false
	}

	// 使用标准库进行IP验证
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	// 确保是IPv4地址
	return ipv4Regex.MatchString(ip)
}

// ValidateChinaMobile 验证中国大陆手机号（可直接调用）
func ValidateChinaMobile(phone string) bool {
	return mobileRegex.MatchString(phone)
}

// ValidatePasswordStrength 验证密码强度（可直接调用）
func ValidatePasswordStrength(password string) (bool, string) {
	// 密码长度至少为8位
	if len(password) < 8 {
		return false, "密码长度不能少于8位"
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return false, "密码必须包含大写字母"
	}
	if !hasLower {
		return false, "密码必须包含小写字母"
	}
	if !hasNumber {
		return false, "密码必须包含数字"
	}
	if !hasSpecial {
		return false, "密码必须包含特殊字符"
	}

	return true, ""
}
