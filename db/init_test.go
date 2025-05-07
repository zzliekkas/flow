package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	// 设置测试模式，避免实际连接数据库
	skipDatabaseConnection = true
}

func TestInitializeDatabase_NestedConfig(t *testing.T) {
	// 测试嵌套配置格式
	nestedConfig := map[string]interface{}{
		"database": map[string]interface{}{
			"default": "mysql",
			"connections": map[string]interface{}{
				"mysql": map[string]interface{}{
					"driver":   "mysql",
					"host":     "localhost",
					"port":     3306,
					"database": "testdb",
					"username": "user",
					"password": "pass",
					"charset":  "utf8mb4",
				},
				"sqlite": map[string]interface{}{
					"driver":   "sqlite3",
					"database": "test.db",
				},
			},
		},
	}

	// 初始化数据库，但不实际连接
	provider, err := InitializeDatabase([]interface{}{nestedConfig})
	assert.NoError(t, err, "初始化数据库配置应该成功")
	assert.NotNil(t, provider, "提供者不应为nil")

	// 验证返回类型
	dbProvider, ok := provider.(*DbProvider)
	assert.True(t, ok, "返回对象应该是*DbProvider类型")
	assert.NotNil(t, dbProvider.Manager, "Manager不应为nil")

	// 检查默认连接
	assert.Equal(t, "mysql", dbProvider.Manager.defaultConnection, "默认连接应该是mysql")

	// 检查配置数量
	assert.Equal(t, 2, len(dbProvider.Manager.configs), "应该有2个连接配置")

	// 验证配置
	mysqlConfig, exists := dbProvider.Manager.configs["mysql"]
	assert.True(t, exists, "mysql配置应该存在")
	assert.Equal(t, "mysql", mysqlConfig.Driver, "mysql驱动应该是mysql")

	sqliteConfig, exists := dbProvider.Manager.configs["sqlite"]
	assert.True(t, exists, "sqlite配置应该存在")
	assert.Equal(t, "sqlite3", sqliteConfig.Driver, "sqlite驱动应该是sqlite3")
}

func TestInitializeDatabase_DirectConfig(t *testing.T) {
	// 测试直接配置格式
	directConfig := map[string]interface{}{
		"default": "default",
		"connections": map[string]interface{}{
			"default": map[string]interface{}{
				"driver":            "mysql",
				"host":              "localhost",
				"port":              3306,
				"database":          "testdb",
				"username":          "user",
				"password":          "pass",
				"charset":           "utf8mb4",
				"max_idle_conns":    10,
				"max_open_conns":    50,
				"conn_max_lifetime": 3600,
			},
		},
	}

	// 初始化数据库
	provider, err := InitializeDatabase([]interface{}{directConfig})
	assert.NoError(t, err, "初始化数据库配置应该成功")
	assert.NotNil(t, provider, "提供者不应为nil")

	// 验证返回类型
	dbProvider, ok := provider.(*DbProvider)
	assert.True(t, ok, "返回对象应该是*DbProvider类型")
	assert.NotNil(t, dbProvider.Manager, "Manager不应为nil")

	// 检查默认连接
	assert.Equal(t, "default", dbProvider.Manager.defaultConnection, "默认连接应该是default")

	// 检查配置数量
	assert.Equal(t, 1, len(dbProvider.Manager.configs), "应该有1个连接配置")

	// 验证配置
	config, exists := dbProvider.Manager.configs["default"]
	assert.True(t, exists, "default配置应该存在")
	assert.Equal(t, "mysql", config.Driver, "驱动应该是mysql")
	assert.Equal(t, 50, config.MaxOpenConns, "最大连接数应该是50")
	assert.Equal(t, 10, config.MaxIdleConns, "最大空闲连接数应该是10")
}

func TestInitializeDatabase_FlatConfig(t *testing.T) {
	// 测试单个连接配置格式
	config := Config{
		Driver:   "postgres",
		Host:     "localhost",
		Port:     5432,
		Database: "testdb",
		Username: "postgres",
		Password: "password",
		SSLMode:  "disable",
	}

	// 初始化数据库
	provider, err := InitializeDatabase([]interface{}{config})
	assert.NoError(t, err, "初始化数据库配置应该成功")
	assert.NotNil(t, provider, "提供者不应为nil")

	// 验证返回类型
	dbProvider, ok := provider.(*DbProvider)
	assert.True(t, ok, "返回对象应该是*DbProvider类型")
	assert.NotNil(t, dbProvider.Manager, "Manager不应为nil")

	// 检查默认连接
	assert.Equal(t, "default", dbProvider.Manager.defaultConnection, "默认连接应该是default")

	// 检查配置数量
	assert.Equal(t, 1, len(dbProvider.Manager.configs), "应该有1个连接配置")

	// 验证配置
	dbConfig, exists := dbProvider.Manager.configs["default"]
	assert.True(t, exists, "default配置应该存在")
	assert.Equal(t, "postgres", dbConfig.Driver, "驱动应该是postgres")
}

func TestCreateConfigFromMap(t *testing.T) {
	// 测试有效的配置
	validConfig := map[string]interface{}{
		"driver":            "mysql",
		"host":              "localhost",
		"port":              3306,
		"database":          "testdb",
		"username":          "user",
		"password":          "pass",
		"charset":           "utf8mb4",
		"max_idle_conns":    20,
		"max_open_conns":    100,
		"conn_max_lifetime": float64(1800),
	}

	config, ok := createConfigFromMap(validConfig)
	assert.True(t, ok, "应该成功创建配置")
	assert.Equal(t, "mysql", config.Driver, "驱动应该是mysql")
	assert.Equal(t, "localhost", config.Host, "主机应该是localhost")
	assert.Equal(t, 3306, config.Port, "端口应该是3306")
	assert.Equal(t, "testdb", config.Database, "数据库名应该是testdb")
	assert.Equal(t, 100, config.MaxOpenConns, "最大连接数应该是100")
	assert.Equal(t, 20, config.MaxIdleConns, "最大空闲连接数应该是20")

	// 测试无效的配置
	invalidConfig := map[string]interface{}{
		"host":     "localhost",
		"database": "testdb",
		// 缺少driver字段
	}

	_, ok = createConfigFromMap(invalidConfig)
	assert.False(t, ok, "缺少driver字段时应该创建失败")

	// 测试非map类型
	_, ok = createConfigFromMap("这不是一个映射")
	assert.False(t, ok, "非map类型时应该创建失败")
}
