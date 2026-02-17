# Plan: Issue #24 — Agent Registry

## Context

The kernel currently supports a single agent created from `Config.Agent` during `New()`. Issue #24 adds an agent registry — named agents with lazy instantiation, capability querying, and config-driven registration. This enables multi-agent scenarios (future #26) where sessions select agents by name or capability.

## Architecture Approach

- **Registry type defined in `agent` package** — manages agents, which is the agent package's domain. The kernel owns an instance, same pattern as `session.Session`.
- **Instance-owned, not global** — unlike the tools global registry, the agent `Registry` is an exported type. Required for test isolation per the issue.
- **Lazy instantiation**: `Register()` stores config, `Get()` calls `agent.New()` on first access.
- **Capability querying** derived from `ModelConfig.Capabilities` map keys — no instantiation required.
- **Config-driven**: new `Agents map[string]config.AgentConfig` field in kernel Config, populated into registry during kernel `New()`.
- **Backward compatible**: existing single-agent `Config.Agent` + `Run()` flow unchanged.

## Implementation

### Step 1: Add sentinel errors — `agent/errors.go`

Add registry sentinel errors alongside the existing structured `AgentError` type:

```go
var (
    ErrAgentNotFound  = errors.New("agent not found")
    ErrAgentExists    = errors.New("agent already registered")
    ErrEmptyAgentName = errors.New("agent name is empty")
)
```

### Step 2: Registry type — `agent/registry.go` (new file)

```go
type AgentInfo struct {
    Name         string
    Capabilities []protocol.Protocol
}

type Registry struct {
    mu      sync.RWMutex
    configs map[string]config.AgentConfig
    agents  map[string]Agent
}
```

Methods:
- `NewRegistry() *Registry`
- `Register(name string, cfg config.AgentConfig) error` — validates non-empty name, rejects duplicates
- `Replace(name string, cfg config.AgentConfig) error` — updates config for existing name, invalidates cached agent (next `Get()` re-instantiates)
- `Get(name string) (Agent, error)` — read-lock fast path for cached agents, write-lock with double-check for lazy instantiation via `New()`
- `List() []AgentInfo` — sorted by name, capabilities from config (no instantiation)
- `Unregister(name string) error` — removes config + cached agent
- `Capabilities(name string) ([]protocol.Protocol, error)` — capabilities from config
- `capabilitiesFromConfig(cfg *config.AgentConfig) []protocol.Protocol` — unexported helper, filters via `protocol.IsValid()`, sorted

Key behaviors:
- `Get()` uses double-checked locking: read-lock fast path, write-lock with re-check for lazy instantiation
- Failed `Get()` calls do NOT cache the failure — next call retries instantiation
- `Replace()` invalidates the cached agent so the new config takes effect on next `Get()`
- `List()` and `Capabilities()` never trigger instantiation
- All outputs sorted for deterministic behavior

### Step 3: Extend kernel config — `kernel/config.go`

Add `Agents` field to `Config`:

```go
Agents map[string]config.AgentConfig `json:"agents,omitempty"`
```

Update `Merge()` to merge per-name agent configs when source has entries.

### Step 4: Wire registry into kernel — `kernel/kernel.go`

- Add `registry *agent.Registry` field to `Kernel` struct
- In `New()`: create registry, iterate `cfg.Agents` and register each
- Add `Registry() *agent.Registry` accessor method
- Add `WithRegistry(r *agent.Registry) Option` for test overrides

The existing `agent` field and `Run()` loop are unchanged.

## Files

| File | Action |
|------|--------|
| `agent/errors.go` | Add 3 sentinel errors |
| `agent/registry.go` | New — Registry type + AgentInfo |
| `kernel/config.go` | Add Agents field, update Merge |
| `kernel/kernel.go` | Add registry field, wire in New, accessor, option |

## Validation Criteria

- [ ] Registry Register/Get/List/Unregister/Capabilities all work correctly
- [ ] Lazy instantiation: agent created on first Get, cached on subsequent calls
- [ ] Thread-safe: concurrent access with no races
- [ ] Config with `agents` map populates registry during kernel New()
- [ ] Existing single-agent configs work unchanged (backward compatible)
- [ ] `go vet ./...` passes
- [ ] `go test ./...` passes
- [ ] `go mod tidy` produces no changes
