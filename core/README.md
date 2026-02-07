# core

Foundational type vocabulary for the TAU kernel — protocol constants, response types, configuration structures, and model types.

## Packages

### protocol

Protocol constants and message types used across the kernel.

- `Protocol` enum: `Chat`, `Vision`, `Tools`, `Embeddings`, `Audio`
- `Message` type: role-tagged messages with optional metadata

### response

Response types returned from LLM operations, with parsing and streaming support.

- `ChatResponse` - Text completion responses with token usage
- `ToolsResponse` - Tool call responses with structured arguments
- `EmbeddingsResponse` - Vector embedding responses
- `AudioResponse` - Audio generation responses
- Streaming support via `StreamHandler`

### config

Configuration types with human-readable durations and clean JSON serialization.

- `AgentConfig` - Top-level agent configuration (name, system prompt, client, provider, model)
- `ProviderConfig` - Provider platform settings (name, base URL, options)
- `ModelConfig` - Model settings with protocol-specific capabilities
- `ClientConfig` - HTTP client settings (timeout, retry, connection pool)
- `Duration` - Human-readable duration strings ("24s", "1m")

### model

Model runtime type bridging configuration to execution.

- `Model` - Runtime representation combining config and protocol support
