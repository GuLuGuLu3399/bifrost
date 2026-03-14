package messenger

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gulugulu3399/bifrost/internal/pkg/xerr"
	"github.com/nats-io/nats.go"
)

// Client 是轻量级 NATS 消息客户端
// 核心原则：Fire-and-Forget，不重试，不持久化
type Client struct {
	conn *nats.Conn
}

// New 连接 NATS 服务
// addr: NATS 服务地址，如 "nats://localhost:4222"
// serviceName: (可选) 服务名，用于 NATS Dashboard 识别
func New(addr string, serviceName ...string) (*Client, error) {
	var opts []nats.Option
	if len(serviceName) > 0 && serviceName[0] != "" {
		opts = append(opts, nats.Name(serviceName[0]))
	} else {
		opts = append(opts, nats.Name("bifrost_service"))
	}

	nc, err := nats.Connect(addr, opts...)
	if err != nil {
		return nil, xerr.Wrap(err, xerr.CodeInternal, "nats connect failed")
	}
	return &Client{conn: nc}, nil
}

// Publish 发布消息到指定主题 (Fire-and-Forget)
// 核心特点：
// 1. 异步发送，不等待回复
// 2. 不保证消息送达（网络问题可能丢失）
// 3. 适合对一致性要求不高的场景（如缓存失效通知）
func (c *Client) Publish(subject string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return xerr.Wrap(err, xerr.CodeInternal, "json marshal failed")
	}
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
//
// - handler: 消息处理回调函数
//
// 返回值：subscription，可用于 Unsubscribe 或查询订阅状态
func (c *Client) Subscribe(subject string, group string, handler Handler) (*nats.Subscription, error) {
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
	if c.conn == nil {
		return nil
	}
	return c.conn.Flush()
}

// Drain 优雅关闭连接：停止接收新消息并等待现有消息处理完成
func (c *Client) Drain() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Drain()
}

// Close 关闭 NATS 连接
func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	// 尽量优雅退出：停止接收新消息并 flush pending publish
	_ = c.conn.Drain()
	c.conn.Close()
	return nil
}

// EnsureStream 校验 JetStream Stream 存在；若不存在则创建。
func (c *Client) EnsureStream(name string, subjects []string) error {
	if c.conn == nil {
		return xerr.New(xerr.CodeInternal, "nats connection is nil")
	}
	if name == "" {
		return xerr.New(xerr.CodeBadRequest, "stream name is required")
	}
	if len(subjects) == 0 {
		return xerr.New(xerr.CodeBadRequest, "stream subjects are required")
	}

	js, err := c.conn.JetStream()
	if err != nil {
		return xerr.Wrap(err, xerr.CodeInternal, "jetstream init failed")
	}

	if _, err := js.StreamInfo(name); err == nil {
		return nil
	}

	if _, err := js.AddStream(&nats.StreamConfig{
		Name:     name,
		Subjects: subjects,
	}); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "already") {
			return nil
		}
		return xerr.Wrap(err, xerr.CodeInternal, fmt.Sprintf("create stream %s failed", name))
	}

	return nil
}

// EnsurePullConsumer 校验 Durable Pull Consumer 存在；若不存在则创建。
func (c *Client) EnsurePullConsumer(streamName, consumerName, filterSubject string) error {
	if c.conn == nil {
		return xerr.New(xerr.CodeInternal, "nats connection is nil")
	}
	if streamName == "" || consumerName == "" || filterSubject == "" {
		return xerr.New(xerr.CodeBadRequest, "stream/consumer/filter are required")
	}

	js, err := c.conn.JetStream()
	if err != nil {
		return xerr.Wrap(err, xerr.CodeInternal, "jetstream init failed")
	}

	if _, err := js.ConsumerInfo(streamName, consumerName); err == nil {
		return nil
	}

	_, err = js.AddConsumer(streamName, &nats.ConsumerConfig{
		Durable:       consumerName,
		FilterSubject: filterSubject,
		AckPolicy:     nats.AckExplicitPolicy,
	})
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "already") {
			return nil
		}
		return xerr.Wrap(err, xerr.CodeInternal, fmt.Sprintf("create consumer %s failed", consumerName))
	}

	return nil
}

// EnsureDefaultContentTopology 确保内容事件链路所需的默认 Stream 与 Consumer 已就绪。
func (c *Client) EnsureDefaultContentTopology() error {
	if err := c.EnsureStream(StreamContent, []string{SubjectContentAll}); err != nil {
		return err
	}
	if err := c.EnsurePullConsumer(StreamContent, ConsumerMirrorIndexer, SubjectPostWildcard); err != nil {
		return err
	}
	return nil
}
