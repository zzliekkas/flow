package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/zzliekkas/flow/cli"
)

var (
	// 支持的资源类型
	resourceTypes = []string{
		"controller",
		"model",
		"middleware",
		"service",
		"repository",
		"migration",
		"seeder",
		"command",
		"event",
		"listener",
	}
)

// NewMakeCommand 创建资源生成命令
func NewMakeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "make [resource-type] [name]",
		Short: "创建一个新的框架资源",
		Long: `生成各种框架资源代码文件，如控制器、模型、中间件等。
		
支持的资源类型:
  - controller   创建控制器
  - model        创建数据模型
  - middleware   创建HTTP中间件
  - service      创建服务
  - repository   创建仓储
  - migration    创建数据库迁移
  - seeder       创建数据库种子
  - command      创建命令
  - event        创建事件
  - listener     创建事件监听器`,
		Args: cobra.MinimumNArgs(2),
		Run:  executeGenerate,
	}

	// 添加标志
	cmd.Flags().StringP("directory", "d", "", "指定生成资源的目录")
	cmd.Flags().StringP("package", "p", "", "指定资源的包名")
	cmd.Flags().StringArrayP("fields", "f", []string{}, "指定模型字段 (格式: name:type)")
	cmd.Flags().Bool("force", false, "强制覆盖已存在的文件")

	return cmd
}

// executeGenerate 执行资源生成
func executeGenerate(cmd *cobra.Command, args []string) {
	resourceType := strings.ToLower(args[0])
	name := args[1]

	// 验证资源类型
	isValid := false
	for _, t := range resourceTypes {
		if t == resourceType {
			isValid = true
			break
		}
	}

	if !isValid {
		cli.PrintError("不支持的资源类型: %s\n支持的类型: %s", resourceType, strings.Join(resourceTypes, ", "))
		return
	}

	// 获取选项
	directory, _ := cmd.Flags().GetString("directory")
	packageName, _ := cmd.Flags().GetString("package")
	fields, _ := cmd.Flags().GetStringArray("fields")
	force, _ := cmd.Flags().GetBool("force")

	// 设置默认包名
	if packageName == "" {
		packageName = resourceType + "s"
	}

	// 根据资源类型选择适当的生成器
	cli.PrintInfo("生成 %s: %s", resourceType, name)

	// 这里将来会集成实际的生成器逻辑
	// 目前只是展示样例实现
	generateStub(resourceType, name, directory, packageName, fields, force)
}

// generateStub 生成代码存根
func generateStub(resourceType, name, directory, packageName string, fields []string, force bool) {
	// 确定文件路径
	fileName := formatFileName(name, resourceType)

	if directory == "" {
		// 使用默认目录
		switch resourceType {
		case "controller":
			directory = "app/controllers"
		case "model":
			directory = "app/models"
		case "middleware":
			directory = "app/middleware"
		case "service":
			directory = "app/services"
		case "repository":
			directory = "app/repositories"
		case "migration":
			directory = "database/migrations"
		case "seeder":
			directory = "database/seeders"
		case "event":
			directory = "app/events"
		case "listener":
			directory = "app/listeners"
		case "command":
			directory = "app/commands"
		default:
			directory = "app/" + resourceType + "s"
		}
	}

	// 确保目录存在
	if err := os.MkdirAll(directory, 0755); err != nil {
		cli.PrintError("创建目录失败: %v", err)
		return
	}

	// 完整文件路径
	filePath := filepath.Join(directory, fileName)

	// 检查文件是否已存在
	if _, err := os.Stat(filePath); err == nil && !force {
		cli.PrintError("文件已存在: %s\n使用 --force 覆盖", filePath)
		return
	}

	// 根据资源类型获取模板
	tmpl, err := getTemplateForResource(resourceType)
	if err != nil {
		cli.PrintError("获取模板失败: %v", err)
		return
	}

	// 准备模板数据
	data := map[string]interface{}{
		"Name":       name,
		"LowerName":  strings.ToLower(name),
		"CamelName":  toCamelCase(name),
		"PascalName": toPascalCase(name),
		"Package":    packageName,
		"Fields":     parseFields(fields),
	}

	// 创建文件
	file, err := os.Create(filePath)
	if err != nil {
		cli.PrintError("创建文件失败: %v", err)
		return
	}
	defer file.Close()

	// 执行模板
	if err := tmpl.Execute(file, data); err != nil {
		cli.PrintError("生成代码失败: %v", err)
		return
	}

	cli.PrintSuccess("文件已生成: %s", filePath)
}

// formatFileName 格式化文件名
func formatFileName(name, resourceType string) string {
	switch resourceType {
	case "migration":
		// 迁移文件格式: timestamp_create_users_table.go
		timestamp := fmt.Sprintf("%d", 0) // 将来使用实际时间戳
		return fmt.Sprintf("%s_create_%s_table.go", timestamp, strings.ToLower(pluralize(name)))
	default:
		// 默认格式: user_controller.go
		return strings.ToLower(name) + "_" + resourceType + ".go"
	}
}

// getTemplateForResource 获取资源类型的模板
func getTemplateForResource(resourceType string) (*template.Template, error) {
	var templateContent string

	// 这里只是示例模板，将来会从实际文件中加载
	switch resourceType {
	case "controller":
		templateContent = `package {{.Package}}

import (
	"github.com/zzliekkas/flow"
)

// {{.PascalName}}Controller 控制器
type {{.PascalName}}Controller struct{}

// Index 列出所有资源
func (c *{{.PascalName}}Controller) Index(ctx *flow.Context) {
	ctx.JSON(200, flow.H{
		"message": "{{.PascalName}} 列表",
	})
}

// Show 显示特定资源
func (c *{{.PascalName}}Controller) Show(ctx *flow.Context) {
	id := ctx.Param("id")
	ctx.JSON(200, flow.H{
		"message": "显示 {{.PascalName}} " + id,
	})
}

// Create 创建新资源
func (c *{{.PascalName}}Controller) Create(ctx *flow.Context) {
	ctx.JSON(201, flow.H{
		"message": "{{.PascalName}} 已创建",
	})
}

// Update 更新特定资源
func (c *{{.PascalName}}Controller) Update(ctx *flow.Context) {
	id := ctx.Param("id")
	ctx.JSON(200, flow.H{
		"message": "{{.PascalName}} " + id + " 已更新",
	})
}

// Delete 删除特定资源
func (c *{{.PascalName}}Controller) Delete(ctx *flow.Context) {
	id := ctx.Param("id")
	ctx.JSON(200, flow.H{
		"message": "{{.PascalName}} " + id + " 已删除",
	})
}
`
	case "model":
		templateContent = `package {{.Package}}

import (
	"time"
)

// {{.PascalName}} 模型
type {{.PascalName}} struct {
	ID        uint      ` + "`json:\"id\" gorm:\"primaryKey\"`" + `
	CreatedAt time.Time ` + "`json:\"created_at\"`" + `
	UpdatedAt time.Time ` + "`json:\"updated_at\"`" + `
{{range .Fields}}	{{.Name}} {{.Type}} ` + "`json:\"{{.JSON}}\" gorm:\"{{.GORM}}\"`" + `
{{end}}
}

// TableName 设置表名
func ({{.LowerName}} *{{.PascalName}}) TableName() string {
	return "{{.LowerName}}s"
}
`
	default:
		templateContent = `package {{.Package}}

// {{.PascalName}} is a {{.Name}} resource.
type {{.PascalName}} struct {
	// TODO: 实现 {{.Name}} 功能
}
`
	}

	return template.New(resourceType).Parse(templateContent)
}

// parseFields 解析字段定义
func parseFields(fieldDefs []string) []map[string]string {
	var fields []map[string]string

	for _, field := range fieldDefs {
		parts := strings.Split(field, ":")
		if len(parts) != 2 {
			continue
		}

		name := toPascalCase(parts[0])
		fieldType := mapFieldType(parts[1])
		jsonName := strings.ToLower(parts[0])
		gormTag := fmt.Sprintf("column:%s", strings.ToLower(parts[0]))

		fields = append(fields, map[string]string{
			"Name": name,
			"Type": fieldType,
			"JSON": jsonName,
			"GORM": gormTag,
		})
	}

	return fields
}

// mapFieldType 将简单类型映射到Go类型
func mapFieldType(simpleType string) string {
	switch strings.ToLower(simpleType) {
	case "string":
		return "string"
	case "int":
		return "int"
	case "bool":
		return "bool"
	case "float":
		return "float64"
	case "time", "date":
		return "time.Time"
	case "json":
		return "map[string]interface{}"
	default:
		return "string"
	}
}

// toCamelCase 将字符串转换为驼峰命名 (小驼峰)
func toCamelCase(s string) string {
	// 处理下划线、破折号等
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.ReplaceAll(s, "-", " ")

	// 转换为标题格式 (每个单词首字母大写)
	s = strings.Title(strings.ToLower(s))

	// 移除空格
	s = strings.ReplaceAll(s, " ", "")

	// 确保第一个字母小写
	if len(s) > 0 {
		return strings.ToLower(s[:1]) + s[1:]
	}
	return s
}

// toPascalCase 将字符串转换为帕斯卡命名 (大驼峰)
func toPascalCase(s string) string {
	// 处理下划线、破折号等
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.ReplaceAll(s, "-", " ")

	// 转换为标题格式 (每个单词首字母大写)
	s = strings.Title(strings.ToLower(s))

	// 移除空格
	s = strings.ReplaceAll(s, " ", "")

	return s
}

// pluralize 简单的复数化函数
func pluralize(s string) string {
	// 简单的复数规则，未考虑不规则复数形式
	if strings.HasSuffix(s, "y") {
		return s[:len(s)-1] + "ies"
	}
	if strings.HasSuffix(s, "s") || strings.HasSuffix(s, "x") ||
		strings.HasSuffix(s, "z") || strings.HasSuffix(s, "ch") ||
		strings.HasSuffix(s, "sh") {
		return s + "es"
	}
	return s + "s"
}
