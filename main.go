package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	cfg := LoadConfig()

	if err := os.MkdirAll(cfg.CacheDir, 0755); err != nil {
		log.Fatalf("failed to create cache dir: %v", err)
	}

	cache := NewCache(cfg.CacheDir)
	proxy := NewProxy(cfg.Upstream)
	handler := NewHandler(cache, proxy, cfg)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("ThinGet starting on %s (upstream: %s, cache: %s)", addr, cfg.Upstream, cfg.CacheDir)
	log.Fatal(http.ListenAndServe(addr, handler))
}
