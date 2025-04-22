package commands

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/zzliekkas/flow/cli"
)

// NewServerCommand 创建服务器命令
func NewServerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "serve",
		Aliases: []string{"server", "start", "run"},
		Short:   "启动Flow应用服务器",
		Long:    `启动HTTP服务器运行Flow应用程序。可以指定端口、主机和其他服务器设置。`,
		Run:     runServer,
	}

	// 添加命令行标志
	cmd.Flags().StringP("host", "H", "localhost", "设置服务器监听的主机")
	cmd.Flags().IntP("port", "p", 8080, "设置服务器监听的端口")
	cmd.Flags().BoolP("production", "P", false, "在生产模式下运行")
	cmd.Flags().BoolP("watch", "w", false, "启用文件监视模式（仅开发环境）")
	cmd.Flags().StringP("config", "c", "", "指定配置文件路径")

	return cmd
}

// runServer 运行服务器命令
func runServer(cmd *cobra.Command, args []string) {
	// 获取命令行参数
	host, _ := cmd.Flags().GetString("host")
	port, _ := cmd.Flags().GetInt("port")
	production, _ := cmd.Flags().GetBool("production")
	watch, _ := cmd.Flags().GetBool("watch")
	configPath, _ := cmd.Flags().GetString("config")

	// 确定环境模式
	mode := "development"
	if production {
		mode = "production"
		// 在生产模式下禁用监视
		watch = false

		// 设置环境变量确保Flow和Gin使用release模式
		os.Setenv("FLOW_MODE", "release")
		os.Setenv("GIN_MODE", "release")

		// 设置Gin模式为release，禁用调试警告
		gin.SetMode(gin.ReleaseMode)
	} else {
		// 明确设置为release模式以禁用警告
		os.Setenv("GIN_MODE", "release")
		gin.SetMode(gin.ReleaseMode)
	}

	// 显示服务器启动信息
	cli.PrintInfo("启动Flow服务器 [%s模式]", mode)
	cli.PrintInfo("监听: %s:%d", host, port)

	if configPath != "" {
		cli.PrintInfo("使用配置文件: %s", configPath)
	}

	if watch {
		cli.PrintInfo("文件监视模式已启用，修改后将自动重启")
	}

	// 这里将来会集成实际的服务器启动逻辑
	// 目前只是一个占位演示
	fmt.Println("\n服务器已启动并运行...")

	// 监听终止信号
	waitForTermination()
}

// waitForTermination 等待终止信号
func waitForTermination() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	cli.PrintInfo("正在关闭服务器...")
	// 这里将来会添加优雅关闭的逻辑
}
