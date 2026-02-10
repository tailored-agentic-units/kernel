# Changelog

## Current

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
