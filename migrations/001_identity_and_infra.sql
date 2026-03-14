-- ============================================================================
-- 001_identity_and_infra.sql
-- 版本: Bifrost v3.2 (Go+Rust Pure Edition)
-- 职责: 基础设施初始化、用户身份核心
-- 设计: 雪花算法 ID + 乐观锁并发控制 + 可观测性埋点
-- ============================================================================

-- 1. 基础设施初始化
-- 开启 pg_trgm 用于后续的模糊搜索加速 (LIKE / ILIKE)
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- ⚠️ 移除 vector 扩展：向量计算已移交 Rust (Mimir/Oracle)，DB 仅作元数据存储
-- ⚠️ 移除 pgcrypto 扩展：ID 生成已移交应用层 (Snowflake)

-- 2. 枚举定义
DO $$ BEGIN
    CREATE TYPE auth_provider AS ENUM ('local', 'github', 'google');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- 3. 用户核心表 (Users)
CREATE TABLE users (
    -- [ID 策略] 使用 BIGINT 承载 Snowflake ID (应用层生成)
    id BIGINT PRIMARY KEY,

    -- [认证信息]
    username VARCHAR(32) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255), -- OAuth 用户可为空

    -- [个人资料]
    nickname VARCHAR(64),
    bio TEXT,
    avatar_key VARCHAR(255), -- 存储 Object Key (如 avatars/u123/me.jpg)

    -- [RBAC & 状态]
    is_admin BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    -- [OAuth 绑定]
    provider auth_provider NOT NULL DEFAULT 'local',
    provider_id VARCHAR(255),

    -- [运维防御字段 - Ops Shield]
    version BIGINT NOT NULL DEFAULT 1,     -- 乐观锁: 防止并发更新覆盖
    last_trace_id VARCHAR(64),              -- 可观测性: 记录最后一次修改的 TraceID
    meta JSONB DEFAULT '{}',                -- 扩展槽: 存非结构化配置 (如 UI偏好/暗黑模式)

    -- [审计时间戳 - 由 App 维护]
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,                 -- 软删除标识

    -- [业务约束]
    CONSTRAINT check_email_fmt CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'),
    CONSTRAINT check_username_len CHECK (char_length(username) >= 3),
    CONSTRAINT check_auth_logic CHECK (
        (provider = 'local' AND password_hash IS NOT NULL) OR
        (provider != 'local' AND provider_id IS NOT NULL)
    )
);

-- 4. 索引策略
-- 唯一索引 (排除软删除数据)
CREATE UNIQUE INDEX idx_users_username_active ON users(username) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_users_email_active ON users(email) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_users_oauth_active ON users(provider, provider_id) WHERE deleted_at IS NULL;

-- 5. 初始超级管理员种子数据 (Seed)
-- 说明: 默认创建一个超级管理员 superadmin (ID=2)，便于开发测试
-- 生产环境建议通过 CLI 工具或运维脚本初始化，设置真实的安全口令
INSERT INTO users (id, username, email, password_hash, nickname, is_admin, version, meta)
VALUES (
    2,
    'superadmin',
    'superadmin@bifrost.local',
    '$2a$10$MWMU/HAKH1acO5vnO5z3mer2CNuHQ9Gvnc/TvhPW8w2prSsBOahCC', -- bcrypt: SuperAdmin@2025!
    'Super Admin',
    TRUE,
    1,
    '{"theme": "dark"}'
)
ON CONFLICT (id) DO NOTHING;