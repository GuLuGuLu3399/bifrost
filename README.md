# BIFROST

Bifrost 是一个 Go + Rust 的内容平台单体仓库（Monorepo），采用 CQRS 分层：

- Go：Gjallar（网关）、Nexus（写）、Beacon（读）
- Rust：Forge（渲染）、Mirror（搜索）、Oracle（异步分析）
- Frontend：Horizon（Nuxt 内容站）、Helm（Tauri 管理端）

## 快速启动

```powershell
# 环境检查
.\manage.ps1 docker-validate

# 构建与启动
.\manage.ps1 docker-build
.\manage.ps1 docker-up

# 查看状态
.\manage.ps1 docker-ps
```

## 服务端口

| 服务 | 端口 | 协议 |
| --- | --- | --- |
| gjallar | 8080 | HTTP |
| nexus | 9001 | gRPC |
| beacon | 9002 | gRPC |
| forge | 9092 | gRPC |
| mirror | 9093 | gRPC |
| oracle | 9094 | gRPC |
| horizon(dev) | 3001 | HTTP |
| helm(dev) | 3000 | HTTP |

## 常用命令

```powershell
.\manage.ps1 docker-up
.\manage.ps1 docker-down
.\manage.ps1 docker-restart
.\manage.ps1 docker-logs
.\manage.ps1 docker-logs gjallar
```

## 文档索引

- [ARCHITECTURE](./docs/ARCHITECTURE.md)
- [EVENT_CONTRACT](./docs/EVENT_CONTRACT.md)
- [FRONTEND_API](./docs/FRONTEND_API.md)
- [GO_SERVICES](./go_services/README.md)
- [RUST_SERVICES](./rust_services/README.md)
- [MIGRATIONS](./migrations/README.md)
