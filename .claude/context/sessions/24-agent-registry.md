# 24 - Agent Registry

## Summary

Added a named agent registry to the `agent` package with lazy instantiation and capability querying. The kernel creates and owns a registry instance, populating it from config during initialization. Agents are registered by name with their configs; actual `Agent` instances are created on first `Get()` call. Capabilities are derived from `ModelConfig.Capabilities` keys without requiring instantiation.

## Key Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Package placement | `agent` package, not `kernel` | Registry manages agents — it's the agent package's domain. Kernel owns an instance, same pattern as `session.Session`. |
| Instance-owned vs global | Exported `Registry` type, instance per kernel | Test isolation — unlike the global tools registry |
| Lazy instantiation | Config stored on Register, agent created on Get | Avoids unnecessary LLM client initialization for unused agents |
| Get() locking | Single write lock | Simpler than read-lock-then-upgrade with double-check; adequate for agent access patterns |
| Config merge for Agents map | Source replaces target wholesale | Consistent with scalar "non-zero source overrides" pattern; avoids surprising partial-merge behavior |
| Replace method | Included, invalidates cached agent | Mirrors tools.Replace pattern; enables config updates at runtime |

## Files Modified

- `agent/errors.go` — added 3 sentinel errors (ErrAgentNotFound, ErrAgentExists, ErrEmptyAgentName)
- `agent/registry.go` — new file: Registry type, AgentInfo, 6 methods + helper
- `agent/registry_test.go` — new file: 15 test cases covering all methods + concurrency
- `kernel/config.go` — added Agents map field, updated Merge
- `kernel/config_test.go` — added 4 config tests (merge, replace, JSON loading)
- `kernel/kernel.go` — added registry field, wiring in New, Registry() accessor, WithRegistry option
- `kernel/kernel_test.go` — added 3 integration tests
- `_project/README.md` — updated agent subsystem description
- `README.md` — updated agent package description
- `.claude/CLAUDE.md` — updated project structure
- `.claude/skills/kernel-dev/SKILL.md` — updated agent package responsibilities

## Patterns Established

- Instance-owned registry type in a subsystem package, kernel owns the instance (vs. global registry in tools)
- Lazy instantiation: store config, create on demand
- Capability querying from config keys without agent instantiation

## Validation Results

- All tests pass (`go test ./...`)
- Race detector clean (`go test -race ./agent/... ./kernel/...`)
- `go vet ./...` clean
- `go mod tidy` no changes
- Registry coverage: 91-100% across methods
- Kernel config coverage: 100%
