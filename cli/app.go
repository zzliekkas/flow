package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// 定义Banner常量
const (
	SmallBanner = `
    ┌─┐┬  ┌─┐┬ ┬
    ├┤ │  │ ││││
    └  ┴─┘└─┘└┴┘
    %s - %s
`
)

// App 表示CLI应用程序
type App struct {
	// 应用名称
	Name string

	// 应用版本
	Version string

	// 应用描述
	Description string

	// 根命令
	rootCmd *cobra.Command

	// 命令集合
	commands []*cobra.Command
}

// NewApp 创建一个新的CLI应用程序
func NewApp(name, version, description string) *App {
	app := &App{
		Name:        name,
		Version:     version,
		Description: description,
		commands:    make([]*cobra.Command, 0),
	}

	// 创建根命令
	app.rootCmd = &cobra.Command{
		Use:     app.Name,
		Short:   app.Description,
		Version: app.Version,
	}

	// 添加版本标志
	app.rootCmd.Flags().BoolP("version", "v", false, "显示版本信息")

	return app
}

// AddCommand 添加一个命令到应用程序
func (a *App) AddCommand(cmd *cobra.Command) {
	a.commands = append(a.commands, cmd)
	a.rootCmd.AddCommand(cmd)
}

// Run 运行CLI应用程序
func (a *App) Run() error {
	// 检查环境变量中的banner大小设置
	bannerSize := os.Getenv("FLOW_BANNER_SIZE")
	if bannerSize == "" || bannerSize == "small" {
		// 使用小型Banner
		fmt.Printf(SmallBanner, a.Version, a.Description)
	} else {
		// 如果用户要求不显示banner
		if bannerSize != "none" {
			fmt.Printf(SmallBanner, a.Version, a.Description)
		}
	}

	// 获取可执行文件名称
	executable := filepath.Base(os.Args[0])

	// 处理不同的可执行文件名
	if strings.HasPrefix(a.rootCmd.Use, "flow") && executable != "flow" {
		a.rootCmd.Use = executable
	}

	return a.rootCmd.Execute()
}

// NewFlowCLI 创建默认的Flow CLI应用程序
func NewFlowCLI() *App {
	return NewApp("flow", "1.0.6", "Flow框架命令行工具")
}

// PrintError 打印错误信息并退出
func PrintError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(1)
}

// PrintSuccess 打印成功信息
func PrintSuccess(format string, args ...interface{}) {
	fmt.Printf("✓ "+format+"\n", args...)
}

// PrintInfo 打印信息
func PrintInfo(format string, args ...interface{}) {
	fmt.Printf("→ "+format+"\n", args...)
}

// PrintWarning 打印警告信息
func PrintWarning(format string, args ...interface{}) {
	fmt.Printf("⚠ "+format+"\n", args...)
}
