package server

import (
	"cache_service/internal/cache"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
	"time"
)

type Server struct {
	cache *cache.LRUCache
	log   *slog.Logger
}

func NewServer(cacheInstance *cache.LRUCache, log *slog.Logger) *chi.Mux {
	server := &Server{
		cache: cacheInstance,
		log:   log,
	}
	r := chi.NewRouter()

	r.Use(server.loggingMiddleware)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	r.Route("/api/lru", func(r chi.Router) {
		r.Post("/", server.CreateLRUHandler)
		r.Get("/{key}", server.GetLRUHandler)
		r.Get("/", server.GetAllLRUHandler)
		r.Delete("/{key}", server.DeleteLRUHandler)
		r.Delete("/", server.DeleteAllLRUHandler)
	})

	return r
}

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
