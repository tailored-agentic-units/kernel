# Plan: #23 — Streaming tools protocol

## Context

The kernel loop needs streaming support for tool calls before the HTTP interface is established (Objective #2). The Agent interface has `ChatStream` and `VisionStream` but no `ToolsStream`. While the streaming routing infrastructure exists (`ParseToolsStreamChunk`, `Protocol.Tools.SupportsStreaming()`, `client.ExecuteStream()`), the `StreamingChunk.Delta` struct lacks a `ToolCalls` field — meaning tool call data from streaming responses would be silently dropped.

Additionally, `ToolCall.UnmarshalJSON` only recognizes the nested format when `function.name` is present. In OpenAI-compatible streaming, continuation chunks carry `function.arguments` without `function.name`, causing those fragments to fall through to the flat format and lose the arguments data. This needs fixing at the core level before `ToolsStream` can faithfully capture tool call deltas.

## Approach

Address gaps bottom-up per the project's design principle: fix core types first, then add the agent interface method. Accumulation of partial tool call arguments into complete `ToolCall` objects is a consumer concern (kernel loop, issue #26) — not part of this issue.

## Implementation

### Step 1: Fix ToolCall.UnmarshalJSON for streaming continuation chunks (`core/protocol/message.go`)

The current condition `nested.Function.Name != ""` skips the nested path for continuation chunks that have `function.arguments` but no `function.name`. Fix by checking for any nested data:

```go
if nested.Function.Name != "" || nested.Function.Arguments != "" {
```

### Step 2: Extend StreamingChunk.Delta with ToolCalls (`core/response/streaming.go`)

Add `ToolCalls` field to the anonymous Delta struct inside StreamingChunk:

```go
Delta struct {
    Role      string              `json:"role,omitempty"`
    Content   string              `json:"content,omitempty"`
    ToolCalls []protocol.ToolCall `json:"tool_calls,omitempty"`
} `json:"delta"`
```

This requires adding `"github.com/tailored-agentic-units/kernel/core/protocol"` to imports.

Add a `ToolCalls()` accessor method (mirrors the `Content()` pattern):

```go
func (c *StreamingChunk) ToolCalls() []protocol.ToolCall {
    if len(c.Choices) > 0 {
        return c.Choices[0].Delta.ToolCalls
    }
    return nil
}
```

Update `ParseToolsStreamChunk` comment to remove the incorrect "same format as chat" statement.

### Step 3: Add ToolsStream to Agent interface (`agent/agent.go`)

Add method signature after the `Tools` method (~line 61):

```go
ToolsStream(ctx context.Context, prompt []protocol.Message, tools []protocol.Tool, opts ...map[string]any) (<-chan *response.StreamingChunk, error)
```

Add implementation after the `Tools` method (~line 237), following the ChatStream pattern:

```go
func (a *agent) ToolsStream(ctx context.Context, prompt []protocol.Message, tools []protocol.Tool, opts ...map[string]any) (<-chan *response.StreamingChunk, error) {
    messages := a.initMessages(prompt)
    options := a.mergeOptions(protocol.Tools, opts...)
    options["stream"] = true

    req := request.NewTools(a.provider, a.model, messages, tools, options)

    return a.client.ExecuteStream(ctx, req)
}
```

### Step 4: Add ToolsStream to MockAgent (`agent/mock/agent.go`)

Add method mirroring ChatStream/VisionStream pattern:

```go
func (m *MockAgent) ToolsStream(ctx context.Context, prompt []protocol.Message, tools []protocol.Tool, opts ...map[string]any) (<-chan *response.StreamingChunk, error) {
    if m.streamError != nil {
        return nil, m.streamError
    }

    ch := make(chan *response.StreamingChunk, len(m.streamChunks))
    for i := range m.streamChunks {
        ch <- &m.streamChunks[i]
    }
    close(ch)

    return ch, nil
}
```

### Step 5: Add NewStreamingToolsAgent helper (`agent/mock/helpers.go`)

Add helper following the NewStreamingChatAgent pattern, producing chunks with tool call deltas.

Update the struct literal in `NewStreamingChatAgent` to include the new `ToolCalls` field in the Delta anonymous struct (required for type compatibility).

## Files Modified

- `core/protocol/message.go` — Fix UnmarshalJSON for streaming continuation chunks
- `core/response/streaming.go` — Extend Delta struct, add ToolCalls() accessor
- `agent/agent.go` — Interface + implementation
- `agent/mock/agent.go` — Mock implementation
- `agent/mock/helpers.go` — NewStreamingToolsAgent helper, update existing struct literals

## Verification

- `go vet ./...`
- `go test ./...`
- New tests for streaming tool call chunks in `core/response/`
- New tests for ToolsStream in agent and mock packages
