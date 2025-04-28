package examples

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/zzliekkas/flow/cloud/storage"
)

// StorageExample 展示如何使用云存储服务
func StorageExample(provider storage.Provider) error {
	ctx := context.Background()

	// 示例参数
	bucketName := "my-test-bucket"
	objectKey := "test-file.txt"
	content := "这是一个测试文件内容"

	// 上传文件
	err := uploadExample(ctx, provider, bucketName, objectKey, content)
	if err != nil {
		return fmt.Errorf("upload failed: %v", err)
	}

	// 获取URL
	url, err := getURLExample(ctx, provider, bucketName, objectKey)
	if err != nil {
		return fmt.Errorf("get URL failed: %v", err)
	}
	fmt.Printf("File URL: %s\n", url)

	// 下载文件
	downloadedContent, err := downloadExample(ctx, provider, bucketName, objectKey)
	if err != nil {
		return fmt.Errorf("download failed: %v", err)
	}
	fmt.Printf("Downloaded content: %s\n", downloadedContent)

	// 列出对象
	err = listObjectsExample(ctx, provider, bucketName, "")
	if err != nil {
		return fmt.Errorf("list objects failed: %v", err)
	}

	// 删除文件
	err = deleteExample(ctx, provider, bucketName, objectKey)
	if err != nil {
		return fmt.Errorf("delete failed: %v", err)
	}

	return nil
}

// 上传文件示例
func uploadExample(ctx context.Context, provider storage.Provider, bucketName, objectKey, content string) error {
	reader := strings.NewReader(content)
	return provider.Upload(ctx, bucketName, objectKey, reader, "text/plain")
}

// 获取文件URL示例
func getURLExample(ctx context.Context, provider storage.Provider, bucketName, objectKey string) (string, error) {
	// 生成一个1小时有效的URL
	return provider.GetURL(ctx, bucketName, objectKey, 1*time.Hour)
}

// 下载文件示例
func downloadExample(ctx context.Context, provider storage.Provider, bucketName, objectKey string) (string, error) {
	reader, err := provider.Download(ctx, bucketName, objectKey)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// 列出对象示例
func listObjectsExample(ctx context.Context, provider storage.Provider, bucketName, prefix string) error {
	objects, err := provider.ListObjects(ctx, bucketName, prefix)
	if err != nil {
		return err
	}

	fmt.Printf("Objects in bucket '%s' with prefix '%s':\n", bucketName, prefix)
	for _, obj := range objects {
		fmt.Printf("  - %s (size: %d bytes, last modified: %s)\n",
			obj.Key, obj.Size, obj.LastModified.Format(time.RFC3339))
	}

	return nil
}

// 删除文件示例
func deleteExample(ctx context.Context, provider storage.Provider, bucketName, objectKey string) error {
	return provider.Delete(ctx, bucketName, objectKey)
}
