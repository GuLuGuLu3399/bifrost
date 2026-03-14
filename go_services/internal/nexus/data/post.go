package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	contentv1 "github.com/gulugulu3399/bifrost/api/content/v1"
	"github.com/gulugulu3399/bifrost/internal/nexus/biz"
	"github.com/gulugulu3399/bifrost/internal/pkg/id"
	"github.com/gulugulu3399/bifrost/internal/pkg/xerr"
	"github.com/jmoiron/sqlx"
)

// postRepo 文章仓储实现
type postRepo struct {
	data      *Data
	snowflake *id.SnowflakeGenerator
}

// NewPostRepo 创建文章仓储
func NewPostRepo(data *Data, snowflake *id.SnowflakeGenerator) biz.PostRepo {
	return &postRepo{
		data:      data,
		snowflake: snowflake,
	}
}

// postPO 文章持久化对象 (Persistence Object)
// 字段类型严格对应数据库 Schema：
// - NOT NULL 字段使用 Go 原生类型 (string, int64 等)
// - NULLABLE 字段使用 sql.Null* 类型
type postPO struct {
	ID          int64  `db:"id"`
	Title       string `db:"title"`        // NOT NULL
	Slug        string `db:"slug"`         // NOT NULL
	RawMarkdown string `db:"raw_markdown"` // NOT NULL

	// NULLABLE 字段
	Summary  sql.NullString `db:"summary"`
	ResourceKey sql.NullString `db:"resource_key"`
	CoverImageKey sql.NullString `db:"cover_image_key"`
	HtmlBody sql.NullString `db:"html_body"`
	TocJson  sql.NullString `db:"toc_json"`

	Status       string        `db:"status"`      // NOT NULL (schema: VARCHAR)
	Visibility   string        `db:"visibility"`  // NOT NULL (schema: VARCHAR)
	AuthorID     int64         `db:"author_id"`   // NOT NULL
	CategoryID   sql.NullInt64 `db:"category_id"` // NULLABLE (FK)
	Version      int64         `db:"version"`     // NOT NULL
	ViewCount    int32         `db:"view_count"`
	LikeCount    int32         `db:"like_count"`
	CommentCount int32         `db:"comment_count"`
	CreatedAt    time.Time     `db:"created_at"` // NOT NULL
	UpdatedAt    time.Time     `db:"updated_at"` // NOT NULL
	PublishedAt  sql.NullTime  `db:"published_at"`
	DeletedAt    sql.NullTime  `db:"deleted_at"`
}

// toEntity 转换为领域实体
func (po *postPO) toEntity() *biz.Post {
	post := &biz.Post{
		ID:           po.ID,
		Title:        po.Title,
		Slug:         po.Slug,
		RawMarkdown:  po.RawMarkdown,
		Status:       dbStatusToProto(po.Status),
		Visibility:   dbVisibilityToProto(po.Visibility),
		AuthorID:     po.AuthorID,
		Version:      po.Version,
		ViewCount:    po.ViewCount,
		LikeCount:    po.LikeCount,
		CommentCount: po.CommentCount,
		CreatedAt:    po.CreatedAt,
		UpdatedAt:    po.UpdatedAt,
	}

	// 处理 NULLABLE 字段
	post.Summary = nullString(po.Summary)
	post.ResourceKey = nullString(po.ResourceKey)
	post.CoverImageKey = nullString(po.CoverImageKey)
	post.HtmlBody = nullString(po.HtmlBody)
	post.TocJson = nullString(po.TocJson)
	post.CategoryID = nullInt64(po.CategoryID)
	post.PublishedAt = nullTime(po.PublishedAt)
	post.DeletedAt = nullTime(po.DeletedAt)
	return post
}

// Create 创建文章
func (r *postRepo) Create(ctx context.Context, post *biz.Post) (int64, error) {
	postID := r.snowflake.GenerateInt64()

	query := `
		INSERT INTO posts (
			id, title, slug, summary, raw_markdown, html_body, toc_json,
			status, visibility, author_id, category_id, version,
			view_count, like_count, comment_count,
			created_at, updated_at, published_at
		) VALUES (
			:id, :title, :slug, :summary, :raw_markdown, :html_body, :toc_json,
			:status, :visibility, :author_id, :category_id, :version,
			:view_count, :like_count, :comment_count,
			:created_at, :updated_at, :published_at
		)
	`

	// 构造 PO，NOT NULL 字段直接赋值
	po := &postPO{
		ID:           postID,
		Title:        post.Title,
		Slug:         post.Slug,
		RawMarkdown:  post.RawMarkdown, // NOT NULL，直接赋值
		Status:       protoStatusToDB(post.Status),
		Visibility:   protoVisibilityToDB(post.Visibility),
		AuthorID:     post.AuthorID,
		Version:      post.Version,
		ViewCount:    0,
		LikeCount:    0,
		CommentCount: 0,
		CreatedAt:    post.CreatedAt,
		UpdatedAt:    post.UpdatedAt,
	}

	// 处理 NULLABLE 字段
	po.Summary = stringToNullString(post.Summary)
	po.ResourceKey = stringToNullString(post.ResourceKey)
	po.CoverImageKey = stringToNullString(post.CoverImageKey)
	po.CategoryID = int64ToNullInt64(post.CategoryID)
	po.PublishedAt = ptrTimeToNullTime(post.PublishedAt)

	db := r.data.DB(ctx)
	_, err := sqlx.NamedExecContext(ctx, db, query, po)
	if err != nil {
		return 0, xerr.Wrap(err, xerr.CodeInternal, "创建文章失败")
	}

	return postID, nil
}

// List 获取文章列表
func (r *postRepo) List(ctx context.Context, filter *biz.PostListFilter) ([]*biz.Post, int32, error) {
	baseQuery := `FROM posts WHERE deleted_at IS NULL`
	args := map[string]any{}

	if filter.Keyword != "" {
		baseQuery += ` AND (title ILIKE :keyword OR raw_markdown ILIKE :keyword)`
		args["keyword"] = "%" + filter.Keyword + "%"
	}

	if filter.CategoryID > 0 {
		baseQuery += ` AND category_id = :category_id`
		args["category_id"] = filter.CategoryID
	}

	if filter.Status != contentv1.PostStatus_POST_STATUS_UNSPECIFIED {
		baseQuery += ` AND status = :status`
		args["status"] = protoStatusToDB(filter.Status)
	}

	// 3. 查询总数
	countQuery := `SELECT count(*) ` + baseQuery
	var total64 int64

	countSQL, countArgs, err := sqlx.Named(countQuery, args)
	if err != nil {
		return nil, 0, xerr.Wrap(err, xerr.CodeInternal, "构建 Count 语句失败")
	}
	countSQL = sqlx.Rebind(sqlx.DOLLAR, countSQL)

	db := r.data.DB(ctx)
	if err := sqlx.GetContext(ctx, db, &total64, countSQL, countArgs...); err != nil {
		return nil, 0, xerr.Wrap(err, xerr.CodeInternal, "查询总数失败")
	}

	// int64 -> int32（保护一下，避免极端情况下溢出）
	var total int32
	if total64 > int64(^uint32(0)>>1) {
		total = int32(^uint32(0) >> 1)
	} else {
		total = int32(total64)
	}

	// 4. 查询数据
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 20
	}
	args["limit"] = filter.PageSize
	args["offset"] = (filter.Page - 1) * filter.PageSize

	selectQuery := `
		SELECT id, title, slug, summary, resource_key, cover_image_key, raw_markdown, html_body, toc_json,
		       status, visibility, author_id, category_id, version,
		       view_count, like_count, comment_count,
		       created_at, updated_at, published_at, deleted_at
	` + baseQuery + ` ORDER BY created_at DESC LIMIT :limit OFFSET :offset`
	selectSQL, selectArgs, err := sqlx.Named(selectQuery, args)
	if err != nil {
		return nil, 0, xerr.Wrap(err, xerr.CodeInternal, "构建 Select 语句失败")
	}
	selectSQL = sqlx.Rebind(sqlx.DOLLAR, selectSQL)

	var pos []postPO
	if err := sqlx.SelectContext(ctx, db, &pos, selectSQL, selectArgs...); err != nil {
		return nil, 0, xerr.Wrap(err, xerr.CodeInternal, "查询文章列表失败")
	}

	entities := make([]*biz.Post, 0, len(pos))
	for _, po := range pos {
		po := po
		entities = append(entities, po.toEntity())
	}

	return entities, total, nil
}

// updatePostPO 用于 Update 操作的 PO，包含乐观锁所需的 OldVersion 字段
type updatePostPO struct {
	ID          int64        `db:"id"`
	Title       string       `db:"title"`        // NOT NULL
	RawMarkdown string       `db:"raw_markdown"` // NOT NULL
	Status      string       `db:"status"`
	OldVersion  int64        `db:"old_version"` // 乐观锁校验用的旧版本号
	UpdatedAt   time.Time    `db:"updated_at"`
	PublishedAt sql.NullTime `db:"published_at"` // NULLABLE
}

// Update 更新文章
// 约定：post.Version 传入的是 oldVersion（数据库当前版本），本方法负责把 version + 1。
func (r *postRepo) Update(ctx context.Context, post *biz.Post) error {
	query := `
		UPDATE posts SET
			title = :title,
			raw_markdown = :raw_markdown,
			status = :status,
			version = :old_version + 1,
			updated_at = :updated_at,
			published_at = :published_at
		WHERE id = :id 
		  AND version = :old_version
		  AND deleted_at IS NULL
	`

	po := updatePostPO{
		ID:          post.ID,
		Title:       post.Title,
		RawMarkdown: post.RawMarkdown,
		Status:      protoStatusToDB(post.Status),
		OldVersion:  post.Version,
		UpdatedAt:   post.UpdatedAt,
		PublishedAt: ptrTimeToNullTime(post.PublishedAt),
	}

	db := r.data.DB(ctx)
	result, err := sqlx.NamedExecContext(ctx, db, query, po)
	if err != nil {
		return xerr.Wrap(err, xerr.CodeInternal, "更新文章失败")
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return xerr.Wrap(err, xerr.CodeInternal, "获取影响行数失败")
	}

	if affected == 0 {
		return biz.ErrVersionConflict
	}

	post.Version = post.Version + 1
	return nil
}

// UpdateRenderedContent 更新渲染后的内容
// 此方法在 CreatePost/UpdatePost 事务中被调用，用于存储 Forge 渲染的 HTML 结果
func (r *postRepo) UpdateRenderedContent(ctx context.Context, id int64, htmlBody, tocJson, summary string) error {
	query := `
		UPDATE posts SET
			html_body = $1,
			toc_json = $2,
			summary = $3,
			updated_at = $4
		WHERE id = $5 AND deleted_at IS NULL
	`

	db := r.data.DB(ctx)
	result, err := db.ExecContext(ctx, query, htmlBody, tocJson, summary, time.Now(), id)
	if err != nil {
		return xerr.Wrap(err, xerr.CodeInternal, "更新渲染内容失败")
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return biz.ErrPostNotFound
	}

	return nil
}

// Delete 软删除文章
func (r *postRepo) Delete(ctx context.Context, id int64) error {
	query := `
		UPDATE posts SET
			deleted_at = $1,
			updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL
	`

	now := time.Now()
	db := r.data.DB(ctx)
	result, err := db.ExecContext(ctx, query, now, now, id)
	if err != nil {
		return xerr.Wrap(err, xerr.CodeInternal, "删除文章失败")
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return biz.ErrPostNotFound
	}

	return nil
}

// GetByID 根据 ID 获取文章
func (r *postRepo) GetByID(ctx context.Context, id int64) (*biz.Post, error) {
	query := `
		SELECT id, title, slug, summary, raw_markdown, html_body, toc_json,
			   status, visibility, author_id, category_id, version,
			   view_count, like_count, comment_count,
			   created_at, updated_at, published_at, deleted_at
		FROM posts
		WHERE id = $1 AND deleted_at IS NULL
	`

	var po postPO
	db := r.data.DB(ctx)
	err := sqlx.GetContext(ctx, db, &po, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, xerr.Wrap(err, xerr.CodeInternal, "查询文章失败")
	}

	return po.toEntity(), nil
}

// GetBySlug 根据 Slug 获取文章
func (r *postRepo) GetBySlug(ctx context.Context, slug string) (*biz.Post, error) {
	query := `
		SELECT id, title, slug, summary, raw_markdown, html_body, toc_json,
			   status, visibility, author_id, category_id, version,
			   view_count, like_count, comment_count,
			   created_at, updated_at, published_at, deleted_at
		FROM posts
		WHERE slug = $1 AND deleted_at IS NULL
	`

	var po postPO
	db := r.data.DB(ctx)
	err := sqlx.GetContext(ctx, db, &po, query, slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, xerr.Wrap(err, xerr.CodeInternal, "查询文章失败")
	}

	return po.toEntity(), nil
}

// ExistsBySlug 检查 Slug 是否存在
func (r *postRepo) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM posts WHERE slug = $1 AND deleted_at IS NULL)`

	var exists bool
	db := r.data.DB(ctx)
	err := sqlx.GetContext(ctx, db, &exists, query, slug)
	if err != nil {
		return false, xerr.Wrap(err, xerr.CodeInternal, "检查 slug 失败")
	}

	return exists, nil
}

func protoStatusToDB(status contentv1.PostStatus) string {
	switch status {
	case contentv1.PostStatus_POST_STATUS_PUBLISHED:
		return "published"
	case contentv1.PostStatus_POST_STATUS_ARCHIVED:
		return "archived"
	case contentv1.PostStatus_POST_STATUS_DRAFT:
		fallthrough
	default:
		return "draft"
	}
}

func dbStatusToProto(status string) contentv1.PostStatus {
	switch status {
	case "published":
		return contentv1.PostStatus_POST_STATUS_PUBLISHED
	case "archived":
		return contentv1.PostStatus_POST_STATUS_ARCHIVED
	case "draft":
		fallthrough
	default:
		return contentv1.PostStatus_POST_STATUS_DRAFT
	}
}

func protoVisibilityToDB(v contentv1.PostVisibility) string {
	switch v {
	case contentv1.PostVisibility_POST_VISIBILITY_HIDDEN:
		return "hidden"
	case contentv1.PostVisibility_POST_VISIBILITY_PASSWORD:
		return "password"
	case contentv1.PostVisibility_POST_VISIBILITY_PUBLIC:
		fallthrough
	default:
		return "public"
	}
}

func dbVisibilityToProto(v string) contentv1.PostVisibility {
	switch v {
	case "hidden":
		return contentv1.PostVisibility_POST_VISIBILITY_HIDDEN
	case "password":
		return contentv1.PostVisibility_POST_VISIBILITY_PASSWORD
	case "public":
		fallthrough
	default:
		return contentv1.PostVisibility_POST_VISIBILITY_PUBLIC
	}
}
