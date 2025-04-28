package storage

import (
	"fmt"
)

// ProviderType 定义云存储提供商类型
type ProviderType string

const (
	// ProviderAWS AWS S3存储
	ProviderAWS ProviderType = "aws"

	// ProviderAliyun 阿里云OSS存储
	ProviderAliyun ProviderType = "aliyun"

	// ProviderHuawei 华为云OBS存储
	ProviderHuawei ProviderType = "huawei"
)

// Factory 存储提供商工厂
type Factory struct {
	providerType ProviderType
	config       Config
}

// NewFactory 创建存储工厂
func NewFactory(providerType ProviderType, config Config) *Factory {
	return &Factory{
		providerType: providerType,
		config:       config,
	}
}

// Create 创建存储提供商实例
func (f *Factory) Create() (Provider, error) {
	switch f.providerType {
	case ProviderAWS:
		return NewAWSStorage(f.config)
	case ProviderAliyun:
		return NewAliyunStorage(f.config)
	case ProviderHuawei:
		return NewHuaweiStorage(f.config)
	default:
		return nil, fmt.Errorf("unsupported storage provider: %s", f.providerType)
	}
}
