# Changelog

## v0.1.0-dev.1.15

### kernel

- Replace `cmd/kernel/main.go` stub with functional CLI entry point (#15)
- Add built-in tools: `datetime`, `read_file`, `list_directory` (#15)
- Add seed memory directory at `cmd/kernel/memory/` for system prompt composition (#15)
- Support unlimited iterations when `maxIterations` is 0 (#15)
- Add `WithLogger` option with `*slog.Logger` for runtime observability (#15)
- Add structured log points in `Run` and `buildSystemContent` (#15)

### core

- Add `ToolCall.MarshalJSON` for nested LLM API format round-trip fidelity (#15)

## v0.1.0-dev.1.14

### kernel

- Add `Kernel` runtime loop with config-driven initialization and observe/think/act/repeat cycle (#14)
- Add `Config`, `DefaultConfig`, `Merge`, `LoadConfig` for kernel configuration (#14)
- Add `Result`, `ToolCallRecord`, `ToolExecutor` interface, and functional options (#14)
- Add `ErrMaxIterations` for loop budget exhaustion (#14)

### core

- Consolidate `response.ToolCall` into `protocol.ToolCall` with custom `UnmarshalJSON` for nested LLM format (#14)
- Add `protocol.InitMessages` convenience wrapper for single-prompt message initialization (#14)
- Evolve `Agent` interface: conversation methods accept `[]protocol.Message` instead of `prompt string` (#14)

### session

- Add `Config`, `DefaultConfig`, `Merge`, `New` for config-driven session creation (#14)

### memory

- Add `Config`, `DefaultConfig`, `Merge`, `NewStore` for config-driven store creation (#14)

## v0.1.0-dev.1.13

### memory

- Add `Store` interface for pluggable persistence with `List`, `Load`, `Save`, `Delete` (#13)
- Add `FileStore` filesystem implementation with atomic writes and hidden file filtering (#13)
- Add `Cache` session-scoped context cache with progressive loading via `Bootstrap`, `Resolve`, `Flush` (#13)
- Add `Entry` type and namespace constants (`memory`, `skills`, `agents`) (#13)
- Consolidate `skills/` skeleton into memory package as a namespace convention (#13)

## v0.1.0-dev.1.12

### core

- Add `protocol.Tool` as canonical tool definition type, replacing `agent.Tool` and `providers.ToolDefinition` (#12)

### tools

- Add global tool registry with `Register`, `Replace`, `Get`, `List`, `Execute` (#12)

## v0.1.0-dev.1.11

### core

- Add `protocol.Role` typed string enum with `RoleSystem`, `RoleUser`, `RoleAssistant`, `RoleTool` constants (#11)
- Add `protocol.ToolCall` struct for tool invocations in conversation history (#11)
- Add `ToolCallID` and `ToolCalls` fields to `protocol.Message` for tool-calling conversations (#11)

### session

- Add `Session` interface with `ID()`, `AddMessage()`, `Messages()`, `Clear()` (#11)
- Add in-memory implementation with concurrent-safe access and defensive copies (#11)

## v0.0.1 - 2026-02-07

Initial kernel repository — consolidated tau-core, tau-agent, tau-orchestrate, and tau-runtime into a single module.

### core

Foundational type vocabulary for the TAU kernel.

- `core/protocol` - Protocol constants (Chat, Vision, Tools, Embeddings, Audio) and Message types
- `core/response` - Response types, parsing, and streaming support
- `core/config` - Configuration types with human-readable durations
- `core/model` - Model runtime type bridging config to execution

### agent

LLM communication layer, extracted from tau-core to establish clean dependency boundaries.

- `agent` - High-level Agent interface with Chat, Vision, Tools, Embed, Audio methods
- `agent/client` - HTTP client with retry logic and exponential backoff
- `agent/mock` - Mock implementations for testing (MockAgent, MockClient, MockProvider)
- `agent/providers` - Provider implementations (Ollama, Azure AI Foundry)
- `agent/request` - Protocol-specific request types

Features:
- Multi-protocol support: Chat, Vision, Tools, Embeddings, Audio
- Multi-provider support: Ollama, Azure (API Key and Entra ID auth)
- Streaming responses for Chat, Vision
- Configuration option merging (model defaults + runtime overrides)
- Thread-safe connection pooling
- Retry with exponential backoff and jitter
- Comprehensive mock implementations for testing

### orchestrate

Multi-agent coordination and workflow orchestration.

- `orchestrate/config` - Configuration structures for all orchestration primitives
- `orchestrate/hub` - Multi-hub agent coordination with message routing
- `orchestrate/messaging` - Message structures, builders, and inter-agent communication
- `orchestrate/observability` - Observer pattern with NoOp, Slog, and Multi observers
- `orchestrate/state` - State graph execution with checkpointing and persistence
- `orchestrate/workflows` - Sequential, parallel, and conditional workflow patterns

Features:
- Multi-hub agent coordination with cross-hub communication
- Four communication patterns: Send, Request/Response, Broadcast, Pub/Sub
- LangGraph-inspired state graph execution with transition predicates
- Sequential chains with state accumulation
- Parallel execution with worker pools and order preservation
- Conditional routing with predicate-based handler selection
- Checkpointing for workflow persistence and recovery
- Composable integration helpers (ChainNode, ParallelNode, ConditionalNode)
- Configurable observability (NoOp, Slog, Multi observers)
- State secrets for sensitive data excluded from serialization

### kernel

- ConnectRPC service definition: `tau.kernel.v1` / `KernelService`
- RPCs: CreateSession, Run (streaming), InjectContext, GetSession
- Proto codegen with buf v2 (protoc-gen-go + protoc-gen-connect-go)
