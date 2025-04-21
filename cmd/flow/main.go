package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

// 定义版本常量
const version = "0.1.0"

// printBanner 打印Flow框架的ASCII艺术banner
func printBanner() {
	banner := `
┌─┐┬  ┌─┐┬ ┬
├┤ │  │ ││││
└  ┴─┘└─┘└┴┘
`
	color.Cyan(banner)
}

func main() {
	// 打印Flow banner
	printBanner()

	// 打印工具名称和版权信息
	color.Cyan("Flow 框架命令行工具 v%s", version)
	color.White("版权所有 © 2024 Flow 框架团队")
	fmt.Println()

	// 解析命令行参数
	args := os.Args[1:]
	if len(args) == 0 {
		printHelp()
		return
	}

	// 处理命令
	command := strings.ToLower(args[0])
	switch command {
	case "version", "-v", "--version":
		fmt.Printf("版本: %s\n", version)
	case "help", "-h", "--help":
		printHelp()
	default:
		color.Red("未知命令: %s", command)
		fmt.Println()
		printHelp()
		os.Exit(1)
	}
}

// 打印帮助信息
func printHelp() {
	fmt.Println("可用命令:")
	fmt.Println("  version, -v, --version   显示版本信息")
	fmt.Println("  help, -h, --help         显示帮助信息")
}
