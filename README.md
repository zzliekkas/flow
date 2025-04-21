# Flow

[![Go Report Card](https://goreportcard.com/badge/github.com/zzliekkas/flow)](https://goreportcard.com/report/github.com/zzliekkas/flow)
[![GoDoc](https://godoc.org/github.com/zzliekkas/flow?status.svg)](https://godoc.org/github.com/zzliekkas/flow)
[![License](https://img.shields.io/github/license/zzliekkas/flow.svg)](https://github.com/zzliekkas/flow/blob/main/LICENSE)

> 优雅、简洁、高效的Go Web应用框架

Flow是一个基于[Gin](https://github.com/gin-gonic/gin)优化的Go语言Web应用框架，旨在提供流畅、简洁而强大的开发体验。Flow框架保持了Gin的高性能特性，同时增加了更多实用功能，使开发者能够更专注于业务逻辑而非框架细节。

## 目录

- [特性](#特性)
- [安装](#安装)
- [快速开始](#快速开始)
- [详细文档](#详细文档)
- [设计理念](#设计理念)
  - [为什么选择Gin作为基础](#为什么选择gin作为基础)
  - [核心设计思想](#核心设计思想)
  - [技术选型理由](#技术选型理由)
- [开发进展](#开发进展)
  - [已完成工作](#已完成工作)
  - [当前状态](#当前状态)
  - [未来计划](#未来计划)
- [项目结构说明](#项目结构说明)
- [贡献](#贡献)
- [许可证](#许可证)

## 特性

- **优雅简洁**：简化配置和路由定义，提供流畅的API设计
- **中间件系统**：强大而灵活的中间件机制，轻松扩展应用功能
- **依赖注入**：内置简洁的依赖注入容器，方便管理服务依赖
- **ORM集成**：无缝集成常用ORM，简化数据库操作
- **日志系统**：可配置的日志记录，支持多种输出方式
- **配置管理**：灵活的配置加载和环境变量支持
- **错误处理**：统一的错误处理机制，提高代码质量
- **完善文档**：详尽的文档和示例代码

## 安装

```bash
go get -u github.com/zzliekkas/flow
```

## 快速开始

创建一个简单的Flow应用只需几行代码：

```go
package main

import (
    "github.com/zzliekkas/flow"
)

func main() {
    app := flow.New()
    
    app.GET("/hello", func(c *flow.Context) {
        c.JSON(200, flow.H{
            "message": "Hello, Flow!",
        })
    })
    
    app.Run(":8080")
}
```

## 详细文档

### 路由组和中间件

```go
package main

import (
    "github.com/zzliekkas/flow"
    "github.com/zzliekkas/flow/middleware"
)

func main() {
    app := flow.New()
    
    // 全局中间件
    app.Use(middleware.Logger())
    app.Use(middleware.Recovery())
    
    // 路由组
    api := app.Group("/api")
    {
        api.GET("/users", GetUsers)
        api.POST("/users", CreateUser)
        
        // 嵌套组
        auth := api.Group("/auth")
        auth.Use(middleware.JWT())
        {
            auth.GET("/profile", GetProfile)
            auth.PUT("/profile", UpdateProfile)
        }
    }
    
    app.Run(":8080")
}
```

### 依赖注入

Flow提供简洁的依赖注入容器：

```go
package main

import (
    "github.com/zzliekkas/flow"
    "github.com/zzliekkas/flow/di"
)

type UserService interface {
    GetUser(id string) (User, error)
}

type userServiceImpl struct {
    // ...
}

func main() {
    app := flow.New()
    
    // 注册服务
    app.Provide(func() UserService {
        return &userServiceImpl{}
    })
    
    app.GET("/users/:id", func(c *flow.Context) {
        var userService UserService
        c.Inject(&userService)
        
        user, err := userService.GetUser(c.Param("id"))
        if err != nil {
            c.Error(err)
            return
        }
        
        c.JSON(200, user)
    })
    
    app.Run(":8080")
}
```

## 设计理念

### 为什么选择Gin作为基础

Flow框架选择基于Gin进行开发有以下几个关键原因：

1. **性能卓越**：Gin是Go语言中性能最出色的Web框架之一，采用定制的httprouter提供极快的HTTP路由。在我们追求高效率的目标中，这是不可或缺的基础。

2. **API成熟度**：Gin有着简洁而成熟的API设计，已被广泛采用并验证。我们不希望重新发明轮子，而是在这个优秀的基础上进行增强。

3. **活跃的生态系统**：Gin拥有活跃的社区和丰富的中间件生态，这为Flow框架提供了良好的兼容性和扩展性基础。

4. **学习曲线低**：相比其他框架，Gin的学习门槛较低，这使得开发者可以快速上手Flow框架。

### 核心设计思想

Flow框架的设计遵循以下核心理念：

1. **扩展而非替代**：我们采用了组合模式来扩展Gin，而非完全重写。这样做的原因是既可以保持Gin的优势，又能添加我们认为缺失的功能。具体实现上，我们通过嵌入`gin.Engine`和`gin.Context`，然后扩展它们的功能。

2. **依赖注入优先**：现代应用开发中，依赖管理是一个核心问题。传统的工厂模式或全局变量方式管理依赖既不优雅也不易测试，因此我们选择了依赖注入模式，并集成了Uber的dig库作为容器。这大大简化了服务的注册、管理和获取，提高了代码的可测试性和可维护性。

3. **约定优于配置**：我们相信良好的默认值可以大幅减少样板代码。Flow提供了合理的默认配置，使开发者可以快速启动项目，同时保留了灵活的自定义能力。

4. **统一错误处理**：错误处理散布在各处是造成代码混乱的主要原因之一。Flow提供了统一的错误处理机制，使错误管理更加一致和可控。

5. **渐进式架构**：框架设计采用渐进式架构，核心功能轻量化，高级功能按需引入。这样既保证了性能，又提供了丰富的功能选择。

### 技术选型理由

Flow框架在各个组件的技术选型上，我们经过了慎重考虑：

1. **依赖注入容器 - Uber的dig**：选择dig而非wire或inject的原因是dig提供了运行时依赖注入，更灵活且无需代码生成步骤。虽然牺牲了一些编译时安全性，但大大提高了开发效率和灵活性。

2. **配置管理 - Viper**：Viper是Go生态中最成熟的配置管理库，支持多种配置源（文件、环境变量、远程配置服务等），提供了配置热重载等高级功能。这些都是现代应用所必需的。

3. **日志系统 - Logrus**：Logrus提供了结构化日志和多种输出格式，便于日志收集和分析。相比标准库的log包，它提供了更丰富的功能，如字段记录、日志级别等。

4. **验证工具 - validator**：数据验证是Web应用的重要环节，validator提供了声明式的标签验证，简化了数据验证逻辑，提高了代码可读性。

## 开发进展

### 已完成工作

Flow框架的开发遵循模块化设计，目前已完成以下核心功能：

#### Phase 1: 核心框架设计与实现 ✅

1. **核心引擎设计**
   - 定义框架基本概念和接口
   - 实现Engine结构，封装Gin引擎
   - 扩展Context结构，增强请求处理能力
   
   **为什么这样做**：核心引擎是框架的基础，我们首先需要设计清晰的接口和结构，以支持后续功能的构建。通过封装而非继承Gin，我们可以在保持兼容性的同时增加新功能。

2. **基础路由系统**
   - 实现HTTP方法处理(GET, POST, PUT, DELETE等)
   - 支持路由组和嵌套分组
   - 路由参数和查询参数处理
   
   **为什么这样做**：路由系统是Web框架的核心组件，我们保留了Gin的高效路由机制，同时扩展了HandlerFunc定义，使其能够与我们的Context协同工作。路由组的实现使API组织更加清晰和模块化。

3. **中间件机制**
   - 日志中间件(Logger)实现
   - 恢复中间件(Recovery)实现
   - 统一中间件接口设计
   
   **为什么这样做**：中间件是处理横切关注点的有效方式，如日志记录、错误恢复、认证等。我们实现了与Gin兼容的中间件接口，同时针对日志和恢复等核心功能提供了默认实现，减轻开发者的负担。

4. **配置管理**
   - 基于viper的配置加载系统
   - 环境变量支持
   - 多环境配置(开发、测试、生产)
   
   **为什么这样做**：现代应用需要灵活的配置管理，能够适应不同环境的需求。通过集成viper，我们提供了强大的配置加载能力，支持YAML/JSON/TOML等多种格式，并能从环境变量中覆盖配置，这对于容器化部署尤为重要。

5. **依赖注入容器**
   - 基于dig的DI系统
   - 服务注册与获取
   - 命名服务支持
   
   **为什么这样做**：依赖注入是实现松耦合和可测试代码的关键。通过提供简洁的DI容器，我们简化了服务之间的依赖关系管理，使代码结构更加清晰，测试更加容易。

6. **错误处理机制**
   - HTTP错误类型定义
   - 统一错误响应格式
   - 友好错误信息生成
   
   **为什么这样做**：一致的错误处理对于API的可用性和可维护性至关重要。我们定义了标准的HTTP错误类型和响应格式，使错误处理更加一致，客户端解析更加容易。

7. **示例应用**
   - RESTful API示例
   - 展示框架主要功能
   - 用户服务实现示例
   
   **为什么这样做**：示例是学习框架的最佳途径，我们提供了完整的示例应用，展示框架的最佳实践和常见使用场景，帮助开发者快速上手。

8. **缓存系统**
   - 缓存接口设计
   - 内存缓存驱动
   - 缓存标签系统
   - 缓存管理器
   - 服务提供者
   
   **为什么这样做**：缓存是提高性能的关键手段，统一的缓存接口使应用可以灵活切换缓存实现而无需修改业务代码。支持内存和Redis等多种驱动满足不同场景的需求，而缓存标签则便于管理相关缓存项的失效。

### 当前状态

目前Flow框架的目录结构如下：

```
flow/
├── app/                # 应用管理
│   ├── bootstrap.go    # 应用引导
│   ├── lifecycle.go    # 生命周期管理
│   ├── environment.go  # 环境信息
│   ├── hooks.go        # 钩子系统
│   ├── provider.go     # 服务提供者
│   └── singleton.go    # 单例管理
├── cache/              # 缓存系统
│   ├── cache.go        # 缓存接口
│   ├── memory.go       # 内存驱动
│   ├── redis.go        # Redis驱动
│   ├── file.go         # 文件驱动
│   ├── tag.go          # 标签系统
│   ├── manager.go      # 缓存管理器
│   └── provider.go     # 缓存提供者
├── config/             # 配置管理
│   ├── config.go       # 配置加载
│   └── app.yaml        # 示例配置
├── di/                 # 依赖注入
│   └── container.go    # DI容器实现
├── event/              # 事件系统
│   ├── event.go        # 事件接口
│   ├── dispatcher.go   # 事件分发器
│   ├── listener.go     # 事件监听器
│   ├── manager.go      # 事件管理器
│   └── provider.go     # 事件提供者
├── middleware/         # 中间件
│   ├── logger.go       # 日志中间件
│   ├── recovery.go     # 恢复中间件
│   ├── cors.go         # CORS中间件
│   ├── jwt.go          # JWT认证中间件
│   ├── csrf.go         # CSRF保护中间件
│   ├── ratelimit.go    # 速率限制中间件
│   └── record.go       # 请求记录中间件
├── utils/              # 工具函数
│   └── validator.go    # 请求验证
├── examples/           # 示例应用
│   ├── simple/         # 简单API示例
│   ├── app/            # 应用管理示例
│   └── cache/          # 缓存系统示例
├── flow.go             # 框架核心
├── error.go            # 错误处理
├── go.mod              # Go模块定义
├── LICENSE             # MIT许可证
└── README.md           # 本文档
```

框架已具备基本Web应用开发能力，可用于构建API服务和简单Web应用。最近完成的工作包括：

1. **错误修复与代码优化**：
   - 修复了`cache/file.go`中的未使用参数问题
   - 解决了中间件模块中的错误常量冲突，通过重命名JWT和CSRF错误常量避免重复声明
   - 优化了JWT错误处理机制
   - 完善了依赖包管理

2. **文件缓存驱动**：
   - 完成了文件系统缓存存储的实现
   - 支持基本的缓存操作，包括获取、设置、删除等
   - 实现了自动垃圾回收机制清理过期缓存
   - 集成了缓存标签系统
   - 提供了文件锁定机制确保并发安全

框架目前处于稳定可用的状态，所有计划的核心功能已经实现完成。

### 未来计划

Flow框架的后续开发计划分为以下几个阶段：

#### Phase 2: 核心功能增强 ✅

1. **应用容器增强** ✅
   - [x] 应用容器完善
   - [x] 服务提供者机制
   - [x] 应用钩子系统
   
   **为什么这样做**：应用容器是框架的核心，负责协调各组件的生命周期和依赖关系。通过完善的容器设计，我们提供了更灵活的应用引导流程和组件注册机制。服务提供者模式让功能模块化，开发者可以按需启用功能，降低了不必要的性能开销。钩子系统则让应用能在关键生命周期事件（启动、关闭等）执行自定义逻辑，极大提高了框架的可扩展性和灵活性。

2. **事件系统** ✅
   - [x] 事件接口定义
   - [x] 事件分发器
   - [x] 事件监听器
   - [x] 异步事件处理
   - [x] 事件管理器
   - [x] 事件服务提供者
   
   **为什么这样做**：事件驱动架构是实现松耦合系统的关键模式。通过事件系统，应用的不同部分可以在不直接依赖的情况下进行通信，一个操作可以触发多个后续行为而无需硬编码这些关系。异步事件处理能力使得耗时操作可以在后台执行，提高应用响应性。事件系统是构建可扩展应用的基石，也为后续的WebSocket和队列系统提供了基础架构支持。

3. **缓存系统** ✅
   - [x] 缓存接口设计
   - [x] 内存缓存驱动
   - [x] 缓存标签系统
   - [x] 缓存管理器
   - [x] 缓存服务提供者
   - [x] Redis 缓存驱动
   - [x] 文件缓存驱动
   
   **为什么这样做**：缓存是提升应用性能的关键技术，尤其在数据频繁读取但不常变化的场景。我们设计了统一的缓存接口，支持多种存储后端，从简单的内存缓存到分布式Redis解决方案。缓存标签系统让相关缓存项能一起失效，解决了缓存一致性难题。多驱动支持则让应用能根据规模和需求选择合适的缓存策略，在开发环境使用文件缓存，生产环境无缝切换至Redis，而无需修改业务代码。

4. **扩展中间件** ✅
   - [x] CORS中间件 (`middleware/cors.go`)
   - [x] JWT认证中间件 (`middleware/jwt.go`)
   - [x] 请求速率限制 (`middleware/ratelimit.go`)
   - [x] 请求/响应记录 (`middleware/record.go`)
   - [x] CSRF保护 (`middleware/csrf.go`)
   
   **为什么这样做**：中间件是Web框架的核心组件，处理横切关注点如安全、日志等。CORS中间件解决了现代Web应用中的跨域问题；JWT实现了无状态认证，特别适合API和微服务架构；速率限制保护应用免受过载和滥用；请求记录提供了可追溯性和调试能力；CSRF保护则防止了常见的Web安全漏洞。这些高质量中间件使开发者能专注于业务逻辑，同时确保应用的安全性和稳定性。

#### Phase 3: 数据访问与集成 🔄 (当前阶段)

1. **数据库集成** (优先级：高)
   - [x] 数据库连接管理 (`db/connection.go`)
   - [x] GORM集成 (`db/gorm.go`)
   - [x] 查询构建器 (`db/query.go`)
   - [x] 迁移工具 (`db/migration.go`)
   - [x] 种子数据 (`db/seeder.go`)

   **为什么这样做**：数据持久化是几乎所有应用的核心需求。我们选择了GORM作为ORM基础，因其在Go社区的广泛采用和成熟度高。通过提供统一的连接管理和查询构建器，开发者可以轻松切换底层数据库而无需修改业务代码。迁移工具和种子数据功能则简化了数据库结构管理和测试数据准备工作，特别适合团队协作和CI/CD环境中使用，确保开发、测试和生产环境的一致性。

2. **测试支持** (优先级：高)
   - [x] 测试助手函数 (`test/helper.go`)
   - [x] HTTP测试客户端 (`test/http.go`)
   - [x] 模拟数据生成 (`test/mock.go`)
   - [x] 单元测试工具 (`test/unit.go`)
   - [x] 集成测试工具 (`test/integration.go`)

   **为什么这样做**：测试是保障代码质量的关键环节，但在Web应用中编写测试常常比较繁琐。我们提供了一系列测试工具，大幅简化HTTP请求模拟、响应断言、数据库状态验证等常见测试场景。模拟数据生成器让创建测试数据变得简单而灵活，单元和集成测试工具则提供了针对性的辅助功能，鼓励开发者养成良好的测试习惯，提高代码质量和可维护性。

3. **仓储模式支持** (优先级：中)
   - [x] 仓储接口定义 (`db/repository.go`)
   - [x] 基础仓储实现 (`db/base_repository.go`)
   - [x] 事务管理 (`db/transaction.go`)

   **为什么这样做**：仓储模式是数据访问的最佳实践之一，它封装了数据存储细节，提供面向领域模型的接口，降低了业务逻辑对数据访问层的依赖。我们实现了通用的仓储接口和基类，让开发者能够轻松实现针对特定模型的仓储，同时支持事务管理，确保数据一致性。这种抽象不仅使代码更易测试，还增强了可维护性，允许底层存储技术在不影响业务逻辑的情况下进行替换。

4. **数据验证增强** (优先级：中)
   - [x] 自定义验证规则 (`validation/rules.go`)
   - [x] 领域验证器 (`validation/domain.go`)
   - [x] 错误消息本地化 (`validation/messages.go`)

   **为什么这样做**：数据验证是确保应用安全和数据完整性的重要环节。我们在标准验证器基础上增加了自定义规则支持，使开发者能够实现特定业务场景的验证逻辑。领域验证器将验证规则与业务模型关联，实现验证逻辑的集中管理。错误消息本地化则提升了用户体验，向不同语言的用户提供更友好的反馈。这些增强功能共同减少了重复代码，提高了验证逻辑的可读性和可维护性。

#### Phase 4: 认证与安全 (已完成) ✅

1. **认证系统框架** (优先级：高) ✅
   - [x] 统一认证抽象层 (`auth/authenticatable.go`)
   - [x] 认证服务提供者 (`auth/provider.go`)
   - [x] 现有JWT中间件集成 (`auth/drivers/jwt_adapter.go`)
   - [x] 会话认证驱动 (`auth/drivers/session.go`)
   - [x] OAuth2客户端 (`auth/drivers/oauth.go`)
   - [x] 社交登录集成 (`auth/drivers/social.go`)
   
   **完成情况**：认证系统框架已完全实现，包括可扩展的认证接口设计、多种认证驱动支持和丰富的集成能力。该框架支持基于JWT的无状态认证、传统的会话认证、OAuth2协议和主流社交平台登录（GitHub、Google、微信等）。核心抽象层确保了不同认证方式之间的一致API，使应用可以轻松切换或组合多种认证策略。

2. **授权系统** (优先级：高) ✅
   - [x] 权限定义 (`auth/permission.go`)
   - [x] 角色管理 (`auth/role.go`)
   - [x] 基于策略的授权 (`auth/policy.go`)
   - [x] 授权中间件 (`middleware/authorize.go`)
   
   **完成情况**：授权系统已完全实现，提供了灵活的权限控制机制。系统支持基于角色的访问控制(RBAC)和更细粒度的基于策略的授权。通过权限定义和角色管理，可以构建复杂的授权层次结构；而基于策略的授权则允许针对特定资源和操作定义自定义授权逻辑，适应各种业务场景需求。

3. **安全框架** (优先级：中) ✅
   - [x] 安全服务提供者 (`security/provider.go`)
   - [x] 现有安全中间件集成 (`security/middleware_integration.go`)
   - [x] XSS防护 (`security/xss.go`)
   - [x] 密码策略管理 (`security/password_policy.go`)
   - [x] 安全头部配置 (`security/headers.go`)
   - [x] 内容安全策略 (`security/csp.go`)
   - [x] 安全配置管理 (`security/config.go`)
   - [x] 安全审计日志 (`security/audit.go`)
   
   **完成情况**：安全框架已全部实现完成，为Flow应用提供了全面的Web安全保障。框架支持内容安全策略(CSP)、XSS防护、HSTS和其他关键安全头部、密码策略验证以及详细的安全审计日志记录。这些功能均采用配置驱动的方式，可以根据应用需求灵活开启和调整，同时提供合理的默认配置，确保即使在最小配置下也能维持基本的安全性。

#### Phase 5: 高级功能 (计划中)

1. **任务队列和调度** (优先级：高)
   - [ ] 队列接口设计 (`queue/queue.go`)
   - [ ] 与事件系统集成 (`queue/event_integration.go`)
   - [ ] 多驱动支持(Redis、RabbitMQ) (`queue/drivers/`)
   - [ ] 任务调度器 (`scheduler/scheduler.go`)
   - [ ] 定时任务 (`scheduler/cron.go`)
   - [ ] 失败任务重试机制 (`queue/retry.go`)
   
   **为什么这样做**：现代应用常需处理大量耗时任务，如邮件发送、报表生成等。任务队列可将这些操作从主请求周期中分离出来，显著提升应用响应性。我们设计了与事件系统集成的队列架构，让开发者能够无缝地将事件转化为异步任务，同时提供定时任务和失败重试机制，确保任务可靠执行。支持多种驱动让应用可以根据规模选择合适的队列实现。

2. **WebSocket支持** (优先级：中)
   - [ ] WebSocket管理器 (`websocket/manager.go`)
   - [ ] 频道系统 (`websocket/channel.go`)
   - [ ] 与事件系统集成 (`websocket/event_integration.go`)
   - [ ] 客户端API (`websocket/client.go`)
   
   **为什么这样做**：实时应用已成为现代Web体验的重要部分，WebSocket是实现这一需求的关键技术。通过集成事件系统和频道机制，我们让开发者能轻松构建聊天室、实时通知等功能，而无需关心底层连接管理的复杂性。管理器负责连接生命周期，频道系统处理消息路由，客户端API则简化前端集成，提供完整的实时通信解决方案。

3. **文件存储系统** (优先级：中)
   - [ ] 统一存储接口 (`storage/filesystem.go`)
   - [ ] 与缓存系统区分的文件管理 (`storage/manager.go`)
   - [ ] 云存储驱动(S3、OSS等) (`storage/cloud/`)
   - [ ] 文件元数据管理 (`storage/metadata.go`)
   - [ ] 文件上传助手 (`storage/uploader.go`)
   
   **为什么这样做**：几乎所有应用都需要管理文件，从简单的头像上传到复杂的文档管理。统一的存储接口让应用代码不需关心底层存储位置，可以无缝地从本地迁移到云存储。文件元数据管理和上传助手则大大简化了文件处理流程，处理安全、验证、格式转换等常见需求，使开发者专注于业务功能而非技术细节。

4. **国际化与本地化** (优先级：低)
   - [ ] 通用国际化接口 (`i18n/translator.go`)
   - [ ] 多语言文件管理 (`i18n/manager.go`)
   - [ ] 复数和日期格式化 (`i18n/formatter.go`)
   - [ ] 本地化中间件 (`middleware/locale.go`)
   - [ ] 与验证系统集成 (`i18n/validation_integration.go`)
   
   **为什么这样做**：全球化应用需要适应不同语言和地区的用户需求。我们的国际化系统不只是简单的文本翻译，还包括复数形式处理、日期格式化等复杂本地化需求。与验证系统的集成确保了错误消息也能正确翻译，提供一致的用户体验。通过本地化中间件，应用可以自动检测用户首选语言，实现无缝的多语言支持。

5. **云原生支持** (优先级：中)
   - [ ] 容器化配置适配 (`cloud/container.go`)
   - [ ] 健康检查机制 (`cloud/health.go`)
   - [ ] 优雅启动/关闭 (`cloud/lifecycle.go`)
   - [ ] 分布式追踪集成 (`cloud/tracing.go`)
   - [ ] 云服务提供商适配器 (`cloud/providers/`)
   
   **为什么这样做**：现代应用越来越多地部署在容器和Kubernetes等环境中。云原生支持模块提供了适应这类环境所需的功能，如健康检查端点、优雅的启动和关闭流程等。分布式追踪集成让开发者可以监控和分析微服务架构中的请求流转，提高系统可观测性。这些功能让Flow应用能够无缝地适应从单体到微服务的各种部署模式，满足各类规模的项目需求。

#### Phase 6: 开发者工具 (计划中)

1. **CLI工具** (优先级：中)
   - [ ] 命令行应用框架 (`cli/app.go`)
   - [ ] 代码生成器 (`cli/generator/`)
   - [ ] 数据库工具命令集成 (`cli/commands/db.go`)
   - [ ] 服务管理命令 (`cli/commands/server.go`)
   - [ ] 资源创建命令 (`cli/commands/make.go`)
   
   **为什么这样做**：命令行工具极大地提高了开发效率，能够自动化许多重复性工作。通过代码生成器，开发者可以快速创建标准化的控制器、模型和服务；数据库工具命令让迁移和种子数据操作变得简单；资源创建命令更是遵循"约定优于配置"的理念，帮助开发者保持代码一致性。这套CLI工具体系将极大地缩短从想法到实现的时间，同时提高代码质量和一致性。

2. **文档生成** (优先级：中)
   - [ ] API文档生成器 (`docs/api.go`)
   - [ ] Swagger集成 (`docs/swagger.go`)
   - [ ] 模型文档 (`docs/model.go`)
   - [ ] 交互式文档UI (`docs/ui/`)
   
   **为什么这样做**：好的文档是优秀项目的标志，也是团队协作的基础。我们的文档生成工具可以直接从代码注释和结构生成API文档，确保文档与代码的一致性。Swagger集成提供了交互式API测试能力，大大简化了前后端协作过程。模型文档则帮助开发者理解复杂的数据结构关系，这些工具共同解决了文档频繁过时的痛点，使开发体验更加流畅。

3. **性能分析** (优先级：低)
   - [ ] 请求性能记录 (`profiler/request.go`)
   - [ ] 数据库查询分析 (`profiler/database.go`)
   - [ ] 内存使用分析 (`profiler/memory.go`)
   - [ ] 性能报告生成 (`profiler/report.go`)
   - [ ] 分析数据导出 (`profiler/export.go`)
   - [ ] 性能指标仪表板 (`profiler/dashboard.go`)
   
   **为什么这样做**：性能问题往往是应用成功的关键障碍，而定位性能瓶颈通常十分困难。我们的性能分析工具集提供了从请求到数据库再到内存使用的全方位监控，自动标记出慢查询和异常请求。通过直观的仪表板展示，开发者可以快速发现并定位性能问题，进行有针对性的优化。这套工具既适用于开发环境的早期发现，也能在生产环境提供持续监控，确保应用始终保持最佳性能。

4. **开发环境增强** (优先级：低)
   - [ ] 热重载支持 (`dev/reload.go`)
   - [ ] 调试信息增强 (`dev/debug.go`)
   - [ ] 开发服务器 (`dev/server.go`)
   - [ ] 开发环境配置 (`dev/config.go`)
   
   **为什么这样做**：高效的开发环境能显著提升开发体验和效率。热重载支持让代码修改即时生效，无需手动重启服务；增强的调试信息提供了更多上下文，加速问题定位；专用的开发服务器则整合了多种开发辅助功能，如自动路由列表、中间件可视化等。这些工具共同打造了一个响应迅速、信息丰富的开发环境，让开发者能够更专注于业务逻辑而非繁琐的操作流程。

#### 下一步工作计划

根据框架发展路线和用户反馈，我们已完成了数据库相关功能开发和认证基础框架，接下来的重点将是：

1. **完成测试支持模块**：提供完整的测试工具集，简化单元测试和集成测试
2. **完善仓储模式**：增强数据库抽象层，包括基础仓储实现和事务管理
3. **完成安全框架**：实现XSS防护、内容安全策略、安全头部配置和审计日志
4. **实现云原生支持**：增加容器化配置和健康检查等功能，以适应现代部署环境

这些功能将使Flow框架在企业级应用开发中更具竞争力，同时保持其简洁易用的特性。

## 项目结构说明

Flow框架的项目结构经过精心设计，旨在提供清晰的代码组织和职责分离：

1. **app/** - 应用容器和生命周期管理
   - 这是框架的核心管理层，负责协调各个组件的初始化和关闭
   - 提供应用级别的事件和钩子机制

2. **cache/** - 缓存系统
   - 提供统一的缓存接口和多种驱动实现
   - 包含内存、Redis和文件缓存驱动
   - 支持缓存标签和自动失效机制

3. **config/** - 配置管理
   - 集中管理所有配置相关功能
   - 提供灵活的配置加载和环境变量支持

4. **db/** - 数据库访问
   - 数据库连接和查询功能
   - ORM集成和迁移工具
   - 仓储模式实现

5. **di/** - 依赖注入
   - 封装依赖注入容器，简化服务管理
   - 提供简洁的API进行服务注册和获取

6. **event/** - 事件系统
   - 事件分发和订阅机制
   - 支持同步和异步事件处理
   - 事件监听器管理

7. **middleware/** - 中间件
   - 包含常用的HTTP中间件实现
   - 提供统一的中间件接口定义

8. **utils/** - 工具函数
   - 各种辅助功能和通用工具
   - 如验证器、加密工具等

9. **examples/** - 示例应用
   - 展示框架的使用方法和最佳实践
   - 提供不同场景的参考实现

10. **auth/** - 认证与授权 (计划中)
    - 统一的认证接口和多驱动支持
    - 基于角色和策略的授权系统
    - 社交登录集成

11. **security/** - 安全功能 (计划中)
    - 安全服务提供者
    - 现有中间件集成与增强
    - 内容安全策略
    - 安全头部配置
    - 安全审计日志

12. **queue/** - 任务队列 (计划中)
    - 异步任务处理
    - 多驱动支持
    - 失败任务重试

13. **storage/** - 文件存储 (计划中)
    - 统一的文件存储接口
    - 云存储集成
    - 文件上传管理

14. **flow.go** - 框架入口
    - 定义核心结构和接口
    - 封装Gin引擎，提供扩展功能

15. **error.go** - 错误处理
    - 统一的错误类型和处理机制
    - 友好的错误响应格式

16. **test/** - 测试工具 (计划中)
    - HTTP测试客户端
    - 模拟数据生成
    - 单元与集成测试助手
    - 数据库测试支持

17. **cloud/** - 云原生支持 (计划中)
    - 容器化配置
    - 健康检查
    - 分布式追踪
    - 云服务提供商适配

这种模块化的结构设计使框架既保持了清晰的组织，又提供了良好的扩展性。每个组件都有明确的职责，遵循单一职责原则，便于测试和维护。

## 贡献

我们欢迎所有形式的贡献，包括但不限于：

- 提交问题和功能请求
- 提交代码改进
- 改进文档
- 分享使用体验

请参阅[贡献指南](./CONTRIBUTING.md)了解更多详情。

## 许可证

Flow使用[MIT许可证](./LICENSE)。

# Flow 框架修复日志

## 已完成修复

### 1. 修复代码重复声明问题

- **validation/domain.go**:
  - 删除重复的`validate`变量声明
  - 将`Initialize`函数重命名为`InitializeDomainValidation`
  - 将`CustomValidationError`类型重命名为`ValidationErrorWithCustomMessage`
  - 确保所有引用这些更改的代码都被正确更新

### 2. 修复未使用变量问题

- **test/helper.go**:
  - 改进了`SetupTestDB`函数的注释，避免了未使用变量的lint错误
  - 移除了未使用的`fmt`导入

### 3. 修复依赖项标记问题

- **go.mod**:
  - 将`github.com/go-playground/locales`和`github.com/go-playground/universal-translator`标记为直接依赖
  - 将`github.com/gin-contrib/sse`正确标记为间接依赖

## 下一步工作

- [ ] 运行完整的测试套件验证修复
- [ ] 检查其他可能的lint错误
- [ ] 考虑更新Go版本（当前为1.20）

# Flow 框架社交登录集成

Flow框架社交登录集成模块为您的应用提供了简单且强大的社交媒体登录功能。支持主流社交平台，如GitHub、Google和微信，让您的用户能够使用他们喜欢的社交账号快速登录。

## 特性

- **多平台支持**：内置支持GitHub、Google和微信登录，可轻松扩展更多平台
- **简单易用的API**：简洁的接口设计，只需几行代码即可集成
- **灵活的用户创建**：自定义用户创建逻辑，轻松适配已有的用户系统
- **无缝集成**：与Flow框架其他组件完美配合
- **类型安全**：完整的类型定义，提供良好的开发体验
- **全面的错误处理**：详细的错误信息，便于调试和处理异常情况

## 安装

通过Go模块安装：

```bash
go get github.com/zzliekkas/flow
```

## 快速开始

### 1. 配置社交登录

```go
package main

import (
    "github.com/zzliekkas/flow/auth/drivers"
)

func main() {
    // 创建用户存储库
    userRepo := &MyUserRepository{}
    
    // 创建社交登录管理器
    socialManager := drivers.NewSocialManager(userRepo)
    
    // 注册GitHub登录提供商
    githubProvider := drivers.NewGitHubProvider(drivers.SocialProviderConfig{
        Provider:     drivers.ProviderGitHub,
        ClientID:     "your-github-client-id",
        ClientSecret: "your-github-client-secret",
        RedirectURL:  "https://your-app.com/auth/github/callback",
        Scopes:       []string{"user:email"},
    })
    
    socialManager.RegisterProvider(githubProvider)
    
    // 设置自定义用户创建回调
    socialManager.SetCreateUserCallback(func(ctx context.Context, socialUser *drivers.SocialUser) (interface{}, error) {
        // 自定义用户创建逻辑
        return createOrFindUser(socialUser), nil
    })
    
    // 继续应用启动流程...
}
```

### 2. 注册路由

使用您的Web框架注册登录和回调路由：

```go
// 使用Gin框架示例
router := gin.Default()

// GitHub登录路由
router.GET("/auth/github", func(c *gin.Context) {
    socialManager.HandleLogin(drivers.ProviderGitHub)(c.Writer, c.Request)
})

// GitHub回调路由
router.GET("/auth/github/callback", func(c *gin.Context) {
    socialManager.HandleCallback(drivers.ProviderGitHub)(c.Writer, c.Request)
})
```

### 3. 创建登录页面

在您的前端页面中添加社交登录按钮：

```html
<a href="/auth/github" class="github-login">
    <img src="/assets/github-logo.png" alt="GitHub登录">
    使用GitHub账号登录
</a>

<a href="/auth/google" class="google-login">
    <img src="/assets/google-logo.png" alt="Google登录">
    使用Google账号登录
</a>

<a href="/auth/wechat" class="wechat-login">
    <img src="/assets/wechat-logo.png" alt="微信登录">
    使用微信账号登录
</a>
```

## 支持的社交平台

### GitHub登录

1. 在[GitHub开发者设置](https://github.com/settings/developers)创建一个OAuth应用
2. 设置应用的回调URL为`https://your-app.com/auth/github/callback`
3. 获取Client ID和Client Secret
4. 配置GitHub提供商：

```go
githubProvider := drivers.NewGitHubProvider(drivers.SocialProviderConfig{
    Provider:     drivers.ProviderGitHub,
    ClientID:     "your-github-client-id",
    ClientSecret: "your-github-client-secret",
    RedirectURL:  "https://your-app.com/auth/github/callback",
    Scopes:       []string{"user:email"},
})
```

### Google登录

1. 在[Google Cloud Console](https://console.cloud.google.com/)创建一个项目
2. 配置OAuth同意屏幕
3. 创建OAuth客户端ID，选择Web应用类型
4. 设置回调URL为`https://your-app.com/auth/google/callback`
5. 配置Google提供商：

```go
googleProvider := drivers.NewGoogleProvider(drivers.SocialProviderConfig{
    Provider:     drivers.ProviderGoogle,
    ClientID:     "your-google-client-id",
    ClientSecret: "your-google-client-secret",
    RedirectURL:  "https://your-app.com/auth/google/callback",
    Scopes:       []string{"profile", "email"},
})
```

### 微信登录

1. 在[微信开放平台](https://open.weixin.qq.com/)注册并创建一个网站应用
2. 获取AppID和AppSecret
3. 设置回调域名
4. 配置微信提供商：

```go
wechatProvider := drivers.NewWeChatProvider(drivers.SocialProviderConfig{
    Provider:     drivers.ProviderWeChat,
    ClientID:     "your-wechat-appid",
    ClientSecret: "your-wechat-secret",
    RedirectURL:  "https://your-app.com/auth/wechat/callback",
    Scopes:       []string{"snsapi_login"},
})
```

## 实现自定义社交平台

如果需要支持其他社交平台，只需实现`SocialProvider`接口：

```go
type MySocialProvider struct {
    config drivers.SocialProviderConfig
}

// 实现SocialProvider接口的所有方法
func (p *MySocialProvider) GetProvider() string {
    return "my-provider"
}

func (p *MySocialProvider) GetAuthURL(state string) string {
    // 实现授权URL生成逻辑
}

func (p *MySocialProvider) ExchangeToken(ctx context.Context, code string) (string, error) {
    // 实现令牌交换逻辑
}

func (p *MySocialProvider) GetUserInfo(ctx context.Context, token string) (*drivers.SocialUser, error) {
    // 实现用户信息获取逻辑
}

// 注册自定义提供商
socialManager.RegisterProvider(&MySocialProvider{
    config: drivers.SocialProviderConfig{
        Provider:     "my-provider",
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
        RedirectURL:  "https://your-app.com/auth/my-provider/callback",
        Scopes:       []string{"required-scopes"},
    },
})
```

## 贡献

欢迎通过Pull Request或Issues贡献代码或提供建议。

## 许可证

本项目采用MIT许可证。请查阅LICENSE文件了解详情。

## 队列系统 (Queue System)

Flow框架提供了功能强大且灵活的任务队列系统，支持异步任务处理、延迟执行、重试机制以及多种队列驱动。

### 核心功能

- **多驱动支持**：内置内存队列和Redis队列驱动，可轻松扩展支持其他后端存储
- **延迟任务**：支持在指定时间或延迟一段时间后执行任务
- **任务调度**：可灵活安排任务的执行时间和顺序
- **重试机制**：任务失败时自动重试，支持配置最大重试次数和重试间隔
- **并发控制**：可配置工作进程数量，控制任务的并发处理能力
- **中间件支持**：通过中间件机制扩展任务处理流程，如日志记录、性能监控等
- **与事件系统集成**：可将事件直接转换为队列任务，实现松耦合的系统架构

### 使用示例

#### 基本用法

```go
// 创建内存队列
memQueue := memory.New(3) // 设置最大重试次数为3

// 创建队列管理器
manager := queue.NewQueueManager()

// 添加队列
manager.AddQueue("default", memQueue)

// 注册任务处理器
manager.Register("send_email", func(ctx context.Context, job *queue.Job) error {
    to, _ := job.Payload["to"].(string)
    subject, _ := job.Payload["subject"].(string)
    log.Printf("发送邮件: 收件人=%v, 主题=%v", to, subject)
    // 执行邮件发送逻辑...
    return nil
})

// 推送任务
payload := map[string]interface{}{
    "to":      "user@example.com",
    "subject": "队列示例",
}
jobID, err := manager.Push(context.Background(), "send_email", payload)

// 启动工作进程
queue, _ := manager.GetQueue("default")
queue.StartWorker(context.Background(), "default", 3) // 启动3个工作进程
```

#### 使用中间件

```go
// 使用中间件注册任务处理器
manager.RegisterWithMiddleware("process_payment",
    paymentHandler,
    queue.LoggingMiddleware(),  // 添加日志中间件
    queue.RetryMiddleware(5),   // 添加重试中间件，最多重试5次
)
```

#### 使用Redis队列（分布式处理）

```go
// 创建Redis队列
options := redis.DefaultOptions()
options.Addr = "localhost:6379"
redisQueue, err := redis.New(options)

// 添加到管理器
manager.AddQueue("distributed", redisQueue)

// 在不同的进程或服务器上启动工作进程
redisQueue.StartWorker(context.Background(), "distributed", 5)
```

#### 与事件系统集成

```go
// 创建队列事件监听器
listener := queue.NewQueueEventListener(manager, "default")

// 注册事件到任务的映射
listener.RegisterEventJob("user.registered", "send_welcome_email")

// 将监听器添加到事件分发器
dispatcher.AddListener("user.registered", listener)
```

### 队列驱动

#### 内存队列 (Memory Queue)

适用于单进程应用，所有任务数据保存在内存中，不支持跨进程或跨服务器的任务处理。优点是速度快，无需外部依赖。

#### Redis队列 (Redis Queue)

基于Redis实现的分布式队列，支持跨进程和跨服务器的任务处理。使用Redis的列表、有序集合和哈希表实现队列、调度和任务数据存储，提供可靠的分布式任务处理能力。

### 队列状态管理

任务可以处于以下状态：
- **pending**: 等待执行
- **scheduled**: 已计划，等待到达指定时间
- **running**: 执行中
- **completed**: 已完成
- **failed**: 执行失败
- **retrying**: 等待重试
- **cancelled**: 已取消

查看任务状态和管理队列：

```go
// 获取任务信息
job, err := queue.Get(ctx, "default", jobID)
fmt.Printf("任务状态: %s\n", job.Status)

// 获取队列大小
size, err := queue.Size(ctx, "default")
fmt.Printf("队列大小: %d\n", size)

// 清空队列
queue.Clear(ctx, "default")
```

## 开发状态

以下是各个模块的开发状态：

| 模块 | 状态 | 说明 |
|------|------|------|
| 基础框架 | ✅ 已完成 | 包含路由、中间件、容器等核心功能 |
| 配置系统 | ✅ 已完成 | 支持多种配置源，环境变量等 |
| 日志系统 | ✅ 已完成 | 支持多种日志驱动和日志级别 |
| 认证系统 | ✅ 已完成 | 包含多种认证驱动和用户权限管理 |
| 数据库 | ✅ 已完成 | 支持MySQL、PostgreSQL、SQLite等 |
| 事件系统 | ✅ 已完成 | 支持事件分发和监听 |
| 任务队列 | ✅ 已完成 | 支持内存队列和Redis队列，延迟任务和重试机制 |
| 安全系统 | ⚠️ 进行中 | 包含CSRF防护、XSS过滤等安全特性 |
| 文件系统 | ⚠️ 进行中 | 支持本地和云存储 |
| 缓存系统 | ⚠️ 进行中 | 支持多种缓存驱动 |
| 邮件系统 | ⚠️ 进行中 | 支持多种邮件驱动 |
| 国际化 | 🔄 计划中 | 支持多语言翻译 |
