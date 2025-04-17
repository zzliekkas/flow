// Package test 提供测试支持工具和辅助函数
package test

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/zzliekkas/flow"
	"github.com/zzliekkas/flow/db"
)

// IntegrationTest 提供集成测试支持
type IntegrationTest struct {
	T         *testing.T
	App       *flow.Engine
	TestDB    *gorm.DB
	Helper    *Helper
	SQLiteDSN string
}

// NewIntegrationTest 创建一个新的集成测试助手
func NewIntegrationTest(t *testing.T) *IntegrationTest {
	helper := NewHelper(t)

	// 创建临时数据库
	dbPath := helper.TempFile("flow-integration-*.db", nil)
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", dbPath)

	return &IntegrationTest{
		T:         t,
		Helper:    helper,
		SQLiteDSN: dsn,
	}
}

// SetupTestApp 设置测试应用
func (i *IntegrationTest) SetupTestApp() *flow.Engine {
	app := flow.New()
	i.App = app
	return app
}

// SetupTestDB 设置测试数据库
func (i *IntegrationTest) SetupTestDB() *gorm.DB {
	// 创建数据库连接
	db, err := gorm.Open(sqlite.Open(i.SQLiteDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(i.T, err, "无法连接测试数据库")

	i.TestDB = db

	// 注册测试结束后的清理函数
	i.T.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	return db
}

// RunWithTimeout 在指定超时时间内运行测试函数
func (i *IntegrationTest) RunWithTimeout(timeout time.Duration, fn func(ctx context.Context)) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	done := make(chan struct{})

	go func() {
		fn(ctx)
		close(done)
	}()

	select {
	case <-done:
		// 测试在超时前完成
	case <-ctx.Done():
		require.Fail(i.T, fmt.Sprintf("测试超时（%v）", timeout))
	}
}

// TruncateTable 清空数据表中的数据
func (i *IntegrationTest) TruncateTable(tableName string) {
	err := i.TestDB.Exec(fmt.Sprintf("DELETE FROM %s", tableName)).Error
	require.NoError(i.T, err, "无法清空表 %s", tableName)

	// 重置自增ID
	err = i.TestDB.Exec(fmt.Sprintf("DELETE FROM sqlite_sequence WHERE name='%s'", tableName)).Error
	if err != nil {
		// 忽略错误，这只是尝试重置自增ID
		i.T.Logf("无法重置表 %s 的自增ID: %v", tableName, err)
	}
}

// CreateTestFixtures 从文件创建测试数据
func (i *IntegrationTest) CreateTestFixtures(fixturePath string, modelType interface{}) {
	data := i.Helper.LoadFixtureFile(fixturePath)
	err := i.TestDB.Model(modelType).Create(data).Error
	require.NoError(i.T, err, "无法创建测试数据")
}

// CreateFixturesDir 创建测试数据文件目录
func (i *IntegrationTest) CreateFixturesDir() string {
	// 获取项目根目录
	root := i.Helper.GetProjectRoot()

	// 创建fixtures目录
	fixturesDir := filepath.Join(root, "test", "fixtures")
	return fixturesDir
}

// ExecuteSQL 执行原始SQL查询
func (i *IntegrationTest) ExecuteSQL(query string, args ...interface{}) {
	err := i.TestDB.Exec(query, args...).Error
	require.NoError(i.T, err, "执行SQL失败: %s", query)
}

// SetupTestTransaction 开始一个测试事务
func (i *IntegrationTest) SetupTestTransaction() *sql.Tx {
	sqlDB, err := i.TestDB.DB()
	require.NoError(i.T, err, "无法获取数据库连接")

	tx, err := sqlDB.Begin()
	require.NoError(i.T, err, "无法开始事务")

	// 注册测试结束后的回滚
	i.T.Cleanup(func() {
		_ = tx.Rollback()
	})

	return tx
}

// SetupRepositoryWithTx 使用事务设置仓储
// 这是一个通用示例，实际使用时需要根据项目的仓储实现进行调整
func (i *IntegrationTest) SetupRepositoryWithTx(modelType interface{}) *db.GenericRepository[interface{}] {
	txDB := i.TestDB.Begin()
	require.NoError(i.T, txDB.Error, "无法开始事务")

	// 注册测试结束后的回滚
	i.T.Cleanup(func() {
		_ = txDB.Rollback()
	})

	// 使用GenericRepository，需要根据实际项目修改
	repo := db.NewGenericRepository[interface{}](txDB)
	return repo
}

// MigrateTestModels 为测试模型创建数据库表
func (i *IntegrationTest) MigrateTestModels(models ...interface{}) {
	err := i.TestDB.AutoMigrate(models...)
	require.NoError(i.T, err, "自动迁移模型失败")
}

// NewHTTPTest 创建用于HTTP测试的客户端
func (i *IntegrationTest) NewHTTPTest() *HTTPClient {
	require.NotNil(i.T, i.App, "应用尚未设置，请先调用SetupTestApp")
	return NewHTTPClient(i.T, i.App)
}

// IntegrationTestSuite 集成测试套件接口
type IntegrationTestSuite interface {
	SetupSuite(t *testing.T)
	TeardownSuite()
	SetupTest(t *testing.T)
	TeardownTest()
}

// RunIntegrationTests 运行集成测试套件中的测试
func RunIntegrationTests(t *testing.T, suite IntegrationTestSuite, tests map[string]func(t *testing.T)) {
	// 设置测试套件
	suite.SetupSuite(t)
	defer suite.TeardownSuite()

	// 执行每个测试
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// 为每个测试设置环境
			suite.SetupTest(t)
			defer suite.TeardownTest()

			// 执行测试
			test(t)
		})
	}
}
