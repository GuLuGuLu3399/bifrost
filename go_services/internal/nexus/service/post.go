package service

import (
	"context"
	"time"

	contentv1 "github.com/gulugulu3399/bifrost/api/content/v1"
	nexusv1 "github.com/gulugulu3399/bifrost/api/content/v1/nexus"
	"github.com/gulugulu3399/bifrost/internal/nexus/biz"
	"github.com/gulugulu3399/bifrost/internal/pkg/contextx"
	"github.com/gulugulu3399/bifrost/internal/pkg/messenger"
	"github.com/gulugulu3399/bifrost/internal/pkg/observability/logger"
	"github.com/gulugulu3399/bifrost/internal/pkg/xerr"
)

const postWriteTimeout = 15 * time.Second

// PostService 文章 gRPC 服务
type PostService struct {
	nexusv1.UnimplementedPostServiceServer
	postUC *biz.PostUseCase
	msgr   *messenger.Client
}

// NewPostService 创建文章服务
func NewPostService(postUC *biz.PostUseCase, msgr *messenger.Client) *PostService {
	return &PostService{postUC: postUC, msgr: msgr}
}

type postEventPayload struct {
	ID   int64  `json:"id"`
	Slug string `json:"slug"`
}

// CreatePost 创建文章
func (s *PostService) CreatePost(ctx context.Context, req *nexusv1.CreatePostRequest) (*nexusv1.CreatePostResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, postWriteTimeout)
	defer cancel()

	// 1. 从 contextx 获取真实的用户 ID
	userID := contextx.UserIDFromContext(ctx)
	if userID == 0 {
		return nil, xerr.New(xerr.CodeUnauthorized, "user not authenticated")
	}

	if req.GetTitle() == "" || req.GetSlug() == "" {
		return nil, xerr.New(xerr.CodeBadRequest, "title and slug are required")
	}
	if req.GetRawMarkdown() == "" {
		return nil, xerr.New(xerr.CodeBadRequest, "raw_markdown is required")
	}

	// structured log: record incoming request with user id
	logger.WithContext(ctx).Info("CreatePost request received", logger.Int64("user_id", userID))

	// 2. 构建输入 (DTO 转换)
	input := &biz.CreatePostInput{
		Title:       req.GetTitle(),
		Slug:        req.GetSlug(),
		RawMarkdown: req.GetRawMarkdown(),
		CategoryID:  req.GetCategoryId(),
		TagNames:    req.GetTagNames(),
		Status:      req.GetStatus(),
		AuthorID:    userID,

		// [新增映射] 把 Proto 里的字段传给 Biz
		CoverImageKey: req.GetCoverImageKey(),
		ResourceKey:   req.GetResourceKey(),
	}

	// 3. 调用用例
	output, err := s.postUC.CreatePost(ctx, input)
	if err != nil {
		return nil, err
	}

	// fire-and-forget event
	if s.msgr != nil {
		id := output.PostID
		slug := input.Slug
		go func() {
			if err := s.msgr.Publish("content.post.created", postEventPayload{ID: id, Slug: slug}); err != nil {
				logger.WithContext(ctx).Warn("publish post.created failed", logger.Err(err))
			}
		}()
	}

	// 4. 返回响应
	return &nexusv1.CreatePostResponse{
		PostId: output.PostID,
	}, nil
}

// UpdatePost 更新文章
func (s *PostService) UpdatePost(ctx context.Context, req *nexusv1.UpdatePostRequest) (*nexusv1.UpdatePostResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, postWriteTimeout)
	defer cancel()

	if req.GetPostId() == 0 {
		return nil, xerr.New(xerr.CodeBadRequest, "post_id is required")
	}
	if req.GetTitle() == "" {
		return nil, xerr.New(xerr.CodeBadRequest, "title is required")
	}
	if req.GetRawMarkdown() == "" {
		return nil, xerr.New(xerr.CodeBadRequest, "raw_markdown is required")
	}

	// 构建输入 (DTO 转换)
	input := &biz.UpdatePostInput{
		ID:          req.GetPostId(),
		Title:       req.GetTitle(),
		RawMarkdown: req.GetRawMarkdown(),
		Status:      req.GetStatus(),

		// [新增映射] 允许更新封面和资源路径
		CoverImageKey: req.GetCoverImageKey(),
		ResourceKey:   req.GetResourceKey(),
	}

	output, err := s.postUC.UpdatePost(ctx, input)
	if err != nil {
		return nil, err
	}

	if s.msgr != nil {
		id := input.ID
		go func() {
			if err := s.msgr.Publish("content.post.updated", postEventPayload{ID: id, Slug: ""}); err != nil {
				logger.WithContext(ctx).Warn("publish post.updated failed", logger.Err(err))
			}
		}()
	}

	return &nexusv1.UpdatePostResponse{
		Version: output.NewVersion,
	}, nil
}

// UpdatePostRenderedContent 更新渲染后的内容 (内部接口)
func (s *PostService) UpdatePostRenderedContent(ctx context.Context, req *nexusv1.UpdatePostRenderedContentRequest) (*nexusv1.UpdatePostRenderedContentResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, postWriteTimeout)
	defer cancel()

	if req.GetPostId() == 0 {
		return nil, xerr.New(xerr.CodeBadRequest, "post_id is required")
	}

	err := s.postUC.UpdateRenderedContent(
		ctx,
		req.GetPostId(),
		req.GetHtmlBody(),
		req.GetTocJson(),
		req.GetSummary(),
	)
	if err != nil {
		return nil, err
	}

	return &nexusv1.UpdatePostRenderedContentResponse{
		Success: true,
	}, nil
}

// DeletePost 删除文章
func (s *PostService) DeletePost(ctx context.Context, req *nexusv1.DeletePostRequest) (*nexusv1.DeletePostResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, postWriteTimeout)
	defer cancel()

	if req.GetPostId() == 0 {
		return nil, xerr.New(xerr.CodeBadRequest, "post_id is required")
	}

	err := s.postUC.DeletePost(ctx, req.GetPostId())
	if err != nil {
		return nil, err
	}

	if s.msgr != nil {
		id := req.GetPostId()
		go func() {
			if err := s.msgr.Publish("content.post.deleted", postEventPayload{ID: id, Slug: ""}); err != nil {
				logger.WithContext(ctx).Warn("publish post.deleted failed", logger.Err(err))
			}
		}()
	}

	return &nexusv1.DeletePostResponse{
		Success: true,
	}, nil
}

// ListPosts 列出文章 (简要信息)
func (s *PostService) ListPosts(ctx context.Context, req *nexusv1.ListPostsRequest) (*nexusv1.ListPostsResponse, error) {
	out, err := s.postUC.ListPosts(ctx, &biz.ListPostsInput{
		Page:       int(req.GetPage()),
		PageSize:   int(req.GetPageSize()),
		Keyword:    req.GetKeyword(),
		CategoryID: req.GetCategoryId(),
		Status:     req.GetStatus(),
	})
	if err != nil {
		return nil, err
	}

	// 转换为 Proto 列表
	protoPosts := make([]*contentv1.Post, 0, len(out.Posts))
	for _, post := range out.Posts {
		protoPosts = append(protoPosts, &contentv1.Post{
			Id:            post.ID,
			Title:         post.Title,
			Slug:          post.Slug,
			Summary:       post.Summary,
			CoverImageKey: post.CoverImageKey,
			ResourceKey:   post.ResourceKey,
		})
	}

	return &nexusv1.ListPostsResponse{
		Posts:      protoPosts,
		TotalCount: out.TotalCount,
	}, nil
}

// GetPost 获取文章 (回显用)
func (s *PostService) GetPost(ctx context.Context, req *nexusv1.GetPostRequest) (*nexusv1.GetPostResponse, error) {
	post, err := s.postUC.GetPost(ctx, req.GetPostId())
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, xerr.New(xerr.CodeNotFound, "post not found")
	}

	// 将 Biz 实体转换为 Proto Message
	return &nexusv1.GetPostResponse{
		Post: &contentv1.Post{
			Id:            post.ID,
			Title:         post.Title,
			Slug:          post.Slug,
			Summary:       post.Summary,
			RawMarkdown:   post.RawMarkdown,
			CoverImageKey: post.CoverImageKey,
			ResourceKey:   post.ResourceKey,
		},
	}, nil
}
