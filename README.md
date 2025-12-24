# 🌈 Bifrost：Go + Rust 混合微服务 CMS 平台

> CQRS + 双引擎驱动的现代内容管理平台
>
> ✨ **高性能渲染** | 🔍 **全文检索** | 📊 **实时分析** | 🎨 **Markdown 渲染** | 📈 **可观测性**

---

## 🚀 快速开始

```powershell
# 1. 验证环境
.\manage.ps1 docker-validate

# 2. 构建镜像
.\manage.ps1 docker-build

# 3. 启动服务
.\manage.ps1 docker-up

# 4. 查看状态
.\manage.ps1 docker-ps

# 5. 查看日志
.\manage.ps1 docker-logs -Service gjallar
```

访问：

- **API 网关**: <http://localhost:8080>
- **Jaeger 追踪**: <http://localhost:16686>
- **MinIO 控制台**: <http://localhost:9001>
- **日志查看**: <http://localhost:9999>

---

## 📋 核心文档

| 文档 | 说明 |
|------|------|
| [系统架构蓝图](./docs/系统架构蓝图.md) | CQRS 架构、服务拓扑、数据流 |
| [Docker 部署指南](./docs/Docker部署指南.md) | 完整部署流程、优化策略 |
| [API 接口规范](./docs/API接口交互规范.md) | HTTP/gRPC 接口文档 |
| [gRPC 集成报告](./docs/GRPC_INTEGRATION_REPORT.md) | Go→Rust 跨语言调用实现 |
| [测试指南](./docs/TESTING_GUIDE.md) | 单元测试、集成测试策略 |
| [数据库架构](./migrations/readme.md) | PostgreSQL 表设计 |
| [代码审计报告](./docs/代码审计报告.md) | 质量评估与改进建议 |

### 服务文档

- [Go Services README](./go_services/README.md) - Go 核心服务架构
- [Rust Services README](./rust_services/README.md) - Rust 微服务构建

---

## 🏗️ 系统架构

详见 **[系统架构蓝图](./docs/系统架构蓝图.md)**

### CQRS + 双引擎模式

```
客户端
  ↓
Gjallar (API 网关) :8080
  ├─ 写请求 → Nexus :9001 → Forge :9092 (同步渲染) → PostgreSQL
  ├─ 读请求 → Beacon :9002 → Redis Cache → PostgreSQL
  └─ 搜索请求 → Mirror :9093 (Tantivy 全文索引)
  
异步链路:
  Nexus → NATS 事件 → Oracle :9094 → 更新 Mirror 索引
```

### 核心服务

| 服务 | 端口 | 语言 | 职责 |
|------|------|------|------|
| **Gjallar** | 8080 (HTTP) | Go | API 网关、路由聚合 |
| **Nexus** | 9001 (gRPC) | Go | 写服务、同步渲染 |
| **Beacon** | 9002 (gRPC) | Go | 读服务、缓存策略 |
| **Forge** | 9092 (gRPC) | Rust | Markdown 渲染引擎 |
| **Mirror** | 9093 (gRPC) | Rust | 全文搜索引擎 |
| **Oracle** | 9094 (gRPC) | Rust | 异步处理、索引更新 |

### 基础设施

| 组件 | 用途 | 访问地址 |
|------|------|----------|
| PostgreSQL 16 | 主数据库 | :5432 |
| Redis 7 | 缓存 | :6379 |
| NATS 2 | 消息队列 | :4222 |
| MinIO | 对象存储 | :9000 (API), :9001 (Console) |
| Jaeger | 链路追踪 | :16686 (UI) |

---

## 💡 关键特性

### ✅ 同步渲染保证一致性

- Nexus 创建/更新文章时**同步调用 Forge**
- 5 秒超时控制，失败自动回滚事务
- 避免"已保存但未渲染"的中间状态

### ⚡ 高性能搜索

- **Mirror** 基于 Tantivy，<50ms 响应
- 实时索引更新（通过 Oracle 消费 NATS 事件）
- 支持分面统计、搜索建议

### 📦 Docker 优化

- **Rust 服务**: 依赖缓存策略，构建时间 5-10x 加速
- **Go 服务**: 静态编译，镜像从 50MB 降至 15MB
- 所有服务健康检查 + 依赖编排

### 🛠️ 统一管理

- **manage.ps1**: 17 个命令管理开发/Docker 全流程
- 参数验证 + Tab 补全
- 一键启动/停止/清理

---

## 📚 数据流

### 写入链路（同步）

```
1. 用户提交文章 → Gjallar
2. Gjallar → Nexus (gRPC)
3. Nexus 开启事务
4. Nexus → Forge (同步 gRPC，5s 超时)
5. Forge 渲染 Markdown → HTML
6. Nexus 存储到 PostgreSQL
7. 提交事务
8. 发布 NATS 事件
9. Oracle 消费 → 更新 Mirror 索引
```

### 读取链路

```
1. 用户查询 → Gjallar
2. Gjallar → Beacon (gRPC)
3. Beacon 检查 Redis 缓存
   ├─ Hit → 返回
   └─ Miss → 查 PostgreSQL → 写缓存
4. 返回结果
```

### 搜索链路（同步）

```
1. 用户搜索 "rust" → Gjallar
2. Gjallar → Mirror (同步 gRPC，5s 超时)
3. Mirror 执行 Tantivy 查询
4. 返回结果 (hits, total, took_ms)
```

---

## 🚀 快速开始

### 前置要求

- **Go 1.21+** (Go 服务)
- **Rust stable** (Rust 服务)
- **Docker & Docker Compose** (推荐)
- **PostgreSQL 14+** / **Redis 6+** (如果不用容器)

### 方法 1：Docker Compose（推荐，一键启动）

```bash
# 启动所有基础设施 + 服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 验证服务
curl http://localhost:8080/health

# 停止
docker-compose down
```

### 方法 2：本地开发（不使用 Docker）

```powershell
# 1. 启动基础设施
.\manage.ps1 docker-up-infra

# 2. Go 服务（需要 3 个终端）
cd go_services
go run cmd/nexus/main.go
go run cmd/beacon/main.go  
go run cmd/gjallar/main.go

# 3. Rust 服务（需要 3 个终端）
cd rust_services
cargo run --bin forge
cargo run --bin mirror
cargo run --bin oracle
```

---

## 🛠️ 开发工具

### manage.ps1 命令

```powershell
# 开发命令
.\manage.ps1 proto-lint         # 检查 .proto 文件
.\manage.ps1 proto-gen          # 生成 gRPC 代码
.\manage.ps1 build-go           # 编译 Go 服务
.\manage.ps1 format             # 格式化代码
.\manage.ps1 clean              # 清理构建产物

# Docker 命令
.\manage.ps1 docker-validate    # 验证 Docker 环境
.\manage.ps1 docker-build       # 构建所有镜像
.\manage.ps1 docker-build-go    # 仅构建 Go 镜像
.\manage.ps1 docker-build-rust  # 仅构建 Rust 镜像
.\manage.ps1 docker-up          # 启动所有服务
.\manage.ps1 docker-up-infra    # 仅启动基础设施
.\manage.ps1 docker-up-go       # 仅启动 Go 服务
.\manage.ps1 docker-up-rust     # 仅启动 Rust 服务
.\manage.ps1 docker-down        # 停止所有服务
.\manage.ps1 docker-restart -Service gjallar  # 重启指定服务
.\manage.ps1 docker-logs -Service nexus       # 查看日志
.\manage.ps1 docker-ps          # 查看服务状态
.\manage.ps1 docker-clean       # 清理未使用资源
.\manage.ps1 docker-clean-all   # 深度清理（含数据卷）
```

### 本地开发（不使用 Docker）

```powershell
# 1. 启动基础设施
.\manage.ps1 docker-up-infra

# 2. Go 服务（需要 3 个终端）
cd go_services
go run cmd/nexus/main.go
go run cmd/beacon/main.go  
go run cmd/gjallar/main.go

# 3. Rust 服务（需要 3 个终端）
cd rust_services
cargo run --bin forge
cargo run --bin mirror
cargo run --bin oracle
```

---

## 🧪 测试

详见 [测试指南](./docs/TESTING_GUIDE.md)

```powershell
# Go 单元测试
cd go_services
go test ./internal/...

# Rust 单元测试
cd rust_services
cargo test

# 集成测试
.\manage.ps1 docker-up
# 运行集成测试脚本...
```

---

## 📊 性能指标

| 服务 | 操作 | P95 延迟 | 优化策略 |
|------|------|----------|----------|
| **Forge** | Markdown 渲染 | 80ms | Rust 异步 + 语法高亮缓存 |
| **Mirror** | 全文搜索 | 35ms | Tantivy 内存索引 + BM25 |
| **Nexus** | 创建文章 | 420ms | 连接池 + 事务优化 |
| **Beacon** | 读取文章 (Cache) | 8ms | Redis 缓存 |
| **Beacon** | 读取文章 (DB) | 65ms | 索引优化 + Read Replica |

> 详细性能分析见 [系统架构蓝图](./docs/系统架构蓝图.md)

---

## 🔒 安全

- JWT 鉴权（HS256）
- XSS 防护（ammonia 库清洗 HTML）
- SQL 注入防护（参数化查询）
- 密码加密（bcrypt）
- HTTPS 支持（生产环境）
- CORS 配置

---

## 📈 监控与追踪

- **Jaeger**: 分布式链路追踪 → <http://localhost:16686>
- **Prometheus**: 指标收集（待实现）
- **Grafana**: 可视化仪表盘（待实现）
- **Dozzle**: 容器日志查看 → <http://localhost:9999>

---

## 🚢 生产部署

详见 [Docker 部署指南](./docs/Docker部署指南.md)

### 推荐配置

| 服务 | CPU | 内存 | 副本数 |
|------|-----|------|--------|
| Gjallar | 2 Core | 2GB | 2-4 |
| Nexus | 2 Core | 2GB | 2-4 |
| Beacon | 2 Core | 4GB | 4-8 |
| Forge | 4 Core | 4GB | 2-4 |
| Mirror | 4 Core | 8GB | 2-4 |
| Oracle | 2 Core | 2GB | 1-2 |
| PostgreSQL | 4 Core | 8GB | 1 (主从) |
| Redis | 2 Core | 4GB | 1 (集群) |

---

## 🤝 贡献

│   ├── search/v1/
│   │   └── mirror.proto          # 搜索 API
│   ├── common/v1/
│   │   └── common.proto          # 枚举 / 错误码
│   └── google/                   # Google API 注解（HTTP 映射）
│
├── go_services/                  # Go 微服务
│   ├── cmd/
│   │   ├── nexus/main.go         # 写入服务入口
│   │   ├── beacon/main.go        # 读取服务入口
│   │   └── gjallar/main.go       # HTTP 网关入口

1. 提交问题 (Issues)
2. Fork 并创建 PR
3. 遵循现有代码风格
4. 添加测试覆盖

---

## 📝 许可

MIT License

---

## 📞 联系方式

- **文档**: [./docs](./docs)
- **Issues**: GitHub Issues
- **Email**: <dev@bifrost.example>

---

## 🗓️ 更新日志

### v3.2 (2024-12-24)

- ✅ 实现 Nexus→Forge 同步渲染集成
- ✅ 实现 Gjallar→Mirror 同步搜索集成
- ✅ 优化 Docker 构建（Rust 1.91.1，Go 1.23）
- ✅ 统一管理脚本 manage.ps1（17 个命令）
- ✅ 完整的健康检查和服务依赖编排
- ✅ 文档整合到 docs/ 目录

### v3.1 (2024-11)

- ✅ CQRS 架构落地
- ✅ Go + Rust 混合服务
- ✅ NATS 消息队列集成
- ✅ OpenTelemetry 追踪

### v3.0 (2024-10)

- 🎉 项目初始化
- ✅ Monorepo 架构设计
- ✅ Proto 定义和代码生成
- ✅ 基础设施搭建

---

**🌈 Bifrost - 连接 Go 与 Rust 的彩虹桥 🌉**

  option (google.api.http) = {
    get: "/v1/posts/{slug_or_id}"
  };
}

```

### 3️⃣ 轻量级消息传递：NATS Fire-and-Forget

替代复杂的 Outbox 发件箱模式，采用更简洁的事件驱动架构。详见 [NATS 消息传递指南](./docs/NATS_MESSAGING_GUIDE.md)。

**核心特点**：

```go
// Nexus 发送端（业务完成后异步发送，不等待）
go func() {
    err := msgr.Publish("content.post.created", payload)
    if err != nil {
        log.Warn("publish failed", err) // 只记日志，不影响业务
    }
}()

// Beacon 消费端（使用 Queue Groups 自动负载均衡）
msgr.Subscribe("content.>", "beacon_service", func(subject, data []byte) {
    // 删除 Redis 缓存
})
```

**为什么适合 CMS？**

- ✅ **代码简洁**：无需 Outbox 表、Relayer 协程、重试逻辑
- ✅ **运维轻松**：NATS 开箱即用，无复杂配置
- ✅ **易于扩展**：新增消费者（如 Audit 服务）只需一行代码，Nexus 无需改动
- ✅ **宽松一致性**：事件丢失概率极低，即使丢失也只是缓存多保留几分钟

### 4️⃣ 高并发写入：Oracle 埋点服务

使用 **批量写缓冲 + DuckDB 列存优化**：

```rust
// Ingestor (后台线程) 缓冲事件
let ingestor = Ingestor::new(1000, Duration::from_secs(30));
// 条件1：缓冲满 (1000条) 或
// 条件2：时间到 (30秒)
// → 批量写入 DuckDB → 磁盘持久化
```

### 5️⃣ 全文检索：Mirror 搜索服务

使用 **Tantivy (Rust 原生搜索库)** + **后台索引 Worker**：

```rust
// 分词 + 倒排索引
"Bifrost 是一个 CMS" → ["bifrost", "cms", ...]
// 查询：AND / OR / 范围 / 模糊匹配 / 排序
```

### 6️⃣ 实时分析：Oracle 后台 Worker

**定期聚合原始日志 → 预计算报表**：

```rust
// 每小时执行
INSERT INTO daily_stats 
  SELECT date, metric, dimension, COUNT(*) 
  FROM raw_events 
  WHERE date = YESTERDAY;

// 前端查询 → 微秒级响应 (只需扫描几百行)
```

### 7️⃣ 可观测性：OpenTelemetry 全链路追踪

- **Trace ID** 自动传播（Context → RPC Header → Database）
- **Span** 记录每个操作（DB Query, gRPC Call, 缓存 Hit/Miss）
- **Metrics** 暴露为 Prometheus
- **日志** 结构化输出 (JSON) 并关联 Trace ID

```bash
# 前端 curl 一次请求
curl -H "X-Trace-Id: abc123" http://localhost:8080/v1/posts/123

# Jaeger 可视化：从前端 → Gjallar → Beacon → PostgreSQL 的完整链路
```

---

## 🧪 测试与质量保证

### 运行单元测试

```bash
# Go 服务
cd go_services
make test           # 全量测试
make test-race      # 竞态条件检测

# Rust 服务
cd rust_services
cargo test --all    # 全量测试
cargo test --doc    # 文档示例测试
```

### 代码质量检查

```bash
# Go 代码风格 & lint
cd go_services
make lint

# Rust 代码风格 & clippy
cd rust_services
cargo clippy --all-targets -- -D warnings
```

### 集成测试

```bash
# 启动完整环境
docker-compose up -d

# 运行集成测试脚本
bash test/integration.sh
```

---

## 📊 性能指标

### 基准测试结果

| 操作                      | 延迟 (P95) | 吞吐量           |
|-------------------------|----------|---------------|
| 获取文章列表 (Beacon + Redis) | 30ms     | 10K req/s     |
| 发布文章 (Nexus + Forge 渲染) | 150ms    | 1K req/s      |
| 全文搜索 (Mirror)           | 50ms     | 5K req/s      |
| 埋点上报 (Oracle 批量缓冲)      | 5ms      | 100K events/s |
| 仪表盘查询 (Oracle 聚合表)      | 10ms     | 50K req/s     |

> 基于本地开发环境测试（Intel i7, 16GB RAM）

---

## 🛠️ 开发与贡献

### 本地开发设置

```bash
# 克隆仓库
git clone https://github.com/gulugulu3399/bifrost.git
cd bifrost

# 安装依赖
## Go
go mod download

## Rust
rustup update

# 生成 proto 代码
make generate-proto

# 或手动
cd go_services && go generate ./...
cd ../rust_services && cargo build
```

### 代码提交规范

```bash
# 格式化代码
make fmt

# 运行 lint
make lint

# 提交前运行测试
make test

# Git 提交
git commit -m "feat(beacon): add post search endpoint"
# 格式: type(scope): subject
```

### 常见开发任务

| 任务         | 命令                                                                 |
|------------|--------------------------------------------------------------------|
| 修改 Proto   | `nano api/content/v1/*/xxx.proto` → `make generate-proto`          |
| 新增 DB 表    | `nano migrations/00X_xxx.sql` → Docker 重启 PostgreSQL               |
| 调试 gRPC    | `grpcurl -plaintext localhost:50051 list`                          |
| 查看数据库      | `docker exec -it bifrost-postgres psql -U bifrost_user -d bifrost` |
| 查看 NATS 消息 | NATS CLI 或 Web UI                                                  |

---

## 🚨 故障排查

### 常见问题

#### Q: Docker Compose 启动失败

```bash
# 检查端口占用
lsof -i :8080
lsof -i :50051

# 清除已有容器
docker-compose down -v

# 重新启动
docker-compose up --build
```

#### Q: gRPC 服务无法连接

```bash
# 检查服务健康
docker-compose logs nexus | tail -50

# 测试连接
grpcurl -plaintext localhost:50051 list

# 检查防火墙
firewall-cmd --list-ports
```

#### Q: PostgreSQL 连接错误

```bash
# 确认数据库创建
docker exec -it bifrost-postgres \
  psql -U bifrost_user -d bifrost -c "SELECT COUNT(*) FROM posts;"

# 查看迁移日志
docker-compose logs postgres
```

#### Q: Rust 编译错误

```bash
# 更新工具链
rustup update stable

# 清除构建缓存
cargo clean
cargo build

# 查看详细错误
RUST_BACKTRACE=1 cargo build
```

### 日志查看

```bash
# 所有服务日志
docker-compose logs -f

# 单个服务
docker-compose logs -f nexus

# 实时过滤错误
docker-compose logs -f --tail=100 | grep -i error
```

---

## 📚 学习资源

### 架构与设计

- [宏服务架构设计文档](./go_services/ADR-005_MACRO_SERVICES_ARCHITECTURE.md)
- [数据库设计与关系图](./migrations/readme.md)
- [HTTP 接口规范](./docs/HTTP_Guide.md)

### 代码示例

- [Go 服务模板](./go_services/internal/)
- [Rust 微服务模板](./rust_services/oracle/)
- [Proto 定义](./api/)

### 参考资源

- [Protocol Buffers 官方文档](https://developers.google.com/protocol-buffers)
- [gRPC 教程](https://grpc.io/docs/languages/go/)
- [PostgreSQL 最佳实践](https://www.postgresql.org/docs/)
- [OpenTelemetry 指南](https://opentelemetry.io/docs/)
- [Tantivy 全文搜索](https://docs.rs/tantivy/)
- [DuckDB 列式存储](https://duckdb.org/docs/)

---

## 📄 许可证

MIT License - 详见 [LICENSE](./LICENSE) 文件

---

## 👥 贡献者

- **GuLuGuLu3399** - 主要开发者
- 所有贡献者欢迎！🙌

---

## 📞 联系与支持

- 📧 Email: <developer@bifrost.example.com>
- 💬 Issues: [GitHub Issues](https://github.com/gulugulu3399/bifrost/issues)
- 📖 Wiki: [Project Wiki](https://github.com/gulugulu3399/bifrost/wiki)

---

## 🎯 项目路线图

### v3.2.0（当前）

- ✅ 核心 CRUD 操作
- ✅ gRPC + HTTP 双协议
- ✅ PostgreSQL 强一致性
- ✅ Redis 缓存层
- ✅ Markdown 渲染 (Forge)
- ✅ 全文检索 (Mirror)
- ✅ 埋点分析 (Oracle)
- ✅ OpenTelemetry 可观测性

### v3.3.0（规划中）

- 📋 用户权限管理 (RBAC)
- 📋 草稿箱与版本控制
- 📋 评论线程优化
- 📋 CDN 集成（图片优化）
- 📋 WebSocket 实时通知

### v4.0.0（远期）

- 🚀 分布式事务 (Saga)
- 🚀 AI 内容推荐
- 🚀 多语言国际化
- 🚀 移动 App 支持
- 🚀 开源社区版

---

**祝你使用愉快！** 🌈✨
