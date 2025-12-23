# Bifrost CMS 数据库架构

## 核心元数据

* **版本**：v3.2 (Go+Rust 微服务重构版)
* **引擎**：PostgreSQL 16+ (Alpine)
* **设计哲学**：哑数据库，强应用，运维优先

## 架构概述

本架构专为 Go 与 Rust 微服务集群设计，采用清晰的 Snowflake ID 主键策略和发件箱模式处理分布式事务，摒弃了数据库触发器和存储过程。

## ER 核心关系图

```mermaid
erDiagram
    USERS {
        BIGINT id PK
        VARCHAR username
        VARCHAR email
        VARCHAR avatar_key
        INT version
        VARCHAR last_trace_id
    }
    CATEGORIES {
        BIGINT id PK
        VARCHAR slug
        INT post_count
    }
    TAGS {
        BIGINT id PK
        VARCHAR slug
        INT post_count
    }
    POSTS {
        BIGINT id PK
        VARCHAR slug
        TEXT raw_markdown
        TEXT html_body
        JSONB toc_json
        VARCHAR status
        INT version
        BIGINT author_id FK
        BIGINT category_id FK
    }
    POST_TAGS {
        BIGINT post_id PK,FK
        BIGINT tag_id PK,FK
    }
    COMMENTS {
        BIGINT id PK
        BIGINT post_id FK
        BIGINT root_id
        TEXT content
        VARCHAR status
    }
    USER_INTERACTIONS {
        BIGINT user_id PK,FK
        BIGINT post_id PK,FK
        VARCHAR type PK
    }
    OUTBOX_EVENTS {
        BIGINT id PK
        VARCHAR topic
        JSONB payload
        VARCHAR status
        VARCHAR locked_by
    }

    USERS ||--o{ POSTS : "撰写"
    CATEGORIES ||--o{ POSTS : "分类"
    POSTS ||--o{ POST_TAGS : "包含"
    TAGS ||--o{ POST_TAGS : "归属"
    USERS ||--o{ COMMENTS : "发表"
    POSTS ||--o{ COMMENTS : "收到"
    POSTS ||--o{ OUTBOX_EVENTS : "触发"
```

## 详细表结构说明

### users 表 (用户身份核心)

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PRIMARY KEY | Snowflake 算法生成的分布式唯一标识 |
| username | VARCHAR(32) | NOT NULL, 长度≥3 | 唯一用户名，用于登录和显示 |
| email | VARCHAR(255) | NOT NULL, 邮箱格式校验 | 用户邮箱地址，用于通知和找回密码 |
| password_hash | VARCHAR(255) | 条件约束 | Bcrypt 加密的密码哈希，OAuth用户可为空 |
| nickname | VARCHAR(64) | NULLABLE | 显示名称，默认为username |
| bio | TEXT | NULLABLE | 个人简介，支持Markdown格式 |
| avatar_key | VARCHAR(255) | NULLABLE | MinIO对象存储键，如`avatars/u123/me.jpg` |
| is_admin | BOOLEAN | NOT NULL DEFAULT FALSE | 管理员权限标志 |
| is_active | BOOLEAN | NOT NULL DEFAULT TRUE | 账户激活状态 |
| provider | auth_provider | NOT NULL DEFAULT 'local' | 认证提供商：local/github/google |
| provider_id | VARCHAR(255) | 条件约束 | OAuth提供商返回的用户标识 |
| version | INTEGER | NOT NULL DEFAULT 1 | 乐观锁版本号，防并发更新覆盖 |
| last_trace_id | VARCHAR(64) | NULLABLE | 最后操作链路追踪ID |
| meta | JSONB | DEFAULT '{}' | 扩展字段：主题偏好、UI配置等 |
| last_login_at | TIMESTAMPTZ | NULLABLE | 最后登录时间 |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() | 记录创建时间 |
| updated_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() | 最后更新时间 |
| deleted_at | TIMESTAMPTZ | NULLABLE | 软删除时间戳 |

### categories 表 (内容分类)

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PRIMARY KEY | Snowflake分布式唯一标识 |
| name | VARCHAR(64) | NOT NULL, UNIQUE | 分类显示名称 |
| slug | VARCHAR(64) | NOT NULL, UNIQUE | URL友好标识符 |
| description | TEXT | NULLABLE | 分类描述文本 |
| post_count | INTEGER | NOT NULL DEFAULT 0 | 文章数量统计（异步维护） |
| version | INTEGER | NOT NULL DEFAULT 1 | 乐观锁版本号 |
| meta | JSONB | DEFAULT '{}' | 扩展字段：图标、颜色、模板配置 |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() | 记录创建时间 |
| updated_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() | 最后更新时间 |

### tags 表 (内容标签)

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PRIMARY KEY | Snowflake分布式唯一标识 |
| name | VARCHAR(64) | NOT NULL, UNIQUE | 标签显示名称 |
| slug | VARCHAR(64) | NOT NULL, UNIQUE | URL友好标识符 |
| post_count | INTEGER | NOT NULL DEFAULT 0 | 文章数量统计（异步维护） |
| version | INTEGER | NOT NULL DEFAULT 1 | 乐观锁版本号 |
| meta | JSONB | DEFAULT '{}' | 扩展字段：图标、颜色等 |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() | 记录创建时间 |
| updated_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() | 最后更新时间 |

### posts 表 (文章核心)

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PRIMARY KEY | Snowflake分布式唯一标识 |
| title | VARCHAR(255) | NOT NULL | 文章标题 |
| slug | VARCHAR(255) | NOT NULL, UNIQUE | URL友好标识符 |
| summary | VARCHAR(500) | NULLABLE | 文章摘要（AI生成或手动填写） |
| resource_key | VARCHAR(255) | NULLABLE | MinIO资源文件夹路径 |
| cover_image_key | VARCHAR(255) | NULLABLE | 封面图存储路径 |
| raw_markdown | TEXT | NOT NULL | 原始Markdown内容（写模型） |
| html_body | TEXT | NULLABLE | 预渲染HTML内容（读模型） |
| toc_json | JSONB | NULLABLE | 目录结构JSON数据 |
| status | VARCHAR(20) | NOT NULL DEFAULT 'draft' | 状态：draft/published/archived |
| visibility | VARCHAR(20) | NOT NULL DEFAULT 'public' | 可见性：public/hidden/password |
| author_id | BIGINT | NOT NULL, FOREIGN KEY | 作者用户ID |
| category_id | BIGINT | NULLABLE, FOREIGN KEY | 分类ID |
| view_count | INTEGER | NOT NULL DEFAULT 0 | 浏览量统计 |
| like_count | INTEGER | NOT NULL DEFAULT 0 | 点赞量统计 |
| comment_count | INTEGER | NOT NULL DEFAULT 0 | 评论量统计 |
| version | INTEGER | NOT NULL DEFAULT 1 | 乐观锁版本号 |
| last_trace_id | VARCHAR(64) | NULLABLE | 最后操作链路追踪ID |
| created_by | BIGINT | NULLABLE | 记录创建人ID |
| updated_by | BIGINT | NULLABLE | 最后修改人ID |
| meta | JSONB | DEFAULT '{}' | 扩展字段：SEO关键词、PDF链接等 |
| published_at | TIMESTAMPTZ | NULLABLE | 发布时间 |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() | 记录创建时间 |
| updated_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() | 最后更新时间 |
| deleted_at | TIMESTAMPTZ | NULLABLE | 软删除时间戳 |

### post_tags 表 (文章-标签关联)

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| post_id | BIGINT | PRIMARY KEY, FOREIGN KEY | 文章ID |
| tag_id | BIGINT | PRIMARY KEY, FOREIGN KEY | 标签ID |
| created_at | TIMESTAMPTZ | DEFAULT NOW() | 关联创建时间 |

### comments 表 (评论系统)

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PRIMARY KEY | Snowflake分布式唯一标识 |
| post_id | BIGINT | NOT NULL, FOREIGN KEY | 所属文章ID |
| user_id | BIGINT | NOT NULL, FOREIGN KEY | 评论用户ID |
| parent_id | BIGINT | NULLABLE, FOREIGN KEY | 父评论ID（邻接表模型） |
| root_id | BIGINT | NULLABLE, FOREIGN KEY | 根评论ID（优化整楼查询） |
| content | TEXT | NOT NULL, 长度1-5000 | 评论内容文本 |
| status | VARCHAR(20) | NOT NULL DEFAULT 'pending' | 状态：pending/approved/spam |
| version | INTEGER | NOT NULL DEFAULT 1 | 乐观锁版本号 |
| last_trace_id | VARCHAR(64) | NULLABLE | 最后操作链路追踪ID |
| ip_address | VARCHAR(45) | NULLABLE | 评论来源IP（IPv6支持） |
| user_agent | TEXT | NULLABLE | 评论来源User-Agent |
| meta | JSONB | DEFAULT '{}' | 扩展字段：举报原因、AI检测分值等 |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() | 记录创建时间 |
| updated_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() | 最后更新时间 |
| deleted_at | TIMESTAMPTZ | NULLABLE | 软删除时间戳 |

### user_interactions 表 (用户互动)

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| user_id | BIGINT | PRIMARY KEY, FOREIGN KEY | 用户ID |
| post_id | BIGINT | PRIMARY KEY, FOREIGN KEY | 文章ID |
| interaction_type | VARCHAR(20) | PRIMARY KEY | 互动类型：like/bookmark/share |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() | 互动创建时间 |
| last_trace_id | VARCHAR(64) | NULLABLE | 操作链路追踪ID |

### ~~outbox_events 表~~ (已废弃 v3.2.1+)

**设计变更说明**：

自 v3.2.1 起，Bifrost 采用 **NATS Fire-and-Forget + Queue Groups** 替代 Outbox 发件箱模式，原因如下：

| 对比维度 | Outbox 模式 | Fire-and-Forget 模式 |
|--------|----------|-----------------|
| 一致性保障 | 强（事务级） | 弱（事件可能丢失） |
| 代码复杂度 | 高（需要 Relayer） | 低（直接 NATS 发送） |
| 数据库压力 | 高（要维护表） | 无（无需持久化） |
| 适用场景 | 金融级（账目对账） | 内容管理（缓存失效） |
| 运维成本 | 高 | 低 |

**为什么适合 CMS？**

* 对于内容管理系统，如果因为网络抖动导致 Redis 缓存没删掉（概率极低），后果仅仅是用户多看几分钟旧文章
* 而不是银行账目对不上 → **不需要引入复杂的事务保障**
* Bifrost 宁可"错杀"（删除不该删的缓存），也不放过（确保缓存一致）

**轻量级方案工作流**：

```
Nexus (业务完成后)
   ↓ go func() 异步发送
NATS (Fan-Out)
   ├→ Beacon (beacon_service 组) → 删除 Redis 缓存
   └→ Mirror (mirror_service 组) → 更新索引
```

详见 [轻量级消息传递方案](#轻量级消息传递方案-nats-fire-and-forget)

## 核心设计原则

### 1. 统一 Snowflake ID

* 全站使用 Twitter Snowflake 算法生成 int64 主键，替代 UUID
* 优势：索引更紧凑，URL友好，自带时间序，分布式生成无冲突

### 2. 内置运维防御

* **version字段**：乐观锁版本号，防止并发更新丢失
* **last_trace_id字段**：链路追踪锚点，便于快速定位问题请求
* **meta字段**：JSONB扩展字段，支持非结构化数据，减少DDL变更频率

### 3. 读写模型分离

* **写模型(raw_markdown)**：由Nexus服务操作，存储原始Markdown内容
* **读模型(html_body,toc_json)**：由Forge服务预渲染，Beacon服务直接读取，实现静态文件级性能

### 4. 轻量级消息传递方案：NATS Fire-and-Forget

替代 Outbox 发件箱模式，采用更简洁的事件驱动架构：

**核心机制**：

1. **Nexus (写服务)** - 业务完成后，通过 `go func()` 异步向 NATS 发送事件，不等待响应
2. **NATS** - 负责消息的扇出 (Fan-Out)，推送给订阅方
3. **Beacon (读服务)** - 订阅 `content.>` 主题，使用 Queue Group `beacon_service` 接收更新
4. **Mirror (搜索服务)** - 订阅 `content.>` 主题，使用 Queue Group `mirror_service` 接收更新

**关键设计要点**：

* **宽松的一致性保证**：事件可能因网络故障丢失，但这对 CMS 可接受（缓存多保留几分钟问题不大）
* **Queue Groups 自动负载均衡**：多个 Beacon 副本自动抢单消费，避免重复处理
* **零数据库压力**：不需要维护 Outbox 表，无 Relayer 协程轮询
* **易于扩展**：新增服务（如 Audit、Notification）只需创建新的 Queue Group，无需改动 Nexus

**事件主题约定**：

```
content.post.created       → 文章发布
content.post.updated       → 文章更新
content.post.deleted       → 文章删除
content.category.updated   → 分类变更
content.tag.updated        → 标签变更
content.comment.created    → 评论新增
content.interaction.*      → 用户互动
```

**消费者示例**：

```go
// Beacon 订阅示例
messenger.Subscribe("content.>", "beacon_service", func(subject, data []byte) {
    // 解析事件，删除相关 Redis 缓存
    // 例：content.post.updated → 删除 post:<id> 和 post:<slug> 缓存
})

// 未来 Rust Mirror 加入，只需新增
// messenger.QueueSubscribe("content.>", "mirror_service", handler)
// Nexus 代码无需任何改动！
```

## 数据库迁移脚本清单

| 顺序 | 文件名 | 核心职责 | 关键点 |
|------|--------|----------|------------|
| 01 | `001_identity_and_infra.sql` | 基础设施与用户表 | Snowflake ID、乐观锁、无 UUID |
| 02 | `002_content_core.sql` | 文章、分类、标签 | 预渲染字段、无触发器、版本控制 |
| 03 | `003_interactions.sql` | 评论、互动 | 邻接表 + 根节点优化、无 Outbox 表 |

**v3.2.1 变更**：删除 Outbox 表，改用 NATS Fire-and-Forget 方案

## 维护说明

* 严禁手动通过SQL直接更新 `post_count`、`view_count` 等统计字段
* 所有统计数据的变更必须通过 NATS 事件驱动相应的业务服务异步完成
* 乐观锁版本号必须在每次更新时自动递增，由应用层保证一致性
* 事件发送使用 `go func()` 异步方式，不阻塞主业务逻辑（Fire-and-Forget 原则）
* 缓存删除策略：宁可多删，不可少删（保证最终一致性）
