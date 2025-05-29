package main

import (
	"container/list"
	"os"
	"path/filepath"
	"sync"
)

type LRUCache struct {
	maxFiles int
	maxBytes int
	curBytes int

	items map[string]*list.Element
	order *list.List
	mu    sync.Mutex
}

type entry struct {
	hash string
	size int
}

func NewLRUCache(maxFiles, maxBytes int) *LRUCache {
	return &LRUCache{
		maxFiles: maxFiles,
		maxBytes: maxBytes,
		items:    make(map[string]*list.Element),
		order:    list.New(),
	}
}

func (c *LRUCache) Touch(hash string, size int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Already in cache → update recency
	if el, ok := c.items[hash]; ok {
		c.order.MoveToFront(el)
		return
	}

	// New cache entry
	e := &entry{hash: hash, size: size}
	el := c.order.PushFront(e)
	c.items[hash] = el
	c.curBytes += size

	// Evict until under limits
	for len(c.items) > c.maxFiles || c.curBytes > c.maxBytes {
		last := c.order.Back()
		if last == nil {
			break
		}
		old := last.Value.(*entry)
		delete(c.items, old.hash)
		c.order.Remove(last)
		c.curBytes -= old.size
		os.Remove(filepath.Join("ssd", old.hash))
	}
}
