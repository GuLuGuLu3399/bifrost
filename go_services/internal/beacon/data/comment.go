package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	beaconv1 "github.com/gulugulu3399/bifrost/api/content/v1/beacon"
	"github.com/gulugulu3399/bifrost/internal/beacon/biz"
	"github.com/gulugulu3399/bifrost/internal/pkg/xerr"
)

type commentRepo struct {
	data *Data
}

func NewCommentRepo(data *Data) biz.CommentRepo { return &commentRepo{data: data} }

// 定义扫描用的 PO
type commentItemPO struct {
	ID        int64         `db:"id"`
	Content   string        `db:"content"`
	CreatedAt time.Time     `db:"created_at"`
	ParentID  sql.NullInt64 `db:"parent_id"`
	RootID    sql.NullInt64 `db:"root_id"`
	// User Info
	UserID    int64          `db:"user_id"`
	Nickname  string         `db:"nickname"`
	AvatarKey sql.NullString `db:"avatar_key"`
}

func (r *commentRepo) ListComments(ctx context.Context, filter *biz.CommentListFilter) ([]*beaconv1.CommentItem, int64, error) {
	// 1. 先查总数
	var total int64
	countQuery := `SELECT count(*) FROM comments WHERE post_id = $1 AND status = 'approved' AND deleted_at IS NULL`
	if err := r.data.DB().GetContext(ctx, &total, countQuery, filter.PostID); err != nil {
		return nil, 0, xerr.Wrap(err, xerr.CodeInternal, "failed to count comments")
	}

	if total == 0 {
		return []*beaconv1.CommentItem{}, 0, nil
	}

	// 2. 分页参数兜底
	page := filter.Page
	pageSize := filter.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// 3. 分页查询详情（JOIN 用户信息，避免 N+1）
	base := `
		SELECT
			c.id, c.content, c.created_at, c.parent_id, c.root_id,
			u.id AS user_id, u.nickname, u.avatar_key
		FROM comments c
		LEFT JOIN users u ON c.user_id = u.id
		WHERE c.post_id = $1 AND c.status = 'approved' AND c.deleted_at IS NULL
	`

	var (
		query string
		args  []any
	)

	// 约定：RootID=0 表示查一级评论；RootID>0 表示查该楼层(root)下的回复
	if filter.RootID > 0 {
		query = fmt.Sprintf(`%s AND c.root_id = $2 ORDER BY c.created_at ASC LIMIT $3 OFFSET $4`, base)
		args = []any{filter.PostID, filter.RootID, pageSize, offset}
	} else {
		query = fmt.Sprintf(`%s AND c.parent_id IS NULL ORDER BY c.created_at DESC LIMIT $2 OFFSET $3`, base)
		args = []any{filter.PostID, pageSize, offset}
	}

	var pos []commentItemPO
	if err := r.data.DB().SelectContext(ctx, &pos, query, args...); err != nil {
		return nil, 0, xerr.Wrap(err, xerr.CodeInternal, "failed to fetch comments")
	}

	// 4. PO 转换为 DTO
	items := make([]*beaconv1.CommentItem, 0, len(pos))
	for _, po := range pos {
		var avatarKey string
		if po.AvatarKey.Valid {
			avatarKey = po.AvatarKey.String
		}

		var parentID int64
		if po.ParentID.Valid {
			parentID = po.ParentID.Int64
		}

		items = append(items, &beaconv1.CommentItem{
			Id:        po.ID,
			Content:   po.Content,
			CreatedAt: po.CreatedAt.Unix(),
			User: &beaconv1.AuthorInfo{
				Id:        po.UserID,
				Nickname:  po.Nickname,
				AvatarKey: avatarKey,
			},
			ParentId: parentID,
		})
	}

	return items, total, nil
}

func (r *commentRepo) CountByPostID(ctx context.Context, postID int64) (int64, error) {
	var count int64
	query := `SELECT count(*) FROM comments WHERE post_id = $1 AND status = 'approved' AND deleted_at IS NULL`
	err := r.data.DB().GetContext(ctx, &count, query, postID)
	return count, err
}
