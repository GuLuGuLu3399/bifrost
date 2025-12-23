package config

import (
	"time"

	"github.com/gulugulu3399/bifrost/internal/pkg/network/grpc"
)

// Config 统一后的网关配置，嵌入 BaseConfig。
type Config struct {
	BaseConfig `mapstructure:",squash" yaml:",inline"`

	HTTP struct {
		Addr         string        `mapstructure:"addr" yaml:"addr"`
		ReadTimeout  time.Duration `mapstructure:"read_timeout" yaml:"read_timeout"`
		WriteTimeout time.Duration `mapstructure:"write_timeout" yaml:"write_timeout"`
		IdleTimeout  time.Duration `mapstructure:"idle_timeout" yaml:"idle_timeout"`
		Timeout      time.Duration `mapstructure:"timeout" yaml:"timeout"`
	} `mapstructure:"http" yaml:"http"`

	RPC struct {
		Nexus  grpc.ClientConfig `mapstructure:"nexus" yaml:"nexus"`
		Beacon grpc.ClientConfig `mapstructure:"beacon" yaml:"beacon"`
	} `mapstructure:"rpc" yaml:"rpc"`
}

// DefaultConfig 提供本地快速启动配置。
func DefaultConfig() *Config {
	return &Config{
		BaseConfig: BaseConfig{
			App: struct {
				Name    string "mapstructure:\"name\" yaml:\"name\""
				Env     string "mapstructure:\"env\" yaml:\"env\""
				Version string "mapstructure:\"version\" yaml:\"version\""
			}{
				Name:    "bifrost-gjallar",
				Env:     "dev",
				Version: "1.0.0",
			},
			Logger: struct {
				Level  string "mapstructure:\"level\" yaml:\"level\""
				Format string "mapstructure:\"format\" yaml:\"format\""
			}{
				Level:  "info",
				Format: "json",
			},
			Observability: struct {
				OtlpEndpoint string "mapstructure:\"otlp_endpoint\" yaml:\"otlp_endpoint\""
			}{
				OtlpEndpoint: "localhost:4317",
			},
		},
		HTTP: struct {
			Addr         string        "mapstructure:\"addr\" yaml:\"addr\""
			ReadTimeout  time.Duration "mapstructure:\"read_timeout\" yaml:\"read_timeout\""
			WriteTimeout time.Duration "mapstructure:\"write_timeout\" yaml:\"write_timeout\""
			IdleTimeout  time.Duration "mapstructure:\"idle_timeout\" yaml:\"idle_timeout\""
			Timeout      time.Duration "mapstructure:\"timeout\" yaml:\"timeout\""
		}{
			Addr:         ":8080",
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  60 * time.Second,
			Timeout:      5 * time.Second,
		},
		RPC: struct {
			Nexus  grpc.ClientConfig "mapstructure:\"nexus\" yaml:\"nexus\""
			Beacon grpc.ClientConfig "mapstructure:\"beacon\" yaml:\"beacon\""
		}{
			Nexus: grpc.ClientConfig{Addr: "localhost:9001", Timeout: 5 * time.Second},
			Beacon: grpc.ClientConfig{Addr: "localhost:9002", Timeout: 5 * time.Second},
		},
	}
}

// LoadGjallarConfig 从环境变量或配置文件加载。
// 关键环境变量示例：
//   BIFROST_RPC_NEXUS_ADDR=nexus:9001
//   BIFROST_RPC_BEACON_ADDR=beacon:9002
//   BIFROST_OBSERVABILITY_OTLP_ENDPOINT=jaeger:4317
func LoadGjallarConfig() (*Config, error) {
	loader := NewLoader(WithDefaults(map[string]any{
		"app.name":                    "bifrost-gjallar",
		"app.env":                     "dev",
		"app.version":                 "1.0.0",
		"logger.level":                "info",
		"logger.format":               "json",
		"observability.otlp_endpoint": "localhost:4317",
		"http.addr":                   ":8080",
		"http.read_timeout":           "5s",
		"http.write_timeout":          "10s",
		"http.idle_timeout":           "60s",
		"http.timeout":                "5s",
		// Provide defaults so Viper binds keys and env can override
		"rpc.nexus.addr":              "localhost:9001",
		"rpc.nexus.timeout":           "5s",
		"rpc.beacon.addr":             "localhost:9002",
		"rpc.beacon.timeout":          "5s",
	}))

	cfg := DefaultConfig()

	v := loader.Viper()
	_ = v.Unmarshal(cfg)

	return cfg, nil
}
