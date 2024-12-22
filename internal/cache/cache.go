package cache

import (
	"context"
	"errors"
	"sync"
	"time"
)

// Ошибки, которые могут возникнуть при работе с кешем
var (
	errEmptyKey    = errors.New("key cannot be empty")    // Ошибка для пустого ключа
	errNegativeTTL = errors.New("ttl cannot be negative") // Ошибка для отрицательного TTL
	errKeyNotFound = errors.New("key not found")          // Ошибка для отсутствующего ключа
	errExpiredKey  = errors.New("key expired")            // Ошибка для истекшего ключа
	errNilNode     = errors.New("node is nil")            // Ошибка для пустого узла
	errEmptyCache  = errors.New("cache is empty")         // Ошибка для пустого кеша
)

// Node представляет собой элемент в кеше, содержащий ключ, значение, время жизни (TTL),
// а также ссылки на предыдущий и следующий элементы в двусвязном списке.
type Node struct {
	key   string      // Ключ элемента в кеше
	value interface{} // Значение элемента
	TTL   time.Time   // Время истечения срока жизни элемента
	prev  *Node       // Указатель на предыдущий элемент в списке
	next  *Node       // Указатель на следующий элемент в списке
}

// LRUCache представляет собой структуру кеша с алгоритмом LRU, поддерживающего TTL для элементов.
type LRUCache struct {
	head       *Node            // Указатель на первый элемент в списке
	tail       *Node            // Указатель на последний элемент в списке
	cache      map[string]*Node // Карта для хранения элементов кеша по ключу
	capacity   int              // Максимальная ёмкость кеша
	defaultTTL time.Duration    // Значение по умолчанию для TTL
	mutex      sync.RWMutex     // Мьютекс для безопасного доступа к кешу
}

// NewLRUCache создает новый LRU кеш с заданной емкостью и значением по умолчанию для TTL.
// Возвращает указатель на новый объект LRUCache.
func NewLRUCache(capacity int, defaultTTL time.Duration) *LRUCache {
	return &LRUCache{
		cache:      make(map[string]*Node),
		capacity:   capacity,
		defaultTTL: defaultTTL,
	}
}

// addNode добавляет новый узел в начало списка.
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

// moveToHead перемещает указанный узел в начало списка (в начало списка недавно использованных элементов).
func (c *LRUCache) moveToHead(node *Node) {
	c.removeNode(node)
	c.addNode(node)
}

// removeNode удаляет узел из списка.
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
	node.prev = nil
	node.next = nil
}

// Put добавляет новый элемент в кеш с заданным ключом, значением и TTL.
// Если элемент с таким ключом уже существует, его значение обновляется и TTL сбрасывается.
// Если кеш переполнен, удаляется наименее недавно использованный элемент.
func (c *LRUCache) Put(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if ctx == nil {
		ctx = context.Background()
	}

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

// Get возвращает значение по ключу из кеша. Также возвращается время истечения срока жизни элемента (TTL).
// Если элемент не найден или его TTL истек, возвращается ошибка.
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

// GetAll возвращает все ключи и значения из кеша.
func (c *LRUCache) GetAll(ctx context.Context) (keys []string, values []interface{}, err error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}

	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if len(c.cache) == 0 {
		return nil, nil, errEmptyCache
	}

	now := time.Now()
	for node := c.head; node != nil; {
		next := node.next
		select {
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		default:
			if now.After(node.TTL) {
				delete(c.cache, node.key)
				c.removeNode(node)
			} else {
				keys = append(keys, node.key)
				values = append(values, node.value)
			}
			node = next
		}
	}
	return keys, values, nil
}

// Evict удаляет элемент из кеша по ключу и возвращает его значение.
// Если элемент не найден, возвращается ошибка.
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

// EvictAll очищает весь кеш.
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

// getTTL возвращает TTL для элемента. Если TTL равен 0, используется значение по умолчанию.
func (c *LRUCache) getTTL(ttl time.Duration) time.Duration {
	if ttl == 0 {
		return c.defaultTTL
	}
	return ttl
}
