package observability

import "context"

// NoOpObserver provides a zero-cost Observer implementation that discards all events.
//
// Use NoOpObserver when observability is not needed to avoid performance overhead.
// The implementation is stateless and can be safely reused across goroutines.
//
// Example:
//
//	observer := observability.NoOpObserver{}
//	state := state.New(observer) // No events emitted, zero overhead
type NoOpObserver struct{}

// OnEvent discards the event without any processing.
func (NoOpObserver) OnEvent(ctx context.Context, event Event) {}
