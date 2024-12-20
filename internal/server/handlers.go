package server

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"net/http"
	"time"
)

func (s *Server) CreateLRUHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	select {
	case <-ctx.Done():
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
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := s.cache.Put(ctx, createRequest.Key, createRequest.Value, time.Duration(createRequest.TTLSeconds)*time.Second); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *Server) GetLRUHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	select {
	case <-ctx.Done():
		http.Error(w, "request cancelled", http.StatusInternalServerError)
		return
	default:
	}
	key := chi.URLParam(r, "key")
	value, expiresAt, err := s.cache.Get(ctx, key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
	}
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
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (s *Server) GetAllLRUHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	select {
	case <-ctx.Done():
		http.Error(w, "request cancelled", http.StatusInternalServerError)
		return
	default:
	}

	keys, values, err := s.cache.GetAll(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNoContent)
	}

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
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (s *Server) DeleteLRUHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	select {
	case <-ctx.Done():
		http.Error(w, "request cancelled", http.StatusInternalServerError)
		return
	default:
	}
	key := chi.URLParam(r, "key")
	_, err := s.cache.Evict(ctx, key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) DeleteAllLRUHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	select {
	case <-ctx.Done():
		http.Error(w, "request cancelled", http.StatusInternalServerError)
		return
	default:
	}

	if err := s.cache.EvictAll(ctx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusNoContent)
}
