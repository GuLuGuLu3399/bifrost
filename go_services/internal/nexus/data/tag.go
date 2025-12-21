package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/gulugulu3399/bifrost/internal/nexus/biz"
	"github.com/gulugulu3399/bifrost/internal/pkg/id"
	"github.com/gulugulu3399/bifrost/internal/pkg/xerr"
	"github.com/jmoiron/sqlx"
)

// tagRepo 标签仓储实现
type tagRepo struct {
	data      *Data
	snowflake *id.SnowflakeGenerator
}

// NewTagRepo 创建标签仓储
func NewTagRepo(data *Data, snowflake *id.SnowflakeGenerator) biz.TagRepo {
	return &tagRepo{
		data:      data,
		snowflake: snowflake,
	}
}

// tagPO 标签持久化对象
type tagPO struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`       // NOT NULL, UNIQUE
	Slug      string    `db:"slug"`       // NOT NULL
	PostCount int32     `db:"post_count"` // NOT NULL, DEFAULT 0
	CreatedAt time.Time `db:"created_at"` // NOT NULL
	UpdatedAt time.Time `db:"updated_at"` // NOT NULL
}

// toEntity 转换为领域实体
func (po *tagPO) toEntity() *biz.Tag {
	return &biz.Tag{
		ID:        po.ID,
		Name:      po.Name,
		Slug:      po.Slug,
		PostCount: po.PostCount,
		CreatedAt: po.CreatedAt,
		UpdatedAt: po.UpdatedAt,
	}
}

// GetOrCreateTags 批量获取或创建标签
// 核心逻辑：利用 ON CONFLICT DO NOTHING 实现并发安全的 "Get-Or-Create"
func (r *tagRepo) GetOrCreateTags(ctx context.Context, names []string) ([]*biz.Tag, error) {
	if len(names) == 0 {
		return nil, nil
	}

	// 1. 内存去重 & 过滤空值
	uniqueNames := make(map[string]struct{})
	var cleanNames []string
	for _, n := range names {
		if n != "" {
			if _, ok := uniqueNames[n]; !ok {
				uniqueNames[n] = struct{}{}
				cleanNames = append(cleanNames, n)
			}
		}
	}
	if len(cleanNames) == 0 {
		return nil, nil
	}

	// 2. 准备插入数据
	// 策略：我们尝试为所有标签生成 ID 并插入。
	// 如果数据库中已存在同名标签 (ON CONFLICT)，则忽略插入，ID 也就被浪费了（Snowflake ID 廉价，允许浪费）。
	// 这样可以避免 "先查-后插" 带来的并发 Race Condition。
	var tagsToInsert []tagPO
	now := time.Now()
	for _, name := range cleanNames {
		tagsToInsert = append(tagsToInsert, tagPO{
			ID:        r.snowflake.GenerateInt64(),
			Name:      name,
			Slug:      name, // 简单处理：默认 Slug 等于 Name (复杂 Slug 逻辑建议在 Biz 层处理)
			PostCount: 0,
			CreatedAt: now,
			UpdatedAt: now,
		})
	}

	// 3. 批量插入 (忽略冲突)
	query := `
		INSERT INTO tags (id, name, slug, post_count, created_at, updated_at)
		VALUES (:id, :name, :slug, :post_count, :created_at, :updated_at)
		ON CONFLICT (name) DO NOTHING
	`

	db := r.data.DB(ctx)
	// 使用 NamedExec 进行批量插入
	_, err := sqlx.NamedExecContext(ctx, db, query, tagsToInsert)
	if err != nil {
		return nil, xerr.Wrap(err, xerr.CodeInternal, "批量创建标签失败")
	}

	// 4. 重新查询所有标签 ID
	// 因为 DO NOTHING 不会返回已存在记录的 ID，所以必须查一次全集
	// sqlx.In 用于处理 IN (?) 查询
	selectQuery, args, err := sqlx.In(`
		SELECT id, name, slug, post_count, created_at, updated_at
		FROM tags
		WHERE name IN (?)
	`, cleanNames)
	if err != nil {
		return nil, xerr.Wrap(err, xerr.CodeInternal, "构造查询参数失败")
	}

	// 这里的 Rebind 是为了适配 Postgres 的 $1, $2 占位符 (sqlx.In 默认生成 ?)
	selectQuery = db.Rebind(selectQuery)

	var pos []tagPO
	err = sqlx.SelectContext(ctx, db, &pos, selectQuery, args...)
	if err != nil {
		return nil, xerr.Wrap(err, xerr.CodeInternal, "查询标签失败")
	}

	// 5. 转换为实体列表
	var entities []*biz.Tag
	for _, po := range pos {
		entities = append(entities, po.toEntity())
	}

	return entities, nil
}

// GetByID 根据 ID 获取标签
func (r *tagRepo) GetByID(ctx context.Context, id int64) (*biz.Tag, error) {
	query := `
		SELECT id, name, slug, post_count, created_at, updated_at
		FROM tags
		WHERE id = $1
	`

	var po tagPO
	db := r.data.DB(ctx)
	err := sqlx.GetContext(ctx, db, &po, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, xerr.Wrap(err, xerr.CodeInternal, "查询标签失败")
	}

	return po.toEntity(), nil
}

// Delete 删除标签
func (r *tagRepo) Delete(ctx context.Context, id int64) error {
	// 直接硬删除，依赖数据库外键 CASCADE 清理 post_tags 表
	query := `DELETE FROM tags WHERE id = $1`

	db := r.data.DB(ctx)
	result, err := db.ExecContext(ctx, query, id)
	if err != nil {
		return xerr.Wrap(err, xerr.CodeInternal, "删除标签失败")
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return biz.ErrTagNotFound // 需要在 biz 层定义这个错误，或者复用 ErrNotFound
	}

	return nil
}
