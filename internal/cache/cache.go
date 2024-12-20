package cache

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	errEmptyKey    = errors.New("key cannot be empty")
	errNegativeTTL = errors.New("ttl cannot be negative")
	errKeyNotFound = errors.New("key not found")
	errExpiredKey  = errors.New("key expired")
	errNilNode     = errors.New("node is nil")
	errEmptyCache  = errors.New("cache is empty")
)

type Node struct {
	key   string
	value interface{}
	TTL   time.Time
	prev  *Node
	next  *Node
}

type LRUCache struct {
	head       *Node
	tail       *Node
	cache      map[string]*Node
	capacity   int
	defaultTTL time.Duration
	mutex      sync.RWMutex
}

func NewLRUCache(capacity int, defaultTTL time.Duration) *LRUCache {
	return &LRUCache{
		cache:      make(map[string]*Node),
		capacity:   capacity,
		defaultTTL: defaultTTL,
	}
}

func (c *LRUCache) addNode(node *Node) {
	node.next = c.head
	if c.head != nil {
		c.head.prev = node
	}
	c.head = node
	if c.tail == nil {
		c.tail = node
	}
}

func (c *LRUCache) moveToHead(node *Node) {
	c.removeNode(node)
	c.addNode(node)
}

func (c *LRUCache) removeNode(node *Node) {
	if node.prev != nil {
		node.prev.next = node.next
	} else {
		c.head = node.next
	}

	if node.next != nil {
		node.next.prev = node.prev
	} else {
		c.tail = node.prev
	}
}

func (c *LRUCache) Put(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if key == "" {
		return errEmptyKey
	}

	if ttl < 0 {
		return errNegativeTTL
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	if node, exists := c.cache[key]; exists {
		node.value = value
		node.TTL = time.Now().Add(c.getTTL(ttl))
		c.moveToHead(node)
		return nil
	}

	if len(c.cache) >= c.capacity {
		if c.tail == nil {
			return errNilNode
		}
		delete(c.cache, c.tail.key)
		c.removeNode(c.tail)
	}

	newNode := &Node{
		key:   key,
		value: value,
		TTL:   time.Now().Add(c.getTTL(ttl)),
	}
	c.cache[key] = newNode
	c.addNode(newNode)
	return nil
}

func (c *LRUCache) Get(ctx context.Context, key string) (value interface{}, expiresAt time.Time, err error) {
	if err := ctx.Err(); err != nil {
		return nil, time.Time{}, err
	}

	if key == "" {
		return nil, time.Time{}, errEmptyKey
	}

	c.mutex.RLock()
	defer c.mutex.RUnlock()

	node, exists := c.cache[key]
	if !exists {
		return nil, time.Time{}, errKeyNotFound
	}

	if time.Now().After(node.TTL) {
		delete(c.cache, key)
		return nil, time.Time{}, errExpiredKey
	}

	if node == nil {
		return nil, time.Time{}, errNilNode
	}

	return node.value, node.TTL, nil
}

func (c *LRUCache) GetAll(ctx context.Context) (keys []string, values []interface{}, err error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}

	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if len(c.cache) == 0 {
		return nil, nil, errEmptyCache
	}

	for node := c.head; node != nil; node = node.next {
		select {
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		default:
			keys = append(keys, node.key)
			values = append(values, node.value)
		}
	}
	return keys, values, nil
}

func (c *LRUCache) Evict(ctx context.Context, key string) (value interface{}, err error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if key == "" {
		return nil, errEmptyKey
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	node, exists := c.cache[key]
	if !exists {
		return nil, errKeyNotFound
	}

	if node == nil {
		return nil, errNilNode
	}

	delete(c.cache, key)
	c.removeNode(node)
	return node.value, nil
}

func (c *LRUCache) EvictAll(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	if len(c.cache) == 0 {
		return errEmptyCache
	}

	c.cache = make(map[string]*Node)
	c.head, c.tail = nil, nil
	return nil
}

func (c *LRUCache) getTTL(ttl time.Duration) time.Duration {
	if ttl == 0 {
		return c.defaultTTL
	}
	return ttl
}
