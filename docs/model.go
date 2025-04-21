package docs

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/zzliekkas/flow/app"
)

// ModelDocGenerator 用于生成数据模型文档的生成器
type ModelDocGenerator struct {
	// 应用实例
	app *app.Application

	// 输出目录
	outputDir string

	// 源码目录
	sourceDir string

	// 模型包路径
	modelPackagePaths []string

	// 标题
	title string

	// 描述
	description string

	// 是否使用Markdown格式
	useMarkdown bool

	// 要忽略的字段
	ignoreFields []string

	// 是否包含私有字段
	includePrivateFields bool

	// 是否包含嵌入字段
	includeEmbeddedFields bool

	// 是否生成ER图
	shouldGenerateERDiagram bool
}

// ModelDoc 表示模型文档
type ModelDoc struct {
	// 文档标题
	Title string `json:"title"`

	// 文档描述
	Description string `json:"description"`

	// 生成时间
	GeneratedAt time.Time `json:"generated_at"`

	// 模型定义
	Models []ModelDefinition `json:"models"`

	// 关系定义
	Relationships []Relationship `json:"relationships,omitempty"`
}

// ModelDefinition 表示模型定义
type ModelDefinition struct {
	// 模型名称
	Name string `json:"name"`

	// 模型描述
	Description string `json:"description"`

	// 模型注释
	Comment string `json:"comment,omitempty"`

	// 模型所在包路径
	Package string `json:"package"`

	// 对应的数据库表（如果有）
	Table string `json:"table,omitempty"`

	// 字段定义
	Fields []FieldDefinition `json:"fields"`
}

// FieldDefinition 表示字段定义
type FieldDefinition struct {
	// 字段名
	Name string `json:"name"`

	// 字段描述
	Description string `json:"description,omitempty"`

	// 字段类型
	Type string `json:"type"`

	// 字段标签
	Tags map[string]string `json:"tags,omitempty"`

	// 是否为主键
	PrimaryKey bool `json:"primary_key,omitempty"`

	// 是否必需
	Required bool `json:"required,omitempty"`

	// 字段默认值
	DefaultValue string `json:"default_value,omitempty"`

	// 是否可为空
	Nullable bool `json:"nullable,omitempty"`

	// 字段注释
	Comment string `json:"comment,omitempty"`

	// 字段验证规则
	ValidationRules string `json:"validation_rules,omitempty"`

	// 字段索引名（如果是索引）
	IndexName string `json:"index_name,omitempty"`

	// 是否唯一
	Unique bool `json:"unique,omitempty"`

	// 关联模型（如果是外键）
	RelatedModel string `json:"related_model,omitempty"`

	// 关联字段（如果是外键）
	RelatedField string `json:"related_field,omitempty"`

	// 字段示例值
	Example string `json:"example,omitempty"`
}

// Relationship 表示模型间关系
type Relationship struct {
	// 源模型
	Source string `json:"source"`

	// 目标模型
	Target string `json:"target"`

	// 关系类型（一对一、一对多、多对多）
	Type string `json:"type"`

	// 源字段
	SourceField string `json:"source_field"`

	// 目标字段
	TargetField string `json:"target_field"`

	// 关系描述
	Description string `json:"description,omitempty"`

	// 中间表（如果是多对多关系）
	JoinTable string `json:"join_table,omitempty"`
}

// NewModelDocGenerator 创建新的模型文档生成器
func NewModelDocGenerator(application *app.Application) *ModelDocGenerator {
	return &ModelDocGenerator{
		app:                     application,
		outputDir:               "./docs/models",
		title:                   "数据模型文档",
		description:             "应用的数据模型和关系定义",
		useMarkdown:             true,
		ignoreFields:            []string{"ID", "CreatedAt", "UpdatedAt", "DeletedAt"},
		includePrivateFields:    false,
		includeEmbeddedFields:   true,
		shouldGenerateERDiagram: true,
	}
}

// SetOutputDir 设置输出目录
func (g *ModelDocGenerator) SetOutputDir(dir string) *ModelDocGenerator {
	g.outputDir = dir
	return g
}

// SetSourceDir 设置源码目录
func (g *ModelDocGenerator) SetSourceDir(dir string) *ModelDocGenerator {
	g.sourceDir = dir
	return g
}

// AddModelPackage 添加模型包路径
func (g *ModelDocGenerator) AddModelPackage(path string) *ModelDocGenerator {
	g.modelPackagePaths = append(g.modelPackagePaths, path)
	return g
}

// SetTitle 设置文档标题
func (g *ModelDocGenerator) SetTitle(title string) *ModelDocGenerator {
	g.title = title
	return g
}

// SetDescription 设置文档描述
func (g *ModelDocGenerator) SetDescription(desc string) *ModelDocGenerator {
	g.description = desc
	return g
}

// UseMarkdown 设置是否使用Markdown格式
func (g *ModelDocGenerator) UseMarkdown(use bool) *ModelDocGenerator {
	g.useMarkdown = use
	return g
}

// IgnoreFields 设置要忽略的字段
func (g *ModelDocGenerator) IgnoreFields(fields ...string) *ModelDocGenerator {
	g.ignoreFields = fields
	return g
}

// IncludePrivateFields 设置是否包含私有字段
func (g *ModelDocGenerator) IncludePrivateFields(include bool) *ModelDocGenerator {
	g.includePrivateFields = include
	return g
}

// IncludeEmbeddedFields 设置是否包含嵌入字段
func (g *ModelDocGenerator) IncludeEmbeddedFields(include bool) *ModelDocGenerator {
	g.includeEmbeddedFields = include
	return g
}

// SetGenerateERDiagram 设置是否生成ER图
func (g *ModelDocGenerator) SetGenerateERDiagram(generate bool) *ModelDocGenerator {
	g.shouldGenerateERDiagram = generate
	return g
}

// Generate 生成模型文档
func (g *ModelDocGenerator) Generate() error {
	// 确保输出目录存在
	if err := os.MkdirAll(g.outputDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	// 解析模型源码
	models, relationships, err := g.parseModels()
	if err != nil {
		return fmt.Errorf("解析模型源码失败: %w", err)
	}

	// 创建文档对象
	doc := ModelDoc{
		Title:         g.title,
		Description:   g.description,
		GeneratedAt:   time.Now(),
		Models:        models,
		Relationships: relationships,
	}

	// 输出JSON文档
	jsonPath := filepath.Join(g.outputDir, "models.json")
	jsonData, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化模型文档失败: %w", err)
	}

	if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
		return fmt.Errorf("写入模型文档失败: %w", err)
	}

	// 生成Markdown文档（如果需要）
	if g.useMarkdown {
		markdownPath := filepath.Join(g.outputDir, "models.md")
		markdownContent, err := g.generateMarkdown(doc)
		if err != nil {
			return fmt.Errorf("生成Markdown文档失败: %w", err)
		}

		if err := os.WriteFile(markdownPath, []byte(markdownContent), 0644); err != nil {
			return fmt.Errorf("写入Markdown文档失败: %w", err)
		}
	}

	// 生成ER图（如果需要）
	if g.shouldGenerateERDiagram {
		if err := g.generateERDiagram(doc); err != nil {
			return fmt.Errorf("生成ER图失败: %w", err)
		}
	}

	fmt.Printf("模型文档已生成: %s\n", g.outputDir)
	return nil
}

// parseModels 解析模型源码
func (g *ModelDocGenerator) parseModels() ([]ModelDefinition, []Relationship, error) {
	var models []ModelDefinition
	var relationships []Relationship

	// 如果没有指定源目录，使用默认目录
	if g.sourceDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, nil, fmt.Errorf("获取当前工作目录失败: %w", err)
		}
		g.sourceDir = wd
	}

	// 如果没有指定模型包路径，使用默认路径
	if len(g.modelPackagePaths) == 0 {
		g.modelPackagePaths = []string{"models", "model", "entity", "entities"}
	}

	fset := token.NewFileSet()

	// 遍历模型包路径
	for _, pkgPath := range g.modelPackagePaths {
		fullPath := filepath.Join(g.sourceDir, pkgPath)
		pkgs, err := parser.ParseDir(fset, fullPath, nil, parser.ParseComments)
		if err != nil {
			// 如果包不存在，继续下一个
			if os.IsNotExist(err) {
				continue
			}
			return nil, nil, fmt.Errorf("解析模型包失败: %w", err)
		}

		// 遍历包
		for pkgName, pkg := range pkgs {
			// 遍历文件
			for _, file := range pkg.Files {
				// 解析模型定义
				fileModels, fileRelationships := g.parseFile(file, pkgName, pkgPath)
				models = append(models, fileModels...)
				relationships = append(relationships, fileRelationships...)
			}
		}
	}

	return models, relationships, nil
}

// parseFile 解析单个文件
func (g *ModelDocGenerator) parseFile(file *ast.File, pkgName, pkgPath string) ([]ModelDefinition, []Relationship) {
	var models []ModelDefinition
	var relationships []Relationship

	// 遍历所有顶级声明
	for _, decl := range file.Decls {
		// 寻找类型声明
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		// 遍历类型规范
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			// 检查是否为结构体
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			// 解析结构体
			model, modelRelationships := g.parseStruct(typeSpec, structType, pkgName, pkgPath, genDecl.Doc)
			models = append(models, model)
			relationships = append(relationships, modelRelationships...)
		}
	}

	return models, relationships
}

// parseStruct 解析结构体
func (g *ModelDocGenerator) parseStruct(typeSpec *ast.TypeSpec, structType *ast.StructType, pkgName, pkgPath string, docGroup *ast.CommentGroup) (ModelDefinition, []Relationship) {
	model := ModelDefinition{
		Name:    typeSpec.Name.Name,
		Package: pkgName,
	}

	// 解析结构体注释
	if docGroup != nil {
		model.Comment = docGroup.Text()
		// 尝试从注释中提取描述
		model.Description = g.extractDescription(docGroup.Text())
	}

	// 尝试从结构体标签中提取表名
	model.Table = g.inferTableName(typeSpec.Name.Name)

	var fields []FieldDefinition
	var relationships []Relationship

	// 遍历字段
	for _, field := range structType.Fields.List {
		// 处理字段名
		var fieldName string
		if len(field.Names) > 0 {
			fieldName = field.Names[0].Name
		} else {
			// 匿名字段使用类型名
			switch t := field.Type.(type) {
			case *ast.Ident:
				fieldName = t.Name
			case *ast.SelectorExpr:
				fieldName = t.Sel.Name
			default:
				continue
			}
		}

		// 是否应该跳过这个字段
		if g.shouldSkipField(fieldName) {
			continue
		}

		// 解析字段定义
		fieldDef, relationship := g.parseField(field, fieldName, model.Name)
		fields = append(fields, fieldDef)

		// 如果存在关系，添加到关系列表
		if relationship.Source != "" {
			relationships = append(relationships, relationship)
		}
	}

	model.Fields = fields
	return model, relationships
}

// parseField 解析字段
func (g *ModelDocGenerator) parseField(field *ast.Field, fieldName, modelName string) (FieldDefinition, Relationship) {
	fieldDef := FieldDefinition{
		Name: fieldName,
	}

	// 初始化空关系
	relationship := Relationship{}

	// 解析字段类型
	fieldDef.Type = g.parseFieldType(field.Type)

	// 解析字段标签
	if field.Tag != nil {
		tag := strings.Trim(field.Tag.Value, "`")
		fieldDef.Tags = g.parseStructTag(tag)

		// 检查是否为主键
		if gormTag, ok := fieldDef.Tags["gorm"]; ok {
			if strings.Contains(gormTag, "primaryKey") {
				fieldDef.PrimaryKey = true
			}
			// 检查是否为外键
			if strings.Contains(gormTag, "foreignKey") {
				relationship.Source = modelName
				relationship.SourceField = fieldName
				// 尝试提取外键信息
				for _, part := range strings.Split(gormTag, ";") {
					if strings.HasPrefix(part, "foreignKey:") {
						fkParts := strings.SplitN(part, ":", 2)
						if len(fkParts) > 1 {
							fieldDef.RelatedField = fkParts[1]
						}
					}
					if strings.HasPrefix(part, "references:") {
						refParts := strings.SplitN(part, ":", 2)
						if len(refParts) > 1 {
							relationship.TargetField = refParts[1]
						}
					}
				}
			}
			// 检查是否有默认值
			if strings.Contains(gormTag, "default:") {
				for _, part := range strings.Split(gormTag, ";") {
					if strings.HasPrefix(part, "default:") {
						defaultParts := strings.SplitN(part, ":", 2)
						if len(defaultParts) > 1 {
							fieldDef.DefaultValue = defaultParts[1]
						}
					}
				}
			}
		}

		// 检查是否必需
		if validTag, ok := fieldDef.Tags["validate"]; ok {
			if strings.Contains(validTag, "required") {
				fieldDef.Required = true
			}
			fieldDef.ValidationRules = validTag
		}

		// 检查是否有描述
		if jsonTag, ok := fieldDef.Tags["json"]; ok {
			parts := strings.Split(jsonTag, ",")
			if len(parts) > 0 && parts[0] != "" && parts[0] != "-" {
				// JSON名可能用作描述的一部分
				if fieldDef.Description == "" {
					fieldDef.Description = "JSON字段: " + parts[0]
				}
			}
		}
	}

	// 解析字段注释
	if field.Comment != nil {
		fieldDef.Comment = field.Comment.Text()
		// 尝试从注释中提取描述
		if fieldDef.Description == "" {
			fieldDef.Description = g.extractDescription(field.Comment.Text())
		}
	}

	// 尝试推断关系类型
	if relationship.Source != "" {
		// 推断目标模型
		// 如果字段类型是指针类型，去掉*
		targetType := fieldDef.Type
		if strings.HasPrefix(targetType, "*") {
			targetType = targetType[1:]
		}
		if strings.HasPrefix(targetType, "[]") {
			// 切片类型表示一对多关系
			targetType = targetType[2:]
			relationship.Type = "one-to-many"
		} else {
			// 单一类型表示一对一关系
			relationship.Type = "one-to-one"
		}
		relationship.Target = targetType

		if relationship.TargetField == "" {
			// 默认目标字段为ID
			relationship.TargetField = "ID"
		}
	}

	return fieldDef, relationship
}

// parseFieldType 解析字段类型
func (g *ModelDocGenerator) parseFieldType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", g.parseFieldType(t.X), t.Sel.Name)
	case *ast.StarExpr:
		return "*" + g.parseFieldType(t.X)
	case *ast.ArrayType:
		return "[]" + g.parseFieldType(t.Elt)
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", g.parseFieldType(t.Key), g.parseFieldType(t.Value))
	case *ast.InterfaceType:
		return "interface{}"
	default:
		return "unknown"
	}
}

// parseStructTag 解析结构体标签
func (g *ModelDocGenerator) parseStructTag(tag string) map[string]string {
	result := make(map[string]string)
	for tag != "" {
		// 跳过前导空格
		i := 0
		for i < len(tag) && tag[i] == ' ' {
			i++
		}
		tag = tag[i:]
		if tag == "" {
			break
		}

		// 扫描到冒号
		i = 0
		for i < len(tag) && tag[i] != ':' {
			i++
		}
		if i >= len(tag) {
			break
		}
		name := tag[:i]
		tag = tag[i+1:]

		// 扫描到下一个空格或引号
		if tag == "" || tag[0] != '"' {
			break
		}
		// 扫描到结束引号
		i = 1
		for i < len(tag) && tag[i] != '"' {
			if tag[i] == '\\' {
				i++
			}
			i++
		}
		if i >= len(tag) {
			break
		}
		value := tag[1:i]
		tag = tag[i+1:]

		result[name] = value
	}
	return result
}

// extractDescription 从注释中提取描述
func (g *ModelDocGenerator) extractDescription(comment string) string {
	if comment == "" {
		return ""
	}

	// 移除注释符号并清理空白
	comment = strings.TrimSpace(comment)
	lines := strings.Split(comment, "\n")
	var result strings.Builder
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// 移除注释前缀
		line = strings.TrimPrefix(line, "//")
		line = strings.TrimPrefix(line, "/*")
		line = strings.TrimSuffix(line, "*/")
		line = strings.TrimSpace(line)
		if line != "" {
			result.WriteString(line)
			result.WriteString(" ")
		}
	}
	return strings.TrimSpace(result.String())
}

// inferTableName 推断表名
func (g *ModelDocGenerator) inferTableName(structName string) string {
	// 默认使用结构体名的蛇形命名
	tableName := ""
	for i, r := range structName {
		if i > 0 && 'A' <= r && r <= 'Z' {
			tableName += "_"
		}
		tableName += strings.ToLower(string(r))
	}
	return tableName + "s" // 加上复数形式
}

// shouldSkipField 判断是否应该跳过字段
func (g *ModelDocGenerator) shouldSkipField(fieldName string) bool {
	// 检查是否在忽略列表
	for _, ignore := range g.ignoreFields {
		if fieldName == ignore {
			return true
		}
	}

	// 检查是否为私有字段
	if !g.includePrivateFields && fieldName != "" && fieldName[0] >= 'a' && fieldName[0] <= 'z' {
		return true
	}

	return false
}

// generateMarkdown 生成Markdown文档
func (g *ModelDocGenerator) generateMarkdown(doc ModelDoc) (string, error) {
	var sb strings.Builder

	// 添加标题和描述
	sb.WriteString(fmt.Sprintf("# %s\n\n", doc.Title))
	sb.WriteString(fmt.Sprintf("%s\n\n", doc.Description))
	sb.WriteString(fmt.Sprintf("生成时间: %s\n\n", doc.GeneratedAt.Format("2006-01-02 15:04:05")))

	// 添加模型列表
	sb.WriteString("## 目录\n\n")
	for _, model := range doc.Models {
		sb.WriteString(fmt.Sprintf("- [%s](#%s)\n", model.Name, strings.ToLower(model.Name)))
	}
	sb.WriteString("\n")

	// 添加模型详情
	for _, model := range doc.Models {
		sb.WriteString(fmt.Sprintf("## %s\n\n", model.Name))

		if model.Description != "" {
			sb.WriteString(fmt.Sprintf("%s\n\n", model.Description))
		}

		if model.Comment != "" {
			sb.WriteString(fmt.Sprintf("注释: %s\n\n", model.Comment))
		}

		if model.Table != "" {
			sb.WriteString(fmt.Sprintf("表名: `%s`\n\n", model.Table))
		}

		// 添加字段表格
		sb.WriteString("### 字段\n\n")
		sb.WriteString("| 字段名 | 类型 | 描述 | 主键 | 必需 | 默认值 | 验证规则 |\n")
		sb.WriteString("|-------|------|------|------|------|--------|----------|\n")

		for _, field := range model.Fields {
			// 格式化字段信息
			primaryKey := ""
			if field.PrimaryKey {
				primaryKey = "✓"
			}

			required := ""
			if field.Required {
				required = "✓"
			}

			// 添加字段行
			sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s | %s | %s |\n",
				field.Name,
				field.Type,
				field.Description,
				primaryKey,
				required,
				field.DefaultValue,
				field.ValidationRules,
			))
		}
		sb.WriteString("\n")
	}

	// 添加关系图
	if len(doc.Relationships) > 0 {
		sb.WriteString("## 模型关系\n\n")
		sb.WriteString("| 源模型 | 关系类型 | 目标模型 | 源字段 | 目标字段 | 描述 |\n")
		sb.WriteString("|-------|----------|---------|-------|---------|------|\n")

		for _, rel := range doc.Relationships {
			sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s | %s |\n",
				rel.Source,
				rel.Type,
				rel.Target,
				rel.SourceField,
				rel.TargetField,
				rel.Description,
			))
		}
		sb.WriteString("\n")
	}

	// 如果生成了ER图，添加链接
	if g.shouldGenerateERDiagram {
		sb.WriteString("## ER图\n\n")
		sb.WriteString("![ER图](./er-diagram.png)\n\n")
	}

	return sb.String(), nil
}

// generateERDiagram 生成ER图
func (g *ModelDocGenerator) generateERDiagram(doc ModelDoc) error {
	// 生成Mermaid图表定义
	var sb strings.Builder
	sb.WriteString("```mermaid\nerDiagram\n")

	// 添加实体
	for _, model := range doc.Models {
		sb.WriteString(fmt.Sprintf("    %s {\n", model.Name))
		for _, field := range model.Fields {
			fieldType := field.Type
			// 简化字段类型显示
			if strings.Contains(fieldType, ".") {
				parts := strings.Split(fieldType, ".")
				fieldType = parts[len(parts)-1]
			}
			pk := ""
			if field.PrimaryKey {
				pk = " PK"
			}
			sb.WriteString(fmt.Sprintf("        %s %s%s\n", fieldType, field.Name, pk))
		}
		sb.WriteString("    }\n")
	}

	// 添加关系
	for _, rel := range doc.Relationships {
		var relType string
		switch rel.Type {
		case "one-to-one":
			relType = "||--||"
		case "one-to-many":
			relType = "||--o{"
		case "many-to-many":
			relType = "}o--o{"
		default:
			relType = "||--o{"
		}
		sb.WriteString(fmt.Sprintf("    %s %s %s : \"%s\"\n",
			rel.Source, relType, rel.Target,
			fmt.Sprintf("%s -> %s", rel.SourceField, rel.TargetField)))
	}

	sb.WriteString("```\n")

	// 输出到Markdown文件
	diagramPath := filepath.Join(g.outputDir, "er-diagram.md")
	if err := os.WriteFile(diagramPath, []byte(sb.String()), 0644); err != nil {
		return fmt.Errorf("写入ER图Markdown文件失败: %w", err)
	}

	// 注意：此处仅生成了Mermaid格式的ER图文本
	// 实际渲染成PNG需要外部工具，如:
	// - 使用Mermaid CLI
	// - 集成到文档生成系统
	// - 使用在线工具转换

	return nil
}
