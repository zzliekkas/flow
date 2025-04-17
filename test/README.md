# Flow 测试工具

Flow 测试工具提供了一套完整的测试支持库，帮助开发者轻松编写单元测试和集成测试。这些工具旨在简化测试过程，提高测试代码的可读性和可维护性。

## 目录结构

```
test/
├── helper.go       # 通用测试助手函数
├── http.go         # HTTP测试客户端
├── mock.go         # 模拟数据生成工具
├── unit.go         # 单元测试支持
├── integration.go  # 集成测试支持
├── fixtures/       # 测试数据文件
│   └── users.json  # 用户测试数据示例
└── example_test.go # 使用示例
```

## 主要功能

### 1. 测试助手 (Helper)

`Helper` 提供通用的测试辅助功能，如临时文件创建、路径解析等。

```go
helper := test.NewHelper(t)

// 获取项目根目录
rootPath := helper.GetProjectRoot()

// 创建临时文件
tmpFile := helper.TempFile("prefix-*.txt", []byte("测试内容"))

// 创建临时目录
tmpDir := helper.TempDir()

// 生成随机字符串
randomStr := helper.RandomString(10)
```

### 2. HTTP测试客户端 (HTTPClient)

`HTTPClient` 简化了API测试过程，提供了一种流畅的方式来模拟HTTP请求和验证响应。

```go
// 创建HTTP测试客户端
client := test.NewHTTPClient(t, app)

// 发送GET请求
response := client.GET("/api/users", nil, nil)

// 验证响应
response.AssertStatus(t, 200)
response.AssertJSONContains(t, "message", "操作成功")

// 发送POST请求
response = client.POST("/api/users", user, nil)

// 绑定响应到结构体
var result UserResponse
response.BindJSON(&result)
```

### 3. 模拟数据生成 (MockGenerator)

`MockGenerator` 用于生成各种类型的随机测试数据，特别适合填充实体对象。

```go
// 创建模拟数据生成器
gen := test.NewMockGenerator()

// 生成随机基本类型
randomInt := gen.RandomInt(1, 100)
randomFloat := gen.RandomFloat(1.0, 100.0)
randomBool := gen.RandomBool()
randomString := gen.RandomString(10)
randomEmail := gen.RandomEmail()
randomName := gen.RandomName()
randomDate := gen.RandomDate(2000, 2023)
randomPhone := gen.RandomPhoneNumber()

// 自动填充结构体
var user User
gen.FillStruct(&user)
```

### 4. 单元测试支持 (UnitTest)

`UnitTest` 提供单元测试相关功能，简化测试编写和断言过程。

```go
// 创建单元测试助手
ut := test.NewUnitTest(t)

// 获取测试函数名称
testName := ut.GetTestName()

// JSON 转换
jsonStr := ut.ToJSON(obj)
ut.FromJSON(jsonStr, &result)
ut.AssertJSONEqual(expected, actual)

// 运行测试用例
ut.RunTestCase("测试用例1", func(t *testing.T) {
    // 测试代码
})
```

### 5. 模拟对象 (Mock)

`Mock` 用于创建模拟对象，记录方法调用并进行验证。

```go
// 创建模拟对象
mock := test.NewMock()

// 记录方法调用
mock.RecordCall("MethodName", arg1, arg2)

// 验证方法调用
mockAssert := test.NewMockAssertions(t, mock)
mockAssert.AssertCalled("MethodName", arg1, arg2)
mockAssert.AssertNotCalled("OtherMethod")
mockAssert.AssertCalledTimes("MethodName", 2)
```

### 6. 集成测试支持 (IntegrationTest)

`IntegrationTest` 提供一套工具，用于进行依赖于数据库和外部服务的集成测试。

```go
// 创建集成测试助手
it := test.NewIntegrationTest(t)

// 设置测试应用
app := it.SetupTestApp()

// 设置测试数据库
db := it.SetupTestDB()

// 创建测试表
it.MigrateTestModels(&User{}, &Product{})

// 清空表数据
it.TruncateTable("users")

// 创建测试数据
it.CreateTestFixtures("users.json", &User{})

// 执行SQL
it.ExecuteSQL("UPDATE users SET status = ? WHERE id = ?", "active", 1)

// HTTP 测试
client := it.NewHTTPTest()
response := client.GET("/api/users", nil, nil)

// 事务测试
tx := it.SetupTestTransaction()
```

## 使用示例

详细示例可参考 `example_test.go` 文件，其中包含了单元测试和集成测试的完整示例。

## 最佳实践

1. **隔离测试环境**: 集成测试应使用独立的测试数据库，避免影响开发环境
2. **合理使用模拟**: 单元测试中通过模拟外部依赖，专注于测试单个组件的功能
3. **使用测试数据文件**: 将测试数据存储在 `fixtures` 目录，便于维护和复用
4. **测试边界条件**: 除了正常场景，也要测试各种错误情况和边界条件
5. **保持测试独立**: 每个测试应该能够独立运行，不依赖其他测试的结果

## 注意事项

- 集成测试通常较慢，可以通过环境变量或命令行标志来控制是否运行
- 尽量使用测试工具提供的辅助函数，减少重复代码
- 测试失败时提供有意义的错误消息，便于定位问题 