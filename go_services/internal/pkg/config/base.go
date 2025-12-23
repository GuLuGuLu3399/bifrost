package config

import (
	"strings"

	"github.com/gulugulu3399/bifrost/internal/pkg/observability/logger"
)

// BaseConfig 定义所有服务共享的基础配置，确保命名一致。
type BaseConfig struct {
	App struct {
		Name    string `mapstructure:"name" yaml:"name"`
		Env     string `mapstructure:"env" yaml:"env"`
		Version string `mapstructure:"version" yaml:"version"`
	} `mapstructure:"app" yaml:"app"`

	Logger struct {
		Level  string `mapstructure:"level" yaml:"level"`
		Format string `mapstructure:"format" yaml:"format"`
	} `mapstructure:"logger" yaml:"logger"`

	Observability struct {
		OtlpEndpoint string `mapstructure:"otlp_endpoint" yaml:"otlp_endpoint"`
	} `mapstructure:"observability" yaml:"observability"`

	// 对象存储配置
	Storage struct {
		Endpoint        string `mapstructure:"endpoint" yaml:"endpoint"`
		AccessKeyID     string `mapstructure:"access_key_id" yaml:"access_key_id"`
		SecretAccessKey string `mapstructure:"secret_access_key" yaml:"secret_access_key"`
		Bucket          string `mapstructure:"bucket" yaml:"bucket"`
		UseSSL          bool   `mapstructure:"use_ssl" yaml:"use_ssl"`
		Region          string `mapstructure:"region" yaml:"region"`
	} `mapstructure:"storage" yaml:"storage"`
}

// LoggerConfig 将基础配置转为统一的 logger.Config。
func (b *BaseConfig) LoggerConfig() *logger.Config {
	lc := logger.DefaultConfig()

	switch strings.ToLower(strings.TrimSpace(b.Logger.Format)) {
	case "console":
		lc.Format = "console"
	default:
		lc.Format = "json"
	}

	switch strings.ToLower(strings.TrimSpace(b.Logger.Level)) {
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
