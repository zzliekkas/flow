package ui

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"time"
)

// Layout 表示文档UI的布局
type Layout struct {
	// 页面标题
	Title string

	// 文档标题
	DocTitle string

	// 项目名称
	ProjectName string

	// 项目版本
	Version string

	// 基础URL
	BaseURL string

	// 导航项目
	NavItems []NavItem

	// CSS文件列表
	CSSFiles []string

	// JS文件列表
	JSFiles []string

	// 自定义CSS代码
	CustomCSS string

	// 自定义JS代码
	CustomJS string

	// 页脚内容
	Footer string

	// Logo URL
	LogoURL string

	// 是否启用搜索
	EnableSearch bool

	// Google Analytics ID
	GoogleAnalyticsID string

	// 输出目录
	OutputDir string

	// 主题
	Theme string
}

// NavItem 表示导航项目
type NavItem struct {
	// 项目标题
	Title string

	// 链接URL
	URL string

	// 是否活跃
	Active bool

	// 子项目
	Children []NavItem

	// 图标（可选）
	Icon string

	// 是否展开
	Expanded bool
}

// NewLayout 创建新的布局
func NewLayout() *Layout {
	return &Layout{
		Title:        "Flow框架文档",
		DocTitle:     "文档",
		ProjectName:  "Flow框架",
		Version:      "v1.0.0",
		BaseURL:      "/docs",
		NavItems:     []NavItem{},
		CSSFiles:     []string{},
		JSFiles:      []string{},
		Footer:       "© " + fmt.Sprint(time.Now().Year()) + " Flow框架团队",
		EnableSearch: true,
		Theme:        "default",
		OutputDir:    "./docs/output",
	}
}

// SetTitle 设置页面标题
func (l *Layout) SetTitle(title string) *Layout {
	l.Title = title
	return l
}

// SetDocTitle 设置文档标题
func (l *Layout) SetDocTitle(title string) *Layout {
	l.DocTitle = title
	return l
}

// SetProjectName 设置项目名称
func (l *Layout) SetProjectName(name string) *Layout {
	l.ProjectName = name
	return l
}

// SetVersion 设置版本
func (l *Layout) SetVersion(version string) *Layout {
	l.Version = version
	return l
}

// SetBaseURL 设置基础URL
func (l *Layout) SetBaseURL(url string) *Layout {
	l.BaseURL = url
	return l
}

// AddNavItem 添加导航项目
func (l *Layout) AddNavItem(item NavItem) *Layout {
	l.NavItems = append(l.NavItems, item)
	return l
}

// AddCSSFile 添加CSS文件
func (l *Layout) AddCSSFile(file string) *Layout {
	l.CSSFiles = append(l.CSSFiles, file)
	return l
}

// AddJSFile 添加JS文件
func (l *Layout) AddJSFile(file string) *Layout {
	l.JSFiles = append(l.JSFiles, file)
	return l
}

// SetCustomCSS 设置自定义CSS
func (l *Layout) SetCustomCSS(css string) *Layout {
	l.CustomCSS = css
	return l
}

// SetCustomJS 设置自定义JS
func (l *Layout) SetCustomJS(js string) *Layout {
	l.CustomJS = js
	return l
}

// SetFooter 设置页脚
func (l *Layout) SetFooter(footer string) *Layout {
	l.Footer = footer
	return l
}

// SetLogoURL 设置Logo URL
func (l *Layout) SetLogoURL(url string) *Layout {
	l.LogoURL = url
	return l
}

// EnableSearchFunction 启用搜索功能
func (l *Layout) EnableSearchFunction(enable bool) *Layout {
	l.EnableSearch = enable
	return l
}

// SetGoogleAnalyticsID 设置Google Analytics ID
func (l *Layout) SetGoogleAnalyticsID(id string) *Layout {
	l.GoogleAnalyticsID = id
	return l
}

// SetOutputDir 设置输出目录
func (l *Layout) SetOutputDir(dir string) *Layout {
	l.OutputDir = dir
	return l
}

// SetTheme 设置主题
func (l *Layout) SetTheme(theme string) *Layout {
	l.Theme = theme
	return l
}

// GenerateLayout 生成布局文件
func (l *Layout) GenerateLayout() error {
	// 确保输出目录存在
	outputDir := filepath.Join(l.OutputDir, "assets")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	// 生成CSS文件
	if err := l.generateCSS(outputDir); err != nil {
		return fmt.Errorf("生成CSS文件失败: %w", err)
	}

	// 生成JS文件
	if err := l.generateJS(outputDir); err != nil {
		return fmt.Errorf("生成JS文件失败: %w", err)
	}

	// 生成模板文件
	if err := l.generateTemplates(l.OutputDir); err != nil {
		return fmt.Errorf("生成模板文件失败: %w", err)
	}

	return nil
}

// generateCSS 生成CSS文件
func (l *Layout) generateCSS(outputDir string) error {
	// 基础CSS
	baseCSS := `
:root {
    --primary-color: #4a6cf7;
    --secondary-color: #6c757d;
    --success-color: #28a745;
    --info-color: #17a2b8;
    --warning-color: #ffc107;
    --danger-color: #dc3545;
    --light-color: #f8f9fa;
    --dark-color: #343a40;
    --body-bg: #ffffff;
    --body-color: #212529;
    --link-color: #4a6cf7;
    --link-hover-color: #0056b3;
    --font-family-sans-serif: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
    --font-family-monospace: SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
    --sidebar-width: 280px;
    --content-max-width: 1200px;
    --header-height: 60px;
    --footer-height: 60px;
}

* {
    box-sizing: border-box;
}

body {
    font-family: var(--font-family-sans-serif);
    background-color: var(--body-bg);
    color: var(--body-color);
    line-height: 1.6;
    margin: 0;
    padding: 0;
}

a {
    color: var(--link-color);
    text-decoration: none;
}

a:hover {
    color: var(--link-hover-color);
    text-decoration: underline;
}

.container {
    display: flex;
    min-height: 100vh;
}

.sidebar {
    width: var(--sidebar-width);
    background-color: var(--light-color);
    border-right: 1px solid #ddd;
    position: fixed;
    height: 100vh;
    overflow-y: auto;
    padding: 20px 0;
    transition: transform 0.3s ease;
}

.sidebar-header {
    padding: 0 20px 20px;
    border-bottom: 1px solid #ddd;
    margin-bottom: 20px;
}

.sidebar-title {
    margin: 0;
    font-size: 1.2rem;
    font-weight: 600;
}

.sidebar-version {
    font-size: 0.8rem;
    color: var(--secondary-color);
}

.sidebar-nav {
    list-style: none;
    padding: 0;
    margin: 0;
}

.sidebar-nav-item {
    padding: 8px 20px;
    display: block;
    color: var(--body-color);
    border-left: 3px solid transparent;
}

.sidebar-nav-item:hover {
    background-color: rgba(0, 0, 0, 0.05);
    text-decoration: none;
}

.sidebar-nav-item.active {
    border-left-color: var(--primary-color);
    background-color: rgba(0, 0, 0, 0.05);
    font-weight: 600;
}

.sidebar-nav-item-child {
    padding-left: 40px;
    font-size: 0.9rem;
}

.content {
    flex: 1;
    margin-left: var(--sidebar-width);
    padding: 20px;
    max-width: calc(100% - var(--sidebar-width));
}

.content-header {
    margin-bottom: 30px;
    padding-bottom: 15px;
    border-bottom: 1px solid #ddd;
}

.content-title {
    margin: 0 0 10px;
    font-size: 2rem;
    font-weight: 600;
}

.content-main {
    max-width: var(--content-max-width);
    margin: 0 auto;
}

.footer {
    text-align: center;
    padding: 20px 0;
    margin-top: 40px;
    border-top: 1px solid #ddd;
    color: var(--secondary-color);
    font-size: 0.9rem;
}

.search-container {
    padding: 0 20px 20px;
    margin-bottom: 20px;
}

.search-input {
    width: 100%;
    padding: 8px 12px;
    border: 1px solid #ddd;
    border-radius: 4px;
    font-size: 0.9rem;
}

@media (max-width: 768px) {
    .sidebar {
        transform: translateX(-100%);
        z-index: 1000;
    }
    
    .sidebar.show {
        transform: translateX(0);
    }
    
    .content {
        margin-left: 0;
        max-width: 100%;
    }
    
    .menu-toggle {
        display: block;
        position: fixed;
        top: 10px;
        left: 10px;
        z-index: 1001;
        background-color: var(--primary-color);
        color: white;
        border: none;
        border-radius: 4px;
        padding: 5px 10px;
        cursor: pointer;
    }
}

/* 代码样式 */
pre {
    background-color: #f5f5f5;
    border: 1px solid #ddd;
    border-radius: 4px;
    padding: 15px;
    overflow-x: auto;
}

code {
    font-family: var(--font-family-monospace);
    background-color: #f5f5f5;
    padding: 2px 4px;
    border-radius: 3px;
    font-size: 0.9em;
}

pre code {
    padding: 0;
    background-color: transparent;
}

/* 表格样式 */
table {
    width: 100%;
    border-collapse: collapse;
    margin-bottom: 20px;
}

th, td {
    padding: 10px;
    border: 1px solid #ddd;
}

th {
    background-color: #f5f5f5;
    font-weight: 600;
    text-align: left;
}

tr:nth-child(even) {
    background-color: rgba(0, 0, 0, 0.02);
}

/* 暗色主题 */
@media (prefers-color-scheme: dark) {
    body.dark-theme {
        --body-bg: #1e1e1e;
        --body-color: #e4e4e4;
        --light-color: #252526;
        --link-color: #569cd6;
        --link-hover-color: #9cdcfe;
    }
    
    body.dark-theme .sidebar {
        background-color: #252526;
        border-right-color: #3e3e42;
    }
    
    body.dark-theme .sidebar-header,
    body.dark-theme .footer {
        border-color: #3e3e42;
    }
    
    body.dark-theme .sidebar-nav-item:hover {
        background-color: rgba(255, 255, 255, 0.05);
    }
    
    body.dark-theme .sidebar-nav-item.active {
        background-color: rgba(255, 255, 255, 0.05);
    }
    
    body.dark-theme pre,
    body.dark-theme code {
        background-color: #2d2d2d;
        border-color: #3e3e42;
    }
    
    body.dark-theme th {
        background-color: #2d2d2d;
    }
    
    body.dark-theme table,
    body.dark-theme th,
    body.dark-theme td {
        border-color: #3e3e42;
    }
}

/* 添加自定义CSS */
` + l.CustomCSS

	// 写入CSS文件
	cssPath := filepath.Join(outputDir, "style.css")
	if err := os.WriteFile(cssPath, []byte(baseCSS), 0644); err != nil {
		return err
	}

	return nil
}

// generateJS 生成JS文件
func (l *Layout) generateJS(outputDir string) error {
	// 基础JS
	baseJS := `
document.addEventListener('DOMContentLoaded', function() {
    // 移动端菜单切换
    const menuToggle = document.querySelector('.menu-toggle');
    const sidebar = document.querySelector('.sidebar');
    
    if (menuToggle) {
        menuToggle.addEventListener('click', function() {
            sidebar.classList.toggle('show');
        });
    }
    
    // 暗色模式切换
    const darkModeToggle = document.querySelector('.dark-mode-toggle');
    
    if (darkModeToggle) {
        darkModeToggle.addEventListener('click', function() {
            document.body.classList.toggle('dark-theme');
            const isDarkMode = document.body.classList.contains('dark-theme');
            localStorage.setItem('darkMode', isDarkMode ? 'enabled' : 'disabled');
        });
        
        // 检查之前是否启用了暗色模式
        if (localStorage.getItem('darkMode') === 'enabled') {
            document.body.classList.add('dark-theme');
        }
    }
    
    // 自动检测系统主题
    const prefersDarkScheme = window.matchMedia('(prefers-color-scheme: dark)');
    
    if (prefersDarkScheme.matches && localStorage.getItem('darkMode') === null) {
        document.body.classList.add('dark-theme');
    }
    
    // 搜索功能
    const searchInput = document.querySelector('.search-input');
    
    if (searchInput) {
        searchInput.addEventListener('input', function() {
            const query = this.value.toLowerCase();
            
            if (query.length < 2) {
                // 重置搜索结果
                document.querySelectorAll('.sidebar-nav-item').forEach(item => {
                    item.style.display = '';
                });
                return;
            }
            
            // 搜索导航项
            document.querySelectorAll('.sidebar-nav-item').forEach(item => {
                const text = item.textContent.toLowerCase();
                item.style.display = text.includes(query) ? '' : 'none';
            });
        });
    }
});

// 导航项折叠/展开
function toggleNav(id) {
    const children = document.getElementById(id);
    if (children) {
        children.style.display = children.style.display === 'none' ? 'block' : 'none';
        
        // 切换箭头图标
        const arrow = document.querySelector('.nav-arrow[data-id="' + id + '"]');
        if (arrow) {
            arrow.textContent = children.style.display === 'none' ? '▶' : '▼';
        }
    }
}

// 高亮当前页面的导航项
function highlightCurrentPage() {
    const currentPath = window.location.pathname;
    document.querySelectorAll('.sidebar-nav-item').forEach(item => {
        if (item.getAttribute('href') === currentPath) {
            item.classList.add('active');
            
            // 展开父级
            const parent = item.closest('.sidebar-nav-item-children');
            if (parent) {
                parent.style.display = 'block';
                const id = parent.id;
                const arrow = document.querySelector('.nav-arrow[data-id="' + id + '"]');
                if (arrow) {
                    arrow.textContent = '▼';
                }
            }
        }
    });
}

// 页面加载时初始化
window.addEventListener('load', highlightCurrentPage);

` + l.CustomJS

	// 写入JS文件
	jsPath := filepath.Join(outputDir, "script.js")
	if err := os.WriteFile(jsPath, []byte(baseJS), 0644); err != nil {
		return err
	}

	return nil
}

// generateTemplates 生成模板文件
func (l *Layout) generateTemplates(outputDir string) error {
	// 基础模板
	baseTemplate := `
<!DOCTYPE html>
<html lang="zh">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{ .Title }}</title>
    
    <!-- CSS文件 -->
    <link rel="stylesheet" href="{{ .BaseURL }}/assets/style.css">
    {{ range .CSSFiles }}
    <link rel="stylesheet" href="{{ . }}">
    {{ end }}
    
    {{ if .GoogleAnalyticsID }}
    <!-- Google Analytics -->
    <script async src="https://www.googletagmanager.com/gtag/js?id={{ .GoogleAnalyticsID }}"></script>
    <script>
        window.dataLayer = window.dataLayer || [];
        function gtag(){dataLayer.push(arguments);}
        gtag('js', new Date());
        gtag('config', '{{ .GoogleAnalyticsID }}');
    </script>
    {{ end }}
</head>
<body>
    <!-- 移动端菜单按钮 -->
    <button class="menu-toggle" style="display: none;">☰</button>
    
    <div class="container">
        <!-- 侧边栏 -->
        <div class="sidebar">
            <div class="sidebar-header">
                {{ if .LogoURL }}
                <img src="{{ .LogoURL }}" alt="{{ .ProjectName }}" class="sidebar-logo" />
                {{ end }}
                <h1 class="sidebar-title">{{ .ProjectName }}</h1>
                <div class="sidebar-version">{{ .Version }}</div>
            </div>
            
            {{ if .EnableSearch }}
            <div class="search-container">
                <input type="text" class="search-input" placeholder="搜索..." />
            </div>
            {{ end }}
            
            <nav class="sidebar-nav">
                {{ template "nav-items" .NavItems }}
            </nav>
            
            <div class="sidebar-footer">
                <button class="dark-mode-toggle">切换主题</button>
            </div>
        </div>
        
        <!-- 内容区域 -->
        <div class="content">
            <div class="content-header">
                <h1 class="content-title">{{ .DocTitle }}</h1>
            </div>
            
            <div class="content-main">
                {{ block "content" . }}{{ end }}
            </div>
            
            <footer class="footer">
                {{ .Footer }}
            </footer>
        </div>
    </div>
    
    <!-- JS文件 -->
    <script src="{{ .BaseURL }}/assets/script.js"></script>
    {{ range .JSFiles }}
    <script src="{{ . }}"></script>
    {{ end }}
</body>
</html>

{{ define "nav-items" }}
    {{ range . }}
        {{ if .Children }}
            <div class="sidebar-nav-item" onclick="toggleNav('{{ .Title }}-children')">
                {{ if .Icon }}<span class="sidebar-nav-icon">{{ .Icon }}</span>{{ end }}
                {{ .Title }}
                <span class="nav-arrow" data-id="{{ .Title }}-children">{{ if .Expanded }}▼{{ else }}▶{{ end }}</span>
            </div>
            <div id="{{ .Title }}-children" class="sidebar-nav-item-children" style="display: {{ if .Expanded }}block{{ else }}none{{ end }};">
                {{ template "nav-items" .Children }}
            </div>
        {{ else }}
            <a href="{{ .URL }}" class="sidebar-nav-item {{ if .Active }}active{{ end }}">
                {{ if .Icon }}<span class="sidebar-nav-icon">{{ .Icon }}</span>{{ end }}
                {{ .Title }}
            </a>
        {{ end }}
    {{ end }}
{{ end }}
`

	// 写入基础模板
	templatePath := filepath.Join(outputDir, "templates", "base.tmpl")
	if err := os.MkdirAll(filepath.Dir(templatePath), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(templatePath, []byte(baseTemplate), 0644); err != nil {
		return err
	}

	// 内容模板
	contentTemplate := `
{{ define "content" }}
<div class="content-body">
    {{ .Content }}
</div>
{{ end }}
`

	// 写入内容模板
	contentTemplatePath := filepath.Join(outputDir, "templates", "content.tmpl")
	if err := os.WriteFile(contentTemplatePath, []byte(contentTemplate), 0644); err != nil {
		return err
	}

	return nil
}

// RenderPage 渲染页面
func (l *Layout) RenderPage(content template.HTML, outputPath string) error {
	// 创建要传递给模板的数据
	data := struct {
		*Layout
		Content template.HTML
	}{
		Layout:  l,
		Content: content,
	}

	// 创建模板
	tmpl, err := template.ParseFiles(
		filepath.Join(l.OutputDir, "templates", "base.tmpl"),
		filepath.Join(l.OutputDir, "templates", "content.tmpl"),
	)
	if err != nil {
		return fmt.Errorf("解析模板失败: %w", err)
	}

	// 创建输出文件
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %w", err)
	}
	defer file.Close()

	// 执行模板
	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("执行模板失败: %w", err)
	}

	return nil
}
