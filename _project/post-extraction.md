# Post-Extraction: Kernel Resumption

Everything needed to resume kernel development after tau/agent and tau/orchestrate libraries are built and marketplace plugins are decomposed.

## Prerequisites

Before any kernel changes:

- [ ] tau/agent v0.1.0 tagged and stable
- [ ] tau/orchestrate v0.1.0 tagged and stable
- [ ] Marketplace plugins decomposed (at minimum, structural move of existing skills)
- [ ] tau-agent and tau-orchestrate marketplace skills written

## Kernel Refactoring

### Package Removal

These packages are extracted to standalone libraries and must be removed from the kernel:

| Remove | Moved To |
|--------|----------|
| `core/config/` | `tau/agent` config/ |
| `core/protocol/` | `tau/agent` protocol/ |
| `core/response/` | `tau/agent` response/ |
| `core/model/` | `tau/agent` model/ |
| `agent/` (all subpackages) | `tau/agent` root + subpackages |
| `observability/` | `tau/orchestrate` observability/ |
| `orchestrate/` (all subpackages) | `tau/orchestrate` packages |

### Import Replacement

Every file that imports extracted packages must be updated:

| Current Import | New Import |
|---------------|------------|
| `github.com/tailored-agentic-units/kernel/core/config` | `github.com/tailored-agentic-units/agent/config` |
| `github.com/tailored-agentic-units/kernel/core/protocol` | `github.com/tailored-agentic-units/agent/protocol` |
| `github.com/tailored-agentic-units/kernel/core/response` | `github.com/tailored-agentic-units/agent/response` |
| `github.com/tailored-agentic-units/kernel/core/model` | `github.com/tailored-agentic-units/agent/model` |
| `github.com/tailored-agentic-units/kernel/agent` | `agent "github.com/tailored-agentic-units/agent"` |
| `github.com/tailored-agentic-units/kernel/agent/client` | `github.com/tailored-agentic-units/agent/client` |
| `github.com/tailored-agentic-units/kernel/agent/mock` | `github.com/tailored-agentic-units/agent/mock` |
| `github.com/tailored-agentic-units/kernel/observability` | `github.com/tailored-agentic-units/orchestrate/observability` |

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
    tc.Function.Arguments
}
```

If tau/agent adopts the unified `Response` with `ContentBlock` interface from go-agents v0.5.0, this entire section must be rewritten to use the new response model. The exact changes depend on the API surface that emerges from the tau/agent concept session.

### go.mod Changes

**Remove**:
```
connectrpc.com/connect v1.19.1
google.golang.org/protobuf v1.36.11
```

**Add**:
```
github.com/tailored-agentic-units/agent v0.1.0
github.com/tailored-agentic-units/orchestrate v0.1.0
```

**Keep**:
```
github.com/google/uuid v1.6.0
```

Note: uuid may also be used by tau/agent for agent IDs. If so, kernel may no longer need it directly — verify after extraction.

### Dead Infrastructure Removal

Per the architecture decision in `_project/objective.md` (pure HTTP + JSON + SSE replaces ConnectRPC):

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
├── cmd/               # Entry points (kernel, prompt-agent)
├── scripts/           # Infrastructure scripts
├── tests/             # Integration tests
├── .claude/           # Configuration and skills
└── .github/           # CI workflows
```

### Post-Refactor go.mod

```
module github.com/tailored-agentic-units/kernel

require (
    github.com/tailored-agentic-units/agent v0.1.0
    github.com/tailored-agentic-units/orchestrate v0.1.0
)
```

## Resuming Phase 1 Objectives

After refactoring, the remaining Phase 1 sub-issues become actionable:

### #26 — Multi-session kernel

Parent/child session relationships for subagent orchestration. Now built on tau/agent's Agent and Registry, and tau/orchestrate's Hub with Participant interface. The kernel is the composition layer where agent.Agent satisfies hub.Participant.

### #27 — HTTP API with SSE streaming

Pure HTTP + JSON + SSE transport replacing ConnectRPC. Go types as source of truth, OpenAPI for schema docs. HTTP handlers in `api/` package. Now built on tau/agent's streaming infrastructure (StreamReader interface).

### #28 — Server entry point

Server binary composing the kernel with HTTP API. Final deliverable for v0.1.0.

## Documentation Updates

These must happen during the refactor, not after:

| Document | Updates Needed |
|----------|---------------|
| `_project/README.md` | Vision statement (remove ConnectRPC), subsystem topology (only kernel-local packages), dependency hierarchy (now includes tau/agent + tau/orchestrate), model compatibility (defer to tau/agent), build order |
| `_project/phase.md` | Phase 1 scope with library extraction as Phase 1A complete, remaining objectives as Phase 1B |
| `_project/objective.md` | Re-scope #26-28 descriptions to reference tau/agent and tau/orchestrate types |
| `.claude/CLAUDE.md` | Project structure (reduced), dependency hierarchy (library-based), commands (remove proto), skills table |
| `.claude/skills/kernel-dev/SKILL.md` | Package responsibilities (reduced to kernel-local), extension patterns (now via library APIs) |
| `README.md` | Architecture description, dependency list, quick start commands |

## Project Management Updates

| Action | Details |
|--------|---------|
| Update Objective #2 status | "In Progress" on project board |
| Reassign #5-9 | Close on kernel, create equivalents on tau/orchestrate repo |
| Reassign #10 | Close on kernel, create equivalent on tau/agent repo |
| Create Phase 1A | "Library Extraction" phase on project board, mark complete |
| Archive concept | Move `_project/library-extraction.md` to `.claude/context/concepts/.archive/` |
| Update kernel-dev skill | Reflect reduced package scope |

## Verification

After all changes:

- [ ] `go test ./...` passes with no extracted package imports
- [ ] `go vet ./...` clean
- [ ] No references to `core/`, `agent/`, `orchestrate/`, `observability/` as local packages
- [ ] No references to `connectrpc`, `protobuf`, or `rpc/`
- [ ] `go.mod` only depends on tau/agent, tau/orchestrate, and direct dependencies
- [ ] All documentation reflects post-extraction architecture
- [ ] Project board reflects Phase 1A complete and Phase 1B in progress
