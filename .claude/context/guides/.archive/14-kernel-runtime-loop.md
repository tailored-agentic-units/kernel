# 14 - Kernel Runtime Loop

## Problem Context

The kernel runtime loop is the central component of Objective #1 (Kernel Core Loop). It composes agent, tools, session, and memory into the observe/think/act/repeat agentic cycle. All three foundation dependencies are complete (#11 session, #12 tools, #13 memory), but two gaps block the kernel implementation:

1. The Agent interface methods accept `prompt string` but the kernel loop needs `[]protocol.Message` for multi-turn tool-use conversations
2. The kernel must initialize from configuration following the cold start pattern — config drives subsystem creation

## Architecture Approach

**Agent evolution:** Conversation-based protocol methods change from `prompt string` to `messages []protocol.Message`. The agent's `initMessages` method retains its name but changes from creating messages internally to prepending the agent's system prompt to caller-provided messages. Non-conversation methods (`Embed`, `Audio`) are unchanged.

**Subsystem configs:** Each subsystem owns a `Config` type and a config-driven constructor following the Default/Merge pattern from `core/config`. The kernel Config embeds subsystem configs and delegates initialization.

**Cold start:** `kernel.New(*Config, ...Option)` creates all subsystems from configuration. Functional options allow tests to override config-created defaults without compromising the config-first design.

**ToolExecutor interface:** The kernel defines `ToolExecutor` with `List()` + `Execute()` since the tools package only exposes global functions. A private `globalToolExecutor` wraps them as the default.

## Implementation

### Step 1: Evolve Agent Interface

**`agent/agent.go`** — Change interface signatures and implementation.

Change the `prompt` parameter type from `string` to `[]protocol.Message` in all 5 conversation method signatures (interface + implementation). The parameter name `prompt` is retained — only the type changes. Method bodies are unchanged since `initMessages` already receives `prompt`.

Update `initMessages` to accept `[]protocol.Message` instead of `string`:

```go
func (a *agent) initMessages(prompt []protocol.Message) []protocol.Message {
	if a.systemPrompt == "" {
		return prompt
	}
	result := make([]protocol.Message, 0, len(prompt)+1)
	result = append(result, protocol.NewMessage(protocol.RoleSystem, a.systemPrompt))
	result = append(result, prompt...)
	return result
}
```

### Step 2: Update Mock Agent

**`agent/mock/agent.go`** — Same change: `prompt string` → `prompt []protocol.Message` in the 5 method signatures. Bodies are unchanged (mock doesn't use the arguments).

### Step 3: Update prompt-agent CLI

**`cmd/prompt-agent/main.go`** — Each execute function builds a message slice from the prompt string. The pattern is the same for all — wrap the prompt in a single user message.

```go
func executeChat(ctx context.Context, agent agent.Agent, prompt string) {
	messages := []protocol.Message{protocol.NewMessage(protocol.RoleUser, prompt)}
	response, err := agent.Chat(ctx, messages)
	// rest unchanged
}

func executeChatStream(ctx context.Context, agent agent.Agent, prompt string) {
	messages := []protocol.Message{protocol.NewMessage(protocol.RoleUser, prompt)}
	stream, err := agent.ChatStream(ctx, messages)
	// rest unchanged
}

func executeVision(ctx context.Context, agent agent.Agent, prompt string, images []string) {
	messages := []protocol.Message{protocol.NewMessage(protocol.RoleUser, prompt)}
	response, err := agent.Vision(ctx, messages, images)
	// rest unchanged
}

func executeVisionStream(ctx context.Context, agent agent.Agent, prompt string, images []string) {
	messages := []protocol.Message{protocol.NewMessage(protocol.RoleUser, prompt)}
	stream, err := agent.VisionStream(ctx, messages, images)
	// rest unchanged
}

func executeTools(ctx context.Context, agent agent.Agent, prompt string, tools []protocol.Tool) {
	messages := []protocol.Message{protocol.NewMessage(protocol.RoleUser, prompt)}
	response, err := agent.Tools(ctx, messages, tools)
	// rest unchanged
}
```

### Step 4: Update Orchestrate Examples

Every `.Chat(ctx, prompt)` call becomes `.Chat(ctx, messages)` where `messages` wraps the prompt in a user message. The pattern for each call site:

```go
// Before
response, err := someAgent.Chat(ctx, prompt)

// After
messages := []protocol.Message{protocol.NewMessage(protocol.RoleUser, prompt)}
response, err := someAgent.Chat(ctx, messages)
```

Files and call sites:

- `orchestrate/examples/phase-01-hubs/main.go` — 4 Chat calls (lines 154, 168, 182, 196)
- `orchestrate/examples/phase-02-03-state-graphs/main.go` — 6 Chat calls (lines 92, 109, 131, 154, 172, 189)
- `orchestrate/examples/phase-04-sequential-chains/main.go` — 1 Chat call (line 164)
- `orchestrate/examples/phase-05-parallel-execution/main.go` — 1 Chat call (line 124)
- `orchestrate/examples/phase-06-checkpointing/main.go` — 4 Chat calls (lines 89, 108, 136, 155)
- `orchestrate/examples/phase-07-conditional-routing/main.go` — 2 Chat calls (lines 218, 272)
- `orchestrate/examples/darpa-procurement/workflow.go` — 8 Chat calls (lines 127, 182, 241, 320, 331, 481, 540, 598)

Each file needs the `protocol` import added if not already present.

### Step 5: Session Configuration

**`session/config.go`** — new file.

```go
package session

type Config struct{}

func DefaultConfig() Config {
	return Config{}
}

func (c *Config) Merge(source *Config) {}

func New(cfg *Config) (Session, error) {
	return NewMemorySession(), nil
}
```

### Step 6: Memory Configuration

**`memory/config.go`** — new file.

```go
package memory

type Config struct {
	Path string `json:"path,omitempty"`
}

func DefaultConfig() Config {
	return Config{}
}

func (c *Config) Merge(source *Config) {
	if source.Path != "" {
		c.Path = source.Path
	}
}

func NewStore(cfg *Config) (Store, error) {
	if cfg.Path == "" {
		return nil, nil
	}
	return NewFileStore(cfg.Path), nil
}
```

### Step 7: Kernel Errors

**`kernel/errors.go`** — new file.

```go
package kernel

import "errors"

var ErrMaxIterations = errors.New("max iterations reached")
```

### Step 8: Kernel Implementation

**`kernel/kernel.go`** — new file.

```go
package kernel

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/tailored-agentic-units/kernel/agent"
	"github.com/tailored-agentic-units/kernel/core/config"
	"github.com/tailored-agentic-units/kernel/core/protocol"
	"github.com/tailored-agentic-units/kernel/core/response"
	"github.com/tailored-agentic-units/kernel/memory"
	"github.com/tailored-agentic-units/kernel/session"
	"github.com/tailored-agentic-units/kernel/tools"
)

const defaultMaxIterations = 10

// --- Configuration ---

type Config struct {
	Agent         config.AgentConfig `json:"agent"`
	Session       session.Config     `json:"session"`
	Memory        memory.Config      `json:"memory"`
	MaxIterations int                `json:"max_iterations,omitempty"`
	SystemPrompt  string             `json:"system_prompt,omitempty"`
}

func DefaultConfig() Config {
	return Config{
		Agent:         config.DefaultAgentConfig(),
		Session:       session.DefaultConfig(),
		Memory:        memory.DefaultConfig(),
		MaxIterations: defaultMaxIterations,
	}
}

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
}

func LoadConfig(filename string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var loaded Config
	if err := json.Unmarshal(data, &loaded); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	cfg.Merge(&loaded)
	return &cfg, nil
}

// --- Types ---

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

// --- ToolExecutor ---

type ToolExecutor interface {
	List() []protocol.Tool
	Execute(ctx context.Context, name string, args json.RawMessage) (tools.Result, error)
}

type globalToolExecutor struct{}

func (globalToolExecutor) List() []protocol.Tool {
	return tools.List()
}

func (globalToolExecutor) Execute(ctx context.Context, name string, args json.RawMessage) (tools.Result, error) {
	return tools.Execute(ctx, name, args)
}

// --- Options ---

type Option func(*Kernel)

func WithAgent(a agent.Agent) Option {
	return func(k *Kernel) { k.agent = a }
}

func WithSession(s session.Session) Option {
	return func(k *Kernel) { k.session = s }
}

func WithToolExecutor(e ToolExecutor) Option {
	return func(k *Kernel) { k.tools = e }
}

func WithMemoryStore(s memory.Store) Option {
	return func(k *Kernel) { k.store = s }
}

// --- Kernel ---

type Kernel struct {
	agent         agent.Agent
	tools         ToolExecutor
	session       session.Session
	store         memory.Store
	maxIterations int
	systemPrompt  string
}

func New(cfg *Config, opts ...Option) (*Kernel, error) {
	a, err := agent.New(&cfg.Agent)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	sess, err := session.New(&cfg.Session)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	store, err := memory.NewStore(&cfg.Memory)
	if err != nil {
		return nil, fmt.Errorf("failed to create memory store: %w", err)
	}

	k := &Kernel{
		agent:         a,
		tools:         globalToolExecutor{},
		session:       sess,
		store:         store,
		maxIterations: cfg.MaxIterations,
		systemPrompt:  cfg.SystemPrompt,
	}

	for _, opt := range opts {
		opt(k)
	}

	return k, nil
}

func (k *Kernel) Run(ctx context.Context, prompt string) (*Result, error) {
	k.session.AddMessage(protocol.NewMessage(protocol.RoleUser, prompt))

	systemContent := k.buildSystemContent(ctx)

	result := &Result{}

	for iteration := range k.maxIterations {
		if err := ctx.Err(); err != nil {
			return result, err
		}

		messages := k.buildMessages(systemContent)

		resp, err := k.agent.Tools(ctx, messages, k.tools.List())
		if err != nil {
			return result, fmt.Errorf("agent call failed: %w", err)
		}

		if len(resp.Choices) == 0 {
			return result, fmt.Errorf("agent returned empty response")
		}

		choice := resp.Choices[0]

		if len(choice.Message.ToolCalls) == 0 {
			k.session.AddMessage(protocol.Message{
				Role:    protocol.RoleAssistant,
				Content: choice.Message.Content,
			})
			result.Response = choice.Message.Content
			result.Iterations = iteration + 1
			return result, nil
		}

		// Assistant message with tool calls
		k.session.AddMessage(protocol.Message{
			Role:      protocol.RoleAssistant,
			Content:   choice.Message.Content,
			ToolCalls: convertToolCalls(choice.Message.ToolCalls),
		})

		// Execute each tool call
		for _, tc := range choice.Message.ToolCalls {
			record := ToolCallRecord{
				Iteration: iteration + 1,
				ID:        tc.ID,
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			}

			toolResult, toolErr := k.tools.Execute(ctx, tc.Function.Name, json.RawMessage(tc.Function.Arguments))

			if toolErr != nil {
				errContent := fmt.Sprintf("error: %s", toolErr)
				k.session.AddMessage(protocol.Message{
					Role:       protocol.RoleTool,
					Content:    errContent,
					ToolCallID: tc.ID,
				})
				record.Result = errContent
				record.IsError = true
			} else {
				k.session.AddMessage(protocol.Message{
					Role:       protocol.RoleTool,
					Content:    toolResult.Content,
					ToolCallID: tc.ID,
				})
				record.Result = toolResult.Content
				record.IsError = toolResult.IsError
			}

			result.ToolCalls = append(result.ToolCalls, record)
		}

		result.Iterations = iteration + 1
	}

	return result, ErrMaxIterations
}

// buildSystemContent combines kernel system prompt with memory context.
func (k *Kernel) buildSystemContent(ctx context.Context) string {
	content := k.systemPrompt

	if k.store == nil {
		return content
	}

	keys, err := k.store.List(ctx)
	if err != nil || len(keys) == 0 {
		return content
	}

	entries, err := k.store.Load(ctx, keys...)
	if err != nil {
		return content
	}

	for _, entry := range entries {
		content += "\n\n" + string(entry.Value)
	}

	return content
}

// buildMessages constructs the full message array: [system] + session history.
func (k *Kernel) buildMessages(systemContent string) []protocol.Message {
	sessionMsgs := k.session.Messages()

	if systemContent == "" {
		return sessionMsgs
	}

	messages := make([]protocol.Message, 0, len(sessionMsgs)+1)
	messages = append(messages, protocol.NewMessage(protocol.RoleSystem, systemContent))
	messages = append(messages, sessionMsgs...)
	return messages
}

func convertToolCalls(toolCalls []response.ToolCall) []protocol.ToolCall {
	result := make([]protocol.ToolCall, len(toolCalls))
	for i, tc := range toolCalls {
		result[i] = protocol.ToolCall{
			ID:        tc.ID,
			Name:      tc.Function.Name,
			Arguments: tc.Function.Arguments,
		}
	}
	return result
}
```

### Step 9: Consolidate ToolCall Types

Eliminate `response.ToolCall` and `response.ToolCallFunction` in favor of `protocol.ToolCall` as the single canonical type. Add a custom unmarshaler to handle the nested JSON format that LLM APIs return.

#### 9a. `core/protocol/message.go` — Custom UnmarshalJSON

Add an unmarshaler on `ToolCall` that handles the nested API response format (`{"id", "type", "function": {"name", "arguments"}}`) and flattens it into the canonical `{ID, Name, Arguments}`:

```go
func (tc *ToolCall) UnmarshalJSON(data []byte) error {
	var nested struct {
		ID       string `json:"id"`
		Function struct {
			Name      string `json:"name"`
			Arguments string `json:"arguments"`
		} `json:"function"`
	}
	if err := json.Unmarshal(data, &nested); err != nil {
		return err
	}

	// Nested format: flatten function fields
	if nested.Function.Name != "" {
		tc.ID = nested.ID
		tc.Name = nested.Function.Name
		tc.Arguments = nested.Function.Arguments
		return nil
	}

	// Flat format: decode directly
	type plain ToolCall
	return json.Unmarshal(data, (*plain)(tc))
}
```

#### 9b. `core/response/tools.go` — Remove ToolCall types, use protocol.ToolCall

Delete `ToolCall` and `ToolCallFunction` types. Update `ToolsResponse` to use `protocol.ToolCall`. Add `protocol` import.

The `ToolsResponse.Choices[].Message.ToolCalls` field changes from `[]ToolCall` to `[]protocol.ToolCall`. The custom unmarshaler on `protocol.ToolCall` handles the nested JSON transparently — `ParseTools` needs no changes.

#### 9c. `kernel/kernel.go` — Remove convertToolCalls, simplify Run

Delete the `convertToolCalls` function and the `core/response` import. In `Run()`, use `protocol.ToolCall` fields directly:

- `tc.Function.Name` → `tc.Name`
- `tc.Function.Arguments` → `tc.Arguments`
- `convertToolCalls(choice.Message.ToolCalls)` → `choice.Message.ToolCalls` (used directly)

#### 9d. `cmd/prompt-agent/main.go` — Flatten field access

```go
// Before
fmt.Printf("  - %s(%s)\n", toolCall.Function.Name, toolCall.Function.Arguments)

// After
fmt.Printf("  - %s(%s)\n", toolCall.Name, toolCall.Arguments)
```

#### 9e. `agent/mock/helpers.go` — Use protocol.ToolCall in helpers

`NewToolsAgent` parameter and inline struct types change from `response.ToolCall` to `protocol.ToolCall`. Same for `NewMultiProtocolAgent`. Replace `response` import with `protocol` where it was the only usage.

```go
func NewToolsAgent(id string, toolCalls []protocol.ToolCall) *MockAgent {
```

All inline struct literals for `ToolCalls` fields change from `[]response.ToolCall` to `[]protocol.ToolCall`.

### Step 10: Propagate Memory Errors

`buildSystemContent` currently silences memory failures. If a store is configured and fails, `Run` should not proceed without the context it was supposed to have.

#### 10a. `kernel/kernel.go` — `buildSystemContent` returns an error

```go
func (k *Kernel) buildSystemContent(ctx context.Context) (string, error) {
	content := k.systemPrompt

	if k.store == nil {
		return content, nil
	}

	keys, err := k.store.List(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list memory keys: %w", err)
	}

	if len(keys) == 0 {
		return content, nil
	}

	entries, err := k.store.Load(ctx, keys...)
	if err != nil {
		return "", fmt.Errorf("failed to load memory entries: %w", err)
	}

	for _, entry := range entries {
		content += "\n\n" + string(entry.Value)
	}

	return content, nil
}
```

#### 10b. `kernel/kernel.go` — `Run` propagates the error

Update the `buildSystemContent` call site in `Run`:

```go
systemContent, err := k.buildSystemContent(ctx)
if err != nil {
	return result, err
}
```

## Validation Criteria

- [ ] `Kernel` struct with config-driven constructor in `kernel/kernel.go`
- [ ] `Config`, `DefaultConfig`, `Merge`, `LoadConfig` follow `core/config` patterns
- [ ] Subsystem configs (`session.Config`, `memory.Config`) with config-driven constructors
- [ ] `Run()` implements observe/think/act/repeat cycle
- [ ] Loop terminates when LLM returns no tool calls (final answer)
- [ ] Loop terminates when MaxIterations reached (returns `ErrMaxIterations` + partial result)
- [ ] Context cancellation stops the loop
- [ ] Tool execution errors are reported to LLM, not fatal
- [ ] Agent interface methods accept `[]protocol.Message` instead of `prompt string`
- [ ] Mock agent and all callers updated for new signatures
- [ ] Functional options allow test overrides of config-created subsystems
- [ ] `protocol.ToolCall` is the single canonical ToolCall type — `response.ToolCall` and `response.ToolCallFunction` deleted
- [ ] `protocol.ToolCall.UnmarshalJSON` handles nested API format
- [ ] Memory load failures in `buildSystemContent` propagate as errors from `Run()`
- [ ] `go vet ./...` passes
