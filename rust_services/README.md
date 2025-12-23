# rust_services

Bifrost 的 Rust 微服务工作区（Cargo Workspace）。当前包含：

- `common`：共享基础库（proto 重导出 / Context 透传 / 错误模型 / tracing 初始化 / OTEL）
- `forge`：Markdown 渲染服务（gRPC）
- `mirror`：全文检索服务（gRPC + Tantivy，含后台索引 worker）
- `oracle`：埋点分析 / BI 服务（gRPC + DuckDB，含高并发批量写入）

> 目录：`D:/dev/bifrost/rust_services`

---

## 1) 代码结构

```
rust_services/
  Cargo.toml        # workspace 入口
  common/
  forge/
  mirror/
  oracle/
  Dockerfile        # 多服务通用镜像构建
```

各服务都是一个 `bin` crate，通过 `tonic` 提供 gRPC 服务。

---

## 2) 前置依赖

- Rust toolchain（建议 stable）
- Windows：生成 proto 可能需要 `protoc`（common/build.rs 会编译 proto 并生成 descriptor）
- Docker（可选，用于容器化）

---

## 3) 构建

在 `rust_services/` 根目录：

```powershell
cargo build
```

仅构建单个服务：

```powershell
cargo build -p forge
cargo build -p mirror
cargo build -p oracle
```

---

## 4) 本地运行（不使用 Docker）

> 三个服务都使用 `common::config::ConfigLoader` 读取配置。约定使用环境变量前缀：
> - `forge`：`APP_FORGE__*`
> - `mirror`：`APP_MIRROR__*`
> - `oracle`：`APP_ORACLE__*`
>
> 日志级别用 `RUST_LOG` 控制。

### 4.1 forge

必需：
- `APP_FORGE__SERVER__ADDR`（例如 `127.0.0.1:50051`）

可选：
- `APP_FORGE__LOG__FORMAT=console|json`

运行：

```powershell
$env:APP_FORGE__SERVER__ADDR="127.0.0.1:50051"
$env:APP_FORGE__LOG__FORMAT="console"
$env:RUST_LOG="info"

cargo run -p forge
```

### 4.2 mirror

必需：
- `APP_MIRROR__SERVER__ADDR`
- `APP_MIRROR__NATS__URL`

可选：
- `MIRROR_INDEX_PATH`（默认 `./data/tantivy_index`）
- `APP_MIRROR__LOG__FORMAT=console|json`

运行：

```powershell
$env:APP_MIRROR__SERVER__ADDR="127.0.0.1:50052"
$env:APP_MIRROR__NATS__URL="nats://127.0.0.1:4222"
$env:MIRROR_INDEX_PATH="./data/tantivy_index"
$env:APP_MIRROR__LOG__FORMAT="console"
$env:RUST_LOG="info"

cargo run -p mirror
```

### 4.3 oracle

必需：
- `APP_ORACLE__SERVER__ADDR`

注意：
- Oracle 当前使用 DuckDB 文件：`data/analytics.db`（代码里固定路径，会自动创建 `data/` 目录）。

运行：

```powershell
$env:APP_ORACLE__SERVER__ADDR="127.0.0.1:50053"
$env:RUST_LOG="info"

cargo run -p oracle
```

---

## 5) 日志 / tracing

三个服务启动时都会调用：

- `common::logger::init_tracing(service_name, env, json)`

其中 `json` 由各自配置的 `APP_*__LOG__FORMAT` 决定：
- `json` => JSON 输出
- 其它 / 未设置 => console 输出

常用：

```powershell
$env:RUST_LOG="info"
# 或者
$env:RUST_LOG="debug"
# 或者仅某些模块
$env:RUST_LOG="common=debug,mirror=debug"
```

---

## 6) Docker（单服务镜像）

仓库根部的 `Dockerfile` 是一个通用多阶段构建文件，通过 `SERVICE` 参数选择构建哪一个二进制。

构建 forge：

```powershell
docker build -t bifrost-forge --build-arg SERVICE=forge .
```

构建 mirror：

```powershell
docker build -t bifrost-mirror --build-arg SERVICE=mirror .
```

构建 oracle：

```powershell
docker build -t bifrost-oracle --build-arg SERVICE=oracle .
```

运行示例（forge）：

```powershell
docker run --rm -p 50051:50051 \
  -e APP_FORGE__SERVER__ADDR=0.0.0.0:50051 \
  -e APP_FORGE__LOG__FORMAT=console \
  -e RUST_LOG=info \
  bifrost-forge
```

---

## 7) 共享库 common

详见：`common/README.md`。

---

## 8) 常见问题

### 8.1 build.rs 报 protoc / descriptor 相关错误

`common/build.rs` 会在编译期生成 proto 代码及 descriptor：
- Windows 上请确保 `protoc` 可用且在 `PATH` 中。

### 8.2 mirror 启动时报 NATS config missing

`mirror/src/main.rs` 会强制要求 `config.nats` 存在：
- 设置 `APP_MIRROR__NATS__URL`（例如 `nats://127.0.0.1:4222`）。

### 8.3 oracle DuckDB 链接问题（Windows）

`oracle` 依赖 `duckdb`，已启用 `bundled` feature，正常情况下会自动编译并链接。

---

## 9) 质量门禁（推荐）

```powershell
cargo fmt
cargo build
```

