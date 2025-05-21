package config

import (
	"fmt"
	"io"
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
	defaultConfig *ConfigManager
	once          sync.Once
)

// ConfigManager 表示配置管理器

type ConfigManager struct {
	viper             *viper.Viper
	configPath        string
	configName        string
	configType        string
	env               string
	loaded            bool
	mu                sync.RWMutex
	onChangeCallbacks []func()
}

// 配置选项函数
type ConfigOption func(*ConfigManager)

// NewConfigManager 创建一个新的配置管理器
func NewConfigManager(options ...ConfigOption) *ConfigManager {
	cfg := &ConfigManager{
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
	return func(c *ConfigManager) {
		c.configPath = path
	}
}

// WithConfigName 设置配置文件名
func WithConfigName(name string) ConfigOption {
	return func(c *ConfigManager) {
		c.configName = name
	}
}

// WithConfigType 设置配置文件类型
func WithConfigType(configType string) ConfigOption {
	return func(c *ConfigManager) {
		c.configType = configType
	}
}

// WithEnvironment 设置环境
func WithEnvironment(env string) ConfigOption {
	return func(c *ConfigManager) {
		c.env = env
	}
}

// Load 加载配置文件
func (c *ConfigManager) Load() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 设置配置文件路径
	if c.configPath != "" {
		c.viper.AddConfigPath(c.configPath)
		if os.Getenv("FLOW_DEBUG") == "true" {
			fmt.Printf("添加配置路径: %s\n", c.configPath)
		}
	} else {
		c.viper.AddConfigPath("./config") // 默认配置目录
		c.viper.AddConfigPath(".")        // 当前目录
	}

	// 设置配置文件名称
	if c.configName != "" {
		c.viper.SetConfigName(c.configName)
		if os.Getenv("FLOW_DEBUG") == "true" {
			fmt.Printf("设置配置文件名: %s\n", c.configName)
		}
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

	// 检查配置文件是否存在并尝试修复
	configType := c.configType
	if configType == "" {
		configType = "yaml"
	}
	configFilePath := filepath.Join(c.configPath, fmt.Sprintf("%s.%s", c.configName, configType))
	if err := c.checkAndFixConfigFile(configFilePath); err != nil && os.Getenv("FLOW_DEBUG") == "true" {
		fmt.Printf("检查配置文件失败: %v\n", err)
	}

	// 尝试加载特定环境的配置文件
	if c.env != "" {
		envConfigName := fmt.Sprintf("%s.%s", c.configName, c.env)
		envConfigPath := filepath.Join(c.configPath, fmt.Sprintf("%s.%s", envConfigName, c.configType))
		if _, err := os.Stat(envConfigPath); err == nil {
			if os.Getenv("FLOW_DEBUG") == "true" {
				fmt.Printf("加载环境特定配置: %s\n", envConfigPath)
			}
			c.viper.SetConfigName(envConfigName)
		}
	}

	// 加载配置
	if err := c.viper.ReadInConfig(); err != nil {
		// 文件不存在，创建默认配置
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if os.Getenv("FLOW_DEBUG") == "true" {
				fmt.Printf("配置文件未找到: %v\n", err)
			}

			// 尝试创建配置目录和文件
			if c.configPath != "" {
				if err := os.MkdirAll(c.configPath, 0755); err == nil {
					defaultConfig := `# Flow应用配置
app:
  name: "flow"
  version: "1.0.0"
  mode: "debug"
  log_level: "info"`

					if err := os.WriteFile(configFilePath, []byte(defaultConfig), 0644); err == nil {
						if os.Getenv("FLOW_DEBUG") == "true" {
							fmt.Printf("已创建默认配置文件: %s\n", configFilePath)
						}
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
		} else {
			// 解析错误，尝试修复
			if os.Getenv("FLOW_DEBUG") == "true" {
				fmt.Printf("配置文件解析错误: %v，尝试修复\n", err)
			}
			if err := c.fixConfigFile(configFilePath); err == nil {
				// 修复成功，重新加载
				if err := c.viper.ReadInConfig(); err == nil {
					c.loaded = true
					c.setupConfigWatch()
					return nil
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

// checkAndFixConfigFile 检查并修复配置文件
func (c *ConfigManager) checkAndFixConfigFile(filePath string) error {
	// 检查文件是否存在
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // 文件不存在，不需要修复
		}
		return err
	}

	// 空文件需要删除并重新创建
	if info.Size() == 0 {
		if err := os.Remove(filePath); err != nil {
			return err
		}
		return nil
	}

	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// 检查文件内容，看是否有BOM标记
	if len(content) >= 3 && content[0] == 0xEF && content[1] == 0xBB && content[2] == 0xBF {
		// 移除BOM标记
		content = content[3:]
		if err := os.WriteFile(filePath, content, info.Mode()); err != nil {
			return err
		}
		if os.Getenv("FLOW_DEBUG") == "true" {
			fmt.Println("已移除配置文件的BOM标记")
		}
	}

	return nil
}

// fixConfigFile 修复配置文件
func (c *ConfigManager) fixConfigFile(filePath string) error {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); err != nil {
		return err
	}

	// 创建备份
	backupFile := filePath + ".bak"
	if err := copyFile(filePath, backupFile); err != nil {
		return err
	}

	// 创建新的配置文件
	defaultConfig := `# Flow应用配置
app:
  name: "flow"
  version: "1.0.0"
  mode: "debug"
  log_level: "info"`

	if err := os.WriteFile(filePath, []byte(defaultConfig), 0644); err != nil {
		return err
	}

	if os.Getenv("FLOW_DEBUG") == "true" {
		fmt.Printf("已修复配置文件: %s，备份保存在: %s\n", filePath, backupFile)
	}
	return nil
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// setupConfigWatch 设置配置文件变更监听
func (c *ConfigManager) setupConfigWatch() {
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
func (c *ConfigManager) OnChange(callback func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onChangeCallbacks = append(c.onChangeCallbacks, callback)
}

// Get 获取指定键的配置值
func (c *ConfigManager) Get(key string) interface{} {
	if c.viper == nil {
		return nil
	}
	return c.viper.Get(key)
}

// GetString 获取字符串配置值
func (c *ConfigManager) GetString(key string) string {
	if c.viper == nil {
		return ""
	}
	return c.viper.GetString(key)
}

// GetInt 获取整数配置值
func (c *ConfigManager) GetInt(key string) int {
	if c.viper == nil {
		return 0
	}
	return c.viper.GetInt(key)
}

// GetBool 获取布尔配置值
func (c *ConfigManager) GetBool(key string) bool {
	if c.viper == nil {
		return false
	}
	return c.viper.GetBool(key)
}

// GetFloat64 获取浮点数配置值
func (c *ConfigManager) GetFloat64(key string) float64 {
	if c.viper == nil {
		return 0
	}
	return c.viper.GetFloat64(key)
}

// GetTime 获取时间配置值
func (c *ConfigManager) GetTime(key string) time.Time {
	if c.viper == nil {
		return time.Time{}
	}
	return c.viper.GetTime(key)
}

// GetDuration 获取时间间隔配置值
func (c *ConfigManager) GetDuration(key string) time.Duration {
	if c.viper == nil {
		return 0
	}
	return c.viper.GetDuration(key)
}

// GetStringSlice 获取字符串切片配置值
func (c *ConfigManager) GetStringSlice(key string) []string {
	if c.viper == nil {
		return []string{}
	}
	return c.viper.GetStringSlice(key)
}

// GetStringMap 获取字符串映射配置值
func (c *ConfigManager) GetStringMap(key string) map[string]interface{} {
	if c.viper == nil {
		return map[string]interface{}{}
	}
	return c.viper.GetStringMap(key)
}

// GetStringMapString 获取字符串映射字符串配置值
func (c *ConfigManager) GetStringMapString(key string) map[string]string {
	if c.viper == nil {
		return map[string]string{}
	}
	return c.viper.GetStringMapString(key)
}

// Unmarshal 将配置解析到结构体
func (c *ConfigManager) Unmarshal(key string, rawVal interface{}) error {
	if c.viper == nil {
		return fmt.Errorf("配置未初始化")
	}
	return c.viper.UnmarshalKey(key, rawVal)
}

// UnmarshalWithOptions 将配置解析到结构体，支持额外选项
func (c *ConfigManager) UnmarshalWithOptions(key string, rawVal interface{}, opts ...viper.DecoderConfigOption) error {
	if c.viper == nil {
		return fmt.Errorf("配置未初始化")
	}
	return c.viper.UnmarshalKey(key, rawVal, opts...)
}

// Set 设置配置值
func (c *ConfigManager) Set(key string, value interface{}) {
	if c.viper == nil {
		c.viper = viper.New()
	}
	c.viper.Set(key, value)
}

// Has 检查是否存在指定键
func (c *ConfigManager) Has(key string) bool {
	if c.viper == nil {
		return false
	}
	return c.viper.IsSet(key)
}

// AllSettings 获取所有配置
func (c *ConfigManager) AllSettings() map[string]interface{} {
	if c.viper == nil {
		return map[string]interface{}{}
	}
	return c.viper.AllSettings()
}

// Sub 获取子配置
func (c *ConfigManager) Sub(key string) *ConfigManager {
	subViper := c.viper.Sub(key)
	if subViper == nil {
		return nil
	}

	return &ConfigManager{
		viper:      subViper,
		configPath: c.configPath,
		configName: c.configName,
		configType: c.configType,
		env:        c.env,
		loaded:     true,
	}
}

// IsLoaded 检查配置是否已加载
func (c *ConfigManager) IsLoaded() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.loaded
}

// Load 加载全局配置
func Load(path string) error {
	once.Do(func() {
		defaultConfig = NewConfigManager(WithConfigPath(path))
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
func Sub(key string) *ConfigManager {
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
		defaultConfig = NewConfigManager()
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
