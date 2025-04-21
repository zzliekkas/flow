package docs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/zzliekkas/flow/app"
)

// DocumentationGenerator 是整体文档生成器，协调各种类型的文档生成
type DocumentationGenerator struct {
	// 应用程序实例
	app *app.Application

	// 文档输出根目录
	outputDir string

	// UI输出目录
	uiDir string

	// 项目名称
	projectName string

	// 项目描述
	description string

	// 项目版本
	version string

	// 文档标题
	title string

	// 生成器列表
	generators []Generator

	// 是否包含API文档
	includeAPI bool

	// 是否包含模块文档
	includeModules bool

	// 是否包含数据库文档
	includeDatabase bool

	// 是否包含CLI文档
	includeCLI bool

	// 是否包含配置文档
	includeConfig bool

	// 是否需要生成UI
	generateUI bool

	// UI主题
	uiTheme string

	// 文档搜索功能
	enableSearch bool

	// 自定义CSS
	customCSS string

	// 自定义JS
	customJS string

	// 自定义Logo
	logoPath string

	// 自定义页脚
	footer string

	// 文档基础URL
	baseURL string

	// Google Analytics ID
	gaID string
}

// Generator 是文档生成器接口
type Generator interface {
	// Generate 生成文档
	Generate() error
}

// NewDocumentationGenerator 创建新的文档生成器
func NewDocumentationGenerator(application *app.Application) *DocumentationGenerator {
	return &DocumentationGenerator{
		app:             application,
		outputDir:       "./docs/output",
		uiDir:           "./docs/ui",
		projectName:     "Flow Framework",
		title:           "Flow Framework Documentation",
		version:         "v1.0.0",
		description:     "Complete documentation for the Flow framework",
		includeAPI:      true,
		includeModules:  true,
		includeDatabase: true,
		includeCLI:      true,
		includeConfig:   true,
		generateUI:      true,
		uiTheme:         "default",
		enableSearch:    true,
		baseURL:         "/docs",
	}
}

// SetOutputDir 设置输出目录
func (g *DocumentationGenerator) SetOutputDir(dir string) *DocumentationGenerator {
	g.outputDir = dir
	return g
}

// SetUIDir 设置UI目录
func (g *DocumentationGenerator) SetUIDir(dir string) *DocumentationGenerator {
	g.uiDir = dir
	return g
}

// SetProjectName 设置项目名称
func (g *DocumentationGenerator) SetProjectName(name string) *DocumentationGenerator {
	g.projectName = name
	return g
}

// SetDescription 设置项目描述
func (g *DocumentationGenerator) SetDescription(desc string) *DocumentationGenerator {
	g.description = desc
	return g
}

// SetVersion 设置项目版本
func (g *DocumentationGenerator) SetVersion(version string) *DocumentationGenerator {
	g.version = version
	return g
}

// SetTitle 设置文档标题
func (g *DocumentationGenerator) SetTitle(title string) *DocumentationGenerator {
	g.title = title
	return g
}

// EnableAPI 启用API文档生成
func (g *DocumentationGenerator) EnableAPI(enable bool) *DocumentationGenerator {
	g.includeAPI = enable
	return g
}

// EnableModules 启用模块文档生成
func (g *DocumentationGenerator) EnableModules(enable bool) *DocumentationGenerator {
	g.includeModules = enable
	return g
}

// EnableDatabase 启用数据库文档生成
func (g *DocumentationGenerator) EnableDatabase(enable bool) *DocumentationGenerator {
	g.includeDatabase = enable
	return g
}

// EnableCLI 启用命令行文档生成
func (g *DocumentationGenerator) EnableCLI(enable bool) *DocumentationGenerator {
	g.includeCLI = enable
	return g
}

// EnableConfig 启用配置文档生成
func (g *DocumentationGenerator) EnableConfig(enable bool) *DocumentationGenerator {
	g.includeConfig = enable
	return g
}

// EnableUI 启用UI生成
func (g *DocumentationGenerator) EnableUI(enable bool) *DocumentationGenerator {
	g.generateUI = enable
	return g
}

// SetUITheme 设置UI主题
func (g *DocumentationGenerator) SetUITheme(theme string) *DocumentationGenerator {
	g.uiTheme = theme
	return g
}

// EnableSearch 启用搜索功能
func (g *DocumentationGenerator) EnableSearch(enable bool) *DocumentationGenerator {
	g.enableSearch = enable
	return g
}

// SetCustomCSS 设置自定义CSS
func (g *DocumentationGenerator) SetCustomCSS(css string) *DocumentationGenerator {
	g.customCSS = css
	return g
}

// SetCustomJS 设置自定义JS
func (g *DocumentationGenerator) SetCustomJS(js string) *DocumentationGenerator {
	g.customJS = js
	return g
}

// SetLogoPath 设置Logo路径
func (g *DocumentationGenerator) SetLogoPath(path string) *DocumentationGenerator {
	g.logoPath = path
	return g
}

// SetFooter 设置页脚
func (g *DocumentationGenerator) SetFooter(footer string) *DocumentationGenerator {
	g.footer = footer
	return g
}

// SetBaseURL 设置基础URL
func (g *DocumentationGenerator) SetBaseURL(url string) *DocumentationGenerator {
	g.baseURL = url
	return g
}

// SetGoogleAnalyticsID 设置Google Analytics ID
func (g *DocumentationGenerator) SetGoogleAnalyticsID(id string) *DocumentationGenerator {
	g.gaID = id
	return g
}

// AddGenerator 添加自定义生成器
func (g *DocumentationGenerator) AddGenerator(generator Generator) *DocumentationGenerator {
	g.generators = append(g.generators, generator)
	return g
}

// Generate 执行文档生成
func (g *DocumentationGenerator) Generate() error {
	// 创建输出目录
	if err := os.MkdirAll(g.outputDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	// 初始化所有生成器
	if err := g.initGenerators(); err != nil {
		return fmt.Errorf("初始化生成器失败: %w", err)
	}

	// 生成各类文档
	if err := g.generateDocs(); err != nil {
		return fmt.Errorf("生成文档失败: %w", err)
	}

	// 生成UI
	if g.generateUI {
		if err := g.generateDocUI(); err != nil {
			return fmt.Errorf("生成文档UI失败: %w", err)
		}
	}

	fmt.Printf("文档生成完成，输出目录: %s\n", g.outputDir)
	if g.generateUI {
		fmt.Printf("文档UI生成完成，访问: %s\n", filepath.Join(g.outputDir, "index.html"))
	}

	return nil
}

// initGenerators 初始化所有文档生成器
func (g *DocumentationGenerator) initGenerators() error {
	// 添加API文档生成器
	if g.includeAPI {
		apiGen := NewAPIDocGenerator(g.app)
		apiGen.SetOutputDir(filepath.Join(g.outputDir, "api"))
		apiGen.SetTitle(fmt.Sprintf("%s API Documentation", g.projectName))
		apiGen.SetDescription(fmt.Sprintf("API documentation for %s", g.projectName))
		apiGen.SetAPIVersion(g.version)
		apiGen.UseMarkdown(true)
		g.generators = append(g.generators, apiGen)
	}

	// 添加模块文档生成器
	if g.includeModules {
		moduleGen := NewModuleDocGenerator(g.app)
		moduleGen.SetOutputDir(filepath.Join(g.outputDir, "modules"))
		g.generators = append(g.generators, moduleGen)
	}

	// 添加数据库文档生成器
	if g.includeDatabase {
		dbGen := NewDatabaseDocGenerator(g.app)
		dbGen.SetOutputDir(filepath.Join(g.outputDir, "database"))
		g.generators = append(g.generators, dbGen)
	}

	// 添加CLI文档生成器
	if g.includeCLI {
		cliGen := NewCLIDocGenerator(g.app)
		cliGen.SetOutputDir(filepath.Join(g.outputDir, "cli"))
		g.generators = append(g.generators, cliGen)
	}

	// 添加配置文档生成器
	if g.includeConfig {
		configGen := NewConfigDocGenerator(g.app)
		configGen.SetOutputDir(filepath.Join(g.outputDir, "config"))
		g.generators = append(g.generators, configGen)
	}

	return nil
}

// generateDocs 生成所有文档
func (g *DocumentationGenerator) generateDocs() error {
	for _, generator := range g.generators {
		if err := generator.Generate(); err != nil {
			return err
		}
	}
	return nil
}

// generateDocUI 生成文档UI
func (g *DocumentationGenerator) generateDocUI() error {
	// 创建UI目录
	if err := os.MkdirAll(g.uiDir, 0755); err != nil {
		return fmt.Errorf("创建UI目录失败: %w", err)
	}

	// 生成UI资源
	if err := g.generateUIResources(); err != nil {
		return fmt.Errorf("生成UI资源失败: %w", err)
	}

	// 生成首页
	if err := g.generateIndexPage(); err != nil {
		return fmt.Errorf("生成首页失败: %w", err)
	}

	// 生成导航
	if err := g.generateNavigation(); err != nil {
		return fmt.Errorf("生成导航失败: %w", err)
	}

	// 复制UI资源到输出目录
	if err := CopyDir(g.uiDir, g.outputDir); err != nil {
		return fmt.Errorf("复制UI资源失败: %w", err)
	}

	return nil
}

// generateUIResources 生成UI资源
func (g *DocumentationGenerator) generateUIResources() error {
	// 创建styles目录
	stylesDir := filepath.Join(g.uiDir, "styles")
	if err := os.MkdirAll(stylesDir, 0755); err != nil {
		return err
	}

	// 创建scripts目录
	scriptsDir := filepath.Join(g.uiDir, "scripts")
	if err := os.MkdirAll(scriptsDir, 0755); err != nil {
		return err
	}

	// 创建images目录
	imagesDir := filepath.Join(g.uiDir, "images")
	if err := os.MkdirAll(imagesDir, 0755); err != nil {
		return err
	}

	// 生成主CSS文件
	mainCss := `
body {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Oxygen, Ubuntu, Cantarell, "Fira Sans", "Droid Sans", "Helvetica Neue", sans-serif;
  line-height: 1.6;
  color: #333;
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}

.container {
  display: flex;
  min-height: 100vh;
}

.sidebar {
  width: 250px;
  padding: 20px;
  border-right: 1px solid #eee;
  position: sticky;
  top: 0;
  height: 100vh;
  overflow-y: auto;
}

.content {
  flex-grow: 1;
  padding: 20px;
}

.header {
  margin-bottom: 20px;
  padding-bottom: 20px;
  border-bottom: 1px solid #eee;
}

.nav-item {
  margin-bottom: 10px;
}

.nav-item a {
  color: #333;
  text-decoration: none;
  display: block;
  padding: 5px 0;
}

.nav-item a:hover {
  color: #0066cc;
}

.nav-group {
  margin-bottom: 15px;
}

.nav-group-title {
  font-weight: bold;
  margin-bottom: 10px;
  color: #666;
}

.footer {
  margin-top: 40px;
  padding-top: 20px;
  border-top: 1px solid #eee;
  color: #666;
  font-size: 14px;
}

h1, h2, h3, h4, h5, h6 {
  margin-top: 1.5em;
  margin-bottom: 0.5em;
}

code {
  background-color: #f5f5f5;
  padding: 0.2em 0.4em;
  border-radius: 3px;
  font-family: SFMono-Regular, Consolas, "Liberation Mono", Menlo, monospace;
}

pre {
  background-color: #f5f5f5;
  padding: 15px;
  border-radius: 5px;
  overflow-x: auto;
}

pre code {
  background-color: transparent;
  padding: 0;
}

table {
  border-collapse: collapse;
  width: 100%;
  margin: 20px 0;
}

th, td {
  border: 1px solid #ddd;
  padding: 8px 12px;
  text-align: left;
}

th {
  background-color: #f5f5f5;
}

blockquote {
  border-left: 4px solid #ddd;
  padding-left: 15px;
  color: #666;
  margin: 20px 0;
}

img {
  max-width: 100%;
}

.search-container {
  margin-bottom: 20px;
}

.search-input {
  width: 100%;
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .container {
    flex-direction: column;
  }
  
  .sidebar {
    width: 100%;
    border-right: none;
    border-bottom: 1px solid #eee;
    position: relative;
    height: auto;
  }
}
`
	if err := os.WriteFile(filepath.Join(stylesDir, "main.css"), []byte(mainCss), 0644); err != nil {
		return err
	}

	// 生成主JS文件
	mainJs := `
document.addEventListener('DOMContentLoaded', function() {
  // 搜索功能
  const searchInput = document.getElementById('search');
  if (searchInput) {
    searchInput.addEventListener('input', function() {
      const value = this.value.toLowerCase();
      const items = document.querySelectorAll('.searchable');
      
      items.forEach(function(item) {
        const text = item.textContent.toLowerCase();
        if (text.includes(value) || value === '') {
          item.style.display = '';
        } else {
          item.style.display = 'none';
        }
      });
    });
  }

  // 移动端导航菜单
  const menuToggle = document.getElementById('menu-toggle');
  const sidebar = document.querySelector('.sidebar');
  
  if (menuToggle && sidebar) {
    menuToggle.addEventListener('click', function() {
      sidebar.classList.toggle('active');
    });
  }

  // 代码高亮
  document.querySelectorAll('pre code').forEach(function(block) {
    hljs.highlightElement(block);
  });
});
`
	if err := os.WriteFile(filepath.Join(scriptsDir, "main.js"), []byte(mainJs), 0644); err != nil {
		return err
	}

	// 添加自定义CSS（如果有）
	if g.customCSS != "" {
		if err := os.WriteFile(filepath.Join(stylesDir, "custom.css"), []byte(g.customCSS), 0644); err != nil {
			return err
		}
	}

	// 添加自定义JS（如果有）
	if g.customJS != "" {
		if err := os.WriteFile(filepath.Join(scriptsDir, "custom.js"), []byte(g.customJS), 0644); err != nil {
			return err
		}
	}

	// 添加默认Logo（如果没有自定义Logo）
	if g.logoPath == "" {
		defaultLogo := `
<svg width="200" height="60" viewBox="0 0 200 60" xmlns="http://www.w3.org/2000/svg">
  <path d="M40 20 L60 20 L60 40 L40 40 Z" fill="#0066cc" />
  <path d="M65 20 L85 20 L85 40 L65 40 Z" fill="#0099cc" />
  <path d="M90 20 L110 20 L110 40 L90 40 Z" fill="#00cccc" />
  <text x="120" y="35" font-family="Arial" font-size="24" font-weight="bold" fill="#333">Flow</text>
</svg>
`
		if err := os.WriteFile(filepath.Join(imagesDir, "logo.svg"), []byte(defaultLogo), 0644); err != nil {
			return err
		}
	} else {
		// 复制自定义Logo
		// TODO: 实现Logo复制
	}

	return nil
}

// generateIndexPage 生成文档首页
func (g *DocumentationGenerator) generateIndexPage() error {
	indexContent := fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>%s</title>
  <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.5.1/styles/default.min.css">
  <link rel="stylesheet" href="styles/main.css">
  %s
  <script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.5.1/highlight.min.js"></script>
  <script src="scripts/main.js"></script>
  %s
  %s
</head>
<body>
  <div class="container">
    <div class="sidebar">
      <div class="header">
        <img src="images/logo.svg" alt="%s" width="200">
      </div>
      
      <div class="search-container">
        <input type="text" id="search" class="search-input" placeholder="搜索文档...">
      </div>
      
      <nav id="navigation">
        <!-- 导航将通过JS加载 -->
        加载导航...
      </nav>
    </div>
    
    <div class="content">
      <div class="header">
        <h1>%s</h1>
        <p>%s</p>
        <p>版本: %s</p>
      </div>
      
      <main>
        <h2>文档目录</h2>
        <ul>
          %s
        </ul>
        
        <h2>快速入门</h2>
        <p>欢迎使用Flow框架文档。本文档提供了框架的详细说明和使用指南。</p>
        
        <h3>安装框架</h3>
        <pre><code class="language-bash">go get github.com/zzliekkas/flow</code></pre>
        
        <h3>创建项目</h3>
        <pre><code class="language-bash">go run github.com/zzliekkas/flow/cmd/flow@latest new myproject</code></pre>
        
        <h3>运行项目</h3>
        <pre><code class="language-bash">cd myproject
go run cmd/myproject/main.go serve</code></pre>
      </main>
      
      <footer class="footer">
        %s
      </footer>
    </div>
  </div>
  
  <script>
    // 加载导航
    fetch('navigation.json')
      .then(response => response.json())
      .then(data => {
        const navElement = document.getElementById('navigation');
        navElement.innerHTML = '';
        
        data.forEach(group => {
          const groupDiv = document.createElement('div');
          groupDiv.className = 'nav-group';
          
          const titleDiv = document.createElement('div');
          titleDiv.className = 'nav-group-title';
          titleDiv.textContent = group.title;
          groupDiv.appendChild(titleDiv);
          
          group.items.forEach(item => {
            const itemDiv = document.createElement('div');
            itemDiv.className = 'nav-item searchable';
            
            const link = document.createElement('a');
            link.href = item.url;
            link.textContent = item.title;
            itemDiv.appendChild(link);
            
            groupDiv.appendChild(itemDiv);
          });
          
          navElement.appendChild(groupDiv);
        });
      })
      .catch(error => {
        console.error('加载导航失败:', error);
        document.getElementById('navigation').innerHTML = '<p>加载导航失败</p>';
      });
  </script>
</body>
</html>
`,
		g.title,
		func() string {
			if g.customCSS != "" {
				return `<link rel="stylesheet" href="styles/custom.css">`
			}
			return ""
		}(),
		func() string {
			if g.customJS != "" {
				return `<script src="scripts/custom.js"></script>`
			}
			return ""
		}(),
		func() string {
			if g.gaID != "" {
				return fmt.Sprintf(`
  <!-- Global site tag (gtag.js) - Google Analytics -->
  <script async src="https://www.googletagmanager.com/gtag/js?id=%s"></script>
  <script>
    window.dataLayer = window.dataLayer || [];
    function gtag(){dataLayer.push(arguments);}
    gtag('js', new Date());
    gtag('config', '%s');
  </script>`, g.gaID, g.gaID)
			}
			return ""
		}(),
		g.projectName,
		g.projectName,
		g.description,
		g.version,
		g.generateDocLinks(),
		func() string {
			if g.footer != "" {
				return g.footer
			}
			return fmt.Sprintf("© %d %s. 保留所有权利。", time.Now().Year(), g.projectName)
		}())

	return os.WriteFile(filepath.Join(g.uiDir, "index.html"), []byte(indexContent), 0644)
}

// generateDocLinks 生成文档链接列表
func (g *DocumentationGenerator) generateDocLinks() string {
	var links []string

	if g.includeAPI {
		links = append(links, `<li><a href="api/index.html">API 文档</a> - 详细的API参考和使用说明</li>`)
	}

	if g.includeModules {
		links = append(links, `<li><a href="modules/index.html">模块文档</a> - 框架核心模块的详细说明</li>`)
	}

	if g.includeDatabase {
		links = append(links, `<li><a href="database/index.html">数据库文档</a> - 数据库模式和模型参考</li>`)
	}

	if g.includeCLI {
		links = append(links, `<li><a href="cli/index.html">命令行工具</a> - CLI命令和选项的完整列表</li>`)
	}

	if g.includeConfig {
		links = append(links, `<li><a href="config/index.html">配置参考</a> - 配置选项和环境变量</li>`)
	}

	return strings.Join(links, "\n          ")
}

// generateNavigation 生成导航JSON
func (g *DocumentationGenerator) generateNavigation() error {
	navigation := []map[string]interface{}{
		{
			"title": "指南",
			"items": []map[string]string{
				{"title": "首页", "url": "index.html"},
				{"title": "快速入门", "url": "index.html#快速入门"},
				{"title": "安装说明", "url": "index.html#安装框架"},
			},
		},
	}

	if g.includeAPI {
		navigation = append(navigation, map[string]interface{}{
			"title": "API 文档",
			"items": []map[string]string{
				{"title": "API 概述", "url": "api/index.html"},
				{"title": "端点参考", "url": "api/api.html"},
				{"title": "认证", "url": "api/auth.html"},
				{"title": "错误处理", "url": "api/errors.html"},
			},
		})
	}

	if g.includeModules {
		navigation = append(navigation, map[string]interface{}{
			"title": "模块",
			"items": []map[string]string{
				{"title": "模块概述", "url": "modules/index.html"},
				{"title": "核心模块", "url": "modules/core.html"},
				{"title": "HTTP", "url": "modules/http.html"},
				{"title": "存储", "url": "modules/storage.html"},
				{"title": "队列", "url": "modules/queue.html"},
				{"title": "缓存", "url": "modules/cache.html"},
			},
		})
	}

	if g.includeDatabase {
		navigation = append(navigation, map[string]interface{}{
			"title": "数据库",
			"items": []map[string]string{
				{"title": "数据库概述", "url": "database/index.html"},
				{"title": "模型定义", "url": "database/models.html"},
				{"title": "查询构建器", "url": "database/queries.html"},
				{"title": "迁移", "url": "database/migrations.html"},
			},
		})
	}

	if g.includeCLI {
		navigation = append(navigation, map[string]interface{}{
			"title": "命令行",
			"items": []map[string]string{
				{"title": "CLI 概述", "url": "cli/index.html"},
				{"title": "服务器命令", "url": "cli/serve.html"},
				{"title": "数据库命令", "url": "cli/db.html"},
				{"title": "生成命令", "url": "cli/make.html"},
				{"title": "队列命令", "url": "cli/queue.html"},
				{"title": "存储命令", "url": "cli/storage.html"},
			},
		})
	}

	if g.includeConfig {
		navigation = append(navigation, map[string]interface{}{
			"title": "配置",
			"items": []map[string]string{
				{"title": "配置概述", "url": "config/index.html"},
				{"title": "环境变量", "url": "config/env.html"},
				{"title": "应用配置", "url": "config/app.html"},
				{"title": "数据库配置", "url": "config/database.html"},
				{"title": "缓存配置", "url": "config/cache.html"},
			},
		})
	}

	// 转换为JSON
	jsonData, err := json.Marshal(navigation)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(g.uiDir, "navigation.json"), jsonData, 0644)
}
