# Library Extraction

Decompose the TAU kernel into three independent libraries so the kernel focuses on agentic harness functionality while lower-level primitives are reusable standalone.

## Architecture

```
tau/agent (primitives — zero TAU deps)
tau/orchestrate (coordination — zero TAU deps)
    ↓ both consumed by ↓
tau/kernel (harness — the composition layer)
```

- **tau/agent**: Model communication, wire formats, streaming, providers, identities. Standalone utility for building model-facing capabilities.
- **tau/orchestrate**: Multi-agent coordination, state graphs, workflows, observability. Standalone framework for orchestrating participants. Uses a local interface for participant identity rather than importing agent types.
- **tau/kernel**: The harness that composes both — runtime loop, sessions, memory, tools, MCP, API. This is where `agent.Agent` satisfies orchestrate's participant contract.

## tau/agent

**Module**: `github.com/tailored-agentic-units/agent`
**Location**: `~/tau/agent`

### Package Layout

Root package defines the Agent interface and constructor. Flat layout — no `core/` grouping.

| Package | Source | Responsibility |
|---------|--------|----------------|
| Root (`agent`) | `kernel/agent/agent.go`, `kernel/agent/registry.go` | Agent interface, New(), Registry, AgentInfo |
| `config/` | `kernel/core/config/` | AgentConfig, ProviderConfig, ModelConfig, ClientConfig, Duration |
| `protocol/` | `kernel/core/protocol/` | Protocol constants, Message, Tool, ToolCall types |
| `response/` | `kernel/core/response/` | Response parsing — must adopt unified model (see Protocol Design) |
| `model/` | `kernel/core/model/` | Model runtime bridge (config → protocol support) |
| `client/` | `kernel/agent/client/` | HTTP transport with retry (429, 502-504), exponential backoff |
| `providers/` | `kernel/agent/providers/` | Provider interface + implementations (Ollama, Azure, Bedrock) |
| `request/` | `kernel/agent/request/` | Protocol-specific request construction |
| `mock/` | `kernel/agent/mock/` | MockAgent, MockClient, MockProvider |
| `format/` | NEW (from go-agents v0.5.0) | Wire format abstraction: Format interface, registry, OpenAI format, Converse format |
| `streaming/` | NEW (from go-agents v0.5.0) | Streaming transport: StreamReader interface, SSE reader, EventStream reader |
| `identities/` | NEW (from go-agents v0.5.0) | Credential management: AWSCredentialSource, Azure managed identity |

### Protocol Design Challenge

The go-agents v0.5.0 protocol restructuring was driven by the need to reduce friction between providers and formats. The key changes:

**Response model**: go-agents replaced separate `ChatResponse`/`ToolsResponse`/`StreamingChunk` with a unified `Response` type using a `ContentBlock` interface (`TextBlock`, `ToolUseBlock`). This was necessary because the format layer needs a single response contract that all formats produce, rather than protocol-specific types.

**Format layer**: Extracted wire marshaling from providers into dedicated `format.Format` interface with a registry. Providers handle transport and auth; formats handle request/response serialization. This separation enables adding new API formats (Anthropic, Google) without touching provider code.

**Streaming transport**: Extracted streaming into a `streaming.StreamReader` interface with SSE and EventStream implementations. Providers no longer implement streaming directly — they compose a StreamReader.

**Three-layer separation** (go-agents pattern):
```
Transport (Provider) → Format (Wire marshaling) → Response (Domain types)
```

**Kernel innovations to preserve**:
- **Agent Registry**: Named agent management with lazy instantiation, capability querying. Not present in go-agents. (See `kernel/agent/registry.go`)
- **Multi-turn flexibility**: Kernel's `Chat(ctx, []Message, ...opts)` passes full message history. go-agents v0.5.0 simplified to `Chat(ctx, string, ...opts)`. The design must support both use cases — harnesses need message-level control, simple callers want string prompts.

**Design goal**: Adopt go-agents v0.5.0's format/streaming/response architecture while preserving the kernel's registry and multi-turn patterns. The concept session must analyze the go-agents packages deeply to find the right integration.

### Reference: go-agents v0.5.0 Key Files

- `~/code/go-agents/pkg/format/format.go` — Format interface
- `~/code/go-agents/pkg/streaming/streaming.go` — StreamReader interface
- `~/code/go-agents/pkg/response/response.go` — Unified Response type
- `~/code/go-agents/pkg/providers/bedrock.go` — Bedrock provider (uses Converse format + EventStream)
- `~/code/go-agents/pkg/identities/` — Credential sourcing
- `~/code/go-agents/ARCHITECTURE.md` — Design patterns

### Reference: Kernel Agent Files

- `kernel/agent/agent.go` — Agent interface, New(), protocol methods
- `kernel/agent/registry.go` — Registry with lazy instantiation, AgentInfo
- `kernel/core/config/` — Configuration types
- `kernel/core/protocol/` — Protocol constants, Message, Tool, ToolCall
- `kernel/core/response/` — ChatResponse, ToolsResponse, StreamingChunk, EmbeddingsResponse, AudioResponse
- `kernel/agent/client/` — HTTP client with retry
- `kernel/agent/providers/` — Provider interface, Ollama, Azure
- `kernel/agent/request/` — Protocol-specific request builders
- `kernel/agent/mock/` — Testing doubles

### Divergence Table

| Aspect | Kernel (current) | go-agents v0.5.0 |
|--------|------------------|-------------------|
| Response types | Separate `ChatResponse`, `ToolsResponse`, `StreamingChunk` | Unified `Response` with `ContentBlock` interface |
| Wire format | Marshaling hardcoded inside each provider | `format.Format` interface with registry |
| Streaming | Provider returns `<-chan any` | `streaming.StreamReader` interface (SSE, EventStream) |
| Providers | Ollama, Azure | Ollama, Azure, Bedrock |
| Auth/Identity | Inline in providers | Dedicated `identities` package |
| Agent.Chat | `Chat(ctx, []Message, ...opts)` | `Chat(ctx, string, ...opts)` |
| Agent registry | `Registry` with lazy instantiation | Not present |
| Embed/Audio | Separate methods returning typed responses | Follows same unified pattern |

## tau/orchestrate

**Module**: `github.com/tailored-agentic-units/orchestrate`
**Location**: `~/tau/orchestrate`

### Package Layout

| Package | Source | Responsibility |
|---------|--------|----------------|
| `observability/` | `kernel/observability/` | Observer interface, Event, Level (OTel-aligned), SlogObserver, NoOpObserver, MultiObserver |
| `config/` | `kernel/orchestrate/config/` | HubConfig, GraphConfig, workflow pattern configs |
| `hub/` | `kernel/orchestrate/hub/` | Multi-agent coordination, messaging, pub/sub |
| `messaging/` | `kernel/orchestrate/messaging/` | Message types, builders, message patterns |
| `state/` | `kernel/orchestrate/state/` | State graphs, checkpoints, immutable state |
| `workflows/` | `kernel/orchestrate/workflows/` | ProcessChain, ProcessParallel, ProcessConditional |
| `examples/` | `kernel/orchestrate/examples/` | Usage examples |

### Dependency Hierarchy (internal)

```
Level 0: observability (no deps)
Level 0: messaging (no deps)
Level 1: config (no deps)
Level 2: hub (config, messaging)
Level 3: state (observability, config)
Level 4: workflows (observability, config, state)
```

### Decoupling from tau/agent

Currently `orchestrate/hub` imports `agent.Agent` for two purposes:

1. **Registration identity**: `ag.ID()` — used for agent lookup, routing, and lifecycle tracking
2. **Handler context**: `MessageContext.Agent` — exposes the full agent to message handlers so they can invoke agent capabilities (Chat, Tools, etc.)

The hub itself only needs `ID()`, but `MessageContext` passes the full agent to handlers. To decouple while preserving handler flexibility:

- Hub defines a local `Participant` interface: `ID() string`
- `MessageContext.Participant` replaces `MessageContext.Agent`
- The kernel (as composition layer) bridges the gap — when registering an agent with a hub, it can wrap or pass the agent since `agent.Agent` satisfies `Participant`
- Handlers that need full agent capabilities receive them through closure binding at registration time, not through `MessageContext`

This makes tau/orchestrate have **zero TAU dependencies** while still supporting the full agent interaction pattern when composed by the kernel.

### Reference: go-agents-orchestration Improvements

The go-agents-orchestration library (v0.3.1) has some refinements worth incorporating:

- `MultiObserver` for event broadcasting (already in kernel's observability)
- `NewGraphWithDeps` for explicit dependency management
- Thread-safe registries throughout
- State fields made public for JSON serialization

**Note**: go-agents-orchestration depends on go-agents v0.3.0 (stale). tau/orchestrate will be built from kernel code + cherry-picked improvements, not directly from go-agents-orchestration.

### Reference: Kernel Orchestrate Files

- `kernel/observability/observer.go` — Observer, Event, Level, EventType
- `kernel/observability/slog.go` — SlogObserver
- `kernel/observability/noop.go` — NoOpObserver
- `kernel/observability/multi.go` — MultiObserver
- `kernel/observability/registry.go` — Global observer registry
- `kernel/orchestrate/hub/hub.go` — Hub interface and implementation
- `kernel/orchestrate/hub/handler.go` — MessageHandler, MessageContext (exposes agent.Agent)
- `kernel/orchestrate/hub/channel.go` — MessageChannel
- `kernel/orchestrate/hub/metrics.go` — Hub metrics
- `kernel/orchestrate/messaging/` — Message types, builders
- `kernel/orchestrate/state/` — State, Graph, CheckpointStore
- `kernel/orchestrate/workflows/` — Chain, Parallel, Conditional patterns
- `kernel/orchestrate/config/` — Configuration types

## Execution Strategy

### Phase 1A: Build New Libraries

Sequential order — tau/agent first since orchestrate tests may need mock agents:

1. **tau/agent**: Initialize repo, port kernel packages with flat layout, integrate go-agents v0.5.0 improvements (format, streaming, response, identities, Bedrock), preserve registry, establish CI and tests, tag v0.1.0
2. **tau/orchestrate**: Initialize repo, port kernel observability + orchestrate packages, replace agent import with local Participant interface, incorporate go-agents-orchestration refinements, establish CI and tests, tag v0.1.0

### Phase 1A: Marketplace Plugins

See `~/tau/tau-marketplace/.claude/marketplace-refactor.md` for the full marketplace decomposition plan. This can proceed in parallel with or after the library work.

### Phase 1B: Kernel Integration

See `_project/post-extraction.md` for the full kernel resumption plan. This follows after libraries and marketplace are complete.

## Industry Context

OpenClaw (local-first AI agent runtime) and NemoClaw (NVIDIA's security-guardrailed agent stack built on OpenClaw) validate the kernel-as-harness pattern. The TAU approach of separating agent primitives from orchestration from the runtime harness maps directly to this industry direction:

- tau/agent ≈ agent primitives layer
- tau/orchestrate ≈ coordination/guardrails layer (NemoClaw's OpenShell)
- tau/kernel ≈ runtime harness (OpenClaw's agent runtime)
