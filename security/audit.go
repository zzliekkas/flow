package security

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

// AuditLogger 安全审计日志记录器接口
type AuditLogger interface {
	// LogEvent 记录安全审计事件
	LogEvent(eventType string, userID string, action string, resource string, success bool, details map[string]interface{}) error

	// LogAuthentication 记录认证事件
	LogAuthentication(userID string, success bool, ipAddress string, details map[string]interface{}) error

	// LogAccessControl 记录访问控制事件
	LogAccessControl(userID string, action string, resource string, success bool, details map[string]interface{}) error

	// LogDataAccess 记录数据访问事件
	LogDataAccess(userID string, action string, resource string, success bool, details map[string]interface{}) error

	// LogSensitiveAction 记录敏感操作事件
	LogSensitiveAction(userID string, action string, resource string, success bool, details map[string]interface{}) error

	// GetLogs 获取指定类型的审计日志（一般用于测试/调试）
	GetLogs(eventType string, limit int) ([]AuditEvent, error)

	// LogRequest 记录HTTP请求
	LogRequest(r *http.Request) error
}

// AuditEvent 审计事件结构
type AuditEvent struct {
	// ID 事件唯一标识
	ID string `json:"id"`

	// Timestamp 事件时间戳
	Timestamp time.Time `json:"timestamp"`

	// EventType 事件类型
	EventType string `json:"event_type"`

	// UserID 用户标识
	UserID string `json:"user_id"`

	// Action 执行的操作
	Action string `json:"action"`

	// Resource 操作的资源
	Resource string `json:"resource"`

	// Success 操作是否成功
	Success bool `json:"success"`

	// Details 事件详细信息
	Details map[string]interface{} `json:"details,omitempty"`
}

// AuditLoggerImpl 审计日志记录器实现
type AuditLoggerImpl struct {
	// 配置
	config AuditConfig

	// 文件句柄（用于文件日志）
	file *os.File

	// 互斥锁（确保并发安全）
	mu sync.Mutex

	// 内存中保存的日志（用于调试/测试）
	logs []AuditEvent
}

// NewAuditLogger 创建审计日志记录器
func NewAuditLogger(config AuditConfig) AuditLogger {
	logger := &AuditLoggerImpl{
		config: config,
		logs:   make([]AuditEvent, 0, 100), // 内存中只保留最近的100条记录
	}

	// 如果目标是文件，确保文件可写
	if config.Enabled && config.Destination == "file" {
		// 确保日志目录存在
		dir := config.FilePath
		if lastSlash := len(dir) - 1; lastSlash >= 0 && dir[lastSlash] != '/' && dir[lastSlash] != '\\' {
			dir = dir[:lastSlash]
		}
		if dir != "" {
			os.MkdirAll(dir, 0755)
		}

		// 打开日志文件，如果不存在则创建
		file, err := os.OpenFile(config.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			logger.file = file
		} else {
			// 打开文件失败，记录错误并退回到内存日志
			fmt.Fprintf(os.Stderr, "Failed to open audit log file: %v\n", err)
		}
	}

	return logger
}

// LogEvent 记录安全审计事件
func (l *AuditLoggerImpl) LogEvent(eventType string, userID string, action string, resource string, success bool, details map[string]interface{}) error {
	if !l.config.Enabled {
		return nil
	}

	// 根据事件类型检查是否需要记录
	switch eventType {
	case "authentication":
		if !l.config.LogAuthenticationEvents {
			return nil
		}
	case "access_control":
		if !l.config.LogAccessControl {
			return nil
		}
	case "data_access":
		if !l.config.LogDataAccess {
			return nil
		}
	case "sensitive_action":
		if !l.config.LogSensitiveActions {
			return nil
		}
	}

	// 创建审计事件
	event := AuditEvent{
		ID:        generateEventID(),
		Timestamp: time.Now(),
		EventType: eventType,
		UserID:    userID,
		Action:    action,
		Resource:  resource,
		Success:   success,
		Details:   details,
	}

	// 记录事件
	return l.logEvent(event)
}

// LogAuthentication 记录认证事件
func (l *AuditLoggerImpl) LogAuthentication(userID string, success bool, ipAddress string, details map[string]interface{}) error {
	if !l.config.Enabled || !l.config.LogAuthenticationEvents {
		return nil
	}

	if details == nil {
		details = make(map[string]interface{})
	}
	details["ip_address"] = ipAddress

	return l.LogEvent("authentication", userID, "login", "", success, details)
}

// LogAccessControl 记录访问控制事件
func (l *AuditLoggerImpl) LogAccessControl(userID string, action string, resource string, success bool, details map[string]interface{}) error {
	if !l.config.Enabled || !l.config.LogAccessControl {
		return nil
	}

	return l.LogEvent("access_control", userID, action, resource, success, details)
}

// LogDataAccess 记录数据访问事件
func (l *AuditLoggerImpl) LogDataAccess(userID string, action string, resource string, success bool, details map[string]interface{}) error {
	if !l.config.Enabled || !l.config.LogDataAccess {
		return nil
	}

	return l.LogEvent("data_access", userID, action, resource, success, details)
}

// LogSensitiveAction 记录敏感操作事件
func (l *AuditLoggerImpl) LogSensitiveAction(userID string, action string, resource string, success bool, details map[string]interface{}) error {
	if !l.config.Enabled || !l.config.LogSensitiveActions {
		return nil
	}

	return l.LogEvent("sensitive_action", userID, action, resource, success, details)
}

// LogRequest 记录HTTP请求
func (l *AuditLoggerImpl) LogRequest(r *http.Request) error {
	if !l.config.Enabled {
		return nil
	}

	// 获取用户ID（实际项目中应从上下文或认证会话中获取）
	userID := "anonymous"

	// 创建事件详情
	details := map[string]interface{}{
		"method":     r.Method,
		"path":       r.URL.Path,
		"remote_ip":  getClientIP(r),
		"user_agent": r.UserAgent(),
	}

	// 记录请求事件
	return l.LogEvent("HTTP_REQUEST", userID, "request", r.URL.Path, true, details)
}

// GetLogs 获取指定类型的审计日志
func (l *AuditLoggerImpl) GetLogs(eventType string, limit int) ([]AuditEvent, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if limit <= 0 || limit > len(l.logs) {
		limit = len(l.logs)
	}

	// 根据事件类型过滤日志
	var filteredLogs []AuditEvent
	if eventType == "" {
		// 复制最近的limit条日志
		filteredLogs = append(filteredLogs, l.logs[len(l.logs)-limit:]...)
	} else {
		// 过滤指定类型的日志
		for i := len(l.logs) - 1; i >= 0 && len(filteredLogs) < limit; i-- {
			if l.logs[i].EventType == eventType {
				filteredLogs = append(filteredLogs, l.logs[i])
			}
		}
	}

	return filteredLogs, nil
}

// logEvent 内部方法，记录事件
func (l *AuditLoggerImpl) logEvent(event AuditEvent) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 将事件转换为JSON
	jsonData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal audit event: %w", err)
	}

	// 根据目标类型记录日志
	switch l.config.Destination {
	case "file":
		if l.file != nil {
			// 写入文件
			if _, err := l.file.Write(append(jsonData, '\n')); err != nil {
				return fmt.Errorf("failed to write audit log: %w", err)
			}
		}
	case "database":
		// 数据库记录逻辑将在后续实现
		// 目前仅保存在内存中
	case "syslog":
		// 系统日志记录逻辑将在后续实现
		// 目前仅保存在内存中
	}

	// 同时在内存中保存一份（最近的100条）
	l.logs = append(l.logs, event)
	if len(l.logs) > 100 {
		l.logs = l.logs[1:]
	}

	return nil
}

// Close 关闭日志记录器，释放资源
func (l *AuditLoggerImpl) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		err := l.file.Close()
		l.file = nil
		return err
	}

	return nil
}

// generateEventID 生成唯一的事件ID
func generateEventID() string {
	return fmt.Sprintf("%d-%x", time.Now().UnixNano(), time.Now().UnixNano()%1000)
}

// getClientIP 获取客户端IP地址
func getClientIP(r *http.Request) string {
	// 检查X-Forwarded-For头
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	// 获取远程地址
	return r.RemoteAddr
}
