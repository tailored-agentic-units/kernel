# agent

LLM communication layer for the TAU kernel — agent interface, HTTP client, provider abstractions (Ollama, Azure), request construction, and response parsing.

## Packages

### agent (root)

High-level Agent interface with protocol methods.

- `Agent` interface: `Chat`, `Vision`, `Tools`, `Embed`, `Audio`, `ChatStream`, `VisionStream`
- `New(config)` constructor with provider registration and model resolution

### client

HTTP client with retry logic, exponential backoff, and connection pooling.

- Automatic retry for transient failures (429, 502, 503, 504, network errors)
- Exponential backoff with optional jitter
- Thread-safe connection pooling

### providers

Provider implementations for LLM platforms.

- `Ollama` - Local model serving via Ollama API
- `Azure` - Azure AI Foundry with API Key and Entra ID authentication
- `Provider` interface and `Registry` for extensibility

### request

Protocol-specific request construction.

- `ChatRequest`, `VisionRequest`, `ToolsRequest`, `EmbeddingsRequest`, `AudioRequest`
- OpenAI-format tool wrapping, vision image embedding

### mock

Mock implementations for testing agent-dependent code without live LLM services.

- `MockAgent` - Complete agent interface implementation
- `MockClient` - Transport client with configurable responses
- `MockProvider` - Provider with endpoint mapping
- Helper constructors: `NewSimpleChatAgent`, `NewStreamingChatAgent`, `NewToolsAgent`, `NewFailingAgent`
