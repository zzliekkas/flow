package docs

import (
	"github.com/zzliekkas/flow/app"
)

// ConfigDocGenerator 用于生成配置文档的生成器
type ConfigDocGenerator struct {
	app       *app.Application
	outputDir string
}

// NewConfigDocGenerator 创建新的配置文档生成器
func NewConfigDocGenerator(application *app.Application) *ConfigDocGenerator {
	return &ConfigDocGenerator{
		app:       application,
		outputDir: "./docs/config",
	}
}

// SetOutputDir 设置输出目录
func (g *ConfigDocGenerator) SetOutputDir(dir string) *ConfigDocGenerator {
	g.outputDir = dir
	return g
}

// Generate 生成配置文档
func (g *ConfigDocGenerator) Generate() error {
	// 这里是生成配置文档的实际逻辑
	// 暂时返回nil，表示没有错误
	return nil
}
