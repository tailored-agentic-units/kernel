# Issue #11 — Session Interface and In-Memory Implementation

## Session Resumption

This plan was created in a prior session. To resume in a new conversation:

1. Run `/dev-workflow task 11` to initialize the task execution session
2. Branch `11-session-interface` already exists — switch to it with `git checkout 11-session-interface`
3. Dev type: **kernel** (load kernel dev-type reference + tau:go-patterns skill)
4. Session skeleton exists at `session/README.md` — no Go files yet
5. Read this plan and present it for review before proceeding to implementation

## Context

The session subsystem manages conversation turns so the kernel loop can maintain context across observe/think/act cycles. This is a sub-issue of #1 (Kernel Core Loop). We need the minimal `Session` interface and an in-memory implementation — just enough for the kernel loop to manage conversation history.

### Why this matters in the larger architecture

The kernel runtime's core cycle is: prompt → reason → tool calls → execute → observe → iterate. Each iteration sends the **full conversation history** to the model. That history includes user messages, assistant responses, tool call requests, and tool results correlated back to those requests. Session is the data structure that holds this history.

`protocol.Message` (Section 5.3 of the kernel concept) cannot represent tool-calling messages — it only has `Role string` and `Content any`. Rather than modifying `protocol.Message` (which would ripple across agent, orchestrate, and providers), session defines its own richer `Message` type. This is the foundation for the full session vision: Token Counter, Context Manager, and Compaction Strategies will build on top of this interface in subsequent issues.

## Design Decisions

### Message type: standalone, not embedding protocol.Message

`protocol.Message.Content` is `any` (supports multimodal content like vision arrays). Session messages need `Content string` for conversation history. Embedding would expose the wrong Content type and gain nothing — session has no need for the `Protocol` type or `NewMessage` constructor. This keeps session at **Level 0** with zero kernel package dependencies.

### ToolCall type: defined in session, not reusing response.ToolCall

`response.ToolCall` is a JSON deserialization struct (nested `Function` struct, `Type` field always `"function"`, json tags). Session needs a flat `{ID, Name, Arguments}` for history tracking. This avoids a dependency on `core/response` (Level 1) and keeps the mapping at the boundary where the kernel loop creates session messages from LLM responses.

### Messages() returns a defensive copy

Following the defensive-copy pattern used throughout the codebase. Copies both the outer slice and each message's `ToolCalls` slice. Since `ToolCall` fields are all strings (value types), shallow copy of ToolCall elements is sufficient.

### Timestamp is caller-provided

`AddMessage` stores the timestamp as given. The session is a store, not a message factory. The kernel loop sets timestamps when creating messages.

## Files

### New: `session/session.go` — Types and interface

- **Role constants**: `RoleSystem`, `RoleUser`, `RoleAssistant`, `RoleTool` (plain `string` constants, matching protocol.Message.Role convention)
- **Session interface**: `ID() string`, `AddMessage(msg Message)`, `Messages() []Message`, `Clear()`
- **Message struct**: `Role string`, `Content string`, `ToolCallID string`, `ToolCalls []ToolCall`, `Timestamp time.Time`
- **ToolCall struct**: `ID string`, `Name string`, `Arguments string`

Only dependency: `time` (stdlib).

### New: `session/memory.go` — In-memory implementation

- **memorySession** private struct: `id string`, `messages []Message`, `mu sync.RWMutex`
- **New() Session** constructor: UUIDv7 via `uuid.Must(uuid.NewV7()).String()`
- `ID()` — returns immutable id (no lock needed)
- `AddMessage()` — write lock, append to slice
- `Messages()` — read lock, deep-enough copy (outer slice + each ToolCalls slice)
- `Clear()` — write lock, reset to empty slice

Dependencies: `sync` (stdlib), `github.com/google/uuid` (already in go.mod).

### New: `session/session_test.go` — Black-box tests

Package `session_test`. Table-driven where applicable. Tests:

| Test | Verifies |
|------|----------|
| `TestNew` | Non-empty ID, empty messages |
| `TestSession_ID_Unique` | Two sessions have different IDs |
| `TestSession_ID_Stable` | Same session returns same ID across calls |
| `TestSession_AddMessage_And_Messages` | Single message round-trip with all fields |
| `TestSession_Messages_Order` | Multiple messages returned in insertion order |
| `TestSession_Messages_Roles` | Table-driven: each role (system, user, assistant, tool) |
| `TestSession_Messages_ToolCalls` | Assistant with ToolCalls + Tool with ToolCallID |
| `TestSession_Messages_DefensiveCopy` | Modifying returned slice doesn't affect session |
| `TestSession_Messages_ToolCalls_DefensiveCopy` | Modifying returned ToolCalls doesn't affect session |
| `TestSession_Clear` | Messages empty after Clear |
| `TestSession_Clear_ThenAdd` | New messages work after Clear |
| `TestSession_Concurrent_AddMessage` | 100 goroutines adding concurrently, all present |
| `TestSession_Concurrent_AddAndRead` | Concurrent Add + Messages, no race |
| `TestSession_Concurrent_AddAndClear` | Concurrent Add + Clear, no panic |

## Implementation Sequence

1. Create `session/session.go` — interface, Message, ToolCall, Role constants
2. Create `session/memory.go` — memorySession + New() constructor

## Validation

```bash
go vet ./session/...
go test ./session/...
go test -race ./session/...
go test ./... # full suite to verify no regressions
go mod tidy   # verify no changes
```

## Reference Files

- `core/protocol/message.go` — Message type this extends conceptually
- `core/response/tools.go` — ToolCall type we intentionally don't reuse
- `agent/agent.go` — Pattern: interface + private struct, UUIDv7 constructor
- `orchestrate/hub/hub.go` — Pattern: sync.RWMutex usage
