package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/zzliekkas/flow/config"
	"github.com/zzliekkas/flow/db"
)

func main() {
	// 启用调试日志
	os.Setenv("FLOW_DB_DEBUG", "true")

	// 设置日志前缀
	log.SetPrefix("[DB_YAML_EXAMPLE] ")
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	fmt.Println("=== 使用YAML配置数据库示例 ===")

	// 获取配置文件路径
	configPath := filepath.Join("examples", "db_yaml_example", "config.yaml")

	// 确认配置文件存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("配置文件不存在: %s", configPath)
	}

	fmt.Printf("使用配置文件: %s\n", configPath)

	// 单独处理一种方法
	fmt.Println("\n=== 使用手动加载配置方式 ===")
	testManualConfig(configPath)

	fmt.Println("\n所有测试完成!")
}

// 手动加载配置方式
func testManualConfig(configPath string) {
	// 为了测试，禁用实际连接
	db.SetTestMode(true)
	defer db.SetTestMode(false)

	// 使用config包加载配置
	cfg := config.NewConfigManager(
		config.WithConfigPath(filepath.Dir(configPath)),
		config.WithConfigName(strings.TrimSuffix(filepath.Base(configPath), filepath.Ext(configPath))),
	)

	fmt.Println("开始加载配置文件...")
	err := cfg.Load()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	fmt.Println("配置文件加载成功")

	// 检查配置内容
	if dbConfig := cfg.Get("database"); dbConfig != nil {
		fmt.Println("数据库配置存在")

		// 打印配置详情
		if defaultConn := cfg.GetString("database.default"); defaultConn != "" {
			fmt.Printf("默认连接: %s\n", defaultConn)
		}

		if connections := cfg.Get("database.connections"); connections != nil {
			fmt.Println("找到数据库连接配置")
		} else {
			fmt.Println("警告: 未找到数据库连接配置")
		}
	} else {
		fmt.Println("警告: 未找到数据库配置部分")
	}

	// 创建数据库管理器
	fmt.Println("创建数据库管理器...")
	manager := db.NewManager()

	// 从配置加载数据库设置
	fmt.Println("从配置加载数据库设置...")
	if err := manager.FromConfig(cfg); err != nil {
		log.Fatalf("从配置加载数据库设置失败: %v", err)
	}
	fmt.Println("数据库设置加载成功")

	// 验证配置
	fmt.Printf("默认连接: %s\n", manager.DefaultConnectionName())
	fmt.Printf("连接数量: %d\n", manager.ConnectionCount())

	// 打印所有连接名称
	fmt.Println("配置的数据库连接:")
	for name := range manager.Configs() {
		fmt.Printf("  - %s\n", name)
	}

	// 使用配置初始化数据库
	fmt.Println("初始化数据库...")
	dbProvider, err := db.InitializeDatabase([]interface{}{})
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	// 检查数据库提供者
	if provider, ok := dbProvider.(*db.DbProvider); ok {
		fmt.Println("数据库提供者创建成功")

		// 检查Manager是否为nil
		if provider.Manager == nil {
			fmt.Println("警告: 数据库管理器为nil")
		} else {
			fmt.Println("数据库管理器已初始化")
		}

		// 检查DB是否为nil
		if provider.DB == nil {
			fmt.Println("注意: DB连接为nil (这是预期的，因为我们启用了测试模式)")
		} else {
			fmt.Println("DB连接已初始化")
		}
	} else {
		fmt.Println("警告: 转换数据库提供者失败")
	}
}
