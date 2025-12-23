use crate::engine::SearchEngine;
use common::config::NatsConfig;
use common::nats::async_nats;
use common::nats::{DeleteEvent, IndexEvent, NatsClient};
use futures::StreamExt;
use std::sync::Arc;
use tracing::{error, info, instrument, warn};

/// NATS 消息消费者，负责监听文章变动并同步到搜索索引
pub struct IndexWorker {
    engine: Arc<SearchEngine>,
    nats: NatsClient,
}

impl IndexWorker {
    pub async fn new(config: &NatsConfig, engine: Arc<SearchEngine>) -> anyhow::Result<Self> {
        let nats = NatsClient::connect(config).await?;
        Ok(Self { engine, nats })
    }

    /// 启动后台消费任务
    pub async fn run(&self) -> anyhow::Result<()> {
        // 创建 Durable Pull Consumer
        let consumer = self
            .nats
            .create_pull_consumer("BIFROST_CONTENT", "mirror_indexer", "post.>")
            .await?;

        info!("IndexWorker started. Listening on 'post.>'");

        let mut messages = consumer.messages().await?;

        while let Some(msg_result) = messages.next().await {
            match msg_result {
                Ok(msg) => {
                    let engine = Arc::clone(&self.engine);
                    // 异步处理每条消息
                    tokio::spawn(async move {
                        if let Err(e) = handle_message(engine, &msg).await {
                            error!("Failed to handle index message: {:?}", e);
                            // 处理失败时 NAK，让消息重新投递
                            if let Err(nak_err) =
                                msg.ack_with(async_nats::jetstream::AckKind::Nak(None))
                                    .await
                            {
                                error!("NAK failed: {:?}", nak_err);
                            }
                        }
                    });
                }
                Err(e) => {
                    warn!("Error receiving message: {:?}", e);
                }
            }
        }

        Ok(())
    }
}

/// 处理单条 NATS 消息
#[instrument(skip(engine, msg), fields(subject = %msg.subject))]
async fn handle_message(
    engine: Arc<SearchEngine>,
    msg: &async_nats::jetstream::Message,
) -> anyhow::Result<()> {
    // 解析 Subject 动作
    let action = common::nats::events::parse_action(&msg.subject);

    match action {
        "created" | "updated" | "published" => {
            let event: IndexEvent = NatsClient::deserialize(&msg.payload)?;

            // 使用 spawn_blocking 包装同步阻塞的索引操作
            let engine_clone = Arc::clone(&engine);

            tokio::task::spawn_blocking(move || {
                engine_clone.index_doc(
                    event.id,
                    &event.title,
                    &event.summary,
                    &event.slug,
                    event.status as i64,
                    event.published_at,
                )
            })
            .await
            .map_err(|e| anyhow::anyhow!("Task join error: {}", e))??;

            info!(post_id = event.id, "Document indexed successfully");
        }
        "deleted" | "unpublished" => {
            let event: DeleteEvent = NatsClient::deserialize(&msg.payload)?;

            // 使用 spawn_blocking 包装同步阻塞的删除操作
            let engine_clone = Arc::clone(&engine);
            let doc_id = event.id;

            tokio::task::spawn_blocking(move || engine_clone.delete_doc(doc_id))
                .await
                .map_err(|e| anyhow::anyhow!("Task join error: {}", e))??;

            info!(post_id = event.id, "Document deleted from index");
        }
        _ => {
            warn!(action = %action, "Unknown action, skipping");
        }
    }

    // 处理成功后 ACK
    msg.ack()
        .await
        .map_err(|e| anyhow::anyhow!("ACK failed: {}", e))?;

    Ok(())
}