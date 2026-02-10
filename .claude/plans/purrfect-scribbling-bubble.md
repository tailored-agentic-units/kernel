# Issue #11 — Session Interface and In-Memory Implementation

## Context

The session subsystem manages conversation turns so the kernel loop can maintain context across observe/think/act cycles. The kernel runtime's core cycle (prompt → reason → tool calls → execute → observe → iterate) sends full conversation history to the model each iteration. Session is the data structure that holds this history.

During planning, we identified that `protocol.Message` lacks tool call/result support (Known Gap #3 in `_project/README.md`). Rather than working around this by defining a separate message type in session, we evolve `protocol.Message` to be the canonical conversation message type — addressing the gap at the correct level of the dependency stack.

This also establishes a design principle for the project: core data structures are the protocol standard; subsystems use them natively rather than creating package-specific mutations.

## Implementation

### Step 1: Add design principle to `.claude/CLAUDE.md`

Add a "Design Principles" section after "Package Dependency Hierarchy":

- Every package's exported data structures are its protocol. Higher-level packages that depend on it consume those types natively.
- `core/` types are the foundational protocol — subsystems build on them, not around them. If a subsystem needs to mutate a core type signature, evolve the core type instead.
- This applies at every level of the hierarchy: if `session` defines `Session`, then `kernel` uses `session.Session` directly — no wrapping, no re-definition.
- When planning implementation, address gaps at the lowest affected dependency level rather than working around them at higher levels.

### Step 2: Evolve `core/protocol/message.go` — Address Known Gap #3

Add tool call support to `protocol.Message`:

- **ToolCall struct**: `ID string`, `Name string`, `Arguments string` — flat struct, not reusing `response.ToolCall` (which is a JSON deserialization struct with nested `Function`)
- **Message fields added**: `ToolCallID string` (json `"tool_call_id,omitempty"`), `ToolCalls []ToolCall` (json `"tool_calls,omitempty"`)
- **Role constants**: `RoleSystem`, `RoleUser`, `RoleAssistant`, `RoleTool` — plain string constants, currently implicit throughout the codebase
- Update `NewMessage` or leave as-is (new fields are zero-valued by default, existing callers unaffected)
- Update existing tests if any exist for `protocol.Message`

`omitempty` on both new fields ensures existing JSON serialization is unchanged when fields are empty.

### Step 3: Update `_project/README.md` — Mark Known Gap #3 resolved

Update Known Gap #3 to reflect that it has been addressed.

### Step 4: Create `session/session.go` — Interface and session-level types

- **Session interface**: `ID() string`, `AddMessage(msg protocol.Message)`, `Messages() []protocol.Message`, `Clear()`
- **Entry struct** (if Timestamp metadata is needed): deferred — the Session interface works with `protocol.Message` directly. Timestamp can be added when the kernel loop needs it.

Only dependency: `core/protocol`.

### Step 5: Create `session/memory.go` — In-memory implementation

- **memorySession** private struct: `id string`, `messages []protocol.Message`, `mu sync.RWMutex`
- **New() Session** constructor: UUIDv7 via `uuid.Must(uuid.NewV7()).String()`
- `ID()` — returns immutable id (no lock needed)
- `AddMessage(msg protocol.Message)` — write lock, append
- `Messages()` — read lock, defensive copy (outer slice + each message's ToolCalls slice; all other fields are value types)
- `Clear()` — write lock, reset to empty slice

Dependencies: `core/protocol`, `sync` (stdlib), `github.com/google/uuid` (already in go.mod).

### Step 6: Create `session/session_test.go` — Black-box tests

Package `session_test`. Table-driven where applicable:

| Test | Verifies |
|------|----------|
| `TestNew` | Non-empty ID, empty messages |
| `TestSession_ID_Unique` | Two sessions have different IDs |
| `TestSession_ID_Stable` | Same session returns same ID across calls |
| `TestSession_AddMessage_And_Messages` | Single message round-trip with all fields |
| `TestSession_Messages_Order` | Multiple messages returned in insertion order |
| `TestSession_Messages_Roles` | Table-driven: each role constant |
| `TestSession_Messages_ToolCalls` | Assistant with ToolCalls + Tool with ToolCallID |
| `TestSession_Messages_DefensiveCopy` | Modifying returned slice doesn't affect session |
| `TestSession_Messages_ToolCalls_DefensiveCopy` | Modifying returned ToolCalls doesn't affect session |
| `TestSession_Clear` | Messages empty after Clear |
| `TestSession_Clear_ThenAdd` | New messages work after Clear |
| `TestSession_Concurrent_AddMessage` | 100 goroutines adding concurrently, all present |
| `TestSession_Concurrent_AddAndRead` | Concurrent Add + Messages, no race |
| `TestSession_Concurrent_AddAndClear` | Concurrent Add + Clear, no panic |

### Step 7: Add protocol.Message tests for new fields

Ensure the new ToolCall, ToolCallID fields and Role constants are tested in `core/protocol/`.

## Verification

```bash
go vet ./...
go test ./core/protocol/...
go test ./session/...
go test -race ./session/...
go test ./...          # full suite — no regressions
go mod tidy            # verify no changes
```

## Files Modified/Created

| File | Action |
|------|--------|
| `.claude/CLAUDE.md` | Add Design Principles section |
| `core/protocol/message.go` | Add ToolCall, ToolCallID, Role constants |
| `core/protocol/message_test.go` | Add/update tests for new fields |
| `_project/README.md` | Update Known Gap #3 |
| `session/session.go` | Create — interface + types |
| `session/memory.go` | Create — in-memory implementation |
| `session/session_test.go` | Create — black-box tests |

## Reference Files

- `core/protocol/message.go` — evolving this type
- `core/response/tools.go` — ToolCall type we intentionally don't reuse (JSON deserialization struct)
- `agent/agent.go` — pattern: interface + private struct, UUIDv7 constructor
- `orchestrate/hub/hub.go` — pattern: sync.RWMutex usage
- `_project/README.md` — Known Gap #3, dependency hierarchy
