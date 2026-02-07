package observability

import (
	"context"
	"time"
)

// Observer receives execution events from orchestration primitives.
//
// Observer implementations can log events, collect metrics, trace execution flow,
// or capture decision points. The interface is intentionally minimal to avoid
// coupling orchestration primitives to specific observability implementations.
//
// Implementations should not affect execution flow - errors or delays in OnEvent
// should not propagate to the caller.
type Observer interface {
	// OnEvent receives an execution event with metadata about what happened.
	// The context provides cancellation/timeout control for expensive operations.
	OnEvent(ctx context.Context, event Event)
}

// Event represents an observable occurrence during workflow execution.
//
// Events capture execution metadata rather than application data. This approach
// enables observability without exposing sensitive information or impacting performance.
type Event struct {
	// Type categorizes the event (state.set, node.execute, etc.)
	Type EventType

	// Timestamp records when the event occurred
	Timestamp time.Time

	// Source identifies the component that emitted the event (state, graph, chain, etc.)
	Source string

	// Data contains metadata about the event (keys changed, duration, progress, etc.)
	// This is execution telemetry, not application data
	Data map[string]any
}

// EventType categorizes observable events across orchestration primitives.
//
// Event types are defined for all phases (2-8) to establish a consistent event
// model across the entire orchestration infrastructure.
type EventType string

const (
	// Phase 2: State operations
	EventStateCreate EventType = "state.create"
	EventStateClone  EventType = "state.clone"
	EventStateSet    EventType = "state.set"
	EventStateMerge  EventType = "state.merge"

	// Phase 3: Graph execution
	EventGraphStart     EventType = "graph.start"
	EventGraphComplete  EventType = "graph.complete"
	EventNodeStart      EventType = "node.start"
	EventNodeComplete   EventType = "node.complete"
	EventEdgeEvaluate   EventType = "edge.evaluate"
	EventEdgeTransition EventType = "edge.transition"
	EventCycleDetected  EventType = "cycle.detected"

	// Phase 4: Sequential chains
	EventChainStart    EventType = "chain.start"
	EventChainComplete EventType = "chain.complete"
	EventStepStart     EventType = "step.start"
	EventStepComplete  EventType = "step.complete"

	// Phase 5: Parallel execution
	EventParallelStart    EventType = "parallel.start"
	EventParallelComplete EventType = "parallel.complete"
	EventWorkerStart      EventType = "worker.start"
	EventWorkerComplete   EventType = "worker.complete"

	// Phase 6: Checkpointing
	EventCheckpointSave   EventType = "checkpoint.save"
	EventCheckpointLoad   EventType = "checkpoint.load"
	EventCheckpointResume EventType = "checkpoint.resume"

	// Phase 7: Conditional routing
	EventRouteEvaluate EventType = "route.evaluate"
	EventRouteSelect   EventType = "route.select"
	EventRouteExecute  EventType = "route.execute"
)
