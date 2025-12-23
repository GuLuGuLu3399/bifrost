# Bifrost 项目测试需求规格书与实施指南

> **版本**：v1.0  
> **时间**：2025年12月23日  
> **状态**：P0 优先级 - 待实施  
> **目标**：补充测试覆盖率，化解零覆盖率风险

---

## 1. 测试目标

### 1.1 总体目标

- **核心业务逻辑覆盖率**：≥ 80%（UseCase 层）
- **数据操作覆盖率**：≥ 70%（Repository 层）
- **公共工具库覆盖率**：≥ 75%（internal/pkg）
- **安全检验覆盖率**：100%（认证、授权、错误处理链路）

### 1.2 质量红线（Gatekeeper）

| 指标 | 门槛 | 说明 |
|------|------|------|
| 整体覆盖率 | ≥ 60% | 优先保障关键路径 |
| 业务层覆盖率 | ≥ 80% | UseCase/Service 必须达到 |
| 错误路径覆盖 | 100% | 所有自定义错误必须有测试 |
| 集成测试数 | ≥ 30 | 覆盖完整业务流程 |
| CI 失败率 | 0% | 代码合并前必须通过所有测试 |

### 1.3 测试进度规划

| 阶段 | 工作量 | 周期 | 目标 |
|------|--------|------|------|
| **第 1 周** | 40% | 单元测试基础设施搭建 | 核心 UseCase 单元测试完成 |
| **第 2 周** | 35% | 集成测试补充 | Repo 层集成测试 + CQRS 链路测试 |
| **第 3 周** | 15% | E2E 测试 + 文档 | 端到端业务流程验证 |
| **第 4 周** | 10% | 回归 + 优化 | 提高覆盖率至 75%+ |

---

## 2. 分层测试策略

### 2.1 测试金字塔模型

```
                      /\
                     /  \          E2E Tests (5%)
                    /----\         - 完整业务流程
                   /      \        - Docker Compose 环境
                  /        \
                 /          \
                /------------\     Integration Tests (30%)
               /              \    - Repository 层 (testcontainers)
              /                \   - NATS 消息流 (mock)
             /                  \  - Cache 操作 (redis mock)
            /                    \
           /                      \
          /------------------------\ Unit Tests (65%)
         /                          \ - UseCase 层 (Repo Mock)
        /                            \ - Service 层 (逻辑验证)
       /                              \ - Data Model 映射
      /                                \
```

### 2.2 分层测试详细规划

#### 2.2.1 单元测试 (Unit Tests)

**覆盖范围**：业务逻辑、数据映射、参数校验

**重点模块**：

| 服务 | 模块 | 关键类 | 优先级 |
|------|------|--------|--------|
| **Nexus** | `biz` | `PostUseCase` | P0 |
| | | `UserUseCase` | P0 |
| | | `CommentUseCase` | P1 |
| **Beacon** | `data` | `PostRepo` | P0 |
| | | `CommentRepo` | P1 |
| **Gjallar** | `middleware` | Auth, Tracing | P1 |
| **Rust** | `forge` | `MarkdownEngine` | P0 |
| | `mirror` | `SearchEngine` | P1 |

**技术栈**：

```
Framework:    stretchr/testify
Assertion:    assert, require
Mocking:      testify/mock (for Repository/Service)
Helper:       Table-Driven Tests
```

**示例用例**：

```go
// go_services/internal/nexus/biz/post_usecase_test.go

func TestPostUseCase_CreatePost(t *testing.T) {
    tests := []struct {
        name      string
        input     *CreatePostInput
        repoMock  func() PostRepo           // Mock 函数
        expectErr bool
        expectOut *CreatePostOutput
    }{
        {
            name: "happy_path_draft",
            input: &CreatePostInput{
                Title:       "Test Post",
                Slug:        "test-post",
                RawMarkdown: "# Hello",
                CategoryID:  1,
                AuthorID:    100,
                Status:      contentv1.PostStatus_POST_STATUS_DRAFT,
            },
            repoMock: func() PostRepo {
                repo := new(MockPostRepo)
                repo.On("ExistsBySlug", mock.Anything, "test-post").Return(false, nil)
                repo.On("Create", mock.Anything, mock.MatchedBy(...)).Return(int64(999), nil)
                return repo
            },
            expectErr: false,
            expectOut: &CreatePostOutput{PostID: 999, Version: 1},
        },
        {
            name: "slug_conflict",
            input: &CreatePostInput{
                Slug: "duplicate-slug",
            },
            repoMock: func() PostRepo {
                repo := new(MockPostRepo)
                repo.On("ExistsBySlug", mock.Anything, "duplicate-slug").Return(true, nil)
                return repo
            },
            expectErr: true,
        },
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            mockRepo := tc.repoMock()
            uc := NewPostUseCase(mockRepo, mockTx, nil)
            
            output, err := uc.CreatePost(context.Background(), tc.input)
            
            if tc.expectErr {
                assert.Error(t, err)
            } else {
                require.NoError(t, err)
                assert.Equal(t, tc.expectOut.PostID, output.PostID)
            }
        })
    }
}
```

#### 2.2.2 集成测试 (Integration Tests)

**覆盖范围**：真实数据库交互、Redis 缓存操作、NATS 事件发布

**重点场景**：

| 场景 | 测试内容 | 容器需求 |
|------|---------|---------|
| Post 完整生命周期 | Create → Update → Delete（软删） | PostgreSQL |
| 乐观锁冲突处理 | 并发更新版本冲突检测 | PostgreSQL |
| 缓存一致性 | 写入触发缓存失效 | PostgreSQL + Redis |
| 事件流完整性 | 发布 → NATS → 消费 | PostgreSQL + NATS |

**技术栈**：

```txt
Container:    testcontainers-go
Database:     PostgreSQL 16 (真实实例)
Cache:        Redis (真实实例)
Message:      NATS (真实实例)
Migration:    预加载 migrations/*.sql
```

**示例用例**：见 2.4.1 代码片段

#### 2.2.3 端到端测试 (E2E Tests)

**覆盖范围**：跨服务调用、完整业务流程、网关响应

**关键路径**：

1. **文章发布流程**：

   ```
   Gjallar (HTTP) 
   → Nexus (gRPC, 创建)
   → Forge (gRPC, 渲染)
   → NATS (事件发布)
   → Beacon (消费)
   → Redis (缓存更新)
   ```

2. **评论互动流程**：

   ```
   Gjallar (HTTP)
   → Nexus (创建评论)
   → 验证文章存在性
   → 发布事件
   → Beacon 更新评论计数
   ```

**技术栈**：

```
Orchestration:    Docker Compose (完整环境)
Test Framework:   Go http client + gRPC client
Assertion:        testify/assert
Data Validation:  SQL 查询验证
```

**运行方式**：

```bash
# 启动全套环境
docker-compose -f docker-compose.test.yml up -d

# 运行 E2E 测试
go test -v -tags=e2e ./tests/e2e/...

# 清理环境
docker-compose -f docker-compose.test.yml down
```

---

## 3. 技术选型与依赖管理

### 3.1 Go 测试工具栈

| 工具 | 用途 | 版本 | 备注 |
|------|------|------|------|
| `testify/assert` | 断言 | v1.8+ | 已在 go.mod |
| `testify/mock` | Mock 对象 | v1.8+ | 已在 go.mod |
| `testify/require` | 前置条件检查 | v1.8+ | 已在 go.mod |
| `go-sqlmock` | SQL Mock | v1.5+ | **需新增** |
| `testcontainers-go` | 容器启动 | v0.20+ | **需新增** |

**新增依赖**：

```bash
cd go_services

# 添加测试依赖
go get -t github.com/DATA-DOG/go-sqlmock@v1.5.2
go get -t github.com/testcontainers/testcontainers-go@v0.31.0
go get -t github.com/testcontainers/testcontainers-go/wait@v0.31.0
```

### 3.2 Rust 测试工具栈

| 工具 | 用途 | 备注 |
|------|------|------|
| `#[test]` | 内置测试 | Rust 标准 |
| `mockall` | Mock trait | **需新增至 Cargo.toml** |
| `tokio::test` | 异步测试 | 已有 tokio |
| `testcontainers-rs` | 容器启动 | **需新增** |

**新增到 workspace.dependencies**（`rust_services/Cargo.toml`）：

```toml
[workspace.dependencies]
# ... 已有的依赖 ...

# 测试依赖
mockall = "0.12.0"
testcontainers = { version = "0.16.0", features = ["default"] }

[dev-dependencies]
mockall = { workspace = true }
testcontainers = { workspace = true }
tokio = { workspace = true, features = ["full"] }
```

### 3.3 版本约束

```
Go:   1.25+
Rust: 2021 edition
PostgreSQL: 16+ (for integration tests)
Redis: 7.0+ (for integration tests)
NATS: 2.10+ (for E2E tests)
```

---

## 4. 测试矩阵（功能点 × 测试类型）

### 4.1 Go 服务矩阵

| 功能 | 单元测试 | 集成测试 | E2E 测试 | 优先级 |
|------|---------|---------|---------|--------|
| **Nexus: 创建文章** | ✓ Mock Repo | ✓ 真实 DB | ✓ 完整链路 | P0 |
| **Nexus: 更新文章** | ✓ 乐观锁测试 | ✓ 版本冲突 | ✓ | P0 |
| **Nexus: 删除文章** | ✓ 软删除逻辑 | ✓ 级联删除 | ✓ | P0 |
| **Nexus: 创建评论** | ✓ 文章存在性检查 | ✓ 关联完整性 | ✓ | P0 |
| **Nexus: 用户注册** | ✓ 密码加密验证 | ✓ Email 唯一性 | ✓ | P0 |
| **Nexus: 用户登录** | ✓ JWT 生成 | ✓ 凭证验证 | ✓ | P0 |
| **Beacon: 查询文章** | ✓ Cache-Aside | ✓ Redis 缓存 | ✓ | P0 |
| **Beacon: 列表文章** | ✓ 分页逻辑 | ✓ 数据库查询 | ✓ | P0 |
| **Gjallar: 认证中间件** | ✓ Token 解析 | ✓ 黑名单检查 | ✓ | P1 |
| **Gjallar: 限流中间件** | ✓ Rate Limiter | ✓ Redis 分布式 | ✓ | P1 |

### 4.2 Rust 服务矩阵

| 功能 | 单元测试 | 集成测试 | 优先级 |
|------|---------|---------|--------|
| **Forge: Markdown 渲染** | ✓ XSS 防护 | ✓ 集成渲染 | P0 |
| **Forge: TOC 生成** | ✓ 标题提取 | ✓ | P0 |
| **Forge: 摘要生成** | ✓ 字符截断 | ✓ | P1 |
| **Mirror: 索引创建** | ✓ Schema 验证 | ✓ 真实 Tantivy | P1 |
| **Mirror: 全文搜索** | ✓ 查询解析 | ✓ 中文分词 | P1 |
| **Oracle: 数据聚合** | ✓ 结果合并 | ✓ | P2 |

---

## 5. CI/CD 集成计划

### 5.1 测试步骤流程

```yaml
# .github/workflows/test.yml (或 .gitlab-ci.yml)

name: Test Pipeline

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_PASSWORD: password
          POSTGRES_DB: bifrost_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432
          
      redis:
        image: redis:7-alpine
        options: >-
          --health-cmd "redis-cli ping"
        ports:
          - 6379:6379
          
      nats:
        image: nats:2.10-alpine
        ports:
          - 4222:4222

    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.25'
          
      - name: Set up Rust
        uses: actions-rs/toolchain@v1
        with:
          toolchain: stable
          
      # ===== Go 测试 =====
      - name: Run Go Unit Tests
        working-directory: go_services
        run: |
          go test -v -race -coverprofile=coverage.out ./...
          go tool cover -func coverage.out | tail -1
          
      - name: Upload Go Coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./go_services/coverage.out
          flags: go
          
      # ===== Rust 测试 =====
      - name: Run Rust Tests
        working-directory: rust_services
        run: |
          cargo test --all --verbose
          
      - name: Check Rust Code Coverage
        working-directory: rust_services
        run: |
          cargo tarpaulin --all --out Xml --timeout 300
          
      - name: Upload Rust Coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./rust_services/cobertura.xml
          flags: rust
          
      # ===== 测试覆盖率门槛检查 =====
      - name: Check Coverage Threshold (Go)
        working-directory: go_services
        run: |
          COVERAGE=$(go tool cover -func coverage.out | tail -1 | awk '{print $3}' | sed 's/%//')
          if (( $(echo "$COVERAGE < 60" | bc -l) )); then
            echo "❌ Go coverage is ${COVERAGE}%, below 60% threshold"
            exit 1
          fi
          echo "✅ Go coverage is ${COVERAGE}%"
          
      - name: Lint Tests (Go)
        working-directory: go_services
        run: |
          go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
          golangci-lint run --timeout 5m
          
  e2e:
    runs-on: ubuntu-latest
    needs: test  # 只有单元测试通过才运行 E2E
    
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.25'
          
      - name: Start Docker Compose Stack
        run: docker-compose -f docker-compose.test.yml up -d
        
      - name: Wait for Services
        run: sleep 10
        
      - name: Run E2E Tests
        working-directory: tests/e2e
        run: go test -v -tags=e2e ./...
        
      - name: Collect Logs on Failure
        if: failure()
        run: docker-compose -f docker-compose.test.yml logs
        
      - name: Cleanup
        if: always()
        run: docker-compose -f docker-compose.test.yml down -v
```

### 5.2 本地开发流程

```bash
# 开发者在提交前运行
cd go_services

# 1. 运行单元测试 (快速反馈)
go test -v -short ./...

# 2. 运行覆盖率检查
go test -cover ./...

# 3. 运行集成测试 (需要 Docker)
go test -v -run Integration ./...

# 4. 完整测试 (含 E2E)
go test -v ./... && docker-compose up -d && go test -tags=e2e ./tests/e2e
```

### 5.3 覆盖率门槛检查脚本

```bash
#!/bin/bash
# scripts/check_coverage.sh

set -e

echo "📊 Checking test coverage thresholds..."

# Go coverage
cd go_services
go test -coverprofile=coverage.out ./...
COVERAGE=$(go tool cover -func coverage.out | tail -1 | awk '{print $3}' | sed 's/%//')
THRESHOLD=60

if (( $(echo "$COVERAGE < $THRESHOLD" | bc -l) )); then
    echo "❌ Go coverage is ${COVERAGE}%, below ${THRESHOLD}% threshold"
    exit 1
else
    echo "✅ Go coverage is ${COVERAGE}% (threshold: ${THRESHOLD}%)"
fi

# Rust coverage
cd ../rust_services
cargo tarpaulin --all --timeout 300 --out Stdout | grep -oP 'Coverage: \K[0-9.]+' > /tmp/rust_coverage.txt
RUST_COVERAGE=$(cat /tmp/rust_coverage.txt | sed 's/%//')

if (( $(echo "$RUST_COVERAGE < 50" | bc -l) )); then
    echo "❌ Rust coverage is ${RUST_COVERAGE}%, below 50% threshold"
    exit 1
else
    echo "✅ Rust coverage is ${RUST_COVERAGE}%"
fi

echo "✅ All coverage checks passed!"
```

---

## 6. 实施示例

### 6.1 Go 集成测试示例：Repo 层 (testcontainers-go)

**文件**：`go_services/internal/nexus/data/post_repo_integration_test.go`

```go
//go:build integration
// +build integration

package data

import (
 "context"
 "testing"
 "time"

 "github.com/stretchr/testify/assert"
 "github.com/stretchr/testify/require"
 "github.com/testcontainers/testcontainers-go"
 "github.com/testcontainers/testcontainers-go/wait"
 "github.com/jmoiron/sqlx"
 _ "github.com/jackc/pgx/v5/stdlib"

 contentv1 "github.com/gulugulu3399/bifrost/api/content/v1"
 "github.com/gulugulu3399/bifrost/internal/pkg/database"
)

// setupPostgresContainer 启动 PostgreSQL 容器并返回数据库连接
func setupPostgresContainer(t *testing.T) (*database.DB, func()) {
 t.Helper()

 ctx := context.Background()

 // 创建 PostgreSQL 容器请求
 req := testcontainers.ContainerRequest{
  Image:        "postgres:16-alpine",
  ExposedPorts: []string{"5432/tcp"},
  Env: map[string]string{
   "POSTGRES_USER":     "testuser",
   "POSTGRES_PASSWORD": "testpass",
   "POSTGRES_DB":       "bifrost_test",
  },
  WaitingFor: wait.ForLog("database system is ready to accept connections").
   WithOccurrence(2).
   WithStartupTimeout(10 * time.Second),
 }

 container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
  ContainerRequest: req,
  Started:          true,
 })
 require.NoError(t, err, "failed to start postgres container")

 // 获取连接字符串
 host, err := container.Host(ctx)
 require.NoError(t, err)

 port, err := container.MappedPort(ctx, "5432")
 require.NoError(t, err)

 dsn := "postgres://testuser:testpass@" + host + ":" + port.Port() + "/bifrost_test?sslmode=disable"

 // 建立数据库连接
 sqlxDB, err := sqlx.ConnectContext(ctx, "pgx", dsn)
 require.NoError(t, err, "failed to connect to postgres")

 db := &database.DB{DB: sqlxDB}

 // 运行迁移脚本
 setupDatabase(t, db)

 // 返回清理函数
 cleanup := func() {
  _ = sqlxDB.Close()
  _ = container.Terminate(ctx)
 }

 return db, cleanup
}

// setupDatabase 加载 SQL 迁移脚本
func setupDatabase(t *testing.T, db *database.DB) {
 t.Helper()

 migrations := []string{
  // 001_identity_and_infra.sql
  `
  CREATE TABLE IF NOT EXISTS users (
   id BIGINT PRIMARY KEY,
   username VARCHAR(100) UNIQUE NOT NULL,
   email VARCHAR(255) UNIQUE NOT NULL,
   password_hash VARCHAR(255) NOT NULL,
   created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
   updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
  );
  `,
  // 002_content_core.sql
  `
  CREATE TABLE IF NOT EXISTS posts (
   id BIGINT PRIMARY KEY,
   title VARCHAR(255) NOT NULL,
   slug VARCHAR(255) UNIQUE NOT NULL,
   raw_markdown TEXT,
   html_body TEXT,
   category_id BIGINT,
   author_id BIGINT NOT NULL REFERENCES users(id),
   status VARCHAR(50) NOT NULL DEFAULT 'DRAFT',
   visibility VARCHAR(50) NOT NULL DEFAULT 'PUBLIC',
   version BIGINT NOT NULL DEFAULT 1,
   published_at TIMESTAMP,
   created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
   updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
  );
  CREATE INDEX IF NOT EXISTS idx_posts_slug ON posts(slug);
  CREATE INDEX IF NOT EXISTS idx_posts_author_id ON posts(author_id);
  `,
 }

 ctx := context.Background()
 for _, migration := range migrations {
  _, err := db.ExecContext(ctx, migration)
  require.NoError(t, err, "failed to run migration")
 }
}

// ============================================================
// 测试用例
// ============================================================

func TestPostRepo_Create_Integration(t *testing.T) {
 if testing.Short() {
  t.Skip("Skipping integration test in short mode")
 }

 db, cleanup := setupPostgresContainer(t)
 defer cleanup()

 // 预先创建用户（外键约束）
 ctx := context.Background()
 authorID := int64(100)
 _, err := db.ExecContext(ctx, `
  INSERT INTO users (id, username, email, password_hash)
  VALUES ($1, 'testuser', 'test@example.com', 'hash')
 `, authorID)
 require.NoError(t, err)

 // 初始化 Repo
 repo := NewPostRepo(NewData(db, nil))

 tests := []struct {
  name      string
  post      *Post
  expectErr bool
  checkFn   func(t *testing.T, id int64, err error)
 }{
  {
   name: "create_draft_post",
   post: &Post{
    Title:       "Test Post",
    Slug:        "test-post-1",
    RawMarkdown: "# Hello World",
    Status:      contentv1.PostStatus_POST_STATUS_DRAFT,
    AuthorID:    authorID,
    Version:     1,
    CreatedAt:   time.Now(),
    UpdatedAt:   time.Now(),
   },
   expectErr: false,
   checkFn: func(t *testing.T, id int64, err error) {
    require.NoError(t, err)
    assert.Greater(t, id, int64(0), "post ID should be generated")

    // 验证数据库中的数据
    var dbPost Post
    err = db.GetContext(ctx, &dbPost, `
     SELECT id, title, slug, status, version FROM posts WHERE id = $1
    `, id)
    require.NoError(t, err)
    assert.Equal(t, "Test Post", dbPost.Title)
    assert.Equal(t, "test-post-1", dbPost.Slug)
    assert.Equal(t, int64(1), dbPost.Version)
   },
  },
  {
   name: "slug_uniqueness_violation",
   post: &Post{
    Title:       "Duplicate",
    Slug:        "test-post-1", // 重复的 slug
    Status:      contentv1.PostStatus_POST_STATUS_DRAFT,
    AuthorID:    authorID,
    Version:     1,
    CreatedAt:   time.Now(),
    UpdatedAt:   time.Now(),
   },
   expectErr: true,
   checkFn: func(t *testing.T, id int64, err error) {
    assert.Error(t, err, "should fail due to slug uniqueness constraint")
   },
  },
 }

 for _, tc := range tests {
  t.Run(tc.name, func(t *testing.T) {
   // 为每个测试创建新事务（数据隔离）
   id, err := repo.Create(ctx, tc.post)

   if tc.expectErr {
    assert.Error(t, err)
   } else {
    require.NoError(t, err)
   }

   tc.checkFn(t, id, err)
  })
 }
}

func TestPostRepo_Update_Version_Conflict(t *testing.T) {
 if testing.Short() {
  t.Skip("Skipping integration test in short mode")
 }

 db, cleanup := setupPostgresContainer(t)
 defer cleanup()

 ctx := context.Background()

 // 创建测试数据
 authorID := int64(101)
 _, err := db.ExecContext(ctx, `
  INSERT INTO users (id, username, email, password_hash)
  VALUES ($1, 'testuser2', 'test2@example.com', 'hash')
 `, authorID)
 require.NoError(t, err)

 // 创建一篇文章
 postID := int64(1001)
 _, err = db.ExecContext(ctx, `
  INSERT INTO posts (id, title, slug, author_id, status, version, created_at, updated_at)
  VALUES ($1, 'Original', 'orig-slug', $2, 'DRAFT', 1, NOW(), NOW())
 `, postID, authorID)
 require.NoError(t, err)

 repo := NewPostRepo(NewData(db, nil))

 // 场景：两个并发请求都想更新 version 1 的文章

 // 第一个请求成功
 post1 := &Post{
  ID:        postID,
  Title:     "Updated by User 1",
  Slug:      "orig-slug",
  Status:    contentv1.PostStatus_POST_STATUS_DRAFT,
  AuthorID:  authorID,
  Version:   1, // 乐观锁版本
  UpdatedAt: time.Now(),
 }

 err1 := repo.Update(ctx, post1)
 require.NoError(t, err1)

 // 验证版本已递增
 var newVersion int64
 err = db.GetContext(ctx, &newVersion, "SELECT version FROM posts WHERE id = $1", postID)
 require.NoError(t, err)
 assert.Equal(t, int64(2), newVersion)

 // 第二个请求尝试更新 version 1（已过期）
 post2 := &Post{
  ID:        postID,
  Title:     "Updated by User 2",
  Slug:      "orig-slug",
  Status:    contentv1.PostStatus_POST_STATUS_DRAFT,
  AuthorID:  authorID,
  Version:   1, // 过期版本
  UpdatedAt: time.Now(),
 }

 err2 := repo.Update(ctx, post2)
 assert.Error(t, err2, "should fail with version conflict")
}
```

**运行方式**：

```bash
# 运行集成测试
cd go_services
go test -v -tags=integration ./internal/nexus/data/...

# 或仅运行特定测试
go test -v -run TestPostRepo_Create_Integration -tags=integration ./internal/nexus/data/...
```

---

### 6.2 Rust 单元测试示例：MarkdownEngine

**文件**：`rust_services/forge/src/engine.rs` (增加测试模块)

```rust
// rust_services/forge/src/engine.rs

// ... 现有代码 ...

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_render_basic_markdown() {
        let engine = MarkdownEngine::new();
        let markdown = "# Hello\n\nThis is a paragraph.";

        let result = engine.render(markdown);
        assert!(result.is_ok(), "render should succeed");

        let (html, toc, summary) = result.unwrap();

        // 验证 HTML 输出
        assert!(html.contains("<h1"));
        assert!(html.contains("Hello"));
        assert!(html.contains("This is a paragraph"));

        // 验证 TOC
        let toc_entries: Vec<TocEntry> = serde_json::from_str(&toc).unwrap();
        assert_eq!(toc_entries.len(), 1);
        assert_eq!(toc_entries[0].level, 1);
        assert_eq!(toc_entries[0].title, "Hello");

        // 验证摘要
        assert!(!summary.is_empty());
        assert!(summary.contains("paragraph"));
    }

    #[test]
    fn test_render_with_multiple_headings() {
        let engine = MarkdownEngine::new();
        let markdown = "# H1\n## H2\n### H3\n## H2 Again";

        let (html, toc, _) = engine.render(markdown).unwrap();

        let toc_entries: Vec<TocEntry> = serde_json::from_str(&toc).unwrap();
        assert_eq!(toc_entries.len(), 4);
        assert_eq!(toc_entries[0].level, 1);
        assert_eq!(toc_entries[1].level, 2);
        assert_eq!(toc_entries[2].level, 3);
        assert_eq!(toc_entries[3].level, 2);
    }

    #[test]
    fn test_xss_prevention_strips_scripts() {
        let engine = MarkdownEngine::new();
        let dangerous = r#"
# Title

<script>alert('xss')</script>

[Link](javascript:void(0))
        "#;

        let (html, _, _) = engine.render(dangerous).unwrap();

        // 验证脚本标签被清除
        assert!(!html.contains("<script>"), "script tags should be removed");
        assert!(!html.contains("alert"), "script content should not appear");

        // 验证标题正常渲染
        assert!(html.contains("<h1"));
        assert!(html.contains("Title"));
    }

    #[test]
    fn test_xss_prevention_strips_event_handlers() {
        let engine = MarkdownEngine::new();
        let dangerous = r#"<img src="x" onerror="alert('xss')" />"#;

        let (html, _, _) = engine.render(dangerous).unwrap();

        // 验证事件处理器被清除
        assert!(!html.contains("onerror"), "event handlers should be removed");
    }

    #[test]
    fn test_summary_truncation() {
        let engine = MarkdownEngine::new();

        // 生成超过 200 字符的内容
        let long_text = "A".repeat(300);
        let markdown = format!("# Header\n\n{}", long_text);

        let (_, _, summary) = engine.render(&markdown).unwrap();

        // 验证摘要被截断
        assert!(summary.len() <= 250, "summary should be truncated"); // 允许一些余量
    }

    #[test]
    fn test_render_with_tables() {
        let engine = MarkdownEngine::new();
        let markdown = r#"
| Header 1 | Header 2 |
|----------|----------|
| Cell 1   | Cell 2   |
        "#;

        let (html, _, _) = engine.render(markdown).unwrap();

        assert!(html.contains("<table"), "should render table");
        assert!(html.contains("<thead"), "should have table head");
        assert!(html.contains("<tbody"), "should have table body");
    }

    #[test]
    fn test_render_with_code_blocks() {
        let engine = MarkdownEngine::new();
        let markdown = r#"
```rust
fn main() {
    println!("Hello");
}
```

        "#;

        let (html, _, _) = engine.render(markdown).unwrap();

        assert!(html.contains("<pre"), "should render code block");
        assert!(html.contains("<code"), "should have code tag");
        assert!(html.contains("println!"), "code content should be preserved");
    }

    #[test]
    fn test_summary_collection_multiline() {
        let engine = MarkdownEngine::new();
        let markdown = "# Title\n\nFirst line.\nSecond line.\nThird line.";

        let (_, _, summary) = engine.render(markdown).unwrap();

        // 摘要应该包含多行文本
        assert!(summary.contains("First line"));
        assert!(summary.contains("Second line"));
    }

    #[test]
    fn test_anchor_generation_from_headings() {
        let engine = MarkdownEngine::new();
        let markdown = "# Getting Started\n## Installation Steps";

        let (_, toc, _) = engine.render(markdown).unwrap();

        let toc_entries: Vec<TocEntry> = serde_json::from_str(&toc).unwrap();

        // 验证 anchor 被正确生成（slug 化）
        assert_eq!(toc_entries[0].anchor, "getting-started");
        assert_eq!(toc_entries[1].anchor, "installation-steps");
    }

    #[test]
    fn test_render_empty_markdown() {
        let engine = MarkdownEngine::new();
        let markdown = "";

        let (html, toc, summary) = engine.render(markdown).unwrap();

        assert!(html.is_empty() || html == "<p></p>\n");
        assert_eq!(toc, "[]");
        assert!(summary.is_empty());
    }

    #[test]
    fn test_render_only_whitespace() {
        let engine = MarkdownEngine::new();
        let markdown = "   \n\n  \n";

        let result = engine.render(markdown);
        assert!(result.is_ok());

        let (html, toc, _) = result.unwrap();
        assert_eq!(toc, "[]", "no headings should be found");
    }
}

```

**运行方式**：

```bash
# 运行 Forge 的所有测试
cd rust_services/forge
cargo test --lib engine

# 仅运行 XSS 防护测试
cargo test --lib engine::tests::test_xss_prevention

# 显示详细输出（包括 println!）
cargo test --lib engine -- --nocapture
```

---

## 7. Mock 对象生成指南

### 7.1 Go Mock 示例 (testify/mock)

如果使用 `mockgen` 自动生成 Mock 类，需安装工具：

```bash
go install github.com/golang/mock/cmd/mockgen@latest
```

**手动创建 Mock**：

```go
// go_services/internal/nexus/biz/post_usecase_test.go

import (
    "github.com/stretchr/testify/mock"
)

// MockPostRepo 是 PostRepo 接口的 Mock 实现
type MockPostRepo struct {
    mock.Mock
}

func (m *MockPostRepo) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
    args := m.Called(ctx, slug)
    return args.Bool(0), args.Error(1)
}

func (m *MockPostRepo) Create(ctx context.Context, post *Post) (int64, error) {
    args := m.Called(ctx, post)
    return args.Get(0).(int64), args.Error(1)
}

// 使用示例
func TestPostUseCase_CreatePost_MockRepo(t *testing.T) {
    mockRepo := new(MockPostRepo)
    
    // 设置预期调用
    mockRepo.On("ExistsBySlug", mock.Anything, "new-slug").Return(false, nil)
    mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(p *Post) bool {
        return p.Title == "Test"
    })).Return(int64(123), nil)
    
    // 创建 UseCase
    uc := NewPostUseCase(mockRepo, mockTx, nil)
    
    // 执行
    output, err := uc.CreatePost(context.Background(), &CreatePostInput{
        Title: "Test",
        Slug:  "new-slug",
    })
    
    // 验证
    require.NoError(t, err)
    assert.Equal(t, int64(123), output.PostID)
    mockRepo.AssertExpectations(t)
}
```

### 7.2 Rust Mock 示例 (mockall)

```rust
// rust_services/mirror/src/engine_test.rs

#[cfg(test)]
mod tests {
    use super::*;
    use mockall::predicate::*;

    #[tokio::test]
    async fn test_search_with_mock_index() {
        // mockall 会自动为 trait 生成 Mock 类型
        // 例如，如果有 trait IndexReader，则生成 MockIndexReader

        // 设置 Mock 行为
        let mut mock_reader = MockIndexReader::new();

        mock_reader
            .expect_search()
            .with(eq("test query"), always())
            .times(1)
            .returning(|_, _| Ok(vec![]));

        // 将 Mock 传入被测函数
        // let result = search(&mock_reader, "test query").await;
        // assert!(result.is_ok());
    }
}
```

---

## 8. 常见问题与排查

### 8.1 问题排查表

| 问题 | 原因 | 解决方案 |
|------|------|--------|
| `testcontainers` 启动超时 | Docker daemon 未运行 | `docker ps` 确保 Docker 启动 |
| sqlmock 期望不匹配 | SQL 字符串不完全相同（空格/参数） | 使用 `sqlmock.WithoutQuoteMatcher` 或 `MatcherFunc` |
| Go 测试中 context timeout | 数据库操作过慢 | 增大 timeout 或检查索引 |
| Rust 异步测试 panic | 没有使用 `#[tokio::test]` | 确保标注 `#[tokio::test]` |
| 覆盖率统计不准 | 测试构建标签未正确配置 | 使用 `// +build !test_skip` 标注 |

### 8.2 调试技巧

**Go**：

```bash
# 显示详细 SQL 日志
export LOG_LEVEL=debug
go test -v -run TestName ./...

# 保留容器便于检查
go test -v -run TestName ./... # 在测试失败后，容器仍在运行

# 进入容器调查
docker exec -it <container_id> psql -U testuser -d bifrost_test
```

**Rust**：

```bash
# 显示日志输出
RUST_LOG=debug cargo test -- --nocapture

# 单线程运行（便于日志阅读）
cargo test -- --test-threads=1 --nocapture
```

---

## 9. 成功标准与验收清单

### 9.1 本周期验收标准

- [ ] **Go 单元测试**：核心 UseCase 覆盖率 ≥ 80%
  - [ ] `PostUseCase` (创建、更新、删除)
  - [ ] `UserUseCase` (注册、登录、密码变更)
  - [ ] `CommentUseCase` (创建、删除)
  
- [ ] **Go 集成测试**：Repo 层 ≥ 70% 覆盖率
  - [ ] `PostRepo.Create`, `Update`, `Delete`
  - [ ] `UserRepo.Create`, `GetByEmail`
  - [ ] 乐观锁版本冲突处理

- [ ] **Rust 单元测试**：核心逻辑 100% 覆盖
  - [ ] `MarkdownEngine.render` (正常 + XSS)
  - [ ] `MarkdownEngine.summary` (截断验证)
  
- [ ] **测试基础设施**：
  - [ ] `Makefile` 或 `scripts/test.sh` for 便捷执行
  - [ ] CI/CD 集成：测试失败阻止合并
  - [ ] 覆盖率报告自动生成

### 9.2 质量门槛

```
✅ 所有测试通过（无 flaky 测试）
✅ 覆盖率 ≥ 60% (Go) / ≥ 50% (Rust)
✅ 没有 hardcoded mock 数据
✅ Mock 接口与真实接口同步
✅ 集成测试使用真实数据库连接
```

---

## 10. 实施时间表

| 周数 | 任务 | 交付物 | 负责人 |
|------|------|--------|--------|
| **第 1 周** | 搭建测试框架，编写 Nexus PostUseCase 单元测试 | post_usecase_test.go (60+ 行) | @dev |
| | 编写 Rust MarkdownEngine 单元测试 | engine_test.rs (150+ 行) | @rust-dev |
| **第 2 周** | 编写 Repo 层集成测试（testcontainers） | *_integration_test.go (200+ 行) | @dev |
| | 配置 CI/CD 流程 | GitHub Actions YAML | @devops |
| **第 3 周** | 补充 E2E 测试（CQRS 链路） | tests/e2e/ | @dev |
| | 完善 Mock 代码生成脚本 | Makefile | @dev |
| **第 4 周** | 性能调优、flaky 测试修复 | 测试报告 | @qa |
| | 文档完善、onboarding 培训 | TEST_GUIDE.md | @architect |

---

## 11. 附录：文件清单

创建以下新文件和目录：

```
go_services/
├── Makefile                          # 新增：测试快捷命令
├── internal/
│   ├── nexus/
│   │   ├── biz/
│   │   │   ├── post_usecase_test.go          # 新增
│   │   │   ├── user_usecase_test.go          # 新增
│   │   │   └── comment_usecase_test.go       # 新增
│   │   └── data/
│   │       ├── post_repo_integration_test.go # 新增
│   │       └── user_repo_integration_test.go # 新增
│   └── beacon/
│       └── data/
│           └── post_repo_integration_test.go # 新增
└── scripts/
    └── check_coverage.sh            # 新增：覆盖率门槛脚本

rust_services/
├── forge/
│   └── src/
│       ├── engine_test.rs           # 新增测试模块
│       └── server_test.rs           # 新增
├── mirror/
│   └── src/
│       └── engine_test.rs           # 新增
└── Cargo.toml                        # 修改：添加 mockall, testcontainers

.github/
└── workflows/
    └── test.yml                     # 新增：CI/CD 配置
```

---

## 12. 参考资源

- [testify 官方文档](https://pkg.go.dev/github.com/stretchr/testify)
- [go-sqlmock 使用指南](https://github.com/DATA-DOG/go-sqlmock)
- [testcontainers-go](https://golang.testcontainers.org/)
- [Rust mockall](https://docs.rs/mockall/)
- [Table-Driven Tests in Go](https://dave.cheney.net/2019/05/07/patterns-for-composite-errors)

---

**文档版本控制**：

- v1.0 - 初版，包含完整策略和代码示例
- 更新日期：2025-12-23
