package messenger

import (
	"encoding/json"

	"github.com/gulugulu3399/bifrost/internal/pkg/xerr"
	"github.com/nats-io/nats.go"
)

type Client struct {
	conn *nats.Conn
}

// New 连接 NATS，保持最简
func New(addr string) (*Client, error) {
	nc, err := nats.Connect(addr, nats.Name("bifrost_service"))
	if err != nil {
		return nil, xerr.Wrap(err, xerr.CodeInternal, "nats connect failed")
	}
	return &Client{conn: nc}, nil
}

// Publish 发送消息 (Fire-and-Forget)
func (c *Client) Publish(subject string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return xerr.Wrap(err, xerr.CodeInternal, "marshal failed")
	}
	return c.conn.Publish(subject, data)
}

// Subscribe 队列组订阅 (核心方法)
// group: 消费者组名 (如 "beacon_service")
// handler: 业务回调
func (c *Client) Subscribe(subject string, group string, handler func(subject string, data []byte)) (*nats.Subscription, error) {
	return c.conn.QueueSubscribe(subject, group, func(msg *nats.Msg) {
		handler(msg.Subject, msg.Data)
	})
}

// Flush 用于确保缓冲区消息尽快发出（可选）。
func (c *Client) Flush() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Flush()
}

func (c *Client) Drain() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Drain()
}

func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	// 尽量优雅退出：停止接收新消息并 flush pending publish
	_ = c.conn.Drain()
	c.conn.Close()
	return nil
}
