# 23 - Streaming tools protocol

## Problem Context

The kernel loop needs streaming support for tool calls before the HTTP interface is established (Objective #2). The Agent interface has `ChatStream` and `VisionStream` but no `ToolsStream`. While streaming routing infrastructure exists (`ParseToolsStreamChunk`, `Protocol.Tools.SupportsStreaming()`, `client.ExecuteStream()`), the `StreamingChunk.Delta` struct lacks a `ToolCalls` field — tool call data from streaming responses would be silently dropped.

Additionally, `ToolCall` uses a flat canonical format with custom `MarshalJSON`/`UnmarshalJSON` to bridge the nested LLM API format. This indirection introduced a streaming bug (continuation chunks with `function.arguments` but no `function.name` were silently dropped) and adds serialization complexity that deviates from the external standard. Adopting the native LLM API format eliminates both issues.

## Architecture Approach

Address gaps bottom-up per the project's design principle: fix core types at the lowest affected dependency level, then add the agent interface method.

- **ToolCall adopts native LLM API format** — eliminates custom JSON methods, aligns with OpenAI/Ollama wire format, streaming continuation chunks work naturally.
- **`ToolFunction` named type** — avoids verbose anonymous struct literals.
- **`NewToolCall` constructor** — encapsulates `Type: "function"` default, keeps call sites clean.
- **`ToolCallRecord` embeds `protocol.ToolCall`** — eliminates field duplication between the wire type and the execution record.
- **Accumulation of partial tool call arguments** into complete `ToolCall` objects is a consumer concern (kernel loop, issue #26).

## Implementation

### Step 1: Refactor ToolCall to native LLM API format

**File:** `core/protocol/message.go`

Remove the `encoding/json` import.

Replace the `ToolCall` struct (lines 19-23), `MarshalJSON` method (lines 27-46), and `UnmarshalJSON` method (lines 51-72) with:

```go
type ToolFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

func NewToolCall(id, name, arguments string) ToolCall {
	return ToolCall{
		ID:   id,
		Type: "function",
		Function: ToolFunction{
			Name:      name,
			Arguments: arguments,
		},
	}
}
```

### Step 2: Embed ToolCall in ToolCallRecord and update kernel loop

**File:** `kernel/kernel.go`

Replace `ToolCallRecord` (lines 33-40) with:

```go
type ToolCallRecord struct {
	protocol.ToolCall
	Iteration int    // Loop cycle in which the call occurred.
	Result    string // Tool execution output.
	IsError   bool   // Whether execution returned an error.
}
```

In the tool call processing loop (lines 193-228), update record construction (lines 196-201) from:

```go
record := ToolCallRecord{
	Iteration: iteration + 1,
	ID:        tc.ID,
	Name:      tc.Name,
	Arguments: tc.Arguments,
}
```

to:

```go
record := ToolCallRecord{
	ToolCall:  tc,
	Iteration: iteration + 1,
}
```

Update field access in the loop:
- Line 194: `tc.Name` → `tc.Function.Name`
- Line 205: `tc.Name` → `tc.Function.Name`
- Line 206: `json.RawMessage(tc.Arguments)` → `json.RawMessage(tc.Function.Arguments)`

### Step 3: Update CLI output for embedded ToolCallRecord

**File:** `cmd/kernel/main.go`

Line 78 — update ToolCallRecord field access:

```go
fmt.Printf("  [%d] %s(%s)\n", i+1, tc.Function.Name, tc.Function.Arguments)
```

**File:** `cmd/prompt-agent/main.go`

Line 186 — update protocol.ToolCall field access:

```go
fmt.Printf("  - %s(%s)\n", toolCall.Function.Name, toolCall.Function.Arguments)
```

### Step 4: Extend StreamingChunk with ToolCalls support

**File:** `core/response/streaming.go`

Add `protocol` import:

```go
import (
	"encoding/json"
	"fmt"

	"github.com/tailored-agentic-units/kernel/core/protocol"
)
```

Add `ToolCalls` field to the anonymous Delta struct inside `StreamingChunk` (lines 18-21):

```go
Delta struct {
	Role      string              `json:"role,omitempty"`
	Content   string              `json:"content,omitempty"`
	ToolCalls []protocol.ToolCall `json:"tool_calls,omitempty"`
} `json:"delta"`
```

Add `ToolCalls()` accessor method after the existing `Content()` method:

```go
func (c *StreamingChunk) ToolCalls() []protocol.ToolCall {
	if len(c.Choices) > 0 {
		return c.Choices[0].Delta.ToolCalls
	}
	return nil
}
```

Update `ParseToolsStreamChunk` comment to replace "Tools protocol uses the same streaming format as chat." with "Tools streaming chunks include tool call deltas in the Delta field."

### Step 5: Add ToolsStream to Agent interface and implementation

**File:** `agent/agent.go`

Add to the `Agent` interface after the `Tools` method (after line 61):

```go
ToolsStream(ctx context.Context, prompt []protocol.Message, tools []protocol.Tool, opts ...map[string]any) (<-chan *response.StreamingChunk, error)
```

Add implementation after the `Tools` method (after line 237):

```go
func (a *agent) ToolsStream(ctx context.Context, prompt []protocol.Message, tools []protocol.Tool, opts ...map[string]any) (<-chan *response.StreamingChunk, error) {
	messages := a.initMessages(prompt)
	options := a.mergeOptions(protocol.Tools, opts...)
	options["stream"] = true

	req := request.NewTools(a.provider, a.model, messages, tools, options)

	return a.client.ExecuteStream(ctx, req)
}
```

### Step 6: Add ToolsStream to MockAgent

**File:** `agent/mock/agent.go`

Add after the `Tools` method (after line 204):

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

## Validation Criteria

- [ ] `ToolCall` uses native LLM API format with `ToolFunction` named type
- [ ] No custom `MarshalJSON`/`UnmarshalJSON` on `ToolCall`
- [ ] `NewToolCall` constructor produces correctly-typed instances
- [ ] `ToolCallRecord` embeds `protocol.ToolCall`, no duplicate fields
- [ ] All runtime field access sites updated (`.Function.Name`, `.Function.Arguments`)
- [ ] `StreamingChunk.Delta` includes `ToolCalls` field
- [ ] `StreamingChunk.ToolCalls()` accessor returns tool calls from first choice
- [ ] `ToolsStream` method added to `Agent` interface
- [ ] `ToolsStream` implementation follows ChatStream/VisionStream pattern
- [ ] Mock agent updated with `ToolsStream` support
- [ ] All existing tests pass
- [ ] `go vet ./...` passes
