package observability

import "context"

// MultiObserver broadcasts events to multiple wrapped observers.
//
// Use MultiObserver when events need to reach multiple destinations simultaneously,
// such as persisting to a database while also streaming to clients. The implementation
// is not thread-safe for modification after construction; all observers should be
// provided at creation time.
type MultiObserver struct {
	observers []Observer
}

// NewMultiObserver creates a MultiObserver that broadcasts to all provided observers.
// Nil observers are filtered out during construction to prevent nil pointer panics.
func NewMultiObserver(observers ...Observer) *MultiObserver {
	filtered := make([]Observer, 0, len(observers))
	for _, obs := range observers {
		if obs != nil {
			filtered = append(filtered, obs)
		}
	}
	return &MultiObserver{observers: filtered}
}

// OnEvent forwards the event to all wrapped observers sequentially.
// The context is propagated to each observer for cancellation support.
func (m *MultiObserver) OnEvent(ctx context.Context, event Event) {
	for _, obs := range m.observers {
		obs.OnEvent(ctx, event)
	}
}
