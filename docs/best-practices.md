# Flow框架最佳实践

本文档提供了使用Flow框架开发应用的最佳实践和建议，帮助您写出更加清晰、高效和易维护的代码。

## 项目组织

### 目录结构

推荐使用以下目录结构组织项目：

```
myapp/
├── config/                # 配置文件
├── app/                   # 应用核心
│   ├── controllers/       # 控制器
│   ├── models/            # 数据模型
│   ├── repositories/      # 数据仓库
│   ├── services/          # 业务服务
│   └── middleware/        # 中间件
├── routes/                # 路由定义
├── database/              # 数据库相关
│   ├── migrations/        # 数据库迁移
│   └── seeds/             # 数据种子
├── resources/             # 资源文件
│   ├── templates/         # 视图模板
│   ├── lang/              # 语言文件
│   └── assets/            # 静态资源
├── storage/               # 存储目录
│   ├── logs/              # 日志文件
│   └── uploads/           # 上传文件
├── tests/                 # 测试文件
├── main.go                # 应用入口
├── go.mod                 # Go模块文件
└── go.sum                 # 依赖校验文件
```

### 包命名

- 使用单数形式：`controller`, `model`, `service`，而不是`controllers`
- 避免使用通用名称如`util`，尽量用更具体的名称如`fileutil`, `stringutil`
- 包名应该简短、清晰且与其功能相关

## 代码组织

### 控制器

控制器应当保持简洁，主要负责：
- 解析请求参数
- 调用适当的服务
- 返回响应

**推荐写法：**

```go
// user_controller.go
package controller

import (
    "github.com/zzliekkas/flow"
    "myapp/app/service"
)

type UserController struct {
    userService *service.UserService
}

func NewUserController(userService *service.UserService) *UserController {
    return &UserController{userService: userService}
}

func (c *UserController) GetUsers(ctx *flow.Context) {
    // 获取查询参数
    limit := ctx.QueryInt("limit", 10)
    page := ctx.QueryInt("page", 1)
    
    // 调用服务
    users, err := c.userService.GetUsers(page, limit)
    if err != nil {
        ctx.JSON(500, flow.H{"error": err.Error()})
        return
    }
    
    // 返回响应
    ctx.JSON(200, users)
}

// 注册路由
func (c *UserController) RegisterRoutes(router *flow.RouterGroup) {
    router.GET("/users", c.GetUsers)
    router.GET("/users/:id", c.GetUser)
    router.POST("/users", c.CreateUser)
    router.PUT("/users/:id", c.UpdateUser)
    router.DELETE("/users/:id", c.DeleteUser)
}
```

### 服务

服务层包含业务逻辑：
- 从仓库获取数据
- 处理业务规则
- 封装事务逻辑

**推荐写法：**

```go
// user_service.go
package service

import (
    "myapp/app/model"
    "myapp/app/repository"
)

type UserService struct {
    userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
    return &UserService{userRepo: userRepo}
}

func (s *UserService) GetUsers(page, limit int) ([]model.User, error) {
    return s.userRepo.FindAll(page, limit)
}

func (s *UserService) CreateUser(data map[string]interface{}) (*model.User, error) {
    // 验证业务规则
    if err := s.validateUserData(data); err != nil {
        return nil, err
    }
    
    // 创建用户
    user := &model.User{
        Name:  data["name"].(string),
        Email: data["email"].(string),
    }
    
    return s.userRepo.Create(user)
}

func (s *UserService) validateUserData(data map[string]interface{}) error {
    // 实现验证逻辑
    return nil
}
```

### 仓库

仓库负责处理数据存储和检索：
- 封装数据库操作
- 实现数据查询
- 隐藏数据访问细节

**推荐写法：**

```go
// user_repository.go
package repository

import (
    "gorm.io/gorm"
    "myapp/app/model"
)

type UserRepository struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
    return &UserRepository{db: db}
}

func (r *UserRepository) FindAll(page, limit int) ([]model.User, error) {
    var users []model.User
    offset := (page - 1) * limit
    
    err := r.db.Offset(offset).Limit(limit).Find(&users).Error
    return users, err
}

func (r *UserRepository) FindByID(id uint) (*model.User, error) {
    var user model.User
    err := r.db.First(&user, id).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *UserRepository) Create(user *model.User) (*model.User, error) {
    err := r.db.Create(user).Error
    return user, err
}

func (r *UserRepository) Update(user *model.User) error {
    return r.db.Save(user).Error
}

func (r *UserRepository) Delete(id uint) error {
    return r.db.Delete(&model.User{}, id).Error
}
```

## 依赖注入

Flow框架提供了依赖注入容器，推荐使用以下方式组织依赖：

```go
// 在main.go或初始化文件中注册依赖
func registerServices(app *flow.Engine) {
    // 注册仓库
    app.Provide(repository.NewUserRepository)
    
    // 注册服务
    app.Provide(service.NewUserService)
    
    // 注册控制器
    app.Provide(controller.NewUserController)
}

// 使用依赖
func setupRoutes(app *flow.Engine) {
    app.Invoke(func(userController *controller.UserController) {
        api := app.Group("/api")
        userController.RegisterRoutes(api)
    })
}
```

## 错误处理

### 统一错误处理

创建一个中央错误处理器：

```go
// error/handler.go
package error

import (
    "github.com/zzliekkas/flow"
)

// AppError 自定义错误类型
type AppError struct {
    Code    int
    Message string
    Err     error
}

func (e *AppError) Error() string {
    return e.Message
}

// NewAppError 创建应用错误
func NewAppError(code int, message string, err error) *AppError {
    return &AppError{
        Code:    code,
        Message: message,
        Err:     err,
    }
}

// ErrorHandler 错误处理中间件
func ErrorHandler(c *flow.Context) {
    c.Next()
    
    // 处理已经设置的错误
    if len(c.Errors) > 0 {
        err := c.Errors[0].Err
        
        // 根据错误类型返回不同响应
        switch e := err.(type) {
        case *AppError:
            c.JSON(e.Code, flow.H{"error": e.Message})
        default:
            c.JSON(500, flow.H{"error": "Internal Server Error"})
        }
    }
}
```

在控制器中使用：

```go
func (c *UserController) GetUser(ctx *flow.Context) {
    id := ctx.ParamUint("id")
    
    user, err := c.userService.GetUser(id)
    if err != nil {
        // 添加错误，会被错误处理中间件捕获
        ctx.Error(err)
        return
    }
    
    ctx.JSON(200, user)
}
```

## 验证

### 请求验证

使用结构体标签进行请求验证：

```go
// 用户创建请求
type CreateUserRequest struct {
    Name     string `json:"name" binding:"required,min=2,max=50"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=8"`
    Age      int    `json:"age" binding:"required,gte=18"`
}

func (c *UserController) CreateUser(ctx *flow.Context) {
    var req CreateUserRequest
    
    // 绑定并验证请求
    if err := ctx.BindJSON(&req); err != nil {
        ctx.JSON(400, flow.H{"error": err.Error()})
        return
    }
    
    // 请求验证通过，继续处理
    user, err := c.userService.CreateUser(map[string]interface{}{
        "name":     req.Name,
        "email":    req.Email,
        "password": req.Password,
        "age":      req.Age,
    })
    
    if err != nil {
        ctx.JSON(500, flow.H{"error": err.Error()})
        return
    }
    
    ctx.JSON(201, user)
}
```

## 配置管理

### 使用层次结构配置

在config.yaml中使用层次结构：

```yaml
app:
  name: "MyApp"
  env: "development"
  debug: true
  
server:
  host: "localhost"
  port: 8080
  
database:
  driver: "mysql"
  host: "localhost"
  port: 3306
  name: "myapp"
  user: "root"
  password: "secret"
  
mail:
  driver: "smtp"
  host: "smtp.example.com"
  port: 587
  username: "no-reply@example.com"
  password: "secret"
```

在代码中访问：

```go
// 加载配置
config.Load("config.yaml")

// 访问配置
appName := config.GetString("app.name")
dbConfig := config.GetStringMap("database")
```

## 中间件使用

### 分组中间件

根据功能分组使用中间件：

```go
// 全局中间件
app.Use(middleware.Recovery())
app.Use(middleware.Logger())

// API中间件组
api := app.Group("/api")
api.Use(middleware.CORS())
api.Use(middleware.RateLimit(100))

// 认证中间件组
auth := api.Group("/")
auth.Use(middleware.JWT())

// 管理员中间件组
admin := auth.Group("/admin")
admin.Use(middleware.AdminOnly())
```

## 数据库最佳实践

### 使用事务

封装事务操作：

```go
func (s *UserService) TransferFunds(fromID, toID uint, amount float64) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        // 从账户扣款
        if err := tx.Model(&model.Account{}).
            Where("id = ? AND balance >= ?", fromID, amount).
            UpdateColumn("balance", gorm.Expr("balance - ?", amount)).Error; err != nil {
            return err
        }
        
        // 检查是否成功扣款
        var fromAccount model.Account
        if err := tx.First(&fromAccount, fromID).Error; err != nil {
            return err
        }
        
        // 给目标账户增加余额
        if err := tx.Model(&model.Account{}).
            Where("id = ?", toID).
            UpdateColumn("balance", gorm.Expr("balance + ?", amount)).Error; err != nil {
            return err
        }
        
        // 记录交易
        txn := model.Transaction{
            FromAccountID: fromID,
            ToAccountID:   toID,
            Amount:        amount,
        }
        if err := tx.Create(&txn).Error; err != nil {
            return err
        }
        
        return nil
    })
}
```

### 使用索引

为常查询字段添加索引：

```go
type User struct {
    ID        uint      `gorm:"primaryKey"`
    Email     string    `gorm:"uniqueIndex"`
    Name      string    `gorm:"index"`
    CreatedAt time.Time `gorm:"index"`
}
```

## 日志最佳实践

### 结构化日志

使用结构化日志便于分析：

```go
// 初始化结构化日志器
logger := logger.New(
    logger.WithLevel(logger.InfoLevel),
    logger.WithFormat(logger.JSONFormat),
    logger.WithOutput("logs/app.log"),
)

// 记录信息
logger.Info("用户登录", logger.Fields{
    "user_id": user.ID,
    "email":   user.Email,
    "ip":      ctx.ClientIP(),
})

// 记录错误
logger.Error("数据库查询失败", logger.Fields{
    "error":   err.Error(),
    "query":   query,
    "params":  params,
})
```

## 测试最佳实践

### 单元测试

为每个组件编写单元测试：

```go
// user_service_test.go
package service_test

import (
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    
    "myapp/app/model"
    "myapp/app/service"
    "myapp/mocks"
)

func TestGetUser(t *testing.T) {
    // 创建模拟仓库
    mockRepo := new(mocks.UserRepository)
    mockRepo.On("FindByID", uint(1)).Return(&model.User{
        ID:    1,
        Name:  "John Doe",
        Email: "john@example.com",
    }, nil)
    
    // 创建服务并注入模拟仓库
    userService := service.NewUserService(mockRepo)
    
    // 测试方法
    user, err := userService.GetUser(1)
    
    // 断言结果
    assert.NoError(t, err)
    assert.NotNil(t, user)
    assert.Equal(t, "John Doe", user.Name)
    
    // 验证模拟调用
    mockRepo.AssertExpectations(t)
}
```

### API测试

测试API端点：

```go
// api_test.go
package test

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/stretchr/testify/assert"
    
    "myapp/app"
)

func TestGetUsers(t *testing.T) {
    // 创建测试应用
    app := app.SetupTestApp()
    
    // 创建测试请求
    req := httptest.NewRequest("GET", "/api/users", nil)
    
    // 记录响应
    w := httptest.NewRecorder()
    app.ServeHTTP(w, req)
    
    // 检查状态码
    assert.Equal(t, http.StatusOK, w.Code)
    
    // 解析响应体
    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    
    // 检查响应内容
    assert.NoError(t, err)
    assert.Contains(t, response, "data")
}
```

## 性能优化

### 缓存使用

使用缓存减少数据库查询：

```go
func (s *UserService) GetUser(id uint) (*model.User, error) {
    // 缓存键
    cacheKey := fmt.Sprintf("user:%d", id)
    
    // 尝试从缓存获取
    var user model.User
    found, err := s.cache.Get(cacheKey, &user)
    if err != nil {
        return nil, err
    }
    
    // 缓存命中
    if found {
        return &user, nil
    }
    
    // 缓存未命中，从数据库获取
    user, err = s.userRepo.FindByID(id)
    if err != nil {
        return nil, err
    }
    
    // 保存到缓存
    s.cache.Set(cacheKey, user, 15*time.Minute)
    
    return &user, nil
}
```

### 分页查询

始终使用分页查询大数据集：

```go
func (r *ProductRepository) FindAll(page, limit int) ([]model.Product, error) {
    var products []model.Product
    offset := (page - 1) * limit
    
    err := r.db.Offset(offset).Limit(limit).Find(&products).Error
    return products, err
}
```

## 安全最佳实践

### 数据验证和清理

始终验证和清理输入数据：

```go
// 验证并清理输入
func sanitizeInput(input string) string {
    // 移除HTML标签
    input = html.EscapeString(input)
    
    // 其他清理操作...
    
    return input
}

func (c *UserController) CreateUser(ctx *flow.Context) {
    var req CreateUserRequest
    
    // 绑定请求
    if err := ctx.BindJSON(&req); err != nil {
        ctx.JSON(400, flow.H{"error": err.Error()})
        return
    }
    
    // 清理输入
    req.Name = sanitizeInput(req.Name)
    req.Email = strings.ToLower(strings.TrimSpace(req.Email))
    
    // 继续处理...
}
```

### 敏感数据处理

谨慎处理敏感数据：

```go
// 在模型中隐藏敏感字段
type User struct {
    ID        uint      `json:"id" gorm:"primaryKey"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    Password  string    `json:"-" gorm:"not null"` // 在JSON中隐藏
    CreatedAt time.Time `json:"created_at"`
}

// 加密敏感数据
func hashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
    return string(bytes), err
}

func verifyPassword(hashedPassword, password string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
    return err == nil
}
```

## 部署最佳实践

### 使用环境变量

优先使用环境变量进行配置：

```go
// 从环境变量加载配置
func loadConfig() *AppConfig {
    return &AppConfig{
        AppName:    getEnv("APP_NAME", "MyApp"),
        AppEnv:     getEnv("APP_ENV", "development"),
        ServerPort: getEnvAsInt("SERVER_PORT", 8080),
        DBHost:     getEnv("DB_HOST", "localhost"),
        DBPort:     getEnvAsInt("DB_PORT", 3306),
        DBName:     getEnv("DB_NAME", "myapp"),
        DBUser:     getEnv("DB_USER", "root"),
        DBPassword: getEnv("DB_PASSWORD", ""),
    }
}

// 获取环境变量，如不存在则返回默认值
func getEnv(key, defaultValue string) string {
    if value, exists := os.LookupEnv(key); exists {
        return value
    }
    return defaultValue
}

// 获取环境变量并转换为整数
func getEnvAsInt(key string, defaultValue int) int {
    valueStr := getEnv(key, "")
    if value, err := strconv.Atoi(valueStr); err == nil {
        return value
    }
    return defaultValue
}
```

### 健康检查

实现健康检查端点：

```go
// health.go
package controller

import (
    "github.com/zzliekkas/flow"
)

func HealthCheck(c *flow.Context) {
    // 检查数据库连接
    if err := c.DB().Ping(); err != nil {
        c.JSON(500, flow.H{
            "status":  "error",
            "message": "Database connection failed",
        })
        return
    }
    
    // 检查缓存服务
    if err := c.Cache().Ping(); err != nil {
        c.JSON(500, flow.H{
            "status":  "error",
            "message": "Cache service unavailable",
        })
        return
    }
    
    // 所有服务正常
    c.JSON(200, flow.H{
        "status":  "ok",
        "version": "1.0.0",
    })
}

// 注册健康检查路由
func RegisterHealthCheck(app *flow.Engine) {
    app.GET("/health", HealthCheck)
}
```

### 优雅关闭

实现优雅关闭：

```go
func main() {
    // 创建应用
    app := flow.New()
    
    // 设置路由和中间件...
    
    // 创建服务器
    server := &http.Server{
        Addr:    ":8080",
        Handler: app,
    }
    
    // 启动服务器
    go func() {
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("listen: %s\n", err)
        }
    }()
    
    // 等待中断信号
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    log.Println("Shutting down server...")
    
    // 设置超时上下文
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    // 优雅关闭
    if err := server.Shutdown(ctx); err != nil {
        log.Fatal("Server forced to shutdown:", err)
    }
    
    log.Println("Server exiting")
}
```

## 总结

遵循这些最佳实践，你将能够创建可维护、高效和安全的Flow应用。记住，最佳实践是指导性的，应该根据项目的具体需求进行调整。随着项目的发展，不断优化和完善这些实践，以适应新的挑战和需求。 