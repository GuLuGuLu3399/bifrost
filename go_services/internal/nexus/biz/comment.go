package biz

import (
	"context"
	"time"

	contentv1 "github.com/gulugulu3399/bifrost/api/content/v1"
)

// ===========================================
// 领域实体 (Domain Entity)
// ===========================================

type Comment struct {
	ID       int64
	PostID   int64
	UserID   int64
	ParentID int64 // 0 表示顶层评论
	RootID   int64 // 顶层评论的 RootID 等于自身 ID
	Content  string
	Status   contentv1.CommentStatus

	// 乐观锁
	Version int64

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// ===========================================
// 仓储接口
// ===========================================

type CommentRepo interface {
	// Create 创建评论
	Create(ctx context.Context, c *Comment) (int64, error)

	// Delete 软删除 (需要校验权限)
	Delete(ctx context.Context, id int64) error

	// GetByID 获取单个评论 (用于回复时计算 RootID 或删除校验)
	GetByID(ctx context.Context, id int64) (*Comment, error)

	// CountByPostID 统计文章评论数 (可选，用于辅助校验)
	CountByPostID(ctx context.Context, postID int64) (int64, error)
}
