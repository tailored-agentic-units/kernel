# Objective: Kernel Core Loop

**Issue:** #1
**Phase:** Phase 1 — Foundation (v0.1.0)

## Scope

Implement the agentic processing loop — the core observe/think/act/repeat cycle that composes the kernel subsystems into a functioning runtime.

## Sub-Issues

| # | Title | Status |
|---|-------|--------|
| 11 | Session interface and in-memory implementation | Closed |
| 12 | Tool registry interface and execution | Closed |
| 13 | Memory store interface and filesystem implementation | Closed |
| 14 | Kernel runtime loop | Closed |
| 15 | Runnable kernel CLI with built-in tools | Open |

## Architecture Decisions

- Session, tools, and memory define interfaces first — kernel runtime wires them together
- Foundation subsystems (11-13) can be implemented in parallel
- Kernel runtime (#14) depends on all three interface definitions
- Runnable kernel CLI (#15) provides runtime validation against a real LLM — supersedes mock-based integration tests per the agent-lab principle
