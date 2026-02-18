# 25 - Kernel Observer

## Problem Context

The kernel's ad-hoc `*slog.Logger` and the orchestrate Observer pattern share the same integration points. The observability package is tightly coupled to `orchestrate/`, but both kernel and orchestrate need it. This promotes observability to a root-level package with OTel-compatible severity levels, integrates it into the kernel runtime loop, and migrates all orchestrate imports.

## Architecture Approach

- **Root-level `observability/` package** — foundation-level (Level 0), no internal dependencies
- **OTel-aligned severity levels** — Level values ARE OTel SeverityNumbers (5=DEBUG, 9=INFO, 13=WARN, 17=ERROR)
- **Level-aware SlogObserver** — maps event levels to slog levels, uses event type as log message
- **Separate `node.state` event** — state snapshots split from `node.complete` into dedicated event type
- **Decentralized event types** — each package defines its own event type constants using `observability.EventType`
- **SlogObserver default** — observable out of the box via `slog.Default()`

## Implementation

### Step 1: Create `observability/` package

```bash
mkdir -p observability
```

#### observability/observer.go

```go
package observability

import (
	"context"
	"log/slog"
	"time"
)

type Level int

const (
	LevelVerbose Level = 5
	LevelInfo    Level = 9
	LevelWarning Level = 13
	LevelError   Level = 17
)

func (l Level) String() string {
	switch {
	case l <= 4:
		return "TRACE"
	case l <= 8:
		return "DEBUG"
	case l <= 12:
		return "INFO"
	case l <= 16:
		return "WARN"
	case l <= 20:
		return "ERROR"
	default:
		return "FATAL"
	}
}

func (l Level) SlogLevel() slog.Level {
	switch {
	case l <= 8:
		return slog.LevelDebug
	case l <= 12:
		return slog.LevelInfo
	case l <= 16:
		return slog.LevelWarn
	default:
		return slog.LevelError
	}
}

type EventType string

type Event struct {
	Type      EventType
	Level     Level
	Timestamp time.Time
	Source    string
	Data      map[string]any
}

type Observer interface {
	OnEvent(ctx context.Context, event Event)
}
```

#### observability/noop.go

```go
package observability

import "context"

type NoOpObserver struct{}

func (NoOpObserver) OnEvent(ctx context.Context, event Event) {}
```

#### observability/multi.go

```go
package observability

import "context"

type MultiObserver struct {
	observers []Observer
}

func NewMultiObserver(observers ...Observer) *MultiObserver {
	filtered := make([]Observer, 0, len(observers))
	for _, obs := range observers {
		if obs != nil {
			filtered = append(filtered, obs)
		}
	}
	return &MultiObserver{observers: filtered}
}

func (m *MultiObserver) OnEvent(ctx context.Context, event Event) {
	for _, obs := range m.observers {
		obs.OnEvent(ctx, event)
	}
}
```

#### observability/slog.go

```go
package observability

import (
	"context"
	"log/slog"
)

type SlogObserver struct {
	logger *slog.Logger
}

func NewSlogObserver(logger *slog.Logger) *SlogObserver {
	return &SlogObserver{logger: logger}
}

func (o *SlogObserver) OnEvent(ctx context.Context, event Event) {
	attrs := make([]slog.Attr, 0, len(event.Data)+1)
	attrs = append(attrs, slog.String("source", event.Source))
	for k, v := range event.Data {
		attrs = append(attrs, slog.Any(k, v))
	}

	o.logger.LogAttrs(ctx, event.Level.SlogLevel(), string(event.Type), attrs...)
}
```

#### observability/registry.go

```go
package observability

import (
	"fmt"
	"log/slog"
	"sync"
)

var (
	observers = map[string]Observer{
		"noop": NoOpObserver{},
		"slog": NewSlogObserver(slog.Default()),
	}
	mutex sync.RWMutex
)

func GetObserver(name string) (Observer, error) {
	mutex.RLock()
	defer mutex.RUnlock()

	obs, exists := observers[name]
	if !exists {
		return nil, fmt.Errorf("unknown observer: %s", name)
	}
	return obs, nil
}

func RegisterObserver(name string, observer Observer) {
	mutex.Lock()
	defer mutex.Unlock()

	observers[name] = observer
}
```

### Step 2: Migrate orchestrate imports

Every orchestrate file changes its import path, event type constants move from `observability/` to the emitting package, and every `observability.Event{}` literal gains a `Level` field.

#### orchestrate/state/events.go

New file. Event type constants for state, graph, checkpoint, and node operations:

```go
package state

import "github.com/tailored-agentic-units/kernel/observability"

const (
	EventStateCreate EventType = "state.create"
	EventStateClone  EventType = "state.clone"
	EventStateSet    EventType = "state.set"
	EventStateMerge  EventType = "state.merge"

	EventGraphStart     EventType = "graph.start"
	EventGraphComplete  EventType = "graph.complete"
	EventNodeStart      EventType = "node.start"
	EventNodeComplete   EventType = "node.complete"
	EventNodeState      EventType = "node.state"
	EventEdgeEvaluate   EventType = "edge.evaluate"
	EventEdgeTransition EventType = "edge.transition"
	EventCycleDetected  EventType = "cycle.detected"

	EventCheckpointSave   EventType = "checkpoint.save"
	EventCheckpointLoad   EventType = "checkpoint.load"
	EventCheckpointResume EventType = "checkpoint.resume"
)

type EventType = observability.EventType
```

The type alias `EventType = observability.EventType` lets the constants use the unqualified name while remaining the same type.

#### orchestrate/workflows/events.go

New file. Event type constants for chains, parallel, and conditional routing:

```go
package workflows

import "github.com/tailored-agentic-units/kernel/observability"

const (
	EventChainStart    EventType = "chain.start"
	EventChainComplete EventType = "chain.complete"
	EventStepStart     EventType = "step.start"
	EventStepComplete  EventType = "step.complete"

	EventParallelStart    EventType = "parallel.start"
	EventParallelComplete EventType = "parallel.complete"
	EventWorkerStart      EventType = "worker.start"
	EventWorkerComplete   EventType = "worker.complete"

	EventRouteEvaluate EventType = "route.evaluate"
	EventRouteSelect   EventType = "route.select"
	EventRouteExecute  EventType = "route.execute"
)

type EventType = observability.EventType
```

#### orchestrate/state/state.go

Import change:

```go
	"github.com/tailored-agentic-units/kernel/observability"
```

`EventStateCreate` in `New()`:

```go
	observer.OnEvent(context.Background(), observability.Event{
		Type:      EventStateCreate,
		Level:     observability.LevelVerbose,
		Timestamp: s.Timestamp,
		Source:    "state",
		Data:      map[string]any{},
	})
```

`EventStateClone` in `Clone()`:

```go
	s.Observer.OnEvent(context.Background(), observability.Event{
		Type:      EventStateClone,
		Level:     observability.LevelVerbose,
		Timestamp: time.Now(),
		Source:    "state",
		Data:      map[string]any{"keys": len(newState.Data)},
	})
```

`EventStateSet` in `Set()`:

```go
	s.Observer.OnEvent(context.Background(), observability.Event{
		Type:      EventStateSet,
		Level:     observability.LevelVerbose,
		Timestamp: time.Now(),
		Source:    "state",
		Data:      map[string]any{"key": key},
	})
```

`EventStateMerge` in `Merge()`:

```go
	s.Observer.OnEvent(context.Background(), observability.Event{
		Type:      EventStateMerge,
		Level:     observability.LevelVerbose,
		Timestamp: time.Now(),
		Source:    "state",
		Data:      map[string]any{"keys": len(other.Data)},
	})
```

#### orchestrate/state/graph.go

Import change:

```go
	"github.com/tailored-agentic-units/kernel/observability"
```

`EventCheckpointLoad` in `Resume()`:

```go
	g.observer.OnEvent(ctx, observability.Event{
		Type:      EventCheckpointLoad,
		Level:     observability.LevelInfo,
		Timestamp: time.Now(),
		Source:    g.name,
		Data: map[string]any{
			"node":   state.CheckpointNode,
			"run_id": runID,
		},
	})
```

`EventCheckpointResume` in `Resume()`:

```go
	g.observer.OnEvent(ctx, observability.Event{
		Type:      EventCheckpointResume,
		Level:     observability.LevelInfo,
		Timestamp: time.Now(),
		Source:    g.name,
		Data: map[string]any{
			"checkpoint_node": state.CheckpointNode,
			"resume_node":     nextNode,
			"run_id":          runID,
		},
	})
```

`EventGraphStart` in `execute()`:

```go
	g.observer.OnEvent(ctx, observability.Event{
		Type:      EventGraphStart,
		Level:     observability.LevelInfo,
		Timestamp: time.Now(),
		Source:    g.name,
		Data: map[string]any{
			"entry_point": g.entryPoint,
			"run_id":      initialState.RunID,
			"exit_points": len(g.exitPoints),
		},
	})
```

`EventCycleDetected` in `execute()`:

```go
			g.observer.OnEvent(ctx, observability.Event{
				Type:      EventCycleDetected,
				Level:     observability.LevelWarning,
				Timestamp: time.Now(),
				Source:    g.name,
				Data: map[string]any{
					"node":        current,
					"visit_count": visited[current],
					"iteration":   iterations,
					"path_length": len(path),
				},
			})
```

`EventNodeStart` in `execute()` — remove `input_snapshot`:

```go
		g.observer.OnEvent(ctx, observability.Event{
			Type:      EventNodeStart,
			Level:     observability.LevelVerbose,
			Timestamp: time.Now(),
			Source:    g.name,
			Data: map[string]any{
				"node":      current,
				"iteration": iterations,
			},
		})
```

`EventNodeComplete` in `execute()` — remove `output_snapshot`, add new `EventNodeState` emission immediately after:

```go
		g.observer.OnEvent(ctx, observability.Event{
			Type:      EventNodeComplete,
			Level:     observability.LevelVerbose,
			Timestamp: time.Now(),
			Source:    g.name,
			Data: map[string]any{
				"node":      current,
				"iteration": iterations,
				"error":     err != nil,
			},
		})

		g.observer.OnEvent(ctx, observability.Event{
			Type:      EventNodeState,
			Level:     observability.LevelVerbose,
			Timestamp: time.Now(),
			Source:    g.name,
			Data: map[string]any{
				"node":            current,
				"iteration":       iterations,
				"input_snapshot":  maps.Clone(state.Data),
				"output_snapshot": maps.Clone(newState.Data),
			},
		})
```

`EventCheckpointSave` in `execute()`:

```go
			g.observer.OnEvent(ctx, observability.Event{
				Type:      EventCheckpointSave,
				Level:     observability.LevelInfo,
				Timestamp: time.Now(),
				Source:    g.name,
				Data: map[string]any{
					"node":   current,
					"run_id": state.RunID,
				},
			})
```

`EventGraphComplete` in `execute()`:

```go
			g.observer.OnEvent(ctx, observability.Event{
				Type:      EventGraphComplete,
				Level:     observability.LevelInfo,
				Timestamp: time.Now(),
				Source:    g.name,
				Data: map[string]any{
					"exit_point":  current,
					"iterations":  iterations,
					"path_length": len(path),
				},
			})
```

`EventEdgeEvaluate` in `execute()`:

```go
			g.observer.OnEvent(ctx, observability.Event{
				Type:      EventEdgeEvaluate,
				Level:     observability.LevelVerbose,
				Timestamp: time.Now(),
				Source:    g.name,
				Data: map[string]any{
					"from":          edge.From,
					"to":            edge.To,
					"edge_index":    i,
					"has_predicate": edge.Predicate != nil,
				},
			})
```

`EventEdgeTransition` in `execute()`:

```go
				g.observer.OnEvent(ctx, observability.Event{
					Type:      EventEdgeTransition,
					Level:     observability.LevelVerbose,
					Timestamp: time.Now(),
					Source:    g.name,
					Data: map[string]any{
						"from":             edge.From,
						"to":               edge.To,
						"edge_index":       i,
						"predicate_name":   edge.Name,
						"predicate_result": true,
					},
				})
```

#### orchestrate/workflows/chain.go

Import change:

```go
	"github.com/tailored-agentic-units/kernel/observability"
```

`EventChainStart` in `ProcessChain()`:

```go
	observer.OnEvent(ctx, observability.Event{
		Type:      EventChainStart,
		Level:     observability.LevelInfo,
		Timestamp: time.Now(),
		Source:    "workflows.ProcessChain",
		Data: map[string]any{
			"item_count":            len(items),
			"has_progress_callback": progress != nil,
			"capture_intermediate":  cfg.CaptureIntermediateStates,
		},
	})
```

`EventChainComplete` — empty items path:

```go
		observer.OnEvent(ctx, observability.Event{
			Type:      EventChainComplete,
			Level:     observability.LevelInfo,
			Timestamp: time.Now(),
			Source:    "workflows.ProcessChain",
			Data: map[string]any{
				"steps_completed": 0,
				"error":           false,
			},
		})
```

`EventChainComplete` — cancellation path:

```go
			observer.OnEvent(ctx, observability.Event{
				Type:      EventChainComplete,
				Level:     observability.LevelInfo,
				Timestamp: time.Now(),
				Source:    "workflows.ProcessChain",
				Data: map[string]any{
					"steps_completed": i,
					"error":           true,
					"error_type":      "cancellation",
				},
			})
```

`EventStepStart`:

```go
		observer.OnEvent(ctx, observability.Event{
			Type:      EventStepStart,
			Level:     observability.LevelVerbose,
			Timestamp: time.Now(),
			Source:    "workflows.ProcessChain",
			Data: map[string]any{
				"step_index":  i,
				"total_steps": len(items),
			},
		})
```

`EventStepComplete` — error path:

```go
			observer.OnEvent(ctx, observability.Event{
				Type:      EventStepComplete,
				Level:     observability.LevelVerbose,
				Timestamp: time.Now(),
				Source:    "workflows.ProcessChain",
				Data: map[string]any{
					"step_index":  i,
					"total_steps": len(items),
					"error":       true,
				},
			})
```

`EventChainComplete` — processor error path:

```go
			observer.OnEvent(ctx, observability.Event{
				Type:      EventChainComplete,
				Level:     observability.LevelInfo,
				Timestamp: time.Now(),
				Source:    "workflows.ProcessChain",
				Data: map[string]any{
					"steps_completed": i,
					"error":           true,
					"error_type":      "processor",
				},
			})
```

`EventStepComplete` — success path:

```go
		observer.OnEvent(ctx, observability.Event{
			Type:      EventStepComplete,
			Level:     observability.LevelVerbose,
			Timestamp: time.Now(),
			Source:    "workflows.ProcessChain",
			Data: map[string]any{
				"step_index":  i,
				"total_steps": len(items),
				"error":       false,
			},
		})
```

`EventChainComplete` — success path:

```go
	observer.OnEvent(ctx, observability.Event{
		Type:      EventChainComplete,
		Level:     observability.LevelInfo,
		Timestamp: time.Now(),
		Source:    "workflows.ProcessChain",
		Data: map[string]any{
			"steps_completed": len(items),
			"error":           false,
		},
	})
```

#### orchestrate/workflows/conditional.go

Import change:

```go
	"github.com/tailored-agentic-units/kernel/observability"
```

`EventRouteEvaluate`:

```go
	observer.OnEvent(ctx, observability.Event{
		Type:      EventRouteEvaluate,
		Level:     observability.LevelVerbose,
		Timestamp: time.Now(),
		Source:    "conditional",
		Data: map[string]any{
			"route_count": len(routes.Handlers),
		},
	})
```

`EventRouteSelect`:

```go
	observer.OnEvent(ctx, observability.Event{
		Type:      EventRouteSelect,
		Level:     observability.LevelVerbose,
		Timestamp: time.Now(),
		Source:    "conditional",
		Data: map[string]any{
			"route":       route,
			"has_default": routes.Default != nil,
		},
	})
```

`EventRouteExecute`:

```go
	observer.OnEvent(ctx, observability.Event{
		Type:      EventRouteExecute,
		Level:     observability.LevelVerbose,
		Timestamp: time.Now(),
		Source:    "conditional",
		Data: map[string]any{
			"route": route,
			"error": false,
		},
	})
```

#### orchestrate/workflows/parallel.go

Import change:

```go
	"github.com/tailored-agentic-units/kernel/observability"
```

`EventParallelStart` — empty items path:

```go
		observer.OnEvent(ctx, observability.Event{
			Type:      EventParallelStart,
			Level:     observability.LevelInfo,
			Timestamp: time.Now(),
			Source:    "workflows.ProcessParallel",
			Data: map[string]any{
				"item_count":            0,
				"worker_count":          0,
				"fail_fast":             cfg.FailFast(),
				"has_progress_callback": progress != nil,
			},
		})
```

`EventParallelComplete` — empty items path:

```go
		observer.OnEvent(ctx, observability.Event{
			Type:      EventParallelComplete,
			Level:     observability.LevelInfo,
			Timestamp: time.Now(),
			Source:    "workflows.ProcessParallel",
			Data: map[string]any{
				"items_processed": 0,
				"items_failed":    0,
				"error":           false,
			},
		})
```

`EventParallelStart` — normal path:

```go
	observer.OnEvent(ctx, observability.Event{
		Type:      EventParallelStart,
		Level:     observability.LevelInfo,
		Timestamp: time.Now(),
		Source:    "workflows.ProcessParallel",
		Data: map[string]any{
			"item_count":            len(items),
			"worker_count":          workerCount,
			"fail_fast":             cfg.FailFast(),
			"has_progress_callback": progress != nil,
		},
	})
```

All 4 `EventParallelComplete` emissions (collector error, context cancelled, fail-fast/all-failed, success) follow the same pattern — add `Level: observability.LevelInfo,`:

```go
		observer.OnEvent(ctx, observability.Event{
			Type:      EventParallelComplete,
			Level:     observability.LevelInfo,
			Timestamp: time.Now(),
			Source:    "workflows.ProcessParallel",
			Data: map[string]any{
				"items_processed": len(results),
				"items_failed":    len(errors),
				"error":           ...,  // true or false depending on path
			},
		})
```

`EventWorkerStart` in `processWorker()`:

```go
			observer.OnEvent(ctx, observability.Event{
				Type:      EventWorkerStart,
				Level:     observability.LevelVerbose,
				Timestamp: time.Now(),
				Source:    "workflows.ProcessParallel",
				Data: map[string]any{
					"worker_id":   workerID,
					"item_index":  work.index,
					"total_items": total,
				},
			})
```

`EventWorkerComplete` in `processWorker()`:

```go
			observer.OnEvent(ctx, observability.Event{
				Type:      EventWorkerComplete,
				Level:     observability.LevelVerbose,
				Timestamp: time.Now(),
				Source:    "workflows.ProcessParallel",
				Data: map[string]any{
					"worker_id":   workerID,
					"item_index":  work.index,
					"total_items": total,
					"error":       err != nil,
				},
			})
```

#### orchestrate/examples/

All 5 example programs need only the import path change. Replace:

```go
	"github.com/tailored-agentic-units/kernel/orchestrate/observability"
```

with:

```go
	"github.com/tailored-agentic-units/kernel/observability"
```

Files:
- `orchestrate/examples/phase-02-03-state-graphs/main.go`
- `orchestrate/examples/phase-04-sequential-chains/main.go`
- `orchestrate/examples/phase-05-parallel-execution/main.go`
- `orchestrate/examples/phase-06-checkpointing/main.go`
- `orchestrate/examples/darpa-procurement/main.go`

### Step 3: Delete `orchestrate/observability/`

```bash
rm -rf orchestrate/observability
```

### Step 4: Create `kernel/observer.go`

```go
package kernel

import "github.com/tailored-agentic-units/kernel/observability"

const (
	EventRunStart       observability.EventType = "kernel.run.start"
	EventRunComplete    observability.EventType = "kernel.run.complete"
	EventIterationStart observability.EventType = "kernel.iteration.start"
	EventToolCall       observability.EventType = "kernel.tool.call"
	EventToolComplete   observability.EventType = "kernel.tool.complete"
	EventResponse       observability.EventType = "kernel.response"
	EventError          observability.EventType = "kernel.error"
)
```

### Step 5: Modify `kernel/kernel.go`

**Imports**: Remove `"io"` and `"log/slog"`. Add `"github.com/tailored-agentic-units/kernel/observability"`.

**Struct field**: Replace `log *slog.Logger` with `observer observability.Observer`.

**Option**: Replace `WithLogger` with:

```go
func WithObserver(o observability.Observer) Option {
	return func(k *Kernel) { k.observer = o }
}
```

**Default in `New`**: Replace `slog.New(slog.NewTextHandler(io.Discard, nil))` with `observability.NewSlogObserver(slog.Default())`. Keep the `"log/slog"` import.

**Event emissions** — replace all `k.log.*` calls:

`Run` method — run started (line 168):
```go
	k.observer.OnEvent(ctx, observability.Event{
		Type:      EventRunStart,
		Level:     observability.LevelInfo,
		Timestamp: time.Now(),
		Source:    "kernel.Run",
		Data: map[string]any{
			"prompt_length":  len(prompt),
			"max_iterations": k.maxIterations,
			"tools":          len(k.tools.List()),
		},
	})
```

`Run` method — iteration started (line 175):
```go
		k.observer.OnEvent(ctx, observability.Event{
			Type:      EventIterationStart,
			Level:     observability.LevelVerbose,
			Timestamp: time.Now(),
			Source:    "kernel.Run",
			Data:      map[string]any{"iteration": iteration + 1},
		})
```

`Run` method — final response (line 198, replace `k.log.Info("run complete", ...)`):
```go
			k.observer.OnEvent(ctx, observability.Event{
				Type:      EventResponse,
				Level:     observability.LevelInfo,
				Timestamp: time.Now(),
				Source:    "kernel.Run",
				Data: map[string]any{
					"iteration":       iteration + 1,
					"response_length": len(result.Response),
				},
			})
```

`Run` method — tool call (line 210):
```go
			k.observer.OnEvent(ctx, observability.Event{
				Type:      EventToolCall,
				Level:     observability.LevelVerbose,
				Timestamp: time.Now(),
				Source:    "kernel.Run",
				Data: map[string]any{
					"iteration": iteration + 1,
					"name":      tc.Function.Name,
				},
			})
```

After tool execution (after the if/else block that handles toolErr, around line 242), add tool complete event:
```go
			k.observer.OnEvent(ctx, observability.Event{
				Type:      EventToolComplete,
				Level:     observability.LevelVerbose,
				Timestamp: time.Now(),
				Source:    "kernel.Run",
				Data: map[string]any{
					"iteration": iteration + 1,
					"name":      tc.Function.Name,
					"error":     record.IsError,
				},
			})
```

`Run` method — max iterations (line 248):
```go
	k.observer.OnEvent(ctx, observability.Event{
		Type:      EventError,
		Level:     observability.LevelWarning,
		Timestamp: time.Now(),
		Source:    "kernel.Run",
		Data: map[string]any{
			"error":      "max iterations reached",
			"iterations": k.maxIterations,
		},
	})
```

`buildSystemContent` — memory loaded (line 287): Remove the `k.log.Debug("memory loaded", ...)` call. Memory loading is an internal implementation detail not covered by the kernel event types.

Add `"time"` to the import list.

### Step 6: Modify `cmd/kernel/main.go`

Replace logger creation and `WithLogger` usage. Change:

```go
import (
	...
	"log/slog"
	...
	"github.com/tailored-agentic-units/kernel/kernel"
)
```

to:

```go
import (
	...
	"log/slog"
	...
	"github.com/tailored-agentic-units/kernel/kernel"
	"github.com/tailored-agentic-units/kernel/observability"
)
```

Replace:
```go
	runtime, err := kernel.New(cfg, kernel.WithLogger(logger))
```

with:
```go
	runtime, err := kernel.New(cfg, kernel.WithObserver(observability.NewSlogObserver(logger)))
```

## Validation Criteria

- [ ] `go vet ./...` passes
- [ ] `go test ./...` passes
- [ ] `go mod tidy` produces no changes
- [ ] No `orchestrate/observability` imports remain in codebase
- [ ] Kernel event types cover all 7 loop integration points
- [ ] `WithObserver` option replaces `WithLogger`
- [ ] Default `SlogObserver` uses `slog.Default()` — observable out of the box
- [ ] SlogObserver respects event levels (Info vs Debug filtering)
- [ ] Level values match OTel SeverityNumber ranges
- [ ] `node.state` events separate from `node.complete`
