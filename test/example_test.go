package test_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zzliekkas/flow"
	"github.com/zzliekkas/flow/test"
)

// User 测试用户模型
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:50;not null" json:"name"`
	Email     string    `gorm:"size:100;uniqueIndex;not null" json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ExampleService 示例服务
type ExampleService struct {
	Mock *test.Mock
}

// NewExampleService 创建示例服务
func NewExampleService() *ExampleService {
	return &ExampleService{
		Mock: test.NewMock(),
	}
}

// GetUserByID 模拟获取用户方法
func (s *ExampleService) GetUserByID(id uint) *User {
	s.Mock.RecordCall("GetUserByID", id)
	return &User{
		ID:    id,
		Name:  "测试用户",
		Email: "test@example.com",
	}
}

// 单元测试示例
func TestUnitExample(t *testing.T) {
	// 创建单元测试助手
	ut := test.NewUnitTest(t)

	t.Run("测试模拟对象", func(t *testing.T) {
		// 创建服务实例
		service := NewExampleService()

		// 调用要测试的方法
		user := service.GetUserByID(1)

		// 创建模拟断言
		mockAssert := test.NewMockAssertions(t, service.Mock)

		// 验证方法是否被调用
		mockAssert.AssertCalled("GetUserByID", uint(1))
		assert.Equal(t, uint(1), user.ID, "用户ID应该匹配")
		assert.Equal(t, "测试用户", user.Name, "用户名称应该匹配")
	})

	t.Run("测试JSON处理", func(t *testing.T) {
		// 测试ToJSON
		user := User{ID: 1, Name: "测试用户", Email: "test@example.com"}
		jsonStr := ut.ToJSON(user)

		// 测试FromJSON
		var newUser User
		ut.FromJSON(jsonStr, &newUser)

		assert.Equal(t, user.ID, newUser.ID, "解析后的ID应匹配")
		assert.Equal(t, user.Name, newUser.Name, "解析后的名称应匹配")
		assert.Equal(t, user.Email, newUser.Email, "解析后的邮箱应匹配")
	})
}

// 模拟数据生成示例
func TestMockGenerator(t *testing.T) {
	// 创建模拟数据生成器
	gen := test.NewMockGenerator()

	t.Run("生成随机基本类型", func(t *testing.T) {
		randomInt := gen.RandomInt(1, 100)
		assert.GreaterOrEqual(t, randomInt, 1, "随机整数应大于等于1")
		assert.LessOrEqual(t, randomInt, 100, "随机整数应小于等于100")

		randomStr := gen.RandomString(10)
		assert.Len(t, randomStr, 10, "随机字符串长度应为10")

		randomEmail := gen.RandomEmail()
		assert.Contains(t, randomEmail, "@", "随机邮箱应包含@符号")
	})

	t.Run("填充结构体", func(t *testing.T) {
		var user User
		gen.FillStruct(&user)

		assert.NotEmpty(t, user.Name, "名称不应为空")
		assert.NotEmpty(t, user.Email, "邮箱不应为空")
		assert.Contains(t, user.Email, "@", "邮箱应包含@符号")
	})
}

// 集成测试示例
func TestIntegrationExample(t *testing.T) {
	// 跳过集成测试，除非显式启用
	// 实际项目中可以通过环境变量来控制是否运行集成测试
	t.Skip("跳过集成测试，这只是一个示例")

	// 创建集成测试助手
	it := test.NewIntegrationTest(t)

	// 设置测试应用
	app := it.SetupTestApp()

	// 设置路由
	app.GET("/test", func(c *flow.Context) {
		c.JSON(http.StatusOK, map[string]string{"message": "Hello, Test!"})
	})

	// 设置测试数据库
	db := it.SetupTestDB()

	// 创建测试表
	it.MigrateTestModels(&User{})

	t.Run("测试HTTP请求", func(t *testing.T) {
		// 创建HTTP测试客户端
		client := it.NewHTTPTest()

		// 发送GET请求
		res := client.GET("/test", nil, nil)

		// 验证响应
		res.AssertStatus(t, http.StatusOK)
		res.AssertJSONContains(t, "message", "Hello, Test!")
	})

	t.Run("测试数据库操作", func(t *testing.T) {
		// 创建测试用户
		user := User{
			Name:  "集成测试用户",
			Email: "integration@example.com",
		}

		// 插入数据库
		err := db.Create(&user).Error
		require.NoError(t, err, "创建用户记录应成功")

		// 查询验证
		var found User
		err = db.First(&found, user.ID).Error
		require.NoError(t, err, "查询用户记录应成功")
		assert.Equal(t, user.Name, found.Name, "查询到的用户名称应匹配")
	})

	t.Run("测试带超时的操作", func(t *testing.T) {
		// 使用超时运行测试函数
		it.RunWithTimeout(2*time.Second, func(ctx context.Context) {
			// 模拟耗时操作
			select {
			case <-time.After(1 * time.Second):
				// 完成操作
			case <-ctx.Done():
				t.Fail()
			}
		})
	})
}
