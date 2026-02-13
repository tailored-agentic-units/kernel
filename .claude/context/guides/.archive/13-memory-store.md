# 13 - Memory Store Interface and Filesystem Implementation

## Problem Context

The memory subsystem is the single context composition pipeline for the TAU kernel. It provides a hierarchical key-value namespace backed by pluggable storage, with session-scoped caching and progressive loading. The kernel loads context at session initialization and accesses it locally — no constant I/O to external systems.

The memory package currently contains only a README.md skeleton. This implementation establishes the full architecture: persistence interface (Store), filesystem backend (FileStore), and session-scoped cache with progressive loading (Cache).

## Architecture Approach

Two types work together:

- **Store** — stateless persistence abstraction. Translates between external storage and internal `[]byte` entries keyed by `/`-separated paths. The kernel requires `[]byte` values; external systems handle transformations.
- **Cache** — session-scoped cache. Bootstraps an index of available keys, progressively loads content on demand, tracks dirty state, and flushes changes back through the Store.

Keys are `/`-separated hierarchical paths that map 1:1 to filesystem paths in the FileStore. Three namespace conventions: `memory/` (core agent memory), `skills/` (skill definitions), `agents/` (sub-agent profiles).

Progressive loading: Bootstrap populates the key index and eagerly loads specified prefixes. Additional content is resolved on demand via `Resolve`. Reads never trigger I/O.

## Implementation

### Step 1: Sentinel Errors (`memory/errors.go`)

New file:

```go
package memory

import "errors"

var (
	ErrKeyNotFound = errors.New("key not found")
	ErrLoadFailed  = errors.New("load failed")
	ErrSaveFailed  = errors.New("save failed")
)
```

### Step 2: Entry Type and Namespace Constants (`memory/entry.go`)

New file:

```go
package memory

const (
	NamespaceMemory = "memory"
	NamespaceSkills = "skills"
	NamespaceAgents = "agents"
)

type Entry struct {
	Key   string
	Value []byte
}
```

### Step 3: Store Interface (`memory/store.go`)

New file:

```go
package memory

import "context"

type Store interface {
	List(ctx context.Context) ([]string, error)
	Load(ctx context.Context, keys ...string) ([]Entry, error)
	Save(ctx context.Context, entries ...Entry) error
	Delete(ctx context.Context, keys ...string) error
}
```

### Step 4: FileStore Implementation (`memory/filestore.go`)

New file. The FileStore maps keys 1:1 to filesystem paths under a root directory.

```go
package memory

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type fileStore struct {
	root string
}

func NewFileStore(root string) Store {
	return &fileStore{root: root}
}

func (s *fileStore) List(_ context.Context) ([]string, error) {
	var keys []string

	err := filepath.WalkDir(s.root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) && path == s.root {
				return fs.SkipAll
			}
			return err
		}

		// Skip hidden files and directories
		if strings.HasPrefix(d.Name(), ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(s.root, path)
		if err != nil {
			return err
		}
		keys = append(keys, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrLoadFailed, err)
	}

	return keys, nil
}

func (s *fileStore) Load(_ context.Context, keys ...string) ([]Entry, error) {
	entries := make([]Entry, 0, len(keys))

	for _, key := range keys {
		path := filepath.Join(s.root, filepath.FromSlash(key))
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("%w: %s", ErrKeyNotFound, key)
			}
			return nil, fmt.Errorf("%w: %s: %v", ErrLoadFailed, key, err)
		}
		entries = append(entries, Entry{Key: key, Value: data})
	}

	return entries, nil
}

func (s *fileStore) Save(_ context.Context, entries ...Entry) error {
	for _, e := range entries {
		path := filepath.Join(s.root, filepath.FromSlash(e.Key))

		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("%w: %s: %v", ErrSaveFailed, e.Key, err)
		}

		// Atomic write: temp file in same directory, then rename
		tmp, err := os.CreateTemp(dir, ".tmp-*")
		if err != nil {
			return fmt.Errorf("%w: %s: %v", ErrSaveFailed, e.Key, err)
		}
		tmpName := tmp.Name()

		if _, err := tmp.Write(e.Value); err != nil {
			tmp.Close()
			os.Remove(tmpName)
			return fmt.Errorf("%w: %s: %v", ErrSaveFailed, e.Key, err)
		}
		if err := tmp.Close(); err != nil {
			os.Remove(tmpName)
			return fmt.Errorf("%w: %s: %v", ErrSaveFailed, e.Key, err)
		}

		if err := os.Rename(tmpName, path); err != nil {
			os.Remove(tmpName)
			return fmt.Errorf("%w: %s: %v", ErrSaveFailed, e.Key, err)
		}
	}

	return nil
}

func (s *fileStore) Delete(_ context.Context, keys ...string) error {
	for _, key := range keys {
		path := filepath.Join(s.root, filepath.FromSlash(key))
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("delete failed: %s: %w", key, err)
		}

		// Clean up empty parent directories up to root
		dir := filepath.Dir(path)
		for dir != s.root {
			if err := os.Remove(dir); err != nil {
				break // Not empty or other error — stop climbing
			}
			dir = filepath.Dir(dir)
		}
	}

	return nil
}
```

### Step 5: Cache Implementation (`memory/cache.go`)

New file. The Cache is the session-scoped cache with progressive loading.

```go
package memory

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strings"
	"sync"
)

type Cache struct {
	store   Store
	cache   map[string][]byte
	index   map[string]bool
	dirty   map[string]bool
	removed map[string]bool
	mu      sync.RWMutex
}

func NewCache(store Store) *Cache {
	return &Cache{
		store:   store,
		cache:   make(map[string][]byte),
		index:   make(map[string]bool),
		dirty:   make(map[string]bool),
		removed: make(map[string]bool),
	}
}

// Bootstrap loads the store index and caches entries matching the given prefixes.
// With no prefixes, only the index is populated.
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

	// Collect keys matching any prefix
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

// Resolve loads specific entries from the store into the cache.
// Already-cached keys are skipped.
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

// Flush persists dirty entries and removes deleted entries through the store.
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
```

## Validation Criteria

- [ ] `Store` interface defined in `memory/store.go`
- [ ] `FileStore` implementation in `memory/filestore.go`
- [ ] `Cache` with progressive loading in `memory/context.go`
- [ ] `List`/`Load`/`Save`/`Delete` round-trips data correctly
- [ ] Handles empty/missing root directory on first List (empty index, no error)
- [ ] Progressive loading: Bootstrap loads index + prefixes, Resolve loads on demand
- [ ] Thread-safe for concurrent Get/Set
- [ ] Flush persists only dirty entries
- [ ] Tests co-located, black-box style (`memory_test` package)
- [ ] `go vet ./memory/...` passes
- [ ] `go mod tidy` produces no changes
