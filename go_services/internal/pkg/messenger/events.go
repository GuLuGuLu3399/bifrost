package messenger

// 事件载荷定义 (建议按服务分组定义)
// 这些结构体应该与 Rust 服务保持兼容

// PostEventPayload 文章事件载荷
type PostEventPayload struct {
	ID          int64  `json:"id"`
	Slug        string `json:"slug"`
	Title       string `json:"title,omitempty"`
	Summary     string `json:"summary,omitempty"`
	Status      int32  `json:"status,omitempty"`
	PublishedAt int64  `json:"published_at,omitempty"`
}

// CategoryEventPayload 分类事件载荷
type CategoryEventPayload struct {
	ID   int64  `json:"id"`
	Slug string `json:"slug"`
}

// TagEventPayload 标签事件载荷
type TagEventPayload struct {
	ID   int64  `json:"id"`
	Slug string `json:"slug"`
}

// CommentEventPayload 评论事件载荷
type CommentEventPayload struct {
	ID     int64 `json:"id"`
	PostID int64 `json:"post_id"`
	UserID int64 `json:"user_id"`
}

// InteractionEventPayload 用户互动事件
type InteractionEventPayload struct {
	UserID int64  `json:"user_id"`
	PostID int64  `json:"post_id"`
	Type   string `json:"type"` // "like", "bookmark", "share"
}

// 常见事件主题常量
const (
	SubjectPostCreated    = "content.post.created"
	SubjectPostUpdated    = "content.post.updated"
	SubjectPostDeleted    = "content.post.deleted"
	SubjectPostWildcard   = "content.post.>"
	SubjectContentAll     = "content.>"
	SubjectCategoryUpdate = "content.category.updated"
	SubjectTagUpdate      = "content.tag.updated"
	SubjectCommentCreated = "content.comment.created"
	SubjectInteraction    = "content.interaction.updated"

	// NATS Stream / Consumer 常量
	StreamContent        = "BIFROST_CONTENT"
	ConsumerMirrorIndexer = "mirror_indexer"

	// Queue Group 常量
	GroupBeacon = "beacon_service"
	GroupMirror = "mirror_service"
	GroupAudit  = "audit_service" // 未来服务
)
