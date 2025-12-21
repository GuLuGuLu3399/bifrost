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

type categoryRepo struct {
	data      *Data
	snowflake *id.SnowflakeGenerator
}

// NewCategoryRepo 构造函数
func NewCategoryRepo(data *Data, snowflake *id.SnowflakeGenerator) biz.CategoryRepo {
	return &categoryRepo{
		data:      data,
		snowflake: snowflake,
	}
}

// categoryPO 数据库持久化对象
type categoryPO struct {
	ID          int64          `db:"id"`
	Name        string         `db:"name"`
	Slug        string         `db:"slug"`
	Description sql.NullString `db:"description"` // Nullable
	PostCount   int32          `db:"post_count"`
	Version     int64          `db:"version"` // 乐观锁
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at"`
	// meta 字段暂时忽略，如果未来需要支持扩展属性可加上
}

// toEntity 转换为领域实体
func (po *categoryPO) toEntity() *biz.Category {
	c := &biz.Category{
		ID:        po.ID,
		Name:      po.Name,
		Slug:      po.Slug,
		PostCount: po.PostCount,
		Version:   po.Version,
		CreatedAt: po.CreatedAt,
		UpdatedAt: po.UpdatedAt,
	}
	if po.Description.Valid {
		c.Description = po.Description.String
	}
	return c
}

// Create 创建分类
func (r *categoryRepo) Create(ctx context.Context, c *biz.Category) (int64, error) {
	// 1. 生成 ID
	c.ID = r.snowflake.GenerateInt64()

	query := `
		INSERT INTO categories (
			id, name, slug, description, post_count, version, created_at, updated_at
		) VALUES (
			:id, :name, :slug, :description, :post_count, :version, :created_at, :updated_at
		)
	`

	po := &categoryPO{
		ID:          c.ID,
		Name:        c.Name,
		Slug:        c.Slug,
		Description: stringToNullString(c.Description),
		PostCount:   0,
		Version:     1, // 初始版本
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}

	db := r.data.DB(ctx)
	_, err := sqlx.NamedExecContext(ctx, db, query, po)
	if err != nil {
		return 0, xerr.Wrap(err, xerr.CodeInternal, "创建分类失败")
	}
	return c.ID, nil
}

// Update 更新分类 (带乐观锁)
// 约定：c.Version 传入的是 oldVersion（数据库当前版本），本方法负责把 version + 1。
func (r *categoryRepo) Update(ctx context.Context, c *biz.Category) error {
	query := `
		UPDATE categories 
		SET name = :name, 
		    slug = :slug, 
		    description = :description, 
		    version = :old_version + 1,
		    updated_at = :updated_at
		WHERE id = :id AND version = :old_version
	`

	po := struct {
		ID          int64          `db:"id"`
		Name        string         `db:"name"`
		Slug        string         `db:"slug"`
		Description sql.NullString `db:"description"`
		OldVersion  int64          `db:"old_version"`
		UpdatedAt   time.Time      `db:"updated_at"`
	}{
		ID:          c.ID,
		Name:        c.Name,
		Slug:        c.Slug,
		Description: stringToNullString(c.Description),
		OldVersion:  c.Version,
		UpdatedAt:   c.UpdatedAt,
	}

	db := r.data.DB(ctx)
	res, err := sqlx.NamedExecContext(ctx, db, query, po)
	if err != nil {
		return xerr.Wrap(err, xerr.CodeInternal, "更新分类失败")
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return xerr.Wrap(err, xerr.CodeInternal, "获取影响行数失败")
	}
	if rows == 0 {
		return xerr.New(xerr.CodeConflict, "更新失败，分类可能已被修改或删除")
	}

	c.Version = c.Version + 1
	return nil
}

// Delete 删除分类
func (r *categoryRepo) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM categories WHERE id = $1`

	db := r.data.DB(ctx)
	res, err := db.ExecContext(ctx, query, id)
	if err != nil {
		return xerr.Wrap(err, xerr.CodeInternal, "删除分类失败")
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return xerr.New(xerr.CodeNotFound, "分类不存在")
	}

	return nil
}

// GetByID 根据 ID 获取
func (r *categoryRepo) GetByID(ctx context.Context, id int64) (*biz.Category, error) {
	query := `SELECT * FROM categories WHERE id = $1`

	var po categoryPO
	db := r.data.DB(ctx)
	err := sqlx.GetContext(ctx, db, &po, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, xerr.Wrap(err, xerr.CodeInternal, "查询分类失败")
	}
	return po.toEntity(), nil
}

// ListAll 获取所有分类
func (r *categoryRepo) ListAll(ctx context.Context) ([]*biz.Category, error) {
	// 按创建时间倒序，方便管理
	query := `SELECT * FROM categories ORDER BY created_at DESC`

	var pos []categoryPO
	db := r.data.DB(ctx)
	err := sqlx.SelectContext(ctx, db, &pos, query)
	if err != nil {
		return nil, xerr.Wrap(err, xerr.CodeInternal, "获取分类列表失败")
	}

	var entities []*biz.Category
	for _, po := range pos {
		entities = append(entities, po.toEntity())
	}
	return entities, nil
}

// ExistsBySlug 检查 Slug 是否存在
func (r *categoryRepo) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM categories WHERE slug = $1)`

	var exists bool
	db := r.data.DB(ctx)
	err := sqlx.GetContext(ctx, db, &exists, query, slug)
	if err != nil {
		return false, xerr.Wrap(err, xerr.CodeInternal, "检查Slug失败")
	}
	return exists, nil
}
