# RUST COMMON

> Rust 服务共享基础库。

## 目录

- [提供能力](#提供能力)
- [典型用法](#典型用法)
- [实践建议](#实践建议)

## 提供能力

- Proto 类型聚合与重导出
- gRPC metadata Context 透传
- 统一错误模型（业务码到 gRPC/HTTP 映射）
- tracing 初始化（console/json）
- OpenTelemetry OTLP 初始化

## 典型用法

```toml
[dependencies]
common = { workspace = true }
```

```rust
use common::{ContextData, CodeError, ErrorCode};
```

## 实践建议

- 服务启动统一走 common 日志/追踪初始化。
- 对外错误统一为 `CodeError`，避免泄露内部细节。
- 服务间调用尽量透传 `ContextData`。
