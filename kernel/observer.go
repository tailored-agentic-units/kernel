package kernel

import "github.com/tailored-agentic-units/kernel/observability"

// Kernel event types emitted during the agentic loop.
const (
	EventRunStart       observability.EventType = "kernel.run.start"
	EventRunComplete    observability.EventType = "kernel.run.complete"
	EventIterationStart observability.EventType = "kernel.iteration.start"
	EventToolCall       observability.EventType = "kernel.tool.call"
	EventToolComplete   observability.EventType = "kernel.tool.complete"
	EventResponse       observability.EventType = "kernel.response"
	EventError          observability.EventType = "kernel.error"
)
