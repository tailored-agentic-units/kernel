# kernel — Project Vision

## Vision

A self-contained agent runtime that can run locally, in containers, or as a distributed service — with external infrastructure connecting exclusively through the ConnectRPC interface.

## Architecture

The kernel consolidates all TAU subsystems into a single Go module:

- **core** — Foundational type vocabulary (protocols, responses, config, models)
- **agent** — LLM communication layer (agent interface, client, providers)
- **tools** — Tool execution interface, registry, and permissions
- **session** — Conversation history and context management
- **memory** — Filesystem-based persistent memory
- **skills** — Progressive disclosure skill system
- **mcp** — MCP client with transport abstraction
- **orchestrate** — Multi-agent coordination and workflow patterns
- **kernel** — Agent runtime: closed-loop processing with ConnectRPC interface

Dependencies flow in one direction: core → capabilities → composition.

## Runtime Boundary

- The kernel is a closed-loop I/O system with zero extension awareness
- The ConnectRPC interface (`tau.kernel.v1.KernelService`) is the sole extensibility boundary
- External services (persistence, IAM, containers, observability, UI) connect through the interface — the kernel never reaches out to them
- The same kernel serves embedded, desktop, server, or cloud deployments — only the extensions change

## Principles

- Each subsystem has a single clear responsibility
- Dependencies flow in one direction: core → capabilities → composition
- The kernel has zero awareness of what connects to its interface
- Local development works without containers or external infrastructure
- Single module, single version — no dependency cascade
