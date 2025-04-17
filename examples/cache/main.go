package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/zzliekkas/flow"
	"github.com/zzliekkas/flow/app"
	"github.com/zzliekkas/flow/cache"
	"github.com/zzliekkas/flow/middleware"
)

// 产品结构体
type Product struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	CreatedAt string  `json:"created_at"`
}

// 模拟数据库
var products = map[string]Product{
	"1": {ID: "1", Name: "手机", Price: 3999.00, CreatedAt: time.Now().Format(time.RFC3339)},
	"2": {ID: "2", Name: "电脑", Price: 6999.00, CreatedAt: time.Now().Format(time.RFC3339)},
	"3": {ID: "3", Name: "耳机", Price: 999.00, CreatedAt: time.Now().Format(time.RFC3339)},
}

func main() {
	// 创建Flow实例
	flowEngine := flow.New()

	// 添加中间件
	flowEngine.Use(middleware.Logger())
	flowEngine.Use(middleware.Recovery())

	// 创建应用实例
	application := app.New(flowEngine)

	// 注册缓存服务提供者
	application.RegisterProvider(cache.NewCacheProvider())

	// 设置路由
	setupRoutes(flowEngine)

	// 启动应用
	log.Println("缓存示例应用启动中，访问: http://localhost:8080")
	if err := application.Run(":8080"); err != nil {
		log.Fatalf("应用启动失败: %v", err)
	}
}

// 设置路由
func setupRoutes(e *flow.Engine) {
	// 首页
	e.GET("/", func(c *flow.Context) {
		c.JSON(http.StatusOK, flow.H{
			"message": "欢迎使用Flow缓存演示",
			"routes": []string{
				"/products - 获取所有产品（使用缓存）",
				"/products/:id - 获取单个产品（使用缓存）",
				"/products/:id/clear - 清除单个产品缓存",
				"/products/clear - 清除所有产品缓存",
				"/counter - 计数器示例",
				"/counter/increment - 增加计数器",
				"/counter/decrement - 减少计数器",
				"/tags - 标签示例",
			},
		})
	})

	// 获取所有产品（使用缓存）
	e.GET("/products", func(c *flow.Context) {
		var manager *cache.Manager
		if err := c.Inject(&manager); err != nil {
			c.JSON(http.StatusInternalServerError, flow.H{"error": "缓存服务不可用"})
			return
		}

		// 尝试从缓存获取
		ctx := context.Background()
		cacheKey := "products:all"
		cached, err := manager.Get(ctx, cacheKey)

		if err == nil {
			// 缓存命中
			c.JSON(http.StatusOK, flow.H{
				"source":   "cache",
				"products": cached,
			})
			return
		}

		// 缓存未命中，从"数据库"获取
		allProducts := make([]Product, 0, len(products))
		for _, product := range products {
			allProducts = append(allProducts, product)
		}

		// 存入缓存
		_ = manager.Set(ctx, cacheKey, allProducts, cache.WithExpiration(1*time.Minute), cache.WithTags("products"))

		c.JSON(http.StatusOK, flow.H{
			"source":   "database",
			"products": allProducts,
		})
	})

	// 获取单个产品（使用缓存）
	e.GET("/products/:id", func(c *flow.Context) {
		id := c.Param("id")

		var manager *cache.Manager
		if err := c.Inject(&manager); err != nil {
			c.JSON(http.StatusInternalServerError, flow.H{"error": "缓存服务不可用"})
			return
		}

		// 尝试从缓存获取
		ctx := context.Background()
		cacheKey := fmt.Sprintf("products:%s", id)
		cached, err := manager.Get(ctx, cacheKey)

		if err == nil {
			// 缓存命中
			c.JSON(http.StatusOK, flow.H{
				"source":  "cache",
				"product": cached,
			})
			return
		}

		// 缓存未命中，从"数据库"获取
		product, exists := products[id]
		if !exists {
			c.JSON(http.StatusNotFound, flow.H{"error": "产品不存在"})
			return
		}

		// 存入缓存
		_ = manager.Set(ctx, cacheKey, product,
			cache.WithExpiration(1*time.Minute),
			cache.WithTags("products", fmt.Sprintf("product:%s", id)))

		c.JSON(http.StatusOK, flow.H{
			"source":  "database",
			"product": product,
		})
	})

	// 清除单个产品缓存
	e.GET("/products/:id/clear", func(c *flow.Context) {
		id := c.Param("id")

		var manager *cache.Manager
		if err := c.Inject(&manager); err != nil {
			c.JSON(http.StatusInternalServerError, flow.H{"error": "缓存服务不可用"})
			return
		}

		// 清除缓存
		ctx := context.Background()
		cacheKey := fmt.Sprintf("products:%s", id)
		_ = manager.Delete(ctx, cacheKey)

		c.JSON(http.StatusOK, flow.H{
			"message": fmt.Sprintf("产品 %s 的缓存已清除", id),
		})
	})

	// 清除所有产品缓存
	e.GET("/products/clear", func(c *flow.Context) {
		var manager *cache.Manager
		if err := c.Inject(&manager); err != nil {
			c.JSON(http.StatusInternalServerError, flow.H{"error": "缓存服务不可用"})
			return
		}

		// 使用标签删除所有产品缓存
		ctx := context.Background()
		_ = manager.TaggedDelete(ctx, "products")

		c.JSON(http.StatusOK, flow.H{
			"message": "所有产品缓存已清除",
		})
	})

	// 计数器示例
	e.GET("/counter", func(c *flow.Context) {
		var manager *cache.Manager
		if err := c.Inject(&manager); err != nil {
			c.JSON(http.StatusInternalServerError, flow.H{"error": "缓存服务不可用"})
			return
		}

		// 获取计数器值
		ctx := context.Background()
		count, err := manager.Get(ctx, "counter")
		if err != nil {
			count = 0
			// 初始化计数器
			_ = manager.Set(ctx, "counter", count)
		}

		c.JSON(http.StatusOK, flow.H{
			"counter": count,
		})
	})

	// 增加计数器
	e.GET("/counter/increment", func(c *flow.Context) {
		var manager *cache.Manager
		if err := c.Inject(&manager); err != nil {
			c.JSON(http.StatusInternalServerError, flow.H{"error": "缓存服务不可用"})
			return
		}

		// 获取增量值
		value := 1
		if v := c.Query("value"); v != "" {
			if parsed, err := strconv.Atoi(v); err == nil {
				value = parsed
			}
		}

		// 增加计数器
		ctx := context.Background()
		count, _ := manager.Increment(ctx, "counter", int64(value))

		c.JSON(http.StatusOK, flow.H{
			"counter": count,
			"message": fmt.Sprintf("计数器增加了 %d", value),
		})
	})

	// 减少计数器
	e.GET("/counter/decrement", func(c *flow.Context) {
		var manager *cache.Manager
		if err := c.Inject(&manager); err != nil {
			c.JSON(http.StatusInternalServerError, flow.H{"error": "缓存服务不可用"})
			return
		}

		// 获取减量值
		value := 1
		if v := c.Query("value"); v != "" {
			if parsed, err := strconv.Atoi(v); err == nil {
				value = parsed
			}
		}

		// 减少计数器
		ctx := context.Background()
		count, _ := manager.Decrement(ctx, "counter", int64(value))

		c.JSON(http.StatusOK, flow.H{
			"counter": count,
			"message": fmt.Sprintf("计数器减少了 %d", value),
		})
	})

	// 标签示例
	e.GET("/tags", func(c *flow.Context) {
		var manager *cache.Manager
		if err := c.Inject(&manager); err != nil {
			c.JSON(http.StatusInternalServerError, flow.H{"error": "缓存服务不可用"})
			return
		}

		// 获取指定标签的所有缓存项
		ctx := context.Background()
		tag := c.DefaultQuery("tag", "products")
		items, _ := manager.TaggedGet(ctx, tag)

		c.JSON(http.StatusOK, flow.H{
			"tag":   tag,
			"items": items,
		})
	})
}
