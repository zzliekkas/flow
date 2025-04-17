// Package test 提供测试支持工具和辅助函数
package test

import (
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// Helper 测试助手结构
type Helper struct {
	T *testing.T
}

// NewHelper 创建一个新的测试助手实例
func NewHelper(t *testing.T) *Helper {
	return &Helper{T: t}
}

// GetProjectRoot 获取项目根目录路径
func (h *Helper) GetProjectRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	return filepath.Dir(dir) // 返回test目录的父目录，即项目根目录
}

// LoadFixtureFile 加载测试数据文件
func (h *Helper) LoadFixtureFile(relativePath string) []byte {
	root := h.GetProjectRoot()
	path := filepath.Join(root, "test", "fixtures", relativePath)
	data, err := os.ReadFile(path)
	require.NoError(h.T, err, "无法读取测试数据文件: %s", path)
	return data
}

// TempFile 创建临时文件并返回其路径，测试完成后会自动清理
func (h *Helper) TempFile(pattern string, content []byte) string {
	tmpFile, err := os.CreateTemp("", pattern)
	require.NoError(h.T, err, "无法创建临时文件")

	defer tmpFile.Close()

	_, err = tmpFile.Write(content)
	require.NoError(h.T, err, "无法写入临时文件内容")

	h.T.Cleanup(func() {
		_ = os.Remove(tmpFile.Name())
	})

	return tmpFile.Name()
}

// TempDir 创建临时目录并返回其路径，测试完成后会自动清理
func (h *Helper) TempDir() string {
	dir, err := os.MkdirTemp("", "flow-test-*")
	require.NoError(h.T, err, "无法创建临时目录")

	h.T.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})

	return dir
}

// RandomString 生成指定长度的随机字符串
func (h *Helper) RandomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

// SetupTestDB 设置测试数据库连接
// 注意：此函数仅为示例，实际实现需要进行适当修改
// 通常的实现方式为：
// 1. 创建临时数据库文件
// 2. 构建数据库连接字符串
// 3. 建立数据库连接
// 4. 根据需要进行数据库初始化
func (h *Helper) SetupTestDB() *gorm.DB {
	// 由于这是示例实现，我们返回nil
	// 在实际使用时，你需要返回真实的数据库连接
	return nil
}

// CleanupTestDB 清理测试数据库
func (h *Helper) CleanupTestDB(db *gorm.DB) {
	// 实际实现取决于所使用的数据库类型
	sqlDB, err := db.DB()
	if err == nil {
		_ = sqlDB.Close()
	}
}

// AssertContainsAll 断言一个字符串包含所有指定的子字符串
func (h *Helper) AssertContainsAll(str string, substrings ...string) {
	for _, substring := range substrings {
		require.True(h.T, strings.Contains(str, substring),
			"期望字符串包含 %q，但不包含", substring)
	}
}

// AssertNotContainsAny 断言一个字符串不包含任何指定的子字符串
func (h *Helper) AssertNotContainsAny(str string, substrings ...string) {
	for _, substring := range substrings {
		require.False(h.T, strings.Contains(str, substring),
			"期望字符串不包含 %q，但包含了", substring)
	}
}

// RunParallel 并行运行测试函数
func (h *Helper) RunParallel(fn func(h *Helper)) {
	h.T.Parallel()
	fn(h)
}

// SetupAndTeardown 设置测试的前置和后置处理
func (h *Helper) SetupAndTeardown(setup func(), teardown func()) {
	if setup != nil {
		setup()
	}

	if teardown != nil {
		h.T.Cleanup(teardown)
	}
}
