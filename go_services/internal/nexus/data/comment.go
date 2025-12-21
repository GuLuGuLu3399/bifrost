package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	contentv1 "github.com/gulugulu3399/bifrost/api/content/v1"
	"github.com/gulugulu3399/bifrost/internal/nexus/biz"
	"github.com/gulugulu3399/bifrost/internal/pkg/id"
	"github.com/gulugulu3399/bifrost/internal/pkg/observability/logger"
	"github.com/gulugulu3399/bifrost/internal/pkg/xerr"
	"github.com/jmoiron/sqlx"
)

type commentRepo struct {
	data      *Data
	snowflake *id.SnowflakeGenerator
}

func NewCommentRepo(data *Data, snowflake *id.SnowflakeGenerator) biz.CommentRepo {
	return &commentRepo{
		data:      data,
		snowflake: snowflake,
	}
}

type commentPO struct {
	ID        int64         `db:"id"`
	PostID    int64         `db:"post_id"`
	UserID    int64         `db:"user_id"`
	ParentID  sql.NullInt64 `db:"parent_id"`
	RootID    sql.NullInt64 `db:"root_id"`
	Content   string        `db:"content"`
	Status    string        `db:"status"` // string enum
	Version   int64         `db:"version"`
	CreatedAt time.Time     `db:"created_at"`
	UpdatedAt time.Time     `db:"updated_at"`
	DeletedAt sql.NullTime  `db:"deleted_at"`
}

func (po *commentPO) toEntity() *biz.Comment {
	// 简单的 Enum 转换：对未知值做兜底处理，避免 silent fallback 到 0
	var status contentv1.CommentStatus
	switch po.Status {
	case "pending":
		status = contentv1.CommentStatus_COMMENT_STATUS_PENDING
	case "approved":
		status = contentv1.CommentStatus_COMMENT_STATUS_APPROVED
	case "rejected":
		status = contentv1.CommentStatus_COMMENT_STATUS_SPAM
	default:
		status = contentv1.CommentStatus_COMMENT_STATUS_PENDING
		logger.Warn("unknown comment status in db; fallback to pending",
			logger.Int64("comment_id", po.ID),
			logger.String("status", po.Status),
		)
	}

	c := &biz.Comment{
		ID:        po.ID,
		PostID:    po.PostID,
		UserID:    po.UserID,
		Content:   po.Content,
		Status:    status,
		Version:   po.Version,
		CreatedAt: po.CreatedAt,
		UpdatedAt: po.UpdatedAt,
	}
	if po.ParentID.Valid {
		c.ParentID = po.ParentID.Int64
	}
	if po.RootID.Valid {
		c.RootID = po.RootID.Int64
	}
	if po.DeletedAt.Valid {
		t := po.DeletedAt.Time
		c.DeletedAt = &t
	}
	return c
}

func (r *commentRepo) Create(ctx context.Context, c *biz.Comment) (int64, error) {
	c.ID = r.snowflake.GenerateInt64()

	// [关键逻辑] 如果是顶层评论，RootID = 自己的 ID
	if c.RootID == 0 {
		c.RootID = c.ID
	}

	// 统一：新建评论版本号固定从 1 开始，并同步回实体
	c.Version = 1

	query := `
		INSERT INTO comments (
			id, post_id, user_id, parent_id, root_id, content, status, version, created_at, updated_at
		) VALUES (
			:id, :post_id, :user_id, :parent_id, :root_id, :content, :status, :version, :created_at, :updated_at
		)
	`

	po := &commentPO{
		ID:      c.ID,
		PostID:  c.PostID,
		UserID:  c.UserID,
		Content: c.Content,
		// Status 转 string
		Status:    "pending",
		Version:   c.Version,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
		ParentID:  int64ToNullInt64(c.ParentID),
		RootID:    int64ToNullInt64(c.RootID),
	}

	if c.Status == contentv1.CommentStatus_COMMENT_STATUS_APPROVED {
		po.Status = "approved"
	}

	db := r.data.DB(ctx)
	_, err := sqlx.NamedExecContext(ctx, db, query, po)
	if err != nil {
		return 0, xerr.Wrap(err, xerr.CodeInternal, "创建评论失败")
	}
	return c.ID, nil
}

func (r *commentRepo) Delete(ctx context.Context, id int64) error {
	query := `UPDATE comments SET deleted_at = $1, updated_at = $2 WHERE id = $3 AND deleted_at IS NULL`
	now := time.Now()
	res, err := r.data.DB(ctx).ExecContext(ctx, query, now, now, id)
	if err != nil {
		return xerr.Wrap(err, xerr.CodeInternal, "删除评论失败")
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return xerr.New(xerr.CodeNotFound, "评论不存在或已删除")
	}
	return nil
}

func (r *commentRepo) GetByID(ctx context.Context, id int64) (*biz.Comment, error) {
	// 查询时要包含 deleted_at IS NULL，但如果业务要求能看到"该评论已删除"的占位符，
	// 则 Data 层可以查出来，由 Biz 层判断。这里暂定只查未删除的。
	query := `SELECT * FROM comments WHERE id = $1 AND deleted_at IS NULL`

	var po commentPO
	db := r.data.DB(ctx)
	err := sqlx.GetContext(ctx, db, &po, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, xerr.Wrap(err, xerr.CodeInternal, "查询评论失败")
	}
	return po.toEntity(), nil
}

func (r *commentRepo) CountByPostID(ctx context.Context, postID int64) (int64, error) {
	query := `SELECT count(*) FROM comments WHERE post_id = $1 AND deleted_at IS NULL`
	var count int64
	if err := sqlx.GetContext(ctx, r.data.DB(ctx), &count, query, postID); err != nil {
		return 0, xerr.Wrap(err, xerr.CodeInternal, "统计评论数失败")
	}
	return count, nil
}
