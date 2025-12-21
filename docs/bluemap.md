# Bifrost CMS v3.2 (Pure Edition) - 架构蓝图

## 1. 战略规划 (Strategy)

### 核心设计哲学

本架构严格遵循 **CQRS (命令查询职责分离)** 模式，配合 **Go (业务编排)** 与 **Rust (高性能计算)** 的双引擎驱动。

### 代码库策略: Monorepo

我们将所有服务统一管理，确保协议（Proto）与依赖的一致性。

```
bifrost-v3/
├── api/                  # gRPC Protobuf 定义 (通用语言)
│   ├── common/           # 共享消息类型
│   ├── write/            # [Nexus] 写服务接口
│   ├── read/             # [Beacon] 读服务接口
│   ├── search/           # [Mirror] 搜索服务接口
│   └── analysis/         # [Oracle] BI 分析接口
├── build/                # Dockerfiles & CI/CD
├── cmd/                  # 服务入口
│   ├── gjallar/          # [Go] 业务网关 (BFF)
│   ├── nexus/            # [Go] 写服务 (Command Side)
│   ├── beacon/           # [Go] 读服务 (Query Side)
│   ├── forge/            # [Rust] 渲染工作者
│   ├── mirror/           # [Rust] 搜索引擎服务
│   └── oracle/           # [Rust] BI 数据分析服务
├── configs/              # OpenResty & App Configs
│   ├── heimdall/         # [OpenResty] Lua 脚本与 Nginx 配置
│   └── ...
├── deploy/               # K8s / Helm
├── internal/             # Go 业务逻辑
├── rust_crates/          # Rust 业务逻辑
├── web/                  # 前端 (Next.js)
└── go.mod
```

## 2. 兵力部署 (Service Topology)

### 🛡️ 第一道防线：边缘与网关

#### A. Heimdall (海姆达尔) - 守门人 (Ingress Gateway)

- **技术栈**: **OpenResty (Nginx + Lua)**
- **职责**:
  - **流量清洗**: WAF 防护、黑白名单、防爬虫。
  - **身份验签**: 在边缘层校验 JWT 签名（Lua 脚本），非法请求直接拒绝，减轻后端压力。
  - **限流熔断**: 基于 Token Bucket 的全局限流。
  - **SSL 卸载**: 证书管理与 HTTPS 终止。

#### B. Gjallar (加拉尔) - 咆哮者 (API Gateway / BFF)

- **技术栈**: **Go**
- **职责**:
  - **聚合层 (BFF)**: 面向前端（Web/Mobile）的 API 适配。例如，一个“文章详情页”请求，由 Gjallar 并行调用 Beacon（查内容）和 Mirror（查相关推荐），聚合后返回。
  - **协议转换**: HTTP/JSON <-> gRPC。
  - **短缓存**: 对高频 API 进行秒级微缓存。

### 🧠 核心阵营：Go (业务逻辑 CQRS)

#### C. Nexus (连结者) - 写服务 (Command Side)

- **技术栈**: **Go**
- **职责 (Write Only)**:
  - **全权负责写入**: 用户注册、发布文章、发表评论、点赞。
  - **身份管理 (Identity)**: 处理用户注册、密码修改等写入逻辑。
  - **事务核心**: 维护 `users`, `posts`, `comments` 的主表写入。
  - **发件箱 (Outbox)**: 确保所有写操作都生成事件 (`post.created`, `user.registered`) 推送至 NATS。

#### D. Beacon (灯塔) - 读服务 (Query Side)

- **技术栈**: **Go**
- **职责 (Read Only)**:
  - **高性能读取**: 提供文章列表、详情、用户信息、评论列表的查询。
  - **读模型优化**: 直接读取经过 Forge 渲染好的 `html_body` 和 `toc_json`。
  - **多级缓存**:
    - L1: 本地内存缓存 (BigCache)。
    - L2: 分布式缓存 (Redis)。
  - **视图表**: 维护专门用于列表展示的非规范化数据（如有必要）。

### ⚔️ 力量阵营：Rust (计算与数据)

#### E. Forge (熔炉) - 渲染服务 (Worker)

- **技术栈**: **Rust**
- **职责**:
  - **事件驱动**: 监听 NATS `post.created` / `post.updated`。
  - **Markdown 编译**: 使用 `pulldown-cmark` 将 Nexus 写入的 Raw Markdown 转换为 HTML。
  - **数据清洗**: 防止 XSS 注入。
  - **回写**: 将渲染结果通过 gRPC 回写给 Nexus（或直接更新读库，视一致性要求而定）。

#### F. Mirror (镜像) - 搜索服务 (Search)

- **技术栈**: **Rust**
- **职责**:
  - **全文检索**: 维护倒排索引 (Tantivy / MeiliSearch)。
  - **数据同步**: 消费 NATS 事件，实时构建索引。
  - **搜索接口**: 提供高性能的文章搜索、自动补全、相关性推荐。
  - **替代原静态资源功能**: 之前定义的静态资源代理功能可由 Nginx (Heimdall) 直接承担，Mirror 专注搜索。

#### G. Oracle (神谕) - BI 服务 (Analysis)

- **技术栈**: **Rust**
- **职责**:
  - **数据仓库**: 聚合用户行为数据（浏览、点赞、停留时长）。
  - **离线/实时计算**: 每日生成热度榜单、用户画像分析、内容质量评分。
  - **大屏报表**: 为管理员后台提供数据可视化 API。
  - **定时任务**: 负责清理过期日志、归档旧数据。

## 3. 数据流转 (Data Flow)

### 场景：发布文章 (The Write Path)

1. **User** -> **Heimdall** (校验 JWT) -> **Gjallar** (转发) -> **Nexus** (gRPC CreatePost)。
2. **Nexus**:
   - 生成 Snowflake ID。
   - 写入 PostgreSQL `posts` 表 (Status=Draft, Content=Markdown)。
   - 写入 `outbox_events` 表 (Topic=`post.created`)。
   - 返回 "发布成功" 给用户。
3. **Nexus Poller**: 扫描 Outbox，推送消息到 NATS。
4. **Forge (Rust)**: 收到消息 -> 拉取 Markdown -> 渲染 HTML -> 调用 Nexus 回写 HTML -> 标记 Outbox 为 Done。

### 场景：浏览文章 (The Read Path)

1. **User** -> **Heimdall** -> **Gjallar** -> **Beacon** (gRPC GetPost)。
2. **Beacon**:
   - 查 Redis 缓存。
   - (Miss) 查 PostgreSQL `html_body` 字段。
   - 返回 HTML 给用户。

### 场景：搜索文章 (The Search Path)

1. **User** -> **Heimdall** -> **Gjallar** -> **Mirror** (gRPC Search)。
2. **Mirror**:
   - 在 Tantivy 索引中查找。
   - 返回 ID 列表。
3. **Gjallar**: 拿着 ID 列表去 **Beacon** 批量查询文章简要信息 -> 聚合返回。

## 4. 基础设施 (Infrastructure)

- **Heimdall**: OpenResty 官方镜像 + Lua 脚本挂载。
- **Message Queue**: NATS JetStream (持久化、高性能)。
- **Database**: PostgreSQL 16 (Nexus/Beacon 共用实例，但逻辑分离；或物理分离)。
- **Cache**: Redis 7 Cluster。
- **Search**: Tantivy (嵌入式 Rust 库) 。