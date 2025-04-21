package commands

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zzliekkas/flow/cli"
)

// NewConfigCommand 创建配置管理命令
func NewConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "config",
		Aliases: []string{"conf", "cfg"},
		Short:   "管理应用配置",
		Long:    `查看、修改和重置应用配置信息。`,
	}

	// 添加子命令
	cmd.AddCommand(newConfigGetCommand())
	cmd.AddCommand(newConfigSetCommand())
	cmd.AddCommand(newConfigListCommand())
	cmd.AddCommand(newConfigCacheCommand())
	cmd.AddCommand(newConfigClearCommand())

	return cmd
}

// newConfigGetCommand 配置获取命令
func newConfigGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [key]",
		Short: "获取配置值",
		Long:  `获取指定键的配置值，支持点号表示法访问嵌套配置，如 "app.name"。`,
		Args:  cobra.ExactArgs(1),
		Run:   getConfig,
	}

	cmd.Flags().BoolP("default", "d", false, "显示默认值（如果没有设置）")

	return cmd
}

// newConfigSetCommand 配置设置命令
func newConfigSetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set [key] [value]",
		Short: "设置配置值",
		Long:  `设置指定键的配置值，支持点号表示法设置嵌套配置，如 "app.name"。`,
		Args:  cobra.ExactArgs(2),
		Run:   setConfig,
	}

	cmd.Flags().BoolP("env", "e", false, "同时更新环境变量")

	return cmd
}

// newConfigListCommand 配置列表命令
func newConfigListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "列出所有配置",
		Long:    `列出所有应用配置，可选择按路径筛选。`,
		Run:     listConfig,
	}

	cmd.Flags().StringP("filter", "f", "", "按键名筛选（支持部分匹配）")
	cmd.Flags().BoolP("hide-sensitive", "s", true, "隐藏敏感信息（如密码和密钥）")

	return cmd
}

// newConfigCacheCommand 配置缓存命令
func newConfigCacheCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cache",
		Short: "缓存配置文件",
		Long:  `将当前配置缓存到文件，以加快应用启动速度。`,
		Run:   cacheConfig,
	}

	return cmd
}

// newConfigClearCommand 清除配置缓存命令
func newConfigClearCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "clear",
		Aliases: []string{"clear-cache"},
		Short:   "清除配置缓存",
		Long:    `清除之前缓存的配置文件。`,
		Run:     clearConfigCache,
	}

	return cmd
}

// getConfig 获取配置的实现
func getConfig(cmd *cobra.Command, args []string) {
	key := args[0]
	showDefault, _ := cmd.Flags().GetBool("default")

	// 这里应该连接到实际的配置系统
	// 以下仅为示例实现
	value, exists := getConfigValue(key)

	if !exists {
		if showDefault {
			defaultValue, hasDefault := getDefaultConfigValue(key)
			if hasDefault {
				cli.PrintInfo("配置 '%s' 未设置，默认值为: %v", key, defaultValue)
				return
			}
		}
		cli.PrintError("配置 '%s' 不存在", key)
		os.Exit(1)
	}

	cli.PrintSuccess("配置 '%s' 的值为: %v", key, value)
}

// setConfig 设置配置的实现
func setConfig(cmd *cobra.Command, args []string) {
	key := args[0]
	value := args[1]
	updateEnv, _ := cmd.Flags().GetBool("env")

	// 这里应该连接到实际的配置系统
	// 以下仅为示例实现
	success := setConfigValue(key, value)

	if !success {
		cli.PrintError("无法设置配置 '%s'", key)
		os.Exit(1)
	}

	cli.PrintSuccess("成功设置配置 '%s' 为: %v", key, value)

	if updateEnv {
		// 更新环境变量
		envKey := configKeyToEnvVar(key)
		os.Setenv(envKey, value)
		cli.PrintInfo("已更新环境变量 %s", envKey)
	}
}

// listConfig 列出配置的实现
func listConfig(cmd *cobra.Command, args []string) {
	filter, _ := cmd.Flags().GetString("filter")
	hideSensitive, _ := cmd.Flags().GetBool("hide-sensitive")

	// 这里应该连接到实际的配置系统
	// 以下仅为示例实现
	configs := getAllConfigs()

	// 过滤配置
	var filteredConfigs []configItem
	for _, cfg := range configs {
		if filter != "" && !strings.Contains(strings.ToLower(cfg.Key), strings.ToLower(filter)) {
			continue
		}

		// 隐藏敏感信息
		if hideSensitive && isSensitiveKey(cfg.Key) {
			cfg.Value = "******"
		}

		filteredConfigs = append(filteredConfigs, cfg)
	}

	// 排序
	sort.Slice(filteredConfigs, func(i, j int) bool {
		return filteredConfigs[i].Key < filteredConfigs[j].Key
	})

	if len(filteredConfigs) == 0 {
		cli.PrintInfo("没有找到匹配的配置")
		return
	}

	cli.PrintSuccess("找到 %d 个配置项", len(filteredConfigs))
	fmt.Println()

	// 打印配置
	for _, cfg := range filteredConfigs {
		fmt.Printf("%s: %v\n", cfg.Key, cfg.Value)
	}
}

// cacheConfig 缓存配置的实现
func cacheConfig(cmd *cobra.Command, args []string) {
	// 这里应该连接到实际的配置系统
	// 以下仅为示例实现
	success := createConfigCache()

	if !success {
		cli.PrintError("无法缓存配置")
		os.Exit(1)
	}

	cli.PrintSuccess("配置已成功缓存")
}

// clearConfigCache 清除配置缓存的实现
func clearConfigCache(cmd *cobra.Command, args []string) {
	// 这里应该连接到实际的配置系统
	// 以下仅为示例实现
	success := clearConfigCacheInternal()

	if !success {
		cli.PrintError("无法清除配置缓存")
		os.Exit(1)
	}

	cli.PrintSuccess("配置缓存已成功清除")
}

// clearConfigCacheInternal 内部清除缓存实现
func clearConfigCacheInternal() bool {
	// 在实际应用中，这里应该清除配置缓存
	return true
}

// 配置项结构体
type configItem struct {
	Key   string
	Value interface{}
}

// 以下是辅助函数，实际应用中需要替换为真实实现

// 获取配置值
func getConfigValue(key string) (interface{}, bool) {
	// 示例配置
	configs := map[string]interface{}{
		"app.name":        "Flow Application",
		"app.env":         "development",
		"app.debug":       true,
		"database.driver": "mysql",
		"database.host":   "localhost",
		"database.port":   3306,
		"mail.host":       "smtp.example.com",
		"mail.password":   "secret-password",
	}

	value, exists := configs[key]
	return value, exists
}

// 获取默认配置值
func getDefaultConfigValue(key string) (interface{}, bool) {
	// 示例默认配置
	defaults := map[string]interface{}{
		"app.name":        "Flow App",
		"app.env":         "production",
		"app.debug":       false,
		"database.driver": "sqlite",
		"database.host":   "localhost",
		"database.port":   3306,
	}

	value, exists := defaults[key]
	return value, exists
}

// 设置配置值
func setConfigValue(key string, value interface{}) bool {
	// 在实际应用中，这里应该更新配置
	return true
}

// 获取所有配置
func getAllConfigs() []configItem {
	// 示例配置
	configs := []configItem{
		{Key: "app.name", Value: "Flow Application"},
		{Key: "app.env", Value: "development"},
		{Key: "app.debug", Value: true},
		{Key: "app.url", Value: "http://localhost:8080"},
		{Key: "database.driver", Value: "mysql"},
		{Key: "database.host", Value: "localhost"},
		{Key: "database.port", Value: 3306},
		{Key: "database.username", Value: "root"},
		{Key: "database.password", Value: "password"},
		{Key: "mail.host", Value: "smtp.example.com"},
		{Key: "mail.port", Value: 587},
		{Key: "mail.username", Value: "user@example.com"},
		{Key: "mail.password", Value: "secret-password"},
		{Key: "mail.encryption", Value: "tls"},
	}

	return configs
}

// 判断是否是敏感配置键
func isSensitiveKey(key string) bool {
	sensitivePatterns := []string{
		"password",
		"secret",
		"key",
		"token",
		"auth",
		"credentials",
		"private",
	}

	lowKey := strings.ToLower(key)
	for _, pattern := range sensitivePatterns {
		if strings.Contains(lowKey, pattern) {
			return true
		}
	}

	return false
}

// 配置键转为环境变量
func configKeyToEnvVar(key string) string {
	// 将点号分隔的键转换为下划线分隔的大写键
	// 例如: app.debug -> APP_DEBUG
	envKey := strings.ReplaceAll(strings.ToUpper(key), ".", "_")
	return envKey
}

// 创建配置缓存
func createConfigCache() bool {
	// 在实际应用中，这里应该创建配置缓存
	return true
}
