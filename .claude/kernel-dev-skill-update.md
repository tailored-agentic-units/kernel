# Session Init: Update kernel-dev Skill (C2)

## Goal

Update the kernel-dev skill to reflect the post-extraction architecture. After C1 (post-extraction refactor), the kernel's package structure, dependency hierarchy, and extension patterns will have changed significantly. The skill must be updated before any subsequent kernel development so that every Phase 1-3 task benefits from accurate context.

## What Changes

### Package Structure (Post-Extraction)

The kernel sheds its local copies of core, agent, and orchestrate packages. The post-extraction structure is:

```
kernel/
├── kernel/         # Runtime loop
├── session/        # Conversation management
├── memory/         # Context composition (FileStore, Cache)
├── tools/          # Tool registry and execution
├── mcp/            # MCP client
├── api/            # HTTP + SSE handlers (NEW, replaces rpc/)
├── cmd/            # Entry points
├── tests/          # Integration tests
```

Removed: `core/`, `agent/`, `orchestrate/`, `observability/`, `rpc/`

### Dependency Hierarchy

The current skill documents a 9-level internal hierarchy. Post-extraction, the hierarchy flattens dramatically:

- **External libraries**: protocol, format, provider, agent, orchestrate (imported, not local)
- **Kernel-local packages**: session, memory, tools, mcp, api, kernel (runtime loop)
- **Foundation level**: memory, tools, session (no internal cross-deps)
- **Integration level**: mcp (depends on tools)
- **Composition level**: kernel, api (depend on everything above)

### Extension Patterns

Current patterns (adding providers, observers, workflow patterns) move to library-dev skills:
- Adding a provider → provider-dev skill
- Adding an observer → orchestrate-dev skill
- Adding a workflow → orchestrate-dev skill

New kernel-specific extension patterns:
- Adding a built-in tool (tools/ package)
- Adding an API endpoint (api/ package)
- Adding a session strategy (session/ package)
- Adding a memory namespace (memory/ package)
- MCP transport implementation (mcp/ package)

### Interface Changes

- `protocol.Tool` → `format.ToolDefinition`
- `response.ChatResponse` / `response.ToolsResponse` → `response.Response` (unified)
- `response.StreamingChunk` → `response.StreamingResponse`
- ConnectRPC service → HTTP + SSE handlers
- Agent import: `kernel/agent` → `github.com/tailored-agentic-units/agent`

## Source Documents

- Post-extraction plan: ~/tau/kernel/_project/post-extraction.md (comprehensive migration guide)
- Current kernel-dev skill: ~/tau/kernel/.claude/skills/kernel-dev/SKILL.md
- kernel CLAUDE.md: ~/tau/kernel/.claude/CLAUDE.md (also needs updating, but that's part of C1)

## Process

1. Complete C1 (post-extraction refactor) first — the skill update must reflect actual code, not planned changes
2. Read the refactored codebase to verify the new structure
3. Rewrite SKILL.md sections:
   - Architecture → post-extraction package structure
   - Package Responsibilities → kernel-local packages only
   - Extension Patterns → kernel-specific patterns (tools, API, session, memory, MCP)
   - Dependency Hierarchy → flattened, library-based
   - Testing Strategy → updated for new import paths and mock locations

## Definition of Done

- [ ] SKILL.md reflects post-extraction architecture
- [ ] All references to removed packages (core/, agent/, orchestrate/, observability/) are gone
- [ ] Extension patterns cover kernel-specific contribution types
- [ ] Dependency hierarchy shows external library imports + kernel-local packages
- [ ] Testing strategy references tau/agent/mock and tau/orchestrate test utilities
