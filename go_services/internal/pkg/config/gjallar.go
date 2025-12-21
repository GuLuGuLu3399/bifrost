package config

import (
	"time"

	"github.com/gulugulu3399/bifrost/internal/pkg/network/grpc"
	"github.com/gulugulu3399/bifrost/internal/pkg/network/http"
)

type Config struct {
	Http      http.ServerConfig `yaml:"http"`
	NexusRPC  grpc.ClientConfig `yaml:"nexus_rpc"`
	BeaconRPC grpc.ClientConfig `yaml:"beacon_rpc"`
}

// DefaultConfig 方便本地开发直接跑
func DefaultConfig() *Config {
	return &Config{
		Http: http.ServerConfig{
			Addr:         ":8080",
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			Timeout:      5 * time.Second, // 优雅关闭时间
		},
		NexusRPC: grpc.ClientConfig{
			Addr:    "localhost:9090",
			Timeout: 5 * time.Second,
		},
		BeaconRPC: grpc.ClientConfig{
			Addr:    "localhost:9091",
			Timeout: 5 * time.Second,
		},
	}
}
