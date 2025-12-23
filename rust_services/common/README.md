# common

Bifrost 的 Rust 共享基础库，给各个服务（`forge` / `mirror` / `oracle` 等）提供统一的：

- **Proto 类型聚合与重导出**（`tonic::include_proto!` 生成的类型）
- **跨服务 Context 透传**（从/向 gRPC Metadata 解析与注入）
- **统一错误模型**（业务错误码 ↔ gRPC / HTTP 映射）
- **tracing 日志初始化**（支持 console / JSON）
- **OpenTelemetry OTLP** 初始化（tracing → OTLP exporter）

> 目标：让每个服务只关心业务逻辑，基础设施能力在 `common` 里统一。

---

## 1. 依赖引入

在你的服务 crate 的 `Cargo.toml`：

- workspace 用法（推荐）：

```toml
[dependencies]
common = { workspace = true }
```

然后在代码里：

```rust
use common::{ContextData, CodeError, ErrorCode};
```

---

## 2. Proto：如何使用 re-export 的类型

`common` 将不同 package 的 proto 聚合并重导出，避免上层写长路径。

在 `common/src/lib.rs` 中已做这些 re-export：

- `common::common` → `bifrost.common.v1`
- `common::models` → `bifrost.content.v1`
- `common::forge` → `bifrost.content.v1.forge`
- `common::oracle` → `bifrost.analysis.v1.oracle`
- `common::search` → `bifrost.search.v1`

### 示例

```rust
use common::models;

fn example(m: models::ContentModel) {
    // ...
}
```

如果你想用更完整的路径也可以：

```rust
use common::api::content::v1 as content_v1;
```

---

## 3. Context 透传（gRPC Metadata）

`common::ctx::ContextData` 用于服务间透传用户、请求链路等元信息。

### Header 常量

定义在 `common/src/ctx.rs`：

- `x-user-id`
- `x-request-id`
- `authorization`
- `x-locale`
- `x-is-admin`

### 从请求中读取 Context

服务端拦截器 / handler 内：

```rust
use common::ContextData;
use tonic::Request;

pub async fn handler(req: Request<()>) {
    let ctx = ContextData::from_metadata(req.metadata());
    // ctx.user_id / ctx.request_id / ctx.token / ctx.locale / ctx.is_admin
}
```

### 向下游请求注入 Context

调用下游 gRPC 前：

```rust
use common::ContextData;
use tonic::Request;

let ctx = ContextData {
    user_id: Some(123),
    request_id: Some("req-xxx".to_string()),
    token: Some("Bearer ...".to_string()),
    locale: Some("zh-CN".to_string()),
    is_admin: false,
};

let mut req = Request::new(());
ctx.inject_request(&mut req);
```

---

## 4. 统一错误模型（ErrorCode / CodeError / ErrorResponse）

### ErrorCode

`common::error::ErrorCode` 是与 Go 端对齐的业务错误码：

- `BadRequest(400)`, `Unauthorized(401)`, `Forbidden(403)`, `NotFound(404)`
- `Conflict(409)`, `Validation(422)`
- `Internal(500)`, `ServiceUnavailable(503)`, `Timeout(504)`

并提供映射：

- `to_grpc() -> tonic::Code`
- `to_http() -> http::StatusCode`

### CodeError

`common::CodeError` 是统一业务错误类型：

- `CodeError::new(code, msg)`
- 一组快捷构造：`bad_request/unauthorized/forbidden/not_found/conflict/validation/internal`
- `to_status()`：转换成 `tonic::Status`，用于 gRPC 返回
- `to_response()`：转换成 `ErrorResponse`（给 gateway/HTTP 使用）

#### gRPC handler 示例

```rust
use common::CodeError;
use tonic::{Response, Status};

pub async fn get_something() -> Result<Response<()>, Status> {
    // ...
    Err(CodeError::not_found("resource not found").to_status())
}
```

#### HTTP/Gateway 场景

```rust
use common::{CodeError, ErrorResponse};

let err = CodeError::validation("invalid input");
let body: ErrorResponse = err.to_response();
let http_status = body.status_code();
```

---

## 5. tracing / 日志初始化（logger）

入口在 `common::logger`：

- `init_tracing(service_name, env, json) -> anyhow::Result<()>`

行为：

- 读取 `RUST_LOG`（`tracing_subscriber::EnvFilter::try_from_default_env()`）
- 默认 fallback 到 `info`
- `json = true` 输出 JSON；`json = false` 输出 console
- 使用 RFC3339 时间戳
- 使用 `OnceLock` 保证同进程只初始化一次

### 示例

```rust
common::logger::init_tracing("forge", "dev", false)?;
```

#### 常用环境变量

- `RUST_LOG=info`
- `RUST_LOG=debug`
- `RUST_LOG=common=debug,forge=debug`

---

## 6. tracing + OpenTelemetry OTLP（otel）

入口在 `common::otel`：

- `init_tracing_with_otel(service_name, env, json, otlp_endpoint) -> anyhow::Result<OtelGuard>`

返回 `OtelGuard`：**必须持有到进程结束**，用于在 drop 时 `shutdown()` flush。

### 示例

```rust
let _guard = common::otel::init_tracing_with_otel(
    "forge",
    "dev",
    true,
    "http://localhost:4317",
)?;
```

---

## 7. build.rs：proto 生成与 pbjson

`common/build.rs` 在编译期做两件事：

1. 用 `tonic_prost_build` 编译 proto（生成 Rust 类型）
2. 生成并读取 `file_descriptor_set`，再交给 `pbjson-build` 用于 JSON/serde 支持

### 触发重建

build 脚本会对以下路径做 `cargo:rerun-if-changed`：

- `../../api`
- 以及列出的各个 proto 文件

### 常见问题

- **`unable to open file_descriptor_set_path`**：通常是 `OUT_DIR` 不存在/不可写或 `protoc` 缺失。
  - 脚本里已经 `create_dir_all(OUT_DIR)`
  - Windows 下请确保 `protoc` 可用并在 `PATH` 中

---

## 8. 模块一览

- `common::config`：`AppConfig` 及基础设施配置结构体（示例）
- `common::ctx`：`ContextData` 与 gRPC metadata 透传
- `common::error`：统一错误码/错误类型
- `common::logger`：tracing 初始化（console / json）
- `common::otel`：tracing + OpenTelemetry OTLP 初始化

---

## 9. 约定建议（团队用法）

- 每个服务启动时二选一：
  - 本地/简单场景：`logger::init_tracing(...)`
  - 需要链路追踪：`otel::init_tracing_with_otel(...)` 并持有 `OtelGuard`
- 所有 gRPC 调用都尽量透传 `ContextData`（请求链路、鉴权信息等）。
- handler 返回错误时优先使用 `CodeError`，避免散落的 `anyhow::Error` 直接对外暴露。

