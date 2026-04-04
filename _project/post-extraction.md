# Post-Extraction: Kernel Resumption

Everything needed to resume kernel development after the five TAU libraries are built.

## Prerequisites

Before any kernel changes:

- [ ] tau/protocol v0.1.0 tagged and stable
- [ ] tau/format v0.1.0 tagged and stable
- [ ] tau/provider v0.1.0 tagged and stable
- [ ] tau/agent v0.1.0 tagged and stable
- [ ] tau/orchestrate v0.1.0 tagged and stable

## Kernel Refactoring

### Package Removal

These packages are extracted to standalone libraries and must be removed from the kernel:

| Remove | Moved To |
|--------|----------|
| `core/config/` | tau/protocol config/ |
| `core/protocol/` | tau/protocol root |
| `core/response/` | tau/protocol response/ |
| `core/model/` | tau/protocol model/ |
| `agent/` (all subpackages) | tau/agent root + subpackages |
| `observability/` | tau/orchestrate observability/ |
| `orchestrate/` (all subpackages) | tau/orchestrate packages |

### Import Replacement

| Current Import | New Import | Module |
|---------------|------------|--------|
| `kernel/core/config` | `github.com/tailored-agentic-units/protocol/config` | tau/protocol |
| `kernel/core/protocol` | `protocol "github.com/tailored-agentic-units/protocol"` | tau/protocol |
| `kernel/core/response` | `github.com/tailored-agentic-units/protocol/response` | tau/protocol |
| `kernel/core/model` | `github.com/tailored-agentic-units/protocol/model` | tau/protocol |
| `kernel/agent` | `agent "github.com/tailored-agentic-units/agent"` | tau/agent |
| `kernel/agent/client` | `github.com/tailored-agentic-units/agent/client` | tau/agent |
| `kernel/agent/mock` | `github.com/tailored-agentic-units/agent/mock` | tau/agent |
| `kernel/agent/registry` (was in root) | `github.com/tailored-agentic-units/agent/registry` | tau/agent |
| `kernel/observability` | `github.com/tailored-agentic-units/orchestrate/observability` | tau/orchestrate |
| N/A (new) | `github.com/tailored-agentic-units/format` | tau/format |
| N/A (new) | `github.com/tailored-agentic-units/provider` | tau/provider |
| `protocol.Tool` | `format.ToolDefinition` | tau/format |
| `response.ChatResponse` | `response.Response` | tau/protocol |
| `response.ToolsResponse` | `response.Response` | tau/protocol |
| `response.StreamingChunk` | `response.StreamingResponse` | tau/protocol |

### Response Model Migration

The kernel runtime loop (`kernel/kernel.go`) directly accesses OpenAI-shaped response fields:

```go
// Current (lines 203-235 of kernel/kernel.go):
resp, err := k.agent.Tools(ctx, messages, k.tools.List())
choice := resp.Choices[0]
if len(choice.Message.ToolCalls) == 0 {
    result.Response = choice.Message.Content
}
for _, tc := range choice.Message.ToolCalls {
    tc.Function.Name
    tc.Function.Arguments  // string
}
```

Rewrite to unified Response model:

```go
// New (tau/protocol response types):
resp, err := k.agent.Tools(ctx, messages, tools)
if len(resp.ToolCalls()) == 0 {
    result.Response = resp.Text()
}
for _, tc := range resp.ToolCalls() {
    tc.Name
    tc.Input  // map[string]any — must json.Marshal before tools.Execute
}
```

Key change: `ToolUseBlock.Input` is `map[string]any`, not a JSON string. The kernel must `json.Marshal(tc.Input)` before passing to `tools.Execute(ctx, name, json.RawMessage)`.

### Session + Context Management

The agent interface now takes `[]Message` directly — the agent is stateless transport. The kernel already builds messages via `k.buildMessages()`. The change is that the agent no longer does internal message construction. Context management strategies (sliding window, summarization) are explicit kernel responsibilities.

### Tool Type Migration

The kernel's `tools/` package currently uses `protocol.Tool` for tool schemas. Post-extraction, tool definitions are `format.ToolDefinition` from tau/format:

```go
// Current:
k.tools.List() returns []protocol.Tool

// New:
k.tools.List() returns []format.ToolDefinition
```

### go.mod Changes

**Remove**:
```
connectrpc.com/connect v1.19.1
google.golang.org/protobuf v1.36.11
```

**Add**:
```
github.com/tailored-agentic-units/protocol v0.1.0
github.com/tailored-agentic-units/format v0.1.0
github.com/tailored-agentic-units/provider v0.1.0  // transitively via agent
github.com/tailored-agentic-units/agent v0.1.0
github.com/tailored-agentic-units/orchestrate v0.1.0
```

**Keep**:
```
github.com/google/uuid v1.6.0
```

### Dead Infrastructure Removal

Per the architecture decision (pure HTTP + JSON + SSE replaces ConnectRPC):

- Remove `rpc/` directory (buf configs, proto definitions, generated code)
- Remove `proto` label from the repository
- Remove ConnectRPC references from `_project/README.md` vision statement

### Post-Refactor Kernel Structure

```
kernel/
├── _project/          # Project management
├── kernel/            # Runtime loop
├── session/           # Conversation management
├── memory/            # Context composition (FileStore, Cache)
├── tools/             # Tool registry and execution
├── mcp/               # MCP client
├── api/               # HTTP + SSE handlers (new, replaces rpc/)
├── cmd/               # Entry points
├── scripts/           # Infrastructure scripts
├── tests/             # Integration tests
├── .claude/           # Configuration and skills
└── .github/           # CI workflows
```

### Post-Refactor go.mod

```
module github.com/tailored-agentic-units/kernel

require (
    github.com/tailored-agentic-units/protocol v0.1.0
    github.com/tailored-agentic-units/format v0.1.0
    github.com/tailored-agentic-units/agent v0.1.0
    github.com/tailored-agentic-units/orchestrate v0.1.0
)
```

## Resuming Phase 1 Objectives

After refactoring, the remaining Phase 1 sub-issues become actionable:

### #26 — Multi-session kernel

Parent/child session relationships for subagent orchestration. Now built on tau/agent's Agent and registry, and tau/orchestrate's Hub with Participant interface. The kernel is the composition layer where agent.Agent satisfies hub.Participant.

### #27 — HTTP API with SSE streaming

Pure HTTP + JSON + SSE transport replacing ConnectRPC. Go types as source of truth, OpenAPI for schema docs. HTTP handlers in `api/` package. Now built on tau/agent's streaming infrastructure (StreamReader interface from tau/protocol).

### #28 — Server entry point

Server binary composing the kernel with HTTP API. Final deliverable for v0.1.0.

## Documentation Updates

These must happen during the refactor, not after:

| Document | Updates Needed |
|----------|---------------|
| `_project/README.md` | Vision (remove ConnectRPC), subsystem topology (kernel-local only), dependency hierarchy (5-library architecture), build order |
| `_project/phase.md` | Phase 1 scope post-extraction |
| `_project/objective.md` | Re-scope #26-28 to reference tau library types |
| `.claude/CLAUDE.md` | Project structure (reduced), dependency hierarchy (library-based), commands (remove proto), skills table |
| `.claude/skills/kernel-dev/SKILL.md` | Package responsibilities (kernel-local), extension patterns (via library APIs) |
| `README.md` | Architecture description, dependency list, quick start commands |

## Project Management Updates

| Action | Details |
|--------|---------|
| Reassign #5-9 | Close on kernel, create equivalents on tau/orchestrate repo |
| Reassign #10 | Close on kernel, create equivalent on tau/agent repo |
| Archive `_project/library-extraction.md` | Concept fully realized in library repos |
| Update kernel-dev skill | Reflect reduced package scope |

## Verification

After all changes:

- [ ] `go test ./...` passes with no extracted package imports
- [ ] `go vet ./...` clean
- [ ] No references to `core/`, `agent/`, `orchestrate/`, `observability/` as local packages
- [ ] No references to `connectrpc`, `protobuf`, or `rpc/`
- [ ] `go.mod` depends on tau/protocol, tau/format, tau/agent, tau/orchestrate
- [ ] All documentation reflects post-extraction architecture
