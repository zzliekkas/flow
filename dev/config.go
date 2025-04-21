package dev

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Config 开发环境配置
type Config struct {
	// 应用根目录
	RootDir string

	// 热重载配置
	Reload ReloadConfig

	// 调试配置
	Debug DebugConfig

	// 开发服务器配置
	Server ServerConfig

	// 环境变量
	Env map[string]string

	// 自定义设置
	Settings map[string]interface{}
}

// ReloadConfig 热重载配置
type ReloadConfig struct {
	// 是否启用热重载
	Enabled bool

	// 监视的目录列表
	WatchDirs []string

	// 忽略的文件模式
	IgnorePatterns []string

	// 重载延迟时间（毫秒）
	DelayMilliseconds int

	// 重载回调
	OnReload func()

	// 最大触发频率（秒）
	MaxFrequency int
}

// DebugConfig 调试配置
type DebugConfig struct {
	// 是否启用详细日志
	VerboseLogging bool

	// 是否显示路由信息
	ShowRoutes bool

	// 是否显示SQL查询
	ShowSQL bool

	// 是否显示HTTP请求和响应
	ShowHTTP bool

	// 是否启用性能分析
	EnableProfiler bool

	// 性能分析端口
	ProfilerPort int

	// 是否启用调试控制台
	EnableConsole bool

	// 控制台端口
	ConsolePort int
}

// ServerConfig 开发服务器配置
type ServerConfig struct {
	// 服务器地址
	Host string

	// 服务器端口
	Port int

	// 是否启用HTTPS
	EnableHTTPS bool

	// 是否启用自动打开浏览器
	OpenBrowser bool

	// 静态文件目录
	StaticDir string

	// 代理设置
	Proxies map[string]string

	// 自定义中间件
	Middlewares []interface{}

	// 是否启用WebSocket
	EnableWebSocket bool

	// 是否启用跨域
	EnableCORS bool

	// 自定义响应头
	Headers map[string]string
}

// NewConfig 创建新的开发环境配置
func NewConfig() *Config {
	// 获取工作目录
	rootDir, _ := os.Getwd()

	return &Config{
		RootDir: rootDir,
		Reload: ReloadConfig{
			Enabled:           true,
			WatchDirs:         []string{"."},
			IgnorePatterns:    []string{".git", "node_modules", "vendor", "tmp", ".idea", ".vscode"},
			DelayMilliseconds: 500,
			MaxFrequency:      1,
		},
		Debug: DebugConfig{
			VerboseLogging: true,
			ShowRoutes:     true,
			ShowSQL:        true,
			ShowHTTP:       true,
			EnableProfiler: true,
			ProfilerPort:   6060,
			EnableConsole:  true,
			ConsolePort:    8888,
		},
		Server: ServerConfig{
			Host:            "localhost",
			Port:            8080,
			EnableHTTPS:     false,
			OpenBrowser:     true,
			StaticDir:       "public",
			EnableWebSocket: true,
			EnableCORS:      true,
			Headers:         map[string]string{},
			Proxies:         map[string]string{},
		},
		Env:      map[string]string{},
		Settings: map[string]interface{}{},
	}
}

// SetRootDir 设置应用根目录
func (c *Config) SetRootDir(dir string) *Config {
	c.RootDir = dir
	return c
}

// EnableHotReload 启用热重载
func (c *Config) EnableHotReload(enabled bool) *Config {
	c.Reload.Enabled = enabled
	return c
}

// AddWatchDir 添加监视目录
func (c *Config) AddWatchDir(dir string) *Config {
	c.Reload.WatchDirs = append(c.Reload.WatchDirs, dir)
	return c
}

// SetIgnorePatterns 设置忽略的文件模式
func (c *Config) SetIgnorePatterns(patterns ...string) *Config {
	c.Reload.IgnorePatterns = patterns
	return c
}

// SetReloadDelay 设置重载延迟时间
func (c *Config) SetReloadDelay(delay time.Duration) *Config {
	c.Reload.DelayMilliseconds = int(delay.Milliseconds())
	return c
}

// SetReloadCallback 设置重载回调
func (c *Config) SetReloadCallback(callback func()) *Config {
	c.Reload.OnReload = callback
	return c
}

// EnableVerboseLogging 启用详细日志
func (c *Config) EnableVerboseLogging(enabled bool) *Config {
	c.Debug.VerboseLogging = enabled
	return c
}

// EnableShowRoutes 启用显示路由信息
func (c *Config) EnableShowRoutes(enabled bool) *Config {
	c.Debug.ShowRoutes = enabled
	return c
}

// EnableShowSQL 启用显示SQL查询
func (c *Config) EnableShowSQL(enabled bool) *Config {
	c.Debug.ShowSQL = enabled
	return c
}

// EnableShowHTTP 启用显示HTTP请求和响应
func (c *Config) EnableShowHTTP(enabled bool) *Config {
	c.Debug.ShowHTTP = enabled
	return c
}

// EnableProfiler 启用性能分析
func (c *Config) EnableProfiler(enabled bool) *Config {
	c.Debug.EnableProfiler = enabled
	return c
}

// SetProfilerPort 设置性能分析端口
func (c *Config) SetProfilerPort(port int) *Config {
	c.Debug.ProfilerPort = port
	return c
}

// EnableConsole 启用调试控制台
func (c *Config) EnableConsole(enabled bool) *Config {
	c.Debug.EnableConsole = enabled
	return c
}

// SetConsolePort 设置控制台端口
func (c *Config) SetConsolePort(port int) *Config {
	c.Debug.ConsolePort = port
	return c
}

// SetServerHost 设置服务器地址
func (c *Config) SetServerHost(host string) *Config {
	c.Server.Host = host
	return c
}

// SetServerPort 设置服务器端口
func (c *Config) SetServerPort(port int) *Config {
	c.Server.Port = port
	return c
}

// EnableHTTPS 启用HTTPS
func (c *Config) EnableHTTPS(enabled bool) *Config {
	c.Server.EnableHTTPS = enabled
	return c
}

// EnableOpenBrowser 启用自动打开浏览器
func (c *Config) EnableOpenBrowser(enabled bool) *Config {
	c.Server.OpenBrowser = enabled
	return c
}

// SetStaticDir 设置静态文件目录
func (c *Config) SetStaticDir(dir string) *Config {
	c.Server.StaticDir = dir
	return c
}

// AddProxy 添加代理设置
func (c *Config) AddProxy(path, target string) *Config {
	c.Server.Proxies[path] = target
	return c
}

// EnableWebSocket 启用WebSocket
func (c *Config) EnableWebSocket(enabled bool) *Config {
	c.Server.EnableWebSocket = enabled
	return c
}

// EnableCORS 启用跨域
func (c *Config) EnableCORS(enabled bool) *Config {
	c.Server.EnableCORS = enabled
	return c
}

// AddHeader 添加自定义响应头
func (c *Config) AddHeader(key, value string) *Config {
	c.Server.Headers[key] = value
	return c
}

// SetEnv 设置环境变量
func (c *Config) SetEnv(key, value string) *Config {
	c.Env[key] = value
	return c
}

// AddSetting 添加自定义设置
func (c *Config) AddSetting(key string, value interface{}) *Config {
	c.Settings[key] = value
	return c
}

// Load 从配置文件加载配置
func (c *Config) Load(configFile string) error {
	// 配置文件加载逻辑
	// 这里简化处理，实际场景可以使用json、yaml等格式配置文件
	return nil
}

// Save 保存配置到文件
func (c *Config) Save(configFile string) error {
	// 配置保存逻辑
	return nil
}

// String 返回配置的字符串表示
func (c *Config) String() string {
	sb := strings.Builder{}

	sb.WriteString("开发环境配置:\n")
	sb.WriteString(fmt.Sprintf("  根目录: %s\n", c.RootDir))

	sb.WriteString("\n热重载配置:\n")
	sb.WriteString(fmt.Sprintf("  启用: %v\n", c.Reload.Enabled))
	sb.WriteString(fmt.Sprintf("  监视目录: %v\n", c.Reload.WatchDirs))
	sb.WriteString(fmt.Sprintf("  忽略模式: %v\n", c.Reload.IgnorePatterns))
	sb.WriteString(fmt.Sprintf("  延迟时间: %d ms\n", c.Reload.DelayMilliseconds))

	sb.WriteString("\n调试配置:\n")
	sb.WriteString(fmt.Sprintf("  详细日志: %v\n", c.Debug.VerboseLogging))
	sb.WriteString(fmt.Sprintf("  显示路由: %v\n", c.Debug.ShowRoutes))
	sb.WriteString(fmt.Sprintf("  显示SQL: %v\n", c.Debug.ShowSQL))
	sb.WriteString(fmt.Sprintf("  显示HTTP: %v\n", c.Debug.ShowHTTP))
	sb.WriteString(fmt.Sprintf("  性能分析: %v (端口: %d)\n", c.Debug.EnableProfiler, c.Debug.ProfilerPort))
	sb.WriteString(fmt.Sprintf("  调试控制台: %v (端口: %d)\n", c.Debug.EnableConsole, c.Debug.ConsolePort))

	sb.WriteString("\n服务器配置:\n")
	sb.WriteString(fmt.Sprintf("  地址: %s:%d\n", c.Server.Host, c.Server.Port))
	sb.WriteString(fmt.Sprintf("  HTTPS: %v\n", c.Server.EnableHTTPS))
	sb.WriteString(fmt.Sprintf("  自动打开浏览器: %v\n", c.Server.OpenBrowser))
	sb.WriteString(fmt.Sprintf("  静态文件目录: %s\n", c.Server.StaticDir))
	sb.WriteString(fmt.Sprintf("  WebSocket: %v\n", c.Server.EnableWebSocket))
	sb.WriteString(fmt.Sprintf("  CORS: %v\n", c.Server.EnableCORS))

	if len(c.Server.Proxies) > 0 {
		sb.WriteString("  代理设置:\n")
		for path, target := range c.Server.Proxies {
			sb.WriteString(fmt.Sprintf("    %s -> %s\n", path, target))
		}
	}

	if len(c.Server.Headers) > 0 {
		sb.WriteString("  自定义响应头:\n")
		for key, value := range c.Server.Headers {
			sb.WriteString(fmt.Sprintf("    %s: %s\n", key, value))
		}
	}

	return sb.String()
}

// GetWatchDirsAbsolute 获取绝对路径的监视目录列表
func (c *Config) GetWatchDirsAbsolute() []string {
	results := make([]string, 0, len(c.Reload.WatchDirs))

	for _, dir := range c.Reload.WatchDirs {
		if filepath.IsAbs(dir) {
			results = append(results, dir)
		} else {
			absDir, err := filepath.Abs(filepath.Join(c.RootDir, dir))
			if err == nil {
				results = append(results, absDir)
			}
		}
	}

	return results
}

// ShouldIgnoreFile 检查是否应该忽略文件
func (c *Config) ShouldIgnoreFile(path string) bool {
	for _, pattern := range c.Reload.IgnorePatterns {
		matched, err := filepath.Match(pattern, filepath.Base(path))
		if err == nil && matched {
			return true
		}

		if strings.Contains(path, pattern) {
			return true
		}
	}

	return false
}
