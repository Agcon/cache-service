package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	os.Setenv("CACHE_SIZE", "50")
	defer os.Unsetenv("CACHE_SIZE")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.CacheSize != 50 {
		t.Errorf("expected cache size 50, got %v", cfg.CacheSize)
	}
}
