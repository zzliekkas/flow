package db

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestIntegrationExample 集成测试示例
// 这不是实际执行的测试，只是展示如何使用数据库模块
func ExampleDatabaseIntegration() {
	// 跳过数据库连接，这只是示例
	skipDatabaseConnection = true

	// 方式1: 直接使用数据库配置
	dbConfig := Config{
		Driver:   "mysql",
		Host:     "localhost",
		Port:     3306,
		Database: "my_app",
		Username: "root",
		Password: "password",
		Charset:  "utf8mb4",
	}
	provider, err := InitializeDatabase([]interface{}{dbConfig})
	if err != nil {
		fmt.Println("初始化数据库失败:", err)
		return
	}

	dbProvider := provider.(*DbProvider)
	fmt.Printf("数据库连接成功: %s\n", dbProvider.Manager.defaultConnection)

	// 方式2: 使用嵌套配置格式
	nestedConfig := map[string]interface{}{
		"database": map[string]interface{}{
			"default": "master",
			"connections": map[string]interface{}{
				"master": map[string]interface{}{
					"driver":         "mysql",
					"host":           "master.example.com",
					"port":           3306,
					"database":       "my_app",
					"username":       "root",
					"password":       "secret",
					"charset":        "utf8mb4",
					"max_open_conns": 100,
					"max_idle_conns": 10,
				},
				"slave": map[string]interface{}{
					"driver":         "mysql",
					"host":           "slave.example.com",
					"port":           3306,
					"database":       "my_app",
					"username":       "readonly",
					"password":       "readonly",
					"charset":        "utf8mb4",
					"max_open_conns": 50,
					"max_idle_conns": 5,
				},
			},
		},
	}

	provider, err = InitializeDatabase([]interface{}{nestedConfig})
	if err != nil {
		fmt.Println("初始化嵌套配置数据库失败:", err)
		return
	}

	dbProvider = provider.(*DbProvider)
	fmt.Printf("嵌套配置数据库连接成功: %s\n", dbProvider.Manager.defaultConnection)
	fmt.Printf("配置的连接数: %d\n", len(dbProvider.Manager.configs))

	// 恢复标志，避免影响其他测试
	skipDatabaseConnection = false

	// Output:
	// 数据库连接成功: default
	// 嵌套配置数据库连接成功: master
	// 配置的连接数: 2
}

// TestConfigurationExamples 展示各种配置示例
func TestConfigurationExamples(t *testing.T) {
	// 启用测试模式
	skipDatabaseConnection = true
	defer func() {
		skipDatabaseConnection = false
	}()

	t.Run("使用标准配置", func(t *testing.T) {
		// 创建SQLite内存数据库配置
		sqliteConfig := Config{
			Driver:   "sqlite3",
			Database: ":memory:",
		}

		provider, err := InitializeDatabase([]interface{}{sqliteConfig})
		assert.NoError(t, err)
		assert.NotNil(t, provider)

		dbProvider := provider.(*DbProvider)
		assert.Equal(t, "default", dbProvider.Manager.defaultConnection)
		assert.Equal(t, 1, len(dbProvider.Manager.configs))
	})

	t.Run("多数据库配置", func(t *testing.T) {
		// 添加多个配置
		masterConfig := Config{
			Driver:   "mysql",
			Host:     "master.db",
			Port:     3306,
			Database: "app_db",
			Username: "app_user",
			Password: "app_pass",
		}

		// 创建连接选项
		multipleConnOpt := func(m *Manager) {
			m.Register("master", masterConfig)
			m.Register("slave", Config{
				Driver:   "mysql",
				Host:     "slave.db",
				Port:     3306,
				Database: "app_db_slave",
				Username: "readonly",
				Password: "readonly",
			})
			m.SetDefaultConnection("master")
		}

		provider, err := InitializeDatabase([]interface{}{multipleConnOpt})
		assert.NoError(t, err)
		assert.NotNil(t, provider)

		dbProvider := provider.(*DbProvider)
		assert.Equal(t, "master", dbProvider.Manager.defaultConnection)
		assert.Equal(t, 2, len(dbProvider.Manager.configs))
		assert.Contains(t, dbProvider.Manager.configs, "master")
		assert.Contains(t, dbProvider.Manager.configs, "slave")
	})
}
