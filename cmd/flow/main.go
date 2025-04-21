package main

import (
	"os"

	"github.com/zzliekkas/flow/cli"
	"github.com/zzliekkas/flow/cli/commands"
)

func main() {
	// 创建Flow CLI应用
	app := cli.NewFlowCLI()

	// 注册所有命令
	commands.RegisterCommands(app)

	// 运行应用
	if err := app.Run(); err != nil {
		cli.PrintError("执行命令时出错: %v", err)
		os.Exit(1)
	}
}
