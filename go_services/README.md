# GO SERVICES

> Bifrost Go 服务实现目录（Gateway + Command + Query）。

## 目录

- [模块说明](#模块说明)
- [目录结构](#目录结构)
- [本地运行](#本地运行)
- [Feature 开关](#feature-开关)
- [现状说明](#现状说明)
- [测试](#测试)

## 模块说明

- `gjallar`: HTTP 网关（鉴权、路由聚合、CORS）
- `nexus`: 写服务（增删改、事件发布）
- `beacon`: 读服务（详情、列表、评论查询）

## 目录结构

```txt
go_services/
├── cmd/
│   ├── gjallar/
│   ├── nexus/
│   └── beacon/
├── configs/
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

## Feature 开关

- `features.enable_messenger`
- `features.enable_storage`
- `features.enable_search`

## 现状说明

- Beacon `ListPosts` 查询与可空字段扫描问题已修复。
- Gjallar CORS 来源改为配置化（默认 3000/3001/3002）。
- Nexus 侧可按开关启用 `StorageService`。
- Gjallar 目前未注册 `StorageService` 的 gateway handler，`/v1/storage/upload_ticket` 经网关返回 `404`。

## 测试

```powershell
cd go_services
go test ./...
```

## 相关文档

- [ARCHITECTURE](../docs/ARCHITECTURE.md)
- [EVENT_CONTRACT](../docs/EVENT_CONTRACT.md)
- [FRONTEND_API](../docs/FRONTEND_API.md)
