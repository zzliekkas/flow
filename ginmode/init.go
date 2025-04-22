// Package ginmode 确保在gin包初始化前设置正确的模式
package ginmode

import (
	"os"
)

func init() {
	// 检查是否已存在用户设置的GIN_MODE或FLOW_MODE
	ginMode := os.Getenv("GIN_MODE")
	flowMode := os.Getenv("FLOW_MODE")

	// 如果用户已经设置了GIN_MODE，尊重用户设置
	if ginMode != "" {
		return
	}

	// 如果用户设置了FLOW_MODE，则使用它来设置GIN_MODE，保持一致性
	if flowMode != "" {
		// 将flow模式映射到gin模式
		switch flowMode {
		case "release", "production":
			os.Setenv("GIN_MODE", "release")
		case "test":
			os.Setenv("GIN_MODE", "test")
		case "debug", "development":
			os.Setenv("GIN_MODE", "debug")
		default:
			// 对于未知模式，默认使用release避免警告
			os.Setenv("GIN_MODE", "release")
		}
		return
	}

	// 如果没有设置任何模式，默认使用release模式避免警告
	// 但不影响用户后续可以通过WithMode()或配置文件修改
	os.Setenv("GIN_MODE", "release")
}
