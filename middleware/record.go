package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// 日志记录输出位置
const (
	LogToConsole = iota
	LogToFile
	LogToDatabase
)

// 日志详细程度级别
const (
	LogLevelBasic = iota
	LogLevelStandard
	LogLevelDetailed
	LogLevelFull
)

// 敏感字段正则表达式
var sensitiveFieldsRegex = regexp.MustCompile(`(?i)(password|token|secret|key|auth|credit|card)`)

// ErrorLogWriterFunc 是错误日志写入函数
type ErrorLogWriterFunc func(error)

// RecordLogWriter 接口定义了日志写入器
type RecordLogWriter interface {
	WriteLog(entry *RequestLogEntry) error
	Close() error
}

// RequestLogEntry 包含请求和响应日志信息
type RequestLogEntry struct {
	Timestamp       time.Time     `json:"timestamp"`
	RequestID       string        `json:"request_id"`
	ClientIP        string        `json:"client_ip"`
	Method          string        `json:"method"`
	Path            string        `json:"path"`
	Query           string        `json:"query,omitempty"`
	UserAgent       string        `json:"user_agent,omitempty"`
	RequestHeaders  interface{}   `json:"request_headers,omitempty"`
	RequestBody     interface{}   `json:"request_body,omitempty"`
	ResponseStatus  int           `json:"response_status"`
	ResponseSize    int           `json:"response_size,omitempty"`
	ResponseHeaders interface{}   `json:"response_headers,omitempty"`
	ResponseBody    interface{}   `json:"response_body,omitempty"`
	Latency         time.Duration `json:"latency"`
	Error           string        `json:"error,omitempty"`
}

// RecordConfig 配置请求记录中间件的行为
type RecordConfig struct {
	// 日志输出位置: LogToConsole, LogToFile, LogToDatabase
	LogDestination int
	// 日志文件路径，当LogDestination为LogToFile时使用
	LogFilePath string
	// 自定义日志写入器，如果提供则忽略LogDestination
	LogWriter RecordLogWriter
	// 日志详细级别: LogLevelBasic, LogLevelStandard, LogLevelDetailed, LogLevelFull
	LogLevel int
	// 是否屏蔽敏感信息
	MaskSensitiveData bool
	// 自定义敏感字段列表，补充默认的正则表达式
	SensitiveFields []string
	// 请求路径排除列表，这些路径不会被记录
	SkipPaths []string
	// 记录请求体的最大大小(字节)
	MaxBodySize int
	// 错误处理函数
	ErrorLogWriter ErrorLogWriterFunc
	// 是否记录健康检查请求
	SkipHealthChecks bool
	// 健康检查路径
	HealthCheckPath string
}

// DefaultRecordConfig 返回默认记录配置
func DefaultRecordConfig() RecordConfig {
	return RecordConfig{
		LogDestination:    LogToConsole,
		LogLevel:          LogLevelStandard,
		MaskSensitiveData: true,
		MaxBodySize:       1024 * 10, // 10KB默认
		SkipHealthChecks:  true,
		HealthCheckPath:   "/health",
		ErrorLogWriter: func(err error) {
			fmt.Fprintf(os.Stderr, "Record middleware error: %v\n", err)
		},
	}
}

// ConsoleLogWriter 实现控制台日志写入
type ConsoleLogWriter struct{}

// WriteLog 将日志条目写入控制台
func (w *ConsoleLogWriter) WriteLog(entry *RequestLogEntry) error {
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

// Close 实现接口
func (w *ConsoleLogWriter) Close() error {
	return nil
}

// FileLogWriter 实现文件日志写入
type FileLogWriter struct {
	file *os.File
}

// NewFileLogWriter 创建新的文件日志写入器
func NewFileLogWriter(filePath string) (*FileLogWriter, error) {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &FileLogWriter{file: file}, nil
}

// WriteLog 将日志条目写入文件
func (w *FileLogWriter) WriteLog(entry *RequestLogEntry) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	_, err = w.file.Write(append(data, '\n'))
	return err
}

// Close 关闭文件
func (w *FileLogWriter) Close() error {
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

// maskSensitiveData 遮盖敏感数据
func maskSensitiveData(data map[string]interface{}, sensitiveFields []string) map[string]interface{} {
	result := make(map[string]interface{})

	// 创建自定义敏感字段映射，便于快速查询
	customSensitiveMap := make(map[string]bool)
	for _, field := range sensitiveFields {
		customSensitiveMap[strings.ToLower(field)] = true
	}

	for k, v := range data {
		// 检查是否为敏感字段
		if customSensitiveMap[strings.ToLower(k)] || sensitiveFieldsRegex.MatchString(k) {
			// 对不同类型进行不同处理
			switch val := v.(type) {
			case string:
				if len(val) > 0 {
					result[k] = "***MASKED***"
				} else {
					result[k] = val
				}
			case map[string]interface{}:
				result[k] = maskSensitiveData(val, sensitiveFields)
			default:
				result[k] = "***MASKED***"
			}
		} else if nestedMap, ok := v.(map[string]interface{}); ok {
			// 递归处理嵌套结构
			result[k] = maskSensitiveData(nestedMap, sensitiveFields)
		} else {
			result[k] = v
		}
	}

	return result
}

// ResponseBodyWriter 自定义响应体写入器，用于捕获响应
type ResponseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write 捕获写入的响应体
func (w ResponseBodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// WriteString 捕获写入的字符串
func (w ResponseBodyWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

// Record 返回记录请求和响应的中间件
func Record(config RecordConfig) gin.HandlerFunc {
	var logWriter RecordLogWriter

	if config.LogWriter != nil {
		logWriter = config.LogWriter
	} else {
		switch config.LogDestination {
		case LogToFile:
			var err error
			logWriter, err = NewFileLogWriter(config.LogFilePath)
			if err != nil && config.ErrorLogWriter != nil {
				config.ErrorLogWriter(fmt.Errorf("failed to create file log writer: %w", err))
				logWriter = &ConsoleLogWriter{} // 回退到控制台
			}
		default:
			logWriter = &ConsoleLogWriter{}
		}
	}

	return func(c *gin.Context) {
		// 检查是否跳过此路径
		path := c.Request.URL.Path

		// 跳过健康检查路径
		if config.SkipHealthChecks && path == config.HealthCheckPath {
			c.Next()
			return
		}

		// 检查跳过路径
		for _, skipPath := range config.SkipPaths {
			if path == skipPath {
				c.Next()
				return
			}
		}

		// 记录开始时间
		startTime := time.Now()

		// 准备日志条目
		entry := &RequestLogEntry{
			Timestamp: startTime,
			RequestID: c.GetString("RequestID"),
			ClientIP:  c.ClientIP(),
			Method:    c.Request.Method,
			Path:      path,
			UserAgent: c.Request.UserAgent(),
		}

		// 记录查询参数
		if len(c.Request.URL.RawQuery) > 0 {
			entry.Query = c.Request.URL.RawQuery
		}

		// 记录请求头
		if config.LogLevel >= LogLevelStandard {
			headers := make(map[string]interface{})
			for k, v := range c.Request.Header {
				if len(v) == 1 {
					headers[k] = v[0]
				} else {
					headers[k] = v
				}
			}

			if config.MaskSensitiveData {
				entry.RequestHeaders = maskSensitiveData(headers, config.SensitiveFields)
			} else {
				entry.RequestHeaders = headers
			}
		}

		// 记录请求体
		if config.LogLevel >= LogLevelDetailed && c.Request.Body != nil && c.Request.ContentLength > 0 {
			if c.Request.ContentLength <= int64(config.MaxBodySize) {
				// 读取请求体
				bodyBytes, _ := io.ReadAll(c.Request.Body)
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // 重置请求体供后续中间件使用

				// 尝试解析JSON
				var bodyData map[string]interface{}
				if err := json.Unmarshal(bodyBytes, &bodyData); err == nil {
					// 成功解析为JSON
					if config.MaskSensitiveData {
						entry.RequestBody = maskSensitiveData(bodyData, config.SensitiveFields)
					} else {
						entry.RequestBody = bodyData
					}
				} else {
					// 不是JSON，使用字符串
					entry.RequestBody = string(bodyBytes)
				}
			} else {
				entry.RequestBody = fmt.Sprintf("[Content too large: %d bytes]", c.Request.ContentLength)
			}
		}

		// 捕获响应
		if config.LogLevel >= LogLevelStandard {
			respBodyWriter := &ResponseBodyWriter{
				ResponseWriter: c.Writer,
				body:           bytes.NewBufferString(""),
			}
			c.Writer = respBodyWriter

			// 处理请求
			c.Next()

			// 记录响应状态
			entry.ResponseStatus = c.Writer.Status()
			entry.ResponseSize = c.Writer.Size()
			entry.Latency = time.Since(startTime)

			// 记录响应头
			if config.LogLevel >= LogLevelDetailed {
				headers := make(map[string]interface{})
				for k, v := range c.Writer.Header() {
					if len(v) == 1 {
						headers[k] = v[0]
					} else {
						headers[k] = v
					}
				}

				if config.MaskSensitiveData {
					entry.ResponseHeaders = maskSensitiveData(headers, config.SensitiveFields)
				} else {
					entry.ResponseHeaders = headers
				}
			}

			// 记录响应体
			if config.LogLevel >= LogLevelFull && respBodyWriter.body.Len() > 0 && respBodyWriter.body.Len() <= config.MaxBodySize {
				bodyBytes := respBodyWriter.body.Bytes()

				// 尝试解析JSON
				var bodyData map[string]interface{}
				if err := json.Unmarshal(bodyBytes, &bodyData); err == nil {
					// 成功解析为JSON
					if config.MaskSensitiveData {
						entry.ResponseBody = maskSensitiveData(bodyData, config.SensitiveFields)
					} else {
						entry.ResponseBody = bodyData
					}
				} else {
					// 不是JSON，使用字符串
					entry.ResponseBody = string(bodyBytes)
				}
			} else if respBodyWriter.body.Len() > config.MaxBodySize {
				entry.ResponseBody = fmt.Sprintf("[Content too large: %d bytes]", respBodyWriter.body.Len())
			}
		} else {
			// 如果不记录详细信息，只需处理请求并记录基本信息
			c.Next()

			entry.ResponseStatus = c.Writer.Status()
			entry.ResponseSize = c.Writer.Size()
			entry.Latency = time.Since(startTime)
		}

		// 记录错误
		if len(c.Errors) > 0 {
			entry.Error = c.Errors.String()
		}

		// 写入日志
		if err := logWriter.WriteLog(entry); err != nil && config.ErrorLogWriter != nil {
			config.ErrorLogWriter(fmt.Errorf("failed to write log: %w", err))
		}
	}
}

// WithRequestID 添加请求ID的中间件
func WithRequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取请求ID，如果没有则生成一个
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Unix())
		}

		// 设置请求ID到上下文
		c.Set("RequestID", requestID)

		// 添加请求ID到响应头
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}
