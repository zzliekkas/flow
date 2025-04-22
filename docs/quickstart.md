# Flow框架快速入门指南

本指南将帮助您快速上手Flow框架，从安装到创建一个简单但功能完整的Web应用。

## 安装Flow框架

### 前提条件

- Go 1.16+
- Git

### 安装步骤

1. 使用go get安装Flow框架：

```bash
go get -u github.com/zzliekkas/flow
```

2. 验证安装：

```bash
go run -v github.com/zzliekkas/flow/cmd/flow version
```

如果显示版本信息，说明安装成功。

## 创建第一个项目

### 使用CLI创建项目

Flow框架提供了便捷的CLI工具来创建项目：

```bash
# 创建一个新项目
go run github.com/zzliekkas/flow/cmd/flow make:project myapp

# 进入项目目录
cd myapp

# 初始化模块
go mod tidy
```

### 手动创建项目

如果您希望手动创建项目，可以按照以下步骤：

1. 创建项目目录和go.mod：

```bash
mkdir myapp
cd myapp
go mod init myapp
```

2. 创建main.go文件：

```go
package main

import (
    "github.com/zzliekkas/flow"
)

func main() {
    // 创建应用
    app := flow.New()
    
    // 注册路由
    app.GET("/", func(c *flow.Context) {
        c.String(200, "Hello, Flow!")
    })
    
    // 启动服务器
    app.Run(":8080")
}
```

3. 运行应用：

```bash
go run main.go
```

访问 http://localhost:8080 查看结果。

## 基本概念

### 路由

Flow提供了便捷的路由注册方法：

```go
// 基本路由
app.GET("/hello", HelloHandler)
app.POST("/users", CreateUserHandler)
app.PUT("/users/:id", UpdateUserHandler)
app.DELETE("/users/:id", DeleteUserHandler)

// 路由组
api := app.Group("/api")
{
    api.GET("/users", GetUsersHandler)
    api.GET("/products", GetProductsHandler)
}

// 路由处理函数
func HelloHandler(c *flow.Context) {
    c.String(200, "Hello World!")
}

// 获取URL参数
func GetUserHandler(c *flow.Context) {
    userID := c.Param("id")
    c.JSON(200, map[string]string{"id": userID})
}

// 获取查询参数
func SearchHandler(c *flow.Context) {
    query := c.Query("q")
    c.JSON(200, map[string]string{"query": query})
}
```

### 中间件

中间件可以处理请求前后的逻辑：

```go
// 全局中间件
app.Use(middleware.Logger())
app.Use(middleware.Recovery())

// 路由组中间件
admin := app.Group("/admin", middleware.Auth())

// 自定义中间件
func MyMiddleware(c *flow.Context) {
    // 请求前逻辑
    fmt.Println("Before request")
    
    // 继续处理
    c.Next()
    
    // 请求后逻辑
    fmt.Println("After request")
}
```

### 请求和响应

Flow提供了丰富的请求处理和响应生成方法：

```go
// 处理JSON请求
func CreateUser(c *flow.Context) {
    var user User
    if err := c.BindJSON(&user); err != nil {
        c.JSON(400, map[string]string{"error": err.Error()})
        return
    }
    
    // 处理用户创建...
    
    c.JSON(201, user)
}

// 处理表单请求
func Login(c *flow.Context) {
    username := c.PostForm("username")
    password := c.PostForm("password")
    
    // 验证登录...
    
    c.JSON(200, map[string]string{"token": "..."})
}

// 文件上传
func UploadFile(c *flow.Context) {
    file, err := c.FormFile("file")
    if err != nil {
        c.String(400, "文件上传失败")
        return
    }
    
    // 保存文件
    dst := filepath.Join("uploads", file.Filename)
    c.SaveUploadedFile(file, dst)
    
    c.String(200, "文件上传成功")
}

// 各种响应类型
func Responses(c *flow.Context) {
    // 字符串响应
    c.String(200, "Hello World")
    
    // JSON响应
    c.JSON(200, map[string]interface{}{
        "name": "John",
        "age": 30,
    })
    
    // XML响应
    c.XML(200, MyStruct{...})
    
    // HTML响应
    c.HTML(200, "index.html", map[string]interface{}{
        "title": "My Page",
    })
    
    // 文件下载
    c.File("path/to/file.pdf")
    
    // 重定向
    c.Redirect(302, "/login")
}
```

## 完整示例

下面是一个更完整的示例，包含了路由、中间件、数据库和模板：

```go
package main

import (
    "log"
    
    "github.com/zzliekkas/flow"
    "github.com/zzliekkas/flow/middleware"
)

type User struct {
    ID    uint   `json:"id" gorm:"primaryKey"`
    Name  string `json:"name" gorm:"not null"`
    Email string `json:"email" gorm:"unique;not null"`
}

func main() {
    // 创建应用
    app := flow.New(
        flow.WithConfig("config.yaml"),
        flow.WithDatabase(),
        flow.WithTemplates("templates/**/*.html"),
    )
    
    // 全局中间件
    app.Use(middleware.Logger())
    app.Use(middleware.Recovery())
    
    // 注册路由
    app.GET("/", func(c *flow.Context) {
        c.HTML(200, "index.html", flow.H{
            "title": "Flow Framework",
        })
    })
    
    // API路由组
    api := app.Group("/api")
    {
        api.GET("/users", GetUsers)
        api.POST("/users", CreateUser)
        api.GET("/users/:id", GetUser)
    }
    
    // 启动服务器
    if err := app.Run(":8080"); err != nil {
        log.Fatal(err)
    }
}

// 获取所有用户
func GetUsers(c *flow.Context) {
    var users []User
    db := c.DB()
    
    if err := db.Find(&users).Error; err != nil {
        c.JSON(500, flow.H{"error": "Failed to get users"})
        return
    }
    
    c.JSON(200, users)
}

// 创建用户
func CreateUser(c *flow.Context) {
    var user User
    if err := c.BindJSON(&user); err != nil {
        c.JSON(400, flow.H{"error": err.Error()})
        return
    }
    
    db := c.DB()
    if err := db.Create(&user).Error; err != nil {
        c.JSON(500, flow.H{"error": "Failed to create user"})
        return
    }
    
    c.JSON(201, user)
}

// 获取单个用户
func GetUser(c *flow.Context) {
    id := c.Param("id")
    var user User
    
    db := c.DB()
    if err := db.First(&user, id).Error; err != nil {
        c.JSON(404, flow.H{"error": "User not found"})
        return
    }
    
    c.JSON(200, user)
}
```

## 项目结构

Flow框架推荐的项目结构如下：

```
myapp/
├── config/                  # 配置文件
│   ├── app.yaml
│   └── database.yaml
├── controllers/             # 控制器
│   ├── auth_controller.go
│   └── user_controller.go
├── models/                  # 数据模型
│   └── user.go
├── repositories/            # 数据仓库
│   └── user_repository.go
├── services/                # 业务逻辑
│   └── auth_service.go
├── middleware/              # 自定义中间件
│   └── auth_middleware.go
├── routes/                  # 路由定义
│   └── api.go
├── templates/               # 视图模板
│   └── index.html
├── public/                  # 静态资源
│   ├── css/
│   └── js/
├── tests/                   # 测试文件
│   └── api_test.go
├── main.go                  # 应用入口
├── go.mod                   # Go模块文件
└── go.sum                   # Go依赖校验文件
```

## 下一步

- 阅读[完整文档](./README.md)了解更多功能
- 查看[示例应用](../examples/)学习实际使用方法
- 了解[最佳实践](./best-practices.md)
- 参考[API文档](./api-reference.md)了解详细接口 