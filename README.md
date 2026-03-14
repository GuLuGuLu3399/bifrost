# BIFROST

Bifrost 是一个面向内容管理与分发场景的多语言 Monorepo 项目，核心目标是将：

- Go 的业务编排能力
- Rust 的高性能处理能力
- 前端多端交互能力（Web + Tauri）

整合为一套可演进的内容平台。

当前架构采用 CQRS 思路进行职责拆分：

- Go：Gjallar（网关）、Nexus（写链路）、Beacon（读链路）
- Rust：Forge（渲染）、Mirror（搜索）、Oracle（异步分析）
- Frontend：Horizon（Nuxt 内容站）、Helm（Tauri 管理端）

---

## 项目特点

### 1. 多语言协作但协议统一

- 以 Protobuf 作为服务契约，Go/Rust 共用 API 定义。
- 对外统一由 Gjallar 暴露 HTTP `/v1/*`，内部通过 gRPC 协作。

### 2. 读写链路分离

- 写链路由 Nexus 负责，支持文章创建、更新、删除和管理接口。
- 读链路由 Beacon 负责，提供详情、列表、评论、分类标签等查询能力。
- 搜索能力独立到 Mirror，降低主业务链路耦合。

### 3. 事件驱动扩展

- 通过 NATS 进行内容事件传播（`content.post.*`）。
- Mirror/Oracle 等服务可独立消费事件进行索引与异步任务处理。

### 4. 前后端并行演进

- Horizon 面向内容浏览与 SEO。
- Helm 面向后台管理与运维联调（含 Tauri 命令桥接和接口实验能力）。

### 5. 运维入口统一

- 根目录 `manage.ps1` 提供 Docker 构建、启动、日志与状态查看。
- 便于本地一键拉起完整依赖并进行链路联调。

---

## 仓库结构概览

```txt
bifrost/
├── api/                 # Protobuf 契约定义
├── go_services/         # Go 服务（网关/读写）
├── rust_services/       # Rust 服务（渲染/搜索/异步）
├── frontend/            # 前端工作区（horizon + helm）
├── migrations/          # PostgreSQL 迁移脚本
├── docs/                # 架构、接口、事件契约文档
├── docker-compose.yml   # 本地编排
└── manage.ps1           # 运维脚本
```

---

## 当前状态说明

项目处于持续迭代阶段，整体功能链路已打通，但仍有部分模块在完善中：

- 网关、读写服务、搜索服务、前端页面已具备基础可用性。
- 管理端功能在持续补齐，部分接口联调仍在推进。
- 文档已按当前代码进行同步，但后续版本可能继续快速变化。

---

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

| 服务         | 端口 | 协议 |
| ------------ | ---- | ---- |
| gjallar      | 8080 | HTTP |
| nexus        | 9001 | gRPC |
| beacon       | 9002 | gRPC |
| forge        | 9092 | gRPC |
| mirror       | 9093 | gRPC |
| oracle       | 9094 | gRPC |
| horizon(dev) | 3001 | HTTP |
| helm(dev)    | 3000 | HTTP |

---

## 常用命令

```powershell
.\manage.ps1 docker-up
.\manage.ps1 docker-down
.\manage.ps1 docker-restart
.\manage.ps1 docker-logs
.\manage.ps1 docker-logs gjallar
```

常见调试方式：

- 查看全量日志：`./manage.ps1 docker-logs`
- 查看单服务日志：`./manage.ps1 docker-logs <service>`
- 检查服务状态：`./manage.ps1 docker-ps`

---

## 文档索引

- [ARCHITECTURE](./docs/ARCHITECTURE.md)
- [EVENT_CONTRACT](./docs/EVENT_CONTRACT.md)
- [FRONTEND_API](./docs/FRONTEND_API.md)
- [GO_SERVICES](./go_services/README.md)
- [RUST_SERVICES](./rust_services/README.md)
- [MIGRATIONS](./migrations/README.md)

---

## 已知风险与注意事项

- 部分接口和网关注册状态可能与预期不完全一致，联调时请优先以实际返回为准。
- 一些功能分支刚完成重构，回归测试覆盖仍不完整。
- 本仓库包含多个子模块并行开发，某些 README/注释可能滞后于实现细节。

---

## 作者说明

作者精力有限，本项目还有许多 bug 未完成测试和修改。
bug应该很多，前面做系统构思的时候想的太多了，想写的东西太多了，后续写下来太累了，跟马拉松一样，有点写不动了，项目对我来说内容很多，需要测试的东西也很多，本来就是想学一下微服务架构的开发还有，monorepo的开发，就这样吧，以后看我心情更新仓库。
如果你也对这个项目感兴趣，欢迎一起参与进来，提交 PR 或者 issue。
