package cache

import (
    "sync"
    "time"
)

type Item struct {
    Data       interface{}
    Expiration int64
}

type Cache struct {
    items map[string]*Item
    mu    sync.RWMutex
}

func New() *Cache {
    c := &Cache{
        items: make(map[string]*Item),
    }
    go c.cleanup()
    return c
}

func (c *Cache) Set(key string, value interface{}, ttlSeconds int) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    c.items[key] = &Item{
        Data:       value,
        Expiration: time.Now().Unix() + int64(ttlSeconds),
    }
}

func (c *Cache) Get(key string) (interface{}, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    item, found := c.items[key]
    if !found {
        return nil, false
    }
    
    if time.Now().Unix() > item.Expiration {
        return nil, false
    }
    
    return item.Data, true
}

func (c *Cache) Delete(key string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    delete(c.items, key)
}

func (c *Cache) cleanup() {
    ticker := time.NewTicker(5 * time.Minute)
    for range ticker.C {
        c.mu.Lock()
        for key, item := range c.items {
            if time.Now().Unix() > item.Expiration {
                delete(c.items, key)
            }
        }
        c.mu.Unlock()
    }
}