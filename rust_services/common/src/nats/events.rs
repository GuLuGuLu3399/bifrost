//! NATS 事件定义
//!
//! 定义跨服务通信的事件结构，用于 JetStream 消息传递。

use serde::{Deserialize, Serialize};

// ============ 文章索引相关事件 ============

/// 文章索引事件：用于通知搜索服务索引文章
///
/// 发送方: Beacon (文章服务)
/// 接收方: Mirror (搜索服务)
/// Subject: post.created, post.updated, post.published
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct IndexEvent {
    /// 文章 ID (Snowflake)
    pub id: i64,
    /// 文章标题
    pub title: String,
    /// URL slug
    pub slug: String,
    /// 文章摘要 (用于索引)
    pub summary: String,
    /// 文章状态 (PostStatus 枚举值: 1=Draft, 2=Published, 3=Archived)
    pub status: i32,
    /// 发布时间 (Unix 时间戳，秒)
    pub published_at: i64,
}

impl IndexEvent {
    pub fn new(
        id: i64,
        title: impl Into<String>,
        slug: impl Into<String>,
        summary: impl Into<String>,
        status: i32,
        published_at: i64,
    ) -> Self {
        Self {
            id,
            title: title.into(),
            slug: slug.into(),
            summary: summary.into(),
            status,
            published_at,
        }
    }
}

/// 文章删除事件：用于通知搜索服务删除索引
///
/// 发送方: Beacon (文章服务)
/// 接收方: Mirror (搜索服务)
/// Subject: post.deleted, post.unpublished
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DeleteEvent {
    /// 文章 ID
    pub id: i64,
}

impl DeleteEvent {
    pub fn new(id: i64) -> Self {
        Self { id }
    }
}

// ============ 通用事件工具 ============

/// 从 NATS subject 中解析动作名称
///
/// 例如: "post.created" -> "created"
pub fn parse_action(subject: &str) -> &str {
    subject.split('.').last().unwrap_or("unknown")
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_parse_action() {
        assert_eq!(parse_action("post.created"), "created");
        assert_eq!(parse_action("post.deleted"), "deleted");
        assert_eq!(parse_action("single"), "single");
        assert_eq!(parse_action(""), "");
    }

    #[test]
    fn test_index_event_serde() {
        let event = IndexEvent::new(123, "Test Title", "test-slug", "Summary", 2, 1703318400);
        let json = serde_json::to_string(&event).unwrap();
        let parsed: IndexEvent = serde_json::from_str(&json).unwrap();
        assert_eq!(parsed.id, 123);
        assert_eq!(parsed.title, "Test Title");
    }
}
