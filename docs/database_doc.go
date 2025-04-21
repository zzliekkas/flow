package docs

import (
	"github.com/zzliekkas/flow/app"
)

// DatabaseDocGenerator 用于生成数据库文档的生成器
type DatabaseDocGenerator struct {
	app       *app.Application
	outputDir string
}

// NewDatabaseDocGenerator 创建新的数据库文档生成器
func NewDatabaseDocGenerator(application *app.Application) *DatabaseDocGenerator {
	return &DatabaseDocGenerator{
		app:       application,
		outputDir: "./docs/database",
	}
}

// SetOutputDir 设置输出目录
func (g *DatabaseDocGenerator) SetOutputDir(dir string) *DatabaseDocGenerator {
	g.outputDir = dir
	return g
}

// Generate 生成数据库文档
func (g *DatabaseDocGenerator) Generate() error {
	// 这里是生成数据库文档的实际逻辑
	// 暂时返回nil，表示没有错误
	return nil
}
