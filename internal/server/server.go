package server

import (
	"cache_service/internal/cache"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
	"time"
)

// Server содержит зависимости для работы HTTP-сервера.
type Server struct {
	cache *cache.LRUCache // Экземпляр LRU-кэша
	log   *slog.Logger    // Логгер для записи сообщений
}

// NewServer создаёт HTTP-сервер с поддержкой маршрутов для работы с кэшем.
//
// Параметры:
// - cacheInstance: экземпляр LRU-кэша.
// - log: экземпляр логгера.
func NewServer(cacheInstance *cache.LRUCache, log *slog.Logger) *chi.Mux {
	server := &Server{
		cache: cacheInstance,
		log:   log,
	}
	r := chi.NewRouter()

	// Middleware
	r.Use(server.loggingMiddleware) // Логирование входящих запросов
	r.Use(middleware.Recoverer)     // Перехват паник
	r.Use(middleware.RequestID)     // Генерация Request ID

	//Маршруты
	r.Route("/api/lru", func(r chi.Router) {
		r.Post("/", server.CreateLRUHandler)
		r.Get("/{key}", server.GetLRUHandler)
		r.Get("/", server.GetAllLRUHandler)
		r.Delete("/{key}", server.DeleteLRUHandler)
		r.Delete("/", server.DeleteAllLRUHandler)
	})

	return r
}

// loggingMiddleware логирует все входящие HTTP-запросы.
//
// Логи включают:
// - Метод запроса.
// - Путь запроса.
// - Время обработки.
//
// Логи пишутся на уровне DEBUG.
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)

		s.log.Debug("Request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"duration", duration.String(),
		)
	})
}
