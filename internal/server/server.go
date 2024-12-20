package server

import (
	"cache_service/internal/cache"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	cache *cache.LRUCache
}

func NewServer(cacheInstance *cache.LRUCache) *chi.Mux {
	server := &Server{
		cache: cacheInstance,
	}
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api/lru", func(r chi.Router) {
		r.Post("/", server.CreateLRUHandler)
		r.Get("/{key}", server.GetLRUHandler)
		r.Get("/", server.GetAllLRUHandler)
		r.Delete("/{key}", server.DeleteLRUHandler)
		r.Delete("/", server.DeleteAllLRUHandler)
	})

	return r
}
