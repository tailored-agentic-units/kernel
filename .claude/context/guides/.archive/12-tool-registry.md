# 12 - Tool Registry Interface and Execution

## Problem Context

The tools subsystem needs a registry for tool handlers that maps LLM tool calls to their execution. Currently `agent.Tool` and `providers.ToolDefinition` are identical types in two packages. Adding a third in `tools` would create three-way duplication, violating the principle that exported data structures are the protocol — defined once, consumed directly at higher levels.

The registry follows the global singleton catalog pattern established by `agent/providers/registry.go`. It catalogs all available tools — the kernel selects which subset to offer the LLM per turn. Built-in tools (future) self-register via `init()` in sub-packages; external libraries extend the catalog by calling `tools.Register()`.

## Architecture Approach

1. Define `protocol.Tool` in `core/protocol` as the canonical tool definition type
2. Replace `agent.Tool` and `providers.ToolDefinition` with `protocol.Tool` throughout
3. Build the global tool registry in `tools/` using `protocol.Tool`, following the providers registry pattern

## Implementation

### Step 1: Define `protocol.Tool` in core

**New file: `core/protocol/tool.go`**

```go
package protocol

// Tool defines a function that can be called by the LLM.
type Tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}
```

### Step 2: Update `agent/providers/data.go`

Remove the `ToolDefinition` struct and update `ToolsData` to use `protocol.Tool`:

```go
// ToolsData contains the data needed to marshal a tools request.
type ToolsData struct {
	Model    string
	Messages []protocol.Message
	Tools    []protocol.Tool
	Options  map[string]any
}
```

Remove lines 29-36 (`ToolDefinition` struct).

### Step 3: Update `agent/request/tools.go`

Replace `[]providers.ToolDefinition` with `[]protocol.Tool`. The `providers` import is still needed for `providers.Provider` and `providers.ToolsData`.

```go
type ToolsRequest struct {
	messages []protocol.Message
	tools    []protocol.Tool
	options  map[string]any
	provider providers.Provider
	model    *model.Model
}

func NewTools(p providers.Provider, m *model.Model, messages []protocol.Message, tools []protocol.Tool, opts map[string]any) *ToolsRequest {
	return &ToolsRequest{
		messages: messages,
		tools:    tools,
		options:  opts,
		provider: p,
		model:    m,
	}
}
```

### Step 4: Update `agent/agent.go`

Update the `Agent` interface method signature (line 61):

```go
Tools(ctx context.Context, prompt string, tools []protocol.Tool, opts ...map[string]any) (*response.ToolsResponse, error)
```

Update the implementation method (lines 216-247). Remove the conversion loop — pass `[]protocol.Tool` directly to the request:

```go
func (a *agent) Tools(ctx context.Context, prompt string, tools []protocol.Tool, opts ...map[string]any) (*response.ToolsResponse, error) {
	messages := a.initMessages(prompt)
	options := a.mergeOptions(protocol.Tools, opts...)

	req := request.NewTools(a.provider, a.model, messages, tools, options)

	result, err := a.client.Execute(ctx, req)
	if err != nil {
		return nil, err
	}

	resp, ok := result.(*response.ToolsResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", result)
	}

	return resp, nil
}
```

Remove the `Tool` struct at lines 327-340.

### Step 5: Update `agent/mock/agent.go`

Update the `Tools` method signature (line 202):

```go
func (m *MockAgent) Tools(ctx context.Context, prompt string, tools []protocol.Tool, opts ...map[string]any) (*response.ToolsResponse, error) {
	return m.toolsResponse, m.toolsError
}
```

The `agent` import can be removed since `MockAgent` no longer references `agent.Tool`. The interface assertion `var _ agent.Agent = (*MockAgent)(nil)` still needs the import.

### Step 6: Update `cmd/prompt-agent/main.go`

Update imports — replace `agent` with `protocol`:

```go
import (
	// ...existing imports...
	"github.com/tailored-agentic-units/kernel/agent"
	"github.com/tailored-agentic-units/kernel/core/config"
	"github.com/tailored-agentic-units/kernel/core/protocol"
)
```

Update `executeTools` signature (line 159):

```go
func executeTools(ctx context.Context, agent agent.Agent, prompt string, tools []protocol.Tool) {
```

Update `loadTools` function (lines 266-278):

```go
func loadTools(filename string) []protocol.Tool {
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("Failed to read tools file: %v", err)
	}

	var tools []protocol.Tool
	if err := json.Unmarshal(data, &tools); err != nil {
		log.Fatalf("Failed to parse tools file: %v", err)
	}

	return tools
}
```

### Step 7: Update `agent/doc.go`

Update the Tools Protocol section. Change `agent.Tool` references to `protocol.Tool`:

Line 21: `Tools(ctx context.Context, prompt string, tools []protocol.Tool, opts ...map[string]any) (*types.ToolsResponse, error)`

Line 123: `tools := []protocol.Tool{`

Lines 213-221 (Tool Definitions section): Update the struct reference to show `protocol.Tool`.

### Step 8: Update `_project/README.md` Known Gaps

Replace lines 164-170:

```markdown
## Known Gaps

One limitation identified in the current codebase, deferred to per-subsystem concept sessions:

1. **Agent methods create fresh messages** — The `Agent` interface methods accept a `prompt string` and internally create a fresh message list. Incompatible with multi-turn conversations where full history must be passed. Deferred to agent subsystem redesign.
```

### Step 9: Create `tools/errors.go`

**New file: `tools/errors.go`**

```go
package tools

import "errors"

var (
	ErrNotFound      = errors.New("tool not found")
	ErrAlreadyExists = errors.New("tool already registered")
	ErrEmptyName     = errors.New("tool name is empty")
)
```

### Step 10: Create `tools/registry.go`

**New file: `tools/registry.go`**

Follow the `agent/providers/registry.go` pattern exactly:

```go
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/tailored-agentic-units/kernel/core/protocol"
)

type Handler func(ctx context.Context, args json.RawMessage) (Result, error)

type Result struct {
	Content string
	IsError bool
}

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

func Register(tool protocol.Tool, handler Handler) error {
	if tool.Name == "" {
		return ErrEmptyName
	}

	register.mu.Lock()
	defer register.mu.Unlock()

	if _, exists := register.entries[tool.Name]; exists {
		return fmt.Errorf("%w: %s", ErrAlreadyExists, tool.Name)
	}

	register.entries[tool.Name] = entry{tool: tool, handler: handler}
	return nil
}

func Replace(tool protocol.Tool, handler Handler) error {
	if tool.Name == "" {
		return ErrEmptyName
	}

	register.mu.Lock()
	defer register.mu.Unlock()

	if _, exists := register.entries[tool.Name]; !exists {
		return fmt.Errorf("%w: %s", ErrNotFound, tool.Name)
	}

	register.entries[tool.Name] = entry{tool: tool, handler: handler}
	return nil
}

func Get(name string) (Handler, bool) {
	register.mu.RLock()
	defer register.mu.RUnlock()

	e, exists := register.entries[name]
	if !exists {
		return nil, false
	}
	return e.handler, true
}

func List() []protocol.Tool {
	register.mu.RLock()
	defer register.mu.RUnlock()

	tools := make([]protocol.Tool, 0, len(register.entries))
	for _, e := range register.entries {
		tools = append(tools, e.tool)
	}
	return tools
}

func Execute(ctx context.Context, name string, args json.RawMessage) (Result, error) {
	register.mu.RLock()
	e, exists := register.entries[name]
	register.mu.RUnlock()

	if !exists {
		return Result{}, fmt.Errorf("%w: %s", ErrNotFound, name)
	}

	result, err := e.handler(ctx, args)
	if err != nil {
		return Result{}, fmt.Errorf("tool %s execution failed: %w", name, err)
	}

	return result, nil
}
```

### Step 11: Update `tools/README.md`

Replace the existing skeleton content with updated status and usage.

## Validation Criteria

- [ ] `go vet ./...` passes
- [ ] `go test ./...` passes (all existing tests updated for `protocol.Tool`)
- [ ] `go mod tidy` produces no changes
- [ ] `protocol.Tool` is the single canonical tool definition type
- [ ] `agent.Tool` and `providers.ToolDefinition` no longer exist
- [ ] `tools.Register`, `tools.Get`, `tools.List`, `tools.Execute` work correctly
- [ ] Registry is thread-safe for concurrent access
- [ ] Known Gaps updated in `_project/README.md`
