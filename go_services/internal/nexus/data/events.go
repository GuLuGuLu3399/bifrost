package data

import (
	"context"

	"github.com/gulugulu3399/bifrost/internal/pkg/messenger"
	"github.com/gulugulu3399/bifrost/internal/pkg/observability/logger"
)

// ============ 事件发送示例 ============
// 这些函数应该在 Nexus 服务的 Data Layer 中使用

// PublishPostCreated 发布"文章已创建"事件
// 由 Nexus 服务在成功保存文章后调用
func (d *Data) PublishPostCreated(ctx context.Context, msgr *messenger.Client, id int64, slug string) {
	// Fire-and-Forget 原则：使用 go func() 异步发送，不阻塞主业务逻辑
	go func() {
		payload := messenger.PostEventPayload{
			ID:   id,
			Slug: slug,
		}
		if err := msgr.Publish(messenger.SubjectPostCreated, payload); err != nil {
			// 只记日志，不返回错误给客户端
			logger.Warn("failed to publish post created event",
				logger.Err(err),
				logger.Int64("post_id", id),
			)
		}
	}()
}

// PublishPostUpdated 发布"文章已更新"事件
func (d *Data) PublishPostUpdated(ctx context.Context, msgr *messenger.Client, id int64, slug string) {
	go func() {
		payload := messenger.PostEventPayload{
			ID:   id,
			Slug: slug,
		}
		if err := msgr.Publish(messenger.SubjectPostUpdated, payload); err != nil {
			logger.Warn("failed to publish post updated event",
				logger.Err(err),
				logger.Int64("post_id", id),
			)
		}
	}()
}

// PublishPostDeleted 发布"文章已删除"事件
func (d *Data) PublishPostDeleted(ctx context.Context, msgr *messenger.Client, id int64, slug string) {
	go func() {
		payload := messenger.PostEventPayload{
			ID:   id,
			Slug: slug,
		}
		if err := msgr.Publish(messenger.SubjectPostDeleted, payload); err != nil {
			logger.Warn("failed to publish post deleted event",
				logger.Err(err),
				logger.Int64("post_id", id),
			)
		}
	}()
}

// PublishCategoryUpdated 发布"分类已更新"事件
func (d *Data) PublishCategoryUpdated(ctx context.Context, msgr *messenger.Client, id int64, slug string) {
	go func() {
		payload := messenger.CategoryEventPayload{
			ID:   id,
			Slug: slug,
		}
		if err := msgr.Publish(messenger.SubjectCategoryUpdate, payload); err != nil {
			logger.Warn("failed to publish category updated event",
				logger.Err(err),
				logger.Int64("category_id", id),
			)
		}
	}()
}

// PublishCommentCreated 发布"评论已创建"事件
func (d *Data) PublishCommentCreated(ctx context.Context, msgr *messenger.Client, id, postID, userID int64) {
	go func() {
		payload := messenger.CommentEventPayload{
			ID:     id,
			PostID: postID,
			UserID: userID,
		}
		if err := msgr.Publish(messenger.SubjectCommentCreated, payload); err != nil {
			logger.Warn("failed to publish comment created event",
				logger.Err(err),
				logger.Int64("comment_id", id),
			)
		}
	}()
}

// PublishInteraction 发布"用户互动"事件（点赞、收藏、分享）
func (d *Data) PublishInteraction(ctx context.Context, msgr *messenger.Client, userID, postID int64, iType string) {
	go func() {
		payload := messenger.InteractionEventPayload{
			UserID: userID,
			PostID: postID,
			Type:   iType,
		}
		if err := msgr.Publish(messenger.SubjectInteraction, payload); err != nil {
			logger.Warn("failed to publish interaction event",
				logger.Err(err),
				logger.Int64("user_id", userID),
				logger.String("type", iType),
			)
		}
	}()
}
