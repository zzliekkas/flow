package security

import (
	"net/http"
	"sync"
	"time"
)

// BasicAuditLogger 提供 AuditLogger 接口的简单实现
type BasicAuditLogger struct {
	// 配置
	config AuditConfig
	// 互斥锁（确保并发安全）
	mu sync.Mutex
	// 内存中保存的日志（用于调试/测试）
	logs []AuditEvent
}

// NewBasicAuditLogger 创建简单审计日志记录器
func NewBasicAuditLogger(config AuditConfig) *BasicAuditLogger {
	return &BasicAuditLogger{
		config: config,
		logs:   make([]AuditEvent, 0, 100),
	}
}

// LogEvent 记录安全审计事件
func (l *BasicAuditLogger) LogEvent(eventType string, userID string, action string, resource string, success bool, details map[string]interface{}) error {
	// 简单实现，实际项目中应根据配置写入不同的目标（文件、数据库等）
	if !l.config.Enabled {
		return nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

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

	// 将事件添加到内存日志中
	l.logs = append(l.logs, event)
	// 保持日志不超过最大容量
	if len(l.logs) > 100 {
		l.logs = l.logs[1:]
	}

	// 实际项目中应实现持久化逻辑
	return nil
}

// LogRequest 记录HTTP请求
func (l *BasicAuditLogger) LogRequest(r *http.Request) error {
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

// LogAuthentication 记录认证事件
func (l *BasicAuditLogger) LogAuthentication(userID string, success bool, ipAddress string, details map[string]interface{}) error {
	if details == nil {
		details = make(map[string]interface{})
	}
	details["ip_address"] = ipAddress
	return l.LogEvent("authentication", userID, "authentication", "", success, details)
}

// LogAccessControl 记录访问控制事件
func (l *BasicAuditLogger) LogAccessControl(userID string, action string, resource string, success bool, details map[string]interface{}) error {
	return l.LogEvent("access_control", userID, action, resource, success, details)
}

// LogDataAccess 记录数据访问事件
func (l *BasicAuditLogger) LogDataAccess(userID string, action string, resource string, success bool, details map[string]interface{}) error {
	return l.LogEvent("data_access", userID, action, resource, success, details)
}

// LogSensitiveAction 记录敏感操作事件
func (l *BasicAuditLogger) LogSensitiveAction(userID string, action string, resource string, success bool, details map[string]interface{}) error {
	return l.LogEvent("sensitive_action", userID, action, resource, success, details)
}

// GetLogs 获取指定类型的审计日志
func (l *BasicAuditLogger) GetLogs(eventType string, limit int) ([]AuditEvent, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if limit <= 0 || limit > len(l.logs) {
		limit = len(l.logs)
	}

	result := make([]AuditEvent, 0, limit)
	count := 0

	// 从最新的日志开始遍历
	for i := len(l.logs) - 1; i >= 0 && count < limit; i-- {
		if eventType == "" || l.logs[i].EventType == eventType {
			result = append(result, l.logs[i])
			count++
		}
	}

	return result, nil
}

// Close 关闭日志记录器，释放资源
func (l *BasicAuditLogger) Close() error {
	// 在基础实现中无需关闭任何资源
	return nil
}
