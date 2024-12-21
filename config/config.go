package config

import (
	"flag"
	"github.com/caarlos0/env/v9"
	_ "github.com/caarlos0/env/v9"
	"time"
)

// Config описывает параметры конфигурации приложения.
type Config struct {
	ServerHostPort  string        `env:"SERVER_HOST_PORT" envDefault:"localhost:8080"` // Адрес и порт сервера
	CacheSize       int           `env:"CACHE_SIZE" envDefault:"10"`                   // Размер кэша
	DefaultCacheTTL time.Duration `env:"DEFAULT_CACHE_TTL" envDefault:"1m"`            // Время жизни элемента по умолчанию
	LogLevel        string        `env:"LOG_LEVEL" envDefault:"WARN"`                  // Уровень логирования
}

// LoadConfig загружает конфигурацию из флагов, переменных окружения или значений по умолчанию.
//
// Возвращает:
// - Указатель на структуру Config с заполненными параметрами.
// - Ошибку, если загрузка конфигурации завершилась неудачно.
func LoadConfig() (*Config, error) {
	hostPort := flag.String("server-host-port", "", "Server host and port (e.g., localhost:8080)")
	cacheSize := flag.Int("cache-size", 0, "Cache size")
	defaultTTL := flag.Duration("default-cache-ttl", 0, "Default cache TTL (e.g., 1m, 30s)")
	logLevel := flag.String("log-level", "", "Log level (e.g., DEBUG, INFO, WARN)")

	flag.Parse()

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	if *hostPort != "" {
		cfg.ServerHostPort = *hostPort
	}
	if *cacheSize != 0 {
		cfg.CacheSize = *cacheSize
	}
	if *defaultTTL != 0 {
		cfg.DefaultCacheTTL = *defaultTTL
	}
	if *logLevel != "" {
		cfg.LogLevel = *logLevel
	}

	return cfg, nil
}
