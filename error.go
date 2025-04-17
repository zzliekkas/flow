package flow

import (
	"fmt"
	"net/http"
)

// HTTPError 表示HTTP错误
type HTTPError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// Error 实现error接口
func (e *HTTPError) Error() string {
	return fmt.Sprintf("code=%d, message=%s", e.Code, e.Message)
}

// NewHTTPError 创建一个新的HTTP错误
func NewHTTPError(code int, message string) *HTTPError {
	return &HTTPError{
		Code:    code,
		Message: message,
	}
}

// NewHTTPErrorWithDetails 创建一个带有详细信息的HTTP错误
func NewHTTPErrorWithDetails(code int, message string, details interface{}) *HTTPError {
	return &HTTPError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// DefaultHTTPErrorHandler 是默认的HTTP错误处理函数
func DefaultHTTPErrorHandler(err error, c *Context) {
	code := http.StatusInternalServerError
	message := "内部服务器错误"
	details := make(map[string]interface{})

	if he, ok := err.(*HTTPError); ok {
		code = he.Code
		message = he.Message
		if he.Details != nil {
			details["details"] = he.Details
		}
	} else {
		details["error"] = err.Error()
	}

	// 发送错误响应
	if !c.Writer.Written() {
		c.JSON(code, H{
			"error":   message,
			"details": details,
		})
	}
}
