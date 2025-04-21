package security

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// PasswordPolicy 密码策略接口
type PasswordPolicy interface {
	// Validate 验证密码是否符合策略
	Validate(password string) (bool, []string)

	// GetRequirements 获取密码要求描述
	GetRequirements() []string

	// GetDefaultError 获取默认错误消息
	GetDefaultError() string
}

// PasswordPolicyImpl 密码策略实现
type PasswordPolicyImpl struct {
	// 配置
	config PasswordConfig

	// 常见密码列表 (真实实现应加载更大的列表)
	commonPasswords map[string]bool
}

// NewPasswordPolicy 创建密码策略实例
func NewPasswordPolicy(config PasswordConfig) PasswordPolicy {
	// 初始化简化版常见密码列表
	commonPasswords := map[string]bool{
		"password":    true,
		"123456":      true,
		"qwerty":      true,
		"admin":       true,
		"welcome":     true,
		"123456789":   true,
		"12345678":    true,
		"abc123":      true,
		"letmein":     true,
		"monkey":      true,
		"1234567":     true,
		"dragon":      true,
		"123123":      true,
		"football":    true,
		"iloveyou":    true,
		"000000":      true,
		"654321":      true,
		"sunshine":    true,
		"master":      true,
		"hottie":      true,
		"princess":    true,
		"baseball":    true,
		"shadow":      true,
		"superman":    true,
		"trustno1":    true,
		"welcome1":    true,
		"admin123":    true,
		"password1":   true,
		"password123": true,
	}

	return &PasswordPolicyImpl{
		config:          config,
		commonPasswords: commonPasswords,
	}
}

// Validate 验证密码是否符合策略
func (p *PasswordPolicyImpl) Validate(password string) (bool, []string) {
	if password == "" {
		return false, []string{"密码不能为空"}
	}

	var errors []string

	// 检查长度
	if len(password) < p.config.MinLength {
		errors = append(errors, fmt.Sprintf("密码长度必须至少为 %d 个字符", p.config.MinLength))
	}

	// 检查大写字母
	if p.config.RequireUpper && !containsUppercase(password) {
		errors = append(errors, "密码必须包含至少一个大写字母")
	}

	// 检查小写字母
	if p.config.RequireLower && !containsLowercase(password) {
		errors = append(errors, "密码必须包含至少一个小写字母")
	}

	// 检查数字
	if p.config.RequireNumber && !containsDigit(password) {
		errors = append(errors, "密码必须包含至少一个数字")
	}

	// 检查特殊字符
	specialChars := "!@#$%^&*()_-+={}[]\\|:;\"'<>,.?/"
	if p.config.RequireSpecial && !containsSpecialChar(password, specialChars) {
		errors = append(errors, fmt.Sprintf("密码必须包含至少一个特殊字符（%s）", specialChars))
	}

	// 检查常见密码
	if p.isCommonPassword(password) {
		errors = append(errors, "请勿使用常见密码，请选择更强的密码")
	}

	// 如果有错误，返回false和错误列表
	if len(errors) > 0 {
		return false, errors
	}

	return true, nil
}

// GetRequirements 获取密码要求描述
func (p *PasswordPolicyImpl) GetRequirements() []string {
	var requirements []string

	requirements = append(requirements, fmt.Sprintf("密码长度必须至少为 %d 个字符", p.config.MinLength))

	if p.config.RequireUpper {
		requirements = append(requirements, "密码必须包含至少一个大写字母")
	}

	if p.config.RequireLower {
		requirements = append(requirements, "密码必须包含至少一个小写字母")
	}

	if p.config.RequireNumber {
		requirements = append(requirements, "密码必须包含至少一个数字")
	}

	specialChars := "!@#$%^&*()_-+={}[]\\|:;\"'<>,.?/"
	if p.config.RequireSpecial {
		requirements = append(requirements, fmt.Sprintf("密码必须包含至少一个特殊字符（%s）", specialChars))
	}

	if len(p.commonPasswords) > 0 {
		requirements = append(requirements, "禁止使用常见密码")
	}

	if p.config.MaxAge > 0 {
		requirements = append(requirements, fmt.Sprintf("密码有效期为 %d 天", p.config.MaxAge))
	}

	if p.config.HistoryCount > 0 {
		requirements = append(requirements, fmt.Sprintf("新密码不能与最近 %d 个旧密码相同", p.config.HistoryCount))
	}

	return requirements
}

// GetDefaultError 获取默认错误消息
func (p *PasswordPolicyImpl) GetDefaultError() string {
	return "密码不符合安全要求，请选择更强的密码"
}

// 检查是否包含大写字母
func containsUppercase(s string) bool {
	for _, r := range s {
		if unicode.IsUpper(r) {
			return true
		}
	}
	return false
}

// 检查是否包含小写字母
func containsLowercase(s string) bool {
	for _, r := range s {
		if unicode.IsLower(r) {
			return true
		}
	}
	return false
}

// 检查是否包含数字
func containsDigit(s string) bool {
	for _, r := range s {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

// 检查是否包含特殊字符
func containsSpecialChar(s string, specialChars string) bool {
	for _, c := range s {
		if strings.ContainsRune(specialChars, c) {
			return true
		}
	}
	return false
}

// 检查是否为常见密码
func (p *PasswordPolicyImpl) isCommonPassword(password string) bool {
	// 转为小写比较
	lowerPassword := strings.ToLower(password)
	return p.commonPasswords[lowerPassword]
}

// PasswordStrength 密码强度级别
type PasswordStrength int

const (
	// 非常弱
	VeryWeak PasswordStrength = iota
	// 弱
	Weak
	// 中等
	Medium
	// 强
	Strong
	// 非常强
	VeryStrong
)

// 获取密码强度描述
func (s PasswordStrength) String() string {
	switch s {
	case VeryWeak:
		return "非常弱"
	case Weak:
		return "弱"
	case Medium:
		return "中等"
	case Strong:
		return "强"
	case VeryStrong:
		return "非常强"
	default:
		return "未知"
	}
}

// PasswordMeter 密码强度测量工具
type PasswordMeter struct {
	policy PasswordPolicy
}

// NewPasswordMeter 创建密码强度测量工具
func NewPasswordMeter(policy PasswordPolicy) *PasswordMeter {
	return &PasswordMeter{
		policy: policy,
	}
}

// GetStrength 评估密码强度
func (m *PasswordMeter) GetStrength(password string) PasswordStrength {
	if password == "" {
		return VeryWeak
	}

	// 基本分数
	score := 0

	// 密码长度评分
	length := len(password)
	if length >= 12 {
		score += 2
	} else if length >= 8 {
		score += 1
	}

	// 字符类型多样性评分
	if containsLowercase(password) {
		score++
	}
	if containsUppercase(password) {
		score++
	}
	if containsDigit(password) {
		score++
	}
	// 使用正则检查特殊字符
	specialCharRegex := regexp.MustCompile(`[^a-zA-Z0-9]`)
	if specialCharRegex.MatchString(password) {
		score++
	}

	// 连续相同字符会减分
	if regexp.MustCompile(`(.)\1{2,}`).MatchString(password) {
		score--
	}

	// 根据分数返回强度级别
	switch {
	case score <= 1:
		return VeryWeak
	case score == 2:
		return Weak
	case score == 3:
		return Medium
	case score == 4:
		return Strong
	default:
		return VeryStrong
	}
}
