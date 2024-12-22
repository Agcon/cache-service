package server

import (
	"bytes"
	"cache_service/internal/cache"
	"cache_service/internal/logger"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServer_DeleteAll(t *testing.T) {
	cacheInstance := cache.NewLRUCache(10, 0)
	log := logger.NewLogger("DEBUG")
	r := NewServer(cacheInstance, log)

	// Добавляем элемент
	reqBody := []byte(`{"key":"key1","value":"value1"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/lru", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}

	// Удаляем все элементы
	req = httptest.NewRequest(http.MethodDelete, "/api/lru", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", w.Code)
	}

	// Проверяем, что кэш пуст
	req = httptest.NewRequest(http.MethodGet, "/api/lru/key1", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestServer_InvalidPostRequest(t *testing.T) {
	cacheInstance := cache.NewLRUCache(10, 0)
	log := logger.NewLogger("DEBUG")
	r := NewServer(cacheInstance, log)

	// Пустое тело
	req := httptest.NewRequest(http.MethodPost, "/api/lru", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	// Некорректный JSON
	req = httptest.NewRequest(http.MethodPost, "/api/lru", bytes.NewBuffer([]byte(`invalid`)))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestServer_GetAll(t *testing.T) {
	cacheInstance := cache.NewLRUCache(10, 0)
	log := logger.NewLogger("DEBUG")
	r := NewServer(cacheInstance, log)

	// Добавляем элементы
	_ = cacheInstance.Put(nil, "key1", "value1", 0)
	_ = cacheInstance.Put(nil, "key2", "value2", 0)

	// Проверяем GET /api/lru
	req := httptest.NewRequest(http.MethodGet, "/api/lru", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response struct {
		Keys   []string      `json:"keys"`
		Values []interface{} `json:"values"`
	}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(response.Keys) != 2 || len(response.Values) != 2 {
		t.Errorf("expected 2 keys and values, got %d and %d", len(response.Keys), len(response.Values))
	}
}
