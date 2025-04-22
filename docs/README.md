# Flow框架文档

Flow是一个基于Gin构建的现代化Go Web应用框架，旨在提供更流畅、简洁而强大的开发体验。通过平衡易用性和灵活性，Flow框架让开发者可以快速构建高性能、可维护的Web应用。

## 框架理念

Flow框架的设计理念是"流畅的开发体验"。我们认为框架应该：

- **简单但不简陋**：提供清晰的API，减少模板代码，但不牺牲功能
- **灵活但有约束**：允许自定义，但提供合理的默认值和最佳实践
- **强大但轻量**：提供丰富功能，但保持核心轻量化
- **高性能但可维护**：优化性能，同时保持代码的可读性和可维护性

## 快速开始

### 安装

```bash
go get github.com/zzliekkas/flow
```

### 创建第一个应用

```go
package main

import (
    "github.com/zzliekkas/flow"
)

func main() {
    // 创建Flow应用
    app := flow.New()
    
    // 注册路由
    app.GET("/", func(c *flow.Context) {
        c.String(200, "Hello, Flow!")
    })
    
    // 启动服务器
    app.Run(":8080")
}
```

### 使用启动选项

```go
app := flow.New(
    flow.WithConfig("config.yaml"),
    flow.WithDatabase(),
    flow.WithLogger("info"),
    flow.WithMiddleware(middleware.Recovery()),
)
```

## 核心功能

### 路由系统

Flow提供直观的路由管理：

```go
// 注册路由
app.GET("/users", GetUsers)
app.POST("/users", CreateUser)
app.GET("/users/:id", GetUser)

// 路由组
api := app.Group("/api")
{
    api.GET("/products", GetProducts)
    api.POST("/products", CreateProduct)
}

// 中间件
auth := app.Group("/auth", middleware.JWT())
{
    auth.GET("/profile", GetProfile)
}
```

### 配置管理

Flow提供灵活的配置系统：

```go
// 加载配置
config.Load("config.yaml")

// 获取配置值
appName := config.GetString("app.name")
port := config.GetInt("server.port")
debug := config.GetBool("app.debug")
```

### 依赖注入

Flow集成了依赖注入容器：

```go
// 注册服务
app.Provide(NewUserService)
app.Provide(NewProductService)

// 使用服务
app.GET("/users", func(c *flow.Context) {
    var userService *UserService
    c.Inject(&userService)
    users := userService.GetAll()
    c.JSON(200, users)
})
```

### 数据库操作

Flow简化了数据库操作：

```go
// 模型定义
type User struct {
    ID        uint   `gorm:"primaryKey"`
    Name      string `gorm:"not null"`
    Email     string `gorm:"unique;not null"`
    CreatedAt time.Time
}

// 查询
var users []User
db.Find(&users)

// 创建
db.Create(&User{Name: "John", Email: "john@example.com"})
```

### 中间件

使用内置中间件：

```go
// 全局中间件
app.Use(middleware.Logger())
app.Use(middleware.Recovery())

// 路由组中间件
api := app.Group("/api", middleware.CORS())
```

## 框架模块

Flow框架由多个模块组成，每个模块都可以独立使用：

### 认证系统 (auth/)

认证模块支持多种身份验证方式和权限管理：

```go
// 配置JWT认证
auth := auth.NewJWTAuth(auth.Config{
    SecretKey: "your-secret-key",
    Expire:    24 * time.Hour,
})

// 保护路由
app.GET("/protected", auth.Middleware(), func(c *flow.Context) {
    c.String(200, "Protected resource")
})

// 获取当前用户
user := auth.CurrentUser(c)
```

### 数据库支持 (db/)

数据库模块提供GORM增强功能和事务管理：

```go
// 初始化数据库
db.Init(db.Config{
    Driver:   "mysql",
    Host:     "localhost",
    Port:     3306,
    Username: "root",
    Password: "password",
    Database: "flow_app",
})

// 使用事务
db.Transaction(func(tx *gorm.DB) error {
    // 执行事务操作
    return nil
})
```

### 缓存系统 (cache/)

缓存模块支持多种驱动和便捷API：

```go
// 使用Redis缓存
cache := cache.New(cache.Redis{
    Addr: "localhost:6379",
})

// 存储数据
cache.Set("key", "value", 10*time.Minute)

// 获取数据
val, err := cache.Get("key")
```

### 消息队列 (queue/)

队列模块用于异步任务处理：

```go
// 创建队列
queue := queue.New(queue.Config{
    Driver: "redis",
    Addr:   "localhost:6379",
})

// 注册任务处理器
queue.Register("send_email", SendEmailHandler)

// 分发任务
queue.Dispatch("send_email", map[string]interface{}{
    "to":      "user@example.com",
    "subject": "Welcome",
})
```

### 国际化 (i18n/)

国际化模块支持多语言翻译：

```go
// 初始化
i18n.Init("locales")

// 设置当前语言
i18n.SetLocale("zh-CN")

// 翻译
message := i18n.T("welcome", map[string]interface{}{
    "name": "John",
})
```

### WebSocket支持 (websocket/)

WebSocket模块提供实时通信支持：

```go
// 创建WebSocket管理器
ws := websocket.New()

// 注册处理器
app.GET("/ws", ws.Handler())

// 广播消息
ws.Broadcast("event", data)
```

### 开发工具 (dev/)

开发工具模块提供热重载和调试功能：

```go
// 启用开发工具
if app.IsDebug() {
    dev.EnableHotReload("./")
    dev.EnableDebugger()
}
```

### 性能分析 (profiler/)

性能分析模块提供应用性能监控：

```go
// 创建性能分析器
profiler := profiler.New()

// 启动分析
profiler.Start()

// 导出报告
report, err := profiler.Stop()
```

## 高级主题

### 部署指南

详细介绍了如何将Flow应用部署到不同环境：

- Docker容器化
- Kubernetes部署
- 传统服务器部署

### 性能优化

提供Flow应用的性能优化建议：

- 数据库查询优化
- 缓存策略
- 并发处理

### 测试策略

介绍Flow应用的测试最佳实践：

- 单元测试
- 集成测试
- 端到端测试

## 贡献指南

如何为Flow框架做出贡献：

- 代码风格规范
- Pull Request流程
- 问题报告指南

## API参考

完整的API文档，包括：

- 核心API
- 各模块API
- 辅助函数

## 示例应用

提供完整的示例应用：

- 简单API服务器
- 完整Web应用
- 微服务架构示例 