package service

import (
	"context"
	"time"

	nexusv1 "github.com/gulugulu3399/bifrost/api/content/v1/nexus"
	"github.com/gulugulu3399/bifrost/internal/nexus/biz"
	"github.com/gulugulu3399/bifrost/internal/pkg/contextx"
	"github.com/gulugulu3399/bifrost/internal/pkg/xerr"
)

const commentWriteTimeout = 15 * time.Second

type CommentService struct {
	nexusv1.UnimplementedCommentServiceServer
	commentUC *biz.CommentUseCase
}

func NewCommentService(commentUC *biz.CommentUseCase) *CommentService {
	return &CommentService{commentUC: commentUC}
}

func (s *CommentService) CreateComment(ctx context.Context, req *nexusv1.CreateCommentRequest) (*nexusv1.CreateCommentResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, commentWriteTimeout)
	defer cancel()

	// 1. 获取当前用户
	userID := contextx.UserIDFromContext(ctx)
	if userID == 0 {
		return nil, xerr.New(xerr.CodeUnauthorized, "请先登录")
	}

	// 2. 参数校验
	if req.GetPostId() == 0 {
		return nil, xerr.New(xerr.CodeBadRequest, "post_id is required")
	}
	if len(req.GetContent()) == 0 {
		return nil, xerr.New(xerr.CodeBadRequest, "comment content cannot be empty")
	}

	// 3. 构造 Input
	input := &biz.CreateCommentInput{
		PostID:   req.GetPostId(),
		UserID:   userID,
		ParentID: req.GetParentId(),
		Content:  req.GetContent(),
	}

	// 4. 调用业务
	comment, err := s.commentUC.CreateComment(ctx, input)
	if err != nil {
		return nil, err
	}

	return &nexusv1.CreateCommentResponse{
		CommentId: comment.ID,
		Status:    comment.Status,
	}, nil
}

func (s *CommentService) DeleteComment(ctx context.Context, req *nexusv1.DeleteCommentRequest) (*nexusv1.DeleteCommentResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, commentWriteTimeout)
	defer cancel()

	userID := contextx.UserIDFromContext(ctx)
	if userID == 0 {
		return nil, xerr.New(xerr.CodeUnauthorized, "请先登录")
	}
	if req.GetCommentId() == 0 {
		return nil, xerr.New(xerr.CodeBadRequest, "comment_id is required")
	}

	isAdmin := contextx.IsAdminFromContext(ctx)

	err := s.commentUC.DeleteComment(ctx, req.GetCommentId(), userID, isAdmin)
	if err != nil {
		return nil, err
	}

	return &nexusv1.DeleteCommentResponse{Success: true}, nil
}
