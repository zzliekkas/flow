package app

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Environment 环境信息结构体
type Environment struct {
	// 系统信息
	GoVersion   string            // Go版本
	GOOS        string            // 操作系统
	GOARCH      string            // 系统架构
	NumCPU      int               // CPU核心数
	StartTime   time.Time         // 启动时间
	Hostname    string            // 主机名
	WorkingDir  string            // 工作目录
	Executable  string            // 可执行文件路径
	Environment map[string]string // 环境变量

	// 应用配置
	AppEnv     string // 应用环境 (development, testing, production)
	AppVersion string // 应用版本
	AppName    string // 应用名称
	Debug      bool   // 是否为调试模式
}

// NewEnvironment 创建一个新的环境信息实例
func NewEnvironment() *Environment {
	hostname, _ := os.Hostname()
	wd, _ := os.Getwd()
	exe, _ := os.Executable()

	// 获取环境变量
	envVars := make(map[string]string)
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if len(pair) == 2 {
			envVars[pair[0]] = pair[1]
		}
	}

	// 确定应用环境
	appEnv := os.Getenv("FLOW_ENV")
	if appEnv == "" {
		appEnv = "development"
	}

	// 确定应用版本
	appVersion := os.Getenv("FLOW_VERSION")
	if appVersion == "" {
		appVersion = "0.1.0" // 默认版本
	}

	// 确定应用名称
	appName := os.Getenv("FLOW_APP")
	if appName == "" {
		appName = filepath.Base(exe) // 使用可执行文件名作为应用名
	}

	// 调试模式
	debug := appEnv == "development" || strings.ToLower(os.Getenv("FLOW_DEBUG")) == "true"

	return &Environment{
		GoVersion:   runtime.Version(),
		GOOS:        runtime.GOOS,
		GOARCH:      runtime.GOARCH,
		NumCPU:      runtime.NumCPU(),
		StartTime:   time.Now(),
		Hostname:    hostname,
		WorkingDir:  wd,
		Executable:  exe,
		Environment: envVars,
		AppEnv:      appEnv,
		AppVersion:  appVersion,
		AppName:     appName,
		Debug:       debug,
	}
}

// IsDevelopment 检查是否为开发环境
func (e *Environment) IsDevelopment() bool {
	return e.AppEnv == "development"
}

// IsProduction 检查是否为生产环境
func (e *Environment) IsProduction() bool {
	return e.AppEnv == "production"
}

// IsTesting 检查是否为测试环境
func (e *Environment) IsTesting() bool {
	return e.AppEnv == "testing"
}

// GetEnv 获取环境变量，如果不存在则返回默认值
func (e *Environment) GetEnv(key, defaultValue string) string {
	if value, exists := e.Environment[key]; exists {
		return value
	}
	return defaultValue
}

// Uptime 获取应用运行时间
func (e *Environment) Uptime() time.Duration {
	return time.Since(e.StartTime)
}

// Summary 获取环境摘要信息
func (e *Environment) Summary() string {
	summary := strings.Builder{}
	summary.WriteString(fmt.Sprintf("应用: %s (版本: %s)\n", e.AppName, e.AppVersion))
	summary.WriteString(fmt.Sprintf("环境: %s (调试模式: %v)\n", e.AppEnv, e.Debug))
	summary.WriteString(fmt.Sprintf("系统: %s/%s (Go %s)\n", e.GOOS, e.GOARCH, e.GoVersion))
	summary.WriteString(fmt.Sprintf("CPU核心: %d\n", e.NumCPU))
	summary.WriteString(fmt.Sprintf("主机名: %s\n", e.Hostname))
	summary.WriteString(fmt.Sprintf("工作目录: %s\n", e.WorkingDir))
	summary.WriteString(fmt.Sprintf("可执行文件: %s\n", e.Executable))
	summary.WriteString(fmt.Sprintf("启动时间: %s\n", e.StartTime.Format(time.RFC3339)))
	return summary.String()
}
