# BIFROST ARCHITECTURE

> 最后更新：2026-03-14

## 目录

- [架构总览](#架构总览)
- [服务拓扑](#服务拓扑)
- [关键链路](#关键链路)
- [实现要点](#实现要点)
- [运行方式](#运行方式)

## 架构总览

Bifrost 采用 CQRS 分层，将写入、读取、搜索和异步处理拆分为独立服务。

```txt
bifrost/
├── go_services/      # gjallar / nexus / beacon
├── rust_services/    # forge / mirror / oracle
├── frontend/         # horizon / helm
├── migrations/
└── docs/
```

## 服务拓扑

| 层级 | 服务 | 端口 | 说明 |
| --- | --- | --- | --- |
| 接入层 | Gjallar | 8080/HTTP | API 网关、鉴权、CORS |
| 写链路 | Nexus | 9001/gRPC | 内容写入、事件发布 |
| 读链路 | Beacon | 9002/gRPC | 查询与分页 |
| 渲染层 | Forge | 9092/gRPC | Markdown 渲染 |
| 搜索层 | Mirror | 9093/gRPC | 全文检索与建议 |
| 异步层 | Oracle | 9094/gRPC | 异步处理与统计 |
| 前端 | Horizon | 3001/dev | Nuxt 内容站 |
| 前端 | Helm | 3000/dev | Tauri 管理端 |

## 关键链路

### 写链路

1. Client -> Gjallar
2. Gjallar -> Nexus
3. Nexus -> Forge（同步渲染）
4. Nexus 写入 PostgreSQL
5. Nexus 发布 `content.post.*` 事件

### 读链路

1. Client -> Gjallar -> Beacon
2. Beacon 优先读取 Redis
3. 未命中回源 PostgreSQL 并回填缓存

### 搜索链路

1. Client -> Gjallar `/v1/search`
2. Gjallar -> Mirror
3. Mirror 返回 hits / facets / took_ms

## 实现要点

- Beacon `ListPosts` 已修复 JOIN/WHERE 组装与可空字段扫描问题。
- Gjallar CORS 来源为配置驱动（默认 3000/3001/3002）。
- Gjallar 会清洗空查询参数（空串/null/undefined）再进入 gRPC-Gateway。
- Gjallar 搜索链路支持降级，Mirror 不可用时返回 `503`。
- `StorageService` 在 Nexus 侧可开启，但当前 Gjallar 未注册对应 gateway handler，`/v1/storage/upload_ticket` 经网关返回 `404`。

## 运行方式

```powershell
.\manage.ps1 docker-validate
.\manage.ps1 docker-build
.\manage.ps1 docker-up
.\manage.ps1 docker-ps
```
