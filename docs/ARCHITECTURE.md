# BIFROST ARCHITECTURE

> 最后更新：2026-03-14  
> 架构模式：CQRS + Go(业务编排) + Rust(高性能计算)

---

## 1. 仓库结构

```txt
bifrost/
├── api/                  # Protobuf 契约
├── go_services/          # Go 服务：gjallar / nexus / beacon
├── rust_services/        # Rust 服务：forge / mirror / oracle
├── frontend/             # 前端工作区：helm / horizon
├── migrations/           # PostgreSQL 迁移脚本
├── docker-compose.yml    # 本地编排
└── docs/                 # 文档
```

## 2. 服务拓扑

| 层级 | 服务 | 端口 | 说明 |
| --- | --- | --- | --- |
| 接入层 | Gjallar | 8080/HTTP | API 网关，鉴权、路由聚合、CORS |
| 写链路 | Nexus | 9001/gRPC | 内容写入、发布事件、调用 Forge |
| 读链路 | Beacon | 9002/gRPC | 内容读取、分页、缓存协作 |
| 渲染层 | Forge | 9092/gRPC | Markdown 渲染为 HTML + TOC |
| 搜索层 | Mirror | 9093/gRPC | 全文检索、建议词、分面 |
| 异步层 | Oracle | 9094/gRPC | 异步统计与分析任务 |
| 前端 | Horizon | 3001/dev | Nuxt 内容站 |
| 前端 | Helm | 3000/dev | Tauri 管理端 |

## 3. 关键链路

### 3.1 写链路（同步渲染 + 异步事件）

1. 客户端请求 Gjallar。
2. Gjallar 转发到 Nexus。
3. Nexus 在事务内同步调用 Forge 渲染 Markdown。
4. Nexus 持久化 `raw_markdown`、`html_body`、`toc_json` 到 PostgreSQL。
5. 事务提交后发布 NATS 事件（`content.post.*`）。
6. Mirror/Oracle 消费事件做索引与异步任务。

### 3.2 读链路（缓存优先）

1. 客户端经 Gjallar 请求 Beacon。
2. Beacon 读取缓存（Redis）命中则直接返回。
3. 未命中时查询 PostgreSQL 并回填缓存。
4. 返回 `PostSummary` / `PostDetail`。

### 3.3 搜索链路

1. 客户端调用 Gjallar 的 `/v1/search`。
2. Gjallar 同步调用 Mirror。
3. Mirror 基于 Tantivy 执行检索并返回 `hits + facets + took_ms`。

## 4. 当前实现要点

- Beacon `ListPosts` 查询已修复：JOIN 与 WHERE 顺序正确，LEFT JOIN 字段按可空类型扫描。
- Gjallar CORS 改为配置驱动：默认允许 `localhost:3000/3001/3002`。
- Horizon 登录页改用 `$fetch` 触发登录请求，避免在事件处理器中误用 `useFetch`。
- 网关层当前仍未注册 `StorageService`，`/v1/storage/upload_ticket` 通过 Gjallar 调用会 404。

## 5. 数据与事件

- 主库：PostgreSQL 16（内容、用户、关系数据）
- 缓存：Redis（读链路缓存）
- 消息：NATS（`content.post.created|updated|deleted`）
- 搜索：Tantivy（Mirror 本地索引）

## 6. 本地运行

```powershell
# 校验环境
.\manage.ps1 docker-validate

# 构建并启动
.\manage.ps1 docker-build
.\manage.ps1 docker-up

# 查看状态/日志
.\manage.ps1 docker-ps
.\manage.ps1 docker-logs
```

## 7. 相关文档

- `docs/EVENT_CONTRACT.md`
- `docs/FRONTEND_API.md`
- `go_services/README.md`
- `rust_services/README.md`
- `migrations/README.md`
