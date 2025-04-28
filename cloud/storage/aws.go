package storage

import (
	"context"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// AWSStorage AWS S3存储实现
type AWSStorage struct {
	client *s3.Client
	region string
}

// NewAWSStorage 创建AWS S3存储客户端
func NewAWSStorage(cfg Config) (*AWSStorage, error) {
	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKey,
			cfg.SecretKey,
			"",
		)),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(awsCfg)

	return &AWSStorage{
		client: client,
		region: cfg.Region,
	}, nil
}

// Upload 上传文件到AWS S3
func (s *AWSStorage) Upload(ctx context.Context, bucketName, objectKey string, reader io.Reader, contentType string) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(objectKey),
		Body:        reader,
		ContentType: aws.String(contentType),
	})
	return err
}

// Download 从AWS S3下载文件
func (s *AWSStorage) Download(ctx context.Context, bucketName, objectKey string) (io.ReadCloser, error) {
	output, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return nil, err
	}
	return output.Body, nil
}

// Delete 从AWS S3删除文件
func (s *AWSStorage) Delete(ctx context.Context, bucketName, objectKey string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})
	return err
}

// GetURL 获取AWS S3文件URL
func (s *AWSStorage) GetURL(ctx context.Context, bucketName, objectKey string, expire time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s.client)

	presignResult, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expire
	})

	if err != nil {
		return "", err
	}

	return presignResult.URL, nil
}

// ListObjects 列出AWS S3存储桶中的对象
func (s *AWSStorage) ListObjects(ctx context.Context, bucketName, prefix string) ([]ObjectInfo, error) {
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	}

	if prefix != "" {
		input.Prefix = aws.String(prefix)
	}

	result, err := s.client.ListObjectsV2(ctx, input)
	if err != nil {
		return nil, err
	}

	var objects []ObjectInfo
	for _, obj := range result.Contents {
		objects = append(objects, ObjectInfo{
			Key:          *obj.Key,
			Size:         *obj.Size,
			LastModified: *obj.LastModified,
			ETag:         *obj.ETag,
			ContentType:  "", // S3 ListObjectsV2 不返回ContentType
		})
	}

	return objects, nil
}
