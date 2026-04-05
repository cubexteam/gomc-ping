package cache

import (
	"sync"
	"time"

	"github.com/cubexteam/gomc-ping/models"
)

type item struct {
	response   *models.Response
	expiration int64
}

type Cache struct {
	mu         sync.RWMutex
	items      map[string]item
	defaultTTL time.Duration
	stop       chan struct{}
}

func New(defaultTTL, cleanupInterval time.Duration) *Cache {
	c := &Cache{
		items:      make(map[string]item),
		defaultTTL: defaultTTL,
		stop:       make(chan struct{}),
	}
	go c.janitor(cleanupInterval)
	return c
}

func (c *Cache) Get(key string) (*models.Response, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	it, ok := c.items[key]
	if !ok {
		return nil, false
	}

	if time.Now().UnixNano() > it.expiration {
		return nil, false
	}

	return it.response, true
}

func (c *Cache) Set(key string, response *models.Response) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = item{
		response:   response,
		expiration: time.Now().Add(c.defaultTTL).UnixNano(),
	}
}

func (c *Cache) Close() {
	close(c.stop)
}

func (c *Cache) janitor(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.mu.Lock()
			now := time.Now().UnixNano()
			for k, v := range c.items {
				if now > v.expiration {
					delete(c.items, k)
				}
			}
			c.mu.Unlock()
		case <-c.stop:
			return
		}
	}
}
