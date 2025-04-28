package storage

import (
	"context"
	"io"
	"time"
)

// Provider 定义对象存储服务的通用接口
type Provider interface {
	// Upload 上传文件到云存储
	Upload(ctx context.Context, bucketName, objectKey string, reader io.Reader, contentType string) error

	// Download 从云存储下载文件
	Download(ctx context.Context, bucketName, objectKey string) (io.ReadCloser, error)

	// Delete 从云存储删除文件
	Delete(ctx context.Context, bucketName, objectKey string) error

	// GetURL 获取访问URL
	GetURL(ctx context.Context, bucketName, objectKey string, expire time.Duration) (string, error)

	// ListObjects 列出存储桶中的对象
	ListObjects(ctx context.Context, bucketName, prefix string) ([]ObjectInfo, error)
}

// ObjectInfo 表示存储对象的基本信息
type ObjectInfo struct {
	Key          string    // 对象键名
	Size         int64     // 对象大小（字节）
	LastModified time.Time // 最后修改时间
	ContentType  string    // 内容类型
	ETag         string    // ETag标识
}

// Config 存储配置基础结构
type Config struct {
	AccessKey string
	SecretKey string
	Region    string
	Endpoint  string
}
