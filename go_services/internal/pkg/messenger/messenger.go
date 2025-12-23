package messenger

import (
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/gulugulu3399/bifrost/internal/pkg/xerr"
)

// Client 是轻量级 NATS 消息客户端
// 核心原则：Fire-and-Forget，不重试，不持久化
type Client struct {
	conn *nats.Conn
}

// New 连接 NATS 服务
// addr: NATS 服务地址，如 "nats://localhost:4222"
// serviceName: 服务名（用于 NATS Dashboard 识别），如 "bifrost_nexus"
func New(addr string, serviceName string) (*Client, error) {
	nc, err := nats.Connect(
		addr,
		nats.Name(serviceName), // 便于在 NATS Dashboard 里排查问题
	)
	if err != nil {
		return nil, xerr.Wrap(err, xerr.CodeInternal, "nats connect failed")
	}
	return &Client{conn: nc}, nil
}

// Close 关闭 NATS 连接
func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

// Publish 发布消息到指定主题 (Fire-and-Forget)
// 核心特点：
// 1. 异步发送，不等待回复
// 2. 不保证消息送达（网络问题可能丢失）
// 3. 适合对一致性要求不高的场景（如缓存失效通知）
//
// 使用示例：
// messenger.Publish("content.post.updated", map[string]interface{}{
//     "id": 12345,
//     "slug": "my-article",
// })
func (c *Client) Publish(subject string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return xerr.Wrap(err, xerr.CodeInternal, "json marshal failed")
	}
	// 发送到 NATS，不等待响应
	return c.conn.Publish(subject, data)
}

// Handler 是消息处理函数签名
type Handler func(subject string, data []byte)

// Subscribe 订阅主题，使用 Queue Group 实现负载均衡
//
// 参数说明：
// - subject: 订阅主题，支持通配符 "content.>" 监听所有 content 事件
// - group: Queue Group 名称（同一组的多个消费者会负载均衡）
//   - "beacon_service" 用于缓存失效
//   - "mirror_service" 用于索引更新
// - handler: 消息处理回调函数
//
// 使用示例：
// sub, err := messenger.Subscribe("content.>", "beacon_service", func(subject string, data []byte) {
//     // 根据 subject 路由到不同的处理器
//     switch subject {
//     case "content.post.updated":
//         // 删除缓存
//     case "content.category.updated":
//         // 清空分类缓存
//     }
// })
//
// 返回值：subscription，可用于 Unsubscribe 或查询订阅状态
func (c *Client) Subscribe(subject string, group string, handler Handler) (*nats.Subscription, error) {
	// QueueSubscribe 用法：
	// - 同一个 group 的多个订阅者，NATS 会自动负载均衡，每条消息只投递给一个订阅者
	// - 这样即使启动多个 Beacon 副本，也不会重复处理同一消息
	sub, err := c.conn.QueueSubscribe(subject, group, func(msg *nats.Msg) {
		handler(msg.Subject, msg.Data)
	})
	if err != nil {
		return nil, xerr.Wrap(err, xerr.CodeInternal, fmt.Sprintf("queue subscribe to %s failed", subject))
	}
	return sub, nil
}

// Unsubscribe 取消订阅
func (c *Client) Unsubscribe(sub *nats.Subscription) error {
	if sub == nil {
		return nil
	}
	return sub.Unsubscribe()
}

// Flush 等待所有发布的消息被服务器确认
// 通常在关闭连接前调用，确保消息不丢失
func (c *Client) Flush() error {
	return c.conn.Flush()
}
