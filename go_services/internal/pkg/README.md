# INTERNAL PKG

`internal/pkg` 提供 Go 服务共享基础能力，避免业务代码重复造轮子。

## 包分层

- `xerr`：统一错误码与错误包装
- `observability`：日志、追踪、指标相关
- `network`：HTTP/gRPC 基础设施与中间件
- `database`：数据库连接与基础封装
- `cache`：缓存客户端封装
- `security`：JWT、密码相关安全工具
- `id`：分布式 ID 能力
- `lifecycle`：组件生命周期管理
- `middleware`：通用中间件

## 使用建议

- 对外返回统一用 `xerr`，避免裸露底层错误。
- 调试环境可提升日志级别并输出细节；生产环境建议降噪。
- 复用 `network` 与 `observability` 初始化流程，保持服务行为一致。

## 约束

- 仅放“跨服务复用”的基础能力。
- 禁止在这里放具体业务逻辑。
