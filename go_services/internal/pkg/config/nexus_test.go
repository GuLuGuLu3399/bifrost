package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadNexus_Validate(t *testing.T) {
	// Create a temporary YAML file with minimal fields
	dir := t.TempDir()
	p := filepath.Join(dir, "nexus.yaml")
	content := `app:
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
    dsn: "postgres://user:pass@localhost:5432/dbname?sslmode=disable"
`
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp config failed: %v", err)
	}

	cfg, err := LoadNexus(p)
	if err != nil {
		t.Fatalf("LoadNexus failed: %v", err)
	}

	if cfg.Server.GRPCAddr == "" {
		t.Fatalf("expected grpc addr to be set")
	}

	// Missing JWT secret should cause validation error
	bad := filepath.Join(dir, "bad.yaml")
	badContent := `app:
  name: "bifrost-nexus"
  env: dev
  version: "1.0"
server:
  grpc_addr: ":9001"
data:
  database:
    driver: "postgres"
    dsn: "postgres://user:pass@localhost:5432/dbname?sslmode=disable"
`
	if err := os.WriteFile(bad, []byte(badContent), 0o644); err != nil {
		t.Fatalf("write bad config failed: %v", err)
	}

	if _, err := LoadNexus(bad); err == nil {
		t.Fatalf("expected LoadNexus to fail when jwt_secret missing")
	}
}
