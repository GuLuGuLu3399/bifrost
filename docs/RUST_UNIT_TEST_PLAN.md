# Bifrost Rust 单元测试建设与实施计划书

> **版本**：v1.0  
> **日期**：2025年12月23日  
> **适用范围**：Forge, Mirror, Oracle 服务  
> **目标**：消除零覆盖率风险、消除 Panic 隐患、确保核心算法正确性

---

## 1. 测试目标

### 1.1 覆盖率指标

| 模块 | 覆盖率目标 | 优先级 | 说明 |
|------|-----------|--------|------|
| **Forge** (MarkdownEngine) | ≥ 90% | P0 | 核心渲染逻辑，关乎数据安全 (XSS) |
| **Mirror** (SearchEngine) | ≥ 80% | P0 | 搜索索引构建，数据完整性 |
| **Oracle** (AggregationService) | ≥ 70% | P1 | 搜索聚合逻辑 |
| **Common** (Config, Logger, OTEL) | ≥ 75% | P1 | 基础设施代码 |
| **整体** | ≥ 75% | 门槛 | 全仓库覆盖率不低于此线 |

### 1.2 "零 Panic" 目标

**当前隐患分析**：

根据代码审计，Rust 代码中存在以下 panic 风险：

```
rust_services/forge/src/main.rs:28          .expect("Server config missing")
rust_services/oracle/src/main.rs:41         .unwrap()
rust_services/common/src/nats/events.rs:91-92  两次 unwrap()
```

**目标**：

- [ ] 消除所有业务逻辑中的 `unwrap()` / `expect()` 调用
- [ ] 所有配置加载失败返回有意义的 `anyhow::Error`，不 panic
- [ ] 所有 JSON 序列化失败由中间件捕获，返回 HTTP 5xx 而非 panic
- [ ] 所有数据库查询失败返回 `Result<T, Error>`，不 panic

### 1.3 测试方法论

采用**金字塔模型**：

```
        ╱╲
       ╱  ╲         Unit Tests: 单个函数、算法验证
      ╱────╲        - MarkdownEngine::render() 的各种输入
     ╱      ╲       - Tokenizer 的分词结果
    ╱────────╲      - Index 的构建逻辑
   ╱          ╲
  ╱────────────╲    Integration Tests: 跨模块交互
 ╱              ╲   - 完整的 Markdown 渲染链路（包括 XSS 清理）
╱────────────────╲  - 索引构建 + 搜索查询的往返
```

**质量门槛**：

```
✅ 所有核心算法的 Happy Path + 3 个 Edge Case
✅ 所有 Result<T, E> 都有错误路径测试
✅ 所有 Option<T> 都有 None 路径测试
✅ 没有 #[should_panic] 测试（应该验证 Err，不应该测试 panic）
```

---

## 2. 工具链规范

### 2.1 Cargo.toml 配置

#### 2.1.1 工作区级别配置

**文件**：`rust_services/Cargo.toml`

```toml
[workspace]
resolver = "2"
members = [
    "common",
    "forge",
    "mirror",
    "oracle",
]

[workspace.package]
version = "3.2.0"
edition = "2021"
authors = ["Bifrost Team"]
license = "MIT"

# =====================================================================
# [新增] 测试和开发依赖统一管理
# =====================================================================
[workspace.dependencies]
# 核心依赖（保持现有）
tokio = { version = "1.48.0", features = ["full"] }
tonic = "0.14.2"
prost = "0.14.1"
prost-types = "0.14.1"
futures = "0.3.31"
serde = { version = "1.0.228", features = ["derive"] }
serde_json = "1.0.146"
anyhow = "1.0.100"
thiserror = "2.0.17"
tracing = "0.1.44"
tracing-subscriber = { version = "0.3.22", features = ["env-filter"] }
pulldown-cmark = { version = "0.13.0", default-features = false, features = ["simd", "html"] }
ammonia = "4.1.2"
slug = "0.1.6"
tantivy = "0.25.0"
tonic-build = "0.14.2"
opentelemetry = "0.31.0"
opentelemetry_sdk = { version = "0.31.0", features = ["rt-tokio"] }
opentelemetry-otlp = { version = "0.31.0", features = ["grpc-tonic"] }
opentelemetry-semantic-conventions = "0.31.0"
tracing-opentelemetry = "0.32.0"
http = "1.4.0"

# =====================================================================
# [新增] 测试依赖
# =====================================================================
mockall = "0.12.0"
pretty_assertions = "1.4.0"
proptest = "1.4.0"  # 属性测试框架，用于模糊测试
tokio-test = "0.4.1"
tempfile = "3.10.0"  # 临时文件，用于集成测试
common = { path = "./common" }
```

#### 2.1.2 单个包级别配置

**文件示例**：`rust_services/forge/Cargo.toml`

```toml
[package]
name = "forge"
version.workspace = true
edition.workspace = true
authors.workspace = true
license.workspace = true

[[bin]]
name = "forge"
path = "src/main.rs"

[dependencies]
common = { workspace = true }
tokio = { workspace = true }
tonic = { workspace = true }
prost = { workspace = true }
tracing = { workspace = true }
tracing-subscriber = { workspace = true }
serde = { workspace = true }
serde_json = { workspace = true }
anyhow = { workspace = true }

# Forge 特有依赖
pulldown-cmark = { workspace = true }
ammonia = { workspace = true }
slug = { workspace = true }

[build-dependencies]
tonic-build = { workspace = true }

# =====================================================================
# [新增] 开发依赖
# =====================================================================
[dev-dependencies]
mockall = { workspace = true }
pretty_assertions = { workspace = true }
proptest = { workspace = true }
tokio-test = { workspace = true }
tempfile = { workspace = true }
```

**Mirror 和 Oracle 类似配置**（只需修改依赖名称）。

### 2.2 测试配置文件

#### 2.2.1 Cargo.toml 中的测试配置

```toml
# 文件：rust_services/Cargo.toml

# 定义测试 Profile，优化测试编译时间
[profile.test]
opt-level = 1  # 启用基础优化，加速测试但保留调试信息

[profile.test.package."*"]
opt-level = 3  # 依赖包使用完全优化

# 集成测试配置（若需要独立的 tests/ 目录）
[[test]]
name = "integration"
path = "tests/integration_tests.rs"
```

### 2.3 测试运行命令速查表

```bash
# 1. 运行所有测试
cargo test --all

# 2. 运行单个包的测试
cargo test -p forge --lib

# 3. 运行特定测试函数
cargo test -p forge engine::tests::test_render_basic

# 4. 显示输出（包括 println! 和 eprintln!）
cargo test -- --nocapture

# 5. 单线程运行（便于查看日志顺序）
cargo test -- --test-threads=1

# 6. 生成覆盖率报告（使用 tarpaulin）
cargo install cargo-tarpaulin
cargo tarpaulin --all --out Html --output-dir coverage

# 7. 检查 unwrap/expect 等风险调用
cargo clippy --all -- -W clippy::unwrap_used -W clippy::expect_used

# 8. 运行测试，失败时停止
cargo test -- --test-threads=1 --fail-fast
```

---

## 3. 测试范围与策略

### 3.1 Forge 服务 (Markdown 渲染)

#### 3.1.1 核心模块

| 模块 | 函数/结构 | 测试重点 | 优先级 |
|------|----------|---------|--------|
| **engine** | `MarkdownEngine::render()` | 渲染正确性、XSS 防护 | P0 |
| | `MarkdownEngine::new()` | 初始化、Sanitizer 配置 | P0 |
| | TOC 生成 | 标题提取、anchor 生成、层级 | P0 |
| | 摘要生成 | 字符截断、多行文本收集 | P0 |

#### 3.1.2 测试用例设计

**Happy Path**：

```rust
#[test]
fn test_render_basic_markdown() {
    // 验证基础 Markdown 渲染（标题、段落、加粗、斜体）
    // 预期输出：有效的 HTML，无 XSS 内容
}

#[test]
fn test_render_with_code_blocks() {
    // 验证代码块渲染（```rust 语法高亮）
}

#[test]
fn test_render_with_tables() {
    // 验证表格渲染（markdown 表格格式）
}
```

**Edge Cases / Error Scenarios**：

```rust
#[test]
fn test_xss_prevention_script_tags() {
    // 恶意输入：<script>alert('xss')</script>
    // 预期：脚本标签被清除，内容不出现在输出中
}

#[test]
fn test_xss_prevention_event_handlers() {
    // 恶意输入：<img src=x onerror="alert()">
    // 预期：事件处理器被移除
}

#[test]
fn test_xss_prevention_javascript_url() {
    // 恶意输入：<a href="javascript:alert()">click</a>
    // 预期：href 被清除或重写
}

#[test]
fn test_empty_markdown() {
    // 输入：""（空字符串）
    // 预期：不 panic，返回 Ok((html, toc, summary))
}

#[test]
fn test_very_long_markdown() {
    // 输入：10MB 的 Markdown 文本
    // 预期：能够处理（或返回合理的错误）
}

#[test]
fn test_malformed_heading() {
    // 输入：# # # (多个 #，非标准 Markdown)
    // 预期：合理降级或返回错误
}
```

**属性测试（PBT）**：

```rust
#[cfg(test)]
mod property_tests {
    use proptest::prelude::*;

    proptest! {
        #[test]
        fn render_never_panics(markdown in ".*") {
            // 任意字符串输入都不应 panic
            let engine = MarkdownEngine::new();
            let _ = engine.render(&markdown);
            // 测试通过意味着 render 对所有输入都健壮
        }
    }
}
```

#### 3.1.3 测试覆盖清单

- [ ] `MarkdownEngine::new()` - 初始化
- [ ] `MarkdownEngine::render()` - 基础渲染（段落、列表、代码）
- [ ] TOC 生成 - 多级标题、anchor 生成、JSON 序列化
- [ ] 摘要生成 - 字符收集、截断逻辑
- [ ] XSS 防护 - 脚本标签、事件处理器、javascript: 协议
- [ ] 边界条件 - 空输入、超长输入、非 UTF-8 输入
- [ ] 属性测试 - 任意输入都不 panic

---

### 3.2 Mirror 服务 (搜索索引)

#### 3.2.1 核心模块

| 模块 | 函数/结构 | 测试重点 | 优先级 |
|------|----------|---------|--------|
| **engine** | `SearchEngine::new()` | 索引初始化、Schema 定义 | P0 |
| | `SearchEngine::index_post()` | 文档索引、字段映射 | P0 |
| | `SearchEngine::search()` | 查询解析、排序、分页 | P0 |
| **tokenizer** | `JiebaTokenizer::tokenize()` | 中文分词、词汇覆盖 | P0 |
| | 停用词处理 | 去除 a/the 等词 | P0 |

#### 3.2.2 测试用例设计

**Happy Path**：

```rust
#[test]
fn test_search_engine_indexing() {
    // 1. 创建索引
    // 2. 索引一篇文章
    // 3. 验证文档被正确存储
}

#[test]
fn test_chinese_tokenization() {
    // 输入：中文词汇（例如 "Rust 编程语言"）
    // 预期：正确分词为 ["Rust", "编程", "语言"]
}

#[test]
fn test_search_ranking() {
    // 索引多篇文章，搜索相关关键词
    // 验证排序：相关性高的文章排在前
}
```

**Error & Edge Cases**：

```rust
#[test]
fn test_index_with_mock_persistence() {
    // 使用 Mock 模拟文件系统故障
    // 预期：返回 Err，不 panic
}

#[test]
fn test_search_empty_index() {
    // 在空索引上搜索
    // 预期：返回空结果，不 panic
}

#[test]
fn test_search_special_characters() {
    // 查询：包含特殊字符（& * ! ?）
    // 预期：合理处理或返回错误
}

#[test]
fn test_concurrent_indexing() {
    // 并发索引多篇文章
    // 预期：使用 RwLock 保证数据一致性
}
```

#### 3.2.3 Mock 使用示例

```rust
#[cfg(test)]
mod tests {
    use super::*;
    use mockall::predicate::*;

    #[test]
    fn test_search_with_storage_failure() {
        // Mock 一个存储层 trait
        let mut mock_storage = MockStorage::new();

        mock_storage
            .expect_persist()
            .with(eq("test_index"))
            .return_once(|_| {
                Err(anyhow::anyhow!("Disk full"))
            });

        let engine = SearchEngine::with_storage(mock_storage);
        let result = engine.save("test_index");

        assert!(result.is_err());
        assert_eq!(result.unwrap_err().to_string(), "Disk full");
    }
}
```

#### 3.2.4 测试覆盖清单

- [ ] 索引创建 - 初始化、Schema 定义
- [ ] 文档索引 - 字段映射、数据类型转换
- [ ] 中文分词 - 多字词、单字词、专有名词
- [ ] 查询执行 - 精确查询、模糊查询、布尔查询
- [ ] 排序与分页 - 相关性排序、分页偏移
- [ ] 并发访问 - 多线程读写、锁竞争
- [ ] 持久化 - 索引保存、加载、损坏恢复

---

### 3.3 Oracle 服务 (搜索聚合)

#### 3.3.1 核心模块

| 模块 | 函数/结构 | 测试重点 | 优先级 |
|------|----------|---------|--------|
| **aggregation** | `AggregationService::search()` | 多源聚合、合并排序 | P0 |
| | 去重逻辑 | 相同文档去重 | P0 |
| | 元数据聚合 | 评论数、点赞数的合并 | P1 |

#### 3.3.2 测试用例设计

```rust
#[test]
fn test_aggregation_search_results() {
    // 从 Mirror 和其他源获取搜索结果
    // 预期：正确合并、去重、排序
}

#[test]
fn test_deduplication_logic() {
    // 多个源返回相同的文章 ID
    // 预期：去重后只返回一份
}

#[test]
fn test_metadata_merge() {
    // 来自 Beacon 的评论数和 Mirror 的搜索得分合并
    // 预期：聚合正确的元数据
}

#[test]
fn test_search_timeout_handling() {
    // 模拟 Mirror 服务超时
    // 预期：降级处理（返回部分结果或错误）
}
```

---

## 4. 目录与文件结构

### 4.1 推荐的测试文件组织

#### 选项 A：内联测试（推荐用于单元测试）

```
rust_services/
├── forge/
│   └── src/
│       ├── main.rs
│       ├── server.rs
│       └── engine.rs
│           └── [内含 mod tests { ... }]
│
├── mirror/
│   └── src/
│       ├── main.rs
│       ├── server.rs
│       ├── engine.rs
│       │   └── [内含 mod tests { ... }]
│       └── tokenizer.rs
│           └── [内含 mod tests { ... }]
│
└── oracle/
    └── src/
        ├── main.rs
        └── aggregation.rs
            └── [内含 mod tests { ... }]
```

**优点**：

- 测试与实现在同一文件，易于维护
- 可以访问 `private` 函数
- 编译快

**示例**：

```rust
// src/engine.rs

pub struct MarkdownEngine { ... }

impl MarkdownEngine {
    pub fn render(&self, md: &str) -> Result<...> { ... }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_render_basic() {
        // ...
    }
}
```

#### 选项 B：独立测试文件（用于集成测试）

```
rust_services/
├── forge/
│   ├── src/
│   │   ├── lib.rs
│   │   └── engine.rs
│   └── tests/
│       ├── integration_test.rs       # 完整链路测试
│       └── fixtures/
│           └── sample_markdown.md    # 测试数据
│
└── [其他服务]
```

**优点**：

- 独立编译，便于组织大规模测试
- 可以测试公共 API（`pub` 接口）
- 便于共享测试数据（fixtures）

**示例**：

```rust
// tests/integration_test.rs

use forge::MarkdownEngine;

#[test]
fn test_full_rendering_pipeline() {
    // 完整的渲染链路测试
}
```

### 4.2 推荐方案

**单元测试** → 使用选项 A（内联 mod tests）  
**集成测试** → 使用选项 B（tests/ 目录）

### 4.3 文件模板

**标准的 Rust 单元测试模块**：

```rust
#[cfg(test)]
mod tests {
    use super::*;
    use pretty_assertions::assert_eq;  // 更好的 assert 输出
    
    // ======================================
    // 辅助函数
    // ======================================
    
    fn setup() -> MarkdownEngine {
        // 公共的测试初始化
        MarkdownEngine::new()
    }

    // ======================================
    // 单元测试
    // ======================================

    #[test]
    fn test_happy_path() {
        let engine = setup();
        let result = engine.render("# Hello");
        
        assert!(result.is_ok());
        let (html, _, _) = result.unwrap();
        assert!(html.contains("<h1>"));
    }

    #[test]
    fn test_error_case_empty_input() {
        let engine = setup();
        let result = engine.render("");
        
        // 应该返回 Ok，而不是 Err 或 panic
        assert!(result.is_ok());
    }

    #[test]
    #[should_panic]  // 仅在期望 panic 时使用（应该很少）
    fn test_expected_panic() {
        // ...
    }
}
```

---

## 5. 实施步骤

### 5.1 第一阶段：基础设施配置（1 周）

#### 步骤 1：更新 Cargo.toml

- [ ] 在 `rust_services/Cargo.toml` 的 `[workspace.dependencies]` 中添加测试依赖
- [ ] 在各包的 `Cargo.toml` 中添加 `[dev-dependencies]`（复用 workspace 版本）

```bash
# 验证配置
cd rust_services
cargo check --all
```

#### 步骤 2：安装测试工具链

```bash
# 1. 安装覆盖率工具
cargo install cargo-tarpaulin

# 2. 安装代码分析工具（检查 unwrap/expect）
cargo install clippy  # 通常已随 rustup 安装

# 3. 验证工具可用
cargo tarpaulin --version
cargo clippy --version
```

#### 步骤 3：创建测试目录结构

```bash
# 创建 tests 目录（用于集成测试）
mkdir -p rust_services/{forge,mirror,oracle}/tests
mkdir -p rust_services/{forge,mirror,oracle}/tests/fixtures

# 创建 fixtures 示例文件
cat > rust_services/forge/tests/fixtures/sample.md << 'EOF'
# Getting Started

This is a sample markdown document for testing.

## Installation

```bash
cargo install forge
```

EOF

```

#### 步骤 4：编写首个测试

在 `rust_services/forge/src/engine.rs` 中添加基础测试模块：

```rust
#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_markdown_engine_creation() {
        let engine = MarkdownEngine::new();
        // 仅验证能够创建实例
        assert_eq!(std::mem::size_of_val(&engine) > 0, true);
    }
}
```

**验证**：

```bash
cd rust_services/forge
cargo test
```

---

### 5.2 第二阶段：核心逻辑覆盖（2 周）

#### 步骤 1：Forge 单元测试

**目标**：MarkdownEngine 覆盖率 ≥ 90%

创建文件：`rust_services/forge/src/engine.rs` (完整测试模块)

```rust
#[cfg(test)]
mod tests {
    use super::*;
    use pretty_assertions::assert_eq;

    // ====== Basic Rendering ======
    
    #[test]
    fn test_render_heading_h1() { /* ... */ }

    #[test]
    fn test_render_heading_multi_level() { /* ... */ }

    #[test]
    fn test_render_paragraph() { /* ... */ }

    #[test]
    fn test_render_bold_italic() { /* ... */ }

    #[test]
    fn test_render_code_block() { /* ... */ }

    #[test]
    fn test_render_list_unordered() { /* ... */ }

    #[test]
    fn test_render_list_ordered() { /* ... */ }

    #[test]
    fn test_render_table() { /* ... */ }

    // ====== XSS Prevention ======

    #[test]
    fn test_xss_script_tags_removed() { /* ... */ }

    #[test]
    fn test_xss_event_handlers_removed() { /* ... */ }

    #[test]
    fn test_xss_javascript_protocol_removed() { /* ... */ }

    #[test]
    fn test_xss_iframe_tags_removed() { /* ... */ }

    // ====== TOC Generation ======

    #[test]
    fn test_toc_generation_single_heading() { /* ... */ }

    #[test]
    fn test_toc_generation_nested_headings() { /* ... */ }

    #[test]
    fn test_toc_anchor_generation() { /* ... */ }

    #[test]
    fn test_toc_json_format() { /* ... */ }

    // ====== Summary Generation ======

    #[test]
    fn test_summary_text_extraction() { /* ... */ }

    #[test]
    fn test_summary_truncation_at_limit() { /* ... */ }

    #[test]
    fn test_summary_multiline_text() { /* ... */ }

    // ====== Edge Cases ======

    #[test]
    fn test_empty_input() { /* ... */ }

    #[test]
    fn test_very_long_input() { /* ... */ }

    #[test]
    fn test_malformed_markdown() { /* ... */ }

    // ====== Property Testing ======

    #[test]
    fn prop_render_never_panics() {
        // 使用 proptest 进行模糊测试
    }
}
```

#### 步骤 2：Mirror 单元测试

**目标**：SearchEngine 和 Tokenizer 覆盖率 ≥ 80%

创建测试文件：

- `rust_services/mirror/src/engine.rs` - 搜索引擎测试
- `rust_services/mirror/src/tokenizer.rs` - 分词器测试

**tokenizer 测试示例**：

```rust
#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_tokenize_english_simple() {
        let tokenizer = JiebaTokenizer::new();
        let tokens = tokenizer.tokenize("hello world");
        assert_eq!(tokens, vec!["hello", "world"]);
    }

    #[test]
    fn test_tokenize_chinese() {
        let tokenizer = JiebaTokenizer::new();
        let tokens = tokenizer.tokenize("Rust 编程语言");
        assert_eq!(tokens, vec!["Rust", "编程", "语言"]);
    }

    #[test]
    fn test_tokenize_stopwords_removed() {
        let tokenizer = JiebaTokenizer::new();
        let tokens = tokenizer.tokenize("the quick brown fox");
        // "the" 应该被移除（停用词）
        assert!(!tokens.contains(&"the"));
        assert!(tokens.contains(&"quick"));
    }
}
```

#### 步骤 3：Oracle 单元测试

**目标**：AggregationService 覆盖率 ≥ 70%

---

### 5.3 第三阶段：Error Handling 专项测试（1 周）

#### 目标

确保所有 `Result<T, E>` 都有测试覆盖，消除 `unwrap()` 滥用。

#### 步骤 1：扫描所有 unwrap/expect

```bash
cd rust_services

# 检查所有 unwrap 调用
cargo clippy --all -- -W clippy::unwrap_used

# 检查所有 expect 调用
cargo clippy --all -- -W clippy::expect_used
```

**输出示例**：

```
warning: called `unwrap` on an `Option` value
  --> rust_services/forge/src/main.rs:28:18
   |
28 |     let cfg = config.parse::<ServerConfig>()?
   |              .expect("Server config missing");
   |
```

#### 步骤 2：为每个 unwrap 编写测试

**原则**：用测试覆盖 `Err` 分支，而非 `unwrap()`。

**示例**：

```rust
// 修改前（不好）
let json = serde_json::to_string(&event).unwrap();

// 修改后（好）
let json = serde_json::to_string(&event)
    .map_err(|e| anyhow::anyhow!("Failed to serialize event: {}", e))?;

// 测试
#[test]
fn test_event_serialization_error_handling() {
    // 创建一个无法序列化的对象（或 Mock）
    let event = UnserializableEvent::new();
    let result = serde_json::to_string(&event);
    
    assert!(result.is_err());
    let err = result.unwrap_err();
    assert!(err.to_string().contains("serialize"));
}
```

#### 步骤 3：Error 路径覆盖清单

| 模块 | 错误场景 | 测试函数 |
|------|---------|---------|
| Forge | JSON 序列化失败 | `test_toc_serialization_error` |
| | Config 加载失败 | `test_config_missing` |
| Mirror | 索引保存失败 | `test_index_persist_error` |
| | 文件权限不足 | `test_file_permission_error` |
| Oracle | 服务超时 | `test_aggregation_timeout` |
| | 网络错误 | `test_network_unreachable` |

---

### 5.4 第四阶段：集成测试与文档（1 周）

#### 步骤 1：编写集成测试

在 `tests/` 目录中编写端到端测试：

```
rust_services/forge/tests/integration_test.rs
```

```rust
use forge::MarkdownEngine;

#[test]
fn test_full_rendering_pipeline() {
    let engine = MarkdownEngine::new();
    
    let markdown = r#"
# Article Title

## Introduction
This is the introduction paragraph.

### Subsection
Some content here with **bold** and *italic*.

```rust
fn main() {
    println!("Hello");
}
```

## Conclusion

End of article.
    "#;

    let (html, toc, summary) = engine.render(markdown)
        .expect("Rendering failed");

    // 验证 HTML
    assert!(html.contains("<h1"));
    assert!(html.contains("Article Title"));

    // 验证 TOC
    let toc_entries: Vec<_> = serde_json::from_str(&toc)
        .expect("Invalid TOC JSON");
    assert!(toc_entries.len() >= 3);

    // 验证摘要
    assert!(!summary.is_empty());
    assert!(summary.contains("introduction"));
}

```

#### 步骤 2：编写覆盖率报告脚本

创建 `scripts/test_coverage.sh`：

```bash
#!/bin/bash

set -e

echo "📊 Running Rust coverage tests..."

cd rust_services

# 生成 HTML 覆盖率报告
cargo tarpaulin \
    --all \
    --out Html \
    --output-dir coverage \
    --timeout 300 \
    --exclude-files tests/ \
    --verbose

echo ""
echo "✅ Coverage report generated at: coverage/index.html"

# 显示覆盖率摘要
echo ""
echo "📈 Coverage Summary:"
cargo tarpaulin --all --out Stdout | tail -20
```

使用：

```bash
chmod +x scripts/test_coverage.sh
./scripts/test_coverage.sh
```

#### 步骤 3：文档化

为每个服务编写 `TEST.md`：

```
rust_services/forge/TEST.md
rust_services/mirror/TEST.md
rust_services/oracle/TEST.md
```

**内容**：

- 如何运行测试
- 测试覆盖范围
- 已知的测试限制
- 如何添加新测试

---

## 6. 代码示例

### 6.1 Cargo.toml 完整配置示例

**文件**：`rust_services/Cargo.toml`

```toml
[workspace]
resolver = "2"
members = [
    "common",
    "forge",
    "mirror",
    "oracle",
]

[workspace.package]
version = "3.2.0"
edition = "2021"
authors = ["Bifrost Team"]
license = "MIT"

# ================================================================
# 所有依赖统一管理
# ================================================================
[workspace.dependencies]
# === 核心运行时 ===
tokio = { version = "1.48.0", features = ["full"] }
tonic = "0.14.2"
prost = "0.14.1"
prost-types = "0.14.1"
futures = "0.3.31"

# === 序列化与错误处理 ===
serde = { version = "1.0.228", features = ["derive"] }
serde_json = "1.0.146"
anyhow = "1.0.100"
thiserror = "2.0.17"

# === 日志与追踪 ===
tracing = "0.1.44"
tracing-subscriber = { version = "0.3.22", features = ["env-filter"] }

# === 业务特定 ===
pulldown-cmark = { version = "0.13.0", default-features = false, features = ["simd", "html"] }
ammonia = "4.1.2"
slug = "0.1.6"
tantivy = "0.25.0"

# === 构建工具 ===
tonic-build = "0.14.2"

# === 可观测性 ===
opentelemetry = "0.31.0"
opentelemetry_sdk = { version = "0.31.0", features = ["rt-tokio"] }
opentelemetry-otlp = { version = "0.31.0", features = ["grpc-tonic"] }
opentelemetry-semantic-conventions = "0.31.0"
tracing-opentelemetry = "0.32.0"

# === HTTP ===
http = "1.4.0"

# === 内部库 ===
common = { path = "./common" }

# ================================================================
# [新增] 测试依赖 (所有包共用)
# ================================================================
mockall = "0.12.0"
pretty_assertions = "1.4.0"
proptest = "1.4.0"
tokio-test = "0.4.1"
tempfile = "3.10.0"

# ================================================================
# 测试编译优化
# ================================================================
[profile.test]
opt-level = 1

[profile.test.package."*"]
opt-level = 3
```

### 6.2 使用 mockall 的完整示例

**场景**：Mock 一个外部的存储服务，测试 SearchEngine 的持久化逻辑。

**步骤 1**：定义可 Mock 的 Trait

```rust
// rust_services/mirror/src/storage.rs

use anyhow::Result;

/// 存储层 trait，便于 Mock
#[mockall::automock]
pub trait Storage: Send + Sync {
    /// 保存索引到磁盘
    fn persist(&self, index_path: &str, index_data: &[u8]) -> Result<()>;
    
    /// 从磁盘加载索引
    fn load(&self, index_path: &str) -> Result<Vec<u8>>;
    
    /// 删除索引
    fn delete(&self, index_path: &str) -> Result<()>;
}

/// 真实的文件系统实现
pub struct FileSystemStorage;

impl Storage for FileSystemStorage {
    fn persist(&self, path: &str, data: &[u8]) -> Result<()> {
        std::fs::write(path, data)?;
        Ok(())
    }

    fn load(&self, path: &str) -> Result<Vec<u8>> {
        Ok(std::fs::read(path)?)
    }

    fn delete(&self, path: &str) -> Result<()> {
        std::fs::remove_file(path)?;
        Ok(())
    }
}
```

**步骤 2**：修改 SearchEngine 接受 Storage trait

```rust
// rust_services/mirror/src/engine.rs

pub struct SearchEngine {
    index: Index,
    storage: Box<dyn Storage>,
}

impl SearchEngine {
    pub fn new(storage: Box<dyn Storage>) -> Self {
        // ...
    }

    pub fn save(&self, path: &str) -> anyhow::Result<()> {
        let index_data = self.serialize_index()?;
        self.storage.persist(path, &index_data)?;
        Ok(())
    }
}
```

**步骤 3**：编写单元测试

```rust
// rust_services/mirror/src/engine.rs

#[cfg(test)]
mod tests {
    use super::*;
    use mockall::predicate::*;

    #[test]
    fn test_save_success() {
        let mut mock_storage = MockStorage::new();

        // 设置期望：persist 被调用一次，参数为 "test_index"
        mock_storage
            .expect_persist()
            .with(eq("test_index"), always())
            .times(1)
            .returning(|_, _| Ok(()));

        let engine = SearchEngine::new(Box::new(mock_storage));
        let result = engine.save("test_index");

        assert!(result.is_ok());
    }

    #[test]
    fn test_save_disk_full_error() {
        let mut mock_storage = MockStorage::new();

        // 模拟磁盘满的错误
        mock_storage
            .expect_persist()
            .with(eq("test_index"), always())
            .returning(|_, _| {
                Err(anyhow::anyhow!("No space left on device"))
            });

        let engine = SearchEngine::new(Box::new(mock_storage));
        let result = engine.save("test_index");

        assert!(result.is_err());
        assert_eq!(
            result.unwrap_err().to_string(),
            "No space left on device"
        );
    }

    #[test]
    fn test_save_permission_denied() {
        let mut mock_storage = MockStorage::new();

        // 模拟权限错误
        mock_storage
            .expect_persist()
            .returning(|_, _| {
                Err(anyhow::anyhow!("Permission denied"))
            });

        let engine = SearchEngine::new(Box::new(mock_storage));
        let result = engine.save("test_index");

        assert!(result.is_err());
    }

    #[test]
    fn test_load_missing_index() {
        let mut mock_storage = MockStorage::new();

        // 模拟文件不存在
        mock_storage
            .expect_load()
            .with(eq("nonexistent"))
            .returning(|_| {
                Err(anyhow::anyhow!("File not found"))
            });

        let engine = SearchEngine::new(Box::new(mock_storage));
        let result = engine.load_from_disk("nonexistent");

        assert!(result.is_err());
    }

    #[test]
    fn test_concurrent_persist_calls() {
        let mut mock_storage = MockStorage::new();

        // 允许多次调用，验证线程安全性
        mock_storage
            .expect_persist()
            .returning(|_, _| Ok(()));

        let engine = std::sync::Arc::new(
            SearchEngine::new(Box::new(mock_storage))
        );

        let handles: Vec<_> = (0..10)
            .map(|i| {
                let engine = engine.clone();
                std::thread::spawn(move || {
                    engine.save(&format!("index_{}", i))
                })
            })
            .collect();

        for handle in handles {
            assert!(handle.join().unwrap().is_ok());
        }
    }
}
```

---

## 7. 常见问题与最佳实践

### 7.1 常见错误

| 错误 | 原因 | 解决方案 |
|------|------|--------|
| `error: cannot find attribute`automock`in this scope` | 未导入 mockall | 添加 `use mockall::automock;` |
| `test panicked` (在 unwrap 处) | 测试覆盖了错误路径但代码 unwrap 了 | 改为 `?` 操作符或显式错误处理 |
| 测试超时 | 文件 I/O 或网络调用阻塞 | 使用 Mock 或 `#[tokio::test]` 注解 |
| Mock 期望不匹配 | 调用参数与期望不符 | 使用 `eq()`, `always()` predicates 或 `MatcherFunc` |

### 7.2 最佳实践

#### 原则 1：使用 Result，不要 unwrap

**不好**：

```rust
let value = json_string.parse::<Value>().unwrap();
```

**好**：

```rust
let value = json_string.parse::<Value>()?;

// 或者带上下文
let value = json_string.parse::<Value>()
    .map_err(|e| anyhow::anyhow!("Failed to parse JSON: {}", e))?;
```

#### 原则 2：测试错误路径

**不好**：

```rust
#[test]
fn test_function() {
    let result = some_function();
    assert!(result.is_ok());  // 只测试正常情况
}
```

**好**：

```rust
#[test]
fn test_function_success() {
    let result = some_function();
    assert!(result.is_ok());
}

#[test]
fn test_function_invalid_input() {
    let result = some_function_with_invalid_input();
    assert!(result.is_err());
    // 验证错误信息
    assert!(result.unwrap_err().to_string().contains("invalid"));
}
```

#### 原则 3：使用 pretty_assertions 获得更好的失败输出

**不好**：

```rust
assert_eq!(actual, expected);  // 差的错误信息
```

**好**：

```rust
use pretty_assertions::assert_eq;
assert_eq!(actual, expected);  // 彩色、结构化的差异显示
```

#### 原则 4：为复杂逻辑使用属性测试

**不好**：

```rust
#[test]
fn test_function_many_cases() {
    assert!(test_case1());
    assert!(test_case2());
    // ... 10 个 case ...
}
```

**好**：

```rust
use proptest::prelude::*;

proptest! {
    #[test]
    fn test_function_property(input in any::<String>()) {
        // 自动生成 100+ 个随机输入
        let result = function(&input);
        prop_assert!(result.is_ok() || result.is_err());  // 验证不 panic
    }
}
```

---

## 8. 成功标准与验收清单

### 8.1 P0 必须完成

- [ ] **Forge MarkdownEngine**：≥ 90% 覆盖率
  - [ ] Happy Path 测试（基础渲染）
  - [ ] XSS 防护测试（所有攻击向量）
  - [ ] Edge Cases（空输入、超长输入）
  - [ ] 属性测试（任意输入都不 panic）

- [ ] **Mirror SearchEngine**：≥ 80% 覆盖率
  - [ ] 索引创建和文档插入
  - [ ] 查询执行和结果排序
  - [ ] 并发访问安全性
  - [ ] 错误处理（磁盘满、权限错误）

- [ ] **消除 unwrap/expect**：
  - [ ] 扫描所有代码，找出所有 `unwrap()` 调用
  - [ ] 为每个 unwrap 编写对应的错误路径测试
  - [ ] 将 `unwrap()` 替换为 `?` 或 `.map_err()`

### 8.2 质量门槛

```
✅ 所有测试通过（零 panic）
✅ 整体覆盖率 ≥ 75%
✅ Forge: ≥ 90%, Mirror: ≥ 80%, Oracle: ≥ 70%
✅ 没有 #[should_panic]（应该验证 Err，不应该测试 panic）
✅ 所有 Result<T, E> 都有 Err 路径测试
```

### 8.3 验收流程

```bash
# 1. 运行所有测试
cargo test --all --verbose

# 2. 检查覆盖率
cargo tarpaulin --all --out Html

# 3. 检查代码质量（unwrap/expect）
cargo clippy --all -- -W clippy::unwrap_used -W clippy::expect_used

# 4. 格式检查
cargo fmt --all --check

# 5. 生成文档
cargo doc --all --no-deps --open
```

---

## 9. 实施时间表与分工

| 周数 | 任务 | 交付物 | 负责人 |
|------|------|--------|--------|
| **W1** | Cargo.toml 配置、测试框架搭建 | 更新的 Cargo.toml、首个测试通过 | @rust-arch |
| | Forge 基础单元测试（XSS、渲染） | forge_test.rs (200+ 行) | @forge-dev |
| **W2** | Mirror 单元测试（搜索、分词） | mirror_engine_test.rs (150+ 行) | @mirror-dev |
| | Oracle 单元测试 | oracle_aggregation_test.rs (100+ 行) | @oracle-dev |
| **W3** | Error handling 专项测试 | unwrap 消除、错误路径覆盖 | @qa |
| | 集成测试补充 | tests/ 目录下的集成测试 | @all |
| **W4** | 覆盖率报告、文档完善 | coverage 报告、TEST.md | @doc |

---

## 10. 参考文档与资源

- [Rust Book - Testing Chapter](https://doc.rust-lang.org/book/ch11-00-testing.html)
- [mockall 官方文档](https://docs.rs/mockall/latest/mockall/)
- [proptest 属性测试](https://docs.rs/proptest/latest/proptest/)
- [Rust API Guidelines - Error Handling](https://rust-lang.github.io/api-guidelines/type-safety.html)
- [OWASP - XSS 防护](https://owasp.org/www-community/attacks/xss/)

---

## 附录：快速参考

### 常用命令速查

```bash
# 运行所有测试
cargo test --all

# 单个包的测试
cargo test -p forge

# 特定测试函数
cargo test -p forge render_basic

# 显示 println! 输出
cargo test -- --nocapture

# 单线程运行（便于调试）
cargo test -- --test-threads=1

# 生成覆盖率报告
cargo tarpaulin --all --out Html

# 检查代码问题
cargo clippy --all

# 格式化代码
cargo fmt --all

# 生成文档
cargo doc --all --no-deps --open
```

### 测试模板速查

```rust
// 基础单元测试
#[test]
fn test_feature() {
    let result = function();
    assert!(result.is_ok());
}

// 错误路径测试
#[test]
fn test_feature_error() {
    let result = function_with_error();
    assert!(result.is_err());
}

// 异步测试
#[tokio::test]
async fn test_async_function() {
    let result = async_function().await;
    assert!(result.is_ok());
}

// Mock 测试
#[test]
fn test_with_mock() {
    let mut mock = MockTrait::new();
    mock.expect_method().return_const(value);
    
    // 使用 mock
}

// 属性测试
proptest! {
    #[test]
    fn test_property(x in any::<Type>()) {
        prop_assert!(predicate);
    }
}
```

---

**文档版本控制**：

- v1.0 - 初版，完整的 Rust 测试计划
- 更新日期：2025-12-23
