# Plan: Issue #14 — Kernel Runtime Loop

## Context

The kernel runtime loop composes agent, tools, session, and memory into the observe/think/act/repeat agentic cycle. All three foundation dependencies are complete (#11 session, #12 tools, #13 memory).

Two gaps must be resolved: (1) the `Agent` interface methods accept `prompt string` but the kernel loop needs `[]protocol.Message` for multi-turn tool-use conversations, and (2) the kernel must initialize from configuration following the cold start pattern established in agent-lab — config drives subsystem creation, not the caller.

## Approach

Two parts: (1) evolve the Agent interface, (2) implement the config-driven kernel runtime.

### Part 1: Agent Interface Evolution

**What changes:** Conversation-based protocol methods (`Chat`, `ChatStream`, `Vision`, `VisionStream`, `Tools`) change their first content parameter from `prompt string` to `messages []protocol.Message`. Non-conversation methods (`Embed`, `Audio`) remain unchanged.

**System prompt behavior:** `initMessages()` is renamed to `prependSystemPrompt()`. It prepends the agent's configured system prompt (if any) to the caller-provided messages. Callers gain full message control while agents retain their identity prompt.

**Files modified:**

| File | Change |
|------|--------|
| `agent/agent.go` | Interface signatures + implementation; rename `initMessages` → `prependSystemPrompt` |
| `agent/mock/agent.go` | Mock method signatures |
| `agent/mock/helpers.go` | No changes needed (helpers configure responses, not call sites) |
| `cmd/prompt-agent/main.go` | Build message slices from CLI prompt string |
| `orchestrate/examples/**/*.go` | Update `.Chat(ctx, prompt)` → `.Chat(ctx, messages)` call sites |

### Part 2: Subsystem Configurations

Each subsystem owns its configuration and provides a config-driven constructor. This establishes extension points and follows the agent-lab cold start pattern — systems initialize from their own configs.

**New files:**

| File | Contents |
|------|----------|
| `session/config.go` | `Config` type, `DefaultConfig()`, `Config.Merge()` |
| `memory/config.go` | `Config` type, `DefaultConfig()`, `Config.Merge()`, `NewStore()` |

#### `session.Config`

```go
type Config struct {
    // Extension point for future session backends.
    // Currently only in-memory sessions are supported.
}

func DefaultConfig() Config
func (c *Config) Merge(source *Config)
func New(cfg *Config) (Session, error)  // returns NewMemorySession() for now
```

#### `memory.Config`

```go
type Config struct {
    Path string `json:"path,omitempty"`  // FileStore root; empty = no memory
}

func DefaultConfig() Config
func (c *Config) Merge(source *Config)
func NewStore(cfg *Config) (Store, error)  // returns nil Store if Path empty
```

### Part 3: Kernel Runtime Loop

**New files in `kernel/`:**

| File | Contents |
|------|----------|
| `kernel.go` | `Kernel` struct, `Config`, `Result`, `ToolCallRecord`, `ToolExecutor` interface, `Option` type, `New()`, `Run()` |
| `errors.go` | `ErrMaxIterations` |

#### Configuration (Cold Start)

The kernel initializes purely from configuration. `New(*Config, ...Option)` delegates to each subsystem's config-driven constructor — callers never construct dependencies manually.

```go
type Config struct {
    Agent         config.AgentConfig  `json:"agent"`
    Session       session.Config      `json:"session"`
    Memory        memory.Config       `json:"memory"`
    MaxIterations int                 `json:"max_iterations,omitempty"`
    SystemPrompt  string              `json:"system_prompt,omitempty"`
}
```

Config lifecycle (matching `core/config` patterns):
- `DefaultConfig()` — sensible defaults (e.g., `MaxIterations: 10`), calls subsystem `DefaultConfig()` functions
- `Config.Merge(*Config)` — delegates to each subsystem's `Merge`, overwrites non-zero kernel fields
- `LoadConfig(filename string) (*Config, error)` — load JSON, merge with defaults

Cold start in `New`:
1. Create agent from `cfg.Agent` via `agent.New()`
2. Create session from `cfg.Session` via `session.New()`
3. Create memory store from `cfg.Memory` via `memory.NewStore()` (returns nil if path empty)
4. Default tool executor wraps the global `tools` package
5. Store config values (`MaxIterations`, `SystemPrompt`)

#### Functional Options (Test Overrides)

Options allow tests to override config-created subsystems:

```go
type Option func(*Kernel)

WithAgent(a agent.Agent) Option
WithSession(s session.Session) Option
WithToolExecutor(e ToolExecutor) Option
WithMemoryStore(s memory.Store) Option
```

Applied after cold start — overrides replace the config-created defaults.

#### ToolExecutor Interface

The tools package exposes global functions, not an instantiable type. The kernel defines a `ToolExecutor` interface for testability:

```go
type ToolExecutor interface {
    List() []protocol.Tool
    Execute(ctx context.Context, name string, args json.RawMessage) (tools.Result, error)
}
```

A private `globalToolExecutor` struct wraps `tools.List()` and `tools.Execute()` as the default implementation created during cold start.

#### Run() Loop

```
Run(ctx context.Context, prompt string) (*Result, error)

1. Add user prompt as message to session
2. Load memory context (if store != nil) and build system message
3. Loop:
   a. Build messages: [system] + session.Messages()
   b. Call agent.Tools(ctx, messages, executor.List())
   c. If response has tool calls:
      - Append assistant message (with tool calls) to session
      - For each tool call: execute via executor, append tool result to session
      - Record ToolCallRecord entries
      - Increment iteration, check MaxIterations → return ErrMaxIterations + partial Result
      - Check ctx.Err() → return context error
      - Continue loop
   d. If no tool calls (final answer):
      - Append assistant message to session
      - Return Result
```

#### Types

```go
type Result struct {
    Response   string
    Iterations int
    ToolCalls  []ToolCallRecord
}

type ToolCallRecord struct {
    Iteration int
    ID        string
    Name      string
    Arguments string
    Result    string
    IsError   bool
}
```

#### Error Handling

- `ErrMaxIterations` — returned alongside partial `Result`
- Context cancellation — checked at loop top, returns `ctx.Err()`
- Tool execution errors — reported to LLM as tool result message (not fatal)
- Agent errors — returned immediately (unrecoverable)

#### Response Type Mapping

- `response.ToolCall` (nested `Function.Name`/`Function.Arguments`) → `protocol.ToolCall` (flat `Name`/`Arguments`) for session messages
- Tool result messages use `protocol.RoleTool` with `ToolCallID`
