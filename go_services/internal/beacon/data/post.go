package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	beaconv1 "github.com/gulugulu3399/bifrost/api/content/v1/beacon"
	"github.com/gulugulu3399/bifrost/internal/beacon/biz"
	"github.com/gulugulu3399/bifrost/internal/pkg/cache"
	"github.com/jmoiron/sqlx"
)

type postRepo struct {
	data *Data
}

func NewPostRepo(data *Data) biz.PostRepo { return &postRepo{data: data} }

// 定义内部扫描用的 PO
type postDetailPO struct {
	ID            int64          `db:"id"`
	Title         string         `db:"title"`
	Slug          string         `db:"slug"`
	Summary       sql.NullString `db:"summary"`
	CoverImageKey sql.NullString `db:"cover_image_key"`
	HtmlBody      sql.NullString `db:"html_body"`
	TocJson       sql.NullString `db:"toc_json"`
	ViewCount     int32          `db:"view_count"`
	LikeCount     int32          `db:"like_count"`
	PublishedAt   sql.NullTime   `db:"published_at"`
	UpdatedAt     time.Time      `db:"updated_at"`
	AuthorID      sql.NullInt64  `db:"author_id"`
	AuthorNick    sql.NullString `db:"author_nickname"`
	AuthorAvatar  sql.NullString `db:"author_avatar_key"`
	CategoryID    sql.NullInt64  `db:"category_id"`
	CategoryName  sql.NullString `db:"category_name"`
	CategorySlug  sql.NullString `db:"category_slug"`
}

func (r *postRepo) GetPost(ctx context.Context, slug string) (*beaconv1.PostDetail, error) {
	key := KeyPostDetail(slug)

	return cache.Fetch(ctx, r.data.Cache(), key, PostCacheTTL, func() (*beaconv1.PostDetail, error) {
		var po postDetailPO
		query := `
			SELECT 
				p.id, p.title, p.slug, p.summary, p.cover_image_key, 
				p.html_body, p.toc_json, p.view_count, p.like_count, p.published_at, p.updated_at,
				u.id AS author_id, u.nickname AS author_nickname, u.avatar_key AS author_avatar_key,
				c.id AS category_id, c.name AS category_name, c.slug AS category_slug
			FROM posts p
			LEFT JOIN users u ON p.author_id = u.id
			LEFT JOIN categories c ON p.category_id = c.id
			WHERE p.slug = $1
			  AND p.status = 'published'
			  AND p.deleted_at IS NULL
		`
		if err := r.data.DB().GetContext(ctx, &po, query, slug); err != nil {
			return nil, err
		}

		// 转换 PO 为 DTO
		detail := &beaconv1.PostDetail{
			Id:            po.ID,
			Title:         po.Title,
			Slug:          po.Slug,
			Summary:       po.Summary.String,
			CoverImageKey: po.CoverImageKey.String,
			HtmlBody:      po.HtmlBody.String,
			TocJson:       po.TocJson.String,
			ViewCount:     po.ViewCount,
			LikeCount:     po.LikeCount,
			UpdatedAt:     po.UpdatedAt.Unix(),
			Author: &beaconv1.AuthorInfo{
				Id:        po.AuthorID.Int64,
				Nickname:  po.AuthorNick.String,
				AvatarKey: po.AuthorAvatar.String,
			},
		}
		if po.PublishedAt.Valid {
			detail.PublishedAt = po.PublishedAt.Time.Unix()
		}
		if po.CategoryID.Valid {
			detail.Category = &beaconv1.CategoryItem{
				Id:   po.CategoryID.Int64,
				Name: po.CategoryName.String,
				Slug: po.CategorySlug.String,
			}
		}
		return detail, nil
	})
}

type postSummaryPO struct {
	ID            int64          `db:"id"`
	Title         string         `db:"title"`
	Slug          string         `db:"slug"`
	Summary       sql.NullString `db:"summary"`
	CoverImageKey sql.NullString `db:"cover_image_key"`
	ViewCount     int32          `db:"view_count"`
	LikeCount     int32          `db:"like_count"`
	CommentCount  int32          `db:"comment_count"`
	PublishedAt   sql.NullTime   `db:"published_at"`
	AuthorID      sql.NullInt64  `db:"author_id"`
	AuthorNick    sql.NullString `db:"author_nickname"`
	AuthorAvatar  sql.NullString `db:"author_avatar_key"`
	CategoryID    sql.NullInt64  `db:"category_id"`
	CategoryName  sql.NullString `db:"category_name"`
	CategorySlug  sql.NullString `db:"category_slug"`
}

func (po *postSummaryPO) toProto() *beaconv1.PostSummary {
	s := &beaconv1.PostSummary{
		Id:            po.ID,
		Title:         po.Title,
		Slug:          po.Slug,
		Summary:       po.Summary.String,
		CoverImageKey: po.CoverImageKey.String,
		ViewCount:     po.ViewCount,
		LikeCount:     po.LikeCount,
		CommentCount:  po.CommentCount,
		Author: &beaconv1.AuthorInfo{
			Id:        po.AuthorID.Int64,
			Nickname:  po.AuthorNick.String,
			AvatarKey: po.AuthorAvatar.String,
		},
	}
	if po.PublishedAt.Valid {
		s.PublishedAt = po.PublishedAt.Time.Unix()
	}
	if po.CategoryID.Valid {
		s.Category = &beaconv1.CategoryItem{
			Id:   po.CategoryID.Int64,
			Name: po.CategoryName.String,
			Slug: po.CategorySlug.String,
		}
	}
	return s
}

// ListPosts 获取列表 (Direct DB)
func (r *postRepo) ListPosts(ctx context.Context, filter *biz.PostListFilter) ([]*beaconv1.PostSummary, int64, error) {
	if filter == nil {
		filter = &biz.PostListFilter{Page: 1, PageSize: 20}
	}
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 20
	}

	where := `WHERE p.status = 'published' AND p.deleted_at IS NULL`
	var args []any

	idx := 1
	if filter.CategoryID > 0 {
		where += ` AND p.category_id = $` + fmt.Sprint(idx)
		args = append(args, filter.CategoryID)
		idx++
	}
	if filter.AuthorID > 0 {
		where += ` AND p.author_id = $` + fmt.Sprint(idx)
		args = append(args, filter.AuthorID)
		idx++
	}

	// 1) count
	countSQL := `SELECT count(*) FROM posts p ` + where
	var total int64
	if err := r.data.DB().GetContext(ctx, &total, countSQL, args...); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []*beaconv1.PostSummary{}, 0, nil
	}

	// 2) data
	offset := (filter.Page - 1) * filter.PageSize
	args2 := append([]any{}, args...)
	args2 = append(args2, filter.PageSize, offset)

	listSQL := fmt.Sprintf(`
		SELECT
			p.id, p.title, p.slug, p.summary, p.cover_image_key,
			p.view_count, p.like_count, p.comment_count, p.published_at,
			u.id AS author_id, u.nickname AS author_nickname, u.avatar_key AS author_avatar_key,
			c.id AS category_id, c.name AS category_name, c.slug AS category_slug
		FROM posts p
		LEFT JOIN users u ON p.author_id = u.id
		LEFT JOIN categories c ON p.category_id = c.id
		%s
		ORDER BY p.published_at DESC NULLS LAST, p.id DESC
		LIMIT $%d OFFSET $%d
	`, where, idx, idx+1)

	var pos []postSummaryPO
	if err := r.data.DB().SelectContext(ctx, &pos, listSQL, args2...); err != nil {
		return nil, 0, err
	}

	items := make([]*beaconv1.PostSummary, 0, len(pos))
	for i := range pos {
		items = append(items, pos[i].toProto())
	}
	return items, total, nil
}

// BatchGetPosts 批量获取 (Search 聚合用)
func (r *postRepo) BatchGetPosts(ctx context.Context, ids []int64) ([]*beaconv1.PostSummary, error) {
	if len(ids) == 0 {
		return []*beaconv1.PostSummary{}, nil
	}

	query, args, err := sqlx.In(`
		SELECT
			p.id, p.title, p.slug, p.summary, p.cover_image_key,
			p.view_count, p.like_count, p.comment_count, p.published_at,
			u.id AS author_id, u.nickname AS author_nickname, u.avatar_key AS author_avatar_key,
			c.id AS category_id, c.name AS category_name, c.slug AS category_slug
		FROM posts p
		LEFT JOIN users u ON p.author_id = u.id
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.id IN (?)
		  AND p.status = 'published'
		  AND p.deleted_at IS NULL
	`, ids)
	if err != nil {
		return nil, err
	}
	query = r.data.DB().Rebind(query)

	var pos []postSummaryPO
	if err := r.data.DB().SelectContext(ctx, &pos, query, args...); err != nil {
		return nil, err
	}

	items := make([]*beaconv1.PostSummary, 0, len(pos))
	for i := range pos {
		items = append(items, pos[i].toProto())
	}
	return items, nil
}
