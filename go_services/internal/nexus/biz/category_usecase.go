package biz

import (
	"context"
	"time"
)

// CategoryUseCase 分类业务逻辑
type CategoryUseCase struct {
	repo CategoryRepo
	tx   Transaction
}

// NewCategoryUseCase 构造函数
func NewCategoryUseCase(repo CategoryRepo, tx Transaction) *CategoryUseCase {
	return &CategoryUseCase{repo: repo, tx: tx}
}

// ===========================================
// 创建分类 (Create)
// ===========================================

type CreateCategoryInput struct {
	Name        string
	Slug        string
	Description string
}

func (uc *CategoryUseCase) CreateCategory(ctx context.Context, input *CreateCategoryInput) (int64, error) {
	// 1. [校验] 检查 Slug 唯一性
	// 这是一个典型的 "Check-Then-Act" 场景，虽然数据库有唯一索引兜底，
	// 但业务层预检查能提供更友好的错误提示。
	exists, err := uc.repo.ExistsBySlug(ctx, input.Slug)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, ErrSlugConflict
	}

	// 2. 构建实体
	now := time.Now()
	category := &Category{
		Name:        input.Name,
		Slug:        input.Slug,
		Description: input.Description,
		PostCount:   0,
		Version:     1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// 3. 落库
	// 分类创建通常不涉及复杂的跨表事务（除非你要发 Outbox 事件通知 Search 服务立刻重建索引）
	// 这里为了简单，暂不开启显式事务，直接调用 Repo
	return uc.repo.Create(ctx, category)
}

// ===========================================
// 更新分类 (Update)
// ===========================================

type UpdateCategoryInput struct {
	ID          int64
	Name        string
	Slug        string
	Description string
}

func (uc *CategoryUseCase) UpdateCategory(ctx context.Context, input *UpdateCategoryInput) error {
	// 1. [检查] 目标分类是否存在
	old, err := uc.repo.GetByID(ctx, input.ID)
	if err != nil {
		return err
	}
	if old == nil {
		return ErrCategoryNotFound
	}

	// 2. [校验] 如果修改了 Slug，需要查重
	if input.Slug != old.Slug {
		exists, err := uc.repo.ExistsBySlug(ctx, input.Slug)
		if err != nil {
			return err
		}
		if exists {
			return ErrSlugConflict
		}
	}

	// 3. 更新实体字段
	old.Name = input.Name
	old.Slug = input.Slug
	old.Description = input.Description
	old.UpdatedAt = time.Now()

	return uc.repo.Update(ctx, old)
}

// ===========================================
// 删除分类 (Delete)
// ===========================================

func (uc *CategoryUseCase) DeleteCategory(ctx context.Context, id int64) error {
	// 1. 检查是否存在
	cat, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if cat == nil {
		return ErrCategoryNotFound
	}

	// 2. 检查是否有文章关联
	if cat.PostCount > 0 {
		return ErrCategoryNotEmpty
	}
	// 3. 执行删除
	return uc.repo.Delete(ctx, id)
}

// ===========================================
// 获取列表 (List)
// ===========================================

func (uc *CategoryUseCase) ListCategories(ctx context.Context) ([]*Category, error) {
	return uc.repo.ListAll(ctx)
}
