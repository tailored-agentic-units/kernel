# 24 - Agent Registry

## Problem Context

The kernel currently supports a single agent created from `Config.Agent` during `New()`. Callers shouldn't need to feed full agent configurations to the kernel for every operation. A registry provides named agent registration with capability awareness, enabling future multi-session and multi-agent scenarios.

## Architecture Approach

The registry is defined in the `agent` package as an exported, instance-owned type — the same pattern as `session.Session`. The kernel creates and owns a registry instance. Agents are registered by name with their configs; actual `Agent` instances are created lazily on first `Get()` call. Capabilities are derived from `ModelConfig.Capabilities` keys without requiring instantiation.

## Implementation

### Step 1: Add sentinel errors — `agent/errors.go`

Add registry sentinel errors after the existing `NewAgentLLMError` function (before the closing of the file). Add the `errors` import.

```go
import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tailored-agentic-units/kernel/core/config"
)

// ...existing code...

var (
	ErrAgentNotFound  = errors.New("agent not found")
	ErrAgentExists    = errors.New("agent already registered")
	ErrEmptyAgentName = errors.New("agent name is empty")
)
```

### Step 2: Registry type — `agent/registry.go` (new file)

Complete implementation:

```go
package agent

import (
	"fmt"
	"sort"
	"sync"

	"github.com/tailored-agentic-units/kernel/core/config"
	"github.com/tailored-agentic-units/kernel/core/protocol"
)

type AgentInfo struct {
	Name         string
	Capabilities []protocol.Protocol
}

type Registry struct {
	mu      sync.RWMutex
	configs map[string]config.AgentConfig
	agents  map[string]Agent
}

func NewRegistry() *Registry {
	return &Registry{
		configs: make(map[string]config.AgentConfig),
		agents:  make(map[string]Agent),
	}
}

func (r *Registry) Register(name string, cfg config.AgentConfig) error {
	if name == "" {
		return ErrEmptyAgentName
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.configs[name]; exists {
		return fmt.Errorf("%w: %s", ErrAgentExists, name)
	}

	r.configs[name] = cfg
	return nil
}

func (r *Registry) Replace(name string, cfg config.AgentConfig) error {
	if name == "" {
		return ErrEmptyAgentName
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.configs[name]; !exists {
		return fmt.Errorf("%w: %s", ErrAgentNotFound, name)
	}

	r.configs[name] = cfg
	delete(r.agents, name)
	return nil
}

func (r *Registry) Get(name string) (Agent, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, registered := r.configs[name]; !registered {
		return nil, fmt.Errorf("%w: %s", ErrAgentNotFound, name)
	}

	if a, exists := r.agents[name]; exists {
		return a, nil
	}

	cfg := r.configs[name]
	a, err := New(&cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent %q: %w", name, err)
	}

	r.agents[name] = a
	return a, nil
}

func (r *Registry) List() []AgentInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	infos := make([]AgentInfo, 0, len(r.configs))
	for name, cfg := range r.configs {
		infos = append(infos, AgentInfo{
			Name:         name,
			Capabilities: capabilitiesFromConfig(&cfg),
		})
	}

	sort.Slice(infos, func(i, j int) bool {
		return infos[i].Name < infos[j].Name
	})

	return infos
}

func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.configs[name]; !exists {
		return fmt.Errorf("%w: %s", ErrAgentNotFound, name)
	}

	delete(r.configs, name)
	delete(r.agents, name)
	return nil
}

func (r *Registry) Capabilities(name string) ([]protocol.Protocol, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cfg, exists := r.configs[name]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrAgentNotFound, name)
	}

	return capabilitiesFromConfig(&cfg), nil
}

func capabilitiesFromConfig(cfg *config.AgentConfig) []protocol.Protocol {
	if cfg.Model == nil || len(cfg.Model.Capabilities) == 0 {
		return nil
	}

	caps := make([]protocol.Protocol, 0, len(cfg.Model.Capabilities))
	for key := range cfg.Model.Capabilities {
		if protocol.IsValid(key) {
			caps = append(caps, protocol.Protocol(key))
		}
	}

	sort.Slice(caps, func(i, j int) bool {
		return string(caps[i]) < string(caps[j])
	})

	return caps
}
```

### Step 3: Extend kernel config — `kernel/config.go`

Add `Agents` field to the `Config` struct:

```go
type Config struct {
	Agent         config.AgentConfig            `json:"agent"`
	Agents        map[string]config.AgentConfig  `json:"agents,omitempty"`
	Session       session.Config                 `json:"session"`
	Memory        memory.Config                  `json:"memory"`
	MaxIterations int                            `json:"max_iterations,omitempty"`
	SystemPrompt  string                         `json:"system_prompt,omitempty"`
}
```

Add agents merge logic at the end of `Merge()`:

```go
func (c *Config) Merge(source *Config) {
	c.Agent.Merge(&source.Agent)
	c.Session.Merge(&source.Session)
	c.Memory.Merge(&source.Memory)

	if source.MaxIterations > 0 {
		c.MaxIterations = source.MaxIterations
	}
	if source.SystemPrompt != "" {
		c.SystemPrompt = source.SystemPrompt
	}

	if len(source.Agents) > 0 {
		c.Agents = source.Agents
	}
}
```

### Step 4: Wire registry into kernel — `kernel/kernel.go`

Add `registry` field to the `Kernel` struct:

```go
type Kernel struct {
	agent         agent.Agent
	registry      *agent.Registry
	session       session.Session
	store         memory.Store
	tools         ToolExecutor
	log           *slog.Logger
	maxIterations int
	systemPrompt  string
}
```

In `New()`, create and populate the registry after the existing subsystem initialization, before applying options:

```go
reg := agent.NewRegistry()
for name, agentCfg := range cfg.Agents {
	if err := reg.Register(name, agentCfg); err != nil {
		return nil, fmt.Errorf("failed to register agent %q: %w", name, err)
	}
}
```

Include `registry: reg` in the kernel struct literal.

Add accessor and option:

```go
func (k *Kernel) Registry() *agent.Registry {
	return k.registry
}

func WithRegistry(r *agent.Registry) Option {
	return func(k *Kernel) { k.registry = r }
}
```

## Validation Criteria

- [ ] Registry Register/Get/Replace/List/Unregister/Capabilities all work correctly
- [ ] Lazy instantiation: agent created on first Get, cached on subsequent calls
- [ ] Replace invalidates cached agent
- [ ] Thread-safe: concurrent access with no races
- [ ] Config with `agents` map populates registry during kernel New()
- [ ] Existing single-agent configs work unchanged (backward compatible)
- [ ] `go vet ./...` passes
- [ ] `go test ./...` passes
- [ ] `go mod tidy` produces no changes
