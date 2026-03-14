# 🌈 BIFROST

> 🚀 **构建高性能、可扩展的现代内容聚合平台**
>
> 🟢 **Go** (业务编排) ✖️ 🦀 **Rust** (高性能计算) ✖️ 🎨 **Vue/Nuxt** (多端呈现)

## 📑 目录

- [✨ 项目简介](#-项目简介)
- [🎯 核心特点](#-核心特点)
- [🏗️ 架构与模块](#-架构与模块)
- [🚀 快速启动](#-快速启动)
- [🔌 服务端口](#-服务端口)
- [🛠️ 常用命令](#-常用命令)
- [📚 文档索引](#-文档索引)
- [🚧 当前状态](#-当前状态)
- [⚠️ 维护说明](#-维护说明)

## ✨ 项目简介

**Bifrost** 是一个以内容管理与分发为目标的工程化实验项目，核心探索微服务架构与多语言协同下的开发范式：

- 🟢 **Go 侧**：负责 API 网关管理与核心业务逻辑编排。
- 🦀 **Rust 侧**：负责 Markdown 渲染、全文搜索、异步任务等 CPU 密集及高性能需求处理。
- 🎨 **前端侧**：包含面向普通用户的内容站（Horizon / Nuxt）与面向管理员的管理端（Helm / Tauri）。

## 🎯 核心特点

- 🤝 **协议统一**：以 Protobuf 作为跨语言契约，拉平不同技术栈间的通讯壁垒。
- ⚖️ **读写分离**：严格遵循 **CQRS** 分层架构，Nexus 专职负责写（Command），Beacon 专职负责读（Query）。
- 📨 **事件驱动**：借助 NATS 实现高度解耦的内容事件异步传播分发。
- 📱 **多端协作**：现代 Web 服务页面与 Tauri 原生应用管理端并行演进。
- 🐳 **运维统一**：提供化繁为简的 manage.ps1 脚本，一键接管本地容器编排与日志输出。

## 🏗️ 架构与模块

| 层级 | 模块 | 核心职责与说明 |
| :--- | :--- | :--- |
| 🌐 **Gateway** | Gjallar | HTTP 统一入口、JWT 鉴权拦截、跨域处理与路由聚合 |
| ✍️ **Command** | Nexus | 核心写链路、内容变更防腐、数据持久化与事件发布 |
| 📖 **Query** | Beacon | 核心读链路、列表与详情高频查询、缓存击穿与回源策略 |
| ⚡ **Render** | Forge | _(Rust)_ Markdown AST 解析与高速 HTML / TOC 渲染 |
| 🔍 **Search** | Mirror | _(Rust)_ 基于 Tantivy 的全文检索索引、检索与分词建议 |
| ⏱️ **Async** | Oracle | _(Rust)_ 消息总线异步分析、定期数据报表统计处理 |
| 🌍 **Frontend** | Horizon | 基于 Nuxt 构建的重首屏 SSR 内容呈现页面 |
| 🖥️ **Frontend** | Helm | 基于 Tauri + Vue 构建的跨平台重交互桌面管理端 |

## 🚀 快速启动

确保你已成功安装 Docker 和 PowerShell：

\\\powershell
# 1) 🔍 环境检查
.\manage.ps1 docker-validate

# 2) 📦 构建所有服务镜像
.\manage.ps1 docker-build

# 3) 🚀 启动完整微服务集群
.\manage.ps1 docker-up

# 4) 📊 查看容器运行心跳与状态
.\manage.ps1 docker-ps
\\\

## 🔌 服务端口

| 服务名 | 端口 | 协议 | 连通说明 |
| :--- | :--- | :--- | :--- |
| 🌐 gjallar | 8080 | HTTP | 前端及外部应用的**统一网关入站口** |
| ✍️ 
exus | 9001 | gRPC | 内部服务互联 |
| 📖 eacon | 9002 | gRPC | 内部服务互联 |
| ⚡ orge | 9092 | gRPC | 内部服务互联 |
| 🔍 mirror | 9093 | gRPC | 内部服务互联 |
| ⏱️ oracle | 9094 | gRPC | 内部服务互联 |
| 🌍 horizon(dev)| 3001 | HTTP | Nuxt SSR 内容站开发模式监听端口 |
| 🖥️ helm(dev) | 3000 | HTTP | Tauri Web 管理端开发模式监听端口 |

## 🛠️ 常用命令

所有本地运维操作收口于根目录 manage.ps1，提升开发心流：

\\\powershell
.\manage.ps1 docker-up            # 启动所有服务
.\manage.ps1 docker-down          # 销毁所有服务与网络体系
.\manage.ps1 docker-restart       # 热重启所有服务状态
.\manage.ps1 docker-logs          # 观测全局容器混合日志流
.\manage.ps1 docker-logs gjallar  # 特写追踪单一服务日志
\\\

## 📚 文档索引

如需深入了解各个子模块与领域的设计意图，请参阅：

- 📐 [架构细节 (ARCHITECTURE)](./docs/ARCHITECTURE.md)
- 🤝 [事件契约 (EVENT_CONTRACT)](./docs/EVENT_CONTRACT.md)
- 🔌 [前端接口 (FRONTEND_API)](./docs/FRONTEND_API.md)
- 🟢 [Go 服务指南 (GO_SERVICES)](./go_services/README.md)
- 🦀 [Rust 服务指南 (RUST_SERVICES)](./rust_services/README.md)
- 🗄️ [数据库变迁记录 (MIGRATIONS)](./migrations/README.md)

## 🚧 当前状态

项目正在火热的持续迭代期，主干基础链路跑通，但各模块边界依旧在补齐中：

- 🟢 网关、读写服务、搜索服务已打通联调，可单机全量部署。
- 🚧 桌面应用（Tauri）业务侧：管理功能的页面重构与能力移植持续进行中。
- 📝 各模块的详细文档会跟随代码 main 分支的推行持续演进。

## ⚠️ 维护说明

🙏 **作者精力受限，全系功能当前还有许多由于未经历系统化联调测试而遗留的 Bug**。

这本是一个从理念构思到架构验证的超大拉练，沿途涉及的技术广度较大，一个人做起来就如同马拉松长跑。
若您对当前的 Go + Rust 多语言微服务协作以及 Monorepo 的架构开发同样感兴趣，**非常欢迎提交 Issue 和 PR！** 🤝
