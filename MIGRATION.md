# Flow v1 → v2 迁移指南

## 概述

Flow v2 是一次重大重构，聚焦于更清晰的架构、更好的模块化和真正的优雅关闭支持。
**大部分 API 保持向后兼容**，迁移主要涉及 import 路径变更。

---

## 1. Import 路径变更（必须）

所有 import 路径从 `github.com/zzliekkas/flow` 变为 `github.com/zzliekkas/flow/v2`：

```go
// v1
import "github.com/zzliekkas/flow"
import "github.com/zzliekkas/flow/middleware"

// v2
import "github.com/zzliekkas/flow/v2"
import "github.com/zzliekkas/flow/v2/middleware"
```

**go.mod 变更：**
```
// v1
require github.com/zzliekkas/flow v1.1.12

// v2
require github.com/zzliekkas/flow/v2 v2.0.0
```

**批量替换命令（PowerShell）：**
```powershell
Get-ChildItem -Filter "*.go" -Recurse | ForEach-Object {
    $content = Get-Content $_.FullName -Raw
    if ($content -match 'github\.com/zzliekkas/flow(?!/v2)') {
        $newContent = $content -replace 'github\.com/zzliekkas/flow(?!/v2)', 'github.com/zzliekkas/flow/v2'
        Set-Content -Path $_.FullName -Value $newContent -NoNewline
    }
}
```

---

## 2. Engine 不再是单例（行为变更）

v1 中 `flow.New()` 使用 `sync.Once`，多次调用返回同一实例。
v2 中每次调用 `flow.New()` 都返回新实例。

```go
// v2 新增：获取全局默认实例
engine := flow.Default()
```

如果你的代码依赖单例行为，改用 `flow.Default()`。

---

## 3. DI 容器访问方式变更

```go
// v1: 直接获取 dig 容器
container := engine.Container() // *dig.Container

// v2: 推荐使用增强的 DI 容器
di := engine.DI() // *di.Container — 提供 ProvideNamed, ProvideValue, Extract 等

// v2: 向后兼容仍可用
container := engine.Container() // *dig.Container（内部调用 di.Dig()）
```

---

## 4. 生命周期钩子增强

```go
// v1: 只有 OnShutdown
engine.OnShutdown(func() { ... })

// v2: 新增 OnStart，支持优先级（数值越小越先执行）
engine.OnStart(func() { ... }, 10)    // priority=10
engine.OnShutdown(func() { ... }, 20) // priority=20

// 不传优先级默认为 100
engine.OnShutdown(func() { ... })
```

---

## 5. Module 接口（新增）

v2 新增标准化的模块注册机制：

```go
type MyModule struct{}

func (m *MyModule) Name() string { return "my-module" }

func (m *MyModule) Init(e *flow.Engine) error {
    return e.Provide(func() *MyService { return NewMyService() })
}

// 可选：实现 RoutableModule 自动注册路由
func (m *MyModule) RegisterRoutes(e *flow.Engine) {
    e.GET("/my-endpoint", myHandler)
}

// 注册
engine.RegisterModule(&MyModule{})
engine.RegisterModules(&ModuleA{}, &ModuleB{})
```

---

## 6. Logger 接口（新增）

v2 提供可替换的日志接口：

```go
// 自定义日志实现
type myLogger struct { ... }
func (l *myLogger) Info(args ...interface{})                 { ... }
func (l *myLogger) Infof(format string, args ...interface{}) { ... }
// ... 实现 Debug/Warn/Error 系列方法

// 设置全局日志
flow.SetLogger(&myLogger{})

// 或通过选项
engine := flow.New(flow.WithLogger(&myLogger{}))
```

---

## 7. 数据库初始化简化

v2 移除了 `flow` 与 `db` 包之间的双向回调通信：

```go
// v1: 使用 init() 回调模式
engine.RegisterDatabaseInitializer() // 已移除

// v2: 直接使用 WithDatabase（行为不变，内部实现更简洁）
engine.WithDatabase(dbConfig)
// 或通过 Option
engine := flow.New(flow.WithDatabase(dbConfig))
```

---

## 8. Shutdown 行为改进

v2 的 `Run()` 内部创建 `http.Server`，`Shutdown()` 现在真正执行优雅关闭：

```go
// v2: Shutdown 会真正关闭 HTTP 服务器
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
engine.Shutdown(ctx) // 执行关闭钩子 + 停止 HTTP 服务器
```

---

## 9. 文件结构变更

| v1 | v2 | 说明 |
|----|-----|------|
| `flow.go` (786行) | `flow.go` (400行) | 只保留 Engine 核心 + Options |
| — | `context.go` | Context 结构体和辅助方法 |
| — | `router.go` | RouterGroup + 路由方法 |
| — | `lifecycle.go` | Run/Shutdown/生命周期 |
| — | `logger.go` | Logger 接口 + 默认实现 |
| — | `module.go` | Module 接口 |

---

## 10. 未来规划 (v2.1)

以下可选包将在 v2.1 中拆分为独立 Go 模块（参见 MODULES.md）：
- `cloud/` → `github.com/zzliekkas/flow-cloud`
- `payment/` → `github.com/zzliekkas/flow-payment`
- `storage/` → `github.com/zzliekkas/flow-storage`
- `websocket/` → `github.com/zzliekkas/flow-websocket`

当前版本中这些包仍随框架一起提供，无需额外操作。

---

## 11. 可选迁移：WebSocket 独立模块（flow-websocket）

从 v2.0.0 开始，`flow/v2/websocket` 的代码已经在独立模块中提供：

- 模块：`github.com/zzliekkas/flow-websocket`（当前版本：`v0.1.0`）

如果你希望减少对 `flow/v2` 的依赖，或希望单独升级 WebSocket 能力，可将 import 从：

```go
import "github.com/zzliekkas/flow/v2/websocket"
```

改为：

```go
import websocket "github.com/zzliekkas/flow-websocket"
```

并在 `go.mod` 增加依赖：

```
require github.com/zzliekkas/flow-websocket v0.1.0
```

本地联调可用 `replace`：

```
replace github.com/zzliekkas/flow-websocket => ../flow-websocket
```

说明：为了向后兼容，`flow/v2/websocket` 目前仍保留，后续会在 v2.x 的某个版本标记 deprecated。
