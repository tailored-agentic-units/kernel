# 25 - Kernel Observer

## Summary

Promoted `orchestrate/observability/` to a root-level `observability/` package with OTel-aligned severity levels, integrated it into the kernel runtime loop replacing the ad-hoc `*slog.Logger`, and migrated all orchestrate imports. Event types are decentralized — each package defines its own constants.

## Key Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Package location | Root-level `observability/` | Foundation package needed by both kernel and orchestrate — can't be nested under either |
| Severity levels | OTel SeverityNumbers (5, 9, 13, 17) | Zero translation for OTel collectors; maps cleanly to slog levels |
| Event type ownership | Decentralized — each package defines its own | Consistent with kernel event types in `kernel/observer.go`; each package owns its domain |
| Default observer | `SlogObserver(slog.Default())` | Observable out of the box; `WithObserver` allows override |
| State snapshots | Separate `node.state` event | Keeps `node.complete` clean; event type is discriminator for snapshot filtering |
| Level filtering | slog handler responsibility | Observer translates levels; handler filters — single responsibility |

## Files Modified

- `observability/observer.go` — Core types: Observer, Event, EventType, Level (OTel-aligned)
- `observability/noop.go` — NoOpObserver
- `observability/multi.go` — MultiObserver (fan-out)
- `observability/slog.go` — Level-aware SlogObserver
- `observability/registry.go` — Global observer registry
- `observability/observer_test.go` — Full test suite (100% coverage)
- `orchestrate/state/events.go` — State/graph/checkpoint event constants
- `orchestrate/workflows/events.go` — Chain/parallel/conditional event constants
- `orchestrate/state/state.go` — Import migration + Level field
- `orchestrate/state/graph.go` — Import migration + Level field + node.state event
- `orchestrate/workflows/chain.go` — Import migration + Level field
- `orchestrate/workflows/conditional.go` — Import migration + Level field
- `orchestrate/workflows/parallel.go` — Import migration + Level field
- `orchestrate/examples/*/main.go` — Import migration (5 files)
- `orchestrate/state/*_test.go` — Import + event type reference updates (5 files)
- `orchestrate/workflows/*_test.go` — Import + event type reference updates (3 files)
- `kernel/observer.go` — Kernel event type constants
- `kernel/kernel.go` — Observer replaces slog logger
- `kernel/kernel_test.go` — TestWithObserver replaces TestWithLogger
- `cmd/kernel/main.go` — WithObserver replaces WithLogger
- `_project/README.md` — Architecture, topology, hierarchy updates
- `_project/objective.md` — Status update, known gaps section
- `.claude/CLAUDE.md` — Structure and hierarchy updates
- `.claude/skills/kernel-dev/SKILL.md` — Package responsibilities and hierarchy updates
- `README.md` — Subsystem descriptions
- Deleted: `orchestrate/observability/` (entire directory)

## Patterns Established

- **OTel-aligned severity**: Level values ARE OTel SeverityNumbers — no translation layer needed
- **Decentralized event types**: Each package defines `observability.EventType` constants for its domain
- **Type alias for event constants**: `type EventType = observability.EventType` in event files allows unqualified constant names
- **Level-based log filtering**: Info for lifecycle boundaries, Verbose for execution details, Warning for anomalies
- **Separate state snapshot events**: `node.state` splits from `node.complete` for clean observer filtering

## Validation Results

- `go vet ./...` — passes
- `go test ./...` — all 18 test packages pass
- `go mod tidy` — no changes
- `observability/` coverage: 100%
- No `orchestrate/observability` imports remain in `.go` files
