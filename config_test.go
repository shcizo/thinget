package main

import (
	"os"
	"testing"
)

func TestLoadConfigDefaults(t *testing.T) {
	os.Unsetenv("THINGET_PORT")
	os.Unsetenv("THINGET_UPSTREAM")
	os.Unsetenv("THINGET_CACHE_DIR")

	cfg := LoadConfig()

	if cfg.Port != "5555" {
		t.Errorf("expected port 5555, got %s", cfg.Port)
	}
	if cfg.Upstream != "https://api.nuget.org" {
		t.Errorf("expected default upstream, got %s", cfg.Upstream)
	}
	if cfg.CacheDir != "/cache" {
		t.Errorf("expected /cache, got %s", cfg.CacheDir)
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	t.Setenv("THINGET_PORT", "9999")
	t.Setenv("THINGET_UPSTREAM", "https://custom.nuget.org")
	t.Setenv("THINGET_CACHE_DIR", "/tmp/nuget")

	cfg := LoadConfig()

	if cfg.Port != "9999" {
		t.Errorf("expected port 9999, got %s", cfg.Port)
	}
	if cfg.Upstream != "https://custom.nuget.org" {
		t.Errorf("expected custom upstream, got %s", cfg.Upstream)
	}
	if cfg.CacheDir != "/tmp/nuget" {
		t.Errorf("expected /tmp/nuget, got %s", cfg.CacheDir)
	}
}
