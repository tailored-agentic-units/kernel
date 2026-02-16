# kernel

TAU (Tailored Agentic Units) kernel — agent runtime with integrated subsystems and a ConnectRPC extension boundary.

## Quick Reference

| Action | Command |
|--------|---------|
| Test | `go test ./...` |
| Coverage (all) | `go test ./... -coverprofile=coverage.out` |
| Coverage (pkg) | `go test ./<package>/... -coverprofile=c.out` |
| Coverage report | `go tool cover -func=coverage.out` |
| HTML report | `go tool cover -html=coverage.out -o coverage.html` |
| Validate | `go vet ./...` |
| Proto lint | `cd rpc && buf lint` |
| Proto generate | `cd rpc && buf generate` |
| Kernel (Ollama) | `go run ./cmd/kernel/ -config cmd/kernel/agent.ollama.qwen3.json -prompt "..."` |
| Prompt (Ollama) | `go run cmd/prompt-agent/main.go -config cmd/prompt-agent/agent.ollama.qwen3.json -prompt "..." -stream` |
| Ollama | `docker compose up -d` |

## Module

```
github.com/tailored-agentic-units/kernel
```

Single Go module. All packages share one version. No dependency cascade.

## Project Structure

```
kernel/
├── _project/          # Project identity, phase, and objective context
├── core/               # Foundational types: protocol, response, config, model
├── agent/              # LLM communication: agent interface, client, providers, request, mock
├── orchestrate/        # Multi-agent coordination: hub, messaging, state, workflows, observability
├── memory/             # Unified context composition: Store, FileStore, Cache. Namespaces: memory/, skills/, agents/
├── tools/              # Tool execution: global registry with Register, Execute, List
├── session/            # Conversation management: Session interface, in-memory implementation
├── mcp/                # MCP client (skeleton)
├── kernel/             # Agent runtime loop with config-driven initialization
├── rpc/                # ConnectRPC infrastructure (proto, buf configs, generated code)
├── cmd/                # Entry points (kernel, prompt-agent)
├── scripts/            # Infrastructure scripts (Azure)
├── .claude/            # Claude Code configuration and skills
└── .github/            # CI workflows
```

## Package Dependency Hierarchy

```
Level 0: core/config, core/protocol
Level 1: core/response, core/model
Level 2: agent/providers, agent/request, agent/client
Level 3: agent (root)
Level 4: agent/mock
Level 5: orchestrate/observability, orchestrate/messaging, orchestrate/config
Level 6: orchestrate/hub, orchestrate/state
Level 7: orchestrate/workflows

Foundation (Level 0 — depend only on core/protocol):
  memory, tools, session

Level 8: kernel (depends on agent, session, memory, tools, core)
```

## Design Principles

- Every package's exported data structures are its protocol. Higher-level packages that depend on it consume those types natively.
- `core/` types are the foundational protocol — subsystems build on them, not around them. If a subsystem needs to mutate a core type signature, evolve the core type instead.
- This applies at every level of the hierarchy: if `session` defines `Session`, then `kernel` uses `session.Session` directly — no wrapping, no re-definition.
- When planning implementation, address gaps at the lowest affected dependency level rather than working around them at higher levels.

## Testing

- Tests are co-located with source in each package (`*_test.go` alongside source files)
- Black-box testing using `package_test` suffix
- Top-level `tests/` reserved for kernel-wide integration tests only
- Table-driven test patterns

## Versioning

Single version for the entire kernel:
- Phase target: `v<major>.<minor>.<patch>` (e.g., `v0.1.0`)
- Dev pre-release: `v<target>-dev.<objective>.<issue>` (e.g., `v0.1.0-dev.3.7`)

## Skills

| Skill | Source | Use When |
|-------|--------|----------|
| tau:dev-workflow | plugin | Development sessions: concept, plan, task, review, release |
| tau:go-patterns | plugin | Go design patterns, interfaces, error handling |
| tau:project-management | plugin | GitHub Projects v2, phases, objectives |
| tau:kernel | plugin | Building applications with the kernel |
| kernel-dev | local | Contributing to the kernel, architecture, testing |

## Context Documents

Project knowledge artifacts stored in `.claude/context/`:

| Directory | Contents | Naming |
|-----------|----------|--------|
| `concepts/` | Architectural concept documents | `[slug].md` |
| `guides/` | Active implementation guides | `[issue-number]-[slug].md` |
| `sessions/` | Session summaries | `[issue-number]-[slug].md` |
| `reviews/` | Project review reports | `[YYYY-MM-DD]-[scope].md` |

Concepts, guides, and sessions have `.archive/` subdirectories for completed documents. Reviews are permanent records and are not archived. Directories are created on demand.

## Task Session: Documentation Review

During a `tau:dev-workflow` task execution session, Phase 7 (Documentation) must include a review of project context documents for any revisions necessitated by the implementation. Check the following files and update any stale descriptions, statuses, or references:

- `_project/README.md` — subsystem topology statuses, known gaps, build order descriptions
- `_project/objective.md` — sub-issue statuses
- `README.md` — subsystem descriptions
- `.claude/CLAUDE.md` — project structure, dependency hierarchy
- `.claude/skills/kernel-dev/SKILL.md` — package responsibilities, dependency hierarchy, extension patterns

This review happens before Phase 8 (Closeout) to ensure all project documentation stays consistent with the codebase.

## Session Continuity

Plan files in `.claude/plans/` enable session continuity across machines.

### Saving Session State

When pausing work, append a context snapshot to the active plan file:

```markdown
## Context Snapshot - [YYYY-MM-DD HH:MM]

**Current State**: [Brief description]

**Files Modified**:
- [List of files changed]

**Next Steps**:
- [Immediate next action]
- [Subsequent actions]

**Key Decisions**:
- [Decisions made and rationale]

**Blockers/Questions**:
- [Unresolved issues]
```

### Restoring Session State

1. Read the plan file to restore context
2. Review the most recent Context Snapshot
3. Resume from the documented Next Steps
4. Update the snapshot when pausing again

## Dependencies

- `github.com/google/uuid` - Agent identification (UUIDv7)
- `google.golang.org/protobuf` - Protocol buffers runtime
- `connectrpc.com/connect` - ConnectRPC framework
