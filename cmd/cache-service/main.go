package main

import (
	"cache_service/config"
	"cache_service/internal/cache"
	"cache_service/internal/logger"
	"cache_service/internal/server"
	"log"
	"net/http"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	logg := logger.NewLogger(cfg.LogLevel)

	cacheInstance := cache.NewLRUCache(cfg.CacheSize, cfg.DefaultCacheTTL)

	r := server.NewServer(cacheInstance, logg)

	logg.Info("Starting server",
		"host", cfg.ServerHostPort,
		"log_level", cfg.LogLevel,
	)

	if err := http.ListenAndServe(cfg.ServerHostPort, r); err != nil {
		logg.Error("Server failed to start", "error", err)
	}
}
