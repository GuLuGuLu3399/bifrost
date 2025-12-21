# 🌈 Bifrost Core Module

> **The Go-based heart of Bifrost**  
> 多语言单体仓库下的宏服务架构 (Polyglot Monorepo with Macro-Services)

---

## 📖 核心文档

| 文档                                                         | 用途                             |
|-----------------------------------------------------------|--------------------------------|
| [ADR-005: 宏服务架构](./ADR-005_MACRO_SERVICES_ARCHITECTURE.md) | ⭐ **必读**：理解 Bifrost Core 的架构本质 |
| [HTTP.md](./HTTP.md)                                      | 📡 API 接口文档                    |
| [Makefile](./Makefile)                                    | 🛠️ 构建和运行命令速查                  |

---

## 🏛️ 架构理念：不是单体，也不是微服务

Bifrost Core 是一个**"黄金平衡点"**的架构：

### 部署维度 🟢 微服务式

```txt
Nexus ─┐     Beacon ─┐     Forge (Rust)     Oracle (Python)
       └─ shared DB ─┘
```

- ✅ 独立容器化
- ✅ 故障隔离
- ✅ 按需扩容

### 数据维度 🟡 单体式

```txt
Nexus + Beacon ── 共享 PostgreSQL ── ACID 事务
                                   └─ 无分布式复杂性
```

- ✅ 强一致性
- ✅ 避免分布式事务
- ✅ 开发效率高

### 代码维度 🟡 单体式

```txt
core/
├── internal/biz      ← 共享业务实体
├── internal/data     ← 共享数据访问
├── pkg/              ← 共享工具库
└── cmd/
    ├── nexus/        ← 写入专用
    ├── beacon/       ← 读取专用
    └── portal/       ← 网关
```

- ✅ 代码复用
- ✅ 原子性提交
- ✅ 本地完整验证

**结论**：兼得单体和微服务的优点，绕过各自的坑。

---

## 🚀 快速启动

### 前置要求

- Go 1.21+
- PostgreSQL 14+
- Redis 6+ (可选)
- MeiliSearch (可选)

### 方法 1: 使用 Makefile（推荐）

```bash
# 查看所有可用命令
make help

# 构建所有服务
make build-all

# 在不同终端运行服务
# 终端 1: Nexus (写服务)
make run-nexus

# 终端 2: Beacon (读服务)
make run-beacon

# 终端 3: Gjallar (网关)
make run-gjallar
```

### 方法 2: 使用 PowerShell 脚本（自动启动全部）

```powershell
# Windows PowerShell
.\scripts\run-all.ps1
```

### 方法 3: 手动启动

```bash
# 终端 1: Nexus (写入服务) - 端口 9001
go run ./cmd/nexus -f configs/nexus.yaml

# 终端 2: Beacon (读取服务) - 端口 9002
go run ./cmd/beacon -f configs/beacon.yaml

# 终端 3: Gjallar (网关) - 端口 8080
go run ./cmd/gjallar
```

### 测试 API

所有请求通过 Gjallar 网关 (`:8080`)：

```bash
# 读操作 → 自动路由到 Beacon
curl http://localhost:8080/api/v1/posts
curl http://localhost:8080/api/v1/posts/1

# 写操作 → 自动路由到 Nexus (需要认证)
curl -X POST http://localhost:8080/api/v1/posts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer dev-secret" \
  -d '{
    "title": "Hello Bifrost",
    "content": "Testing multi-module architecture",
    "author_id": 1,
    "category_id": 1
  }'
```

### 验证服务状态

```bash
# 健康检查
curl http://localhost:8080/health

# Prometheus Metrics
curl http://localhost:8081/metrics  # Nexus
curl http://localhost:8082/metrics  # Beacon
```

---

## 📁 目录结构

```txt
core/
├── go.mod                    # 统一 Go 模块
├── Makefile                  # 中央构建
│
├── cmd/                      # 三个独立服务入口
│   ├── nexus/                # CMS 管理后台（写主导）
│   │   ├── main.go           # 启动入口
│   │   ├── wire.go           # 依赖注入定义
│   │   └── internal/handler/ # 业务处理器
│   ├── beacon/               # Web 前台（读主导）
│   │   ├── main.go
│   │   ├── wire.go
│   │   └── internal/handler/
│   └── portal/               # Gateway 网关
│       ├── main.go
│       └── config.yaml
│
├── internal/                 # 共享内部核心
│   ├── conf/                 # 配置定义
│   ├── data/                 # 数据访问层
│   │   ├── data.go           # DB 初始化
│   │   ├── schema/           # GORM Schema
│   │   └── *.go              # DAO 实现
│   ├── biz/                  # 业务实体与接口
│   │   └── biz.go            # 领域模型
│   ├── service/              # 业务服务层
│   ├── app/                  # 应用程序逻辑
│   ├── clients/              # 外部服务客户端
│   ├── middleware/           # HTTP 中间件
│   ├── handler/              # 共享 HTTP 处理器（可选）
│   └── ...
│
└── pkg/                      # 共享工具库
    ├── logger/               # 日志
    ├── auth/                 # 认证
    ├── httpx/                # HTTP 工具
    ├── rpcx/                 # RPC 工具
    ├── security/             # 安全库
    ├── xerr/                 # 错误处理
    └── circuitbreaker/       # 熔断器
```

---

## 🛠️ 核心职责分工

### 🟦 Nexus (写主导)

- 所有 POST/PUT/DELETE 操作
- 业务状态验证
- 调用 Forge 进行内容清洗
- **监听端口**: 8081

### 🟦 Beacon (读主导)

- 所有 GET 操作
- **CDN URL 拼接**：`storage_path` → `https://cdn.example.com/path`
- HTTP 缓存头设置
- SEO 友好的数据格式
- **监听端口**: 8082

### 🟦 Portal (网关)

- 统一认证 (JWT)
- 路由分发 (写→Nexus, 读→Beacon)
- 限流与熔断
- **监听端口**: 8080

### 🟧 Forge (Rust 引擎)

- HTML/Markdown 清洗 (XSS 防御)
- 文件魔数校验
- PDF 生成
- **通信**: gRPC

### 🟨 Oracle (Python 决策)

- 异步任务消费
- 数据分析与 BI
- Julia 脚本执行（高性能计算）
- **通信**: 消息队列 + gRPC

---

## 📋 关键设计原则

### 1. 代码分层 (Layered Architecture)

```txt
HTTP Request
    ↓
├─ Handler (cmd/nexus/internal/handler)  ← 路由 & 请求验证
├─ Service (internal/service)            ← 业务逻辑
├─ Biz (internal/biz)                    ← 业务实体 & 接口
├─ Data (internal/data)                  ← 数据访问
└─ DB (PostgreSQL)
```

### 2. 不能import的规则 (Import Restrictions)

```txt
❌ cmd/nexus → import cmd/beacon (私有代码)
❌ internal/data → import cmd/nexus (业务逻辑)
❌ pkg/* → import internal/* (纯工具库)

✅ cmd/nexus → import internal/biz
✅ cmd/nexus → import internal/data
✅ cmd/nexus → import pkg/*
```

### 3. 协议优先 (Protocol-First)

- **Go ↔ Go**：直接共享代码（`internal` 包）
- **Go ↔ Rust/Python**：gRPC
- **Async 通信**：消息队列 (PostgreSQL Outbox)

---

## 🔄 典型工作流

### 创建新的业务实体

1. **定义业务实体**

   ```go
   // internal/biz/article.go
   type Article struct {
       ID    uint
       Title string
   }
   
   type ArticleRepo interface {
       GetArticle(ctx context.Context, id uint) (*Article, error)
       CreateArticle(ctx context.Context, a *Article) error
   }
   ```

2. **定义数据库 Schema**

   ```go
   // internal/data/schema/article.go
   type Article struct {
       gorm.Model
       Title string
   }
   ```

3. **实现 Repository**

   ```go
   // internal/data/article.go
   func (r *articleRepo) GetArticle(ctx context.Context, id uint) (*biz.Article, error) {
       var schema schema.Article
       if err := r.data.db.First(&schema, id).Error; err != nil {
           return nil, err
       }
       return &biz.Article{ID: schema.ID, Title: schema.Title}, nil
   }
   ```

4. **创建 Service**

   ```go
   // internal/service/article_service.go
   type ArticleService struct {
       repo biz.ArticleRepo
   }
   
   func (s *ArticleService) GetArticle(ctx context.Context, id uint) (*biz.Article, error) {
       return s.repo.GetArticle(ctx, id)
   }
   ```

5. **创建 Handler**

   ```go
   // cmd/nexus/internal/handler/article.go (如果是写操作)
   // cmd/beacon/internal/handler/article.go (如果是读操作)
   
   func (h *ArticleHandler) GetArticle(c *gin.Context) {
       article, err := h.service.GetArticle(c.Context(), id)
       c.JSON(200, article)
   }
   ```

---

## 📊 性能考量

| 维度            | 特点                         | 优化策略                            |
|---------------|----------------------------|---------------------------------|
| **Nexus 吞吐**  | 受 DB 写入限制                  | 批量操作、异步队列                       |
| **Beacon 吞吐** | 受 DB 连接数限制                 | Connection Pool、Read Replica、缓存 |
| **一致性**       | 强一致（ACID）                  | 本地缓存 TTL 需注意                    |
| **跨进程通信**     | Nexus ↔ Beacon 无 RPC（代码共享） | 通过 DB 同步                        |

---

## ⚖️ 何时考虑进一步拆分

当以下指标达到时，考虑进一步拆分为真正的微服务：

- 📈 **QPS > 100k/s**：单 PostgreSQL 实例性能瓶颈
- 📈 **数据量 > 10TB**：分库分表变得必须
- 📈 **团队 > 20 人**：分布式开发需求增加
- 📈 **语言多样化**：超过 3 种语言，单体仓库管理成本高

**此时的进化路径**：

```txt
当前状态                 →  未来状态
┌────────┐              ┌─────────────┐
│ Nexus  │──────┐       │ Nexus DB A  │
│ Beacon │──┬───┼──→    │ Beacon DB B │
└────────┘  │   │       │ (分库)      │
        shared DB       └─────────────┘
```

---

## 🐛 故障恢复

### 场景 1：Nexus 故障

```txt
写操作失败 ─ 返回错误给客户端
读操作正常 ─ Beacon 继续服务最后一致的数据
```

**恢复**：修复 Nexus，重新启动 → 继续接收写操作

### 场景 2：Beacon 故障

```txt
读操作失败 ─ 返回错误给客户端
写操作正常 ─ Nexus 继续写入，数据持久化到 DB
```

**恢复**：修复 Beacon，重新启动 → 读取最新数据

### 场景 3：PostgreSQL 故障

```txt
系统完全不可用 ─ 需要 DB 恢复或故障转移
```

**防护**：

- ✅ 主从复制 + Failover
- ✅ 定期备份
- ✅ PgBouncer 连接池

---

## 📚 延伸阅读

- **[ADR-005: 宏服务架构](./ADR-005_MACRO_SERVICES_ARCHITECTURE.md)** - 完整的架构决策记录
- **[HTTP.md](./HTTP.md)** - HTTP API 接口文档

---

## 🚀 贡献指南

修改 Core 模块时，请遵守：

1. **代码分层**：
   - ❌ 不要在 Handler 中写 SQL 查询
   - ✅ 通过 Repository 访问数据
   - ❌ 不要在 Service 中做路由逻辑

2. **跨服务通信**：
   - ✅ Nexus ↔ Beacon：通过共享 `internal/biz` 和数据库
   - ✅ Nexus ↔ Forge：通过 gRPC
   - ❌ 避免 HTTP 调用同一 Core 的其他服务

3. **测试**：

   ```bash
   go test ./...
   ```

---

**Bifrost Core：开发爽，运维爽。** 🌈
