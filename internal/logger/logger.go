package logger

import (
	"log/slog"
	"os"
)

// NewLogger создаёт новый экземпляр структурированного логгера.
//
// Параметры:
// - level: строковый уровень логирования (DEBUG, INFO, WARN, ERROR).
//
// Возвращает:
// - Экземпляр логгера slog.
func NewLogger(level string) *slog.Logger {
	var lvl slog.Level
	switch level {
	case "DEBUG":
		lvl = slog.LevelDebug
	case "INFO":
		lvl = slog.LevelInfo
	case "WARN":
		lvl = slog.LevelWarn
	case "ERROR":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelWarn
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: lvl})
	return slog.New(handler)
}
