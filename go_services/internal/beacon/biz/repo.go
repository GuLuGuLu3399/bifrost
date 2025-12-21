package biz

import (
	"context"

	beaconv1 "github.com/gulugulu3399/bifrost/api/content/v1/beacon"
)

// =============================================================================
// PostRepo: 文章相关
// =============================================================================

type PostListFilter struct {
	Page       int
	PageSize   int
	CategoryID int64
	TagID      int64
	AuthorID   int64
}

type PostRepo interface {
	// GetPost 获取详情 (Cache-Aside)
	GetPost(ctx context.Context, slugOrId string) (*beaconv1.PostDetail, error)
	// ListPosts 获取列表 (Direct DB)
	ListPosts(ctx context.Context, filter *PostListFilter) ([]*beaconv1.PostSummary, int64, error)
	// BatchGetPosts 批量获取 (Search 聚合用)
	BatchGetPosts(ctx context.Context, ids []int64) ([]*beaconv1.PostSummary, error)
}

// =============================================================================
// MetaRepo: 分类与标签 (Category & Tag)
// =============================================================================

type MetaRepo interface {
	// ListCategories 获取所有分类 (通常数据少，可全量缓存)
	ListCategories(ctx context.Context) ([]*beaconv1.CategoryItem, error)
	// ListTags 获取标签列表 (可能支持热门/全部)
	ListTags(ctx context.Context) ([]*beaconv1.TagItem, error)
}

// =============================================================================
// UserRepo: 用户/作者 (User)
// =============================================================================

type UserRepo interface {
	// GetUser 获取用户公开资料 (Cache-Aside)
	GetUser(ctx context.Context, userID int64) (*beaconv1.UserProfile, error)
}

// =============================================================================
// CommentRepo: 评论 (Comment)
// =============================================================================

type CommentListFilter struct {
	PostID   int64
	RootID   int64 // 0 表示查一级评论
	Page     int
	PageSize int
}

type CommentRepo interface {
	// ListComments 获取评论列表 (通常直接查库，或只缓存第一页)
	ListComments(ctx context.Context, filter *CommentListFilter) ([]*beaconv1.CommentItem, int64, error)
	// CountByPostID 获取评论总数 (辅助用)
	CountByPostID(ctx context.Context, postID int64) (int64, error)
}
