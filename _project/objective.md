# Objective: Kernel Core Loop

**Issue:** #1
**Phase:** Phase 1 — Foundation (v0.1.0)

## Scope

Implement the agentic processing loop — the core observe/think/act/repeat cycle that composes the kernel subsystems into a functioning runtime.

## Sub-Issues

| # | Title | Status |
|---|-------|--------|
| 11 | Session interface and in-memory implementation | Open |
| 12 | Tool registry interface and execution | Open |
| 13 | Memory store interface and filesystem implementation | Open |
| 14 | Kernel runtime loop | Open |
| 15 | Kernel integration tests | Open |

## Architecture Decisions

- Session, tools, and memory define interfaces first — kernel runtime wires them together
- Foundation subsystems (11-13) can be implemented in parallel
- Kernel runtime (#14) depends on all three interface definitions
- Integration tests (#15) validate the composed system
