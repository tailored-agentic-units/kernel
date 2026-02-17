# Session: Streaming Tools Protocol

**Issue:** #23
**Branch:** `23-streaming-tools-protocol`
**Objective:** #2 — Kernel Interface

## Summary

Added `ToolsStream` to the Agent interface and refactored `ToolCall` from a flat canonical format with custom JSON methods to the native LLM API format, aligning with the OpenAI-compatible wire format used by both Ollama and Azure providers.

## Changes

### Core type refactoring (`core/protocol/message.go`)

- Replaced flat `ToolCall{ID, Name, Arguments}` with native format: `ToolCall{ID, Type, Function}` where `Function` is the new `ToolFunction{Name, Arguments}` named type
- Removed custom `MarshalJSON`/`UnmarshalJSON` methods — standard `encoding/json` handles the native format directly
- Added `NewToolCall(id, name, arguments)` constructor that sets `Type: "function"` automatically

### Streaming tool call support (`core/response/streaming.go`)

- Extended `StreamingChunk.Delta` with `ToolCalls []protocol.ToolCall` field
- Added `ToolCalls()` accessor method (mirrors the `Content()` pattern)
- Updated `ParseToolsStreamChunk` comment

### Agent interface (`agent/agent.go`)

- Added `ToolsStream` method to the `Agent` interface
- Added implementation following the `ChatStream`/`VisionStream` pattern

### Mock agent (`agent/mock/agent.go`, `agent/mock/helpers.go`)

- Added `ToolsStream` method to `MockAgent`
- Added `NewStreamingToolsAgent` helper for streaming tool call test scenarios
- Updated `NewStreamingChatAgent` anonymous struct to include `ToolCalls` field

### Kernel runtime (`kernel/kernel.go`)

- `ToolCallRecord` now embeds `protocol.ToolCall` instead of duplicating fields
- Updated field access throughout: `tc.Function.Name`, `tc.Function.Arguments`

### CLI updates (`cmd/kernel/main.go`, `cmd/prompt-agent/main.go`)

- Updated ToolCall field access in output formatting

### Test updates (16 files modified)

- Updated all ToolCall struct literals across test files to use `NewToolCall` or native format
- Updated all `.Name`/`.Arguments` field access to `.Function.Name`/`.Function.Arguments`
- Rewrote marshal/unmarshal tests for native format (removed flat format tests)
- Added new tests: `TestNewToolCall`, streaming chunk tool call tests, `ToolsStream` tests

## Design Decisions

- **Native LLM API format over flat canonical**: Eliminates custom JSON methods, fixes streaming continuation bug, aligns with external standards
- **`ToolFunction` named type**: Avoids verbose anonymous struct literals in call sites
- **`ToolCallRecord` embedding**: Eliminates field duplication between wire type and execution record
- **Accumulation deferred to #26**: Reassembling partial streaming arguments is a consumer concern for the kernel loop
- **CLI streaming deferred to #26**: `cmd/prompt-agent` and `cmd/kernel` streaming tool support deferred

## Scope Boundary

Tool call accumulation (reassembling partial streaming arguments into complete `ToolCall` objects) is explicitly not part of this issue — it belongs to #26 (Multi-session kernel) where the streaming-first kernel loop is implemented.
