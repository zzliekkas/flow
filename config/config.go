package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Manager 是配置管理器结构体
type Manager struct {
	v          *viper.Viper
	configPath string
	env        string
}

// NewManager 创建一个新的配置管理器
func NewManager(configPath string) *Manager {
	v := viper.New()
	env := os.Getenv("FLOW_ENV")
	if env == "" {
		env = "development"
	}

	// 加载.env文件（如果存在）
	_ = godotenv.Load()

	return &Manager{
		v:          v,
		configPath: configPath,
		env:        env,
	}
}

// Load 加载配置文件
func (m *Manager) Load(configName string) error {
	m.v.SetConfigName(configName)
	m.v.AddConfigPath(m.configPath)
	m.v.SetConfigType("yaml") // 默认使用yaml格式

	// 尝试加载基础配置
	err := m.v.ReadInConfig()
	if err != nil {
		return err
	}

	// 尝试加载环境特定配置
	envConfigName := configName + "." + m.env
	m.v.SetConfigName(envConfigName)
	_ = m.v.MergeInConfig() // 忽略环境配置不存在的错误

	// 允许环境变量覆盖配置
	m.v.AutomaticEnv()
	m.v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	return nil
}

// LoadAll 加载配置目录中的所有配置文件
func (m *Manager) LoadAll() error {
	files, err := filepath.Glob(filepath.Join(m.configPath, "*.yaml"))
	if err != nil {
		return err
	}

	for _, file := range files {
		configName := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))

		// 跳过环境特定的配置文件，它们将在Load方法中合并
		if strings.Contains(configName, ".") {
			continue
		}

		err := m.Load(configName)
		if err != nil {
			return err
		}
	}

	return nil
}

// Get 获取配置值
func (m *Manager) Get(key string) interface{} {
	return m.v.Get(key)
}

// GetString 获取字符串配置值
func (m *Manager) GetString(key string) string {
	return m.v.GetString(key)
}

// GetInt 获取整数配置值
func (m *Manager) GetInt(key string) int {
	return m.v.GetInt(key)
}

// GetBool 获取布尔配置值
func (m *Manager) GetBool(key string) bool {
	return m.v.GetBool(key)
}

// GetStringSlice 获取字符串切片配置值
func (m *Manager) GetStringSlice(key string) []string {
	return m.v.GetStringSlice(key)
}

// GetStringMap 获取字符串映射配置值
func (m *Manager) GetStringMap(key string) map[string]interface{} {
	return m.v.GetStringMap(key)
}

// Set 设置配置值
func (m *Manager) Set(key string, value interface{}) {
	m.v.Set(key, value)
}

// IsSet 检查配置键是否存在
func (m *Manager) IsSet(key string) bool {
	return m.v.IsSet(key)
}

// AllSettings 获取所有配置
func (m *Manager) AllSettings() map[string]interface{} {
	return m.v.AllSettings()
}

// Viper 获取内部viper实例
func (m *Manager) Viper() *viper.Viper {
	return m.v
}

// Env 获取当前环境
func (m *Manager) Env() string {
	return m.env
}

// IsDevelopment 检查是否为开发环境
func (m *Manager) IsDevelopment() bool {
	return m.env == "development"
}

// IsProduction 检查是否为生产环境
func (m *Manager) IsProduction() bool {
	return m.env == "production"
}

// IsTest 检查是否为测试环境
func (m *Manager) IsTest() bool {
	return m.env == "test"
}
