# kernel — Project Identity

## Vision

A self-contained agent runtime that can run locally, in containers, or as a distributed service — with external infrastructure connecting exclusively through the ConnectRPC interface.

Following Anthropic's core principle: **simple, composable patterns over complex frameworks**. Start with the minimum viable abstraction at each layer. Only increase complexity when it demonstrably improves outcomes. Each subsystem should be independently useful, testable, and comprehensible.

## Architecture

The kernel consolidates all TAU subsystems into a single Go module:

- **core** — Foundational type vocabulary (protocols, responses, config, models)
- **agent** — LLM communication layer (agent interface, client, providers)
- **tools** — Tool execution interface, registry, and permissions
- **session** — Conversation history and context management
- **memory** — Unified context composition: persistent memory, skills, agent profiles
- **mcp** — MCP client with transport abstraction
- **observability** — Event-based observability with OTel-aligned severity levels
- **orchestrate** — Multi-agent coordination and workflow patterns
- **kernel** — Agent runtime: closed-loop processing with ConnectRPC interface

Dependencies flow in one direction: core → capabilities → composition.

## Runtime Boundary

The kernel is a closed-loop I/O system with zero extension awareness. The ConnectRPC interface (`tau.kernel.v1.KernelService`) is the sole extensibility boundary. External services connect through it — the kernel never reaches out to them. The same kernel binary serves embedded, desktop, server, or cloud deployments.

Extension ecosystem (external services connecting through the interface):

- **Persistence** — session state storage, memory file management
- **IAM** — authentication, authorization
- **Container/Sandbox** — execution environment management
- **MCP Gateway** — proxies MCP servers for external tool access
- **Observability** — metrics, tracing, logging export
- **UI** — web interface, CLI, or other user-facing interfaces

## Subsystem Topology

| Subsystem | Domain | Depends On | Status |
|-----------|--------|------------|--------|
| **core** | Foundational types: Protocol, Message, Response, Config, Model | uuid | Complete |
| **agent** | LLM client: Agent, Client, Provider, Request, Mock, Registry | core | Complete |
| **observability** | Event-based observability: Observer, Event, Level (OTel-aligned), SlogObserver, registry | *(none)* | Complete |
| **orchestrate** | Coordination: Hub, State, Workflows, Checkpoint | agent, observability | Complete |
| **memory** | Unified context composition: Store interface, FileStore, Cache. Namespaces: `memory/`, `skills/`, `agents/` | *(none)* | Complete |
| **tools** | Tool system: global registry with Register, Execute, List | core | Complete |
| **session** | Conversation management: Session interface, in-memory implementation | core | Complete |
| **mcp** | MCP client: transport abstraction, tool discovery, stdio/SSE | tools | Skeleton |
| **kernel** | Agent runtime: agentic loop, config-driven initialization, observer integration | all above | Complete |

## Dependency Hierarchy

- **kernel** — all subsystems below
  - **orchestrate** — agent, observability
    - **agent** — core
  - **mcp** — tools
    - **tools** — core
  - **observability** — *(no internal dependencies)*
  - **memory** — *(no internal dependencies)*
  - **session** — core
- **core** — *(external: uuid)*

Key properties:

- **Acyclic**: No circular dependencies at any level
- **Shallow**: Maximum depth of 3 (kernel → mcp → tools → core)
- **Independent foundations**: observability and memory have zero internal dependencies; tools and session depend only on core types
- **Clean separation**: Each subsystem owns a single domain with no overlap
- **Enforced by Go**: Import rules and the type system enforce boundaries within the single module

## Build Order

### Phase 0 — Complete

**core, agent, orchestrate** — Migrated into the kernel monorepo with co-located tests, import paths updated, all tests passing.

### Phase 1 — Foundation (independent subsystems)

These subsystems can be built in parallel (no cross-dependencies):

1. **memory** — Filesystem-based persistent memory with zero internal dependencies. Bootstrap loading, working memory, structured notes.
2. **tools** — Tool execution interface, registry, permissions, built-in tools. Depends on core for `protocol.Tool` type only.
3. **session** — Conversation history management with token tracking and compaction. Depends on core for `protocol.Message` type only.

### Phase 2 — Integration (builds on Phase 1)

4. **mcp** — MCP client with transport abstraction. Depends on tools. Can begin after tools is complete.

### Phase 3 — Composition

5. **kernel** — Agentic loop composing all subsystems. Depends on everything above.

```
Phase 0:  [core + agent + orchestrate]          (complete)
              |
              v
Phase 1:  [memory]  [tools]  [session]          (parallel, complete)
                       |
                       v
Phase 2:             [mcp]
                       |
                       v
Phase 3:           [kernel]                     (runtime loop complete)
```

## Model Compatibility

The agent runtime must work across platforms available at all classification levels. Both **tool calling** and **reasoning** are critical — tool calling drives the agentic loop, while reasoning enables effective tool selection, result interpretation, and multi-step planning.

### Ollama (Self-Hosted)

Models supporting both tool calling and reasoning:

| Model | Parameters | Tool Calling | Reasoning | Notes |
|-------|-----------|-------------|-----------|-------|
| **Qwen 3** | 0.6B - 235B | Strong | Strong | Hybrid think/no_think modes. Recommended starting point for local development |
| **GPT-OSS** (OpenAI) | 20B, 120B | Strong | Strong | Adjustable reasoning effort. 20B runs on 16GB devices |
| **Kimi K2 Thinking** (Moonshot AI) | MoE (~32B active) | Excellent | Excellent | Best-in-class agentic capability; supports 200-300 sequential tool calls |
| **DeepSeek-V3.1** | 671B MoE | Strong | Strong | Hybrid thinking/non-thinking modes. Requires significant hardware |
| **Devstral Small 2** (Mistral) | 24B | Strong | Limited | 68% on SWE-Bench Verified. Strong for software engineering agents |
| **Nemotron-3-Nano** (NVIDIA) | 30B | Strong | Strong | Efficient agentic model |

Models with strong tool calling but limited reasoning:

| Model | Parameters | Tool Calling | Reasoning | Notes |
|-------|-----------|-------------|-----------|-------|
| **Qwen3-Coder** | 30B, 480B | Excellent | Limited | Purpose-built for agentic coding workflows |
| **Ministral-3** (Mistral) | 3B, 8B, 14B | Strong | Limited | Edge deployment. Multi-modal, multi-lingual |
| **Granite4** (IBM) | 350M, 1B, 3B | Good | Limited | Ultra-small enterprise models for edge/IoT |

Models with strong reasoning but limited tool calling:

| Model | Parameters | Tool Calling | Reasoning | Notes |
|-------|-----------|-------------|-----------|-------|
| **DeepSeek-R1** | 1.5B - 671B | Poor | Excellent | Tool calling is unreliable |
| **Magistral** (Mistral) | 24B | Limited | Strong | Transparent reasoning traces |
| **Phi-4-Reasoning** (Microsoft) | 14B | Limited | Strong | Rivals much larger models |

**Recommendation**: Qwen 3 (8B or larger) or GPT-OSS 20B for development and testing.

### Azure AI Foundry (Cloud-Hosted)

| Model Family | Tool Calling | Reasoning | Notes |
|-------------|-------------|-----------|-------|
| **GPT-5.2** | Excellent | Excellent | Freeform tool calling. 272K context |
| **Claude Opus 4.5 / Sonnet 4.5** | Excellent | Excellent | Full tool suite including computer use |
| **Grok 4** (xAI) | Excellent | Excellent | First-principles reasoning with native tool use |
| **DeepSeek-V3.2** | Excellent | Excellent | Thinking Retention Mechanism preserves reasoning across tool calls |
| **o3 / o4-mini** | Good | Excellent | Native function calling within chain-of-thought. No parallel tool calls |
| **Mistral Large 3** | Excellent | Very Good | 673B MoE. Best open-weight for multi-tool orchestration |

## Conventions

- **Module**: `github.com/tailored-agentic-units/kernel` — single Go module, single version, no dependency cascade
- **Testing**: Co-located tests (`*_test.go` alongside source), black-box testing using `package_test` suffix, table-driven patterns. Top-level `tests/` reserved for kernel-wide integration tests only
- **No `pkg/` prefix**: Packages live directly under their subsystem root
- **Versioning**: Phase target `v<major>.<minor>.<patch>`, dev pre-release `v<target>-dev.<objective>.<issue>`
- **Package boundaries**: Enforced by Go's import rules and type system — no repository walls needed

## Principles

- Each subsystem has a single clear responsibility
- Dependencies flow in one direction: core → capabilities → composition
- The kernel has zero awareness of what connects to its interface
- Local development works without containers or external infrastructure
- Single module, single version — no dependency cascade
