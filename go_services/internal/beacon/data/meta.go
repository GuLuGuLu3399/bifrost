package data

import (
	"context"

	beaconv1 "github.com/gulugulu3399/bifrost/api/content/v1/beacon"
	"github.com/gulugulu3399/bifrost/internal/beacon/biz"
	"github.com/gulugulu3399/bifrost/internal/pkg/cache"
)

type metaRepo struct {
	data *Data
}

func NewMetaRepo(data *Data) biz.MetaRepo {
	return &metaRepo{data: data}
}

// ListCategories 获取所有分类 (带缓存)
func (r *metaRepo) ListCategories(ctx context.Context) ([]*beaconv1.CategoryItem, error) {
	res, err := cache.Fetch(ctx, r.data.Cache(), KeyCategoryList(), MetaCacheTTL, func() (*[]*beaconv1.CategoryItem, error) {
		type categoryPO struct {
			ID          int64  `db:"id"`
			Name        string `db:"name"`
			Slug        string `db:"slug"`
			Description string `db:"description"`
			PostCount   int32  `db:"post_count"`
		}

		var pos []categoryPO
		query := `
					SELECT id, name, slug, COALESCE(description, '') AS description, post_count
					FROM categories
					ORDER BY id
				`
		if err := r.data.DB().SelectContext(ctx, &pos, query); err != nil {
			return nil, err
		}

		items := make([]*beaconv1.CategoryItem, 0, len(pos))
		for _, po := range pos {
			items = append(items, &beaconv1.CategoryItem{
				Id:          po.ID,
				Name:        po.Name,
				Slug:        po.Slug,
				Description: po.Description,
				PostCount:   po.PostCount,
			})
		}

		return &items, nil
	})
	if err != nil {
		return nil, err
	}
	return *res, nil
}

// ListTags 获取热门标签 (带缓存)
func (r *metaRepo) ListTags(ctx context.Context) ([]*beaconv1.TagItem, error) {
	res, err := cache.Fetch(ctx, r.data.Cache(), KeyTagList(), MetaCacheTTL, func() (*[]*beaconv1.TagItem, error) {
		type tagPO struct {
			ID        int64  `db:"id"`
			Name      string `db:"name"`
			Slug      string `db:"slug"`
			PostCount int32  `db:"post_count"`
		}

		var pos []tagPO
		query := `
			SELECT id, name, slug, post_count
			FROM tags
			WHERE post_count > 0
			ORDER BY post_count DESC
			LIMIT 50
		`
		if err := r.data.DB().SelectContext(ctx, &pos, query); err != nil {
			return nil, err
		}

		items := make([]*beaconv1.TagItem, 0, len(pos))
		for _, po := range pos {
			items = append(items, &beaconv1.TagItem{
				Id:        po.ID,
				Name:      po.Name,
				Slug:      po.Slug,
				PostCount: po.PostCount,
			})
		}

		return &items, nil
	})
	if err != nil {
		return nil, err
	}
	return *res, nil
}
