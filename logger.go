package flow

import (
	"log"
	"os"
)

// Logger 定义Flow框架的日志接口
// 用户可以实现此接口来替换默认的日志行为
type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
}

// defaultLogger 是基于标准库log的默认日志实现
type defaultLogger struct {
	logger *log.Logger
	level  int
}

// 日志级别常量
const (
	LogLevelDebug = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

// newDefaultLogger 创建默认日志实例
func newDefaultLogger() *defaultLogger {
	return &defaultLogger{
		logger: log.New(os.Stdout, "", log.LstdFlags),
		level:  LogLevelInfo,
	}
}

// SetLevel 设置日志级别
func (l *defaultLogger) SetLevel(level int) {
	l.level = level
}

func (l *defaultLogger) Debug(args ...interface{}) {
	if l.level <= LogLevelDebug {
		l.logger.Println(append([]interface{}{"[DEBUG]"}, args...)...)
	}
}

func (l *defaultLogger) Debugf(format string, args ...interface{}) {
	if l.level <= LogLevelDebug {
		l.logger.Printf("[DEBUG] "+format, args...)
	}
}

func (l *defaultLogger) Info(args ...interface{}) {
	if l.level <= LogLevelInfo {
		l.logger.Println(append([]interface{}{"[INFO]"}, args...)...)
	}
}

func (l *defaultLogger) Infof(format string, args ...interface{}) {
	if l.level <= LogLevelInfo {
		l.logger.Printf("[INFO] "+format, args...)
	}
}

func (l *defaultLogger) Warn(args ...interface{}) {
	if l.level <= LogLevelWarn {
		l.logger.Println(append([]interface{}{"[WARN]"}, args...)...)
	}
}

func (l *defaultLogger) Warnf(format string, args ...interface{}) {
	if l.level <= LogLevelWarn {
		l.logger.Printf("[WARN] "+format, args...)
	}
}

func (l *defaultLogger) Error(args ...interface{}) {
	if l.level <= LogLevelError {
		l.logger.Println(append([]interface{}{"[ERROR]"}, args...)...)
	}
}

func (l *defaultLogger) Errorf(format string, args ...interface{}) {
	if l.level <= LogLevelError {
		l.logger.Printf("[ERROR] "+format, args...)
	}
}

// frameworkLogger 框架内部使用的全局日志实例
var frameworkLogger Logger = newDefaultLogger()

// SetLogger 设置框架全局日志实例
func SetLogger(l Logger) {
	if l != nil {
		frameworkLogger = l
	}
}

// GetLogger 获取框架全局日志实例
func GetLogger() Logger {
	return frameworkLogger
}

// WithLogger 返回一个设置日志实例的选项
func WithLogger(l Logger) Option {
	return func(e *Engine) {
		SetLogger(l)
	}
}

// flog 框架内部日志快捷方法（不导出）
var flog = struct {
	Debug  func(args ...interface{})
	Debugf func(format string, args ...interface{})
	Info   func(args ...interface{})
	Infof  func(format string, args ...interface{})
	Warn   func(args ...interface{})
	Warnf  func(format string, args ...interface{})
	Error  func(args ...interface{})
	Errorf func(format string, args ...interface{})
}{
	Debug:  func(args ...interface{}) { frameworkLogger.Debug(args...) },
	Debugf: func(format string, args ...interface{}) { frameworkLogger.Debugf(format, args...) },
	Info:   func(args ...interface{}) { frameworkLogger.Info(args...) },
	Infof:  func(format string, args ...interface{}) { frameworkLogger.Infof(format, args...) },
	Warn:   func(args ...interface{}) { frameworkLogger.Warn(args...) },
	Warnf:  func(format string, args ...interface{}) { frameworkLogger.Warnf(format, args...) },
	Error:  func(args ...interface{}) { frameworkLogger.Error(args...) },
	Errorf: func(format string, args ...interface{}) { frameworkLogger.Errorf(format, args...) },
}
