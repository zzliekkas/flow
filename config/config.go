package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var (
	// 全局配置实例
	defaultConfig *Config
	once          sync.Once
)

// Config 表示配置管理器
type Config struct {
	// viper实例
	viper *viper.Viper

	// 配置文件路径
	configPath string

	// 配置文件名
	configName string

	// 配置文件类型
	configType string

	// 环境
	env string

	// 是否已加载
	loaded bool

	// 锁
	mu sync.RWMutex

	// 配置更改回调
	onChangeCallbacks []func()
}

// 配置选项函数
type ConfigOption func(*Config)

// NewConfig 创建一个新的配置管理器
func NewConfig(options ...ConfigOption) *Config {
	cfg := &Config{
		viper:      viper.New(),
		configPath: "./config",
		configName: "config",
		configType: "yaml",
		env:        "development",
	}

	// 应用选项
	for _, opt := range options {
		opt(cfg)
	}

	return cfg
}

// WithConfigPath 设置配置文件路径
func WithConfigPath(path string) ConfigOption {
	return func(c *Config) {
		c.configPath = path
	}
}

// WithConfigName 设置配置文件名
func WithConfigName(name string) ConfigOption {
	return func(c *Config) {
		c.configName = name
	}
}

// WithConfigType 设置配置文件类型
func WithConfigType(configType string) ConfigOption {
	return func(c *Config) {
		c.configType = configType
	}
}

// WithEnvironment 设置环境
func WithEnvironment(env string) ConfigOption {
	return func(c *Config) {
		c.env = env
	}
}

// Load 加载配置文件
func (c *Config) Load() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 设置配置文件路径
	if c.configPath != "" {
		c.viper.AddConfigPath(c.configPath)
		fmt.Printf("添加配置路径: %s\n", c.configPath)
	} else {
		c.viper.AddConfigPath("./config") // 默认配置目录
		c.viper.AddConfigPath(".")        // 当前目录
	}

	// 设置配置文件名称
	if c.configName != "" {
		c.viper.SetConfigName(c.configName)
		fmt.Printf("设置配置文件名: %s\n", c.configName)
	} else {
		c.viper.SetConfigName("app") // 默认配置文件名
	}

	// 设置配置文件类型
	if c.configType != "" {
		c.viper.SetConfigType(c.configType)
	} else {
		c.viper.SetConfigType("yaml") // 默认使用YAML
	}

	// 加载环境变量
	c.viper.AutomaticEnv()
	c.viper.SetEnvPrefix("FLOW") // 环境变量前缀，如FLOW_APP_NAME
	c.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 尝试加载特定环境的配置文件
	if c.env != "" {
		envConfigName := fmt.Sprintf("%s.%s", c.configName, c.env)
		envConfigPath := filepath.Join(c.configPath, fmt.Sprintf("%s.%s", envConfigName, c.configType))
		if _, err := os.Stat(envConfigPath); err == nil {
			fmt.Printf("加载环境特定配置: %s\n", envConfigPath)
			c.viper.SetConfigName(envConfigName)
		}
	}

	// 加载配置
	if err := c.viper.ReadInConfig(); err != nil {
		// 文件不存在，创建默认配置
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Printf("配置文件未找到: %v\n", err)

			// 尝试创建配置目录和文件
			if c.configPath != "" {
				if err := os.MkdirAll(c.configPath, 0755); err == nil {
					defaultConfig := `app:
  name: "flow"
  version: "1.0.0"
  mode: "debug"
  log_level: "info"`

					configType := c.configType
					if configType == "" {
						configType = "yaml"
					}

					filepath := fmt.Sprintf("%s/%s.%s", c.configPath, c.configName, configType)
					if err := os.WriteFile(filepath, []byte(defaultConfig), 0644); err == nil {
						fmt.Printf("已创建默认配置文件: %s\n", filepath)
						// 重新加载
						if err := c.viper.ReadInConfig(); err == nil {
							c.loaded = true
							// 设置文件变更监听
							c.setupConfigWatch()
							return nil
						}
					}
				}
			}
		}
		return err
	}

	c.loaded = true

	// 设置文件变更监听
	c.setupConfigWatch()

	return nil
}

// setupConfigWatch 设置配置文件变更监听
func (c *Config) setupConfigWatch() {
	c.viper.WatchConfig()
	c.viper.OnConfigChange(func(e fsnotify.Event) {
		c.mu.Lock()
		for _, callback := range c.onChangeCallbacks {
			callback()
		}
		c.mu.Unlock()
	})
}

// OnChange 设置配置变更回调
func (c *Config) OnChange(callback func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onChangeCallbacks = append(c.onChangeCallbacks, callback)
}

// Get 获取指定键的配置值
func (c *Config) Get(key string) interface{} {
	if c.viper == nil {
		return nil
	}
	return c.viper.Get(key)
}

// GetString 获取字符串配置值
func (c *Config) GetString(key string) string {
	if c.viper == nil {
		return ""
	}
	return c.viper.GetString(key)
}

// GetInt 获取整数配置值
func (c *Config) GetInt(key string) int {
	if c.viper == nil {
		return 0
	}
	return c.viper.GetInt(key)
}

// GetBool 获取布尔配置值
func (c *Config) GetBool(key string) bool {
	if c.viper == nil {
		return false
	}
	return c.viper.GetBool(key)
}

// GetFloat64 获取浮点数配置值
func (c *Config) GetFloat64(key string) float64 {
	if c.viper == nil {
		return 0
	}
	return c.viper.GetFloat64(key)
}

// GetTime 获取时间配置值
func (c *Config) GetTime(key string) time.Time {
	if c.viper == nil {
		return time.Time{}
	}
	return c.viper.GetTime(key)
}

// GetDuration 获取时间间隔配置值
func (c *Config) GetDuration(key string) time.Duration {
	if c.viper == nil {
		return 0
	}
	return c.viper.GetDuration(key)
}

// GetStringSlice 获取字符串切片配置值
func (c *Config) GetStringSlice(key string) []string {
	if c.viper == nil {
		return []string{}
	}
	return c.viper.GetStringSlice(key)
}

// GetStringMap 获取字符串映射配置值
func (c *Config) GetStringMap(key string) map[string]interface{} {
	if c.viper == nil {
		return map[string]interface{}{}
	}
	return c.viper.GetStringMap(key)
}

// GetStringMapString 获取字符串映射字符串配置值
func (c *Config) GetStringMapString(key string) map[string]string {
	if c.viper == nil {
		return map[string]string{}
	}
	return c.viper.GetStringMapString(key)
}

// Unmarshal 将配置解析到结构体
func (c *Config) Unmarshal(key string, rawVal interface{}) error {
	if c.viper == nil {
		return fmt.Errorf("配置未初始化")
	}
	return c.viper.UnmarshalKey(key, rawVal)
}

// UnmarshalWithOptions 将配置解析到结构体，支持额外选项
func (c *Config) UnmarshalWithOptions(key string, rawVal interface{}, opts ...viper.DecoderConfigOption) error {
	if c.viper == nil {
		return fmt.Errorf("配置未初始化")
	}
	return c.viper.UnmarshalKey(key, rawVal, opts...)
}

// Set 设置配置值
func (c *Config) Set(key string, value interface{}) {
	if c.viper == nil {
		c.viper = viper.New()
	}
	c.viper.Set(key, value)
}

// Has 检查是否存在指定键
func (c *Config) Has(key string) bool {
	if c.viper == nil {
		return false
	}
	return c.viper.IsSet(key)
}

// AllSettings 获取所有配置
func (c *Config) AllSettings() map[string]interface{} {
	if c.viper == nil {
		return map[string]interface{}{}
	}
	return c.viper.AllSettings()
}

// Sub 获取子配置
func (c *Config) Sub(key string) *Config {
	subViper := c.viper.Sub(key)
	if subViper == nil {
		return nil
	}

	return &Config{
		viper:      subViper,
		configPath: c.configPath,
		configName: c.configName,
		configType: c.configType,
		env:        c.env,
		loaded:     true,
	}
}

// IsLoaded 检查配置是否已加载
func (c *Config) IsLoaded() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.loaded
}

// Load 加载全局配置
func Load(path string) error {
	once.Do(func() {
		defaultConfig = NewConfig(WithConfigPath(path))
	})
	return defaultConfig.Load()
}

// Get 从全局配置中获取值
func Get(key string) interface{} {
	ensureLoaded()
	return defaultConfig.Get(key)
}

// GetString 从全局配置中获取字符串值
func GetString(key string) string {
	ensureLoaded()
	return defaultConfig.GetString(key)
}

// GetInt 从全局配置中获取整数值
func GetInt(key string) int {
	ensureLoaded()
	return defaultConfig.GetInt(key)
}

// GetBool 从全局配置中获取布尔值
func GetBool(key string) bool {
	ensureLoaded()
	return defaultConfig.GetBool(key)
}

// GetFloat64 从全局配置中获取浮点数值
func GetFloat64(key string) float64 {
	ensureLoaded()
	return defaultConfig.GetFloat64(key)
}

// GetTime 从全局配置中获取时间值
func GetTime(key string) time.Time {
	ensureLoaded()
	return defaultConfig.GetTime(key)
}

// GetDuration 从全局配置中获取时间间隔值
func GetDuration(key string) time.Duration {
	ensureLoaded()
	return defaultConfig.GetDuration(key)
}

// GetStringSlice 从全局配置中获取字符串切片值
func GetStringSlice(key string) []string {
	ensureLoaded()
	return defaultConfig.GetStringSlice(key)
}

// GetStringMap 从全局配置中获取字符串映射值
func GetStringMap(key string) map[string]interface{} {
	ensureLoaded()
	return defaultConfig.GetStringMap(key)
}

// GetStringMapString 从全局配置中获取字符串映射字符串值
func GetStringMapString(key string) map[string]string {
	ensureLoaded()
	return defaultConfig.GetStringMapString(key)
}

// Set 设置全局配置值
func Set(key string, value interface{}) {
	ensureLoaded()
	defaultConfig.Set(key, value)
}

// Has 检查全局配置是否存在指定键
func Has(key string) bool {
	ensureLoaded()
	return defaultConfig.Has(key)
}

// OnChange 设置全局配置变更回调
func OnChange(callback func()) {
	ensureLoaded()
	defaultConfig.OnChange(callback)
}

// Unmarshal 将全局配置解析到结构体
func Unmarshal(key string, rawVal interface{}) error {
	ensureLoaded()
	return defaultConfig.Unmarshal(key, rawVal)
}

// UnmarshalWithOptions 将全局配置解析到结构体，支持额外选项
func UnmarshalWithOptions(key string, rawVal interface{}, opts ...viper.DecoderConfigOption) error {
	ensureLoaded()
	return defaultConfig.UnmarshalWithOptions(key, rawVal, opts...)
}

// Sub 获取全局配置的子配置
func Sub(key string) *Config {
	ensureLoaded()
	return defaultConfig.Sub(key)
}

// AllSettings 获取所有全局配置
func AllSettings() map[string]interface{} {
	ensureLoaded()
	return defaultConfig.AllSettings()
}

// ensureLoaded 确保全局配置已加载
func ensureLoaded() {
	if defaultConfig == nil {
		// 如果全局配置为nil，创建一个默认配置而不是抛出panic
		defaultConfig = NewConfig()
		// 设置一些默认值
		defaultConfig.Set("app.name", "flow")
		defaultConfig.Set("app.version", "1.0.0")
		defaultConfig.Set("app.mode", "debug")
		fmt.Println("警告: 配置未初始化，已创建默认配置")
		return
	}

	if !defaultConfig.IsLoaded() {
		// 尝试加载配置，但不因加载失败而panic
		err := defaultConfig.Load()
		if err != nil {
			fmt.Printf("警告: 加载配置失败: %v，将使用默认值\n", err)
			// 确保viper实例已初始化
			if defaultConfig.viper == nil {
				defaultConfig.viper = viper.New()
			}
		}
	}
}
