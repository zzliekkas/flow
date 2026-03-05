package flow

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zzliekkas/flow/v2/config"
	"github.com/zzliekkas/flow/v2/db"
	"gorm.io/gorm"
)

// Context 是Flow框架的上下文结构体，扩展了Gin的Context
type Context struct {
	*gin.Context
	engine *Engine
}

// Inject 向上下文注入依赖
func (c *Context) Inject(target interface{}) error {
	return c.engine.Invoke(func(injected interface{}) {
		*target.(*interface{}) = injected
	})
}

// DB 获取数据库连接
// 这是一个便捷方法，用于从上下文中获取数据库连接
func (c *Context) DB() *gorm.DB {
	var dbProvider *db.DbProvider
	err := c.engine.Invoke(func(p *db.DbProvider) {
		dbProvider = p
	})

	if err != nil || dbProvider == nil {
		return nil
	}

	return dbProvider.DB
}

// Cache 获取缓存实例
// 这是一个便捷方法，用于从上下文中获取缓存实例
func (c *Context) Cache() interface{} {
	var cache interface{}
	c.Inject(&cache)
	return cache
}

// QueryInt 获取查询参数并转换为整数，如果不存在或转换失败则返回默认值
func (c *Context) QueryInt(key string, defaultValue int) int {
	value := c.Query(key)
	if value == "" {
		return defaultValue
	}

	intValue, err := c.IntParam(value)
	if err != nil {
		return defaultValue
	}

	return intValue
}

// ParamUint 获取URL参数并转换为无符号整数，如果不存在或转换失败则返回0
func (c *Context) ParamUint(key string) uint {
	value := c.Param(key)
	if value == "" {
		return 0
	}

	uintValue, err := c.UintParam(value)
	if err != nil {
		return 0
	}

	return uintValue
}

// IntParam 将字符串转换为整数
func (c *Context) IntParam(value string) (int, error) {
	return strconv.Atoi(value)
}

// UintParam 将字符串转换为无符号整数
func (c *Context) UintParam(value string) (uint, error) {
	val, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(val), nil
}

// Config 获取配置实例
func (c *Context) Config() *config.ConfigManager {
	var cfg *config.ConfigManager
	err := c.engine.Invoke(func(c *config.ConfigManager) {
		cfg = c
	})

	if err != nil || cfg == nil {
		// 如果没有注册配置，返回一个具备安全默认值的空配置
		cfg = config.NewConfigManager()
		// 手动初始化 viper，确保不会发生空指针异常
		cfg.Set("app.name", "flow")
		cfg.Set("app.version", Version)
		cfg.Set("app.mode", c.engine.config.Mode)
		cfg.Set("app.log_level", c.engine.config.LogLevel)
	}

	return cfg
}

// ConfigValue 获取指定键的配置值
func (c *Context) ConfigValue(key string) interface{} {
	cfg := c.Config()
	if cfg == nil {
		return nil
	}
	return cfg.Get(key)
}

// ConfigString 获取字符串配置值
func (c *Context) ConfigString(key string, defaultValue string) string {
	cfg := c.Config()
	if cfg == nil {
		return defaultValue
	}

	value := cfg.GetString(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// ConfigInt 获取整数配置值
func (c *Context) ConfigInt(key string, defaultValue int) int {
	cfg := c.Config()
	if cfg == nil {
		return defaultValue
	}

	return cfg.GetInt(key)
}

// ConfigBool 获取布尔配置值
func (c *Context) ConfigBool(key string, defaultValue bool) bool {
	cfg := c.Config()
	if cfg == nil {
		return defaultValue
	}

	return cfg.GetBool(key)
}
