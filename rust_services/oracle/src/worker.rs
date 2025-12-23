// rust_services/oracle/src/worker.rs

use crate::storage::duck::DuckStore;
use std::sync::Arc;
use std::time::Duration;
use tracing::{error, info, instrument};

pub struct AnalyticsWorker {
    store: DuckStore,
}

impl AnalyticsWorker {
    pub fn new(store: DuckStore) -> Self {
        Self { store }
    }

    /// 启动后台任务
    /// 这里的逻辑是：每隔 1 小时执行一次聚合
    pub async fn run(self: Arc<Self>) {
        info!("👷 Analytics Worker started");

        // 1. 定义定时器 (例如每 1 小时触发一次)
        // 在开发环境为了测试效果，可以设为 1 分钟: Duration::from_secs(60)
        let mut interval = tokio::time::interval(Duration::from_secs(3600));

        loop {
            interval.tick().await; // 等待下一次触发

            let worker = self.clone();
            // 放到 blocking 线程去跑 SQL
            if let Err(e) = tokio::task::spawn_blocking(move || {
                worker.aggregate_daily_stats()
            })
            .await
            {
                error!("Worker spawn error: {}", e);
            }
        }
    }

    /// 核心逻辑：执行预聚合 SQL
    #[instrument(skip(self))]
    fn aggregate_daily_stats(&self) -> anyhow::Result<()> {
        info!("Running scheduled aggregation...");

        // 逻辑：将 raw_events 里的数据聚合写入 daily_stats
        // 这里的 SQL 利用了 DuckDB 的强大特性
        let sql = r#"
            INSERT OR REPLACE INTO daily_stats (date, metric, dimension, dimension_value, value)
            SELECT 
                strftime(created_at, '%Y-%m-%d') as date,
                'pv' as metric,
                'global' as dimension,
                'all' as dimension_value,
                count(*) as value
            FROM raw_events
            WHERE DATE(created_at) = CURRENT_DATE - INTERVAL '1' DAY
            
            UNION ALL
            
            SELECT 
                strftime(created_at, '%Y-%m-%d') as date,
                'uv' as metric,
                'global' as dimension,
                'all' as dimension_value,
                count(distinct user_id) as value
            FROM raw_events
            WHERE user_id != 0 
                AND DATE(created_at) = CURRENT_DATE - INTERVAL '1' DAY
            
            UNION ALL
            
            SELECT 
                strftime(created_at, '%Y-%m-%d') as date,
                'pv' as metric,
                'event_type' as dimension,
                event_type as dimension_value,
                count(*) as value
            FROM raw_events
            WHERE DATE(created_at) = CURRENT_DATE - INTERVAL '1' DAY
            GROUP BY event_type;
        "#;

        // 调用 store 执行
        self.store.raw_execute(sql)?;

        info!("✅ Daily stats aggregated successfully");
        Ok(())
    }

    /// (可选) 计算内容热度分
    #[instrument(skip(self))]
    fn _calculate_content_scores(&self) -> anyhow::Result<()> {
        info!("Calculating content scores...");

        let sql = r#"
            INSERT OR REPLACE INTO daily_stats (date, metric, dimension, dimension_value, value)
            SELECT 
                CURRENT_DATE as date,
                'content_score' as metric,
                'target_id' as dimension,
                target_id as dimension_value,
                (
                    COUNT(CASE WHEN event_type = 'post_view' THEN 1 END) * 1 +
                    COUNT(CASE WHEN event_type = 'post_like' THEN 1 END) * 5 +
                    COUNT(CASE WHEN event_type = 'post_comment' THEN 1 END) * 10
                ) as value
            FROM raw_events
            WHERE DATE(created_at) = CURRENT_DATE - INTERVAL '1' DAY
            GROUP BY target_id;
        "#;

        self.store.raw_execute(sql)?;

        info!("✅ Content scores calculated successfully");
        Ok(())
    }

    /// (可选) 数据清理：删除过期的原始日志
    #[instrument(skip(self))]
    fn _cleanup_old_events(&self, retention_days: i32) -> anyhow::Result<()> {
        info!("Cleaning up old events (retention: {} days)...", retention_days);

        let sql = format!(
            "DELETE FROM raw_events WHERE DATE(created_at) < CURRENT_DATE - INTERVAL '{}' DAY;",
            retention_days
        );

        self.store.raw_execute(&sql)?;

        info!("✅ Old events cleaned up successfully");
        Ok(())
    }
}

impl Clone for AnalyticsWorker {
    fn clone(&self) -> Self {
        Self {
            store: self.store.clone(),
        }
    }
}
