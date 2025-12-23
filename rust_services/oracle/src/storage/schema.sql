-- rust_services/oracle/src/storage/schema.sql

-- 原始事件表 (Append-Only)
CREATE TABLE IF NOT EXISTS raw_events (
    -- 主键：使用 BIGINT 而非 UBIGINT，确保跨数据库兼容性
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    
    -- 核心字段
    event_type VARCHAR(50) NOT NULL,           -- 事件类型，如 'post_view', 'post_like'
    target_id VARCHAR(100) NOT NULL,            -- 目标ID，如文章ID
    user_id BIGINT,                             -- 用户ID，NULL 表示匿名用户
    
    -- 元数据
    meta JSONB,                                 -- 元数据，使用 JSONB 而非 JSON
    ip INET,                                    -- 提取 IP 到独立列便于查询
    user_agent TEXT,                            -- 提取 User-Agent 到独立列
    referrer TEXT,                              -- 引用来源
    
    -- 时间相关字段
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),  -- 使用带时区的时间戳
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- 分区键（如果数据量大可考虑按时间分区）
    event_date DATE GENERATED ALWAYS AS (created_at::DATE) STORED
);

-- 索引
CREATE INDEX IF NOT EXISTS idx_raw_events_created_at ON raw_events(created_at);
CREATE INDEX IF NOT EXISTS idx_raw_events_event_type ON raw_events(event_type);
CREATE INDEX IF NOT EXISTS idx_raw_events_user_id ON raw_events(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_raw_events_target_id ON raw_events(target_id);
CREATE INDEX IF NOT EXISTS idx_raw_events_event_date ON raw_events(event_date);

-- 每日聚合表
CREATE TABLE IF NOT EXISTS daily_stats (
    date DATE NOT NULL,
    metric VARCHAR(50) NOT NULL,                -- 指标，如 'pv', 'uv', 'like_count'
    dimension VARCHAR(100) NOT NULL,            -- 维度，如 'global', 'category:tech'
    dimension_value VARCHAR(200),                -- 维度值
    value BIGINT NOT NULL DEFAULT 0,
    
    -- 元数据
    first_event_at TIMESTAMPTZ,                 -- 当天首次事件时间
    last_event_at TIMESTAMPTZ,                  -- 当天最后事件时间
    total_count BIGINT,                         -- 总计数（用于计算平均值等）
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    PRIMARY KEY (date, metric, dimension, dimension_value)
);

-- 索引
CREATE INDEX IF NOT EXISTS idx_daily_stats_date ON daily_stats(date);
CREATE INDEX IF NOT EXISTS idx_daily_stats_metric ON daily_stats(metric);
CREATE INDEX IF NOT EXISTS idx_daily_stats_dimension ON daily_stats(dimension);
CREATE INDEX IF NOT EXISTS idx_daily_stats_composite ON daily_stats(date, metric, dimension);

-- 小时粒度聚合表（可选，用于更细粒度分析）
CREATE TABLE IF NOT EXISTS hourly_stats (
    date_hour TIMESTAMPTZ NOT NULL,            -- 按小时聚合
    metric VARCHAR(50) NOT NULL,
    dimension VARCHAR(100) NOT NULL,
    dimension_value VARCHAR(200),
    value BIGINT NOT NULL DEFAULT 0,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    PRIMARY KEY (date_hour, metric, dimension, dimension_value)
);

-- 用户行为分析表
CREATE TABLE IF NOT EXISTS user_behavior (
    user_id BIGINT NOT NULL,
    date DATE NOT NULL,
    
    -- 行为统计
    pv_count INT DEFAULT 0,                     -- 页面浏览量
    like_count INT DEFAULT 0,                    -- 点赞数
    comment_count INT DEFAULT 0,                 -- 评论数
    share_count INT DEFAULT 0,                   -- 分享数
    
    -- 活跃度指标
    active_duration INTERVAL,                   -- 活跃时长
    first_active_at TIMESTAMPTZ,                -- 首次活跃时间
    last_active_at TIMESTAMPTZ,                 -- 最后活跃时间
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    PRIMARY KEY (user_id, date)
);

-- 触发器函数：自动更新 updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 为每个表添加 updated_at 触发器
CREATE TRIGGER update_raw_events_updated_at
    BEFORE UPDATE ON raw_events
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_daily_stats_updated_at
    BEFORE UPDATE ON daily_stats
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_hourly_stats_updated_at
    BEFORE UPDATE ON hourly_stats
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_behavior_updated_at
    BEFORE UPDATE ON user_behavior
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 物化视图：热门内容统计（每日刷新）
CREATE MATERIALIZED VIEW IF NOT EXISTS hot_content_daily AS
SELECT 
    target_id,
    COUNT(*) as pv_count,
    COUNT(DISTINCT user_id) as uv_count,
    COUNT(CASE WHEN event_type = 'post_like' THEN 1 END) as like_count,
    DATE(created_at) as stat_date
FROM raw_events
WHERE event_type IN ('post_view', 'post_like')
    AND created_at >= CURRENT_DATE - INTERVAL '30 days'
GROUP BY target_id, DATE(created_at)
WITH DATA;

CREATE UNIQUE INDEX IF NOT EXISTS idx_hot_content_daily 
    ON hot_content_daily(target_id, stat_date);

-- 刷新物化视图的函数
CREATE OR REPLACE FUNCTION refresh_hot_content_daily()
RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY hot_content_daily;
END;
$$ LANGUAGE plpgsql;

-- 注释
COMMENT ON TABLE raw_events IS '原始事件表，存储用户行为事件';
COMMENT ON COLUMN raw_events.user_id IS '用户ID，NULL表示匿名用户';
COMMENT ON COLUMN raw_events.meta IS '事件元数据，JSON格式';
COMMENT ON COLUMN raw_events.ip IS '用户IP地址';
COMMENT ON COLUMN raw_events.user_agent IS '浏览器User-Agent';

COMMENT ON TABLE daily_stats IS '每日聚合统计表';
COMMENT ON COLUMN daily_stats.dimension_value IS '维度具体值，如分类的具体名称';

-- 授予权限（根据实际需求调整）
-- GRANT SELECT, INSERT ON raw_events TO analytics_user;
-- GRANT SELECT ON daily_stats TO analytics_user;