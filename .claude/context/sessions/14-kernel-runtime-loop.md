# 14 - Kernel Runtime Loop

## Summary

Implemented the single-agent runtime loop composing agent, tools, session, and memory into the observe/think/act/repeat agentic cycle. The kernel initializes from configuration via `New(*Config, ...Option)` following the cold start pattern — config drives subsystem creation, functional options enable test overrides. `Run()` executes the loop: add user prompt to session, build system content from memory, call `agent.Tools()`, execute any tool calls, repeat until the agent produces a final response or iterations are exhausted.

Additionally evolved the Agent interface (conversation methods accept `[]protocol.Message` instead of `prompt string`), consolidated `response.ToolCall` into `protocol.ToolCall` with a custom `UnmarshalJSON` for transparent nested-to-flat conversion, added subsystem configurations for session and memory, and introduced `protocol.InitMessages` as a convenience wrapper.

## Key Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Agent method signatures | `prompt []protocol.Message` | Enables multi-turn conversations; kernel passes full session history |
| ToolCall consolidation | Single `protocol.ToolCall` with custom `UnmarshalJSON` | Eliminates duplicate types; nested LLM format handled transparently |
| Kernel initialization | Config-driven cold start with functional options | Follows agent-lab pattern; callers never construct dependencies manually |
| ToolExecutor interface | Kernel-local interface wrapping global `tools` package | Testability without changing tools package API |
| Memory error handling | Propagate errors from `buildSystemContent` | Silent failures hide real problems; callers should know |
| `initMessages` naming | Kept original name vs. `prependSystemPrompt` | Developer preference; clearer intent in context |

## Files Modified

### New files
- `kernel/kernel.go` — Kernel struct, types, ToolExecutor, options, New(), Run()
- `kernel/config.go` — Config, DefaultConfig, Merge, LoadConfig
- `kernel/errors.go` — ErrMaxIterations
- `kernel/kernel_test.go` — 14 tests (93.8% coverage)
- `kernel/config_test.go` — 5 tests
- `session/config.go` — Config, DefaultConfig, Merge, New
- `session/config_test.go` — 3 tests
- `memory/config.go` — Config, DefaultConfig, Merge, NewStore
- `memory/config_test.go` — 5 tests

### Modified files
- `core/protocol/message.go` — ToolCall with UnmarshalJSON, InitMessages
- `core/protocol/protocol_test.go` — 6 new tests (UnmarshalJSON + InitMessages)
- `core/response/tools.go` — Uses protocol.ToolCall, removed ToolCall/ToolCallFunction types
- `core/response/response_test.go` — Updated for flat ToolCall
- `agent/agent.go` — Interface methods accept []protocol.Message
- `agent/agent_test.go` — Updated call sites
- `agent/mock/agent.go` — Updated method signatures
- `agent/mock/agent_test.go` — Updated call sites
- `agent/mock/helpers.go` — Uses protocol.ToolCall
- `agent/client/client_test.go` — Updated for protocol.ToolCall + InitMessages
- `agent/providers/base_test.go` — InitMessages adoption
- `agent/providers/ollama_test.go` — InitMessages adoption
- `agent/providers/azure_test.go` — InitMessages adoption
- `cmd/prompt-agent/main.go` — InitMessages adoption
- `orchestrate/examples/` — InitMessages adoption across 7 example files

### Infrastructure files
- `_project/README.md` — Removed Known Gaps, updated subsystem statuses
- `README.md` — Updated subsystem descriptions
- `.claude/CLAUDE.md` — Updated structure, hierarchy, added doc review directive
- `.claude/skills/kernel-dev/SKILL.md` — Added kernel package, updated hierarchy

## Patterns Established

- **Config-driven cold start**: `kernel.New(*Config, ...Option)` — config creates all subsystems, options override for tests
- **`protocol.InitMessages`**: Convenience wrapper replacing verbose `[]protocol.Message{protocol.NewMessage(...)}` pattern
- **Custom UnmarshalJSON on protocol types**: Transparent format conversion at the deserialization boundary
- **Subsystem Config pattern**: Each subsystem owns `Config`, `DefaultConfig()`, `Merge()`, and a config-driven constructor

## Validation Results

- `go vet ./...` — pass
- `go test ./...` — all pass (0 failures)
- `go mod tidy` — no changes
- Kernel package coverage: 93.8% (Run: 100%, config: 100%)
- Protocol package coverage: 92.0%
- Session package coverage: 100%
