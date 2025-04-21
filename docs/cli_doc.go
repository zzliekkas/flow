package docs

import (
	"github.com/zzliekkas/flow/app"
)

// CLIDocGenerator 用于生成命令行工具文档的生成器
type CLIDocGenerator struct {
	app       *app.Application
	outputDir string
}

// NewCLIDocGenerator 创建新的CLI文档生成器
func NewCLIDocGenerator(application *app.Application) *CLIDocGenerator {
	return &CLIDocGenerator{
		app:       application,
		outputDir: "./docs/cli",
	}
}

// SetOutputDir 设置输出目录
func (g *CLIDocGenerator) SetOutputDir(dir string) *CLIDocGenerator {
	g.outputDir = dir
	return g
}

// Generate 生成CLI文档
func (g *CLIDocGenerator) Generate() error {
	// 这里是生成CLI文档的实际逻辑
	// 暂时返回nil，表示没有错误
	return nil
}
