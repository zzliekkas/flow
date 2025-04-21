# Flow框架开发环境增强模块

这是Flow框架的开发环境增强模块，为开发者提供了一系列工具和功能，以提高开发效率和调试能力。

## 功能特点

- **热重载支持**：自动监听文件变更并重新加载应用
- **调试信息增强**：提供详细的日志记录和调试信息展示
- **开发服务器**：集成的开发服务器，支持静态文件、代理和调试功能
- **开发环境配置**：灵活的配置选项，适应不同的开发需求

## 快速开始

### 基本使用

```go
package main

import (
    "github.com/zzliekkas/flow/dev"
)

func main() {
    // 创建配置
    config := dev.NewConfig()
    
    // 配置热重载
    config.EnableHotReload(true)
    config.AddWatchDir("./src")
    
    // 创建开发服务器
    server, err := dev.NewDevServer(config)
    if err != nil {
        panic(err)
    }
    
    // 启动服务器
    if err := server.Start(); err != nil {
        panic(err)
    }
    
    // 保持运行直到程序结束
    select {}
}
```

### 使用热重载功能

```go
// 创建热重载器
config := dev.NewConfig()
reloader, err := dev.NewHotReloader(config)
if err != nil {
    panic(err)
}

// 配置热重载器
reloader.AddFileExtension(".go", ".html", ".css", ".js")
reloader.OnReload(func() {
    // 重新编译和运行应用的逻辑
    println("检测到变更，正在重新加载...")
})

// 启动热重载器
if err := reloader.Start(); err != nil {
    panic(err)
}
```

### 使用调试日志记录器

```go
// 创建日志记录器
logger := dev.NewDebugLogger(nil)

// 设置日志文件
err := logger.SetLogFile("./logs/debug.log")
if err != nil {
    panic(err)
}

// 记录不同级别的日志
logger.Debug("这是一条调试日志")
logger.Info("这是一条信息日志")
logger.Warn("这是一条警告日志")
logger.Error("这是一条错误日志")

// 记录SQL查询
logger.LogSQL("SELECT * FROM users", nil, 10*time.Millisecond, nil)
```

## 详细配置

### 配置选项

开发环境支持以下配置选项：

#### 热重载配置

- `EnableHotReload(bool)`: 启用/禁用热重载
- `AddWatchDir(string)`: 添加要监视的目录
- `SetIgnorePatterns(...string)`: 设置要忽略的文件模式
- `SetReloadDelay(time.Duration)`: 设置重载延迟时间

#### 调试配置

- `EnableVerboseLogging(bool)`: 启用/禁用详细日志
- `EnableShowRoutes(bool)`: 启用/禁用路由信息显示
- `EnableShowSQL(bool)`: 启用/禁用SQL查询显示
- `EnableShowHTTP(bool)`: 启用/禁用HTTP请求和响应显示
- `EnableProfiler(bool)`: 启用/禁用性能分析
- `SetProfilerPort(int)`: 设置性能分析端口
- `EnableConsole(bool)`: 启用/禁用调试控制台
- `SetConsolePort(int)`: 设置控制台端口

#### 服务器配置

- `SetServerHost(string)`: 设置服务器地址
- `SetServerPort(int)`: 设置服务器端口
- `EnableHTTPS(bool)`: 启用/禁用HTTPS
- `EnableOpenBrowser(bool)`: 启用/禁用自动打开浏览器
- `SetStaticDir(string)`: 设置静态文件目录
- `AddProxy(path, target string)`: 添加代理设置
- `EnableWebSocket(bool)`: 启用/禁用WebSocket
- `EnableCORS(bool)`: 启用/禁用跨域
- `AddHeader(key, value string)`: 添加自定义响应头

## 开发服务器功能

开发服务器提供以下主要功能：

1. **静态文件服务**：自动提供静态文件访问
2. **API代理**：支持将请求代理到其他服务器
3. **热重载集成**：自动检测文件变更并重启应用
4. **调试路由**：提供调试信息查看和控制
5. **性能分析**：集成性能分析工具

## 调试功能

调试日志记录器提供以下功能：

1. **多级别日志**：支持Debug、Info、Warn、Error、Fatal等级别
2. **路由记录**：记录注册的路由信息
3. **SQL查询记录**：记录执行的SQL查询、参数和耗时
4. **HTTP请求记录**：记录请求和响应的详细信息
5. **调试信息导出**：支持导出完整的调试信息

## 最佳实践

1. **仅在开发环境使用**：这些功能主要用于开发过程，生产环境应禁用或移除
2. **合理配置监视目录**：只监视必要的目录，避免监视过多文件导致性能问题
3. **使用日志文件**：对于长时间运行的开发会话，建议将日志输出到文件
4. **利用性能分析**：定期检查性能分析结果，及早发现潜在问题

## 常见问题

### 热重载不工作

可能的原因：
- 监视的目录不正确
- 文件变更太频繁，被合并处理
- 系统限制了文件监视数量

解决方法：
- 检查配置的监视目录
- 增加重载延迟时间
- 调整系统文件监视限制 (如Linux上的inotify限制)

### 服务器无法启动

可能的原因：
- 端口被占用
- 权限不足
- 配置错误

解决方法：
- 更改服务器端口
- 以适当权限运行
- 检查配置参数 