# Issue #13 — Memory Store Interface and Filesystem Implementation

## Context

The memory subsystem is the single context composition pipeline for the TAU kernel. It provides a hierarchical key-value namespace backed by pluggable storage, with session-scoped caching and progressive loading. The kernel loads context at session initialization and accesses it locally — no constant I/O to external systems.

This is a Phase 1 foundation subsystem with zero internal kernel dependencies. The memory package currently contains only a README.md skeleton.

**Expanded scope**: Beyond the minimal Store from the original issue, this implementation establishes the memory system as the unified context pipeline for notes, skills, and profiles — with progressive loading support for the three-level skill access pattern.

## Architecture

### Data Model

- **Keys**: `/`-separated hierarchical paths (like relative filesystem paths)
- **Values**: `[]byte` (raw content — the kernel's contract)
- **Namespaces**: `memory/`, `skills/`, `agents/` (top-level conventions)

Example key layout:
```
memory/global.md                    # Always-loaded agent memory
memory/project/kernel.md            # Project-scoped memory
skills/go-patterns/SKILL.md         # Skill definition (metadata + instructions)
skills/go-patterns/resources/...    # Skill resources (on-demand)
agents/explorer.json                # Sub-agent profile
```

### Interfaces

**Store** — persistence abstraction (stateless, handles I/O):
```go
type Store interface {
    List(ctx context.Context) ([]string, error)              // Discover available keys
    Load(ctx context.Context, keys ...string) ([]Entry, error)  // Fetch specific entries
    Save(ctx context.Context, entries ...Entry) error         // Persist entries
    Delete(ctx context.Context, keys ...string) error         // Remove entries
}
```

**Context** — session-scoped cache (stateful, no I/O on reads):
```go
func NewContext(store Store) *Context
func (c *Context) Bootstrap(ctx context.Context, prefixes ...string) error  // Index + load matching prefixes
func (c *Context) Resolve(ctx context.Context, keys ...string) error        // Load specific entries on demand
func (c *Context) Flush(ctx context.Context) error                          // Persist dirty entries

func (c *Context) Get(key string) ([]byte, bool)  // Read cached content
func (c *Context) Set(key string, value []byte)    // Update cache (marks dirty)
func (c *Context) Delete(key string)               // Remove from cache (marks for deletion)
func (c *Context) Has(key string) bool              // Key exists in index (loaded or not)
func (c *Context) Keys() []string                   // All indexed keys
func (c *Context) Entries(prefix string) []Entry    // Cached entries matching prefix
```

### Progressive Loading Flow

```
Bootstrap("memory/")
  → Store.List() → populate index (all available keys)
  → Store.Load("memory/*") → cache matching entries

During session:
  → Has("skills/go-patterns/SKILL.md") → true (in index)
  → Get("skills/go-patterns/SKILL.md") → nil, false (not cached)
  → Resolve(ctx, "skills/go-patterns/SKILL.md") → loads into cache
  → Get("skills/go-patterns/SKILL.md") → content, true

On flush:
  → Flush(ctx) → Save dirty entries, Delete removed entries
```

### FileStore Translation

FileStore maps filesystem structure 1:1 to the key namespace:
- `{root}/memory/global.md` → key `"memory/global.md"`
- `{root}/skills/go-patterns/SKILL.md` → key `"skills/go-patterns/SKILL.md"`
- `{root}/agents/explorer.json` → key `"agents/explorer.json"`

No extension stripping, no path transformation. Keys ARE relative filesystem paths. This is the simplest storage-to-namespace translation — other Store implementations (database, ConnectRPC) would define their own mapping.

## File Plan

| File | Action | Purpose |
|------|--------|---------|
| `memory/entry.go` | Create | Entry type, namespace constants |
| `memory/store.go` | Create | Store interface |
| `memory/filestore.go` | Create | Filesystem Store implementation |
| `memory/context.go` | Create | Session-scoped context cache |
| `memory/errors.go` | Create | Sentinel errors |

## Implementation

### Step 1: Sentinel errors (`memory/errors.go`)

Error variables for storage operations:
- `ErrKeyNotFound` — requested key not in store
- `ErrLoadFailed` — wrapping context for storage read failures
- `ErrSaveFailed` — wrapping context for storage write failures

### Step 2: Entry type and namespace constants (`memory/entry.go`)

```go
type Entry struct {
    Key   string
    Value []byte
}
```

Namespace constants:
```go
const (
    NamespaceMemory = "memory"
    NamespaceSkills = "skills"
    NamespaceAgents = "agents"
)
```

### Step 3: Store interface (`memory/store.go`)

The `Store` interface as defined in the Architecture section above. Stateless persistence contract.

### Step 4: FileStore implementation (`memory/filestore.go`)

Constructor: `NewFileStore(root string) Store`

Behaviors:
- **List**: `filepath.WalkDir` the root, return relative paths for all regular files. Skip hidden files/dirs (`.` prefix).
- **Load**: For each requested key, read `filepath.Join(root, key)`. Return `ErrKeyNotFound` wrapped if file missing.
- **Save**: For each entry, write to `filepath.Join(root, key)`. Create parent directories with `os.MkdirAll`. Use atomic write (temp file + rename).
- **Delete**: For each key, `os.Remove(filepath.Join(root, key))`. Ignore not-found errors. Clean up empty parent directories.

### Step 5: Context implementation (`memory/context.go`)

Internal state:
```go
type Context struct {
    store   Store
    cache   map[string][]byte  // loaded content
    index   map[string]bool    // all known keys
    dirty   map[string]bool    // modified since last flush
    removed map[string]bool    // deleted since last flush
    mu      sync.RWMutex
}
```

Key behaviors:
- **Bootstrap(ctx, prefixes...)**: Calls `Store.List` to populate index. If prefixes given, calls `Store.Load` for matching keys and populates cache.
- **Resolve(ctx, keys...)**: Skips already-cached keys. Loads remaining from Store. Adds to cache.
- **Get(key)**: Read-locked cache lookup. Returns `nil, false` if not loaded.
- **Set(key, value)**: Write-locked. Updates cache, adds to index, marks dirty.
- **Delete(key)**: Write-locked. Removes from cache, marks for deletion.
- **Has(key)**: Read-locked index lookup.
- **Keys()**: Read-locked. Returns sorted slice of all indexed keys.
- **Entries(prefix)**: Read-locked. Returns cached entries whose keys start with prefix.
- **Flush(ctx)**: Calls `Store.Save` for dirty entries, `Store.Delete` for removed entries. Clears dirty/removed tracking.

Thread-safety: `sync.RWMutex` — reads use `RLock`, writes use `Lock`.

## Validation Criteria

- [ ] `Store` interface defined in `memory/store.go`
- [ ] `FileStore` implementation in `memory/filestore.go`
- [ ] `Context` with progressive loading in `memory/context.go`
- [ ] `List`/`Load`/`Save`/`Delete` round-trips data correctly
- [ ] Handles empty/missing root directory on first List (empty index, no error)
- [ ] Progressive loading: Bootstrap loads index + prefixes, Resolve loads on demand
- [ ] Thread-safe for concurrent Get/Set
- [ ] Flush persists only dirty entries
- [ ] Tests co-located, black-box style (`memory_test` package)
- [ ] `go vet ./memory/...` passes
- [ ] `go mod tidy` produces no changes
