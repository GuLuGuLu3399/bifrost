package biz

import (
	"context"
	"time"

	contentv1 "github.com/gulugulu3399/bifrost/api/content/v1"
)

// ===========================================
// 领域实体 (Domain Entity)
// ===========================================

// Post 文章领域实体
type Post struct {
	// 基本信息
	ID          int64
	Title       string
	Slug        string
	Summary     string
	RawMarkdown string

	// [新增字段] 资源与封面
	CoverImageKey string
	ResourceKey   string

	// 渲染内容 (Nexus 负责存储，但通常由 Forge 回写)
	HtmlBody string
	TocJson  string

	Status     contentv1.PostStatus
	Visibility contentv1.PostVisibility
	AuthorID   int64
	CategoryID int64
	TagNames   []string // 业务层辅助字段，Repo层负责处理多对多关系
	Version    int64    // 乐观锁版本号

	// 统计 (Nexus 仅读取或忽略，不负责写入，由异步事件更新)
	ViewCount    int32
	LikeCount    int32
	CommentCount int32

	// 时间
	CreatedAt   time.Time
	UpdatedAt   time.Time
	PublishedAt *time.Time
	DeletedAt   *time.Time
}

// PostListFilter 列表查询条件
type PostListFilter struct {
	Page       int
	PageSize   int
	Keyword    string // 搜索标题或内容
	CategoryID int64
	Status     contentv1.PostStatus
}

// ===========================================
// 仓储接口
// ===========================================

type PostRepo interface {
	// Create 创建文章
	Create(ctx context.Context, post *Post) (int64, error)

	// Update 更新文章基本信息
	// [注意] 实现层应只更新: Title, Slug, Markdown, Status, Cover, Resource, UpdatedAt, Version
	// 严禁更新: ViewCount, LikeCount (防止覆盖异步统计数据)
	Update(ctx context.Context, post *Post) error

	// UpdateRenderedContent 更新渲染后的内容
	UpdateRenderedContent(ctx context.Context, id int64, htmlBody, tocJson, summary string) error

	// Delete 软删除
	Delete(ctx context.Context, id int64) error

	GetByID(ctx context.Context, id int64) (*Post, error)
	GetBySlug(ctx context.Context, slug string) (*Post, error)
	ExistsBySlug(ctx context.Context, slug string) (bool, error)

	// List 文章列表查询
	List(ctx context.Context, filter *PostListFilter) ([]*Post, int32, error)
}
