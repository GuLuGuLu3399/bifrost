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
type EventPayload struct {
	ID   int64  `json:"id"`
	Slug string `json:"slug"`
}

type Consumer struct {
	data *Data
	msgr *messenger.Client

	mu   sync.Mutex
	subs []*nats.Subscription
}

func NewConsumer(data *Data, msgr *messenger.Client) *Consumer {
	return &Consumer{data: data, msgr: msgr}
}

// Start 启动监听（Fire-and-forget 不需要 ack/retry）
func (c *Consumer) Start() error {
	sub, err := c.msgr.Subscribe("content.>", "beacon_service", func(subject string, data []byte) {
		switch subject {
		case "content.post.created", "content.post.updated", "content.post.deleted":
			c.handlePostChange(data)
		case "content.category.updated", "content.category.created", "content.category.deleted":
			c.handleMetaChange()
		}
	})
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.subs = append(c.subs, sub)
	c.mu.Unlock()

	return nil
}

// Close 实现 lifecycle.Closer：取消订阅并尽量优雅退出。
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

	// Consumer 复用注入的 msgr，不负责关闭/Drain 连接（由 main 统一管理），避免重复 Close。
	return nil
}

func (c *Consumer) handlePostChange(data []byte) {
	var p EventPayload
	if err := json.Unmarshal(data, &p); err != nil {
		logger.Warn("beacon consumer unmarshal failed", logger.Err(err))
		return
	}

	ctx := context.Background()
	_ = c.data.Cache().Delete(ctx, KeyPostDetail(fmt.Sprintf("%d", p.ID)))
	if p.Slug != "" {
		_ = c.data.Cache().Delete(ctx, KeyPostDetail(p.Slug))
	}
}

func (c *Consumer) handleMetaChange() {
	_ = c.data.Cache().Delete(context.Background(), KeyCategoryList())
}
