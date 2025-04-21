package examples

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/zzliekkas/flow/storage"
	"github.com/zzliekkas/flow/storage/cloud"
)

// CloudStorageDriverExample 展示云存储的基本用法
func CloudStorageDriverExample() {
	// 创建存储管理器
	manager := storage.NewManager()

	// 创建云存储管理器
	cloudManager := cloud.NewCloudManager()

	// 注册云存储驱动到管理器
	err := cloud.RegisterToManager(manager, cloudManager)
	if err != nil {
		fmt.Printf("注册云存储驱动失败: %v\n", err)
		return
	}

	// 尝试配置并使用S3驱动
	if os.Getenv("S3_ENDPOINT") != "" {
		s3Config := cloud.S3Config{
			Endpoint:  os.Getenv("S3_ENDPOINT"),
			Region:    os.Getenv("S3_REGION"),
			Bucket:    os.Getenv("S3_BUCKET"),
			AccessKey: os.Getenv("S3_ACCESS_KEY"),
			SecretKey: os.Getenv("S3_SECRET_KEY"),
			UseSSL:    strings.ToLower(os.Getenv("S3_USE_SSL")) == "true",
			PublicURL: os.Getenv("S3_PUBLIC_URL"),
		}

		s3Driver, err := cloud.New(s3Config)
		if err != nil {
			fmt.Printf("创建S3驱动失败: %v\n", err)
		} else {
			// 注册S3驱动到云存储管理器
			err = cloudManager.RegisterCloud("s3", s3Driver)
			if err != nil {
				fmt.Printf("注册S3驱动失败: %v\n", err)
			} else {
				// 使用S3存储
				s3Example(manager)
			}
		}
	}

	// 尝试配置并使用OSS驱动
	if os.Getenv("OSS_ENDPOINT") != "" {
		ossConfig := cloud.OSSConfig{
			Endpoint:        os.Getenv("OSS_ENDPOINT"),
			AccessKeyID:     os.Getenv("OSS_ACCESS_KEY_ID"),
			AccessKeySecret: os.Getenv("OSS_ACCESS_KEY_SECRET"),
			Bucket:          os.Getenv("OSS_BUCKET"),
			UseSSL:          strings.ToLower(os.Getenv("OSS_USE_SSL")) == "true",
			PublicURL:       os.Getenv("OSS_PUBLIC_URL"),
		}

		ossDriver, err := cloud.NewOSS(ossConfig)
		if err != nil {
			fmt.Printf("创建OSS驱动失败: %v\n", err)
		} else {
			// 注册OSS驱动到云存储管理器
			// 使用core.FileSystem类型的适配器
			coreAdapter := &storage.StorageToCoreFSAdapter{StorageFS: ossDriver}
			err = cloudManager.RegisterCloud("oss", coreAdapter)
			if err != nil {
				fmt.Printf("注册OSS驱动失败: %v\n", err)
			} else {
				// 使用OSS存储
				ossExample(manager)
			}
		}
	}

	// 如果没有配置任何云存储驱动
	if !cloudManager.HasCloud("s3") && !cloudManager.HasCloud("oss") {
		fmt.Println("未配置任何云存储驱动。请设置以下环境变量：")
		fmt.Println("S3驱动: S3_ENDPOINT, S3_REGION, S3_BUCKET, S3_ACCESS_KEY, S3_SECRET_KEY")
		fmt.Println("OSS驱动: OSS_ENDPOINT, OSS_ACCESS_KEY_ID, OSS_ACCESS_KEY_SECRET, OSS_BUCKET")
	}
}

// s3Example 展示S3存储的使用方法
func s3Example(manager *storage.Manager) {
	ctx := context.Background()

	// 获取S3驱动
	s3, err := manager.Disk("cloud.s3")
	if err != nil {
		log.Fatalf("获取S3驱动失败: %v", err)
	}

	// 创建测试目录
	dirPath := "examples/test-" + fmt.Sprintf("%d", time.Now().Unix())
	err = s3.CreateDirectory(ctx, dirPath)
	if err != nil {
		log.Fatalf("创建目录失败: %v", err)
	}
	fmt.Printf("成功在S3创建目录: %s\n", dirPath)

	// 写入文本文件
	filePath := dirPath + "/example.txt"
	content := []byte("这是S3上的测试文件内容。")
	err = s3.Write(ctx, filePath, content, storage.WithVisibility("public"))
	if err != nil {
		log.Fatalf("写入文件失败: %v", err)
	}
	fmt.Printf("成功写入文件到S3: %s\n", filePath)

	// 获取文件信息
	file, err := s3.Get(ctx, filePath)
	if err != nil {
		log.Fatalf("获取文件失败: %v", err)
	}
	fmt.Printf("S3文件信息:\n")
	fmt.Printf("  路径: %s\n", file.Path())
	fmt.Printf("  名称: %s\n", file.Name())
	fmt.Printf("  大小: %d 字节\n", file.Size())
	fmt.Printf("  类型: %s\n", file.MimeType())
	fmt.Printf("  可见性: %s\n", file.Visibility())
	fmt.Printf("  URL: %s\n", file.URL())

	// 生成临时URL
	tempURL, err := s3.TemporaryURL(ctx, filePath, 1*time.Hour)
	if err != nil {
		log.Printf("生成临时URL失败: %v", err)
	} else {
		fmt.Printf("  临时URL (1小时有效): %s\n", tempURL)
	}

	// 读取文件内容
	data, err := s3.Get(ctx, filePath)
	if err != nil {
		log.Fatalf("读取文件失败: %v", err)
	}

	reader, err := data.ReadStream(ctx)
	if err != nil {
		log.Fatalf("获取读取流失败: %v", err)
	}
	defer reader.Close()

	readData, err := io.ReadAll(reader)
	if err != nil {
		log.Fatalf("读取流失败: %v", err)
	}
	fmt.Printf("  内容: %s\n", string(readData))

	// 复制文件
	copyPath := dirPath + "/example-copy.txt"
	err = s3.Copy(ctx, filePath, copyPath)
	if err != nil {
		log.Fatalf("复制文件失败: %v", err)
	}
	fmt.Printf("成功复制文件到: %s\n", copyPath)

	// 列出目录中的文件
	files, err := s3.Files(ctx, dirPath)
	if err != nil {
		log.Fatalf("列出文件失败: %v", err)
	}
	fmt.Printf("目录 %s 中的文件:\n", dirPath)
	for _, f := range files {
		fmt.Printf("  - %s (%d 字节)\n", f.Name(), f.Size())
	}

	// 计算校验和
	checksum, err := s3.Checksum(ctx, filePath, "md5")
	if err != nil {
		log.Printf("计算校验和失败: %v", err)
	} else {
		fmt.Printf("文件MD5校验和: %s\n", checksum)
	}

	// 设置可见性为私有
	err = s3.SetVisibility(ctx, filePath, "private")
	if err != nil {
		log.Printf("设置可见性失败: %v", err)
	}
	fmt.Println("已将文件设置为私有")

	// 再次获取文件以检查可见性
	file, err = s3.Get(ctx, filePath)
	if err != nil {
		log.Fatalf("获取文件失败: %v", err)
	}
	fmt.Printf("文件可见性现在是: %s\n", file.Visibility())

	// 删除文件
	fmt.Println("正在删除测试文件...")
	err = s3.Delete(ctx, filePath)
	if err != nil {
		log.Fatalf("删除文件失败: %v", err)
	}
	err = s3.Delete(ctx, copyPath)
	if err != nil {
		log.Fatalf("删除文件失败: %v", err)
	}

	// 删除目录
	fmt.Println("正在删除测试目录...")
	err = s3.DeleteDirectory(ctx, dirPath)
	if err != nil {
		log.Fatalf("删除目录失败: %v", err)
	}

	fmt.Println("S3存储示例完成")
}

// ossExample 展示OSS存储的使用方法
func ossExample(manager *storage.Manager) {
	ctx := context.Background()

	// 获取OSS驱动
	oss, err := manager.Disk("cloud.oss")
	if err != nil {
		log.Fatalf("获取OSS驱动失败: %v", err)
	}

	// 创建测试目录
	dirPath := "examples/test-" + fmt.Sprintf("%d", time.Now().Unix())
	err = oss.CreateDirectory(ctx, dirPath)
	if err != nil {
		log.Fatalf("创建目录失败: %v", err)
	}
	fmt.Printf("成功在OSS创建目录: %s\n", dirPath)

	// 写入文本文件
	filePath := dirPath + "/example.txt"
	content := []byte("这是OSS上的测试文件内容。")
	err = oss.Write(ctx, filePath, content, storage.WithVisibility("public"))
	if err != nil {
		log.Fatalf("写入文件失败: %v", err)
	}
	fmt.Printf("成功写入文件到OSS: %s\n", filePath)

	// 获取文件信息
	file, err := oss.Get(ctx, filePath)
	if err != nil {
		log.Fatalf("获取文件失败: %v", err)
	}
	fmt.Printf("OSS文件信息:\n")
	fmt.Printf("  路径: %s\n", file.Path())
	fmt.Printf("  名称: %s\n", file.Name())
	fmt.Printf("  大小: %d 字节\n", file.Size())
	fmt.Printf("  类型: %s\n", file.MimeType())
	fmt.Printf("  可见性: %s\n", file.Visibility())
	fmt.Printf("  URL: %s\n", file.URL())

	// 生成临时URL
	tempURL, err := oss.TemporaryURL(ctx, filePath, 1*time.Hour)
	if err != nil {
		log.Printf("生成临时URL失败: %v", err)
	} else {
		fmt.Printf("  临时URL (1小时有效): %s\n", tempURL)
	}

	// 读取文件内容
	data, err := oss.Get(ctx, filePath)
	if err != nil {
		log.Fatalf("读取文件失败: %v", err)
	}

	reader, err := data.ReadStream(ctx)
	if err != nil {
		log.Fatalf("获取读取流失败: %v", err)
	}
	defer reader.Close()

	readData, err := io.ReadAll(reader)
	if err != nil {
		log.Fatalf("读取流失败: %v", err)
	}
	fmt.Printf("  内容: %s\n", string(readData))

	// 复制文件
	copyPath := dirPath + "/example-copy.txt"
	err = oss.Copy(ctx, filePath, copyPath)
	if err != nil {
		log.Fatalf("复制文件失败: %v", err)
	}
	fmt.Printf("成功复制文件到: %s\n", copyPath)

	// 列出目录中的文件
	files, err := oss.Files(ctx, dirPath)
	if err != nil {
		log.Fatalf("列出文件失败: %v", err)
	}
	fmt.Printf("目录 %s 中的文件:\n", dirPath)
	for _, f := range files {
		fmt.Printf("  - %s (%d 字节)\n", f.Name(), f.Size())
	}

	// 计算校验和
	checksum, err := oss.Checksum(ctx, filePath, "md5")
	if err != nil {
		log.Printf("计算校验和失败: %v", err)
	} else {
		fmt.Printf("文件MD5校验和: %s\n", checksum)
	}

	// 设置可见性为私有
	err = oss.SetVisibility(ctx, filePath, "private")
	if err != nil {
		log.Printf("设置可见性失败: %v", err)
	}
	fmt.Println("已将文件设置为私有")

	// 再次获取文件以检查可见性
	file, err = oss.Get(ctx, filePath)
	if err != nil {
		log.Fatalf("获取文件失败: %v", err)
	}
	fmt.Printf("文件可见性现在是: %s\n", file.Visibility())

	// 删除文件
	fmt.Println("正在删除测试文件...")
	err = oss.Delete(ctx, filePath)
	if err != nil {
		log.Fatalf("删除文件失败: %v", err)
	}
	err = oss.Delete(ctx, copyPath)
	if err != nil {
		log.Fatalf("删除文件失败: %v", err)
	}

	// 删除目录
	fmt.Println("正在删除测试目录...")
	err = oss.DeleteDirectory(ctx, dirPath)
	if err != nil {
		log.Fatalf("删除目录失败: %v", err)
	}

	fmt.Println("OSS存储示例完成")
}
