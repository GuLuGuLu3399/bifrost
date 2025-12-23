package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoader_LoadFile_EnvOverride(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "nexus.yaml")

	// minimal valid config
	yaml := []byte(`app:
  name: "bifrost-nexus"
  env: dev
  version: "1.0"
server:
  grpc_addr: ":9001"
security:
  jwt_secret: "secret"
data:
  database:
    driver: "postgres"
    dsn: "dsn"
logger:
  level: info
  encoding: json
`)
	if err := os.WriteFile(cfgPath, yaml, 0o644); err != nil {
		t.Fatalf("write cfg: %v", err)
	}

	// override grpc addr via env
	_ = os.Setenv("BIFROST_SERVER_GRPC_ADDR", ":9999")
	t.Cleanup(func() { _ = os.Unsetenv("BIFROST_SERVER_GRPC_ADDR") })

	cfg, err := LoadNexus(cfgPath)
	if err != nil {
		t.Fatalf("LoadNexus: %v", err)
	}
	if cfg.Server.GRPCAddr != ":9999" {
		t.Fatalf("expected grpc_addr overridden to :9999, got %q", cfg.Server.GRPCAddr)
	}
}

func TestLoadNexus_ValidateRequired(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "bad.yaml")

	// missing jwt_secret
	yaml := []byte(`app:
  name: "bifrost-nexus"
server:
  grpc_addr: ":9001"
data:
  database:
    driver: "postgres"
    dsn: "dsn"
`)
	if err := os.WriteFile(cfgPath, yaml, 0o644); err != nil {
		t.Fatalf("write cfg: %v", err)
	}

	_, err := LoadNexus(cfgPath)
	if err == nil {
		t.Fatalf("expected validation error, got nil")
	}
}
