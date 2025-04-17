package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/zzliekkas/flow"
	"github.com/zzliekkas/flow/middleware"
)

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// 自定义错误
var (
	ErrUserNotFound = errors.New("user not found")
)

type UserService interface {
	GetUserByID(id string) (User, error)
	GetAllUsers() ([]User, error)
	CreateUser(user User) error
}

type userServiceImpl struct {
	users map[string]User
}

func NewUserService() UserService {
	// 创建一个简单的内存存储用户服务
	users := make(map[string]User)
	users["1"] = User{ID: "1", Name: "张三", Email: "zhangsan@example.com"}
	users["2"] = User{ID: "2", Name: "李四", Email: "lisi@example.com"}

	return &userServiceImpl{
		users: users,
	}
}

func (s *userServiceImpl) GetUserByID(id string) (User, error) {
	user, ok := s.users[id]
	if !ok {
		// 修复：使用自定义错误代替http.ErrNotFound
		return User{}, ErrUserNotFound
	}
	return user, nil
}

func (s *userServiceImpl) GetAllUsers() ([]User, error) {
	users := make([]User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}
	return users, nil
}

func (s *userServiceImpl) CreateUser(user User) error {
	if _, ok := s.users[user.ID]; ok {
		return http.ErrAbortHandler
	}
	s.users[user.ID] = user
	return nil
}

func main() {
	// 创建Flow应用实例
	app := flow.New()

	// 注册服务
	app.Provide(NewUserService)

	// 添加中间件
	app.Use(middleware.Logger())
	app.Use(middleware.Recovery())

	// 定义路由
	app.GET("/", func(c *flow.Context) {
		c.JSON(http.StatusOK, flow.H{
			"message": "欢迎使用Flow框架",
			"version": flow.Version,
		})
	})

	// API路由组
	api := app.Group("/api")
	{
		// 用户相关路由
		users := api.Group("/users")
		{
			// 获取所有用户
			users.GET("", func(c *flow.Context) {
				var userService UserService
				c.Inject(&userService)

				users, err := userService.GetAllUsers()
				if err != nil {
					c.JSON(http.StatusInternalServerError, flow.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, users)
			})

			// 获取指定用户
			users.GET("/:id", func(c *flow.Context) {
				id := c.Param("id")

				var userService UserService
				c.Inject(&userService)

				user, err := userService.GetUserByID(id)
				if err != nil {
					c.JSON(http.StatusNotFound, flow.H{"error": "用户不存在"})
					return
				}

				c.JSON(http.StatusOK, user)
			})

			// 创建用户
			users.POST("", func(c *flow.Context) {
				var user User
				if err := c.ShouldBindJSON(&user); err != nil {
					c.JSON(http.StatusBadRequest, flow.H{"error": err.Error()})
					return
				}

				var userService UserService
				c.Inject(&userService)

				if err := userService.CreateUser(user); err != nil {
					c.JSON(http.StatusInternalServerError, flow.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusCreated, user)
			})
		}
	}

	// 启动服务器
	log.Println("Flow示例服务器启动，监听端口: 8080")
	if err := app.Run(":8080"); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
