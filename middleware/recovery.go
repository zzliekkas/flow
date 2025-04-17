package middleware

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/zzliekkas/flow"
)

// RecoveryConfig 是恢复中间件的配置选项
type RecoveryConfig struct {
	// DisableStackAll 禁用完整堆栈跟踪
	DisableStackAll bool

	// DisablePrintStack 禁用打印堆栈信息
	DisablePrintStack bool

	// MaxStackSize 最大堆栈大小
	MaxStackSize int
}

// RecoveryDefaultConfig 返回恢复中间件的默认配置
func RecoveryDefaultConfig() RecoveryConfig {
	return RecoveryConfig{
		DisableStackAll:   false,
		DisablePrintStack: false,
		MaxStackSize:      2048,
	}
}

// Recovery 返回一个恢复中间件
func Recovery() flow.HandlerFunc {
	return RecoveryWithConfig(RecoveryDefaultConfig())
}

// RecoveryWithConfig 返回一个使用指定配置的恢复中间件
func RecoveryWithConfig(config RecoveryConfig) flow.HandlerFunc {
	return func(c *flow.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 检查是否已经写入响应头
				if c.Writer.Written() {
					return
				}

				stack := make([]byte, config.MaxStackSize)
				stackSize := runtime.Stack(stack, config.DisableStackAll)
				stack = stack[:stackSize]

				// 打印堆栈信息
				if !config.DisablePrintStack {
					fmt.Printf("[Flow] panic recovered:\n%s\n%s\n", err, stack)
				}

				// 创建错误响应
				errMsg := fmt.Sprintf("%v", err)
				httpErr := &flow.HTTPError{
					Code:    http.StatusInternalServerError,
					Message: errMsg,
				}

				// 添加错误到上下文
				c.Error(fmt.Errorf("%v", err))

				// 返回JSON错误响应
				c.JSON(httpErr.Code, flow.H{
					"error": httpErr.Message,
				})
			}
		}()

		c.Next()
	}
}

// HTTPError 表示HTTP错误
type HTTPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Error 实现error接口
func (e *HTTPError) Error() string {
	return fmt.Sprintf("code=%d, message=%s", e.Code, e.Message)
}
