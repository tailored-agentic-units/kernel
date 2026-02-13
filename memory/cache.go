package memory

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strings"
	"sync"
)

// Cache provides session-scoped access to the memory namespace. It maintains
// an index of all available keys and progressively loads content on demand.
// Reads never trigger I/O. All methods are safe for concurrent use.
type Cache struct {
	store   Store
	cache   map[string][]byte
	index   map[string]bool
	dirty   map[string]bool
	removed map[string]bool
	mu      sync.RWMutex
}

// NewCache creates a Cache backed by the given Store.
func NewCache(store Store) *Cache {
	return &Cache{
		store:   store,
		cache:   make(map[string][]byte),
		index:   make(map[string]bool),
		dirty:   make(map[string]bool),
		removed: make(map[string]bool),
	}
}

func (c *Cache) Bootstrap(ctx context.Context, prefixes ...string) error {
	keys, err := c.store.List(ctx)
	if err != nil {
		return fmt.Errorf("bootstrap index: %w", err)
	}

	c.mu.Lock()
	for _, key := range keys {
		c.index[key] = true
	}
	c.mu.Unlock()

	if len(prefixes) == 0 {
		return nil
	}

	var toLoad []string
	for _, key := range keys {
		for _, prefix := range prefixes {
			if strings.HasPrefix(key, prefix) {
				toLoad = append(toLoad, key)
				break
			}
		}
	}

	if len(toLoad) == 0 {
		return nil
	}

	entries, err := c.store.Load(ctx, toLoad...)
	if err != nil {
		return fmt.Errorf("bootstrap load: %w", err)
	}

	c.mu.Lock()
	for _, e := range entries {
		c.cache[e.Key] = e.Value
	}
	c.mu.Unlock()

	return nil
}

func (c *Cache) Resolve(ctx context.Context, keys ...string) error {
	c.mu.RLock()
	var toLoad []string
	for _, key := range keys {
		if _, cached := c.cache[key]; !cached {
			toLoad = append(toLoad, key)
		}
	}
	c.mu.RUnlock()

	if len(toLoad) == 0 {
		return nil
	}

	entries, err := c.store.Load(ctx, toLoad...)
	if err != nil {
		return fmt.Errorf("resolve: %w", err)
	}

	c.mu.Lock()
	for _, e := range entries {
		c.cache[e.Key] = e.Value
		c.index[e.Key] = true
	}
	c.mu.Unlock()

	return nil
}

func (c *Cache) Flush(ctx context.Context) error {
	c.mu.RLock()
	var toSave []Entry
	for key := range c.dirty {
		if val, ok := c.cache[key]; ok {
			toSave = append(toSave, Entry{Key: key, Value: val})
		}
	}
	var toDelete []string
	for key := range c.removed {
		toDelete = append(toDelete, key)
	}
	c.mu.RUnlock()

	if len(toSave) > 0 {
		if err := c.store.Save(ctx, toSave...); err != nil {
			return fmt.Errorf("flush save: %w", err)
		}
	}

	if len(toDelete) > 0 {
		if err := c.store.Delete(ctx, toDelete...); err != nil {
			return fmt.Errorf("flush delete: %w", err)
		}
	}

	c.mu.Lock()
	c.dirty = make(map[string]bool)
	c.removed = make(map[string]bool)
	c.mu.Unlock()

	return nil
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	val, ok := c.cache[key]
	if !ok {
		return nil, false
	}
	return slices.Clone(val), true
}

func (c *Cache) Set(key string, value []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache[key] = slices.Clone(value)
	c.index[key] = true
	c.dirty[key] = true
	delete(c.removed, key)
}

func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.cache, key)
	delete(c.index, key)
	delete(c.dirty, key)
	c.removed[key] = true
}

func (c *Cache) Has(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.index[key]
}

func (c *Cache) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0, len(c.index))
	for key := range c.index {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (c *Cache) Entries(prefix string) []Entry {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var entries []Entry
	for key, val := range c.cache {
		if strings.HasPrefix(key, prefix) {
			entries = append(entries, Entry{Key: key, Value: slices.Clone(val)})
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Key < entries[j].Key
	})
	return entries
}
