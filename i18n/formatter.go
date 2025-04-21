package i18n

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Formatter 提供语言相关的格式化功能
type Formatter struct {
	// 用于翻译日期和时间相关文本
	translator Translator

	// 缓存不同语言的格式
	dateFormats map[string]map[string]string

	// 货币符号和格式
	currencyFormats map[string]map[string]string

	// 数字格式化
	numberFormats map[string]map[string]string
}

// 常用的日期格式名称
const (
	DateFormatShort      = "short"          // 短日期: 2020/01/01
	DateFormatMedium     = "medium"         // 中等日期: Jan 1, 2020
	DateFormatLong       = "long"           // 长日期: January 1, 2020
	DateFormatFull       = "full"           // 完整日期: Wednesday, January 1, 2020
	TimeFormatShort      = "timeShort"      // 短时间: 13:45
	TimeFormatMedium     = "timeMedium"     // 中等时间: 13:45:30
	TimeFormatLong       = "timeLong"       // 长时间: 13:45:30 CST
	DateTimeFormatShort  = "dateTimeShort"  // 短日期时间: 2020/01/01 13:45
	DateTimeFormatMedium = "dateTimeMedium" // 中等日期时间: Jan 1, 2020 13:45:30
	DateTimeFormatLong   = "dateTimeLong"   // 长日期时间: January 1, 2020 13:45:30 CST
	DateTimeFormatFull   = "dateTimeFull"   // 完整日期时间: Wednesday, January 1, 2020 13:45:30 CST
)

// NewFormatter 创建一个新的格式化器
func NewFormatter(translator Translator) *Formatter {
	f := &Formatter{
		translator:      translator,
		dateFormats:     make(map[string]map[string]string),
		currencyFormats: make(map[string]map[string]string),
		numberFormats:   make(map[string]map[string]string),
	}

	// 初始化默认格式
	f.initDefaultFormats()

	return f
}

// 初始化默认的日期格式
func (f *Formatter) initDefaultFormats() {
	// 英语日期格式
	f.dateFormats["en"] = map[string]string{
		DateFormatShort:      "2006/01/02",
		DateFormatMedium:     "Jan 2, 2006",
		DateFormatLong:       "January 2, 2006",
		DateFormatFull:       "Monday, January 2, 2006",
		TimeFormatShort:      "15:04",
		TimeFormatMedium:     "15:04:05",
		TimeFormatLong:       "15:04:05 MST",
		DateTimeFormatShort:  "2006/01/02 15:04",
		DateTimeFormatMedium: "Jan 2, 2006 15:04:05",
		DateTimeFormatLong:   "January 2, 2006 15:04:05 MST",
		DateTimeFormatFull:   "Monday, January 2, 2006 15:04:05 MST",
	}

	// 中文日期格式
	f.dateFormats["zh"] = map[string]string{
		DateFormatShort:      "2006/01/02",
		DateFormatMedium:     "2006年1月2日",
		DateFormatLong:       "2006年1月2日",
		DateFormatFull:       "2006年1月2日 星期Monday",
		TimeFormatShort:      "15:04",
		TimeFormatMedium:     "15:04:05",
		TimeFormatLong:       "15:04:05 MST",
		DateTimeFormatShort:  "2006/01/02 15:04",
		DateTimeFormatMedium: "2006年1月2日 15:04:05",
		DateTimeFormatLong:   "2006年1月2日 15:04:05 MST",
		DateTimeFormatFull:   "2006年1月2日 星期Monday 15:04:05 MST",
	}

	// 设置默认货币格式
	f.currencyFormats["en"] = map[string]string{
		"USD": "$%s",
		"EUR": "€%s",
		"GBP": "£%s",
		"JPY": "¥%s",
		"CNY": "¥%s",
	}

	f.currencyFormats["zh"] = map[string]string{
		"USD": "$%s",
		"EUR": "€%s",
		"GBP": "£%s",
		"JPY": "¥%s",
		"CNY": "￥%s",
	}

	// 设置默认数字格式
	f.numberFormats["en"] = map[string]string{
		"decimal":    ".",
		"thousands":  ",",
		"grouping":   "3",
		"percentage": "%s%%",
	}

	f.numberFormats["zh"] = map[string]string{
		"decimal":    ".",
		"thousands":  ",",
		"grouping":   "3",
		"percentage": "%s%%",
	}
}

// RegisterDateFormat 注册特定语言的日期格式
func (f *Formatter) RegisterDateFormat(locale, formatName, formatString string) {
	if _, exists := f.dateFormats[locale]; !exists {
		f.dateFormats[locale] = make(map[string]string)
	}
	f.dateFormats[locale][formatName] = formatString
}

// FormatDate 格式化日期
func (f *Formatter) FormatDate(ctx context.Context, date time.Time, formatName string) string {
	locale := f.translator.GetLocale(ctx)
	return f.FormatDateWithLocale(locale, date, formatName)
}

// FormatDateWithLocale 使用指定语言格式化日期
func (f *Formatter) FormatDateWithLocale(locale string, date time.Time, formatName string) string {
	// 获取指定语言的格式
	localeFormats, exists := f.dateFormats[locale]
	if !exists {
		// 尝试回退到默认语言
		fallbackLocale := f.translator.GetFallbackLocale()
		localeFormats, exists = f.dateFormats[fallbackLocale]
		if !exists {
			// 使用英语作为最后回退
			localeFormats = f.dateFormats["en"]
		}
	}

	// 获取指定的格式
	format, exists := localeFormats[formatName]
	if !exists {
		// 回退到短日期格式
		format = localeFormats[DateFormatShort]
		if format == "" {
			// 最后的回退
			format = "2006/01/02"
		}
	}

	// 处理中文星期
	if locale == "zh" && strings.Contains(format, "星期") {
		weekday := date.Weekday()
		weekdayStr := ""
		switch weekday {
		case time.Sunday:
			weekdayStr = "日"
		case time.Monday:
			weekdayStr = "一"
		case time.Tuesday:
			weekdayStr = "二"
		case time.Wednesday:
			weekdayStr = "三"
		case time.Thursday:
			weekdayStr = "四"
		case time.Friday:
			weekdayStr = "五"
		case time.Saturday:
			weekdayStr = "六"
		}
		format = strings.Replace(format, "Monday", weekdayStr, 1)
	}

	return date.Format(format)
}

// FormatCurrency 格式化货币
func (f *Formatter) FormatCurrency(ctx context.Context, amount float64, currency string) string {
	locale := f.translator.GetLocale(ctx)
	return f.FormatCurrencyWithLocale(locale, amount, currency)
}

// FormatCurrencyWithLocale 使用指定语言格式化货币
func (f *Formatter) FormatCurrencyWithLocale(locale string, amount float64, currency string) string {
	// 获取货币格式
	localeFormats, exists := f.currencyFormats[locale]
	if !exists {
		// 尝试回退
		fallbackLocale := f.translator.GetFallbackLocale()
		localeFormats, exists = f.currencyFormats[fallbackLocale]
		if !exists {
			// 使用英语作为最后回退
			localeFormats = f.currencyFormats["en"]
		}
	}

	// 获取指定货币的格式
	format, exists := localeFormats[currency]
	if !exists {
		// 如果没有特定格式，使用通用格式
		format = "%s " + currency
	}

	// 格式化金额
	amountStr := f.FormatNumberWithLocale(locale, amount, 2)

	return fmt.Sprintf(format, amountStr)
}

// FormatNumber 格式化数字
func (f *Formatter) FormatNumber(ctx context.Context, number float64, decimals int) string {
	locale := f.translator.GetLocale(ctx)
	return f.FormatNumberWithLocale(locale, number, decimals)
}

// FormatNumberWithLocale 使用指定语言格式化数字
func (f *Formatter) FormatNumberWithLocale(locale string, number float64, decimals int) string {
	// 获取数字格式
	localeFormats, exists := f.numberFormats[locale]
	if !exists {
		// 尝试回退
		fallbackLocale := f.translator.GetFallbackLocale()
		localeFormats, exists = f.numberFormats[fallbackLocale]
		if !exists {
			// 使用英语作为最后回退
			localeFormats = f.numberFormats["en"]
		}
	}

	// 获取格式参数
	decimalSep := localeFormats["decimal"]
	thousandsSep := localeFormats["thousands"]
	groupingStr := localeFormats["grouping"]
	grouping, _ := strconv.Atoi(groupingStr)
	if grouping <= 0 {
		grouping = 3
	}

	// 格式化数字
	// 首先使用默认格式
	result := strconv.FormatFloat(number, 'f', decimals, 64)

	// 分离整数和小数部分
	parts := strings.Split(result, ".")
	intPart := parts[0]

	// 添加千位分隔符
	if thousandsSep != "" {
		intPart = f.addThousandsSeparator(intPart, thousandsSep, grouping)
	}

	// 重新组合
	if len(parts) > 1 {
		result = intPart + decimalSep + parts[1]
	} else {
		result = intPart
	}

	return result
}

// FormatPercentage 格式化百分比
func (f *Formatter) FormatPercentage(ctx context.Context, number float64, decimals int) string {
	locale := f.translator.GetLocale(ctx)
	return f.FormatPercentageWithLocale(locale, number, decimals)
}

// FormatPercentageWithLocale 使用指定语言格式化百分比
func (f *Formatter) FormatPercentageWithLocale(locale string, number float64, decimals int) string {
	// 将数字转换为百分数（乘以100）
	percentage := number * 100

	// 格式化数字
	formattedNumber := f.FormatNumberWithLocale(locale, percentage, decimals)

	// 获取百分比格式
	localeFormats, exists := f.numberFormats[locale]
	if !exists {
		fallbackLocale := f.translator.GetFallbackLocale()
		localeFormats, exists = f.numberFormats[fallbackLocale]
		if !exists {
			localeFormats = f.numberFormats["en"]
		}
	}

	percentFormat := localeFormats["percentage"]
	if percentFormat == "" {
		percentFormat = "%s%%"
	}

	return fmt.Sprintf(percentFormat, formattedNumber)
}

// RegisterCurrencyFormat 注册特定语言的货币格式
func (f *Formatter) RegisterCurrencyFormat(locale, currency, format string) {
	if _, exists := f.currencyFormats[locale]; !exists {
		f.currencyFormats[locale] = make(map[string]string)
	}
	f.currencyFormats[locale][currency] = format
}

// RegisterNumberFormat 注册特定语言的数字格式
func (f *Formatter) RegisterNumberFormat(locale, key, value string) {
	if _, exists := f.numberFormats[locale]; !exists {
		f.numberFormats[locale] = make(map[string]string)
	}
	f.numberFormats[locale][key] = value
}

// addThousandsSeparator 添加千位分隔符
func (f *Formatter) addThousandsSeparator(number, separator string, grouping int) string {
	length := len(number)
	if length <= grouping {
		return number
	}

	// 处理负数符号
	prefix := ""
	if number[0] == '-' {
		prefix = "-"
		number = number[1:]
		length--
	}

	// 分组添加分隔符
	var result []byte
	for i := 0; i < length; i++ {
		if i > 0 && (length-i)%grouping == 0 {
			result = append(result, separator...)
		}
		result = append(result, number[i])
	}

	return prefix + string(result)
}
