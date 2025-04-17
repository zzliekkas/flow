# Flow 缓存系统示例

这个示例展示了 Flow 框架中缓存系统的使用方法。缓存系统是一个高性能、灵活的缓存解决方案，支持多种驱动、标签管理和缓存管理器。

## 功能特点

1. **统一的缓存接口**：所有缓存驱动实现相同的接口，便于切换和测试
2. **多驱动支持**：内置支持内存缓存，可扩展支持 Redis 等
3. **标签管理**：通过标签关联多个缓存项，便于批量操作
4. **缓存管理器**：管理多个缓存存储，支持默认存储和命名存储
5. **键前缀**：支持键前缀，避免键冲突
6. **计数器操作**：内置支持原子递增/递减操作
7. **自动GC**：自动清理过期缓存项

## 示例内容

本示例展示了以下功能：

1. **基本缓存操作**：设置、获取、删除缓存
2. **标签管理**：通过标签批量管理缓存项
3. **计数器功能**：递增和递减计数器
4. **缓存失效**：手动和自动缓存失效

## 目录结构

```
cache/
├── cache.go      # 核心接口定义
├── memory.go     # 内存缓存驱动
├── tag.go        # 缓存标签系统
├── manager.go    # 缓存管理器
└── provider.go   # 缓存服务提供者
```

## 运行示例

```bash
go run examples/cache/main.go
```

然后访问 http://localhost:8080 查看示例。

## 使用方法

### 1. 注册缓存服务提供者

```go
// 创建应用实例
application := app.New(flowEngine)

// 注册缓存服务提供者
application.RegisterProvider(cache.NewCacheProvider())
```

### 2. 配置缓存

在 `config.yaml` 中配置缓存：

```yaml
cache:
  default: "memory"
  stores:
    memory:
      driver: "memory"
      ttl: "5m"
    redis:
      driver: "redis"
      host: "localhost"
      port: 6379
      database: 0
      prefix: "flow:"
      ttl: "10m"
```

### 3. 在控制器中使用缓存

```go
// 在控制器中注入缓存管理器
var manager *cache.Manager
if err := c.Inject(&manager); err != nil {
    return err
}

// 设置缓存
ctx := context.Background()
err := manager.Set(ctx, "key", "value", 
    cache.WithExpiration(5*time.Minute),
    cache.WithTags("tag1", "tag2"))

// 获取缓存
value, err := manager.Get(ctx, "key")

// 删除缓存
err := manager.Delete(ctx, "key")

// 使用标签删除
err := manager.TaggedDelete(ctx, "tag1")

// 使用计数器
count, err := manager.Increment(ctx, "counter", 1)
```

## 扩展缓存系统

你可以通过实现 `cache.Driver` 和 `cache.Store` 接口来添加新的缓存驱动。然后使用 `cache.RegisterDriver()` 注册你的驱动。 