# 🌈 Bifrost：多语言博客/CMS 平台

> 一个基于 **Go + Rust 混合微服务架构** 的现代内容管理与分析平台
>
> ✨ **高性能** | 📊 **实时分析** | 🔍 **全文检索** | 🎨 **内容渲染** | 📈 **可观测性**

---

## 📋 快速导航

| 文档                                                | 用途                   |
|---------------------------------------------------|----------------------|
| [Go Services README](./go_services/README.md)     | Go 核心服务架构 & 部署指南     |
| [Rust Services README](./rust_services/README.md) | Rust 微服务构建 & 配置      |
| [HTTP API 指南](./docs/HTTP_Guide.md)               | 接口交互规范与示例            |
| [数据库架构](./migrations/readme.md)                   | PostgreSQL 表设计 & 关系图 |
| [NATS 消息传递](./docs/NATS_MESSAGING_GUIDE.md)     | Fire-and-Forget 轻量级方案 |
| [代码审计报告](./docs/code_audit_report.md)             | 质量评估与改进建议            |

---

## 🏗️ 项目架构概览

### 核心设计理念：**黄金平衡的宏服务架构**

```txt
┌─────────────────────────────────────────────────────────────────┐
│                        前端应用 (Frontend)                       │
│                    (React/Vue/Next.js)                          │
└────────┬──────────────────────────────────────┬──────────────────┘
         │                                      │
    HTTP/1.1                              HTTP/1.1
     REST                                  REST
         │                                      │
    ┌────▼────────┐                   ┌────────▼─────┐
    │  Gjallar    │                   │  Gjallar     │
    │  (Gateway)  │                   │  (Mirror)    │
    └────┬────────┘                   └────────┬─────┘
         │                                      │
    ┌────▼────────────────┬────────────────────▼─────────┐
    │                     │                              │
   gRPC                  gRPC                          gRPC
    │                     │                              │
    │                     │                              │
┌───▼────┐          ┌────▼───┐                  ┌────────▼────┐
│ Nexus  │◄────────►│ Beacon │                  │   Forge     │
│(写服务)│ PostgreSQL  (读服务)                  │(渲染服务)   │
└───┬────┘          └────┬───┘                  └────────┬────┘
    │                     │ PostgreSQL + Redis          │
    │                     │                              │
    │  ┌──────────────────┴──────────────┐      ┌───────┴────────┐
    │  │                                 │      │                │
   NATS                              ┌───▼─────┴──────┐    ┌──────▼──────┐
    │  │                             │  Oracle (Rust) │    │Mirror(Rust) │
    │  │                             │  (BI 分析)     │    │(全文检索)   │
    │  │                             │  DuckDB        │    │Tantivy      │
    │  │                             │  + Worker      │    │+ Indexer    │
    │  │                             └────────────────┘    └─────────────┘
    │  │
    │  │  Queue Groups 负载均衡：
    │  │  - beacon_service  (缓存失效)
    │  │  - mirror_service  (索引更新)
    │  │  - 可扩展至其他服务
    │  │
    │  └─► NATS (Fire-and-Forget)
    │
    └────────────────────────────────┐
                                      │
                            ┌─────────▼─────────┐
                            │   MinIO (S3)      │
                            │ (头像/附件存储)   │
                            └───────────────────┘

```

### 🟢 部署维度：微服务式

- ✅ 每个服务独立容器（Nexus, Beacon, Forge, Oracle, Mirror）
- ✅ 故障隔离（一个服务宕机不影响其他）
- ✅ 按需扩容（热服务可单独增加副本）

### 🟡 数据维度：单体式

- ✅ Nexus + Beacon 共享单一 PostgreSQL 数据库 → 强一致性
- ✅ 避免分布式事务复杂性，使用**NATS Fire-and-Forget 轻量级方案**
- ✅ Redis 作为 Session/缓存层，非关键数据

### 🟡 代码维度：单体式

- ✅ 所有 Go 服务共享 `common/` 库和 `pkg/` 工具集
- ✅ 所有 Rust 服务共享 `common/` crate
- ✅ 代码复用 > 服务隔离，原子性提交

---

## 🔧 核心服务列表

### Go 服务 (Bifrost Core)

| 服务          | 端口    | 职责                    | 技术栈                                |
|-------------|-------|-----------------------|------------------------------------|
| **Nexus**   | 50051 | 写入核心（发布文章、评论、互动）      | gRPC, PostgreSQL, Redis, NATS      |
| **Beacon**  | 50052 | 读取核心（查询文章、评论、统计）      | gRPC, PostgreSQL, Redis, LRU Cache |
| **Gjallar** | 8080  | HTTP 网关（REST to gRPC） | gRPC Gateway, JWT Auth             |

### Rust 服务 (Bifrost Engines)

| 服务         | 端口    | 职责                                         | 技术栈                          |
|------------|-------|--------------------------------------------|------------------------------|
| **Forge**  | 50053 | Markdown 渲染（Editor Preview / Batch Render） | Tokio, Pulldown-cmark, gRPC  |
| **Mirror** | 50054 | 全文检索（Index / Query）                        | Tokio, Tantivy, NATS Worker  |
| **Oracle** | 50055 | BI 分析（Track Events / Dashboard）            | Tokio, DuckDB, Worker + Cron |

### 基础设施 (Infrastructure)

| 组件             | 用途           | 配置                   |
|----------------|--------------|----------------------|
| **PostgreSQL** | 主数据存储        | 16-Alpine, 共享数据库     |
| **Redis**      | Session / 缓存 | 7-Alpine             |
| **NATS**       | 异步消息队列       | 2-Alpine + JetStream |
| **MinIO**      | 对象存储（头像、附件）  | S3-compatible        |
| **Jaeger**     | 分布式追踪        | OpenTelemetry        |

---

## 📊 数据流与交互模式

### 流写入场景 (典型：发布文章)

```
Frontend
    │ POST /v1/posts
    ▼
  Gjallar (网关)
    │ gRPC
    ▼
  Nexus (写服务)
    │ [1] 写入 posts 表
    │ [2] 调用 Forge 渲染 Markdown → HTML
    │ [3] 事务提交成功
    │ [4] go func() 异步发送 NATS 消息 (post.created)
    ▼
 PostgreSQL 事务成功返回 + NATS 消息分发
    │
    ├─► Beacon (beacon_service) 订阅 → 删除 Redis 缓存
    ├─► Mirror (mirror_service) 订阅 → 更新全文索引
    └─► Oracle (未来) 订阅 → 聚合分析
```

### 快读场景 (典型：查询文章列表)

```txt
Frontend
    │ GET /v1/posts
    ▼
  Gjallar (网关)
    │ gRPC
    ▼
  Beacon (读服务)
    │ [1] 查询 Redis 缓存 (miss → 查数据库)
    │ [2] 执行 PostgreSQL 查询
    │ [3] 填充缓存 (2h TTL)
    ▼
  返回 JSON 响应 (< 100ms)
```

### 分析场景 (典型：获取仪表盘指标)

```txt
前端埋点（通过 Gjallar 转发）
    │ TrackEvent RPC
    ▼
  Oracle (分析服务)
    │ [1] 批量接收事件（缓冲 1000 条或 30s）
    │ [2] 写入 raw_events (DuckDB)
    │ [3] Worker 后台定期聚合 → daily_stats 表
    ▼
  GetDashboardStats
    │ 查询 daily_stats (秒级响应)
    │ 返回 PV / UV / 热点分析
    ▼
  仪表盘展示
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

### 方法 2：本地开发（分离启动）

#### 启动基础设施

```bash
# PostgreSQL + Redis + NATS + MinIO
docker-compose up -d postgres redis nats minio jaeger
```

#### 启动 Go 服务

```bash
cd go_services

# 终端 1: Nexus (写服务)
make run-nexus

# 终端 2: Beacon (读服务)
make run-beacon

# 终端 3: Gjallar (网关)
make run-gjallar
```

#### 启动 Rust 服务

```bash
cd rust_services

# 终端 4: Forge (渲染)
cargo run -p forge

# 终端 5: Mirror (搜索)
cargo run -p mirror

# 终端 6: Oracle (分析)
cargo run -p oracle
```

#### 验证服务健康

```bash
# HTTP REST 接口
curl http://localhost:8080/health

# gRPC 服务
grpcurl -plaintext localhost:50051 list

# Prometheus 指标
curl http://localhost:9090/metrics

# Jaeger 追踪
open http://localhost:16686
```

---

## 📖 API 使用示例

### 1. 用户认证

```bash
# 注册
curl -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alice",
    "email": "alice@example.com",
    "password": "secure_password_123"
  }'

# 登录
curl -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alice",
    "password": "secure_password_123"
  }'
# 返回 JWT token，后续请求使用 Authorization: Bearer <token>
```

### 2. 发布文章

```bash
curl -X POST http://localhost:8080/v1/posts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "title": "Bifrost 架构解析",
    "slug": "bifrost-architecture",
    "raw_markdown": "# Bifrost\n\n这是一个...",
    "category_id": "1001",
    "tags": ["architecture", "golang", "rust"]
  }'
```

### 3. 查询文章（带分页）

```bash
curl "http://localhost:8080/v1/posts?page=1&page_size=20&sort=-created_at"
```

### 4. 获取文章详情

```bash
curl "http://localhost:8080/v1/posts/bifrost-architecture"
```

### 5. 内容预览（实时渲染）

```bash
curl -X POST http://localhost:8080/v1/render/preview \
  -H "Content-Type: application/json" \
  -d '{
    "raw_markdown": "# 标题\n\n**粗体**内容",
    "mode": "article"
  }'
```

### 6. 埋点事件（分析）

```bash
curl -X POST http://localhost:8080/v1/analysis/events \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": "post_view",
    "user_id": 123,
    "target_id": "post-456",
    "meta": {
      "referer": "https://google.com",
      "device": "mobile"
    }
  }'
```

### 7. 获取仪表盘指标（Admin Only）

```bash
curl "http://localhost:8080/v1/analysis/dashboard" \
  -H "Authorization: Bearer <admin_token>"
```

---

## 📂 项目结构

```txt
bifrost/
│
├── api/                          # Proto 定义（Single Source of Truth）
│   ├── content/v1/
│   │   ├── nexus/                # 用户 / 发布 / 写入 API
│   │   ├── beacon/               # 查询 / 读取 API
│   │   ├── forge/                # 渲染 API
│   │   ├── oracle/               # 分析 API
│   │   └── models.proto          # 共享数据模型
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
│   ├── internal/
│   │   ├── biz/                  # 业务逻辑层（shared）
│   │   ├── data/                 # 数据访问层（shared）
│   │   ├── server/               # gRPC 服务实现（shared）
│   │   └── middleware/           # 中间件（logging, auth, etc.）
│   ├── pkg/                      # 共享工具库
│   │   ├── cache/                # LRU / 分布式缓存
│   │   ├── database/             # PostgreSQL 连接 / 迁移
│   │   ├── contextx/             # Context 扩展
│   │   ├── id/                   # Snowflake ID 生成
│   │   ├── xerr/                 # 错误处理
│   │   ├── lifecycle/            # 优雅启动 / 关闭
│   │   └── ... (其他)
│   ├── Makefile                  # 快速命令（build / run / test）
│   ├── go.mod & go.sum
│   └── README.md
│
├── rust_services/                # Rust 微服务 (Cargo Workspace)
│   ├── common/                   # 共享 crate（proto, otel, logger）
│   │   ├── src/
│   │   │   ├── lib.rs            # 重导出 proto 和公用函数
│   │   │   ├── logger.rs         # tracing 初始化
│   │   │   ├── trace.rs          # OpenTelemetry 追踪
│   │   │   ├── lifecycle.rs      # 优雅关闭
│   │   │   └── nats/             # NATS 客户端
│   │   └── build.rs              # proto 代码生成
│   │
│   ├── forge/                    # Markdown 渲染服务
│   │   ├── src/
│   │   │   ├── main.rs
│   │   │   ├── server.rs         # gRPC 服务实现
│   │   │   └── engine.rs         # 渲染引擎核心
│   │   └── Cargo.toml
│   │
│   ├── mirror/                   # 全文搜索引擎
│   │   ├── src/
│   │   │   ├── main.rs
│   │   │   ├── server.rs         # gRPC 服务实现
│   │   │   ├── engine.rs         # Tantivy 搜索核心
│   │   │   ├── tokenizer.rs      # 分词器
│   │   │   └── worker.rs         # 后台索引构建
│   │   └── Cargo.toml
│   │
│   ├── oracle/                   # BI 分析引擎
│   │   ├── src/
│   │   │   ├── main.rs
│   │   │   ├── server.rs         # gRPC 服务实现
│   │   │   ├── ingestion.rs      # 高并发事件接收
│   │   │   ├── worker.rs         # ✨ 后台聚合 Worker
│   │   │   └── storage/
│   │   │       ├── duck.rs       # DuckDB 存储层
│   │   │       └── schema.sql    # 表定义
│   │   └── Cargo.toml
│   │
│   ├── Cargo.toml                # Workspace 入口
│   ├── Dockerfile                # 多服务通用镜像
│   ├── docker-compose.yml        # (备选，部分服务)
│   └── README.md
│
├── migrations/                   # PostgreSQL 数据库迁移
│   ├── 001_identity_and_infra.sql
│   ├── 002_content_core.sql
│   ├── 003_interactions.sql
│   └── readme.md
│
├── docs/                         # 文档
│   ├── HTTP_Guide.md             # 接口规范
│   ├── code_audit_report.md      # 代码审计
│   ├── TEST_SPECIFICATION.md
│   ├── RUST_UNIT_TEST_PLAN.md
│   └── bluemap.md
│
├── docker-compose.yml            # 主编排文件
├── buf.yaml & buf.gen.yaml       # Protocol Buffer 配置
├── manage.ps1                    # PowerShell 管理脚本
└── README.md                     # （本文件）
```

---

## 🔌 核心特性详解

### 1️⃣ 宏服务架构的优势

| 特性     | 传统单体  | 微服务   | 宏服务 (Bifrost) |
|--------|-------|-------|---------------|
| 开发效率   | ⭐⭐⭐⭐⭐ | ⭐⭐    | ⭐⭐⭐⭐⭐         |
| 代码复用   | ⭐⭐⭐⭐⭐ | ⭐⭐    | ⭐⭐⭐⭐⭐         |
| 隔离故障   | ❌     | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐          |
| 分布式复杂性 | ❌     | ⭐⭐⭐   | ⭐⭐            |
| 一致性保障  | ⭐⭐⭐⭐⭐ | ❌     | ⭐⭐⭐⭐          |

### 2️⃣ gRPC + HTTP 双协议

- **内部通信**：gRPC (Protobuf) → 高效、类型安全、自文档化
- **客户端通信**：HTTP REST (JSON) → 易于集成、前端友好
- **网关层**：Gjallar 使用 `grpc-gateway` 自动转换

```proto
rpc GetPost(GetPostRequest) returns (GetPostResponse) {
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
