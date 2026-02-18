# Objective: Kernel Interface

**Issue:** #2
**Phase:** Phase 1 — Foundation (v0.1.0)

## Scope

Establish the kernel's HTTP interface — the sole extensibility boundary through which external services connect. Resolves foundational architecture decisions: agent registry, multi-session runtime, streaming-first loop, observer pattern, and pure HTTP + SSE transport.

## Sub-Issues

| # | Title | Status |
|---|-------|--------|
| 23 | Streaming tools protocol | PR #30 |
| 24 | Agent registry | PR #31 |
| 25 | Kernel observer | PR #32 |
| 26 | Multi-session kernel | Open |
| 27 | HTTP API with SSE streaming | Open |
| 28 | Server entry point | Open |

## Dependency Graph

```
[#23: Streaming Tools] [#24: Agent Registry] [#25: Kernel Observer]
         \                    |                    /
          v                   v                  v
              [#26: Multi-session Kernel]
                       |
                       v
              [#27: HTTP API + SSE]
                       |
                       v
              [#28: Server Entry Point]
```

## Known Gaps

- **Subsystem observability** — foundation packages (`memory`, `tools`, `session`) should accept an Observer and define their own event types, similar to `orchestrate`. The kernel would pass its observer down during initialization. The kernel's `memory loaded` log was removed in #25 because the kernel shouldn't log on behalf of a subsystem — when `memory` gets its own Observer, it would emit a `memory.load` event at the appropriate level. Follow-up to #25.

## Architecture Decisions

- **Agent registry is kernel infrastructure** — named agents (model-aligned: qwen3-8b, llava-13b, gpt-5), capability querying. Instance-owned, not global. The `memory/agents/` namespace is reserved for subagent profile content.
- **Sessions are the context boundary** — all subsystem integrations scoped to sessions. Per-session memory via Cache.
- **Streaming-first** — `ToolsStream()` added to Agent interface. Kernel loop uses streaming by default.
- **Observer replaces logger** — orchestrate Observer pattern with kernel-specific event types. Slog adapter for backward compatibility. Absorbs Objective #4's logger concern.
- **Pure HTTP + SSE** — replaces ConnectRPC. Standard net/http + JSON + Server-Sent Events. Go types as source of truth. OpenAPI for schema docs.
- **HTTP handlers in `api/` package** — replaces `rpc/`. Clean separation from kernel runtime.
- **Child session foundation** — parent ID and inheritance config in session model. Full subagent orchestration deferred to a future objective.
