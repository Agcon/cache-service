package cache

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestLRUCache_PutAndGet(t *testing.T) {
	c := NewLRUCache(2, 1*time.Minute)

	// Добавляем элемент
	err := c.Put(context.Background(), "key1", "value1", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Получаем элемент
	val, expiresAt, err := c.Get(context.Background(), "key1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "value1" {
		t.Errorf("expected value1, got %v", val)
	}
	if expiresAt.Before(time.Now()) {
		t.Errorf("expiresAt is in the past: %v", expiresAt)
	}

	// Обновляем элемент
	err = c.Put(context.Background(), "key1", "newValue", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, _, err = c.Get(context.Background(), "key1")
	if val != "newValue" {
		t.Errorf("expected newValue, got %v", val)
	}
}

func TestLRUCache_KeyExpired(t *testing.T) {
	c := NewLRUCache(1, 1*time.Millisecond)

	// Добавляем элемент
	err := c.Put(context.Background(), "key1", "value1", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Ждём, чтобы TTL истёк
	time.Sleep(2 * time.Millisecond)

	// Проверяем истечение
	_, _, err = c.Get(context.Background(), "key1")
	if !errors.Is(err, errExpiredKey) {
		t.Errorf("expected ErrKeyExpired, got %v", err)
	}
}

func TestLRUCache_EvictAll(t *testing.T) {
	c := NewLRUCache(3, 1*time.Minute)

	// Добавляем элементы
	_ = c.Put(context.Background(), "key1", "value1", 0)
	_ = c.Put(context.Background(), "key2", "value2", 0)
	_ = c.Put(context.Background(), "key3", "value3", 0)

	// Полная очистка
	err := c.EvictAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Проверяем, что кэш пуст
	_, _, err = c.Get(context.Background(), "key1")
	if !errors.Is(err, errKeyNotFound) {
		t.Errorf("expected ErrKeyNotFound, got %v", err)
	}
}

func TestLRUCache_GetAll_RemoveExpired(t *testing.T) {
	cache := NewLRUCache(3, 1*time.Second)

	_ = cache.Put(context.Background(), "key1", "value1", 500*time.Millisecond)
	_ = cache.Put(context.Background(), "key2", "value2", 2*time.Second)

	time.Sleep(1 * time.Second)

	keys, _, err := cache.GetAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(keys) != 1 || keys[0] != "key2" {
		t.Errorf("expected 1 valid key (key2), got keys=%v", keys)
	}
}
