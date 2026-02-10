# session

Conversation history management for the TAU kernel runtime loop.

Provides the `Session` interface and an in-memory implementation. Messages use `protocol.Message` natively, including tool call support for multi-turn agentic conversations.

## Future

- Token counting and context window tracking
- Compaction strategies
- Persistent session storage (via ConnectRPC extension boundary)
