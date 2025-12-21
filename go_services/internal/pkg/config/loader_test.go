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
  grpc_port: ":9001"
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

	// override grpc port via env
	_ = os.Setenv("BIFROST_APP_GRPC_PORT", ":9999")
	t.Cleanup(func() { _ = os.Unsetenv("BIFROST_APP_GRPC_PORT") })

	cfg, err := LoadNexus(cfgPath)
	if err != nil {
		t.Fatalf("LoadNexus: %v", err)
	}
	if cfg.App.GRPCPort != ":9999" {
		t.Fatalf("expected grpc_port overridden to :9999, got %q", cfg.App.GRPCPort)
	}
}

func TestLoadNexus_ValidateRequired(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "bad.yaml")

	// missing jwt_secret
	yaml := []byte(`app:
  grpc_port: ":9001"
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
