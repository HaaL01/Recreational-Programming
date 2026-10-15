package main

import (
	"net/http"
	"sync"
)

// CachedResponse holds everything we need to replay a response later.
type CachedResponse struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

// Cache is an in-memory store of responses, keyed by request.
type Cache struct {
	mu      sync.Mutex
	entries map[string]CachedResponse
}

func NewCache() *Cache {
	return &Cache{entries: make(map[string]CachedResponse)}
}

func (c *Cache) Get(key string) (CachedResponse, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	resp, ok := c.entries[key]
	return resp, ok
}

func (c *Cache) Set(key string, resp CachedResponse) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = resp
}
