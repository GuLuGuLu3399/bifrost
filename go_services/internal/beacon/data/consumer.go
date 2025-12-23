package data

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gulugulu3399/bifrost/internal/pkg/messenger"
	"github.com/gulugulu3399/bifrost/internal/pkg/observability/logger"
	"github.com/nats-io/nats.go"
)

// EventPayload 定义从 NATS 接收的消息结构 (应与 Nexus 发出的一致)
// 这是最小的事件载荷，用于识别哪个资源变更了
type EventPayload struct {
	ID   int64  `json:"id"`
	Slug string `json:"slug"`
}

// Consumer 负责监听 NATS 事件并更新缓存
// 核心原则：Fire-and-Forget，宁可多删缓存，不可少删
type Consumer struct {
	data *Data
	msgr *messenger.Client

	mu   sync.Mutex
	subs []*nats.Subscription
}

// NewConsumer 创建消费者实例
func NewConsumer(data *Data, msgr *messenger.Client) *Consumer {
	return &Consumer{data: data, msgr: msgr}
}

// Start 启动监听
// 关键点：
// 1. 监听 "content.>" 通配符，捕获所有内容变动
// 2. Queue Group "beacon_service" 实现负载均衡
// 3. 处理失败只记日志，不阻塞或返回错误
func (c *Consumer) Start() error {
	sub, err := c.msgr.Subscribe("content.>", messenger.GroupBeacon, func(subject string, data []byte) {
		// 路由事件到对应的处理器
		switch subject {
		case messenger.SubjectPostCreated, messenger.SubjectPostUpdated, messenger.SubjectPostDeleted:
			c.handlePostChange(data)
		case messenger.SubjectCategoryUpdate:
			c.handleCategoryChange()
		case messenger.SubjectTagUpdate:
			c.handleTagChange()
		case messenger.SubjectCommentCreated:
			c.handleCommentChange(data)
		case messenger.SubjectInteraction:
			c.handleInteractionChange(data)
		}
	})
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.subs = append(c.subs, sub)
	c.mu.Unlock()

	logger.Info("beacon consumer started", logger.String("group", messenger.GroupBeacon))
	return nil
}

// Close 优雅关闭消费者
// 实现 lifecycle.Closer 接口
func (c *Consumer) Close() error {
	c.mu.Lock()
	subs := append([]*nats.Subscription(nil), c.subs...)
	c.subs = nil
	c.mu.Unlock()

	for _, s := range subs {
		if s == nil {
			continue
		}
		_ = s.Unsubscribe()
	}

	// Consumer 复用注入的 msgr，不负责关闭连接（由 main 统一管理）
	return nil
}

// ============ 事件处理函数 ============

// handlePostChange 处理文章变更事件
// 策略：删除 ID 缓存、Slug 缓存、以及列表缓存
// 理由：宁可多删（用户多查一次数据库），不可少删（用户看到旧数据）
func (c *Consumer) handlePostChange(data []byte) {
	var p EventPayload
	if err := json.Unmarshal(data, &p); err != nil {
		logger.Warn("beacon consumer: unmarshal post event failed", logger.Err(err))
		return
	}

	ctx := context.Background()

	// 删除 ID 缓存（如 post:12345）
	if p.ID > 0 {
		_ = c.data.Cache().Delete(ctx, KeyPostDetail(fmt.Sprintf("%d", p.ID)))
	}

	// 删除 Slug 缓存（如 post:my-article）
	if p.Slug != "" {
		_ = c.data.Cache().Delete(ctx, KeyPostDetail(p.Slug))
	}

	// 删除文章列表缓存（因为列表排序、计数可能改变）
	_ = c.data.Cache().Delete(ctx, KeyPostList())
}

// handleCategoryChange 处理分类变更事件
func (c *Consumer) handleCategoryChange() {
	ctx := context.Background()
	// 分类变了，把分类列表缓存全清了
	_ = c.data.Cache().Delete(ctx, KeyCategoryList())
}

// handleTagChange 处理标签变更事件
func (c *Consumer) handleTagChange() {
	ctx := context.Background()
	_ = c.data.Cache().Delete(ctx, KeyTagList())
}

// handleCommentChange 处理评论变更事件
// 删除对应文章的缓存（因为评论计数改变）
func (c *Consumer) handleCommentChange(data []byte) {
	var p messenger.CommentEventPayload
	if err := json.Unmarshal(data, &p); err != nil {
		logger.Warn("beacon consumer: unmarshal comment event failed", logger.Err(err))
		return
	}

	ctx := context.Background()
	if p.PostID > 0 {
		_ = c.data.Cache().Delete(ctx, KeyPostDetail(fmt.Sprintf("%d", p.PostID)))
	}
}

// handleInteractionChange 处理用户互动变更事件（点赞、收藏等）
// 删除文章缓存（因为点赞/收藏计数改变）
func (c *Consumer) handleInteractionChange(data []byte) {
	var p messenger.InteractionEventPayload
	if err := json.Unmarshal(data, &p); err != nil {
		logger.Warn("beacon consumer: unmarshal interaction event failed", logger.Err(err))
		return
	}

	ctx := context.Background()
	if p.PostID > 0 {
		_ = c.data.Cache().Delete(ctx, KeyPostDetail(fmt.Sprintf("%d", p.PostID)))
	}
}
