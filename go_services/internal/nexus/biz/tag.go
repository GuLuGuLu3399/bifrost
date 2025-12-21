package biz

import (
	"context"
	"time"
)

// ===========================================
// 领域实体 (Domain Entity)
// ===========================================

type Tag struct {
	ID   int64
	Name string
	Slug string

	// 统计数据 (Nexus 仅读取/初始化，不负责高频更新)
	PostCount int32

	CreatedAt time.Time
	UpdatedAt time.Time
}

// ===========================================
// Outbox Event
// ===========================================

const (
	OutboxTopicTagDeleted = "tag.deleted"
)

type TagEventPayload struct {
	TagID int64  `json:"tag_id"`
	Slug  string `json:"slug"`
}

// ===========================================
// 仓储接口 (Repository Interface)
// ===========================================

type TagRepo interface {
	// GetOrCreateTags 核心方法：根据名字批量获取，不存在的自动创建
	// 用于 PostUseCase 处理文章标签关联
	GetOrCreateTags(ctx context.Context, names []string) ([]*Tag, error)

	// GetByID 获取标签
	GetByID(ctx context.Context, id int64) (*Tag, error)

	// Delete 删除标签 (同时需要清理 post_tags 关联表)
	Delete(ctx context.Context, id int64) error
}
