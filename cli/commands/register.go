package commands

import "github.com/zzliekkas/flow/cli"

// RegisterCommands 将所有命令注册到CLI应用
func RegisterCommands(app *cli.App) {
	// 服务器命令
	app.AddCommand(NewServerCommand())

	// 数据库命令
	app.AddCommand(NewDBCommand())

	// 代码生成命令
	app.AddCommand(NewMakeCommand())

	// 路由命令
	app.AddCommand(NewRoutesCommand())

	// 配置命令
	app.AddCommand(NewConfigCommand())

	// 缓存命令
	app.AddCommand(NewCacheCommand())

	// 日志命令
	app.AddCommand(NewLogsCommand())

	// 队列命令
	app.AddCommand(NewQueueCommand())

	// 存储命令
	app.AddCommand(NewStorageCommand())

	// 可以在此处添加更多命令
	// app.AddCommand(NewStorageCommand())
	// 等等...
}
