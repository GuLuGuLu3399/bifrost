# RUST SERVICES

> Bifrost Rust 工作区（渲染 / 搜索 / 异步处理）。

## 目录

- [服务列表](#服务列表)
- [构建](#构建)
- [运行](#运行)
- [Docker](#docker)

## 服务列表

- `forge`: Markdown 渲染服务
- `mirror`: 全文检索服务（Tantivy）
- `oracle`: 异步分析与处理服务
- `common`: 共享基础库

## 构建

```powershell
cd rust_services
cargo build

# 按服务构建
cargo build -p forge
cargo build -p mirror
cargo build -p oracle
```

## 运行

```powershell
# forge
$env:APP_FORGE__SERVER__ADDR="127.0.0.1:9092"
cargo run -p forge

# mirror
$env:APP_MIRROR__SERVER__ADDR="127.0.0.1:9093"
$env:APP_MIRROR__NATS__URL="nats://127.0.0.1:4222"
# 可选：关闭 NATS worker
# $env:APP_MIRROR__FEATURES__ENABLE_NATS_WORKER="false"
# 可选：覆盖默认拓扑
# $env:APP_MIRROR__NATS__STREAM_NAME="BIFROST_CONTENT"
# $env:APP_MIRROR__NATS__CONSUMER_NAME="mirror_indexer"
# $env:APP_MIRROR__NATS__FILTER_SUBJECT="content.post.>"
cargo run -p mirror

# oracle
$env:APP_ORACLE__SERVER__ADDR="127.0.0.1:9094"
cargo run -p oracle
```

## Docker

```powershell
docker build -f rust_services/Dockerfile -t bifrost-forge --build-arg SERVICE=forge .
docker build -f rust_services/Dockerfile -t bifrost-mirror --build-arg SERVICE=mirror .
docker build -f rust_services/Dockerfile -t bifrost-oracle --build-arg SERVICE=oracle .
```

## 相关文档

- [RUST COMMON](./common/README.md)
- [EVENT_CONTRACT](../docs/EVENT_CONTRACT.md)
- [ARCHITECTURE](../docs/ARCHITECTURE.md)
