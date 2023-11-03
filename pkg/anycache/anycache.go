package anycache

import (
	"errors"
	"sync"
)

// All available package errors.
var (
	ErrMinCacheSize = errors.New("please provide a cache capacity greater than or equal to 1")
	ErrMaxCacheSize = errors.New("error: requested cache capacity too large, must be < 524,288,000 (100MB)")
)

const maxCacheSize = 5 * 100 * 1024 * 1024 // 500MB

// AnyCache is a thread safe implementation of a simple generic cache.
type AnyCache struct {
	items map[any]any
	mu    *sync.Mutex
}

// Item represents a "record" in the cache.
type Item struct {
	Key   any
	Value any
}

// New returns an AnyCache with the given capacity.
func New(capacity int) (*AnyCache, error) {
	if capacity >= maxCacheSize {
		return nil, ErrMaxCacheSize
	}
	if capacity < 1 {
		return nil, ErrMinCacheSize
	}
	ac := AnyCache{
		items: make(map[any]any, capacity),
		mu:    &sync.Mutex{},
	}
	return &ac, nil
}

// Set sets an item in the cache, returning the Item as well as
// a bool representing whether the set overwrote another value.
func (ac *AnyCache) Set(key any, value any) (Item, bool) {
	var overwritten bool

	ac.mu.Lock()
	defer ac.mu.Unlock()

	_, ok := ac.items[key]
	if ok {
		overwritten = true
	}
	ac.items[key] = value

	return Item{Key: key, Value: value}, overwritten
}

// Get gets an Item by its key and returns both the Item
// as well as whether the call was a cache "hit".
func (ac AnyCache) Get(key any) (Item, bool) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	item, ok := ac.items[key]
	if !ok {
		return Item{}, false
	}

	return Item{Key: key, Value: item}, true
}

func (ac AnyCache) Keys() []any {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	ret := make([]any, 0, len(ac.items))
	for k := range ac.items {
		ret = append(ret, k)
	}

	return ret
}

// Flush clears the AnyCache's inner map.
func (ac *AnyCache) Flush() {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	clear(ac.items)
}

// Len retieves the total count of items in the cache.
func (ac AnyCache) Len() int {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	return len(ac.items)
}
