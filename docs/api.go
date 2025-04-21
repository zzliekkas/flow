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

// APIDocGenerator 用于生成API文档的生成器
type APIDocGenerator struct {
	// 应用实例
	app *app.Application

	// 输出目录
	outputDir string

	// 基础URL
	baseURL string

	// API版本
	apiVersion string

	// 项目标题
	title string

	// 项目描述
	description string

	// 作者信息
	author string

	// 联系邮箱
	email string

	// 许可证信息
	license string

	// 源代码目录
	sourceDir string

	// 要解析的目录名称列表
	includeDirs []string

	// 要排除的目录名称列表
	excludeDirs []string

	// 要解析的文件扩展名
	fileExtensions []string

	// 是否包含内部方法（以_开头的方法）
	includeInternal bool

	// 是否包含私有方法（小写开头的方法）
	includePrivate bool

	// 是否使用Markdown格式
	useMarkdown bool

	// 路由前缀
	routePrefix string
}

// APIEndpoint 表示API端点信息
type APIEndpoint struct {
	// 端点路径
	Path string `json:"path"`

	// HTTP方法
	Method string `json:"method"`

	// 端点描述
	Description string `json:"description"`

	// 处理器函数名称
	Handler string `json:"handler"`

	// 请求参数
	RequestParams []APIParam `json:"request_params,omitempty"`

	// 请求体
	RequestBody interface{} `json:"request_body,omitempty"`

	// 响应体
	ResponseBody interface{} `json:"response_body,omitempty"`

	// 可能的响应状态码
	StatusCodes []APIStatusCode `json:"status_codes,omitempty"`

	// 中间件
	Middleware []string `json:"middleware,omitempty"`

	// 分组名
	Group string `json:"group,omitempty"`

	// 端点标签（用于分类）
	Tags []string `json:"tags,omitempty"`

	// 是否废弃
	Deprecated bool `json:"deprecated,omitempty"`

	// 废弃说明
	DeprecationMessage string `json:"deprecation_message,omitempty"`

	// 示例请求
	Examples []APIExample `json:"examples,omitempty"`
}

// APIParam 表示API参数信息
type APIParam struct {
	// 参数名
	Name string `json:"name"`

	// 参数类型
	Type string `json:"type"`

	// 参数描述
	Description string `json:"description"`

	// 是否必需
	Required bool `json:"required"`

	// 默认值
	DefaultValue string `json:"default_value,omitempty"`

	// 参数位置（路径、查询、头部、正文等）
	Location string `json:"location"`

	// 验证规则
	ValidationRules string `json:"validation_rules,omitempty"`

	// 示例值
	Example string `json:"example,omitempty"`
}

// APIStatusCode 表示API响应状态码信息
type APIStatusCode struct {
	// HTTP状态码
	Code int `json:"code"`

	// 状态描述
	Description string `json:"description"`

	// 示例响应
	Example interface{} `json:"example,omitempty"`
}

// APIExample 表示API示例
type APIExample struct {
	// 示例名称
	Name string `json:"name"`

	// 示例请求
	Request interface{} `json:"request"`

	// 示例响应
	Response interface{} `json:"response"`

	// 示例描述
	Description string `json:"description,omitempty"`
}

// APIDocumentation 表示整体API文档
type APIDocumentation struct {
	// 文档标题
	Title string `json:"title"`

	// 文档描述
	Description string `json:"description"`

	// API版本
	Version string `json:"version"`

	// 基础URL
	BaseURL string `json:"base_url"`

	// 生成时间
	GeneratedAt time.Time `json:"generated_at"`

	// 作者
	Author string `json:"author,omitempty"`

	// 联系邮箱
	Email string `json:"email,omitempty"`

	// 许可证
	License string `json:"license,omitempty"`

	// API端点列表
	Endpoints []APIEndpoint `json:"endpoints"`

	// 模型定义
	Models map[string]interface{} `json:"models,omitempty"`
}

// NewAPIDocGenerator 创建新的API文档生成器
func NewAPIDocGenerator(application *app.Application) *APIDocGenerator {
	return &APIDocGenerator{
		app:            application,
		outputDir:      "./docs/api",
		apiVersion:     "v1",
		title:          "API Documentation",
		description:    "REST API Documentation",
		fileExtensions: []string{".go"},
		routePrefix:    "/api",
	}
}

// SetOutputDir 设置输出目录
func (g *APIDocGenerator) SetOutputDir(dir string) *APIDocGenerator {
	g.outputDir = dir
	return g
}

// SetBaseURL 设置基础URL
func (g *APIDocGenerator) SetBaseURL(url string) *APIDocGenerator {
	g.baseURL = url
	return g
}

// SetAPIVersion 设置API版本
func (g *APIDocGenerator) SetAPIVersion(version string) *APIDocGenerator {
	g.apiVersion = version
	return g
}

// SetTitle 设置文档标题
func (g *APIDocGenerator) SetTitle(title string) *APIDocGenerator {
	g.title = title
	return g
}

// SetDescription 设置文档描述
func (g *APIDocGenerator) SetDescription(desc string) *APIDocGenerator {
	g.description = desc
	return g
}

// SetAuthor 设置作者信息
func (g *APIDocGenerator) SetAuthor(author string) *APIDocGenerator {
	g.author = author
	return g
}

// SetEmail 设置联系邮箱
func (g *APIDocGenerator) SetEmail(email string) *APIDocGenerator {
	g.email = email
	return g
}

// SetLicense 设置许可证信息
func (g *APIDocGenerator) SetLicense(license string) *APIDocGenerator {
	g.license = license
	return g
}

// SetSourceDir 设置源代码目录
func (g *APIDocGenerator) SetSourceDir(dir string) *APIDocGenerator {
	g.sourceDir = dir
	return g
}

// IncludeDirectories 设置要包含的目录
func (g *APIDocGenerator) IncludeDirectories(dirs ...string) *APIDocGenerator {
	g.includeDirs = dirs
	return g
}

// ExcludeDirectories 设置要排除的目录
func (g *APIDocGenerator) ExcludeDirectories(dirs ...string) *APIDocGenerator {
	g.excludeDirs = dirs
	return g
}

// SetFileExtensions 设置要解析的文件扩展名
func (g *APIDocGenerator) SetFileExtensions(exts ...string) *APIDocGenerator {
	g.fileExtensions = exts
	return g
}

// IncludeInternalMethods 是否包含内部方法
func (g *APIDocGenerator) IncludeInternalMethods(include bool) *APIDocGenerator {
	g.includeInternal = include
	return g
}

// IncludePrivateMethods 是否包含私有方法
func (g *APIDocGenerator) IncludePrivateMethods(include bool) *APIDocGenerator {
	g.includePrivate = include
	return g
}

// UseMarkdown 是否使用Markdown格式
func (g *APIDocGenerator) UseMarkdown(use bool) *APIDocGenerator {
	g.useMarkdown = use
	return g
}

// SetRoutePrefix 设置路由前缀
func (g *APIDocGenerator) SetRoutePrefix(prefix string) *APIDocGenerator {
	g.routePrefix = prefix
	return g
}

// Generate 生成API文档
func (g *APIDocGenerator) Generate() error {
	// 确保输出目录存在
	if err := os.MkdirAll(g.outputDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	// 如果未设置源目录，默认使用当前目录
	if g.sourceDir == "" {
		g.sourceDir = "."
	}

	// 搜集路由信息
	endpoints, err := g.collectRoutes()
	if err != nil {
		return fmt.Errorf("收集路由信息失败: %w", err)
	}

	// 解析源代码获取更多API信息
	if err := g.parseSourceCode(endpoints); err != nil {
		return fmt.Errorf("解析源代码失败: %w", err)
	}

	// 提取模型定义
	models, err := g.extractModels()
	if err != nil {
		return fmt.Errorf("提取模型定义失败: %w", err)
	}

	// 生成文档
	doc := APIDocumentation{
		Title:       g.title,
		Description: g.description,
		Version:     g.apiVersion,
		BaseURL:     g.baseURL,
		GeneratedAt: time.Now(),
		Author:      g.author,
		Email:       g.email,
		License:     g.license,
		Endpoints:   endpoints,
		Models:      models,
	}

	// 将文档写入文件
	if err := g.writeDocumentation(doc); err != nil {
		return fmt.Errorf("写入文档失败: %w", err)
	}

	return nil
}

// collectRoutes 收集应用的路由信息
func (g *APIDocGenerator) collectRoutes() ([]APIEndpoint, error) {
	endpoints := []APIEndpoint{}

	// TODO: 这里需要根据实际的路由获取逻辑来实现
	// 目前使用简单的示例数据

	// 示例端点
	endpoints = append(endpoints, APIEndpoint{
		Path:        "/api/users",
		Method:      "GET",
		Description: "获取用户列表",
		Handler:     "GetUsers",
		Group:       "用户管理",
		Tags:        []string{"users", "admin"},
		StatusCodes: []APIStatusCode{
			{Code: 200, Description: "成功获取用户列表"},
			{Code: 401, Description: "未授权"},
			{Code: 500, Description: "服务器内部错误"},
		},
		Middleware: []string{"auth", "logging"},
	})

	endpoints = append(endpoints, APIEndpoint{
		Path:        "/api/users/{id}",
		Method:      "GET",
		Description: "获取单个用户",
		Handler:     "GetUser",
		Group:       "用户管理",
		Tags:        []string{"users", "admin"},
		RequestParams: []APIParam{
			{
				Name:        "id",
				Type:        "string",
				Description: "用户ID",
				Required:    true,
				Location:    "path",
			},
		},
		StatusCodes: []APIStatusCode{
			{Code: 200, Description: "成功获取用户信息"},
			{Code: 404, Description: "用户不存在"},
			{Code: 500, Description: "服务器内部错误"},
		},
	})

	// 可以添加更多示例端点...

	return endpoints, nil
}

// parseSourceCode 解析源代码以获取更多API信息
func (g *APIDocGenerator) parseSourceCode(endpoints []APIEndpoint) error {
	// 创建一个文件集合
	fset := token.NewFileSet()

	// 遍历源代码目录
	err := filepath.Walk(g.sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			// 检查是否应该排除该目录
			dir := filepath.Base(path)
			for _, excludeDir := range g.excludeDirs {
				if dir == excludeDir {
					return filepath.SkipDir
				}
			}

			// 如果指定了includeDirs，检查是否应该包含该目录
			if len(g.includeDirs) > 0 {
				included := false
				for _, includeDir := range g.includeDirs {
					if dir == includeDir {
						included = true
						break
					}
				}
				if !included && path != g.sourceDir {
					return filepath.SkipDir
				}
			}

			return nil
		}

		// 检查文件扩展名
		ext := filepath.Ext(path)
		validExt := false
		for _, fileExt := range g.fileExtensions {
			if ext == fileExt {
				validExt = true
				break
			}
		}
		if !validExt {
			return nil
		}

		// 解析Go源文件
		f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			// 跳过解析错误，不中断整个过程
			fmt.Printf("警告: 解析文件 %s 时出错: %v\n", path, err)
			return nil
		}

		// 更新端点信息
		g.updateEndpointsFromFile(f, endpoints)

		return nil
	})

	return err
}

// updateEndpointsFromFile 从解析的源文件中更新端点信息
func (g *APIDocGenerator) updateEndpointsFromFile(f *ast.File, endpoints []APIEndpoint) {
	// 遍历所有声明
	for _, decl := range f.Decls {
		// 只关注函数声明
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		// 获取函数名
		funcName := funcDecl.Name.Name

		// 检查是否应该包含该函数
		if !g.shouldIncludeFunction(funcName) {
			continue
		}

		// 找到对应的端点
		for i, endpoint := range endpoints {
			if endpoint.Handler == funcName {
				// 从注释中获取更多信息
				if funcDecl.Doc != nil {
					g.updateEndpointFromComments(funcDecl.Doc.Text(), &endpoints[i])
				}
				break
			}
		}
	}
}

// shouldIncludeFunction 检查是否应该包含该函数
func (g *APIDocGenerator) shouldIncludeFunction(funcName string) bool {
	// 检查是否为内部函数（以_开头）
	if strings.HasPrefix(funcName, "_") && !g.includeInternal {
		return false
	}

	// 检查是否为私有函数（小写开头）
	if len(funcName) > 0 && funcName[0] >= 'a' && funcName[0] <= 'z' && !g.includePrivate {
		return false
	}

	return true
}

// updateEndpointFromComments 从注释中更新端点信息
func (g *APIDocGenerator) updateEndpointFromComments(comments string, endpoint *APIEndpoint) {
	lines := strings.Split(comments, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// 提取描述信息
		if strings.HasPrefix(line, "@Description") || strings.HasPrefix(line, "@描述") {
			endpoint.Description = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "@Description"), "@描述"))
		}

		// 提取标签信息
		if strings.HasPrefix(line, "@Tag") || strings.HasPrefix(line, "@标签") {
			tag := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "@Tag"), "@标签"))
			if tag != "" {
				endpoint.Tags = append(endpoint.Tags, tag)
			}
		}

		// 提取废弃信息
		if strings.HasPrefix(line, "@Deprecated") || strings.HasPrefix(line, "@废弃") {
			endpoint.Deprecated = true
			message := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "@Deprecated"), "@废弃"))
			if message != "" {
				endpoint.DeprecationMessage = message
			}
		}

		// 提取参数信息
		if strings.HasPrefix(line, "@Param") || strings.HasPrefix(line, "@参数") {
			paramInfo := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "@Param"), "@参数"))
			parts := strings.SplitN(paramInfo, " ", 5)
			if len(parts) >= 3 {
				param := APIParam{
					Name:     parts[0],
					Type:     parts[1],
					Location: parts[2],
					Required: true,
				}
				if len(parts) >= 4 {
					param.Required = parts[3] == "required"
				}
				if len(parts) >= 5 {
					param.Description = parts[4]
				}
				endpoint.RequestParams = append(endpoint.RequestParams, param)
			}
		}

		// 提取响应状态码信息
		if strings.HasPrefix(line, "@Status") || strings.HasPrefix(line, "@状态码") {
			statusInfo := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "@Status"), "@状态码"))
			parts := strings.SplitN(statusInfo, " ", 2)
			if len(parts) >= 1 {
				code := 0
				fmt.Sscanf(parts[0], "%d", &code)
				if code > 0 {
					status := APIStatusCode{
						Code: code,
					}
					if len(parts) >= 2 {
						status.Description = parts[1]
					}
					endpoint.StatusCodes = append(endpoint.StatusCodes, status)
				}
			}
		}
	}
}

// extractModels 提取模型定义
func (g *APIDocGenerator) extractModels() (map[string]interface{}, error) {
	models := make(map[string]interface{})

	// TODO: 实现模型定义的提取
	// 目前使用示例数据

	// 示例用户模型
	models["User"] = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id": map[string]interface{}{
				"type":        "string",
				"description": "用户唯一标识符",
			},
			"name": map[string]interface{}{
				"type":        "string",
				"description": "用户名",
			},
			"email": map[string]interface{}{
				"type":        "string",
				"description": "电子邮箱",
			},
			"created_at": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "创建时间",
			},
		},
		"required": []string{"id", "name", "email"},
	}

	// 可以添加更多示例模型...

	return models, nil
}

// writeDocumentation 将文档写入文件
func (g *APIDocGenerator) writeDocumentation(doc APIDocumentation) error {
	// JSON格式
	jsonData, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}

	jsonFile := filepath.Join(g.outputDir, "api.json")
	if err := os.WriteFile(jsonFile, jsonData, 0644); err != nil {
		return err
	}

	// Markdown格式（如果启用）
	if g.useMarkdown {
		markdownContent, err := g.generateMarkdown(doc)
		if err != nil {
			return err
		}

		markdownFile := filepath.Join(g.outputDir, "api.md")
		if err := os.WriteFile(markdownFile, []byte(markdownContent), 0644); err != nil {
			return err
		}
	}

	return nil
}

// generateMarkdown 生成Markdown格式的文档
func (g *APIDocGenerator) generateMarkdown(doc APIDocumentation) (string, error) {
	var content strings.Builder

	// 标题和描述
	content.WriteString(fmt.Sprintf("# %s\n\n", doc.Title))
	content.WriteString(fmt.Sprintf("%s\n\n", doc.Description))

	// 基本信息
	content.WriteString("## 基本信息\n\n")
	content.WriteString(fmt.Sprintf("- **版本**: %s\n", doc.Version))
	content.WriteString(fmt.Sprintf("- **基础URL**: %s\n", doc.BaseURL))
	content.WriteString(fmt.Sprintf("- **生成时间**: %s\n", doc.GeneratedAt.Format("2006-01-02 15:04:05")))
	if doc.Author != "" {
		content.WriteString(fmt.Sprintf("- **作者**: %s\n", doc.Author))
	}
	if doc.Email != "" {
		content.WriteString(fmt.Sprintf("- **联系邮箱**: %s\n", doc.Email))
	}
	if doc.License != "" {
		content.WriteString(fmt.Sprintf("- **许可证**: %s\n", doc.License))
	}
	content.WriteString("\n")

	// 按分组组织端点
	groups := make(map[string][]APIEndpoint)
	for _, endpoint := range doc.Endpoints {
		group := endpoint.Group
		if group == "" {
			group = "默认"
		}
		groups[group] = append(groups[group], endpoint)
	}

	// 端点信息
	content.WriteString("## API端点\n\n")

	// 目录
	content.WriteString("### 目录\n\n")
	for group := range groups {
		content.WriteString(fmt.Sprintf("- [%s](#%s)\n", group, strings.ToLower(strings.ReplaceAll(group, " ", "-"))))
	}
	content.WriteString("\n")

	// 按分组输出端点详情
	for group, endpoints := range groups {
		content.WriteString(fmt.Sprintf("### %s\n\n", group))

		for _, endpoint := range endpoints {
			// 端点标题
			content.WriteString(fmt.Sprintf("#### `%s` %s\n\n", endpoint.Method, endpoint.Path))

			// 描述
			if endpoint.Description != "" {
				content.WriteString(fmt.Sprintf("%s\n\n", endpoint.Description))
			}

			// 废弃警告
			if endpoint.Deprecated {
				content.WriteString("> **警告**: 此端点已废弃")
				if endpoint.DeprecationMessage != "" {
					content.WriteString(fmt.Sprintf(" - %s", endpoint.DeprecationMessage))
				}
				content.WriteString("\n\n")
			}

			// 标签
			if len(endpoint.Tags) > 0 {
				content.WriteString("**标签**: ")
				for i, tag := range endpoint.Tags {
					if i > 0 {
						content.WriteString(", ")
					}
					content.WriteString(fmt.Sprintf("`%s`", tag))
				}
				content.WriteString("\n\n")
			}

			// 中间件
			if len(endpoint.Middleware) > 0 {
				content.WriteString("**中间件**: ")
				for i, mw := range endpoint.Middleware {
					if i > 0 {
						content.WriteString(", ")
					}
					content.WriteString(fmt.Sprintf("`%s`", mw))
				}
				content.WriteString("\n\n")
			}

			// 请求参数
			if len(endpoint.RequestParams) > 0 {
				content.WriteString("**请求参数**:\n\n")
				content.WriteString("| 名称 | 类型 | 位置 | 必需 | 描述 |\n")
				content.WriteString("|------|------|------|------|------|\n")
				for _, param := range endpoint.RequestParams {
					required := "否"
					if param.Required {
						required = "是"
					}
					content.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
						param.Name,
						param.Type,
						param.Location,
						required,
						param.Description))
				}
				content.WriteString("\n")
			}

			// 请求体
			if endpoint.RequestBody != nil {
				content.WriteString("**请求体**:\n\n")
				content.WriteString("```json\n")
				requestJSON, _ := json.MarshalIndent(endpoint.RequestBody, "", "  ")
				content.WriteString(string(requestJSON))
				content.WriteString("\n```\n\n")
			}

			// 响应状态码
			if len(endpoint.StatusCodes) > 0 {
				content.WriteString("**响应状态码**:\n\n")
				content.WriteString("| 状态码 | 描述 |\n")
				content.WriteString("|--------|------|\n")
				for _, status := range endpoint.StatusCodes {
					content.WriteString(fmt.Sprintf("| %d | %s |\n", status.Code, status.Description))
				}
				content.WriteString("\n")
			}

			// 响应体
			if endpoint.ResponseBody != nil {
				content.WriteString("**响应体**:\n\n")
				content.WriteString("```json\n")
				responseJSON, _ := json.MarshalIndent(endpoint.ResponseBody, "", "  ")
				content.WriteString(string(responseJSON))
				content.WriteString("\n```\n\n")
			}

			// 示例
			if len(endpoint.Examples) > 0 {
				content.WriteString("**示例**:\n\n")
				for i, example := range endpoint.Examples {
					content.WriteString(fmt.Sprintf("**示例 %d**: %s\n\n", i+1, example.Name))
					if example.Description != "" {
						content.WriteString(fmt.Sprintf("%s\n\n", example.Description))
					}

					// 请求示例
					content.WriteString("请求:\n\n")
					content.WriteString("```json\n")
					requestJSON, _ := json.MarshalIndent(example.Request, "", "  ")
					content.WriteString(string(requestJSON))
					content.WriteString("\n```\n\n")

					// 响应示例
					content.WriteString("响应:\n\n")
					content.WriteString("```json\n")
					responseJSON, _ := json.MarshalIndent(example.Response, "", "  ")
					content.WriteString(string(responseJSON))
					content.WriteString("\n```\n\n")
				}
			}
		}
	}

	// 模型定义
	if len(doc.Models) > 0 {
		content.WriteString("## 模型定义\n\n")
		for name, model := range doc.Models {
			content.WriteString(fmt.Sprintf("### %s\n\n", name))
			modelJSON, _ := json.MarshalIndent(model, "", "  ")
			content.WriteString("```json\n")
			content.WriteString(string(modelJSON))
			content.WriteString("\n```\n\n")
		}
	}

	return content.String(), nil
}
