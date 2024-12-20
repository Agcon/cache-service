package main

import (
	"cache_service/config"
	"cache_service/internal/cache"
	"cache_service/internal/server"
	"log"
	"net/http"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	cacheInstance := cache.NewLRUCache(cfg.CacheSize, cfg.DefaultCacheTTL)

	r := server.NewServer(cacheInstance)

	log.Printf("Starting server on %s with log level %s", cfg.ServerHostPort, cfg.LogLevel)
	if err := http.ListenAndServe(cfg.ServerHostPort, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
