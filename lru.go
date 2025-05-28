package main

import (
	"container/list"
	"os"
	"path/filepath"
)

type LRUCache struct {
	capacity int
	items    map[string]*list.Element
	order    *list.List // most recent at front
}

type entry struct {
	key string
}

func NewLRUCache(cap int) *LRUCache {
	return &LRUCache{
		capacity: cap,
		items:    make(map[string]*list.Element),
		order:    list.New(),
	}
}

func (c *LRUCache) Touch(key string) {
	// If already exists, move to front
	if el, ok := c.items[key]; ok {
		c.order.MoveToFront(el)
		return
	}

	// Add new
	el := c.order.PushFront(&entry{key})
	c.items[key] = el

	// Evict if over capacity
	if len(c.items) > c.capacity {
		last := c.order.Back()
		if last != nil {
			ent := last.Value.(*entry)
			delete(c.items, ent.key)
			c.order.Remove(last)

			// Remove file from SSD
			os.Remove(filepath.Join("ssd", ent.key))
		}
	}
}
