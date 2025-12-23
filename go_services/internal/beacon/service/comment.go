package service

import (
	"context"

	commonv1 "github.com/gulugulu3399/bifrost/api/common/v1"
	beaconv1 "github.com/gulugulu3399/bifrost/api/content/v1/beacon"
	"github.com/gulugulu3399/bifrost/internal/beacon/biz"
	"github.com/gulugulu3399/bifrost/internal/pkg/xerr"
)

type CommentService struct {
	repo biz.CommentRepo
}

func NewCommentService(repo biz.CommentRepo) *CommentService {
	return &CommentService{repo: repo}
}

// ListComments 获取评论列表
func (s *CommentService) ListComments(ctx context.Context, req *beaconv1.ListCommentsRequest) (*beaconv1.ListCommentsResponse, error) {
	if req.GetPostId() == 0 {
		return nil, xerr.New(xerr.CodeBadRequest, "post_id is required")
	}

	page, pageSize := parsePage(req.GetPage())

	items, total, err := s.repo.ListComments(ctx, &biz.CommentListFilter{
		PostID:   req.GetPostId(),
		RootID:   0, // beacon.proto 当前不支持 root_id，默认返回一级评论
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, err
	}

	nextToken := nextPageTokenByTotal(page, pageSize, total)
	return &beaconv1.ListCommentsResponse{
		Comments: items,
		Page: &commonv1.PageResponse{
			NextPageToken: nextToken,
			TotalCount:    totalToInt32(total),
		},
	}, nil
}
