package main

import (
	"fmt"
	"log"
	"os"

	"github.com/zzliekkas/flow/db"
)

func main() {
	// 启用调试日志
	os.Setenv("FLOW_DB_DEBUG", "true")

	// 示例1: 使用标准配置
	fmt.Println("\n=== 示例1: 使用标准配置 ===")
	testStandardConfig()

	// 示例2: 使用嵌套配置
	fmt.Println("\n=== 示例2: 使用嵌套配置 ===")
	testNestedConfig()

	// 示例3: 使用函数选项
	fmt.Println("\n=== 示例3: 使用函数选项 ===")
	testFunctionOptions()

	fmt.Println("\n所有测试完成!")
}

// 测试标准配置
func testStandardConfig() {
	// 为了测试，禁用实际连接
	db.SetTestMode(true)
	defer db.SetTestMode(false)

	// 创建SQLite内存数据库配置
	config := db.Config{
		Driver:   "sqlite3",
		Database: ":memory:",
	}

	// 调用InitializeDatabase
	dbProvider, err := db.InitializeDatabase([]interface{}{config})
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	// 检查返回的提供者
	if provider, ok := dbProvider.(*db.DbProvider); ok {
		fmt.Printf("数据库初始化成功!\n")
		fmt.Printf("默认连接: %s\n", provider.Manager.DefaultConnectionName())
		fmt.Printf("连接数量: %d\n", provider.Manager.ConnectionCount())
	} else {
		log.Fatal("数据库初始化失败")
	}
}

// 测试嵌套配置
func testNestedConfig() {
	// 为了测试，禁用实际连接
	db.SetTestMode(true)
	defer db.SetTestMode(false)

	// 嵌套配置
	nestedConfig := map[string]interface{}{
		"database": map[string]interface{}{
			"default": "master",
			"connections": map[string]interface{}{
				"master": map[string]interface{}{
					"driver":         "mysql",
					"host":           "master.db",
					"port":           3306,
					"database":       "app_db",
					"username":       "root",
					"password":       "secret",
					"charset":        "utf8mb4",
					"max_open_conns": 100,
					"max_idle_conns": 10,
				},
				"slave": map[string]interface{}{
					"driver":         "mysql",
					"host":           "slave.db",
					"port":           3306,
					"database":       "app_db",
					"username":       "readonly",
					"password":       "readonly",
					"charset":        "utf8mb4",
					"max_open_conns": 50,
					"max_idle_conns": 5,
				},
			},
		},
	}

	// 调用InitializeDatabase
	dbProvider, err := db.InitializeDatabase([]interface{}{nestedConfig})
	if err != nil {
		log.Fatalf("初始化嵌套配置失败: %v", err)
	}

	// 检查返回的提供者
	if provider, ok := dbProvider.(*db.DbProvider); ok {
		fmt.Printf("嵌套配置初始化成功!\n")
		fmt.Printf("默认连接: %s\n", provider.Manager.DefaultConnectionName())
		fmt.Printf("连接数量: %d\n", provider.Manager.ConnectionCount())

		// 打印所有连接名称
		fmt.Println("配置的数据库连接:")
		for name := range provider.Manager.Configs() {
			fmt.Printf("  - %s\n", name)
		}
	} else {
		log.Fatal("数据库初始化失败")
	}
}

// 测试函数选项
func testFunctionOptions() {
	// 为了测试，禁用实际连接
	db.SetTestMode(true)
	defer db.SetTestMode(false)

	// 创建配置函数
	configFunc := func(m *db.Manager) {
		// 添加主数据库
		m.Register("main", db.Config{
			Driver:   "postgres",
			Host:     "pg.example.com",
			Port:     5432,
			Database: "main_db",
			Username: "postgres",
			Password: "postgres",
			SSLMode:  "disable",
		})

		// 添加分析数据库
		m.Register("analytics", db.Config{
			Driver:   "clickhouse",
			Host:     "ch.example.com",
			Port:     9000,
			Database: "analytics",
			Username: "default",
			Password: "default",
		})

		// 设置默认连接
		m.SetDefaultConnection("main")
	}

	// 调用InitializeDatabase
	dbProvider, err := db.InitializeDatabase([]interface{}{configFunc})
	if err != nil {
		log.Fatalf("初始化函数配置失败: %v", err)
	}

	// 检查返回的提供者
	if provider, ok := dbProvider.(*db.DbProvider); ok {
		fmt.Printf("函数配置初始化成功!\n")
		fmt.Printf("默认连接: %s\n", provider.Manager.DefaultConnectionName())
		fmt.Printf("连接数量: %d\n", provider.Manager.ConnectionCount())

		// 打印所有连接名称
		fmt.Println("配置的数据库连接:")
		for name := range provider.Manager.Configs() {
			fmt.Printf("  - %s\n", name)
		}
	} else {
		log.Fatal("数据库初始化失败")
	}
}
