# Plan: Issue #25 — Kernel Observer

## Context

The kernel's `*slog.Logger` and the orchestrate Observer pattern share the same integration points. The observability package is currently tightly coupled to `orchestrate/`, but both kernel and orchestrate need it. This issue promotes observability to a root-level package with OTel-compatible severity levels, integrates it into the kernel, and migrates orchestrate imports.

## Architecture: Root-Level `observability/`

### Event with OTel-aligned severity

```go
type Level int

const (
    LevelVerbose Level = 5   // OTel DEBUG (5-8), slog.Debug
    LevelInfo    Level = 9   // OTel INFO (9-12), slog.Info
    LevelWarning Level = 13  // OTel WARN (13-16), slog.Warn
    LevelError   Level = 17  // OTel ERROR (17-20), slog.Error
)
```

Level values ARE OTel SeverityNumbers — zero translation for OTel collectors. `Level.SlogLevel()` maps to Go's slog levels. `Level.String()` returns OTel severity text ("DEBUG", "INFO", "WARN", "ERROR").

```go
type Event struct {
    Type      EventType      // OTel EventName
    Level     Level          // OTel SeverityNumber
    Timestamp time.Time      // OTel Timestamp
    Source    string         // OTel InstrumentationScope
    Data      map[string]any // OTel Attributes
}
```

### SlogObserver (level-aware)

Uses `event.Level.SlogLevel()` to emit at the correct slog level. Flattens `Data` keys as top-level slog attributes. Uses `string(event.Type)` as the log message.

```
time=... level=INFO msg=kernel.run.start source=kernel.Run prompt_length=42 tools=3
time=... level=DEBUG msg=kernel.tool.call source=kernel.Run name=greet iteration=1
```

### Log readability via levels

The primary readability improvement is level-based filtering:

- **Info** consumers see lifecycle boundaries only (run start/complete, graph start/complete, chain start/complete)
- **Verbose** consumers see execution details (iterations, tool calls, node/edge/step events)
- State snapshots are separated into dedicated `node.state` events at Verbose — observers can filter by type within the Verbose tier

### Kernel event types (defined in `kernel/observer.go`)

| Constant | Value | Level |
|----------|-------|-------|
| `EventRunStart` | `kernel.run.start` | Info |
| `EventRunComplete` | `kernel.run.complete` | Info |
| `EventIterationStart` | `kernel.iteration.start` | Verbose |
| `EventToolCall` | `kernel.tool.call` | Verbose |
| `EventToolComplete` | `kernel.tool.complete` | Verbose |
| `EventResponse` | `kernel.response` | Info |
| `EventError` | `kernel.error` | Warning |

### Orchestrate event level assignments

| Category | Events | Level |
|----------|--------|-------|
| Lifecycle boundaries | graph.start/complete, chain.start/complete, parallel.start/complete, checkpoint.* | Info |
| Internal execution | state.*, node.start/complete, node.state, edge.*, step.*, worker.*, route.* | Verbose |
| Anomalies | cycle.detected | Warning |

### New event type: `node.state`

Separates state snapshots from `node.complete`:
- **`node.complete`** — clean metadata: node name, iteration, error status (no snapshots)
- **`node.state`** — full state snapshot: emitted after `node.complete`, carries `input_snapshot` and `output_snapshot`

Both at LevelVerbose. The event type is the discriminator for observers that want to filter snapshots.

## Implementation Steps

### Step 1: Create `observability/` package

New root-level package with these files:

| File | Contents |
|------|----------|
| `observability/observer.go` | Observer interface, Event struct (with Level), EventType, Level type (OTel-aligned constants, String(), SlogLevel()) |
| `observability/events.go` | All orchestrate event type constants (moved from orchestrate/observability/observer.go) + new `EventNodeState` |
| `observability/noop.go` | NoOpObserver (zero-cost discard) |
| `observability/multi.go` | MultiObserver (fan-out to multiple observers) |
| `observability/slog.go` | SlogObserver (level-aware, flattened attrs) |
| `observability/registry.go` | Global observer registry (GetObserver, RegisterObserver) |

### Step 2: Migrate orchestrate imports

Update all orchestrate source and test files:

**Source** (import path change + add Level to event emissions):
- `orchestrate/state/state.go` — 4 events, all LevelVerbose
- `orchestrate/state/graph.go` — 10 existing events + 1 new `node.state` event. Remove `input_snapshot`/`output_snapshot` from `node.start`/`node.complete`, add new `node.state` emission after `node.complete`
- `orchestrate/workflows/chain.go` — 5 events, Info for lifecycle, Verbose for steps
- `orchestrate/workflows/conditional.go` — 3 events, all LevelVerbose
- `orchestrate/workflows/parallel.go` — 4 events, Info for lifecycle, Verbose for workers

**Tests** (import path change + add Level to expected events):
- `orchestrate/state/state_test.go`, `graph_test.go`, `edge_test.go`, `node_test.go`, `checkpoint_test.go`
- `orchestrate/workflows/chain_test.go`, `conditional_test.go`, `integration_test.go`

**Examples** (import path change):
- `orchestrate/examples/*/main.go` (5 example programs)

### Step 3: Delete `orchestrate/observability/`

Remove the entire `orchestrate/observability/` directory.

### Step 4: Create `kernel/observer.go`

Kernel event type constants using `observability.EventType`. No kernel-specific SlogObserver needed — the root-level one handles levels correctly.

### Step 5: Modify `kernel/kernel.go`

- Replace `log *slog.Logger` → `observer observability.Observer`
- Replace `WithLogger` → `WithObserver`
- Default: `observability.NewSlogObserver(slog.Default())` (observable out of the box)
- Replace all `k.log.*` calls → `k.observer.OnEvent(ctx, ...)`
- Remove `"io"` and `"log/slog"` imports, add `"github.com/tailored-agentic-units/kernel/observability"`

### Step 6: Modify `cmd/kernel/main.go`

- Replace `kernel.WithLogger(logger)` → `kernel.WithObserver(observability.NewSlogObserver(logger))`

## Dependency Hierarchy Change

```
Before:
  orchestrate/observability → (none)
  orchestrate/state → orchestrate/observability
  kernel → agent, session, memory, tools, core

After:
  observability → (none)                    ← new Level 0 foundation package
  orchestrate/state → observability         ← import path change
  kernel → agent, session, memory, tools, core, observability  ← new dependency
```

## Validation

- [ ] `go vet ./...` passes
- [ ] `go test ./...` passes
- [ ] `go mod tidy` produces no changes
- [ ] No `orchestrate/observability` imports remain
- [ ] Kernel event types cover all 7 loop integration points
- [ ] `WithObserver` option works, `WithLogger` removed
- [ ] Default `SlogObserver` uses `slog.Default()`
- [ ] SlogObserver respects event levels (Info vs Debug filtering works)
- [ ] Level values match OTel SeverityNumber ranges
- [ ] `node.state` events separate from `node.complete` (no inline snapshots)
