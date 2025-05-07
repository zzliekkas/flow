package util

import (
	"net/url"
	"regexp"
	"strings"
)

// MaskDSN 掩盖DSN中的敏感信息（如密码）
func MaskDSN(dsn string) string {
	if dsn == "" {
		return ""
	}

	// 尝试处理标准URI格式的DSN
	if strings.Contains(dsn, "://") {
		u, err := url.Parse(dsn)
		if err == nil && u.User != nil {
			if _, hasPassword := u.User.Password(); hasPassword {
				maskedUser := u.User.Username() + ":********"
				maskedDsn := strings.Replace(dsn, u.User.String(), maskedUser, 1)
				return maskedDsn
			}
		}
	}

	// 处理键值对格式的DSN
	passwordRegex := regexp.MustCompile(`(password|passwd|pwd)=([^;& ]+)`)
	maskedDsn := passwordRegex.ReplaceAllString(dsn, "$1=********")

	// 处理MySQL格式的DSN (username:password@tcp(...))
	mysqlRegex := regexp.MustCompile(`([^:@]+):([^@]+)@`)
	maskedDsn = mysqlRegex.ReplaceAllString(maskedDsn, "$1:********@")

	return maskedDsn
}
