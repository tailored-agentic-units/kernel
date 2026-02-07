# Architecture Detail Reference

Detailed file listings and type signatures for each package.

## Package File Structure

```
pkg/
├── observability/
│   ├── observer.go      # Observer interface, Event, EventType constants
│   ├── noop.go          # NoOpObserver (zero-cost)
│   ├── slog.go          # SlogObserver (structured logging)
│   ├── multi.go         # MultiObserver (broadcast to multiple)
│   ├── registry.go      # GetObserver/RegisterObserver with sync.RWMutex
│   └── doc.go           # Package documentation
│
├── messaging/
│   ├── message.go       # Message struct, helper methods
│   ├── builder.go       # NewRequest, NewResponse, NewNotification, NewBroadcast builders
│   └── types.go         # MessageType (request/response/notification/broadcast), Priority enums
│
├── hub/
│   ├── hub.go           # Hub interface, concrete implementation, New constructor
│   ├── handler.go       # MessageHandler type, MessageContext struct
│   ├── channel.go       # MessageChannel[T] generic wrapper with context-aware Send/Receive
│   └── metrics.go       # Metrics struct (LocalAgents, MessagesSent, MessagesRecv)
│
├── config/
│   ├── hub.go           # HubConfig with DefaultHubConfig()
│   ├── state.go         # GraphConfig, CheckpointConfig with defaults
│   ├── workflows.go     # ChainConfig, ParallelConfig, ConditionalConfig with defaults
│   └── doc.go           # Package documentation
│
├── state/
│   ├── state.go         # State struct (Data, Secrets, Observer, RunID, CheckpointNode, Timestamp)
│   ├── node.go          # StateNode interface, FunctionNode implementation
│   ├── edge.go          # Edge struct, TransitionPredicate, built-in predicates
│   ├── graph.go         # StateGraph interface, concrete stateGraph, NewGraph, NewGraphWithDeps
│   ├── checkpoint.go    # CheckpointStore interface, MemoryCheckpointStore, registry
│   └── error.go         # ExecutionError (NodeName, State, Path, Err)
│
└── workflows/
    ├── chain.go         # ProcessChain[TItem, TContext], ChainResult
    ├── parallel.go      # ProcessParallel[TItem, TResult], ParallelResult
    ├── conditional.go   # ProcessConditional[TState], Routes, RoutePredicate, RouteHandler
    ├── integration.go   # ChainNode, ParallelNode, ConditionalNode (StateNode wrappers)
    ├── progress.go      # ChainProgress, ParallelProgress callback types
    └── error.go         # ChainError, ParallelError, ConditionalError, TaskError
```

## State Type Detail

```go
type State struct {
    Data           map[string]any        `json:"data"`
    Secrets        map[string]any        `json:"-"`
    Observer       observability.Observer `json:"-"`
    RunID          string                `json:"run_id"`
    CheckpointNode string               `json:"checkpoint_node"`
    Timestamp      time.Time            `json:"timestamp"`
}
```

Methods: New, Clone, Get, Set, Merge, GetSecret, SetSecret, DeleteSecret, SetCheckpointNode, Checkpoint

## Edge Predicates

Built-in predicate constructors:
- `AlwaysTransition()` — Always returns true
- `KeyExists(key)` — True if key exists in state data
- `KeyEquals(key, value)` — True if key equals value
- `Not(predicate)` — Logical negation
- `And(predicates...)` — Logical conjunction
- `Or(predicates...)` — Logical disjunction

## Observer Events

```go
// Graph execution
EventGraphStart, EventGraphComplete
EventNodeStart, EventNodeComplete
EventEdgeEvaluate, EventEdgeTransition
EventCycleDetected

// Checkpointing
EventCheckpointSave, EventCheckpointLoad, EventCheckpointResume

// Workflows
EventChainStart, EventChainStepStart, EventChainStepComplete, EventChainComplete
EventParallelStart, EventWorkerStart, EventWorkerComplete, EventParallelComplete
EventRouteEvaluate, EventRouteSelect, EventRouteExecute
```

## Workflow Pattern Signatures

```go
// Sequential chain
func ProcessChain[TItem any, TContext any](
    ctx context.Context,
    cfg config.ChainConfig,
    items []TItem,
    initial TContext,
    processor func(ctx context.Context, item TItem, state TContext) (TContext, error),
    progress ChainProgress[TItem, TContext],
) (ChainResult[TItem, TContext], error)

// Parallel execution
func ProcessParallel[TItem any, TResult any](
    ctx context.Context,
    cfg config.ParallelConfig,
    items []TItem,
    processor func(ctx context.Context, item TItem) (TResult, error),
    progress ParallelProgress[TItem, TResult],
) (ParallelResult[TItem, TResult], error)

// Conditional routing
func ProcessConditional[TState any](
    ctx context.Context,
    cfg config.ConditionalConfig,
    state TState,
    predicate RoutePredicate[TState],
    routes Routes[TState],
) (TState, error)
```

## Configuration Defaults

```go
DefaultHubConfig()        // Name: "hub", Buffer: 100, Timeout: 30s
DefaultGraphConfig()      // MaxIterations: 100, Observer: "slog"
DefaultChainConfig()      // CaptureIntermediateStates: false, Observer: "slog"
DefaultParallelConfig()   // MaxWorkers: 0 (auto), FailFast: true, Observer: "slog"
DefaultConditionalConfig() // Observer: "slog"
```

## Hub Communication Patterns

1. **Send**: Fire-and-forget, async delivery, no response expected
2. **Request/Response**: Synchronous with correlation ID, timeout via context
3. **Broadcast**: All registered agents except sender, best-effort
4. **Pub/Sub**: Topic-based subscription, delivery to all topic subscribers
