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

## 可选模块（Optional — 计划在 v2.1 拆为独立模块）
以下包是可选扩展，引入了较重的第三方依赖。将在 v2.1 版本中拆分为独立 Go 模块：

| 包 | 说明 | 重依赖 | 未来模块路径 |
|---|------|--------|-------------|
| `cloud/` | 云服务集成（AWS, OpenTelemetry） | aws-sdk-go-v2, opentelemetry | `github.com/zzliekkas/flow-cloud` |
| `payment/` | 支付集成（Stripe, PayPal, Alipay, WeChat） | stripe-go, paypal, alipay, wechatpay | `github.com/zzliekkas/flow-payment` |
| `storage/` | 对象存储（OSS, S3, Qiniu, COS） | aliyun-oss, aws-s3, qiniu, cos, imaging | `github.com/zzliekkas/flow-storage` |
| `websocket/` | WebSocket 管理（已抽离，可选迁移） | gorilla/websocket | `github.com/zzliekkas/flow-websocket`（v0.1.0+） |

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
