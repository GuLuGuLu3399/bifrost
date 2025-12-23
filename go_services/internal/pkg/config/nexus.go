package config

import (
	"fmt"
	"time"
)

// NexusConfig 统一后的写服务配置，嵌入 BaseConfig 保证命名一致。
type NexusConfig struct {
	BaseConfig `mapstructure:",squash" yaml:",inline"`

	Server struct {
		GRPCAddr string `mapstructure:"grpc_addr" yaml:"grpc_addr"`
	} `mapstructure:"server" yaml:"server"`

	Security struct {
		JWTSecret     string        `mapstructure:"jwt_secret" yaml:"jwt_secret"`
		JWTExpiration time.Duration `mapstructure:"jwt_expiration" yaml:"jwt_expiration"`
	} `mapstructure:"security" yaml:"security"`

	SnowflakeNode int64 `mapstructure:"snowflake_node" yaml:"snowflake_node"`

	Data struct {
		Database struct {
			Driver          string        `mapstructure:"driver" yaml:"driver"`
			DSN             string        `mapstructure:"dsn" yaml:"dsn"`
			MaxOpenConns    int           `mapstructure:"max_open_conns" yaml:"max_open_conns"`
			MaxIdleConns    int           `mapstructure:"max_idle_conns" yaml:"max_idle_conns"`
			ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime" yaml:"conn_max_lifetime"`
			ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time" yaml:"conn_max_idle_time"`
		} `mapstructure:"database" yaml:"database"`
	} `mapstructure:"data" yaml:"data"`

	Messenger struct {
		Addr string `mapstructure:"addr" yaml:"addr"`
	} `mapstructure:"messenger" yaml:"messenger"`

	// 下游 RPC 地址配置
	RPC struct {
		ForgeAddr string `mapstructure:"forge_addr" yaml:"forge_addr"`
	} `mapstructure:"rpc" yaml:"rpc"`
}

func (c *NexusConfig) validate() error {
	if c.App.Name == "" {
		c.App.Name = "bifrost-nexus"
	}
	if c.Server.GRPCAddr == "" {
		return fmt.Errorf("server.grpc_addr 为必填项")
	}
	if c.Data.Database.Driver == "" {
		return fmt.Errorf("data.database.driver 为必填项")
	}
	if c.Data.Database.DSN == "" {
		return fmt.Errorf("data.database.dsn 为必填项")
	}
	if c.Security.JWTSecret == "" {
		return fmt.Errorf("security.jwt_secret 为必填项")
	}
	if c.SnowflakeNode == 0 {
		c.SnowflakeNode = 1
	}
	if c.Messenger.Addr == "" {
		c.Messenger.Addr = "nats://127.0.0.1:4222"
	}
	if c.Observability.OtlpEndpoint == "" {
		c.Observability.OtlpEndpoint = "localhost:4317"
	}
	// ForgeAddr 可选：若未配置则跳过渲染调用
	return nil
}

// LoadNexus 从 YAML（或环境变量）加载 Nexus 配置。
func LoadNexus(path string) (*NexusConfig, error) {
	l := NewLoader(WithDefaults(map[string]any{
		"app.name":                    "bifrost-nexus",
		"app.env":                     "dev",
		"app.version":                 "1.0.0",
		"logger.level":                "info",
		"logger.format":               "json",
		"observability.otlp_endpoint": "localhost:4317",
		"server.grpc_addr":            ":9001",
		"snowflake_node":              int64(1),
		"security.jwt_expiration":     "24h",
		"data.database.driver":        "pgx",
		"data.database.max_open_conns":     25,
		"data.database.max_idle_conns":     25,
		"data.database.conn_max_lifetime":  "5m",
		"data.database.conn_max_idle_time": "5m",
		"messenger.addr":                   "nats://127.0.0.1:4222",
		"rpc.forge_addr":                    "",
	}))

	var cfg NexusConfig
	if err := l.LoadFile(path, &cfg, cfg.validate); err != nil {
		return nil, err
	}
	return &cfg, nil
}
