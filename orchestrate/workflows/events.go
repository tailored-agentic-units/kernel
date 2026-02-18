package workflows

import "github.com/tailored-agentic-units/kernel/observability"

const (
	// Sequential chains
	EventChainStart    observability.EventType = "chain.start"
	EventChainComplete observability.EventType = "chain.complete"
	EventStepStart     observability.EventType = "step.start"
	EventStepComplete  observability.EventType = "step.complete"

	// Parallel execution
	EventParallelStart    observability.EventType = "parallel.start"
	EventParallelComplete observability.EventType = "parallel.complete"
	EventWorkerStart      observability.EventType = "worker.start"
	EventWorkerComplete   observability.EventType = "worker.complete"

	// Conditional routing
	EventRouteEvaluate observability.EventType = "route.evaluate"
	EventRouteSelect   observability.EventType = "route.select"
	EventRouteExecute  observability.EventType = "route.execute"
)
