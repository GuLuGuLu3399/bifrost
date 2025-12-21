package service

import (
	"context"

	beaconv1 "github.com/gulugulu3399/bifrost/api/content/v1/beacon"
	"github.com/gulugulu3399/bifrost/internal/beacon/biz"
)

type MetaService struct {
	repo biz.MetaRepo
}

func NewMetaService(repo biz.MetaRepo) *MetaService {
	return &MetaService{repo: repo}
}

func (s *MetaService) ListCategories(ctx context.Context, req *beaconv1.ListCategoriesRequest) (*beaconv1.ListCategoriesResponse, error) {
	cats, err := s.repo.ListCategories(ctx)
	if err != nil {
		return nil, err
	}
	return &beaconv1.ListCategoriesResponse{Categories: cats}, nil
}

func (s *MetaService) ListTags(ctx context.Context, req *beaconv1.ListTagsRequest) (*beaconv1.ListTagsResponse, error) {
	tags, err := s.repo.ListTags(ctx)
	if err != nil {
		return nil, err
	}
	return &beaconv1.ListTagsResponse{Tags: tags}, nil
}
