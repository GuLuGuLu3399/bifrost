package service

import (
	"context"
	"time"

	contentv1 "github.com/gulugulu3399/bifrost/api/content/v1"
	nexusv1 "github.com/gulugulu3399/bifrost/api/content/v1/nexus"
	"github.com/gulugulu3399/bifrost/internal/nexus/biz"
	"github.com/gulugulu3399/bifrost/internal/pkg/xerr"
)

const (
	categoryWriteTimeout = 15 * time.Second
	categoryReadTimeout  = 5 * time.Second
)

type CategoryService struct {
	nexusv1.UnimplementedCategoryServiceServer
	useCase *biz.CategoryUseCase
}

// NewCategoryService 构造函数
func NewCategoryService(uc *biz.CategoryUseCase) *CategoryService {
	return &CategoryService{useCase: uc}
}

// CreateCategory 创建
func (s *CategoryService) CreateCategory(ctx context.Context, req *nexusv1.CreateCategoryRequest) (*nexusv1.CreateCategoryResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, categoryWriteTimeout)
	defer cancel()

	if req.GetName() == "" || req.GetSlug() == "" {
		return nil, xerr.New(xerr.CodeBadRequest, "category name and slug are required")
	}

	id, err := s.useCase.CreateCategory(ctx, &biz.CreateCategoryInput{
		Name:        req.GetName(),
		Slug:        req.GetSlug(),
		Description: req.GetDescription(),
	})
	if err != nil {
		return nil, err
	}
	return &nexusv1.CreateCategoryResponse{CategoryId: id}, nil
}

// UpdateCategory 更新
func (s *CategoryService) UpdateCategory(ctx context.Context, req *nexusv1.UpdateCategoryRequest) (*nexusv1.UpdateCategoryResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, categoryWriteTimeout)
	defer cancel()

	if req.GetCategoryId() == 0 {
		return nil, xerr.New(xerr.CodeBadRequest, "category_id is required")
	}
	if req.GetName() == "" || req.GetSlug() == "" {
		return nil, xerr.New(xerr.CodeBadRequest, "category name and slug are required")
	}

	err := s.useCase.UpdateCategory(ctx, &biz.UpdateCategoryInput{
		ID:          req.GetCategoryId(),
		Name:        req.GetName(),
		Slug:        req.GetSlug(),
		Description: req.GetDescription(),
	})
	if err != nil {
		return nil, err
	}
	return &nexusv1.UpdateCategoryResponse{Success: true}, nil
}

// DeleteCategory 删除
func (s *CategoryService) DeleteCategory(ctx context.Context, req *nexusv1.DeleteCategoryRequest) (*nexusv1.DeleteCategoryResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, categoryWriteTimeout)
	defer cancel()

	if req.GetCategoryId() == 0 {
		return nil, xerr.New(xerr.CodeBadRequest, "category_id is required")
	}

	err := s.useCase.DeleteCategory(ctx, req.GetCategoryId())
	if err != nil {
		return nil, err
	}
	return &nexusv1.DeleteCategoryResponse{Success: true}, nil
}

// ListCategories 列表 (管理端)
func (s *CategoryService) ListCategories(ctx context.Context, req *nexusv1.ListCategoriesRequest) (*nexusv1.ListCategoriesResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, categoryReadTimeout)
	defer cancel()

	cats, err := s.useCase.ListCategories(ctx)
	if err != nil {
		return nil, err
	}

	var pbCats []*contentv1.Category
	for _, c := range cats {
		pbCats = append(pbCats, &contentv1.Category{
			Id:          c.ID,
			Name:        c.Name,
			Slug:        c.Slug,
			Description: c.Description,
			PostCount:   c.PostCount,
		})
	}
	return &nexusv1.ListCategoriesResponse{Categories: pbCats}, nil
}
