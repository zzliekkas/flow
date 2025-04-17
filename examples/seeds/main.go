package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/zzliekkas/flow/db"
	"gorm.io/gorm"
)

// 产品模型
type Product struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	Price       float64   `gorm:"not null" json:"price"`
	Inventory   int       `gorm:"not null;default:0" json:"inventory"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// 用户模型
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:50;not null" json:"name"`
	Email     string    `gorm:"size:100;uniqueIndex;not null" json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// 创建用户种子数据
func createUserSeeder(_ *gorm.DB) db.Seeder {
	return db.NewSeeder("users", func(tx *gorm.DB) error {
		users := []User{
			{Name: "张三", Email: "zhangsan@example.com"},
			{Name: "李四", Email: "lisi@example.com"},
			{Name: "王五", Email: "wangwu@example.com"},
		}
		return tx.CreateInBatches(users, 100).Error
	})
}

// 创建产品种子数据（依赖于用户数据）
func createProductSeeder(_ *gorm.DB) db.Seeder {
	return db.NewSeeder("products", func(tx *gorm.DB) error {
		products := []Product{
			{Name: "笔记本电脑", Description: "高性能笔记本电脑", Price: 6999.99, Inventory: 100},
			{Name: "智能手机", Description: "最新款智能手机", Price: 3999.99, Inventory: 200},
			{Name: "耳机", Description: "无线降噪耳机", Price: 999.99, Inventory: 500},
		}
		return tx.CreateInBatches(products, 100).Error
	}, "users") // 依赖于用户种子数据
}

func main() {
	// 获取数据库连接
	dbManager := db.NewManager()
	dbConfig := db.Config{
		Driver:   db.SQLite,
		Database: "seeds_example.db",
	}
	if err := dbManager.Register("default", dbConfig); err != nil {
		log.Fatalf("数据库注册失败: %v", err)
	}

	dbConn, err := dbManager.Connect("default")
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	// 创建表
	if err := dbConn.AutoMigrate(&User{}, &Product{}); err != nil {
		log.Fatalf("表创建失败: %v", err)
	}

	// 创建种子数据管理器
	seederManager := db.NewSeederManager(dbConn)

	// 创建并注册种子数据
	if err := seederManager.Register(createUserSeeder(dbConn)); err != nil {
		log.Fatalf("注册用户种子数据失败: %v", err)
	}
	if err := seederManager.Register(createProductSeeder(dbConn)); err != nil {
		log.Fatalf("注册产品种子数据失败: %v", err)
	}

	// 从JSON文件加载种子数据
	seedDir := "data/seeds"
	if err := os.MkdirAll(seedDir, 0755); err != nil {
		log.Fatalf("创建种子数据目录失败: %v", err)
	}

	// 创建示例JSON种子数据
	productSeedJSON := `[
		{"name": "平板电脑", "description": "高清平板电脑", "price": 2599.99, "inventory": 50},
		{"name": "智能手表", "description": "支持多种运动模式", "price": 1299.99, "inventory": 80}
	]`
	if err := os.WriteFile(filepath.Join(seedDir, "more_products.json"), []byte(productSeedJSON), 0644); err != nil {
		log.Fatalf("创建JSON种子数据文件失败: %v", err)
	}

	// 从目录加载种子数据
	if err := seederManager.LoadSeedersFromDirectory(seedDir, &Product{}); err != nil {
		log.Fatalf("加载种子数据目录失败: %v", err)
	}

	// 获取种子数据状态
	status, err := seederManager.Status()
	if err != nil {
		log.Fatalf("获取种子数据状态失败: %v", err)
	}
	fmt.Println("=== 种子数据状态 ===")
	for _, s := range status {
		fmt.Printf("名称: %s, 状态: %s\n", s["name"], s["status"])
	}

	// 执行种子数据
	fmt.Println("\n=== 执行种子数据 ===")
	if err := seederManager.Run(); err != nil {
		log.Fatalf("执行种子数据失败: %v", err)
	}
	fmt.Println("所有种子数据执行成功")

	// 查询结果
	fmt.Println("\n=== 用户数据 ===")
	var users []User
	if err := dbConn.Find(&users).Error; err != nil {
		log.Fatalf("查询用户失败: %v", err)
	}
	for _, user := range users {
		fmt.Printf("ID: %d, 名称: %s, 邮箱: %s\n", user.ID, user.Name, user.Email)
	}

	fmt.Println("\n=== 产品数据 ===")
	var products []Product
	if err := dbConn.Find(&products).Error; err != nil {
		log.Fatalf("查询产品失败: %v", err)
	}
	for _, product := range products {
		fmt.Printf("ID: %d, 名称: %s, 价格: %.2f, 库存: %d\n", product.ID, product.Name, product.Price, product.Inventory)
	}
}
