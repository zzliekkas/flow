package middleware

import (
	"time"

	"github.com/sirupsen/logrus"
	"github.com/zzliekkas/flow"
)

// LoggerConfig 是日志中间件的配置选项
type LoggerConfig struct {
	// SkipPaths 是不需要记录日志的路径
	SkipPaths []string

	// LogLevel 是日志级别
	LogLevel logrus.Level

	// Formatter 是日志格式化器
	Formatter logrus.Formatter

	// Output 是日志输出目标
	Output logrus.FieldLogger
}

// LoggerDefaultConfig 返回日志中间件的默认配置
func LoggerDefaultConfig() LoggerConfig {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	})

	return LoggerConfig{
		SkipPaths: []string{},
		LogLevel:  logrus.InfoLevel,
		Formatter: &logrus.TextFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
			FullTimestamp:   true,
		},
		Output: logger,
	}
}

// Logger 返回一个日志中间件
func Logger() flow.HandlerFunc {
	return LoggerWithConfig(LoggerDefaultConfig())
}

// LoggerWithConfig 返回一个使用指定配置的日志中间件
func LoggerWithConfig(config LoggerConfig) flow.HandlerFunc {
	// 确保配置有效
	if config.Formatter != nil && config.Output != nil {
		if logger, ok := config.Output.(*logrus.Logger); ok {
			logger.SetFormatter(config.Formatter)
		}
	}

	if config.Output == nil {
		config.Output = logrus.StandardLogger()
	}

	return func(c *flow.Context) {
		// 处理请求开始时间
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// 跳过不需要记录日志的路径
		for _, skip := range config.SkipPaths {
			if path == skip {
				c.Next()
				return
			}
		}

		// 处理请求
		c.Next()

		// 请求结束后记录日志
		end := time.Now()
		latency := end.Sub(start)

		// 获取客户端IP和状态码
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		// 根据状态码选择日志级别
		var logFunc func(format string, args ...interface{})
		switch {
		case statusCode >= 500:
			logFunc = config.Output.Errorf
		case statusCode >= 400:
			logFunc = config.Output.Warnf
		case statusCode >= 300:
			logFunc = config.Output.Infof
		default:
			logFunc = config.Output.Infof
		}

		// 记录日志
		logFunc("[Flow] %s | %3d | %13v | %15s | %-7s %s",
			end.Format("2006/01/02 - 15:04:05"),
			statusCode,
			latency,
			clientIP,
			method,
			path,
		)

		// 如果有错误，记录错误详情
		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				config.Output.Errorf("Error: %v", e.Err)
			}
		}
	}
}
