package config

import (
	"fmt"
	"time"
)

// BeaconConfig 统一后的读服务配置，嵌入 BaseConfig 保持一致性。
type BeaconConfig struct {
	BaseConfig `mapstructure:",squash" yaml:",inline"`

	Server struct {
		GRPCAddr                string        `mapstructure:"grpc_addr" yaml:"grpc_addr"`
		GracefulShutdownTimeout time.Duration `mapstructure:"graceful_shutdown_timeout" yaml:"graceful_shutdown_timeout"`
	} `mapstructure:"server" yaml:"server"`

	Data struct {
		Database struct {
			DSN          string        `mapstructure:"dsn" yaml:"dsn"`
			MaxIdleConns int           `mapstructure:"max_idle_conns" yaml:"max_idle_conns"`
			MaxOpenConns int           `mapstructure:"max_open_conns" yaml:"max_open_conns"`
			MaxLifetime  time.Duration `mapstructure:"max_lifetime" yaml:"max_lifetime"`
		} `mapstructure:"database" yaml:"database"`

		Redis struct {
			Addr         string        `mapstructure:"addr" yaml:"addr"`
			Password     string        `mapstructure:"password" yaml:"password"`
			DB           int           `mapstructure:"db" yaml:"db"`
			DialTimeout  time.Duration `mapstructure:"dial_timeout" yaml:"dial_timeout"`
			ReadTimeout  time.Duration `mapstructure:"read_timeout" yaml:"read_timeout"`
			WriteTimeout time.Duration `mapstructure:"write_timeout" yaml:"write_timeout"`
			PoolSize     int           `mapstructure:"pool_size" yaml:"pool_size"`
			MinIdleConns int           `mapstructure:"min_idle_conns" yaml:"min_idle_conns"`
		} `mapstructure:"redis" yaml:"redis"`
	} `mapstructure:"data" yaml:"data"`

	Messenger struct {
		Addr string `mapstructure:"addr" yaml:"addr"`
	} `mapstructure:"messenger" yaml:"messenger"`
}

func (c *BeaconConfig) validate() error {
	if c.App.Name == "" {
		c.App.Name = "bifrost-beacon"
	}
	if c.Server.GRPCAddr == "" {
		return fmt.Errorf("server.grpc_addr 为必填项")
	}
	if c.Data.Database.DSN == "" {
		return fmt.Errorf("data.database.dsn 为必填项")
	}
	if c.Server.GracefulShutdownTimeout == 0 {
		c.Server.GracefulShutdownTimeout = 10 * time.Second
	}
	if c.Messenger.Addr == "" {
		c.Messenger.Addr = "nats://127.0.0.1:4222"
	}
	if c.Observability.OtlpEndpoint == "" {
		c.Observability.OtlpEndpoint = "localhost:4317"
	}
	return nil
}

func LoadBeacon(path string) (*BeaconConfig, error) {
	l := NewLoader(WithDefaults(map[string]any{
		"app.name":                    "bifrost-beacon",
		"app.env":                     "dev",
		"app.version":                 "1.0.0",
		"logger.level":                "info",
		"logger.format":               "json",
		"observability.otlp_endpoint": "localhost:4317",
		"server.grpc_addr":            ":9002",
		"server.graceful_shutdown_timeout": "10s",
		"data.database.max_idle_conns":     20,
		"data.database.max_open_conns":     200,
		"data.database.max_lifetime":       "1h",
		"data.redis.dial_timeout":          "5s",
		"data.redis.read_timeout":          "3s",
		"data.redis.write_timeout":         "3s",
		"data.redis.pool_size":             20,
		"data.redis.min_idle_conns":        5,
		"messenger.addr":                   "nats://127.0.0.1:4222",
	}))

	var cfg BeaconConfig
	if err := l.LoadFile(path, &cfg, cfg.validate); err != nil {
		return nil, err
	}
	return &cfg, nil
}
