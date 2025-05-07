package db

import (
	"fmt"
	"sync"
)

var (
	// 用于示例和测试的全局状态
	testMode     bool
	testModeLock sync.RWMutex
)

// SetTestMode 设置测试模式状态
// 当测试模式开启时，会跳过实际的数据库连接，使用mock连接
func SetTestMode(enabled bool) {
	testModeLock.Lock()
	defer testModeLock.Unlock()
	testMode = enabled
	skipDatabaseConnection = enabled // 兼容现有代码
}

// IsTestMode 返回当前测试模式状态
func IsTestMode() bool {
	testModeLock.RLock()
	defer testModeLock.RUnlock()
	return testMode
}

// GetConnectionDetails 从配置中获取连接详情，用于测试和调试
func GetConnectionDetails(config Config) map[string]interface{} {
	return map[string]interface{}{
		"driver":            config.Driver,
		"host":              config.Host,
		"port":              config.Port,
		"database":          config.Database,
		"username":          config.Username,
		"password":          "******", // 密码脱敏
		"max_open_conns":    config.MaxOpenConns,
		"max_idle_conns":    config.MaxIdleConns,
		"conn_max_lifetime": config.ConnMaxLifetime,
	}
}

// GetNestedConfigDetails 从嵌套配置中获取所有连接详情
func GetNestedConfigDetails(config map[string]interface{}) (map[string]interface{}, error) {
	// 提取数据库配置
	dbMap, hasNestedStruct := processNestedConfig(config)
	if !hasNestedStruct {
		return nil, fmt.Errorf("未找到嵌套配置结构")
	}

	// 获取默认连接
	defaultConn := "default"
	if defaultVal, ok := dbMap["default"]; ok {
		if strVal, ok := defaultVal.(string); ok {
			defaultConn = strVal
		}
	}

	// 获取连接配置
	result := map[string]interface{}{
		"default": defaultConn,
	}

	// 获取连接配置
	if connsVal, ok := dbMap["connections"]; ok {
		if conns, ok := connsVal.(map[string]interface{}); ok {
			connections := make(map[string]interface{})
			for name, conn := range conns {
				if connMap, ok := conn.(map[string]interface{}); ok {
					// 转换为Config对象以复用连接逻辑
					config, ok := createConfigFromMap(connMap)
					if !ok {
						continue // 跳过无效配置
					}
					connections[name] = GetConnectionDetails(config)
				}
			}
			result["connections"] = connections
		}
	}

	return result, nil
}

// DefaultConnectionName 返回默认连接名称
func (m *Manager) DefaultConnectionName() string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.defaultConnection
}

// ConnectionCount 返回配置的连接数量
func (m *Manager) ConnectionCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.configs)
}

// Configs 返回所有配置的映射
func (m *Manager) Configs() map[string]Config {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// 创建副本避免外部修改
	configsCopy := make(map[string]Config)
	for name, config := range m.configs {
		configsCopy[name] = config
	}

	return configsCopy
}
