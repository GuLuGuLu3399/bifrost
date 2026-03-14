# GO SERVICES

Bifrost 的 Go 服务实现目录，包含：

- `gjallar`：HTTP 网关（统一入口、鉴权、CORS、路由聚合）
- `nexus`：写服务（文章增删改、事务内同步渲染）
- `beacon`：读服务（列表/详情查询、缓存协作）

## 目录结构

```txt
go_services/
├── cmd/
│   ├── gjallar/
│   ├── nexus/
│   └── beacon/
├── configs/
│   ├── gjallar.yaml
│   ├── nexus.yaml
│   └── beacon.yaml
├── internal/
│   ├── gjallar/
│   ├── nexus/
│   ├── beacon/
│   └── pkg/
└── api/
```

## 本地运行

```powershell
cd go_services

go run ./cmd/nexus/main.go -config configs/nexus.yaml
go run ./cmd/beacon/main.go -f configs/beacon.yaml
go run ./cmd/gjallar/main.go -f configs/gjallar.yaml
```

## 当前关键说明

- Beacon `ListPosts` 已修复 SQL 组装与可空字段扫描问题。
- Gjallar CORS 来源改为配置项 `cors.allowed_origins`。
- 默认开发来源建议保留 `3000/3001/3002`。

## Feature 开关

- `features.enable_messenger`
- `features.enable_storage`
- `features.enable_search`

可按 MVP 场景关闭非核心能力以减少依赖（NATS/MinIO/搜索链路）。

## 测试

```powershell
cd go_services
go test ./...
```

## 相关文档

- [ARCHITECTURE](../docs/ARCHITECTURE.md)
- [EVENT_CONTRACT](../docs/EVENT_CONTRACT.md)
- [FRONTEND_API](../docs/FRONTEND_API.md)
