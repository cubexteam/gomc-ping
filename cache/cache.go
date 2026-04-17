package cache

import (
	"hash/fnv"
	"sync"
	"time"

	"github.com/cubexteam/gomc-ping/models"
)

const numShards = 16

type item struct {
	response   *models.Response
	expiration int64
}

type shard struct {
	mu         sync.RWMutex
	items      map[string]item
}

// Cache is a sharded, TTL-based in-memory cache.
// Sharding reduces lock contention when many goroutines access the cache
// concurrently (e.g. batch pinging hundreds of servers).
type Cache struct {
	shards     [numShards]shard
	defaultTTL time.Duration
	stop       chan struct{}
	once       sync.Once
}

func New(defaultTTL, cleanupInterval time.Duration) *Cache {
	c := &Cache{
		defaultTTL: defaultTTL,
		stop:       make(chan struct{}),
	}
	for i := range c.shards {
		c.shards[i].items = make(map[string]item)
	}
	go c.janitor(cleanupInterval)
	return c
}

func (c *Cache) shardFor(key string) *shard {
	h := fnv.New32a()
	h.Write([]byte(key))
	return &c.shards[h.Sum32()%numShards]
}

func (c *Cache) Get(key string) (*models.Response, bool) {
	s := c.shardFor(key)
	s.mu.RLock()
	defer s.mu.RUnlock()

	it, ok := s.items[key]
	if !ok || time.Now().UnixNano() > it.expiration {
		return nil, false
	}
	return it.response, true
}

func (c *Cache) Set(key string, response *models.Response) {
	s := c.shardFor(key)
	s.mu.Lock()
	defer s.mu.Unlock()

	s.items[key] = item{
		response:   response,
		expiration: time.Now().Add(c.defaultTTL).UnixNano(),
	}
}

// Close stops the background janitor. Safe to call multiple times.
func (c *Cache) Close() {
	c.once.Do(func() { close(c.stop) })
}

func (c *Cache) janitor(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now().UnixNano()
			for i := range c.shards {
				s := &c.shards[i]
				s.mu.Lock()
				for k, v := range s.items {
					if now > v.expiration {
						delete(s.items, k)
					}
				}
				s.mu.Unlock()
			}
		case <-c.stop:
			return
		}
	}
}
