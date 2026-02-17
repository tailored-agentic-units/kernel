---
name: kernel-dev
description: >
  REQUIRED for contributing to the TAU kernel. Use when adding providers,
  protocols, workflow patterns, observers, state graph extensions, or modifying
  kernel architecture. Triggers: new provider, new protocol, new workflow,
  Observer interface, architecture, testing, coverage, extension patterns.
---

# kernel Contributor Guide

## When This Skill Applies

- Adding new LLM providers or protocol support to `agent/`
- Adding new workflow patterns, observers, or state graph extensions to `orchestrate/`
- Adding implementation to skeleton packages (mcp)
- Extending the memory context pipeline (Store backends, skill loading, agent profiles)
- Extending the kernel runtime loop or adding ConnectRPC service implementation
- Architectural decisions affecting package boundaries
- Writing tests for any kernel package

## Architecture

The kernel is a single Go module (`github.com/tailored-agentic-units/kernel`) containing integrated subsystems. All packages share one version.

### Package Dependency Hierarchy

```
Level 0: core/config, core/protocol          (zero internal deps)
Level 1: core/response, core/model           (depends on Level 0)
Level 2: agent/providers, agent/request       (depends on Level 0-1)
Level 3: agent (root), agent/client           (depends on Level 0-2)
Level 4: agent/mock                           (depends on Level 0-3)
Level 5: orchestrate/observability            (zero internal deps)
Level 6: orchestrate/messaging, orchestrate/config  (Level 5)
Level 7: orchestrate/hub                      (depends on Level 3-6)
Level 8: orchestrate/state                    (depends on Level 5-6)
Level 9: orchestrate/workflows                (depends on Level 5-8)

Foundation (Level 0 — depend only on core/protocol):
  memory, tools, session

Level 10: kernel (depends on agent, session, memory, tools, core)
```

Dependencies only flow downward. Never import a higher-level package from a lower-level one.

### Package Responsibilities

| Package | Responsibility | Key Interfaces |
|---------|---------------|----------------|
| `core/config` | Configuration types, duration parsing | `AgentConfig`, `ProviderConfig` |
| `core/protocol` | Protocol constants, message types | `Protocol`, `Message` |
| `core/response` | Response parsing, streaming | `ChatResponse`, `ToolsResponse` |
| `core/model` | Model runtime type | `Model` |
| `agent` | Agent interface, lifecycle, named agent registry | `Agent`, `Registry`, `AgentInfo` |
| `agent/client` | HTTP transport, retry | `Client` |
| `agent/providers` | LLM platform adapters | `Provider`, `Registry` |
| `agent/request` | Request construction | `Builder` |
| `agent/mock` | Test doubles | `MockAgent`, `MockProvider` |
| `orchestrate/config` | Orchestration config | `HubConfig`, `GraphConfig` |
| `orchestrate/hub` | Agent coordination | `Hub` |
| `orchestrate/messaging` | Message structures | `Message`, builders |
| `orchestrate/observability` | Execution tracing | `Observer` |
| `orchestrate/state` | State graphs, checkpoints | `State`, `Graph`, `CheckpointStore` |
| `orchestrate/workflows` | Workflow patterns | `ProcessChain`, `ProcessParallel`, `ProcessConditional` |
| `memory` | Context composition pipeline | `Store`, `Cache`, `Entry`, `NewFileStore`, `NewCache` |
| `tools` | Tool execution and registry | `Handler`, `Result`, `Register`, `Execute`, `List` |
| `session` | Conversation management | `Session`, `NewMemorySession` |
| `kernel` | Agent runtime loop | `Kernel`, `Config`, `Result`, `ToolExecutor`, `WithLogger` |

## Extension Patterns

### Adding a Provider

1. Create new file in `agent/providers/` (e.g., `anthropic.go`)
2. Implement `Provider` interface
3. Register in `agent/providers/registry.go` init function
4. Add tests in `agent/providers/` (co-located)

### Adding an Observer

1. Create new file in `orchestrate/observability/`
2. Implement `Observer` interface
3. Add tests alongside implementation

### Adding a Workflow Pattern

1. Create new file in `orchestrate/workflows/`
2. Follow existing pattern: `Process<Pattern>` function + `<Pattern>Node` integration helper
3. Add corresponding config in `orchestrate/config/`
4. Add tests alongside implementation

## Testing Strategy

- Tests co-located with source (`*_test.go` alongside `.go` files)
- Black-box testing using `_test` package suffix
- Table-driven tests for comprehensive coverage
- HTTP mocking with `httptest.Server` for provider tests
- `agent/mock/` package for testing orchestration without live LLMs
- Integration tests in top-level `tests/`

### Coverage Philosophy

Focus test effort on:
1. All public API methods and exported types
2. Error paths and edge cases
3. Concurrency safety (hub, parallel workflows)
4. State transitions and predicates

See `references/` for detailed architecture documentation per subsystem.
