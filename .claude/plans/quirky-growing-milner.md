# Objective Planning: #2 — Kernel Interface

## Context

Objective #1 (Kernel Core Loop) is complete. The kernel has a functioning single-agent runtime loop. Before establishing the HTTP interface (the sole extensibility boundary), we need to resolve foundational architecture decisions that will be difficult to change once the boundary is set:

- **Agent management**: Callers shouldn't need full agent configs. The kernel needs a registry where agents are registered by name with capability awareness.
- **Multi-session**: The kernel is a singleton managing multiple sessions. Sessions are the context boundary for all subsystem integrations.
- **Streaming**: Build around streaming from the start, not retrofit later.
- **Observability**: Unify the kernel's ad-hoc logger with the observer pattern from orchestrate.
- **Transport**: Pure HTTP + SSE replaces ConnectRPC. Go types as source of truth. OpenAPI for schema documentation. Follows agent-lab patterns (`~/code/agent-lab`).

## Architecture

### Revised Kernel Model

```
Kernel (singleton)
├── Agent Registry (kernel-level, like tools registry)
│   ├── "qwen3-8b"  → AgentConfig (qwen3:8b, chat+tools)
│   ├── "llava-13b" → AgentConfig (llava:13b, chat+vision+tools)
│   └── "gpt-5"     → AgentConfig (gpt-5, all capabilities)
├── Tool Registry (shared, existing global)
├── Observer (replaces logger, event-driven)
├── Session Manager
│   ├── Session A
│   │   ├── agent: "default" (resolved from registry)
│   │   ├── memory: session-scoped Cache
│   │   ├── messages: conversation history
│   │   ├── status: active/running/completed/error
│   │   └── parent: nil (root session)
│   └── Session B
│       └── ...
└── Config (kernel-level: system prompt, max iterations, defaults)
```

### HTTP API Surface

```
# Agents
GET    /agents                         List registered agents and capabilities

# Sessions
POST   /sessions                       Create session (agent name, bootstrap context)
GET    /sessions/{id}                  Get session metadata and status
POST   /sessions/{id}/run              Submit prompt → SSE stream of events

# Session Memory (session-scoped context boundary)
GET    /sessions/{id}/memory           List memory keys
GET    /sessions/{id}/memory/{key...}  Get memory entry
PUT    /sessions/{id}/memory/{key...}  Save memory entry
DELETE /sessions/{id}/memory/{key...}  Delete memory entry

# Tools
GET    /tools                          List registered tools
```

SSE event stream for `/run`:
```
event: status
data: {"status":"running","message":"Iteration 1"}

event: tool_call
data: {"tool_name":"read_file","arguments":"{...}","result":"..."}

event: token
data: {"text":"Here is the answer..."}

event: status
data: {"status":"completed","message":"Run complete"}

data: [DONE]
```

### Key Design Decisions

1. **Agent registry is kernel infrastructure** — like the tools registry, but instance-owned (not global). Named agents with configs, capability querying. The `memory/agents/` namespace is reserved for subagent profile *content* (personality, behavior definitions), NOT registration infrastructure.

2. **Sessions are the context boundary** — all subsystem integrations (memory, tools access, agent selection) are scoped to sessions. Per-session memory via Cache. Child sessions inherit parent context with optional restriction.

3. **Streaming-first kernel loop** — `ToolsStream()` added to Agent interface. The kernel loop uses streaming by default, emitting events as they occur. Observer receives structured events.

4. **Observer replaces logger** — the kernel adopts orchestrate's `Observer` pattern. Kernel-specific event types (iteration, tool call, response). Slog adapter provides backward compatibility. Absorbs Objective #4's logger concern.

5. **Pure HTTP + SSE** — remove ConnectRPC dependency. Standard `net/http` handlers with JSON request/response. SSE for streaming. Go structs as source of truth. OpenAPI spec for documentation. Follows agent-lab patterns.

6. **HTTP handlers in `api/` package** — new top-level package replacing `rpc/`. Clean separation: `api/` contains all HTTP infrastructure wired to the kernel. Existing `rpc/` (proto + generated code) removed.

7. **Child session foundation only** — session model includes parent ID and inheritance config. Full subagent orchestration deferred to a future objective (tracked in planning notes).

## Sub-Issues

### Dependency Graph

```
[#1: Streaming Tools] [#2: Agent Registry] [#3: Kernel Observer]
         \                    |                    /
          \                   |                   /
           v                  v                  v
              [#4: Multi-session Kernel]
                       |
                       v
              [#5: HTTP API + SSE]
                       |
                       v
              [#6: Server Entry Point]
```

Sub-issues 1-3 are independent foundations (parallelizable). Sub-issue 4 integrates them. Sub-issues 5-6 build the HTTP layer on top.

---

### Sub-issue 1: Streaming tools protocol

**Package**: `agent/`
**Labels**: `agent`, `feature`

Add `ToolsStream` to the Agent interface, following the established ChatStream/VisionStream pattern.

**Scope**:
- Add `ToolsStream(ctx, messages, tools, opts) (<-chan *StreamingChunk, error)` to Agent interface
- Implement in `agent/agent.go` — sets `stream: true`, calls `client.ExecuteStream()`
- Add to `agent/mock/` mock agent
- All streaming plumbing already exists: `ParseToolsStreamChunk()`, provider `ProcessStreamResponse()`, client `ExecuteStream()`, `Protocol.Tools.SupportsStreaming() == true`

**Key files**: `agent/agent.go`, `agent/mock/mock.go`

---

### Sub-issue 2: Agent registry

**Package**: `kernel/`
**Labels**: `kernel`, `feature`

Agent registration as kernel infrastructure — named agents, config management, capability querying.

**Scope**:
- Registry type with `Register(name, AgentConfig)`, `Get(name) (Agent, error)`, `List() []AgentInfo`, `Unregister(name)`
- Lazy agent instantiation: configs registered, agents created on first use (or session creation)
- Capability querying: which protocols does a named agent support? (derived from `ModelConfig.Capabilities` keys)
- Kernel config extended to include `agents` map (named agent configs)
- Registry owned by Kernel instance (not global) for test isolation

**Key files**: `kernel/registry.go`, `kernel/config.go`

---

### Sub-issue 3: Kernel observer (absorbs Objective #4 logger concern)

**Package**: `kernel/`
**Labels**: `kernel`, `feature`

Replace the kernel's ad-hoc slog logger with the orchestrate Observer pattern. Defines kernel-specific event types.

**Scope**:
- Adopt `observability.Observer` interface from `orchestrate/observability/`
- Define kernel event types: `kernel.iteration.start`, `kernel.tool.call`, `kernel.tool.complete`, `kernel.response`, `kernel.run.start`, `kernel.run.complete`, `kernel.error`
- Replace `k.log.Info/Debug/Warn` calls with `k.observer.OnEvent()`
- `WithObserver(o Observer)` option replaces `WithLogger(l *slog.Logger)`
- Slog adapter: Observer implementation that writes structured events to slog
- Noop observer as default (zero overhead when unobserved)
- Update Objective #4 issue to remove logger concern, add planning note that observer subsumes it

**Key files**: `kernel/observer.go`, `kernel/kernel.go`, `orchestrate/observability/observer.go` (reuse)

---

### Sub-issue 4: Multi-session kernel

**Package**: `kernel/`
**Labels**: `kernel`, `feature`

Refactor the kernel from single-agent/single-session to a multi-session runtime with agent registry integration.

**Scope**:
- **Session manager**: Create, get, list sessions with lifecycle tracking (active → running → completed/error)
- **Session configuration**: Agent name (resolved from registry), memory config, parent session ID (foundation for child sessions)
- **Per-session memory**: Each session gets a scoped `memory.Cache`. InjectContext maps to `cache.Save()`.
- **Kernel.Run refactor**: Takes session ID + prompt. Resolves agent from session config → registry. Uses streaming (ToolsStream). Emits events via observer.
- **Child session foundation**: Session model includes `ParentID` and inheritance config. No subagent spawning or cross-session orchestration. Add planning notes to identify the future objective for full implementation.
- **Kernel.New refactor**: Kernel initialized with shared config + agent registry. Sessions created dynamically, not at init.

**Depends on**: Sub-issues 1, 2, 3
**Key files**: `kernel/kernel.go`, `kernel/config.go`, `kernel/session.go` (new)

---

### Sub-issue 5: HTTP API with SSE streaming

**Package**: `api/` (new, replaces `rpc/`)
**Labels**: `kernel`, `feature`

Pure HTTP handlers with SSE streaming, following agent-lab patterns. Remove ConnectRPC.

**Scope**:
- **Remove ConnectRPC**: Delete `rpc/gen/`, `rpc/proto/`, `rpc/buf.yaml`, `rpc/buf.gen.yaml`. Remove `connectrpc.com/connect` and `google.golang.org/protobuf` from `go.mod`.
- **Request/response types**: Go structs with JSON tags as source of truth (`api/types.go`)
- **HTTP handlers**: Agents (list), Sessions (create, get, run), Memory (list, get, save, delete per session), Tools (list) (`api/handler.go`)
- **SSE streaming**: `writeSSEStream()` utility following agent-lab pattern — typed events (`status`, `tool_call`, `token`), JSON data, flush per event, `[DONE]` marker
- **Response helpers**: `RespondJSON()`, `RespondError()` (`api/respond.go`)
- **Error mapping**: Kernel errors → HTTP status codes
- **Handler tests**: Using `httptest.NewServer` and SSE client parsing
- **OpenAPI spec**: `GET /openapi.json` endpoint served by the kernel. Spec built alongside handlers. The kernel serves the JSON — rendering is handled externally.

**Reference**: `~/code/agent-lab` — `internal/agents/handler.go` (SSE pattern), `pkg/handlers/` (response helpers), `pkg/openapi/` (spec building)

**Depends on**: Sub-issue 4
**Key files**: `api/handler.go`, `api/types.go`, `api/respond.go`, `api/sse.go`, `api/openapi.go`

---

### Sub-issue 6: Server entry point

**Package**: `cmd/server/`
**Labels**: `kernel`, `feature`

Server binary that starts the kernel as an HTTP service.

**Scope**:
- `cmd/server/main.go` — HTTP server with handler routing via `http.ServeMux`
- Server configuration: listen address, kernel config file path, read/write timeouts
- Built-in tool registration (shared with or extracted from `cmd/kernel/tools.go`)
- Graceful shutdown with signal handling (SIGINT/SIGTERM) and shutdown timeout
- Structured logging via observer → slog adapter
- **Scalar UI**: Standalone Bun project in `scalar/` directory. Fetches `GET /openapi.json` from the running kernel server. Fully separate from the Go binary — no embedding, no `//go:embed`. Development-time only.

**Depends on**: Sub-issue 5
**Key files**: `cmd/server/main.go`, `cmd/server/tools.go`, `scalar/` (Bun project)

---

## Scope Boundary Notes

### In scope
- Agent registry as kernel infrastructure
- Multi-session kernel with streaming-first loop
- Observer replacing logger
- Pure HTTP API with SSE streaming + OpenAPI
- Child session model (parent ID, inheritance config)
- Remove ConnectRPC dependency

### Deferred (with planning notes)
- **Full subagent orchestration**: Child session spawning, memory inheritance execution, result propagation. Identify target objective and add planning notes to the issue body.
- **Runtime agent registration API via HTTP**: Managing agent configs through the HTTP interface (vs config file). Future API evolution.
- **Token-level streaming accumulation**: If providers send partial tool call arguments in chunks, handling that accumulation. Initial implementation streams at tool-call-complete granularity.

## Infrastructure

| Field | Value |
|-------|-------|
| Task issue type ID | `IT_kwDOD155C84B2CKc` |
| Bug issue type ID | `IT_kwDOD155C84B2CKd` |
| Milestone | `Phase 1 - Foundation` (#1) |
| Project | TAU Platform (#1) |
| Phase field ID | `PVTSSF_lADOD155C84BN3wGzg8vZF8` |
| Phase 1 option ID | `40668fd6` |
| Backlog option ID | `1724a68a` |
| Parent issue ID | `I_kwDORK9Egc7pITez` (issue #2) |

## Collateral Updates

- **Objective #2 (issue #2)**: Update issue body to reflect pure HTTP + SSE approach and expanded scope (agent registry, multi-session, observer).
- **Objective #4 (issue #3)**: Update issue body to remove deeper logger configuration concern, add note that observer (from this objective) subsumes it.
- **`_project/objective.md`**: Replace with Objective #2 content after sub-issue creation.
- **`_project/phase.md`**: Update Objective #2 status to In Progress.
- **`_project/README.md`**: Update subsystem topology — add `api/` package, remove `rpc/` references as appropriate.
- **Future objective planning note**: Identify which objective gets full child session / subagent orchestration.
