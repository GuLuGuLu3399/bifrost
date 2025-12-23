// rust_services/oracle/src/storage/duck.rs

use anyhow::{anyhow, Result};
use duckdb::{params, Connection};
use std::sync::{Arc, Mutex};

use crate::ingestion::EventPayload; // Service <-> Storage 数据契约
use tracing::info;

#[derive(Clone)]
pub struct DuckStore {
    // DuckDB Connection 本身不是 Sync；用 Mutex 串行化访问。
    conn: Arc<Mutex<Connection>>,
}

impl DuckStore {
    /// 打开或创建数据库文件
    pub fn open(path: &str) -> Result<Self> {
        let conn = Connection::open(path)?;

        // 初始化表结构
        conn.execute_batch(include_str!("schema.sql"))?;

        info!("DuckDB initialized at: {}", path);

        Ok(Self {
            conn: Arc::new(Mutex::new(conn)),
        })
    }

    /// 批量写入事件 (Batch Insert)
    /// 性能关键点：使用事务 + Prepared Statement
    pub fn insert_batch(&self, events: Vec<EventPayload>) -> Result<()> {
        if events.is_empty() {
            return Ok(());
        }

        // PoisonError<MutexGuard<Connection>> 不能直接 `?` 转 anyhow（guard 非 Send）。
        // 这里显式转成 anyhow::Error。
        let mut conn = self
            .conn
            .lock()
            .map_err(|_| anyhow!("duckdb connection mutex poisoned"))?;

        let tx = conn.transaction()?;

        {
            let mut stmt = tx.prepare(
                "INSERT INTO raw_events (event_type, target_id, user_id, meta, created_at)\n                 VALUES (?, ?, ?, ?, ?)",
            )?;

            for e in events {
                // DuckDB rust 绑定层对 chrono::NaiveDateTime 的 ToSql 支持不稳定；
                // 这里统一存 RFC3339 字符串，查询侧用 strptime/strftime 处理。
                let created_at_rfc3339 = e.created_at.to_rfc3339();

                stmt.execute(params![
                    e.event_type,
                    e.target_id,
                    e.user_id,
                    e.meta.to_string(),
                    created_at_rfc3339
                ])?;
            }
        }

        tx.commit()?;
        Ok(())
    }

    /// 查询仪表盘趋势 (OLAP 查询)
    /// 返回: [(日期, PV数)]
    pub fn get_trend(&self, days: i64) -> Result<Vec<(String, i64)>> {
        let conn = self
            .conn
            .lock()
            .map_err(|_| anyhow!("duckdb connection mutex poisoned"))?;

        // created_at 存的是 RFC3339 字符串时，通过 strptime 转成 timestamp 再计算。
        let mut stmt = conn.prepare(
            r#"
            SELECT
                strftime(strptime(created_at, '%Y-%m-%dT%H:%M:%S%.f%z'), '%Y-%m-%d') as day,
                count(*) as pv
            FROM raw_events
            WHERE strptime(created_at, '%Y-%m-%dT%H:%M:%S%.f%z') > now() - INTERVAL ? DAYS
            GROUP BY day
            ORDER BY day ASC
            "#,
        )?;

        let rows = stmt.query_map(params![days], |row| Ok((row.get(0)?, row.get(1)?)))?;

        let mut result = Vec::new();
        for row in rows {
            result.push(row?);
        }
        Ok(result)
    }
}