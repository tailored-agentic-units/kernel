# 11 - Session Interface and In-Memory Implementation

## Problem Context

The session subsystem manages conversation turns so the kernel loop can maintain context across observe/think/act cycles. Each iteration sends the full conversation history to the model — including user messages, assistant responses, tool call requests, and correlated tool results.

`protocol.Message` currently has only `Role` and `Content`, which cannot represent tool-calling messages (Known Gap #3). Rather than defining a parallel message type in session, we evolve `protocol.Message` at the core level so session and all other subsystems consume the canonical type natively.

## Architecture Approach

**Protocol-first**: Core data structures are the protocol standard. Subsystems use them natively. If a subsystem needs a field that doesn't exist on a core type, the core type is evolved — not worked around with package-local alternatives.

This means:
- `protocol.Message` gains tool call fields and role constants (resolving Known Gap #3)
- `session.Session` interface operates on `protocol.Message` directly
- No mapping layer needed between session and protocol types

**Why not reuse `response.ToolCall`**: That type is a JSON deserialization struct (nested `Function` struct, `Type` field always `"function"`, json tags). `protocol.ToolCall` is the flat canonical form: `{ID, Name, Arguments}`. The kernel loop maps from response to protocol when processing LLM output.

## Implementation

### Step 1: Evolve `core/protocol/message.go`

Add typed `Role`, role constants, `ToolCall` struct, and new fields to `Message`:

```go
package protocol

type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// Flat canonical form for tool invocations in conversation history.
// Distinct from response.ToolCall which is a JSON deserialization struct.
type ToolCall struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type Message struct {
	Role       Role       `json:"role"`
	Content    any        `json:"content"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
}

func NewMessage(role Role, content any) Message {
	return Message{Role: role, Content: content}
}
```

`omitempty` on `ToolCallID` and `ToolCalls` ensures existing JSON serialization is unchanged when these fields are empty.

### Step 2: Migrate existing callers to typed Role

All `NewMessage("string", ...)` calls become `NewMessage(protocol.Role*, ...)`.

**`agent/agent.go`** — `initMessages()` (lines 319, 322):

```go
func (a *agent) initMessages(prompt string) []protocol.Message {
	messages := make([]protocol.Message, 0)

	if a.systemPrompt != "" {
		messages = append(messages, protocol.NewMessage(protocol.RoleSystem, a.systemPrompt))
	}

	messages = append(messages, protocol.NewMessage(protocol.RoleUser, prompt))

	return messages
}
```

**`agent/mock/helpers.go`** — `NewSimpleChatAgent` (line 24) and `NewMultiProtocolAgent` (line 150):

```go
// line 24
Message: protocol.NewMessage(protocol.RoleAssistant, content),
```

```go
// line 150
Message: protocol.NewMessage(protocol.RoleAssistant, "Mock chat response"),
```

**`agent/providers/base.go`** — no change needed. Line 124 constructs `protocol.Message{Role: message.Role, ...}` where `message.Role` is already typed from the source message.

### Step 3: Create `session/session.go`

```go
package session

import (
	"github.com/tailored-agentic-units/kernel/core/protocol"
)

type Session interface {
	ID() string
	AddMessage(msg protocol.Message)
	Messages() []protocol.Message
	Clear()
}
```

### Step 4: Create `session/memory.go`

```go
package session

import (
	"slices"
	"sync"

	"github.com/google/uuid"
	"github.com/tailored-agentic-units/kernel/core/protocol"
)

type memorySession struct {
	id       string
	messages []protocol.Message
	mu       sync.RWMutex
}

func New() Session {
	return &memorySession{
		id: uuid.Must(uuid.NewV7()).String(),
	}
}

func (s *memorySession) ID() string {
	return s.id
}

func (s *memorySession) AddMessage(msg protocol.Message) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages = append(s.messages, msg)
}

func (s *memorySession) Messages() []protocol.Message {
	s.mu.RLock()
	defer s.mu.RUnlock()

	copied := make([]protocol.Message, len(s.messages))
	for i, msg := range s.messages {
		copied[i] = msg
		copied[i].ToolCalls = slices.Clone(msg.ToolCalls)
	}
	return copied
}

func (s *memorySession) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages = nil
}
```

**Defensive copy in `Messages()`**: Each message struct is copied by value (which covers `Role`, `Content`, `ToolCallID`), then `ToolCalls` is cloned separately since slices are reference types. Prevents callers from mutating session state.

## Validation Criteria

- [ ] `go vet ./...` passes
- [ ] `go test ./...` passes (no regressions in existing tests)
- [ ] `go test -race ./session/...` passes
- [ ] `go mod tidy` produces no changes
- [ ] `protocol.Message` JSON serialization unchanged for messages without tool fields
- [ ] Session defensive copy prevents external mutation of stored messages
- [ ] Concurrent access to session does not race
