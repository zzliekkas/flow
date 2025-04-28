# Flow框架云服务模块

Flow框架云服务模块提供了统一的云服务抽象层，支持多种云服务提供商。

## 支持的云服务提供商

当前支持以下云服务提供商：

- 阿里云（Aliyun）
- AWS（亚马逊云）
- 华为云（Huawei Cloud）

## 快速开始

### 1. 注册云服务提供者

在应用初始化时，注册云服务提供者：

```go
package main

import (
    "github.com/zzliekkas/flow"
    "github.com/zzliekkas/flow/app"
    "github.com/zzliekkas/flow/cloud"
)

func main() {
    engine := flow.New()
    application := app.New(engine)
    
    // 注册云服务提供者
    application.RegisterProvider(cloud.NewCloudServiceProvider())
    
    // 启动应用
    application.Run(":8080")
}
```

### 2. 配置云服务

创建配置文件 `config/cloud.yaml`：

```yaml
cloud:
  # 当前使用的云服务提供商: aws, aliyun, huawei
  provider: "aliyun"
  
  # AWS S3配置
  aws:
    region: "ap-east-1"
    endpoint: ""
    accessKey: "${AWS_ACCESS_KEY}"
    secretKey: "${AWS_SECRET_KEY}"
    
  # 阿里云OSS配置
  aliyun:
    region: "cn-hangzhou"
    endpoint: "oss-cn-hangzhou.aliyuncs.com"
    accessKey: "${ALIYUN_ACCESS_KEY}"
    secretKey: "${ALIYUN_SECRET_KEY}"
    
  # 华为云OBS配置
  huawei:
    region: "cn-north-4"
    endpoint: "obs.cn-north-4.myhuaweicloud.com"
    accessKey: "${HUAWEI_ACCESS_KEY}"
    secretKey: "${HUAWEI_SECRET_KEY}"
```

### 3. 使用云存储服务

在服务中使用云存储服务：

```go
package services

import (
    "context"
    "io"
    
    "github.com/zzliekkas/flow/cloud/storage"
)

type FileService struct {
    storage storage.Provider // 注入存储提供商
}

// 构造函数，通过依赖注入获取存储提供商
func NewFileService(storageProvider storage.Provider) *FileService {
    return &FileService{
        storage: storageProvider,
    }
}

// 上传文件
func (s *FileService) UploadFile(ctx context.Context, bucketName, objectKey string, reader io.Reader, contentType string) error {
    return s.storage.Upload(ctx, bucketName, objectKey, reader, contentType)
}

// 获取文件URL
func (s *FileService) GetFileURL(ctx context.Context, bucketName, objectKey string, expireTime time.Duration) (string, error) {
    return s.storage.GetURL(ctx, bucketName, objectKey, expireTime)
}

// 下载文件
func (s *FileService) DownloadFile(ctx context.Context, bucketName, objectKey string) (io.ReadCloser, error) {
    return s.storage.Download(ctx, bucketName, objectKey)
}

// 删除文件
func (s *FileService) DeleteFile(ctx context.Context, bucketName, objectKey string) error {
    return s.storage.Delete(ctx, bucketName, objectKey)
}

// 列出文件
func (s *FileService) ListFiles(ctx context.Context, bucketName, prefix string) ([]storage.ObjectInfo, error) {
    return s.storage.ListObjects(ctx, bucketName, prefix)
}
```

## 多云部署注意事项

在多云环境中部署应用时的最佳实践：

1. **通过配置选择云提供商**
   - 为不同环境创建不同的配置文件
   - 使用环境变量替换敏感凭证

2. **使用统一的桶（Bucket）命名策略**
   - 确保不同云环境使用相同的桶名称逻辑
   - 例如：`{项目名}-{环境}-{用途}`

3. **处理云服务特有功能**
   - 避免在业务代码中使用特定云服务的独有特性
   - 如需使用，通过条件判断或策略模式扩展抽象接口

4. **错误处理和重试策略**
   - 实现统一的错误处理机制
   - 针对不同云环境设置合适的重试策略

5. **多云同步和容灾**
   - 考虑关键数据在多云之间的同步机制
   - 实现跨云服务的灾难恢复策略

## 扩展支持更多云服务提供商

要添加新的云服务提供商支持，只需实现对应的接口：

1. 实现 `storage.Provider` 接口
2. 在 `storage/factory.go` 中注册新的提供商
3. 更新配置结构

详细示例可参考现有的实现代码。 