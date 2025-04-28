package storage

import (
	"context"
	"io"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// AliyunStorage 阿里云OSS存储实现
type AliyunStorage struct {
	client   *oss.Client
	endpoint string
}

// NewAliyunStorage 创建阿里云存储客户端
func NewAliyunStorage(config Config) (*AliyunStorage, error) {
	client, err := oss.New(config.Endpoint, config.AccessKey, config.SecretKey)
	if err != nil {
		return nil, err
	}

	return &AliyunStorage{
		client:   client,
		endpoint: config.Endpoint,
	}, nil
}

// Upload 上传文件到阿里云OSS
func (s *AliyunStorage) Upload(ctx context.Context, bucketName, objectKey string, reader io.Reader, contentType string) error {
	bucket, err := s.client.Bucket(bucketName)
	if err != nil {
		return err
	}

	options := []oss.Option{
		oss.ContentType(contentType),
	}

	return bucket.PutObject(objectKey, reader, options...)
}

// Download 从阿里云OSS下载文件
func (s *AliyunStorage) Download(ctx context.Context, bucketName, objectKey string) (io.ReadCloser, error) {
	bucket, err := s.client.Bucket(bucketName)
	if err != nil {
		return nil, err
	}

	return bucket.GetObject(objectKey)
}

// Delete 从阿里云OSS删除文件
func (s *AliyunStorage) Delete(ctx context.Context, bucketName, objectKey string) error {
	bucket, err := s.client.Bucket(bucketName)
	if err != nil {
		return err
	}

	return bucket.DeleteObject(objectKey)
}

// GetURL 获取阿里云OSS文件URL
func (s *AliyunStorage) GetURL(ctx context.Context, bucketName, objectKey string, expire time.Duration) (string, error) {
	bucket, err := s.client.Bucket(bucketName)
	if err != nil {
		return "", err
	}

	signedURL, err := bucket.SignURL(objectKey, oss.HTTPGet, int64(expire.Seconds()))
	if err != nil {
		return "", err
	}

	return signedURL, nil
}

// ListObjects 列出阿里云OSS存储桶中的对象
func (s *AliyunStorage) ListObjects(ctx context.Context, bucketName, prefix string) ([]ObjectInfo, error) {
	bucket, err := s.client.Bucket(bucketName)
	if err != nil {
		return nil, err
	}

	lsRes, err := bucket.ListObjects(oss.Prefix(prefix))
	if err != nil {
		return nil, err
	}

	var objects []ObjectInfo
	for _, object := range lsRes.Objects {
		objects = append(objects, ObjectInfo{
			Key:          object.Key,
			Size:         object.Size,
			LastModified: object.LastModified,
			ETag:         object.ETag,
			ContentType:  object.Type,
		})
	}

	return objects, nil
}
