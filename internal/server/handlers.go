package server

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"net/http"
	"time"
)

// CreateLRUHandler обрабатывает POST-запрос на добавление элемента в кэш.
//
// Метод:
// - POST /api/lru
//
// Тело запроса (JSON):
// - key (string): Ключ элемента.
// - value (interface{}): Значение элемента.
// - ttl_seconds (int, optional): Время жизни элемента в секундах.
//
// Ответы:
// - 201 Created: Элемент успешно добавлен.
// - 400 Bad Request: Некорректный запрос.
// - 500 Internal Server Error: Ошибка сервера.
func (s *Server) CreateLRUHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	s.log.Info("Processing request", "method", r.Method, "path", r.URL.Path)
	select {
	case <-ctx.Done():
		s.log.Warn("Request cancelled", "method", r.Method, "path", r.URL.Path)
		http.Error(w, "request cancelled", http.StatusInternalServerError)
		return
	default:
	}

	var createRequest struct {
		Key        string      `json:"key"`
		Value      interface{} `json:"value"`
		TTLSeconds int64       `json:"ttl_seconds,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&createRequest); err != nil {
		s.log.Error("Invalid request body", "error", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := s.cache.Put(ctx, createRequest.Key, createRequest.Value, time.Duration(createRequest.TTLSeconds)*time.Second); err != nil {
		s.log.Error("Failed to put key in cache", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	s.log.Info("Key added to cache", "key", createRequest.Key)
	w.WriteHeader(http.StatusCreated)
}

// GetLRUHandler обрабатывает GET-запрос на получение элемента по ключу.
//
// Метод:
// - GET /api/lru/{key}
//
// Параметры пути:
// - key (string): Ключ элемента.
//
// Ответы:
// - 200 OK: Успешный ответ с данными элемента.
// - 404 Not Found: Ключ не найден или истёк срок действия.
// - 500 Internal Server Error: Ошибка сервера.
func (s *Server) GetLRUHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	s.log.Info("Processing request", "method", r.Method, "path", r.URL.Path)
	select {
	case <-ctx.Done():
		s.log.Warn("Request cancelled", "method", r.Method, "path", r.URL.Path)
		http.Error(w, "request cancelled", http.StatusInternalServerError)
		return
	default:
	}
	key := chi.URLParam(r, "key")
	value, expiresAt, err := s.cache.Get(ctx, key)
	if err != nil {
		s.log.Error("Failed to get key from cache", "error", err)
		http.Error(w, err.Error(), http.StatusNotFound)
	}

	s.log.Info("Key retrieved from cache", "key", key, "expires_at", expiresAt)
	response := struct {
		Key       string      `json:"key"`
		Value     interface{} `json:"value"`
		ExpiresAt int64       `json:"expires_at"`
	}{
		Key:       key,
		Value:     value,
		ExpiresAt: expiresAt.Unix(),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.log.Error("Failed to encode response", "error", err)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// GetAllLRUHandler обрабатывает GET-запрос на получение всех элементов из кэша.
//
// Метод:
// - GET /api/lru
//
// Ответы:
// - 200 OK: Успешный ответ с данными всех элементов.
// - 204 No Content: Кэш пуст.
// - 500 Internal Server Error: Ошибка сервера.
func (s *Server) GetAllLRUHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	s.log.Info("Processing request", "method", r.Method, "path", r.URL.Path)
	select {
	case <-ctx.Done():
		s.log.Warn("Request cancelled", "method", r.Method, "path", r.URL.Path)
		http.Error(w, "request cancelled", http.StatusInternalServerError)
		return
	default:
	}

	keys, values, err := s.cache.GetAll(ctx)
	if err != nil {
		s.log.Error("Failed to get all keys from cache", "error", err)
		http.Error(w, err.Error(), http.StatusNoContent)
	}

	s.log.Info("All keys retrieved from cache", "count", len(keys))
	response := struct {
		Keys   []string      `json:"keys"`
		Values []interface{} `json:"values"`
	}{
		Keys:   keys,
		Values: values,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.log.Error("Failed to encode response", "error", err)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// DeleteLRUHandler обрабатывает DELETE-запрос на удаление элемента по ключу.
//
// Метод:
// - DELETE /api/lru/{key}
//
// Параметры пути:
// - key (string): Ключ элемента.
//
// Ответы:
// - 204 No Content: Элемент успешно удалён.
// - 404 Not Found: Ключ не найден.
// - 500 Internal Server Error: Ошибка сервера.
func (s *Server) DeleteLRUHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	s.log.Info("Processing request", "method", r.Method, "path", r.URL.Path)
	select {
	case <-ctx.Done():
		s.log.Warn("Request cancelled", "method", r.Method, "path", r.URL.Path)
		http.Error(w, "request cancelled", http.StatusInternalServerError)
		return
	default:
	}
	key := chi.URLParam(r, "key")
	_, err := s.cache.Evict(ctx, key)
	if err != nil {
		s.log.Error("Failed to delete key from cache", "error", err)
		http.Error(w, err.Error(), http.StatusNotFound)
	}
	s.log.Info("Key deleted from cache", "key", key)
	w.WriteHeader(http.StatusNoContent)
}

// DeleteAllLRUHandler обрабатывает DELETE-запрос на удаление всех элементов из кэша.
//
// Метод:
// - DELETE /api/lru
//
// Ответы:
// - 204 No Content: Все элементы успешно удалены.
// - 500 Internal Server Error: Ошибка сервера.
func (s *Server) DeleteAllLRUHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	s.log.Info("Processing request", "method", r.Method, "path", r.URL.Path)
	select {
	case <-ctx.Done():
		s.log.Warn("Request cancelled", "method", r.Method, "path", r.URL.Path)
		http.Error(w, "request cancelled", http.StatusInternalServerError)
		return
	default:
	}

	if err := s.cache.EvictAll(ctx); err != nil {
		s.log.Error("Failed to delete all keys from cache", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	s.log.Info("All keys successfully deleted from cache")
	w.WriteHeader(http.StatusNoContent)
}
