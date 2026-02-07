// Package observability provides minimal observability primitives for orchestration workflows.
//
// This package establishes the foundation for production-grade observability without
// impacting performance when observability is not needed. It defines interfaces and types
// for capturing execution events across state management, graph execution, and workflow patterns.
//
// # Core Components
//
// Observer - Interface for receiving execution events
//
// Event - Structure containing event metadata (type, timestamp, source, data)
//
// EventType - Constants for all observable events across orchestration primitives
//
// NoOpObserver - Zero-cost observer implementation when observability not needed
//
// # Observer Registry
//
// The package provides a registry pattern enabling configuration-driven observer selection:
//
//	observability.RegisterObserver("slog", NewSlogObserver())
//	observer, err := observability.GetObserver("slog")
//
// This enables JSON configuration:
//
//	{"name": "my-graph", "observer": "slog", "max_iterations": 100}
//
// # Usage
//
// Observer hooks are integrated into all orchestration primitives from Phase 2 onwards.
// Events are emitted at key execution points without affecting normal execution flow.
//
// When observability is not needed, use NoOpObserver for zero performance overhead:
//
//	observer := observability.NoOpObserver{}
//	state := state.New(observer) // No events emitted
//
// Phase 8 will provide production Observer implementations (structured logging, metrics,
// trace correlation). The minimal interface established here prevents retrofit friction.
package observability
