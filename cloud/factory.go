package cloud

import (
	"fmt"

	"github.com/zzliekkas/flow/cloud/storage"
	"github.com/zzliekkas/flow/config"
)

// 云服务提供商类型
const (
	ProviderAWS    = "aws"
	ProviderAliyun = "aliyun"
	ProviderHuawei = "huawei"
)

// NewStorageProvider 创建存储提供商实例
func NewStorageProvider(cfg *config.Config) (storage.Provider, error) {
	provider := cfg.GetString("cloud.provider")

	var providerType storage.ProviderType
	switch provider {
	case ProviderAWS:
		providerType = storage.ProviderAWS
	case ProviderAliyun:
		providerType = storage.ProviderAliyun
	case ProviderHuawei:
		providerType = storage.ProviderHuawei
	default:
		return nil, fmt.Errorf("unsupported cloud provider: %s", provider)
	}

	storageConfig := storage.Config{
		AccessKey: cfg.GetString(fmt.Sprintf("cloud.%s.accessKey", provider)),
		SecretKey: cfg.GetString(fmt.Sprintf("cloud.%s.secretKey", provider)),
		Region:    cfg.GetString(fmt.Sprintf("cloud.%s.region", provider)),
		Endpoint:  cfg.GetString(fmt.Sprintf("cloud.%s.endpoint", provider)),
	}

	factory := storage.NewFactory(providerType, storageConfig)
	return factory.Create()
}
