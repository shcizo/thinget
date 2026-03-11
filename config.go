package main

import "os"

// Config holds the runtime configuration for the ThinGet proxy.
type Config struct {
	Port     string
	Upstream string
	CacheDir string
}

// LoadConfig reads configuration from environment variables, falling back to
// sensible defaults when variables are not set.
func LoadConfig() Config {
	return Config{
		Port:     getEnv("THINGET_PORT", "5555"),
		Upstream: getEnv("THINGET_UPSTREAM", "https://api.nuget.org"),
		CacheDir: getEnv("THINGET_CACHE_DIR", "/cache"),
	}
}

// getEnv returns the value of the environment variable named by key, or
// fallback if the variable is unset or empty.
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
