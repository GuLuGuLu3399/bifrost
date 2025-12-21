package service

import (
	"context"
	"strconv"

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
	page := 1
	pageSize := 20
	var nextToken string

	if req.GetPage() != nil {
		if req.GetPage().GetPageSize() > 0 {
			pageSize = int(req.GetPage().GetPageSize())
		}
		if tok := req.GetPage().GetPageToken(); tok != "" {
			if p, err := strconv.Atoi(tok); err == nil && p > 0 {
				page = p
			}
		}
	}

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

	// next_page_token：简单页码策略。没有更多数据就返回空。
	if int64(page*pageSize) < total {
		nextToken = strconv.Itoa(page + 1)
	}

	tc := int32(total)
	if total > int64(^uint32(0)>>1) {
		// 防溢出：total 超过 int32 最大值时，PageResponse.TotalCount 置为最大值
		tc = int32(^uint32(0) >> 1)
	}

	return &beaconv1.ListPostsResponse{
		Posts: items,
		Page: &commonv1.PageResponse{
			NextPageToken: nextToken,
			TotalCount:    tc,
		},
	}, nil
}
