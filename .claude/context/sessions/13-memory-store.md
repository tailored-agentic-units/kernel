# 13 - Memory Store Interface and Filesystem Implementation

## Summary

Implemented the memory subsystem as the unified context composition pipeline for the TAU kernel. The package provides a hierarchical key-value namespace with two core types: Store (persistence abstraction) and Cache (session-scoped cache with progressive loading). Includes a filesystem-based Store implementation (FileStore) that maps keys 1:1 to relative file paths. Removed the standalone `skills/` skeleton package since skills are now a namespace within the memory system.

## Key Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Expanded scope | Memory as unified context pipeline | Consolidates memory, skills, and agent profiles into a single namespace with shared persistence and caching semantics |
| Value type | `[]byte` | Kernel's contract — external systems handle transformations |
| Key format | `/`-separated paths | Maps 1:1 to filesystem paths, natural hierarchy separator for any backend |
| Namespace naming | `memory/`, `skills/`, `agents/` | `memory` for core agent memory, `agents` aligns with Claude convention |
| Cache vs Context naming | `Cache` | Avoids conflation with `context.Context` in function signatures |
| Namespace constants | Untyped strings | Used as prefix conventions in string operations, not discrete enum values |
| Progressive loading | Bootstrap indexes, Resolve loads on demand | Supports three-level skill access pattern without eager loading |
| Store ordering | Unspecified | FileStore naturally returns sorted (WalkDir), but interface doesn't guarantee it |

## Files Modified

- `memory/errors.go` — created: sentinel errors
- `memory/entry.go` — created: Entry type, namespace constants
- `memory/store.go` — created: Store interface, package doc
- `memory/filestore.go` — created: filesystem Store implementation
- `memory/cache.go` — created: session-scoped Cache with progressive loading
- `memory/filestore_test.go` — created: 12 FileStore tests
- `memory/cache_test.go` — created: 19 Cache tests
- `memory/README.md` — updated description
- `README.md` — removed `skills/` row, updated `memory/` description
- `.claude/CLAUDE.md` — removed `skills/` from structure, updated `memory/` description
- `skills/` — removed (skeleton consolidated into memory)

## Patterns Established

- **Store interface**: Stateless persistence with `List`, `Load`, `Save`, `Delete` — any backend implements this to participate in the context pipeline
- **Cache progressive loading**: `Bootstrap` (index + prefixes) → `Resolve` (on demand) → `Get` (local read) → `Flush` (persist dirty)
- **Defensive copies**: `Get` and `Set` clone byte slices to prevent external mutation
- **Namespace conventions**: `memory/`, `skills/`, `agents/` as top-level key prefixes

## Validation Results

- `go vet ./...` — clean
- `go test ./...` — all packages pass
- `go mod tidy` — no changes
- Memory package: 31 tests, 88.2% coverage
- Uncovered: defensive OS-level error paths in Save/List (acceptable)
