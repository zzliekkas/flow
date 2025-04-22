package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/zzliekkas/flow"
	"github.com/zzliekkas/flow/middleware"
)

// 用户模型
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"not null"`
	Email     string    `json:"email" gorm:"unique;not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// 用户控制器
type UserController struct{}

// NewUserController 创建用户控制器
func NewUserController() *UserController {
	return &UserController{}
}

// RegisterRoutes 注册路由
func (c *UserController) RegisterRoutes(router flow.RouterGroup) {
	router.GET("", c.GetUsers)
	router.GET("/:id", c.GetUser)
	router.POST("", c.CreateUser)
	router.PUT("/:id", c.UpdateUser)
	router.DELETE("/:id", c.DeleteUser)
}

// GetUsers 获取所有用户
func (c *UserController) GetUsers(ctx *flow.Context) {
	// 模拟数据库中的用户列表
	users := []User{
		{ID: 1, Name: "张三", Email: "zhangsan@example.com", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: 2, Name: "李四", Email: "lisi@example.com", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: 3, Name: "王五", Email: "wangwu@example.com", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	ctx.JSON(200, flow.H{
		"success": true,
		"data":    users,
	})
}

// GetUser 获取单个用户
func (c *UserController) GetUser(ctx *flow.Context) {
	// 获取URL参数
	id := ctx.Param("id")

	// 模拟查询数据库 - 使用id参数
	var userId uint = 1
	// 尝试将id转换为uint (实际项目中应该进行错误处理)
	if parsedId, err := strconv.ParseUint(id, 10, 64); err == nil {
		userId = uint(parsedId)
	}

	user := User{
		ID:        userId, // 使用转换后的id
		Name:      "张三",
		Email:     "zhangsan@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	ctx.JSON(200, flow.H{
		"success": true,
		"message": fmt.Sprintf("成功获取ID为%s的用户", id),
		"data":    user,
	})
}

// CreateUser 创建用户
func (c *UserController) CreateUser(ctx *flow.Context) {
	// 绑定请求参数
	var user User
	if err := ctx.BindJSON(&user); err != nil {
		ctx.JSON(400, flow.H{
			"success": false,
			"message": "无效的请求数据",
			"error":   err.Error(),
		})
		return
	}

	// 设置ID和时间戳
	user.ID = 4
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	// 返回新创建的用户
	ctx.JSON(201, flow.H{
		"success": true,
		"message": "用户创建成功",
		"data":    user,
	})
}

// UpdateUser 更新用户信息
func (c *UserController) UpdateUser(ctx *flow.Context) {
	// 获取URL参数
	id := ctx.Param("id")

	// 从请求体获取要更新的数据
	var user User
	if err := ctx.BindJSON(&user); err != nil {
		ctx.JSON(400, flow.H{
			"success": false,
			"message": "请求数据格式错误",
			"error":   err.Error(),
		})
		return
	}

	// 模拟更新操作
	// 实际项目中应该进行数据库操作
	user.UpdatedAt = time.Now()

	ctx.JSON(200, flow.H{
		"success": true,
		"message": fmt.Sprintf("成功更新ID为%s的用户", id),
		"data":    user,
	})
}

// DeleteUser 删除用户
func (c *UserController) DeleteUser(ctx *flow.Context) {
	// 获取URL参数
	id := ctx.Param("id")

	// 返回删除结果
	ctx.JSON(200, flow.H{
		"success": true,
		"message": "用户删除成功",
		"data": flow.H{
			"id": id,
		},
	})
}

// 文章控制器
type ArticleController struct{}

// NewArticleController 创建文章控制器
func NewArticleController() *ArticleController {
	return &ArticleController{}
}

// RegisterRoutes 注册路由
func (c *ArticleController) RegisterRoutes(router flow.RouterGroup) {
	router.GET("", c.GetArticles)
	router.GET("/:id", c.GetArticle)
}

// GetArticles 获取所有文章
func (c *ArticleController) GetArticles(ctx *flow.Context) {
	// 模拟文章列表
	articles := []flow.H{
		{"id": 1, "title": "Flow框架入门指南", "content": "这是一篇介绍Flow框架的文章..."},
		{"id": 2, "title": "使用Flow构建RESTful API", "content": "本文将介绍如何使用Flow框架构建RESTful API..."},
		{"id": 3, "title": "Flow框架性能优化技巧", "content": "这里有一些Flow框架性能优化的小技巧..."},
	}

	ctx.JSON(200, flow.H{
		"success": true,
		"data":    articles,
	})
}

// GetArticle 获取单个文章
func (c *ArticleController) GetArticle(ctx *flow.Context) {
	// 获取URL参数
	id := ctx.Param("id")

	// 模拟查询数据库
	article := map[string]interface{}{
		"id":      id, // 使用参数id
		"title":   "Flow框架入门指南",
		"content": "这是一篇关于Flow框架的入门教程，介绍了基本用法和示例代码。",
		"author":  "管理员",
		"created": time.Now().Format(time.RFC3339),
	}

	ctx.JSON(200, flow.H{
		"success": true,
		"message": fmt.Sprintf("获取ID为%s的文章成功", id), // 使用参数id
		"data":    article,
	})
}

// 主函数
func main() {
	// 创建Flow应用
	app := flow.New(
		flow.WithMode("debug"),
		flow.WithLogLevel("debug"),
		flow.WithMiddleware(middleware.Logger()),
		flow.WithMiddleware(middleware.Recovery()),
		flow.WithMiddleware(middleware.CORS()),
	)

	// 注册控制器
	app.Provide(NewUserController)
	app.Provide(NewArticleController)

	// 创建API路由组
	api := app.Group("/api")

	// 注册控制器路由
	app.Invoke(func(userController *UserController, articleController *ArticleController) {
		// 用户路由
		userGroup := api.Group("/users")
		userController.RegisterRoutes(*userGroup)

		// 文章路由
		articleGroup := api.Group("/articles")
		articleController.RegisterRoutes(*articleGroup)
	})

	// 首页路由
	app.GET("/", func(c *flow.Context) {
		c.JSON(200, flow.H{
			"name":        "Flow示例API",
			"version":     "1.0.0",
			"description": "Flow框架的简单示例API",
			"endpoints": map[string]string{
				"用户列表": "/api/users",
				"单个用户": "/api/users/:id",
				"文章列表": "/api/articles",
				"单篇文章": "/api/articles/:id",
			},
		})
	})

	// 启动服务器
	log.Println("Flow示例服务器启动，监听端口: 8080")
	if err := app.Run(":8080"); err != nil {
		log.Fatal("服务器启动失败: ", err)
	}
}
