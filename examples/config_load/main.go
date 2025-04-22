package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/zzliekkas/flow"
	"github.com/zzliekkas/flow/config"
)

func main() {
	// 打印当前工作目录
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("获取当前工作目录失败: %v", err)
	}
	fmt.Printf("当前工作目录: %s\n", cwd)

	// 创建config目录（如果不存在）
	configDir := filepath.Join(cwd, "config")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			log.Fatalf("创建配置目录失败: %v", err)
		}
		fmt.Printf("已创建配置目录: %s\n", configDir)
	}

	// 创建app.yaml配置文件（如果不存在）
	appYamlPath := filepath.Join(configDir, "app.yaml")
	if _, err := os.Stat(appYamlPath); os.IsNotExist(err) {
		// 创建示例配置内容
		yamlContent := `# Flow应用配置示例
app:
  name: "测试应用"
  version: "0.1.0"
  debug: true
  
server:
  port: 8080
  host: "127.0.0.1"
  
log:
  level: "debug"
  format: "text"
`
		// 写入配置文件
		if err := os.WriteFile(appYamlPath, []byte(yamlContent), 0644); err != nil {
			log.Fatalf("创建配置文件失败: %v", err)
		}
		fmt.Printf("已创建配置文件: %s\n", appYamlPath)
	}

	// 方法1: 直接使用config包加载配置
	fmt.Println("\n方法1: 使用config包直接加载")
	cfg := config.NewConfig(
		config.WithConfigPath(configDir),
		config.WithConfigName("app"),
	)

	if err := cfg.Load(); err != nil {
		log.Printf("警告: 加载配置失败: %v", err)
	} else {
		fmt.Printf("应用名称: %s\n", cfg.GetString("app.name"))
		fmt.Printf("应用版本: %s\n", cfg.GetString("app.version"))
		fmt.Printf("服务器端口: %d\n", cfg.GetInt("server.port"))
		fmt.Printf("调试模式: %v\n", cfg.GetBool("app.debug"))
	}

	// 方法2: 通过Flow框架加载配置
	fmt.Println("\n方法2: 通过Flow框架加载")
	app := flow.New(
		flow.WithConfig(configDir),
	)

	// 使用配置
	app.Invoke(func(cfg *config.Config) {
		fmt.Printf("应用名称: %s\n", cfg.GetString("app.name"))
		fmt.Printf("应用版本: %s\n", cfg.GetString("app.version"))
		fmt.Printf("服务器端口: %d\n", cfg.GetInt("server.port"))
		fmt.Printf("调试模式: %v\n", cfg.GetBool("app.debug"))
	})

	fmt.Println("\n配置加载测试完成")
}
