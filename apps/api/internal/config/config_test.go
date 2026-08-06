package config

import "testing"

func TestLoad(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("REDIS_URL", "redis://example")
	t.Setenv("API_ADDRESS", "127.0.0.1:9000")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Address != "127.0.0.1:9000" {
		t.Fatalf("Address = %q", cfg.Address)
	}
}

func TestLoadRequiresDependencies(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("REDIS_URL", "")
	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil")
	}
}
