package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/gulugulu3399/bifrost/internal/pkg/observability/logger"
)

// NexusConfig 对应 configs/nexus.yaml 配置文件（仅包含 nexus 服务当前需要的部分）。
// 我们保持与 YAML 结构一致，以便复用同一配置文件。
//
// 环境变量覆盖规则：
//   - 前缀: BIFROST_
//   - 键名替换: "." -> "_"
//
// 示例：
//
//	BIFROST_APP_GRPC_PORT=":9001"
//	BIFROST_DATA_DATABASE_SOURCE="..."
//
// 提示：viper 支持 duration 类型，如 "10s"、"1h"。
type NexusConfig struct {
	App struct {
		Name          string `mapstructure:"name"`
		Version       string `mapstructure:"version"`
		Env           string `mapstructure:"env"`
		GRPCPort      string `mapstructure:"grpc_port"`
		SnowflakeNode int64  `mapstructure:"snowflake_node"`

		Security struct {
			JWTSecret     string        `mapstructure:"jwt_secret"`
			JWTExpiration time.Duration `mapstructure:"jwt_expiration"`
		} `mapstructure:"security"`
	} `mapstructure:"app"`

	Logger struct {
		Level    string `mapstructure:"level"`    // 日志级别
		Encoding string `mapstructure:"encoding"` // 日志格式，对应 logger.Config 的 Format
		Dazzle   bool   `mapstructure:"dazzle"`   // 兼容字段，映射到 EnableColor
	} `mapstructure:"logger"`

	Data struct {
		Database struct {
			// 与 internal/pkg/database.Config 对齐
			Driver          string        `mapstructure:"driver"`
			DSN             string        `mapstructure:"dsn"`
			MaxOpenConns    int           `mapstructure:"max_open_conns"`
			MaxIdleConns    int           `mapstructure:"max_idle_conns"`
			ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
			ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time"`
		} `mapstructure:"database"`
	} `mapstructure:"data"`

	Messenger struct {
		Addr string `mapstructure:"addr"`
	} `mapstructure:"messenger"`
}

// validate 验证配置的必填项
func (c *NexusConfig) validate() error {
	if c.App.GRPCPort == "" {
		return fmt.Errorf("app.grpc_port 为必填项")
	}
	if c.Data.Database.Driver == "" {
		return fmt.Errorf("data.database.driver 为必填项")
	}
	if c.Data.Database.DSN == "" {
		return fmt.Errorf("data.database.dsn 为必填项")
	}
	if c.App.Security.JWTSecret == "" {
		return fmt.Errorf("app.security.jwt_secret 为必填项")
	}
	if c.App.SnowflakeNode == 0 {
		c.App.SnowflakeNode = 1
	}
	if c.Messenger.Addr == "" {
		c.Messenger.Addr = "nats://127.0.0.1:4222"
	}
	return nil
}

// LoadNexus 从 YAML 文件（如 configs/nexus.yaml）加载配置，支持环境变量覆盖
func LoadNexus(path string) (*NexusConfig, error) {
	l := NewLoader(WithDefaults(map[string]any{
		"app.env":            "dev",
		"app.snowflake_node": int64(1),
		"logger.level":       "info",
		"logger.encoding":    "json",
		"logger.dazzle":      false,
		// database 默认值（新字段）
		"data.database.driver":             "postgres",
		"data.database.max_open_conns":     25,
		"data.database.max_idle_conns":     25,
		"data.database.conn_max_lifetime":  "5m",
		"data.database.conn_max_idle_time": "5m",
		"messenger.addr":                   "nats://127.0.0.1:4222",
	}))

	var cfg NexusConfig
	if err := l.LoadFile(path, &cfg, cfg.validate); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// LoggerConfig 将 Nexus 配置中的日志部分转换为内部 logger.Config
func (c *NexusConfig) LoggerConfig() *logger.Config {
	lc := logger.DefaultConfig()

	// 设置日志格式
	switch strings.ToLower(strings.TrimSpace(c.Logger.Encoding)) {
	case "console":
		lc.Format = "console"
	default:
		lc.Format = "json"
	}

	// 是否启用颜色输出
	lc.EnableColor = c.Logger.Dazzle

	// 设置日志级别
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
