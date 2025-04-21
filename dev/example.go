package dev

import (
	"fmt"
	"log"
	"os"
	"time"
)

// Example 开发环境示例
func Example() {
	// 创建配置
	config := NewConfig()

	// 配置热重载
	config.EnableHotReload(true)
	config.AddWatchDir("./src")
	config.AddWatchDir("./views")
	config.SetReloadDelay(500 * time.Millisecond)

	// 配置调试选项
	config.EnableVerboseLogging(true)
	config.EnableShowRoutes(true)
	config.EnableShowSQL(true)
	config.EnableShowHTTP(true)

	// 配置服务器
	config.SetServerHost("localhost")
	config.SetServerPort(8080)
	config.SetStaticDir("./public")
	config.EnableOpenBrowser(true)

	// 创建开发服务器
	server, err := NewDevServer(config)
	if err != nil {
		log.Fatalf("创建开发服务器失败: %v\n", err)
	}

	// 启动服务器
	fmt.Println("正在启动开发服务器...")
	if err := server.Start(); err != nil {
		log.Fatalf("启动开发服务器失败: %v\n", err)
	}

	// 等待服务器停止
	defer server.Stop()

	// 保持运行直到收到退出信号
	<-make(chan struct{})
}

// StartWithHotReload 使用热重载启动应用的示例
func StartWithHotReload() {
	// 创建热重载器
	config := NewConfig()
	reloader, err := NewHotReloader(config)
	if err != nil {
		log.Fatalf("创建热重载器失败: %v\n", err)
	}

	// 配置热重载器
	reloader.AddFileExtension(".go", ".html", ".css", ".js")
	reloader.OnReload(func() {
		log.Println("代码已更改，正在重新编译和运行...")
		// 这里可以执行编译和重启等操作
	})

	// 启动热重载器
	if err := reloader.Start(); err != nil {
		log.Fatalf("启动热重载器失败: %v\n", err)
	}
	defer reloader.Stop()

	// 监听重载通道
	reloadCh := reloader.GetReloadChannel()

	// 主循环
	for {
		select {
		case <-reloadCh:
			log.Println("正在重启应用...")
			// 这里可以执行应用重启逻辑
		}
	}
}

// UsageWithDebugLogger 调试日志记录器使用示例
func UsageWithDebugLogger() {
	// 创建调试日志记录器
	logger := NewDebugLogger(nil)

	// 设置日志文件
	err := logger.SetLogFile("./logs/debug.log")
	if err != nil {
		log.Fatalf("设置日志文件失败: %v\n", err)
	}
	defer logger.Close()

	// 记录不同级别的日志
	logger.Debug("这是一条调试日志")
	logger.Info("这是一条信息日志")
	logger.Warn("这是一条警告日志")
	logger.Error("这是一条错误日志")

	// 记录SQL查询
	logger.LogSQL("SELECT * FROM users WHERE id = ?", []interface{}{1}, 10*time.Millisecond, nil)

	// 记录HTTP请求和响应
	logger.LogHTTPRequest("req-001", "GET", "/api/users", nil, nil, "127.0.0.1")
	logger.LogHTTPResponse("req-001", 200, nil, []byte(`{"id":1,"name":"John"}`))

	// 导出调试信息
	logger.DumpDebugInfo(os.Stdout)
}

// ConfigExample 配置示例
func ConfigExample() {
	// 创建默认配置
	config := NewConfig()

	// 自定义配置
	config.SetRootDir("./myapp")

	// 热重载配置
	config.EnableHotReload(true)
	config.AddWatchDir("./src")
	config.SetIgnorePatterns(".git", "node_modules", "tmp")
	config.SetReloadDelay(200 * time.Millisecond)

	// 调试配置
	config.EnableVerboseLogging(true)
	config.EnableShowRoutes(true)
	config.EnableShowSQL(true)
	config.EnableProfiler(true)
	config.SetProfilerPort(6060)

	// 服务器配置
	config.SetServerHost("0.0.0.0") // 绑定所有接口
	config.SetServerPort(3000)
	config.EnableOpenBrowser(false)
	config.SetStaticDir("./public")
	config.EnableCORS(true)

	// 代理配置
	config.AddProxy("/api", "http://localhost:8080")

	// 环境变量
	config.SetEnv("NODE_ENV", "development")

	// 打印配置
	fmt.Println(config.String())
}

// IntegratedExample 集成使用示例
func IntegratedExample() {
	// 创建配置
	config := NewConfig()

	// 创建调试日志记录器
	logger := NewDebugLogger(config)
	logger.SetLogFile("./logs/dev.log")

	// 创建热重载器
	reloader, err := NewHotReloader(config)
	if err != nil {
		log.Fatalf("创建热重载器失败: %v\n", err)
	}

	// 设置热重载回调
	reloader.OnReload(func() {
		logger.Info("检测到文件变更，正在重新加载...")
		// 重新加载逻辑
	})

	// 启动热重载器
	if err := reloader.Start(); err != nil {
		logger.Errorf("启动热重载器失败: %v", err)
		os.Exit(1)
	}

	// 创建开发服务器
	server, err := NewDevServer(config)
	if err != nil {
		logger.Errorf("创建开发服务器失败: %v", err)
		os.Exit(1)
	}

	// 启动服务器
	logger.Info("正在启动开发服务器...")
	if err := server.Start(); err != nil {
		logger.Errorf("启动开发服务器失败: %v", err)
		os.Exit(1)
	}

	// 等待退出信号
	<-make(chan struct{})

	// 清理资源
	reloader.Stop()
	server.Stop()
	logger.Close()
}
