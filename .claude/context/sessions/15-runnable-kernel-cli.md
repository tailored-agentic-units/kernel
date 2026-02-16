# 15 — Runnable Kernel CLI with Built-in Tools

## Summary

Replaced the `cmd/kernel/main.go` stub with a functional CLI entry point that exercises the full agentic loop against a real LLM. Added three built-in tools (datetime, read_file, list_directory), seed memory for system prompt composition, and structured logging via `slog`. Fixed a ToolCall serialization bug that blocked provider communication.

## Key Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Built-in tools location | `cmd/kernel/` (not `tools/`) | CLI demo tools, not part of the kernel library API |
| Memory seed directory | `cmd/kernel/memory/` | Co-located with config; exercises full memory → system prompt pipeline |
| Unlimited iterations | `maxIterations=0` means run until done | Clean semantic; context cancellation provides safety net |
| Logger interface | `*slog.Logger` via `WithLogger` option | Standard library, zero dependencies, supports future per-subsystem logging |
| ToolCall MarshalJSON | Value receiver, nested format | Round-trip fidelity with UnmarshalJSON; value receiver works in all serialization contexts |
| CLI max-iterations flag | Default -1 (sentinel) | Distinguishes "not provided" from "unlimited" (0) |

## Files Modified

- `cmd/kernel/main.go` — replaced stub with functional CLI
- `cmd/kernel/tools.go` — built-in tool definitions and registration (new)
- `cmd/kernel/memory/identity.md` — seed memory content (new)
- `cmd/kernel/agent.ollama.qwen3.json` — added memory path
- `kernel/kernel.go` — unlimited iterations, `WithLogger`, slog log points
- `core/protocol/message.go` — `ToolCall.MarshalJSON` for nested API format
- `core/protocol/protocol_test.go` — MarshalJSON and round-trip tests
- `kernel/kernel_test.go` — unlimited iterations and WithLogger tests
- `_project/README.md` — kernel status updated to Complete
- `_project/objective.md` — issue #15 status updated to Closed
- `.claude/CLAUDE.md` — fixed kernel run command
- `.claude/skills/kernel-dev/SKILL.md` — added WithLogger to kernel exports
- `README.md` — fixed kernel run command

## Patterns Established

- **Remediation convention**: Implementation guides gain a Remediation section (R1, R2, ...) between final step and Validation Criteria for blockers discovered during execution
- **slog logger pattern**: Kernel accepts `*slog.Logger` via `WithLogger`; discard logger default; subsystems will follow this pattern (tracked in Objective #4)
- **CLI flag sentinel**: Use -1 default for optional numeric overrides where 0 has semantic meaning

## Validation Results

- `go vet ./...` — clean
- `go test ./...` — all pass
- `go mod tidy` — no changes
- Coverage: kernel.go Run 100%, message.go MarshalJSON 100%, kernel package 94.2%, protocol package 92.3%
- Manual: CLI runs against Ollama/Qwen3 with tool calls, memory loading, and verbose logging
