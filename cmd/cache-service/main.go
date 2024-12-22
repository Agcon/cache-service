// Package main запускает HTTP-сервер для работы с LRU-кэшем.
// Сервер поддерживает операции добавления, получения и удаления данных из кэша.
// Конфигурация задаётся через флаги, переменные окружения или значения по умолчанию.
//
// Пример запуска:
//
//	go run cmd/app/main.go -cache-size=100 -default-cache-ttl=60s -server-host-port=localhost:8080
package main

import (
	"cache_service/config"
	"cache_service/internal/cache"
	"cache_service/internal/logger"
	"cache_service/internal/server"
	"github.com/joho/godotenv"
	"log"
	"net/http"
)

func main() {
	// Загружаем переменные окружения из файла .env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Загружаем конфигурацию
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	// Инициализируем логгер
	logg := logger.NewLogger(cfg.LogLevel)

	// Инициализируем кэш
	cacheInstance := cache.NewLRUCache(cfg.CacheSize, cfg.DefaultCacheTTL)

	// Настраиваем сервер
	r := server.NewServer(cacheInstance, logg)

	// Запуск HTTP-сервера
	logg.Info("Starting server",
		"host", cfg.ServerHostPort,
		"log_level", cfg.LogLevel,
	)

	if err := http.ListenAndServe(cfg.ServerHostPort, r); err != nil {
		logg.Error("Server failed to start", "error", err)
	}
}
