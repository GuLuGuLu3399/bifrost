package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/gulugulu3399/bifrost/internal/pkg/observability/logger"
)

// BeaconConfig 对应 configs/beacon.yaml（只读服务 Beacon）。
// 结构保持与 YAML 一致，便于 viper mapstructure 直接反序列化。
//
// 注意：common.Loader 会自动支持环境变量覆盖：
// - 前缀 BIFROST
// - "." -> "_"
//
// 例如：BIFROST_APP_GRPC_PORT=":9002"
//
//	BIFROST_DATA_DATABASE_SOURCE="..."
type BeaconConfig struct {
	App struct {
		Name     string `mapstructure:"name"`
		Version  string `mapstructure:"version"`
		Env      string `mapstructure:"env"`
		GRPCPort string `mapstructure:"grpc_port"`

		GracefulShutdownTimeout time.Duration `mapstructure:"graceful_shutdown_timeout"`
	} `mapstructure:"app"`

	Logger struct {
		Level    string `mapstructure:"level"`
		Encoding string `mapstructure:"encoding"`
		Dazzle   bool   `mapstructure:"dazzle"`
	} `mapstructure:"logger"`

	Data struct {
		Database struct {
			Source       string        `mapstructure:"source"`
			MaxIdleConns int           `mapstructure:"max_idle_conns"`
			MaxOpenConns int           `mapstructure:"max_open_conns"`
			MaxLifetime  time.Duration `mapstructure:"max_lifetime"`
		} `mapstructure:"database"`

		Redis struct {
			Addr         string        `mapstructure:"addr"`
			Password     string        `mapstructure:"password"`
			DB           int           `mapstructure:"db"`
			DialTimeout  time.Duration `mapstructure:"dial_timeout"`
			ReadTimeout  time.Duration `mapstructure:"read_timeout"`
			WriteTimeout time.Duration `mapstructure:"write_timeout"`
			PoolSize     int           `mapstructure:"pool_size"`
			MinIdleConns int           `mapstructure:"min_idle_conns"`
		} `mapstructure:"redis"`
	} `mapstructure:"data"`

	Messenger struct {
		Addr string `mapstructure:"addr"`
	} `mapstructure:"messenger"`
}

func (c *BeaconConfig) validate() error {
	if c.App.GRPCPort == "" {
		return fmt.Errorf("app.grpc_port 为必填项")
	}
	if c.Data.Database.Source == "" {
		return fmt.Errorf("data.database.source 为必填项")
	}
	if c.App.GracefulShutdownTimeout == 0 {
		c.App.GracefulShutdownTimeout = 10 * time.Second
	}
	if c.Messenger.Addr == "" {
		c.Messenger.Addr = "nats://127.0.0.1:4222"
	}
	return nil
}

func LoadBeacon(path string) (*BeaconConfig, error) {
	l := NewLoader(WithDefaults(map[string]any{
		"app.env":                       "dev",
		"app.graceful_shutdown_timeout": "10s",
		"logger.level":                  "info",
		"logger.encoding":               "json",
		"logger.dazzle":                 false,
		"data.database.max_idle_conns":  20,
		"data.database.max_open_conns":  200,
		"data.database.max_lifetime":    "1h",
		"data.redis.dial_timeout":       "5s",
		"data.redis.read_timeout":       "3s",
		"data.redis.write_timeout":      "3s",
		"data.redis.pool_size":          20,
		"data.redis.min_idle_conns":     5,
		"messenger.addr":                "nats://127.0.0.1:4222",
	}))

	var cfg BeaconConfig
	if err := l.LoadFile(path, &cfg, cfg.validate); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *BeaconConfig) LoggerConfig() *logger.Config {
	lc := logger.DefaultConfig()

	switch strings.ToLower(strings.TrimSpace(c.Logger.Encoding)) {
	case "console":
		lc.Format = "console"
	default:
		lc.Format = "json"
	}

	lc.EnableColor = c.Logger.Dazzle

	switch strings.ToLower(strings.TrimSpace(c.Logger.Level)) {
	case "debug":
		lc.Level = logger.DebugLevel
	case "warn", "warning":
		lc.Level = logger.WarnLevel
	case "error":
		lc.Level = logger.ErrorLevel
	default:
		lc.Level = logger.InfoLevel
	}

	return lc
}
