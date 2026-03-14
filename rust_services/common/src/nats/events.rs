//! NATS 事件定义
//!
//! 定义跨服务通信的事件结构，用于 JetStream 消息传递。

use serde::{Deserialize, Serialize};

// ============ 文章索引相关事件 ============

pub const SUBJECT_POST_CREATED: &str = "content.post.created";
pub const SUBJECT_POST_UPDATED: &str = "content.post.updated";
pub const SUBJECT_POST_DELETED: &str = "content.post.deleted";
pub const SUBJECT_POST_WILDCARD: &str = "content.post.>";

pub const STREAM_CONTENT: &str = "BIFROST_CONTENT";
pub const CONSUMER_MIRROR_INDEXER: &str = "mirror_indexer";

/// 文章索引事件：用于通知搜索服务索引文章
///
/// 发送方: Nexus (文章服务)
/// 接收方: Mirror (搜索服务)
/// Subject: content.post.created, content.post.updated, content.post.published
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PostEventPayload {
    /// 文章 ID (Snowflake)
    pub id: i64,
    /// 文章标题
    #[serde(default)]
    pub title: String,
    /// URL slug
    #[serde(default)]
    pub slug: String,
    /// 文章摘要 (用于索引)
    #[serde(default)]
    pub summary: String,
    /// 文章状态 (PostStatus 枚举值: 1=Draft, 2=Published, 3=Archived)
    #[serde(default)]
    pub status: i32,
    /// 发布时间 (Unix 时间戳，秒)
    #[serde(default)]
    pub published_at: i64,
}

impl PostEventPayload {
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

pub type IndexEvent = PostEventPayload;
pub type DeleteEvent = PostEventPayload;

// ============ 通用事件工具 ============

/// 从 NATS subject 中解析动作名称
///
/// 例如: "content.post.created" -> "created"
pub fn parse_action(subject: &str) -> &str {
    subject.split('.').last().unwrap_or("unknown")
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_parse_action() {
        assert_eq!(parse_action("content.post.created"), "created");
        assert_eq!(parse_action("content.post.deleted"), "deleted");
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

    #[test]
    fn test_compat_min_payload() {
        let parsed: IndexEvent = serde_json::from_str(r#"{"id": 123}"#).unwrap();
        assert_eq!(parsed.id, 123);
        assert_eq!(parsed.title, "");
        assert_eq!(parsed.published_at, 0);
    }

    #[test]
    fn test_subject_and_stream_constants() {
        assert_eq!(SUBJECT_POST_CREATED, "content.post.created");
        assert_eq!(SUBJECT_POST_UPDATED, "content.post.updated");
        assert_eq!(SUBJECT_POST_DELETED, "content.post.deleted");
        assert_eq!(SUBJECT_POST_WILDCARD, "content.post.>");
        assert_eq!(STREAM_CONTENT, "BIFROST_CONTENT");
        assert_eq!(CONSUMER_MIRROR_INDEXER, "mirror_indexer");
    }
}
