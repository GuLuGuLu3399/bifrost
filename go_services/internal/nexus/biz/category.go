package biz

import (
	"context"
	"time"
)

// ===========================================
// 领域实体 (Domain Entity)
// ===========================================

type Category struct {
	ID          int64
	Name        string
	Slug        string
	Description string

	// 统计数据 (Nexus 通常只读，或者在删除时校验用)
	PostCount int32

	// 乐观锁版本号
	Version int64

	CreatedAt time.Time
	UpdatedAt time.Time
}

// ===========================================
// 仓储接口 (Repository Interface)
// ===========================================

// CategoryRepo 定义了对分类数据的持久化操作
// 由 internal/nexus/data/category_repo.go 实现
type CategoryRepo interface {
	// Create 创建分类
	Create(ctx context.Context, c *Category) (int64, error)

	// Update 更新分类 (Name, Slug, Description)
	Update(ctx context.Context, c *Category) error

	// Delete 删除分类
	Delete(ctx context.Context, id int64) error

	// GetByID 获取单个分类
	GetByID(ctx context.Context, id int64) (*Category, error)

	// ListAll 获取所有分类 (用于编辑器下拉列表，量级通常较小，全量返回即可)
	ListAll(ctx context.Context) ([]*Category, error)

	// ExistsBySlug 检查 Slug 是否存在 (用于创建/更新时的校验)
	ExistsBySlug(ctx context.Context, slug string) (bool, error)
}
