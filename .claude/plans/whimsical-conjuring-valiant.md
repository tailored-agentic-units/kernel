# Project Review — kernel

**Date:** 2026-02-16
**Scope:** tailored-agentic-units/kernel
**Session type:** Project Review

## Context

Phase 1 (Foundation) has reached a critical milestone: all foundation subsystems (memory, tools, session) are implemented, and the kernel runtime loop is merged. However, the only way to verify the composed system is through unit tests — there is no runnable entry point that exercises the kernel against a real LLM. This review evaluates the current state, validates the roadmap against a principle established in agent-lab (runtime validation supersedes integration tests), and recommends infrastructure adjustments to get a runnable kernel sooner.

## Findings

### Infrastructure Audit

| Area | Status | Notes |
|------|--------|-------|
| Project board | OK | Phases match `_project/README.md` |
| Labels | OK | Consistent subsystem + type labels |
| Issue types | OK | Objectives and Tasks properly typed |
| Orphaned issues | 6 items | #5-10 (orchestrate observability, Whisper provider) — backlog, not linked to any objective |
| Missing config | Bug | `cmd/prompt-agent/config.ollama.json` referenced in README and CLAUDE.md does not exist |

### Codebase Assessment

| Area | Grade | Notes |
|------|-------|-------|
| Code quality | A | Clean interfaces, acyclic dependencies, consistent patterns |
| Test coverage | B+ | 40 test files across all packages, all passing. No integration tests. |
| Documentation | B | Good exported-type docs. README Quick Start references missing config file. |
| Dependencies | A | Minimal: uuid, protobuf, connect |
| Technical debt | A- | `cmd/kernel/main.go` is a stub. `mcp/` is a skeleton. |

### Context Health

| Area | Status | Notes |
|------|--------|-------|
| Concepts | Current | `orchestrate/advanced-observability.md` active |
| Guides | Properly archived | #11-14 all in `.archive/` |
| Sessions | Complete | Summaries exist for #11-14 |
| CLAUDE.md | Needs update | References missing config file; kernel entry point not documented |
| `_project/objective.md` | Stale | Shows #14 as Open — it is closed |
| `_project/phase.md` | Current | Objective 1 correctly shown as In Progress |
| `_project/README.md` | Needs update | `skills/` skeleton was removed in #13 (skills consolidated into memory namespace); still listed as a separate subsystem in dependency hierarchy |

### Vision Alignment

The Phase 1 vision is sound. The architecture, dependency hierarchy, and subsystem boundaries are working as designed. The gap is in **delivery sequencing**: the build order correctly phased the library construction but didn't include a runnability checkpoint after the core loop was assembled.

**Key insight:** `kernel.New` + `kernel.Run` is fully functional. It composes session, memory, tools, and agent into the agentic cycle. All it lacks is a CLI entry point and a config file.

### Hardware Validation

Qwen3-8B (Q4_K_M quantization via Ollama) on RTX 2080 (8GB VRAM):
- **Model size:** 5.2 GB download, ~5.38 GB VRAM at Q4
- **VRAM usage:** 67.3% of 8 GB — comfortable headroom for KV cache at moderate context lengths
- **Performance (estimated from 3070 Ti 8GB):** ~69 tok/sec generation, ~99ms TTFT at 1K context
- **Verdict:** Solid fit for development and testing. Keep context length modest (8K or below) to stay within VRAM budget.

## Decisions

### 1. Refactor #15 from integration tests to runnable kernel CLI

**Rationale:** The agent-lab principle applies directly — runtime validation through a CLI entry point exercises the full subsystem composition against a real LLM. This catches what mocks cannot: prompt formatting, tool schema compatibility, provider-specific behavior. A running kernel IS the integration test. Integration tests may be revisited at a later point if a concrete need arises.

**Action:** Refactor Issue #15's scope from "kernel integration tests" to "runnable kernel CLI with built-in tools." Update the issue title, body, and acceptance criteria.

### 2. Skills consolidated into memory package

Confirmed from Session #13: the standalone `skills/` skeleton was removed and skills became a namespace within the memory system (`skills/` prefix in the Store). `_project/README.md` and CLAUDE.md dependency hierarchies should reflect this.

### 3. Config naming convention

Config files follow the pattern `agent.[provider].[model].json`. The kernel CLI config will be `cmd/kernel/agent.ollama.qwen3.json`.

### 4. Reviews are permanent records

Review reports in `.claude/context/reviews/` are not archived. They serve as a historical record of what was analyzed and what decisions were made.

### 5. Qwen3-8B confirmed as local development model

Fits RTX 2080 (8GB VRAM) at Q4_K_M quantization with room to spare. Docker Compose already configured to auto-load the model.

## Infrastructure Adjustments

### GitHub Issues

| Action | Target | Details |
|--------|--------|---------|
| Refactor | #15 | Change scope from integration tests to runnable kernel CLI with built-in tools |
| Update | #1 body | Update sub-issue table: #14 closed, #15 scope changed |

### Documentation Updates

| File | Fix |
|------|-----|
| `_project/objective.md` | Mark #14 as Closed. Update #15 title/description to match refactored scope. |
| `_project/README.md` | Remove `skills/` from subsystem topology and dependency hierarchy (consolidated into memory). Update kernel status. |
| `.claude/CLAUDE.md` | Remove `skills/` from project structure. Fix prompt-agent config reference. Add kernel CLI to Quick Reference. Update config naming convention. |
| `README.md` | Update Quick Start to show kernel CLI usage. Fix missing config reference. |

### Review Report

Create `.claude/context/reviews/2026-02-16-kernel.md` capturing this review's findings, decisions, and adjustments.

## Verification

- `gh issue view 15 -R tailored-agentic-units/kernel` shows refactored scope
- `_project/objective.md` shows #14 as Closed and #15 with updated description
- `_project/README.md` no longer references standalone `skills/` package
- Review report exists at `.claude/context/reviews/2026-02-16-kernel.md`
