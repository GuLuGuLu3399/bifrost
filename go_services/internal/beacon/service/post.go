package service

import (
	"context"

	commonv1 "github.com/gulugulu3399/bifrost/api/common/v1"
	beaconv1 "github.com/gulugulu3399/bifrost/api/content/v1/beacon"
	"github.com/gulugulu3399/bifrost/internal/beacon/biz"
	"github.com/gulugulu3399/bifrost/internal/pkg/xerr"
)

type PostService struct {
	repo biz.PostRepo
}

func NewPostService(repo biz.PostRepo) *PostService {
	return &PostService{repo: repo}
}

func (s *PostService) GetPost(ctx context.Context, req *beaconv1.GetPostRequest) (*beaconv1.GetPostResponse, error) {
	post, err := s.repo.GetPost(ctx, req.GetSlugOrId())
	if err != nil {
		return nil, err
	}
	return &beaconv1.GetPostResponse{Post: post}, nil
}

func (s *PostService) ListPosts(ctx context.Context, req *beaconv1.ListPostsRequest) (*beaconv1.ListPostsResponse, error) {
	// PageRequest 语义：PageToken 作为页码/游标，这里按“页码”解析。
	page, pageSize := parsePage(req.GetPage())

	items, total, err := s.repo.ListPosts(ctx, &biz.PostListFilter{
		Page:       page,
		PageSize:   pageSize,
		CategoryID: req.GetCategoryId(),
		TagID:      req.GetTagId(),
		AuthorID:   req.GetAuthorId(),
	})
	if err != nil {
		return nil, xerr.Wrap(err, xerr.CodeInternal, "failed to list posts")
	}

	nextToken := nextPageTokenByTotal(page, pageSize, total)
	return &beaconv1.ListPostsResponse{
		Posts: items,
		Page: &commonv1.PageResponse{
			NextPageToken: nextToken,
			TotalCount:    totalToInt32(total),
		},
	}, nil
}
