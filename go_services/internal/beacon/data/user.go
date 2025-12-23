package data

import (
	"context"
	"database/sql"
	"time"

	beaconv1 "github.com/gulugulu3399/bifrost/api/content/v1/beacon"
	"github.com/gulugulu3399/bifrost/internal/beacon/biz"
	"github.com/gulugulu3399/bifrost/internal/pkg/cache"
)

type userRepo struct {
	data *Data
}

func NewUserRepo(data *Data) biz.UserRepo { return &userRepo{data: data} }

func (r *userRepo) GetUser(ctx context.Context, userID int64) (*beaconv1.UserProfile, error) {
	return cache.Fetch(ctx, r.data.Cache(), KeyUserProfile(userID), UserCacheTTL, func() (*beaconv1.UserProfile, error) {
		// 使用 PO 承载结果
		var row struct {
			ID           int64          `db:"id"`
			Username     string         `db:"username"`
			Nickname     string         `db:"nickname"`
			AvatarKey    sql.NullString `db:"avatar_key"`
			Bio          sql.NullString `db:"bio"`
			RegisteredAt time.Time      `db:"registered_at"`
			PostCount    int32          `db:"post_count"`
		}
		query := `
			SELECT id, username, nickname, avatar_key, bio, created_at as registered_at,
			(SELECT count(*) FROM posts WHERE author_id = $1 AND status = 'published') as post_count
			FROM users WHERE id = $1 AND deleted_at IS NULL
		`
		if err := r.data.DB().GetContext(ctx, &row, query, userID); err != nil {
			return nil, err
		}

		return &beaconv1.UserProfile{
			Id:        row.ID,
			Username:  row.Username,
			Nickname:  row.Nickname,
			AvatarKey: row.AvatarKey.String,
			Bio:       row.Bio.String,
			PostCount: row.PostCount,
			// 注意：RegisteredAt 字段在 proto 定义中不存在，已移除
		}, nil
	})
}
