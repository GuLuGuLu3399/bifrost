-- ============================================================================
-- 002_content_core.sql (Refined)
-- 版本: Bifrost v3.2
-- 变更: 修复软删除唯一性冲突，优化覆盖索引
-- ============================================================================

-- 1. 分类表 (Categories)
CREATE TABLE categories (
    id BIGINT PRIMARY KEY,
    name VARCHAR(64) NOT NULL,
    slug VARCHAR(64) NOT NULL,
    description TEXT,
    post_count INTEGER NOT NULL DEFAULT 0,
    version BIGINT NOT NULL DEFAULT 1,
    meta JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    -- [修改] 移除物理约束，改用下方的部分唯一索引
);

-- [新增] 支持软删除的唯一索引
-- 允许存在多条 slug='tech' 的记录，只要其中 N-1 条的 deleted_at 不为空
CREATE UNIQUE INDEX uq_categories_slug_active ON categories(slug); -- Categories 无软删除逻辑可直接用标准唯一，若有则加 WHERE
CREATE UNIQUE INDEX uq_categories_name_active ON categories(name);

-- 2. 标签表 (Tags)
CREATE TABLE tags (
    id BIGINT PRIMARY KEY,
    name VARCHAR(64) NOT NULL,
    slug VARCHAR(64) NOT NULL,
    post_count INTEGER NOT NULL DEFAULT 0,
    version BIGINT NOT NULL DEFAULT 1,
    meta JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX uq_tags_slug_active ON tags(slug);
CREATE UNIQUE INDEX uq_tags_name_active ON tags(name);

-- 3. 文章核心表 (Posts)
CREATE TABLE posts (
    id BIGINT PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL, -- [注意] 这里不再加 UNIQUE 约束
    summary VARCHAR(500),
    resource_key VARCHAR(255),
    cover_image_key VARCHAR(255),
    raw_markdown TEXT NOT NULL,
    html_body TEXT,
    toc_json JSONB,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    visibility VARCHAR(20) NOT NULL DEFAULT 'public',
    author_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    category_id BIGINT REFERENCES categories(id) ON DELETE SET NULL,
    view_count INTEGER NOT NULL DEFAULT 0,
    like_count INTEGER NOT NULL DEFAULT 0,
    comment_count INTEGER NOT NULL DEFAULT 0,
    version BIGINT NOT NULL DEFAULT 1,
    last_trace_id VARCHAR(64),
    created_by BIGINT,
    updated_by BIGINT,
    meta JSONB DEFAULT '{}',
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- 4. 文章-标签关联表
CREATE TABLE post_tags (
    post_id BIGINT REFERENCES posts(id) ON DELETE CASCADE,
    tag_id BIGINT REFERENCES tags(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (post_id, tag_id)
);

-- ============================================================================
-- 5. 索引策略 (优化版)
-- ============================================================================

-- [核心修复] 软删除下的唯一性
-- 允许 slug="hello-world" 被删除后，再次创建同名文章
CREATE UNIQUE INDEX uq_posts_slug_active ON posts(slug)
    WHERE deleted_at IS NULL;

-- [性能优化] Slug 存在性检查的覆盖索引
-- 当执行 SELECT 1 ... WHERE slug=? 时，无需回表查 Heap，直接读索引
-- 配合 ExistsBySlug 使用
CREATE INDEX idx_posts_slug_covering ON posts(slug) INCLUDE (id)
    WHERE deleted_at IS NULL;

-- [查询优化] 状态 + 时间 (Beacon 首页列表)
CREATE INDEX idx_posts_status_pub ON posts(status, published_at DESC)
    WHERE deleted_at IS NULL;

-- [查询优化] 分类 + 状态 (分类列表页)
CREATE INDEX idx_posts_cat_status ON posts(category_id, status, published_at DESC)
    WHERE deleted_at IS NULL;

-- 归属查询
CREATE INDEX idx_posts_author ON posts(author_id);
CREATE INDEX idx_posts_trace ON posts(last_trace_id);