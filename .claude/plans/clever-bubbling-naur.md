# Project Review: TAU Kernel — Strategic Decomposition

## Context

The kernel project has been paused since Feb 18, 2025 (~6 weeks). During that time, significant improvements were made to the upstream `go-agents` library (v0.5.0) including a format abstraction layer, unified response types, AWS Bedrock provider, streaming transport abstraction, and identity/credential management. The user wants to decompose the kernel's lower-level subsystems into standalone TAU libraries (`tau/agent`, `tau/orchestrate`) so the kernel can focus purely on agentic harness functionality. This aligns with the broader industry direction (OpenClaw/NemoClaw validate the kernel-as-harness pattern). The tau-marketplace also needs decomposition from a monolithic plugin into standalone per-skill plugins.

---

## Phase 1: Infrastructure Audit

### Project Board — "TAU Platform" (org project #1)

| Area | Status | Notes |
|------|--------|-------|
| Project board | Needs attention | Objective #2 shows "Todo" but is 50% complete (3/6 sub-issues merged) — should be "In Progress" |
| Phases | Needs attention | Board has Phase field with "Backlog" and "Phase 1 - Foundation" but no items are phase-assigned |
| Labels | OK | 17 labels with subsystem + type taxonomy |
| Issue types | Needs attention | Objectives (#1-4) have `objective` label. Sub-issues (#5-10) lack type assignment — no `task` label |
| Milestones | Needs attention | Only "Phase 1 - Foundation" exists; issues #5-10 have no milestone |
| Orphaned issues | 6 | #5-10 are on the board but unassigned to any objective or milestone |

### Issue Summary

| Category | Count | Details |
|----------|-------|---------|
| Total issues | 28 | |
| Open | 10 | #2-4 (objectives), #5-10 (orphaned), #26-28 (Kernel Interface sub-issues) |
| Closed | 18 | #1, #11-15, #23-25 + associated PRs |
| On board | 21 | All kernel issues present |

### Issue Disposition (post-decomposition)

- **#5-9** (advanced observability) — Will belong to `tau/orchestrate`
- **#10** (Whisper audio provider) — Will belong to `tau/agent`
- **#26-28** (multi-session, HTTP API, server entry point) — Stay in kernel, target post-decomposition architecture

> **Note**: No changes to kernel project management infrastructure until new libraries are built and architecture is proven.

---

## Phase 2: Codebase Assessment

### Health

| Area | Grade | Notes |
|------|-------|-------|
| Code quality | A | Clean acyclic dependency hierarchy, consistent patterns across 18 packages |
| Test coverage | A- | All 18 test packages pass. `core/model` lacks tests (thin bridging type). `go vet` clean. |
| Dependencies | B+ | `connectrpc.com/connect` and `google.golang.org/protobuf` are dead weight — architecture decision moved to pure HTTP+SSE (#27). `rpc/` directory is orphaned infrastructure. |
| Technical debt | B | Significant divergence from go-agents v0.5.0 — kernel's response types, provider interface, and streaming model are now outdated |

### Critical Divergence: Kernel vs go-agents v0.5.0

| Aspect | Kernel (current) | go-agents v0.5.0 |
|--------|------------------|-------------------|
| Response types | Separate `ChatResponse`, `ToolsResponse`, `StreamingChunk` | Unified `Response` with `ContentBlock` interface |
| Wire format | Marshaling hardcoded inside each provider | `format.Format` interface with registry (OpenAI, Converse) |
| Streaming | Provider returns `<-chan any` | `streaming.StreamReader` interface (SSE, EventStream) |
| Providers | Ollama, Azure | Ollama, Azure, **Bedrock** |
| Auth/Identity | Inline in providers | Dedicated `identities` package |
| Agent.Chat | `Chat(ctx, []Message, ...opts)` | `Chat(ctx, string, ...opts)` |

### Decomposition Opportunity: Decouple orchestrate from agent

`orchestrate/hub` imports the full `agent.Agent` interface but only uses `ag.ID()`. By defining a local `Participant` interface in hub:

```go
type Participant interface {
    ID() string
}
```

tau/orchestrate becomes **fully independent** from tau/agent:

```
tau/agent (primitives — zero TAU deps)
tau/orchestrate (coordination — zero TAU deps)
    ↓ both consumed by ↓
tau/kernel (harness — depends on both, composes them)
```

This is cleaner than the linear chain (`agent → orchestrate → kernel`). The kernel becomes the sole composition layer where `agent.Agent` satisfies `hub.Participant`.

### Package Extraction Map

**To `tau/agent`** (`github.com/tailored-agentic-units/agent`):

Root package defines the Agent interface and constructor (`agent.Agent`, `agent.New()`). Flat layout:
- Root: `agent.Agent` interface, `agent.New()`, `agent.Registry`
- `config/` — agent configuration, duration handling
- `protocol/` — protocol constants, message types
- `response/` — response parsing (unified model from go-agents v0.5.0)
- `model/` — model runtime bridge
- `client/` — HTTP transport with retry
- `providers/` — LLM platform adapters (Ollama, Azure, Bedrock)
- `request/` — request construction
- `mock/` — testing doubles
- `format/` — wire format abstraction (NEW from go-agents)
- `streaming/` — streaming transport abstraction (NEW from go-agents)
- `identities/` — credential management (NEW from go-agents)

**To `tau/orchestrate`** (`github.com/tailored-agentic-units/orchestrate`):
- `observability/` — Observer, Event, Level (OTel-aligned)
- `config/` — hub, state, workflow configuration
- `hub/` — multi-agent coordination (uses local `Participant` interface, not agent.Agent)
- `messaging/` — message structures and builders
- `state/` — state graphs, checkpoints
- `workflows/` — chain, parallel, conditional patterns
- `examples/` — usage examples

**Kernel retains** (post-decomposition):
- `kernel/` — runtime loop (depends on tau/agent + tau/orchestrate)
- `session/` — conversation management
- `memory/` — context composition
- `tools/` — tool registry
- `mcp/` — MCP client skeleton
- `api/` — HTTP+SSE handlers (to be created in #27, replaces removed `rpc/`)
- `cmd/` — entry points

---

## Phase 3: Context Infrastructure Review

| Area | Status | Notes |
|------|--------|-------|
| Concepts | Current | `advanced-observability.md` in concepts/ — will migrate to tau/orchestrate |
| Guides | OK | 7 archived guides for completed issues |
| Sessions | OK | 7 session summaries for completed work |
| Reviews | Empty | No previous review reports exist — this will be the first |
| CLAUDE.md | Stale pending decomposition | Will need update after library extraction, not before |
| Skills | Stale pending decomposition | kernel-dev references packages that will be extracted |

### _project/ Health

| Document | Status | Notes |
|----------|--------|-------|
| README.md | Stale | ConnectRPC references in vision statement should be corrected (decision was pure HTTP+SSE). Architecture/topology will change post-decomposition. |
| phase.md | Stale pending decomposition | Phase 1 scope must be re-evaluated after library extraction. |
| objective.md | Partially stale | Objective #2 is 50% complete. Architecture decisions documented correctly (pure HTTP+SSE). |

> **Note**: Documentation updates deferred until new infrastructure is proven and informs precise changes needed.

---

## Phase 4: Vision Alignment

### Industry Validation

OpenClaw and NemoClaw (NVIDIA) validate the kernel-as-harness approach:
- **OpenClaw**: Local-first AI agent runtime — similar to kernel's vision
- **NemoClaw**: Adds security guardrails and policy enforcement — similar to what kernel's orchestrate subsystem provides (hub coordination, state management, workflow patterns)

The TAU kernel's approach of separating agent primitives from orchestration from the runtime harness is architecturally sound and now validated by major industry players.

### Strategic Direction: Three Independent Libraries

```
tau/agent (primitives — zero TAU deps)
tau/orchestrate (coordination — zero TAU deps)
    ↓ both consumed by ↓
tau/kernel (harness — the composition layer)
```

- **tau/agent**: Model communication, formats, streaming, providers, identities. Standalone utility for anyone building model-facing capabilities.
- **tau/orchestrate**: Multi-agent coordination, state graphs, workflows, observability. Standalone framework for anyone orchestrating agents (uses local `Participant` interface, not agent types).
- **tau/kernel**: The harness that composes both — runtime loop, sessions, memory, tools, MCP, API. This is where `agent.Agent` meets `hub.Participant`.

### Phase Roadmap Revision

**Current Phase 1** (Foundation, v0.1.0): 3/6 sub-issues complete. Remaining 3 (#26-28) form a chain.

**Proposed revision** — insert library extraction before completing Phase 1:

| Phase | Focus | Target |
|-------|-------|--------|
| **Phase 1A: Library Extraction** | Build tau/agent and tau/orchestrate; create standalone marketplace plugins | Pre-requisite for continuing kernel work |
| **Phase 1B: Foundation (resumed)** | Refactor kernel to depend on new libraries; complete #26-28 | v0.1.0 |
| Phase 2 | Skills, MCP, subsystem observability | v0.2.0 |
| Phase 3 | Advanced orchestration, OTel | v0.3.0 |

### Marketplace Decomposition

**Current**: Single `tau` plugin with 6 skills — all-or-nothing install.

**Proposed**: Every skill becomes a standalone plugin. No more monolithic `tau` plugin.

| Plugin | Skill | Purpose |
|--------|-------|---------|
| `tau-dev-workflow` | dev-workflow | Structured development sessions |
| `tau-github-cli` | github-cli | GitHub CLI operations |
| `tau-go-patterns` | go-patterns | Go design patterns |
| `tau-project-management` | project-management | GitHub Projects v2 |
| `tau-overview` | tau-overview | Ecosystem conventions |
| `tau-agent` | agent | tau/agent usage guide |
| `tau-orchestrate` | orchestrate | tau/orchestrate usage guide |
| `tau-kernel` | kernel | tau/kernel usage guide |

Dev skills (`agent-dev`, `orchestrate-dev`, `kernel-dev`) are **not** marketplace plugins — they are co-located in their respective repositories under `.claude/skills/`.

### ConnectRPC Cleanup

Per objective.md, the architecture decision was pure HTTP + JSON + SSE. Dead infrastructure to remove from kernel:
- `rpc/` directory (buf configs, proto definitions, generated code)
- `connectrpc.com/connect` and `google.golang.org/protobuf` dependencies
- ConnectRPC references in `_project/README.md`
- `proto` label on the repo

> **Note**: This cleanup can happen as part of the concept session or as a standalone housekeeping PR. It does not depend on library extraction.

---

## Phase 5: Infrastructure Adjustments

### Decisions Made

1. **Phase order**: Phase 1A (Library Extraction) before resuming #26-28
2. **Package layout**: Flat top-level for tau/agent; root package defines Agent interface
3. **Protocol design**: Best-of-both-worlds — go-agents v0.5.0 restructuring was necessary for format/streaming layers; tau/agent adopts this while preserving kernel innovations (registry, multi-turn flexibility)
4. **Orchestrate independence**: Hub uses local `Participant` interface instead of importing agent — makes tau/orchestrate fully independent from tau/agent
5. **Marketplace**: All skills become standalone plugins; dev skills stay co-located in repos
6. **Kernel changes deferred**: No kernel project management or documentation changes until new libraries are built and architecture is proven
7. **Next step**: Concept development session to formalize the architecture

### Immediate Actions (this review session)

1. Create review report at `.claude/context/reviews/2026-04-01-kernel.md`

### Next Session: Concept Development

`/dev-workflow concept` — "Library Extraction"

Deep analysis of go-agents v0.5.0 protocol layer (format, streaming, response model) to design tau/agent's consolidated protocol. The concept should produce:
- Three-library architecture document with zero-TAU-dep orchestrate design
- Package mapping with flat layout for tau/agent (root = Agent interface)
- Protocol design reconciling go-agents v0.5.0 with kernel innovations
- Repository initialization plans for tau/agent and tau/orchestrate
- Marketplace plugin decomposition plan (all standalone)
- Phased roadmap (Phase 1A → Phase 1B)

### Deferred Actions (after new infrastructure is proven)

- Kernel project board updates (phase assignments, objective status)
- Kernel issue reassignment (#5-10 to new repos)
- Kernel documentation updates (CLAUDE.md, _project/, skills)
- ConnectRPC cleanup
- Re-scoping #26-28 for post-decomposition architecture

---

## Phase 6: Recommendations

1. **Start with a concept development session** to formalize the library extraction architecture. The decomposition touches 3+ repos and creates 2 new ones — it needs a proper concept document before phase/objective planning.

2. **Preserve kernel's agent registry** in tau/agent — go-agents v0.5.0 doesn't have this.

3. **Reconcile protocol layers deeply** — the go-agents v0.5.0 restructuring (format abstraction, streaming transport, unified response with content blocks) was driven by the need to reduce friction between providers and formats. tau/agent must adopt this consolidated protocol while preserving kernel innovations. This is the central design challenge for the concept session.

4. **Exploit the Participant interface pattern** — decoupling orchestrate from agent is a significant architectural improvement that makes both libraries independently useful.

5. **Plan team onboarding** once Phase A infrastructure is complete, before Phase B refactoring.

---

## Verification

- [x] All kernel tests pass (`go test ./...`)
- [x] `go vet ./...` clean
- [x] Project board queried and state documented
- [x] All issues cataloged with status
- [x] go-agents v0.5.0 improvements identified and compared
- [x] go-agents-orchestration state assessed
- [x] tau-marketplace structure analyzed
- [x] OpenClaw/NemoClaw landscape researched
- [x] hub→agent coupling analyzed — only uses ID(), decoupling viable
- [ ] Review report to be created at `.claude/context/reviews/2026-04-01-kernel.md`
