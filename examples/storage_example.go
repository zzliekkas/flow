// Package examples 提供示例代码
package examples

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// StorageExample 展示基本的文件存储操作
func StorageExample() {
	// 在这里提供示例代码的占位符
	// 注意：实际的存储示例需要重构以避免导入循环
	fmt.Println("存储系统示例已经实现，但需要修复导入循环问题。")
	fmt.Println("请查看具体的实现文件了解更多信息。")
}

// LocalStorageExample 展示本地文件存储操作
func LocalStorageExample() {
	// 创建临时目录
	tempDir, err := ioutil.TempDir("", "storage-example")
	if err != nil {
		fmt.Printf("创建临时目录失败: %v\n", err)
		return
	}
	defer os.RemoveAll(tempDir)

	// 写入文件
	testFile := filepath.Join(tempDir, "test.txt")
	content := []byte("这是测试内容")
	err = ioutil.WriteFile(testFile, content, 0644)
	if err != nil {
		fmt.Printf("写入文件失败: %v\n", err)
		return
	}
	fmt.Printf("成功写入文件: %s\n", testFile)

	// 读取文件
	readContent, err := ioutil.ReadFile(testFile)
	if err != nil {
		fmt.Printf("读取文件失败: %v\n", err)
		return
	}
	fmt.Printf("文件内容: %s\n", readContent)

	// 删除文件
	err = os.Remove(testFile)
	if err != nil {
		fmt.Printf("删除文件失败: %v\n", err)
		return
	}
	fmt.Println("成功删除文件")
}
