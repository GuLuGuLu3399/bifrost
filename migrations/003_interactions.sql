-- ============================================================================
-- 003_interactions.sql (Refined)
-- 版本: Bifrost v3.2
-- ============================================================================

-- 1. 评论树表 (Comments)
CREATE TABLE comments (
    id BIGINT PRIMARY KEY,
    post_id BIGINT NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    parent_id BIGINT REFERENCES comments(id) ON DELETE CASCADE,
    root_id BIGINT REFERENCES comments(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    version BIGINT NOT NULL DEFAULT 1,
    last_trace_id VARCHAR(64),
    ip_address VARCHAR(45),
    user_agent TEXT,
    meta JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT check_comment_len CHECK (char_length(content) > 0 AND char_length(content) <= 5000)
);

-- 2. 用户互动表 (User Interactions)
CREATE TABLE user_interactions (
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    post_id BIGINT REFERENCES posts(id) ON DELETE CASCADE,
    interaction_type VARCHAR(20) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_trace_id VARCHAR(64),
    PRIMARY KEY (user_id, post_id, interaction_type)
);


-- ============================================================================
-- 4. 索引策略 (优化版)
-- ============================================================================

-- [评论优化] 状态查询
CREATE INDEX idx_comments_post_status ON comments(post_id, status)
    WHERE deleted_at IS NULL;

-- [评论优化] 楼层展开 + 时间排序
-- 优化 SELECT * FROM comments WHERE root_id = ? ORDER BY created_at ASC
CREATE INDEX idx_comments_root_time ON comments(root_id, created_at ASC);

 -- [优化] 加上这个反向索引，Beacon 查"某标签下的文章列表"速度快 100 倍
CREATE INDEX idx_post_tags_tag_id ON post_tags(tag_id);