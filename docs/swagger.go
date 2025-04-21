package docs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SwaggerGenerator 用于生成Swagger规范文档
type SwaggerGenerator struct {
	// API文档生成器
	apiGenerator *APIDocGenerator

	// 输出目录
	outputDir string

	// 主机名
	host string

	// 协议列表
	schemes []string

	// 安全定义
	securityDefinitions map[string]interface{}

	// 全局参数
	globalParams map[string]interface{}

	// 全局响应
	globalResponses map[string]interface{}
}

// SwaggerDocument 表示Swagger文档
type SwaggerDocument struct {
	Swagger             string                 `json:"swagger"`
	Info                SwaggerInfo            `json:"info"`
	Host                string                 `json:"host,omitempty"`
	BasePath            string                 `json:"basePath,omitempty"`
	Schemes             []string               `json:"schemes,omitempty"`
	Paths               map[string]interface{} `json:"paths"`
	Definitions         map[string]interface{} `json:"definitions,omitempty"`
	SecurityDefinitions map[string]interface{} `json:"securityDefinitions,omitempty"`
	Tags                []SwaggerTag           `json:"tags,omitempty"`
}

// SwaggerInfo 表示Swagger信息
type SwaggerInfo struct {
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Version     string          `json:"version"`
	Contact     *SwaggerContact `json:"contact,omitempty"`
	License     *SwaggerLicense `json:"license,omitempty"`
}

// SwaggerContact 表示联系人信息
type SwaggerContact struct {
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

// SwaggerLicense 表示许可证信息
type SwaggerLicense struct {
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

// SwaggerTag 表示API分组标签
type SwaggerTag struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// NewSwaggerGenerator 创建新的Swagger文档生成器
func NewSwaggerGenerator(apiGenerator *APIDocGenerator) *SwaggerGenerator {
	return &SwaggerGenerator{
		apiGenerator:        apiGenerator,
		outputDir:           "./docs/swagger",
		schemes:             []string{"http", "https"},
		securityDefinitions: make(map[string]interface{}),
		globalParams:        make(map[string]interface{}),
		globalResponses:     make(map[string]interface{}),
	}
}

// SetOutputDir 设置输出目录
func (g *SwaggerGenerator) SetOutputDir(dir string) *SwaggerGenerator {
	g.outputDir = dir
	return g
}

// SetHost 设置主机名
func (g *SwaggerGenerator) SetHost(host string) *SwaggerGenerator {
	g.host = host
	return g
}

// SetSchemes 设置协议
func (g *SwaggerGenerator) SetSchemes(schemes ...string) *SwaggerGenerator {
	g.schemes = schemes
	return g
}

// AddSecurityDefinition 添加安全定义
func (g *SwaggerGenerator) AddSecurityDefinition(name string, definition interface{}) *SwaggerGenerator {
	g.securityDefinitions[name] = definition
	return g
}

// AddJWTSecurityDefinition 添加JWT安全定义
func (g *SwaggerGenerator) AddJWTSecurityDefinition() *SwaggerGenerator {
	g.securityDefinitions["Bearer"] = map[string]interface{}{
		"type":        "apiKey",
		"name":        "Authorization",
		"in":          "header",
		"description": "使用Bearer token进行身份验证，格式: Bearer {token}",
	}
	return g
}

// AddBasicAuthSecurityDefinition 添加基本认证安全定义
func (g *SwaggerGenerator) AddBasicAuthSecurityDefinition() *SwaggerGenerator {
	g.securityDefinitions["BasicAuth"] = map[string]interface{}{
		"type":        "basic",
		"description": "HTTP基本认证",
	}
	return g
}

// AddGlobalParameter 添加全局参数
func (g *SwaggerGenerator) AddGlobalParameter(name string, parameter interface{}) *SwaggerGenerator {
	g.globalParams[name] = parameter
	return g
}

// AddGlobalResponse 添加全局响应
func (g *SwaggerGenerator) AddGlobalResponse(name string, response interface{}) *SwaggerGenerator {
	g.globalResponses[name] = response
	return g
}

// Generate 生成Swagger文档
func (g *SwaggerGenerator) Generate() error {
	// 确保输出目录存在
	if err := os.MkdirAll(g.outputDir, 0755); err != nil {
		return fmt.Errorf("创建Swagger输出目录失败: %w", err)
	}

	// 使用API文档生成器创建文档
	apiDoc, err := g.generateAPIDoc()
	if err != nil {
		return fmt.Errorf("生成API文档失败: %w", err)
	}

	// 转换为Swagger文档
	swaggerDoc, err := g.convertToSwagger(apiDoc)
	if err != nil {
		return fmt.Errorf("转换为Swagger文档失败: %w", err)
	}

	// 写入JSON文件
	outputPath := filepath.Join(g.outputDir, "swagger.json")
	jsonData, err := json.MarshalIndent(swaggerDoc, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化Swagger文档失败: %w", err)
	}

	if err := os.WriteFile(outputPath, jsonData, 0644); err != nil {
		return fmt.Errorf("写入Swagger文档失败: %w", err)
	}

	// 生成Swagger UI
	if err := g.generateSwaggerUI(); err != nil {
		return fmt.Errorf("生成Swagger UI失败: %w", err)
	}

	fmt.Printf("Swagger文档已生成: %s\n", outputPath)
	return nil
}

// generateAPIDoc 生成API文档
func (g *SwaggerGenerator) generateAPIDoc() (APIDocumentation, error) {
	// 创建一个API路由收集器
	collector := &apiDocCollector{
		doc: APIDocumentation{
			Title:       g.apiGenerator.title,
			Description: g.apiGenerator.description,
			Version:     g.apiGenerator.apiVersion,
			BaseURL:     g.apiGenerator.baseURL,
			Author:      g.apiGenerator.author,
			Email:       g.apiGenerator.email,
			License:     g.apiGenerator.license,
			GeneratedAt: time.Now(),
		},
	}

	// 收集路由
	endpoints, err := g.apiGenerator.collectRoutes()
	if err != nil {
		return APIDocumentation{}, err
	}
	collector.doc.Endpoints = endpoints

	// 解析源代码以丰富文档
	err = g.apiGenerator.parseSourceCode(collector.doc.Endpoints)
	if err != nil {
		return APIDocumentation{}, err
	}

	// 提取模型
	models, err := g.apiGenerator.extractModels()
	if err != nil {
		return APIDocumentation{}, err
	}
	collector.doc.Models = models

	return collector.doc, nil
}

// 用于收集API文档的辅助结构
type apiDocCollector struct {
	doc APIDocumentation
}

// convertToSwagger 将API文档转换为Swagger格式
func (g *SwaggerGenerator) convertToSwagger(apiDoc APIDocumentation) (SwaggerDocument, error) {
	swaggerDoc := SwaggerDocument{
		Swagger: "2.0",
		Info: SwaggerInfo{
			Title:       apiDoc.Title,
			Description: apiDoc.Description,
			Version:     apiDoc.Version,
		},
		Host:                g.host,
		BasePath:            g.apiGenerator.routePrefix,
		Schemes:             g.schemes,
		Paths:               make(map[string]interface{}),
		Definitions:         make(map[string]interface{}),
		SecurityDefinitions: g.securityDefinitions,
	}

	// 添加联系人信息（如果有）
	if apiDoc.Author != "" || apiDoc.Email != "" {
		swaggerDoc.Info.Contact = &SwaggerContact{
			Name:  apiDoc.Author,
			Email: apiDoc.Email,
		}
	}

	// 添加许可证信息（如果有）
	if apiDoc.License != "" {
		swaggerDoc.Info.License = &SwaggerLicense{
			Name: apiDoc.License,
		}
	}

	// 提取所有标签
	tagMap := make(map[string]bool)
	var tags []SwaggerTag

	// 转换路径
	for _, endpoint := range apiDoc.Endpoints {
		// 收集标签
		for _, tag := range endpoint.Tags {
			if !tagMap[tag] {
				tagMap[tag] = true
				tags = append(tags, SwaggerTag{Name: tag})
			}
		}

		// 创建路径
		path := endpoint.Path
		method := strings.ToLower(endpoint.Method)

		// 初始化路径对象（如果不存在）
		if _, exists := swaggerDoc.Paths[path]; !exists {
			swaggerDoc.Paths[path] = make(map[string]interface{})
		}

		// 处理路径参数
		parameters := []map[string]interface{}{}
		for _, param := range endpoint.RequestParams {
			swaggerParam := map[string]interface{}{
				"name":        param.Name,
				"in":          param.Location,
				"description": param.Description,
				"required":    param.Required,
			}

			// 设置类型
			switch param.Type {
			case "string":
				swaggerParam["type"] = "string"
			case "int", "int64", "int32":
				swaggerParam["type"] = "integer"
				if param.Type == "int64" {
					swaggerParam["format"] = "int64"
				} else if param.Type == "int32" {
					swaggerParam["format"] = "int32"
				}
			case "float", "float64", "float32":
				swaggerParam["type"] = "number"
				if param.Type == "float64" {
					swaggerParam["format"] = "double"
				} else if param.Type == "float32" {
					swaggerParam["format"] = "float"
				}
			case "bool":
				swaggerParam["type"] = "boolean"
			case "time.Time":
				swaggerParam["type"] = "string"
				swaggerParam["format"] = "date-time"
			default:
				swaggerParam["type"] = "string"
			}

			// 添加默认值（如果有）
			if param.DefaultValue != "" {
				swaggerParam["default"] = param.DefaultValue
			}

			// 添加示例（如果有）
			if param.Example != "" {
				swaggerParam["example"] = param.Example
			}

			parameters = append(parameters, swaggerParam)
		}

		// 创建响应对象
		responses := map[string]interface{}{}
		for _, statusCode := range endpoint.StatusCodes {
			code := fmt.Sprintf("%d", statusCode.Code)
			responses[code] = map[string]interface{}{
				"description": statusCode.Description,
			}
			if statusCode.Example != nil {
				responses[code].(map[string]interface{})["examples"] = map[string]interface{}{
					"application/json": statusCode.Example,
				}
			}
		}

		// 如果没有定义响应，添加默认200响应
		if len(responses) == 0 {
			responses["200"] = map[string]interface{}{
				"description": "成功",
			}
			if endpoint.ResponseBody != nil {
				responses["200"].(map[string]interface{})["examples"] = map[string]interface{}{
					"application/json": endpoint.ResponseBody,
				}
			}
		}

		// 创建操作对象
		operation := map[string]interface{}{
			"summary":     endpoint.Description,
			"description": endpoint.Description,
			"parameters":  parameters,
			"responses":   responses,
			"tags":        endpoint.Tags,
		}

		// 添加废弃标记（如果有）
		if endpoint.Deprecated {
			operation["deprecated"] = true
		}

		// 将操作添加到路径
		pathObj := swaggerDoc.Paths[path].(map[string]interface{})
		pathObj[method] = operation
	}

	// 添加所有标签
	swaggerDoc.Tags = tags

	// 添加模型定义
	for name, model := range apiDoc.Models {
		swaggerDoc.Definitions[name] = model
	}

	return swaggerDoc, nil
}

// generateSwaggerUI 生成Swagger UI
func (g *SwaggerGenerator) generateSwaggerUI() error {
	uiDir := filepath.Join(g.outputDir, "ui")
	if err := os.MkdirAll(uiDir, 0755); err != nil {
		return fmt.Errorf("创建Swagger UI目录失败: %w", err)
	}

	// 创建index.html
	indexHTML := `
<!DOCTYPE html>
<html lang="zh">
<head>
  <meta charset="UTF-8">
  <title>Swagger UI</title>
  <link rel="stylesheet" type="text/css" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.9.0/swagger-ui.css">
  <style>
    html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
    *, *:before, *:after { box-sizing: inherit; }
    body { margin: 0; padding: 0; }
    .topbar { display: none; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.9.0/swagger-ui-bundle.js"></script>
  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.9.0/swagger-ui-standalone-preset.js"></script>
  <script>
    window.onload = function() {
      const ui = SwaggerUIBundle({
        url: "../swagger.json",
        dom_id: "#swagger-ui",
        deepLinking: true,
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        plugins: [
          SwaggerUIBundle.plugins.DownloadUrl
        ],
        layout: "StandaloneLayout"
      });
      window.ui = ui;
    };
  </script>
</body>
</html>
`
	indexPath := filepath.Join(uiDir, "index.html")
	if err := os.WriteFile(indexPath, []byte(indexHTML), 0644); err != nil {
		return fmt.Errorf("写入Swagger UI HTML文件失败: %w", err)
	}

	return nil
}
