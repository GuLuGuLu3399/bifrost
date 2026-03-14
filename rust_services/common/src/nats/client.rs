//! NATS JetStream 客户端封装
//!
//! 提供简化的 JetStream 客户端，支持发布和消费消息。

use crate::config::NatsConfig;
use anyhow::{Context, Result};
use async_nats::jetstream::{self, consumer::PullConsumer, stream::Stream};
use serde::{de::DeserializeOwned, Serialize};
use std::sync::Arc;
use tracing::{debug, info};

/// JetStream Stream 配置
pub struct StreamConfig {
    /// Stream 名称
    pub name: String,
    /// 监听的 subject 模式
    pub subjects: Vec<String>,
}

impl StreamConfig {
    pub fn new(name: impl Into<String>, subjects: Vec<String>) -> Self {
        Self {
            name: name.into(),
            subjects,
        }
    }
}

/// NATS JetStream 客户端封装
#[derive(Clone)]
pub struct NatsClient {
    client: async_nats::Client,
    jetstream: jetstream::Context,
}

impl NatsClient {
    /// 连接到 NATS 服务器
    pub async fn connect(config: &NatsConfig) -> Result<Self> {
        let client = async_nats::connect(&config.url)
            .await
            .with_context(|| format!("Failed to connect to NATS at {}", config.url))?;

        let jetstream = jetstream::new(client.clone());

        info!(url = %config.url, "Connected to NATS");

        Ok(Self { client, jetstream })
    }

    /// 获取底层 NATS 客户端
    pub fn client(&self) -> &async_nats::Client {
        &self.client
    }

    /// 获取 JetStream 上下文
    pub fn jetstream(&self) -> &jetstream::Context {
        &self.jetstream
    }

    /// 获取已存在的 Stream
    pub async fn get_stream(&self, name: &str) -> Result<Stream> {
        self.jetstream
            .get_stream(name)
            .await
            .with_context(|| format!("Failed to get stream '{}'. Make sure it exists.", name))
    }

    /// 校验 Stream 存在，不存在时自动创建。
    pub async fn ensure_stream(&self, config: &StreamConfig) -> Result<Stream> {
        match self.jetstream.get_stream(&config.name).await {
            Ok(stream) => {
                info!(stream = %config.name, "Using existing stream");
                Ok(stream)
            }
            Err(_) => {
                let stream = self
                    .jetstream
                    .create_stream(jetstream::stream::Config {
                        name: config.name.clone(),
                        subjects: config.subjects.clone(),
                        ..Default::default()
                    })
                    .await
                    .with_context(|| format!("Failed to create stream '{}'", config.name))?;

                info!(
                    stream = %config.name,
                    subjects = ?config.subjects,
                    "Created stream"
                );

                Ok(stream)
            }
        }
    }

    /// 创建 Durable Pull Consumer
    pub async fn create_pull_consumer(
        &self,
        stream_name: &str,
        consumer_name: &str,
        filter_subject: &str,
    ) -> Result<PullConsumer> {
        let stream = self.get_stream(stream_name).await?;

        let consumer = stream
            .create_consumer(jetstream::consumer::pull::Config {
                durable_name: Some(consumer_name.to_string()),
                filter_subject: filter_subject.to_string(),
                ..Default::default()
            })
            .await
            .with_context(|| {
                format!(
                    "Failed to create consumer '{}' on stream '{}'",
                    consumer_name, stream_name
                )
            })?;

        info!(
            stream = %stream_name,
            consumer = %consumer_name,
            filter = %filter_subject,
            "Created pull consumer"
        );

        Ok(consumer)
    }

    /// 发布消息到指定 subject
    pub async fn publish<T: Serialize>(&self, subject: &str, payload: &T) -> Result<()> {
        let bytes = serde_json::to_vec(payload)
            .with_context(|| "Failed to serialize message payload")?;

        self.jetstream
            .publish(subject.to_string(), bytes.into())
            .await
            .with_context(|| format!("Failed to publish to '{}'", subject))?
            .await
            .with_context(|| format!("Failed to confirm publish to '{}'", subject))?;

        debug!(subject = %subject, "Message published");

        Ok(())
    }

    /// 从消息 payload 反序列化
    pub fn deserialize<T: DeserializeOwned>(payload: &[u8]) -> Result<T> {
        serde_json::from_slice(payload).with_context(|| "Failed to deserialize message payload")
    }

    /// 包装为 Arc 便于共享
    pub fn into_arc(self) -> Arc<Self> {
        Arc::new(self)
    }
}
