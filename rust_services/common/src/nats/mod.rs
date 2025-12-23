//! NATS JetStream 通用封装
//!
//! 提供跨服务的 NATS 客户端、事件定义和消息处理工具。

mod client;
pub mod events;

pub use async_nats;
pub use client::{NatsClient, StreamConfig};
pub use events::{DeleteEvent, IndexEvent};
