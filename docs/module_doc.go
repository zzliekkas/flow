package docs

import (
	"github.com/zzliekkas/flow/app"
)

// ModuleDocGenerator 用于生成模块文档的生成器
type ModuleDocGenerator struct {
	app       *app.Application
	outputDir string
}

// NewModuleDocGenerator 创建新的模块文档生成器
func NewModuleDocGenerator(application *app.Application) *ModuleDocGenerator {
	return &ModuleDocGenerator{
		app:       application,
		outputDir: "./docs/modules",
	}
}

// SetOutputDir 设置输出目录
func (g *ModuleDocGenerator) SetOutputDir(dir string) *ModuleDocGenerator {
	g.outputDir = dir
	return g
}

// Generate 生成模块文档
func (g *ModuleDocGenerator) Generate() error {
	// 这里是生成模块文档的实际逻辑
	// 暂时返回nil，表示没有错误
	return nil
}
