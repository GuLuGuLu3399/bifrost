package biz

import (
	"context"
	"time"

	contentv1 "github.com/gulugulu3399/bifrost/api/content/v1"
)

type CommentUseCase struct {
	repo     CommentRepo
	postRepo PostRepo // 需要检查文章是否存在
	tx       Transaction
}

func NewCommentUseCase(repo CommentRepo, postRepo PostRepo, tx Transaction) *CommentUseCase {
	return &CommentUseCase{
		repo:     repo,
		postRepo: postRepo,
		tx:       tx,
	}
}

// CreateCommentInput 创建输入
type CreateCommentInput struct {
	PostID   int64
	UserID   int64
	ParentID int64 // 可选，0 表示盖楼
	Content  string
}

func (uc *CommentUseCase) CreateComment(ctx context.Context, input *CreateCommentInput) (*Comment, error) {
	var comment *Comment

	err := uc.tx.ExecTx(ctx, func(txCtx context.Context) error {
		// 1. 检查文章是否存在 (Fail fast)
		// 这一步可以用 Cache 优化，但 Nexus 写频低，直接查库更安全
		post, err := uc.postRepo.GetByID(txCtx, input.PostID)
		if err != nil {
			return err
		}
		if post == nil {
			return ErrPostNotFound
		}

		// 2. 准备实体
		now := time.Now()
		comment = &Comment{
			PostID:    input.PostID,
			UserID:    input.UserID,
			ParentID:  input.ParentID,
			Content:   input.Content,
			Status:    contentv1.CommentStatus_COMMENT_STATUS_PENDING, // 默认 Pending，等待审核
			Version:   1,
			CreatedAt: now,
			UpdatedAt: now,
		}

		// 3. 计算 RootID (关键逻辑)
		if input.ParentID == 0 {
			// A. 顶层评论：RootID 暂定为 0，等生成 ID 后回填，或者如果 ID 是预生成的，可以直接填
			// 在 Data 层，我们通常会先生成 ID，所以这里留给 Data 层处理，
			// 或者我们约定：ParentID=0 时，Data 层把 RootID 设为 ID。
			comment.RootID = 0
		} else {
			// B. 回复评论：必须查出父评论的 RootID
			parent, err := uc.repo.GetByID(txCtx, input.ParentID)
			if err != nil {
				return err
			}
			if parent == nil {
				return ErrCommentReplyForbidden
			}
			if parent.PostID != input.PostID {
				return ErrCommentParentPostMismatch
			}
			// 继承父节点的家族 (RootID)
			comment.RootID = parent.RootID
		}

		// 4. 落库
		id, err := uc.repo.Create(txCtx, comment)
		if err != nil {
			return err
		}
		comment.ID = id
		return nil
	})

	return comment, err
}

// DeleteComment 删除评论
func (uc *CommentUseCase) DeleteComment(ctx context.Context, id, operatorID int64, isAdmin bool) error {
	return uc.tx.ExecTx(ctx, func(txCtx context.Context) error {
		// 1. 查评论
		comment, err := uc.repo.GetByID(txCtx, id)
		if err != nil {
			return err
		}
		if comment == nil {
			return ErrCommentNotFound
		}

		// 2. 鉴权: 只有作者或管理员能删
		if comment.UserID != operatorID && !isAdmin {
			return ErrCommentConflict
		}

		// 3. 执行删除
		return uc.repo.Delete(txCtx, id)
	})
}
