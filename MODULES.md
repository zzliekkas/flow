# Flow Framework — 模块分类

## 核心模块（Core）
以下包是 flow 框架的核心组成部分，始终随框架一起提供：

| 包 | 说明 | 主要依赖 |
|---|------|---------|
| `flow` (根包) | Engine, Context, Router, Lifecycle, Module, Logger | gin, dig |
| `config/` | 配置管理（Viper） | viper |
| `db/` | 数据库连接管理 | gorm, mysql/postgres/sqlite drivers |
| `di/` | 依赖注入容器封装 | dig |
| `middleware/` | 常用中间件（Logger, Recovery, CORS 等） | gin |
| `cache/` | 缓存管理 | go-redis |
| `auth/` | 认证（JWT, OAuth, Social Login） | golang-jwt |
| `validation/` | 请求验证 | go-playground/validator |
| `i18n/` | 国际化 | — |
| `error.go` | 错误处理 | — |

## 可选模块（已拆为独立 Go 模块，v0.1.0+）
以下包已提取为独立仓库和 Go 模块，可按需引入：

| 原包 | 独立模块 | 版本 | 重依赖 |
|------|---------|------|--------|
| `cloud/` | [`github.com/zzliekkas/flow-cloud`](https://github.com/zzliekkas/flow-cloud) | v0.1.0 | aws-sdk-go-v2, opentelemetry |
| `payment/` | [`github.com/zzliekkas/flow-payment`](https://github.com/zzliekkas/flow-payment) | v0.1.0 | stripe-go, paypal, alipay, wechatpay |
| `storage/` | [`github.com/zzliekkas/flow-storage`](https://github.com/zzliekkas/flow-storage) | v0.1.0 | aliyun-oss, aws-s3, qiniu, cos, imaging |
| `websocket/` | [`github.com/zzliekkas/flow-websocket`](https://github.com/zzliekkas/flow-websocket) | v0.1.0 | gorilla/websocket |

> 注意：为了向后兼容，`flow/v2` 中仍保留这些子包的源码，后续版本会标记 deprecated。

## 工具模块（Utility）
| 包 | 说明 |
|---|------|
| `app/` | 高级应用容器（ServiceProvider, Lifecycle, Hooks） |
| `cli/` | CLI 命令行工具 |
| `event/` | 事件系统 |
| `queue/` | 消息队列 |
| `security/` | 安全工具 |
| `profiler/` | 性能分析 |
| `utils/` | 通用工具函数 |
