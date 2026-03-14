# 🌈 BIFROST

> 🚀 **一个折腾 Go x Rust 微服务与 Monorepo 的全栈实验场**
> 🟢 **Go** (业务编排) ✖️ 🦀 **Rust** (高性能计算) ✖️ 🎨 **Vue/Nuxt** (多端呈现)

## 📑 目录

* [✨ 项目简介](https://www.google.com/search?q=%23-%E9%A1%B9%E7%9B%AE%E7%AE%80%E4%BB%8B)
* [🎯 核心特点](https://www.google.com/search?q=%23-%E6%A0%B8%E5%BF%83%E7%89%B9%E7%82%B9)
* [🏗️ 架构与模块](https://www.google.com/search?q=%23-%E6%9E%B6%E6%9E%84%E4%B8%8E%E6%A8%A1%E5%9D%97)
* [🚀 快速启动](https://www.google.com/search?q=%23-%E5%BF%AB%E9%80%9F%E5%90%AF%E5%8A%A8)
* [🔌 服务端口](https://www.google.com/search?q=%23-%E6%9C%8D%E5%8A%A1%E7%AB%AF%E5%8F%A3)
* [🛠️ 常用命令](https://www.google.com/search?q=%23-%E5%B8%B8%E7%94%A8%E5%91%BD%E4%BB%A4)
* [📚 文档索引](https://www.google.com/search?q=%23-%E6%96%87%E6%A1%A3%E7%B4%A2%E5%BC%95)
* [🚧 当前状态与碎碎念](https://www.google.com/search?q=%23-%E5%BD%93%E5%89%8D%E7%8A%B6%E6%80%81%E4%B8%8E%E7%A2%8E%E7%A2%8E%E5%BF%B5)

## ✨ 项目简介

**Bifrost** 是一个以内容管理与分发为目标的工程化实验项目，核心旨在探索**微服务架构**与**多语言协同**下的现代开发范式：

* 🟢 **Go 侧**：负责 API 网关管理与核心业务逻辑的调度编排。
* 🦀 **Rust 侧**：承担 Markdown 渲染、全文搜索、异步任务等 CPU 密集型及对性能要求极高的底层处理。
* 🎨 **前端侧**：包含面向普通用户的内容消费站（Horizon / Nuxt）与面向管理员的跨平台管理端（Helm / Tauri）。

## 🎯 核心特点

* 🤝 **协议统一**：以 Protobuf 作为跨语言契约，彻底拉平不同技术栈间的通讯壁垒。
* ⚖️ **读写分离**：严格遵循 **CQRS** 分层架构，Nexus 专职负责写（Command），Beacon 专职负责读（Query）。
* 📨 **事件驱动**：借助 NATS 实现高度解耦的内容事件异步传播与分发。
* 📱 **多端协作**：现代 Web SSR 服务与 Tauri 原生桌面应用并行演进。
* 🐳 **极简运维**：提供化繁为简的 `manage.ps1` 脚本，一键接管本地容器编排与日志观测。

## 🏗️ 架构与模块

| 层级 | 模块名称 | 核心职责与说明 |
| --- | --- | --- |
| 🌐 **Gateway** | **Gjallar** | HTTP 统一入口、JWT 鉴权拦截、跨域处理与路由聚合 |
| ✍️ **Command** | **Nexus** | 核心写链路、内容变更防腐、数据持久化与事件发布 |
| 📖 **Query** | **Beacon** | 核心读链路、列表与详情高频查询、缓存击穿与回源策略 |
| ⚡ **Render** | **Forge** | *(Rust)* Markdown AST 解析与高速 HTML / TOC 渲染 |
| 🔍 **Search** | **Mirror** | *(Rust)* 基于 Tantivy 的全文索引、内容检索与分词建议 |
| ⏱️ **Async** | **Oracle** | *(Rust)* 消息总线异步分析、定期数据报表统计与处理 |
| 🌍 **Frontend** | **Horizon** | 基于 Nuxt 3 构建的重首屏 SSR 内容呈现页面 |
| 🖥️ **Frontend** | **Helm** | 基于 Tauri + Vue 构建的跨平台、重交互桌面管理端 |

## 🚀 快速启动

确保您的本地环境已成功安装 Docker 和 PowerShell：

```powershell
# 1) 🔍 环境检查：验证 Docker 及相关依赖状态
.\manage.ps1 docker-validate

# 2) 📦 构建镜像：编译并构建所有微服务镜像
.\manage.ps1 docker-build

# 3) 🚀 启动集群：一键拉起完整微服务体系
.\manage.ps1 docker-up

# 4) 📊 状态观测：查看容器运行心跳与端口映射
.\manage.ps1 docker-ps

```

## 🔌 服务端口

| 服务名 | 端口 | 协议 | 连通说明 |
| --- | --- | --- | --- |
| 🌐 **gjallar** | `8080` | HTTP | 前端及外部应用的**统一网关入站口** |
| ✍️ **nexus** | `9001` | gRPC | 内部服务互联 |
| 📖 **beacon** | `9002` | gRPC | 内部服务互联 |
| ⚡ **forge** | `9092` | gRPC | 内部服务互联 |
| 🔍 **mirror** | `9093` | gRPC | 内部服务互联 |
| ⏱️ **oracle** | `9094` | gRPC | 内部服务互联 |
| 🌍 **horizon** | `3001` | HTTP | Nuxt SSR 内容站开发模式监听端口 |
| 🖥️ **helm** | `3000` | HTTP | Tauri Web 管理端开发模式监听端口 |

## 🛠️ 常用命令

所有本地运维操作均已收口于根目录的 `manage.ps1`，助你保持专注的开发心流：

```powershell
.\manage.ps1 docker-up            # 启动所有服务
.\manage.ps1 docker-down          # 销毁所有服务与网络体系
.\manage.ps1 docker-restart       # 热重启所有服务状态
.\manage.ps1 docker-logs          # 观测全局容器混合日志流
.\manage.ps1 docker-logs gjallar  # 特写追踪单一服务日志 (例: gjallar)

```

## 📚 文档索引

如需深入了解各个子模块与领域的设计意图，请参阅：

* 📐 [架构细节 (ARCHITECTURE)](https://www.google.com/search?q=./docs/ARCHITECTURE.md)
* 🤝 [事件契约 (EVENT_CONTRACT)](https://www.google.com/search?q=./docs/EVENT_CONTRACT.md)
* 🔌 [前端接口 (FRONTEND_API)](https://www.google.com/search?q=./docs/FRONTEND_API.md)
* 🟢 [Go 服务指南 (GO_SERVICES)](https://www.google.com/search?q=./go_services/README.md)
* 🦀 [Rust 服务指南 (RUST_SERVICES)](https://www.google.com/search?q=./rust_services/README.md)
* 🗄️ [数据库变迁记录 (MIGRATIONS)](https://www.google.com/search?q=./migrations/README.md)

## 🚧 当前状态与碎碎念

🙏 **本项目本质上是一个个人的技术探索与学习「试验田」，目前遗留了较多 Bug，请谨慎参考，切勿直接用于生产环境。**

作为一个从传统 SpringBoot + Vue3 架构向现代微服务（Go + Rust）转型的练手工程，项目的技术栈铺得非常广，涵盖了 **Monorepo、微服务架构、Go、Rust、Vue、Nuxt、NATS、Redis、PostgreSQL** 等众多技术和中间件。但这种“大而全”也带来了一些现实的开发困境：

* **精力透支与 Bug 遗留**：单兵作战完成这样一个多语言的全栈微服务项目，平时还要兼顾繁杂的课业和像 OOAD 这样的硬核考试，个人的精力和业余时间实在被榨干了，导致许多功能至今仍遗留了不少已知或未知的 Bug。
* **架构繁琐与调试地狱**：为了追求微服务的完整体验，引入了过多的组件。现在回头看，整体架构有些**过度设计且过于繁琐**。日常排查问题需要频繁穿梭于 Jaeger（链路追踪）、Dozzle（容器日志）以及各个微服务节点之间，极其冗长的 Debug 链路让单人快速联调变得步履维艰，严重拖慢了开发节奏。
* **前端优化停滞**：由于绝大部分脑力都消耗在了后端架构的搭建、跨语言通信以及容器编排的折腾上，目前实在没有余力再去精雕细琢前端（Horizon / Helm）的 UI 展示效果和交互体验了。

**总而言之**，这是一场挑战自我技术广度的“超大拉练”。如果您同样在学习这套 Go + Rust + 微服务的技术栈，或者对排雷、重构、简化架构感兴趣，**非常欢迎提交 Issue 和 PR！** 哪怕只是帮忙修一修前端页面的样式，或者优化一下本地部署的体验，我都感激不尽！🤝
