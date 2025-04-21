package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/zzliekkas/flow/cli"
)

// NewStorageCommand 创建存储管理命令
func NewStorageCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "storage",
		Aliases: []string{"store", "disk"},
		Short:   "管理文件存储",
		Long:    `管理文件存储，包括本地和云存储驱动，执行存储操作和查看信息。`,
	}

	// 添加子命令
	cmd.AddCommand(newStorageListCommand())
	cmd.AddCommand(newStorageInfoCommand())
	cmd.AddCommand(newStorageCopyCommand())
	cmd.AddCommand(newStorageMoveCommand())
	cmd.AddCommand(newStorageDeleteCommand())
	cmd.AddCommand(newStorageMakePublicCommand())
	cmd.AddCommand(newStorageMakePrivateCommand())
	cmd.AddCommand(newStorageUrlCommand())

	return cmd
}

// newStorageListCommand 创建存储列表命令
func newStorageListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list [directory]",
		Aliases: []string{"ls", "files"},
		Short:   "列出存储中的文件",
		Long:    `列出存储中指定目录下的文件和子目录。`,
		Run:     listStorage,
		Args:    cobra.MaximumNArgs(1),
	}

	cmd.Flags().StringP("disk", "d", "", "存储磁盘名称")
	cmd.Flags().BoolP("recursive", "r", false, "递归列出所有子目录中的文件")
	cmd.Flags().StringP("format", "f", "table", "输出格式 (table, json)")
	cmd.Flags().BoolP("human", "H", true, "以人类可读格式显示文件大小")
	cmd.Flags().StringP("sort", "s", "name", "排序方式 (name, size, time)")
	cmd.Flags().BoolP("reverse", "R", false, "反向排序")

	return cmd
}

// newStorageInfoCommand 创建存储信息命令
func newStorageInfoCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "info",
		Aliases: []string{"disks", "status"},
		Short:   "显示存储磁盘信息",
		Long:    `显示所有已配置的存储磁盘及其当前状态。`,
		Run:     showStorageInfo,
	}

	cmd.Flags().StringP("disk", "d", "", "存储磁盘名称")
	cmd.Flags().StringP("format", "f", "table", "输出格式 (table, json)")
	cmd.Flags().BoolP("stats", "s", false, "显示详细统计信息")

	return cmd
}

// newStorageCopyCommand 创建文件复制命令
func newStorageCopyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "copy <source> <destination>",
		Aliases: []string{"cp"},
		Short:   "复制文件",
		Long:    `复制存储中的文件或目录到新位置。`,
		Run:     copyStorage,
		Args:    cobra.ExactArgs(2),
	}

	cmd.Flags().StringP("source-disk", "s", "", "源存储磁盘名称")
	cmd.Flags().StringP("destination-disk", "d", "", "目标存储磁盘名称")
	cmd.Flags().BoolP("recursive", "r", false, "递归复制目录")
	cmd.Flags().BoolP("force", "f", false, "覆盖已存在的文件")

	return cmd
}

// newStorageMoveCommand 创建文件移动命令
func newStorageMoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "move <source> <destination>",
		Aliases: []string{"mv"},
		Short:   "移动文件",
		Long:    `移动存储中的文件或目录到新位置。`,
		Run:     moveStorage,
		Args:    cobra.ExactArgs(2),
	}

	cmd.Flags().StringP("source-disk", "s", "", "源存储磁盘名称")
	cmd.Flags().StringP("destination-disk", "d", "", "目标存储磁盘名称")
	cmd.Flags().BoolP("force", "f", false, "覆盖已存在的文件")

	return cmd
}

// newStorageDeleteCommand 创建文件删除命令
func newStorageDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete <path>",
		Aliases: []string{"rm", "remove"},
		Short:   "删除文件",
		Long:    `删除存储中的文件或目录。`,
		Run:     deleteStorage,
		Args:    cobra.ExactArgs(1),
	}

	cmd.Flags().StringP("disk", "d", "", "存储磁盘名称")
	cmd.Flags().BoolP("recursive", "r", false, "递归删除目录")
	cmd.Flags().BoolP("force", "f", false, "不提示确认直接删除")

	return cmd
}

// newStorageMakePublicCommand 创建设置公开访问命令
func newStorageMakePublicCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "make-public <path>",
		Short: "设置文件为公开访问",
		Long:  `设置存储中的文件或目录为公开可访问。`,
		Run:   makeStoragePublic,
		Args:  cobra.ExactArgs(1),
	}

	cmd.Flags().StringP("disk", "d", "", "存储磁盘名称")
	cmd.Flags().BoolP("recursive", "r", false, "递归处理目录")

	return cmd
}

// newStorageMakePrivateCommand 创建设置私有访问命令
func newStorageMakePrivateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "make-private <path>",
		Short: "设置文件为私有访问",
		Long:  `设置存储中的文件或目录为私有访问。`,
		Run:   makeStoragePrivate,
		Args:  cobra.ExactArgs(1),
	}

	cmd.Flags().StringP("disk", "d", "", "存储磁盘名称")
	cmd.Flags().BoolP("recursive", "r", false, "递归处理目录")

	return cmd
}

// newStorageUrlCommand 创建获取URL命令
func newStorageUrlCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "url <path>",
		Aliases: []string{"link", "share"},
		Short:   "获取文件URL",
		Long:    `获取存储中文件的访问URL，可生成临时URL。`,
		Run:     getStorageUrl,
		Args:    cobra.ExactArgs(1),
	}

	cmd.Flags().StringP("disk", "d", "", "存储磁盘名称")
	cmd.Flags().IntP("expires", "e", 60, "临时URL有效期（分钟）")
	cmd.Flags().BoolP("temporary", "t", false, "生成临时URL")

	return cmd
}

// listStorage 列出存储中的文件
func listStorage(cmd *cobra.Command, args []string) {
	disk, _ := cmd.Flags().GetString("disk")
	recursive, _ := cmd.Flags().GetBool("recursive")
	format, _ := cmd.Flags().GetString("format")
	human, _ := cmd.Flags().GetBool("human")
	sort, _ := cmd.Flags().GetString("sort")
	reverse, _ := cmd.Flags().GetBool("reverse")

	// 确定要列出的目录
	directory := "/"
	if len(args) > 0 {
		directory = args[0]
	}

	// 构建描述
	diskDesc := "默认磁盘"
	if disk != "" {
		diskDesc = fmt.Sprintf("磁盘 '%s'", disk)
	}

	// 显示操作信息
	cli.PrintInfo("列出%s中的文件 (目录: %s)", diskDesc, directory)
	if recursive {
		cli.PrintInfo("递归模式：包括所有子目录")
	}

	// 输出格式信息
	cli.PrintInfo("输出格式：%s", format)

	// 排序信息
	sortOrder := "升序"
	if reverse {
		sortOrder = "降序"
	}
	cli.PrintInfo("按%s%s排序", sort, sortOrder)

	// 在实际应用中，此处应该连接到实际的存储系统
	// 以下是示例输出
	fmt.Println("\n文件列表:")
	fmt.Println("名称\t\t\t大小\t\t修改时间\t\t类型")
	fmt.Println("-------------------------------------------------------------------------------------")

	// 生成一些示例文件
	files := []struct {
		name    string
		size    int64
		modTime time.Time
		isDir   bool
	}{
		{"images", 0, time.Now().Add(-24 * time.Hour), true},
		{"documents", 0, time.Now().Add(-48 * time.Hour), true},
		{"uploads", 0, time.Now().Add(-12 * time.Hour), true},
		{"logo.png", 45920, time.Now().Add(-5 * time.Hour), false},
		{"report.pdf", 1258000, time.Now().Add(-2 * time.Hour), false},
		{"data.json", 8720, time.Now().Add(-30 * time.Minute), false},
		{"config.yaml", 3500, time.Now().Add(-1 * time.Hour), false},
		{"backup.zip", 15680000, time.Now().Add(-3 * time.Hour), false},
	}

	// 显示文件列表
	for _, file := range files {
		sizeStr := fmt.Sprintf("%d B", file.size)
		if human {
			sizeStr = humanReadableSize(file.size)
		}

		fileType := "文件"
		if file.isDir {
			fileType = "目录"
		}

		fmt.Printf("%-20s\t%-10s\t%s\t%s\n",
			file.name,
			sizeStr,
			file.modTime.Format("2006-01-02 15:04:05"),
			fileType)
	}
}

// showStorageInfo 显示存储磁盘信息
func showStorageInfo(cmd *cobra.Command, args []string) {
	disk, _ := cmd.Flags().GetString("disk")
	format, _ := cmd.Flags().GetString("format")
	stats, _ := cmd.Flags().GetBool("stats")

	// 构建描述
	if disk == "" {
		cli.PrintInfo("显示所有存储磁盘的信息")
	} else {
		cli.PrintInfo("显示存储磁盘 '%s' 的信息", disk)
	}

	// 输出格式信息
	cli.PrintInfo("输出格式：%s", format)

	// 在实际应用中，此处应该查询实际的存储系统
	// 以下是示例输出
	fmt.Println("\n存储磁盘信息:")
	fmt.Println("名称\t\t驱动\t\t状态\t\t空间使用\t默认")
	fmt.Println("-------------------------------------------------------------------------------------")

	// 生成一些示例磁盘
	disks := []struct {
		name      string
		driver    string
		status    string
		usage     float64
		total     int64
		isDefault bool
	}{
		{"local", "local", "在线", 45.2, 100 * 1024 * 1024 * 1024, true},
		{"public", "local", "在线", 12.8, 200 * 1024 * 1024 * 1024, false},
		{"s3", "s3", "在线", 22.5, 1000 * 1024 * 1024 * 1024, false},
		{"oss", "oss", "离线", 0, 500 * 1024 * 1024 * 1024, false},
	}

	// 显示磁盘列表
	for _, d := range disks {
		// 如果指定了特定磁盘，只显示该磁盘
		if disk != "" && disk != d.name {
			continue
		}

		usageStr := fmt.Sprintf("%.1f%% (%s/%s)",
			d.usage,
			humanReadableSize(int64(float64(d.total)*d.usage/100.0)),
			humanReadableSize(d.total))

		defaultMark := ""
		if d.isDefault {
			defaultMark = "✓"
		}

		fmt.Printf("%-10s\t%-10s\t%-10s\t%-20s\t%s\n",
			d.name,
			d.driver,
			d.status,
			usageStr,
			defaultMark)
	}

	// 如果请求显示详细统计
	if stats && disk != "" {
		fmt.Println("\n详细统计信息:")
		fmt.Println("-------------------------------------------------------------------------------------")
		fmt.Printf("总文件数: %d\n", 1254)
		fmt.Printf("总目录数: %d\n", 56)
		fmt.Printf("平均文件大小: %s\n", humanReadableSize(256*1024))
		fmt.Printf("最大文件: %s (%s)\n", "backup.zip", humanReadableSize(150*1024*1024))
		fmt.Printf("最后写入: %s\n", time.Now().Add(-15*time.Minute).Format("2006-01-02 15:04:05"))
	}
}

// copyStorage 复制存储文件
func copyStorage(cmd *cobra.Command, args []string) {
	sourceDisk, _ := cmd.Flags().GetString("source-disk")
	destDisk, _ := cmd.Flags().GetString("destination-disk")
	recursive, _ := cmd.Flags().GetBool("recursive")
	force, _ := cmd.Flags().GetBool("force")

	source := args[0]
	destination := args[1]

	// 构建描述
	sourceDesc := fmt.Sprintf("'%s'", source)
	if sourceDisk != "" {
		sourceDesc = fmt.Sprintf("'%s' (磁盘: %s)", source, sourceDisk)
	}

	destDesc := fmt.Sprintf("'%s'", destination)
	if destDisk != "" {
		destDesc = fmt.Sprintf("'%s' (磁盘: %s)", destination, destDisk)
	}

	// 显示操作信息
	cli.PrintInfo("复制 %s 到 %s", sourceDesc, destDesc)
	if recursive {
		cli.PrintInfo("递归模式：复制所有子目录和文件")
	}
	if force {
		cli.PrintInfo("强制模式：覆盖已存在的文件")
	}

	// 在实际应用中，此处应该连接到实际的存储系统
	// 以下是示例实现
	time.Sleep(1 * time.Second)
	cli.PrintSuccess("复制完成：5个文件已复制")
}

// moveStorage 移动存储文件
func moveStorage(cmd *cobra.Command, args []string) {
	sourceDisk, _ := cmd.Flags().GetString("source-disk")
	destDisk, _ := cmd.Flags().GetString("destination-disk")
	force, _ := cmd.Flags().GetBool("force")

	source := args[0]
	destination := args[1]

	// 构建描述
	sourceDesc := fmt.Sprintf("'%s'", source)
	if sourceDisk != "" {
		sourceDesc = fmt.Sprintf("'%s' (磁盘: %s)", source, sourceDisk)
	}

	destDesc := fmt.Sprintf("'%s'", destination)
	if destDisk != "" {
		destDesc = fmt.Sprintf("'%s' (磁盘: %s)", destination, destDisk)
	}

	// 显示操作信息
	cli.PrintInfo("移动 %s 到 %s", sourceDesc, destDesc)
	if force {
		cli.PrintInfo("强制模式：覆盖已存在的文件")
	}

	// 在实际应用中，此处应该连接到实际的存储系统
	// 以下是示例实现
	time.Sleep(800 * time.Millisecond)
	cli.PrintSuccess("移动完成")
}

// deleteStorage 删除存储文件
func deleteStorage(cmd *cobra.Command, args []string) {
	disk, _ := cmd.Flags().GetString("disk")
	recursive, _ := cmd.Flags().GetBool("recursive")
	force, _ := cmd.Flags().GetBool("force")

	path := args[0]

	// 构建描述
	pathDesc := fmt.Sprintf("'%s'", path)
	if disk != "" {
		pathDesc = fmt.Sprintf("'%s' (磁盘: %s)", path, disk)
	}

	// 显示操作信息
	cli.PrintInfo("删除 %s", pathDesc)

	// 如果是目录且没有指定递归
	// 在实际应用中，应该先检查此路径是否为目录
	if strings.HasSuffix(path, "/") && !recursive {
		cli.PrintError("'%s' 是一个目录，请使用 -r 标志递归删除", path)
		return
	}

	// 如果不是强制模式，显示确认提示
	if !force {
		fmt.Printf("确定要删除 %s 吗? (y/n): ", pathDesc)
		var response string
		fmt.Scanln(&response)
		confirmed := strings.ToLower(response) == "y" || strings.ToLower(response) == "yes"

		if !confirmed {
			cli.PrintInfo("操作已取消")
			return
		}
	}

	// 在实际应用中，此处应该连接到实际的存储系统
	// 以下是示例实现
	time.Sleep(500 * time.Millisecond)

	if recursive && strings.HasSuffix(path, "/") {
		cli.PrintSuccess("目录 '%s' 及其内容已删除", path)
	} else {
		cli.PrintSuccess("文件 '%s' 已删除", path)
	}
}

// makeStoragePublic 设置存储为公开访问
func makeStoragePublic(cmd *cobra.Command, args []string) {
	disk, _ := cmd.Flags().GetString("disk")
	recursive, _ := cmd.Flags().GetBool("recursive")

	path := args[0]

	// 构建描述
	pathDesc := fmt.Sprintf("'%s'", path)
	if disk != "" {
		pathDesc = fmt.Sprintf("'%s' (磁盘: %s)", path, disk)
	}

	// 显示操作信息
	cli.PrintInfo("设置 %s 为公开访问", pathDesc)

	if recursive && strings.HasSuffix(path, "/") {
		cli.PrintInfo("递归模式：包括所有子目录和文件")
	}

	// 在实际应用中，此处应该连接到实际的存储系统
	// 以下是示例实现
	time.Sleep(300 * time.Millisecond)

	cli.PrintSuccess("可见性已更新：%s 现在为公开访问", pathDesc)
}

// makeStoragePrivate 设置存储为私有访问
func makeStoragePrivate(cmd *cobra.Command, args []string) {
	disk, _ := cmd.Flags().GetString("disk")
	recursive, _ := cmd.Flags().GetBool("recursive")

	path := args[0]

	// 构建描述
	pathDesc := fmt.Sprintf("'%s'", path)
	if disk != "" {
		pathDesc = fmt.Sprintf("'%s' (磁盘: %s)", path, disk)
	}

	// 显示操作信息
	cli.PrintInfo("设置 %s 为私有访问", pathDesc)

	if recursive && strings.HasSuffix(path, "/") {
		cli.PrintInfo("递归模式：包括所有子目录和文件")
	}

	// 在实际应用中，此处应该连接到实际的存储系统
	// 以下是示例实现
	time.Sleep(300 * time.Millisecond)

	cli.PrintSuccess("可见性已更新：%s 现在为私有访问", pathDesc)
}

// getStorageUrl 获取存储URL
func getStorageUrl(cmd *cobra.Command, args []string) {
	disk, _ := cmd.Flags().GetString("disk")
	expires, _ := cmd.Flags().GetInt("expires")
	temporary, _ := cmd.Flags().GetBool("temporary")

	path := args[0]

	// 构建描述
	pathDesc := fmt.Sprintf("'%s'", path)
	if disk != "" {
		pathDesc = fmt.Sprintf("'%s' (磁盘: %s)", path, disk)
	}

	// 显示操作信息
	if temporary {
		cli.PrintInfo("生成 %s 的临时URL (有效期: %d分钟)", pathDesc, expires)
	} else {
		cli.PrintInfo("获取 %s 的URL", pathDesc)
	}

	// 在实际应用中，此处应该连接到实际的存储系统
	// 以下是示例实现
	time.Sleep(200 * time.Millisecond)

	baseUrl := "https://storage.example.com"
	if temporary {
		token := fmt.Sprintf("%x", time.Now().UnixNano())
		validUntil := time.Now().Add(time.Duration(expires) * time.Minute)

		url := fmt.Sprintf("%s/%s?token=%s&expires=%d",
			baseUrl, path, token, validUntil.Unix())

		cli.PrintSuccess("临时URL: %s", url)
		cli.PrintInfo("URL有效期至: %s", validUntil.Format("2006-01-02 15:04:05"))
	} else {
		url := fmt.Sprintf("%s/%s", baseUrl, path)
		cli.PrintSuccess("URL: %s", url)
	}
}

// humanReadableSize 格式化文件大小为人类可读格式
func humanReadableSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}
