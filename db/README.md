# Flow框架数据库模块

Flow框架数据库模块提供了强大的数据库连接和ORM管理功能，支持多种数据库系统，并提供了便捷的配置方式。

## 配置格式

Flow数据库模块现在支持两种配置格式：

### 1. 嵌套格式（推荐）

```yaml
# config.yaml
database:
  default: "mysql"  # 默认连接名
  connections:
    mysql:
      driver: "mysql"
      host: "localhost"
      port: 3306
      database: "mydatabase"
      username: "root"
      password: "password"
      charset: "utf8mb4"
      # 其他配置...
    
    postgres:
      driver: "postgres"
      host: "postgres_host"
      port: 5432
      database: "postgres_db"
      username: "postgres_user"
      password: "postgres_password"
      # 其他配置...
```

### 2. 平铺格式（旧版，保留向后兼容）

```yaml
# config.yaml
database:
  driver: "mysql"
  host: "localhost"
  port: 3306
  database: "mydatabase"
  username: "root"
  password: "password"
  charset: "utf8mb4"
  # 其他配置...
```

## 在应用中使用

### 基本用法

```go
package main

import (
    "github.com/zzliekkas/flow"
    "github.com/zzliekkas/flow/config"
)

func main() {
    // 创建engine
    engine := flow.New()
    
    // 使用配置文件初始化数据库
    engine.WithConfig("config.yaml").WithDatabase()
    
    // 或者直接传入配置
    engine.WithDatabase(map[string]interface{}{
        "database": map[string]interface{}{
            "default": "mysql",
            "connections": map[string]interface{}{
                "mysql": map[string]interface{}{
                    "driver": "mysql",
                    "host": "localhost",
                    "port": 3306,
                    "database": "testdb",
                    "username": "root",
                    "password": "password",
                },
            },
        },
    })
    
    // 启动应用
    engine.Run(":8080")
}
```

### 依赖注入使用

```go
type UserRepository struct {
    DB *gorm.DB
}

func NewUserRepository(dbProvider *db.DbProvider) *UserRepository {
    return &UserRepository{
        DB: dbProvider.DB,
    }
}

// 在engine上注册
engine.Provide(NewUserRepository)
```

## 支持的数据库类型

- MySQL
- PostgreSQL
- SQLite

## 连接池配置

数据库连接池可以通过以下选项配置：

```yaml
database:
  connections:
    mysql:
      # ...基本配置
      max_idle_conns: 10     # 最大空闲连接数
      max_open_conns: 100    # 最大打开连接数
      conn_max_lifetime: 1h  # 连接最大生存时间
      conn_max_idle_time: 30m # 空闲连接最大存活时间
```

## 健康检查配置

您还可以配置数据库连接的健康检查：

```yaml
database:
  connections:
    mysql:
      # ...基本配置
      health_check: true                 # 启用健康检查
      health_check_period: 30s           # 健康检查周期
      health_check_timeout: 5s           # 健康检查超时时间
      health_check_sql: "SELECT 1"       # 健康检查SQL
```

## 错误处理

数据库模块定义了以下错误类型以便于错误处理：

- `ErrUnsupportedDriver` - 不支持的数据库驱动类型
- `ErrConnectionNotFound` - 连接未找到
- `ErrInvalidConfiguration` - 无效的数据库配置
- `ErrDatabaseNotFound` - 未找到指定的数据库
- `ErrConnectionFailed` - 数据库连接失败

## 扩展与自定义

您可以通过以下方式扩展数据库功能：

1. 创建自定义的连接配置选项
2. 实现自定义的Repository
3. 添加数据库中间件

请参考文档以获取更多详细信息。 