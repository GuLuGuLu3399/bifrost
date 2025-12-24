package biz

import (
	"context"
	"fmt"
	"time"

	contentv1 "github.com/gulugulu3399/bifrost/api/content/v1"
	forgev1 "github.com/gulugulu3399/bifrost/api/content/v1/forge"
)

// PostUseCase 文章用例
type PostUseCase struct {
	repo PostRepo
	tx   Transaction
	// 可选的 Forge 渲染客户端（未配置则为 nil，跳过渲染）
	forgeClient forgev1.RenderServiceClient
}

func NewPostUseCase(repo PostRepo, tx Transaction, forgeClient forgev1.RenderServiceClient) *PostUseCase {
	return &PostUseCase{
		repo:        repo,
		tx:          tx,
		forgeClient: forgeClient,
	}
}

// CreatePostInput 创建文章输入
type CreatePostInput struct {
	Title       string
	Slug        string
	RawMarkdown string
	CategoryID  int64
	TagNames    []string
	Status      contentv1.PostStatus
	AuthorID    int64

	// [新增]
	CoverImageKey string
	ResourceKey   string
}

type CreatePostOutput struct {
	PostID  int64
	Version int64
}

// CreatePost 创建文章
func (uc *PostUseCase) CreatePost(ctx context.Context, input *CreatePostInput) (*CreatePostOutput, error) {
	var output *CreatePostOutput

	err := uc.tx.ExecTx(ctx, func(txCtx context.Context) error {
		// 1. 检查 slug 是否已存在
		exists, err := uc.repo.ExistsBySlug(txCtx, input.Slug)
		if err != nil {
			return err
		}
		if exists {
			return ErrSlugConflict
		}

		// 2. 构建领域实体
		now := time.Now()
		post := &Post{
			Title:         input.Title,
			Slug:          input.Slug,
			RawMarkdown:   input.RawMarkdown,
			Status:        input.Status,
			CategoryID:    input.CategoryID,
			AuthorID:      input.AuthorID,
			CoverImageKey: input.CoverImageKey,
			ResourceKey:   input.ResourceKey,
			TagNames:      input.TagNames,

			Visibility: contentv1.PostVisibility_POST_VISIBILITY_PUBLIC,
			Version:    1,
			CreatedAt:  now,
			UpdatedAt:  now,
		}

		if input.Status == contentv1.PostStatus_POST_STATUS_PUBLISHED {
			post.PublishedAt = &now
		}

		// 3. 创建文章
		postID, err := uc.repo.Create(txCtx, post)
		if err != nil {
			return err
		}
		post.ID = postID

		// 3.5 同步调用 Forge 渲染（如果配置了客户端）
		// 渲染失败将阻止文章发布，确保数据一致性
		if uc.forgeClient != nil && post.RawMarkdown != "" {
			renderCtx, cancel := context.WithTimeout(txCtx, 5*time.Second)
			defer cancel()
			
			resp, rerr := uc.forgeClient.Render(renderCtx, &forgev1.RenderRequest{RawMarkdown: post.RawMarkdown})
			if rerr != nil {
				return fmt.Errorf("forge render failed: %w", rerr)
			}
			
			// 更新渲染后的内容
			if resp != nil {
				if err := uc.repo.UpdateRenderedContent(txCtx, postID, resp.GetHtmlBody(), resp.GetTocJson(), resp.GetSummary()); err != nil {
					return fmt.Errorf("update rendered content failed: %w", err)
				}
			}
		}

		output = &CreatePostOutput{PostID: postID, Version: post.Version}
		return nil
	})

	return output, err
}

// UpdatePostInput 更新文章输入
type UpdatePostInput struct {
	ID          int64
	Title       string
	RawMarkdown string
	Status      contentv1.PostStatus

	// [新增补全] 允许修改分类和标签
	CategoryID int64
	TagNames   []string

	// [新增]
	CoverImageKey string
	ResourceKey   string
}

type UpdatePostOutput struct {
	NewVersion int64
}

// ListPostsInput 列表查询输入（可选参数）
// 说明：Go 没有真正的可选参数，常用做法是用一个输入结构体承载可选字段。
type ListPostsInput struct {
	Page       int
	PageSize   int
	Keyword    string
	CategoryID int64
	Status     contentv1.PostStatus
}

type ListPostsOutput struct {
	Posts      []*Post
	TotalCount int32
}

// UpdatePost 更新文章
func (uc *PostUseCase) UpdatePost(ctx context.Context, input *UpdatePostInput) (*UpdatePostOutput, error) {
	var output *UpdatePostOutput

	err := uc.tx.ExecTx(ctx, func(txCtx context.Context) error {
		// 1. 获取现有文章
		post, err := uc.repo.GetByID(txCtx, input.ID)
		if err != nil {
			return err
		}
		if post == nil {
			return ErrPostNotFound
		}

		// 2. 更新字段
		now := time.Now()
		post.Title = input.Title
		post.RawMarkdown = input.RawMarkdown
		post.Status = input.Status
		post.CategoryID = input.CategoryID
		post.TagNames = input.TagNames
		post.CoverImageKey = input.CoverImageKey
		post.ResourceKey = input.ResourceKey
		post.UpdatedAt = now

		if input.Status == contentv1.PostStatus_POST_STATUS_PUBLISHED && post.PublishedAt == nil {
			post.PublishedAt = &now
		}

		// 3. 保存更新
		if err := uc.repo.Update(txCtx, post); err != nil {
			return err
		}

		// 3.5 同步重新渲染（当 RawMarkdown 变更时）
		// 渲染失败将阻止更新，确保数据一致性
		if uc.forgeClient != nil && post.RawMarkdown != "" {
			renderCtx, cancel := context.WithTimeout(txCtx, 5*time.Second)
			defer cancel()
			
			resp, rerr := uc.forgeClient.Render(renderCtx, &forgev1.RenderRequest{RawMarkdown: post.RawMarkdown})
			if rerr != nil {
				return fmt.Errorf("forge render failed: %w", rerr)
			}
			
			// 更新渲染后的内容
			if resp != nil {
				if err := uc.repo.UpdateRenderedContent(txCtx, post.ID, resp.GetHtmlBody(), resp.GetTocJson(), resp.GetSummary()); err != nil {
					return fmt.Errorf("update rendered content failed: %w", err)
				}
			}
		}

		output = &UpdatePostOutput{NewVersion: post.Version}
		return nil
	})

	return output, err
}

// DeletePost 删除文章 (软删除)
func (uc *PostUseCase) DeletePost(ctx context.Context, id int64) error {
	return uc.tx.ExecTx(ctx, func(txCtx context.Context) error {
		// 1. 检查文章是否存在
		post, err := uc.repo.GetByID(txCtx, id)
		if err != nil {
			return err
		}
		if post == nil {
			return ErrPostNotFound
		}

		// 2. 软删除
		if err := uc.repo.Delete(txCtx, id); err != nil {
			return err
		}

		return nil
	})
}

// ListPosts 列出文章
func (uc *PostUseCase) ListPosts(ctx context.Context, input *ListPostsInput) (*ListPostsOutput, error) {
	// 默认值
	in := &ListPostsInput{Page: 1, PageSize: 20}
	if input != nil {
		*in = *input
		if in.Page <= 0 {
			in.Page = 1
		}
		if in.PageSize <= 0 {
			in.PageSize = 20
		}
	}

	posts, total, err := uc.repo.List(ctx, &PostListFilter{
		Page:       in.Page,
		PageSize:   in.PageSize,
		Keyword:    in.Keyword,
		CategoryID: in.CategoryID,
		Status:     in.Status,
	})
	if err != nil {
		return nil, err
	}

	return &ListPostsOutput{Posts: posts, TotalCount: total}, nil
}

// GetPost 获取文章
func (uc *PostUseCase) GetPost(ctx context.Context, id int64) (*Post, error) {
	post, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, ErrPostNotFound
	}
	return post, nil
}
