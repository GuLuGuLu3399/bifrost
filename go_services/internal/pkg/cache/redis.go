package cache

import (
	"context"
	"encoding/json"
	"errors"

	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"

	"github.com/gulugulu3399/bifrost/internal/pkg/xerr"
)

var (
	ErrCacheMiss = xerr.New(xerr.CodeNotFound, "cache miss")
	// 全局 singleflight 组即可，按 key 隔离逻辑，无需 map
	sfGroup singleflight.Group
)

// Config Redis 配置
type Config struct {
	Addr         string        `yaml:"addr"`
	Password     string        `yaml:"password"`
	DB           int           `yaml:"db"`
	PoolSize     int           `yaml:"pool_size"`
	MinIdleConns int           `yaml:"min_idle_conns"`
	DialTimeout  time.Duration `yaml:"dial_timeout"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Addr:         "localhost:6379",
		DB:           0,
		PoolSize:     10,
		MinIdleConns: 5,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}
}

// Client Redis 客户端封装
type Client struct {
	rdb *redis.Client
}

// Close closes the underlying redis client (connection pool).
// Call this on application shutdown.
func (c *Client) Close() error {
	if c == nil || c.rdb == nil {
		return nil
	}
	return c.rdb.Close()
}

// New 创建新的 Redis 客户端
func New(cfg *Config) (*Client, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})

	// 验证连接
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, xerr.Wrap(err, xerr.CodeInternal, "failed to connect to redis")
	}

	return &Client{rdb: rdb}, nil
}

// NewClient 从已有的 redis.Client 创建客户端
func NewClient(rdb *redis.Client) *Client {
	return &Client{rdb: rdb}
}

// Fetch 核心方法：支持泛型、自动序列化、Singleflight
// T 代表返回的数据类型
func Fetch[T any](ctx context.Context, c *Client, key string, ttl time.Duration, fetcher func() (*T, error)) (*T, error) {
	// 1. 尝试从缓存读取
	val, err := c.Get(ctx, key)
	if err == nil {
		var data T
		if err := json.Unmarshal([]byte(val), &data); err == nil {
			return &data, nil
		}
		// JSON 解析失败，记录错误但继续执行回源逻辑
	}

	// 2. 缓存未命中，使用 singleflight 保护后端 (如 DB)
	// 使用 sfGroup.Do，key 相同的请求会被合并成一个
	res, err, _ := sfGroup.Do(key, func() (any, error) {
		// 再次检查缓存 (Double Check)，防止并发穿透
		if val, err := c.Get(ctx, key); err == nil {
			var data T
			if err := json.Unmarshal([]byte(val), &data); err == nil {
				return &data, nil
			}
		}

		// 执行真正的回源逻辑 (例如查数据库)
		data, err := fetcher()
		if err != nil {
			return nil, err
		}

		// 异步回写缓存，不阻塞主流程
		go func() {
			// 序列化并写入
			if buf, err := json.Marshal(data); err == nil {
				_ = c.Set(context.Background(), key, string(buf), ttl)
			}
		}()

		return data, nil
	})

	if err != nil {
		return nil, err
	}

	return res.(*T), nil
}

// Get 基础操作
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	val, err := c.rdb.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrCacheMiss
	}
	if err != nil {
		return "", xerr.Wrap(err, xerr.CodeInternal, fmt.Sprintf("failed to get key %s", key))
	}
	return val, nil
}

// Set 基础操作
func (c *Client) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	if err := c.rdb.Set(ctx, key, value, expiration).Err(); err != nil {
		return xerr.Wrap(err, xerr.CodeInternal, fmt.Sprintf("failed to set key %s", key))
	}
	return nil
}

// Delete 基础操作
func (c *Client) Delete(ctx context.Context, key string) error {
	if err := c.rdb.Del(ctx, key).Err(); err != nil {
		return xerr.Wrap(err, xerr.CodeInternal, fmt.Sprintf("failed to delete key %s", key))
	}
	return nil
}

// Exists 检查 key 是否存在
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	count, err := c.rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, xerr.Wrap(err, xerr.CodeInternal, fmt.Sprintf("failed to check key %s", key))
	}
	return count > 0, nil
}

// Expire 设置 key 过期时间
func (c *Client) Expire(ctx context.Context, key string, expiration time.Duration) error {
	if err := c.rdb.Expire(ctx, key, expiration).Err(); err != nil {
		return xerr.Wrap(err, xerr.CodeInternal, fmt.Sprintf("failed to expire key %s", key))
	}
	return nil
}

// GetTTL 获取 key 剩余过期时间
func (c *Client) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := c.rdb.TTL(ctx, key).Result()
	if err != nil {
		return 0, xerr.Wrap(err, xerr.CodeInternal, fmt.Sprintf("failed to get ttl for key %s", key))
	}
	return ttl, nil
}
