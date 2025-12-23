// rust_services/oracle/src/ingestion.rs

use crate::storage::duck::DuckStore;
use chrono::{DateTime, Utc};
use serde_json::Value;
use tokio::sync::mpsc;
use std::time::Duration;
use tracing::{info, error, instrument};

/// 内部传输的事件载荷
/// 这是 Service 层和 Storage 层之间的数据契约
#[derive(Debug)]
pub struct EventPayload {
    pub event_type: String, // e.g. "post_view"
    pub target_id: String,  // e.g. "1024"
    pub user_id: i64,       // e.g. 10086
    pub meta: Value,        // e.g. {"ip": "127.0.0.1"}
    pub created_at: DateTime<Utc>,
}

/// 摄入器：负责接收高并发写入，并在后台批量落盘
#[derive(Clone)]
pub struct Ingestor {
    // 使用 MPSC Channel (Multi-Producer, Single-Consumer)
    // Service 层是 Producers，后台任务是 Consumer
    sender: mpsc::Sender<EventPayload>,
}

impl Ingestor {
    /// 启动后台摄入任务，并返回 Ingestor 句柄
    pub fn new(store: DuckStore) -> Self {
        // 创建缓冲区容量为 10000 的通道
        // 如果瞬间积压超过 10000 条，submit 会变慢（Backpressure），保护内存不爆
        let (tx, mut rx) = mpsc::channel::<EventPayload>(10000);

        // 启动后台批处理任务 (Fire-and-forget background task)
        tokio::spawn(async move {
            let mut buffer = Vec::with_capacity(500);
            // 设定强制刷盘间隔：1秒
            // 意味着数据最多在内存里飘 1 秒，即使没凑够 500 条也会落盘
            let mut interval = tokio::time::interval(Duration::from_secs(1));

            loop {
                tokio::select! {
                    // 1. 接收到新事件
                    Some(event) = rx.recv() => {
                        buffer.push(event);
                        // 策略：积攒够 500 条，立即刷盘
                        if buffer.len() >= 500 {
                            flush(&store, &mut buffer).await;
                        }
                    }
                    // 2. 定时器触发 (处理低频写入场景)
                    _ = interval.tick() => {
                        if !buffer.is_empty() {
                            flush(&store, &mut buffer).await;
                        }
                    }
                }
            }
        });

        Self { sender: tx }
    }

    /// 对外暴露的极速写入接口
    /// 这里的开销极小，仅仅是把 struct 塞进 channel
    #[instrument(skip(self, event), fields(event_type = %event.event_type))]
    pub async fn submit(&self, event: EventPayload) {
        if let Err(e) = self.sender.send(event).await {
            // 只有当 channel 关闭（程序退出）时才会走到这里
            error!("Failed to submit event to ingestion buffer: {}", e);
        }
    }
}

/// 核心：执行批量刷盘
async fn flush(store: &DuckStore, buffer: &mut Vec<EventPayload>) {
    let batch_size = buffer.len();
    // 把 buffer 的所有权移出来，清空 buffer 供下次使用
    let events: Vec<_> = buffer.drain(..).collect();

    // 克隆 store 句柄以移动到 blocking 线程
    let store_clone = store.clone();

    // [关键架构点] IO 隔离
    // DuckDB 的写入是同步阻塞的 (CPU & Disk IO 密集)
    // 绝对不能在 Tokio 的异步线程里直接跑，否则会卡死其他 gRPC 请求
    // 必须用 spawn_blocking 扔到专用线程池
    let res = tokio::task::spawn_blocking(move || {
        store_clone.insert_batch(events)
    }).await;

    match res {
        Ok(Ok(_)) => {
            info!("✅ Flushed {} events to DuckDB", batch_size);
        },
        Ok(Err(e)) => {
            // 业务逻辑错误（如 SQL 错误、磁盘满）
            error!("❌ DuckDB insert batch failed: {}", e);
            // 在生产环境，这里应该把 events 写入一个死信队列 (DLQ) 或文件备份
        },
        Err(e) => {
            // 线程池错误 (Panic)
            error!("❌ Spawn blocking join error: {}", e);
        }
    }
}