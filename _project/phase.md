# Phase 1 — Foundation

**Version target:** v0.1.0

## Scope

Implement the foundation subsystems that have no dependencies on each other: memory, tools, and session. These three subsystems plus the kernel runtime loop and integration tests compose the first functional kernel.

## Objectives

| # | Objective | Status |
|---|-----------|--------|
| 1 | Kernel Core Loop | In Progress |
| 2 | Kernel ConnectRPC Interface | Planned |
| 3 | Skills and MCP Integration | Planned |
| 4 | Local Development Mode | Planned |

## Constraints

- memory, tools, session can be built in parallel (no cross-dependencies)
- kernel runtime depends on all three foundation subsystems
- Single module — all subsystems share one version tag
