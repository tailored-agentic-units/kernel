# Issue #12 — Tool Registry Interface and Execution

## Context

The tools subsystem needs a registry for tool handlers that maps LLM tool calls to their execution. Currently `agent.Tool` and `providers.ToolDefinition` are identical types defined in two places — introducing a third type would create three-way duplication.

Per the project principle that *"exported data structures are the protocol — define once, consume directly at higher levels"*, the canonical `Tool` type belongs in `core/protocol` (the lowest affected level). Both `agent` and `tools` then consume it directly, resolving the Known Gap.

The tools registry follows the **global singleton catalog pattern** established by `agent/providers/registry.go` — a thread-safe global registry with package-level functions (`Register`, `Get`, `List`, `Execute`). The registry is a catalog of all available tools, not the active set for a session. The kernel runtime selects which tools to offer the LLM from this catalog.

Built-in tools (future) will self-register via `init()` in sub-packages. External libraries extend the catalog by calling `tools.Register()`. This matches the providers extensibility model exactly.

## Step 1: Establish `protocol.Tool` as Canonical Type

Move the tool definition type to `core/protocol` and update all consumers.

### New file: `core/protocol/tool.go`

```go
type Tool struct {
    Name        string         `json:"name"`
    Description string         `json:"description"`
    Parameters  map[string]any `json:"parameters"`
}
```

### Update `agent` package

- **`agent/agent.go`**: Remove `Tool` struct (lines 327-340). Update `Agent` interface and `agent.Tools()` method to use `protocol.Tool`. Remove the `[]Tool` → `[]providers.ToolDefinition` conversion loop (lines 224-231) — pass `[]protocol.Tool` through directly.
- **`agent/doc.go`**: Update documentation references from `agent.Tool` to `protocol.Tool`.
- **`agent/mock/agent.go`**: Update `MockAgent.Tools()` signature to `[]protocol.Tool`.

### Update `providers` package

- **`agent/providers/data.go`**: Remove `ToolDefinition` struct. Update `ToolsData.Tools` field to `[]protocol.Tool`.
- **`agent/request/tools.go`**: Update `NewTools()` parameter and `toolsRequest.tools` field to `[]protocol.Tool`.

### Update consumers

- **`cmd/prompt-agent/main.go`**: Update `executeTools()` and `loadTools()` to use `[]protocol.Tool`.

### Update tests

- `agent/agent_test.go`: `[]agent.Tool` → `[]protocol.Tool`
- `agent/mock/agent_test.go`: Update import if needed
- `agent/client/client_test.go`: `[]providers.ToolDefinition` → `[]protocol.Tool`
- `agent/providers/base_test.go`: `[]providers.ToolDefinition` → `[]protocol.Tool`

### Update Known Gaps

- **`_project/README.md`**: Remove the "Tool type duplication" Known Gap entry (item 2), since it's resolved.

## Step 2: Tool Registry (Global Singleton)

Build the global tool registry following the `agent/providers/registry.go` pattern. The tools package depends on `core/protocol` only.

### New file: `tools/errors.go`

Package-level sentinel errors:

```go
var (
    ErrNotFound      = errors.New("tool not found")
    ErrAlreadyExists = errors.New("tool already registered")
    ErrEmptyName     = errors.New("tool name is empty")
)
```

### New file: `tools/registry.go`

**Types:**

```go
// Handler is the function signature for tool implementations.
type Handler func(ctx context.Context, args json.RawMessage) (Result, error)

// Result is the tool execution output that feeds back into the next LLM turn.
type Result struct {
    Content string
    IsError bool
}
```

**Global registry** (following providers/registry.go pattern):

```go
type entry struct {
    tool    protocol.Tool
    handler Handler
}

type registry struct {
    entries map[string]entry
    mu      sync.RWMutex
}

var register = &registry{
    entries: make(map[string]entry),
}
```

**Package-level functions:**

- `Register(tool protocol.Tool, handler Handler) error` — validates non-empty name, rejects duplicates, stores entry
- `Get(name string) (Handler, bool)` — RLock, lookup by name
- `List() []protocol.Tool` — RLock, return all registered tool definitions (defensive copy)
- `Execute(ctx context.Context, name string, args json.RawMessage) (Result, error)` — RLock to get handler, release lock, call handler, wrap errors with context

### File updates

- **`tools/README.md`**: Update status and add usage section showing the registration pattern.

## Validation

- `go vet ./...`
- `go test ./...` (all existing + new tests pass)
- `go mod tidy` produces no changes
