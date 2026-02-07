package observability_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/tailored-agentic-units/kernel/orchestrate/observability"
)

type captureObserver struct {
	mu     sync.Mutex
	events []observability.Event
}

func (o *captureObserver) OnEvent(ctx context.Context, event observability.Event) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.events = append(o.events, event)
}

func (o *captureObserver) getEvents() []observability.Event {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.events
}

func TestMultiObserver_BroadcastsToAllObservers(t *testing.T) {
	obs1 := &captureObserver{}
	obs2 := &captureObserver{}
	obs3 := &captureObserver{}

	multi := observability.NewMultiObserver(obs1, obs2, obs3)

	event := observability.Event{
		Type:      observability.EventNodeStart,
		Timestamp: time.Now(),
		Source:    "test",
		Data:      map[string]any{"key": "value"},
	}

	multi.OnEvent(context.Background(), event)

	observers := []*captureObserver{obs1, obs2, obs3}
	for i, obs := range observers {
		events := obs.getEvents()
		if len(events) != 1 {
			t.Errorf("Observer %d: got %d events, want 1", i, len(events))
		}
		if events[0].Type != observability.EventNodeStart {
			t.Errorf("Observer %d: got type %v, want %v", i, events[0].Type, observability.EventNodeStart)
		}
	}
}

func TestMultiObserver_EmptyObservers(t *testing.T) {
	multi := observability.NewMultiObserver()

	event := observability.Event{
		Type:      observability.EventNodeStart,
		Timestamp: time.Now(),
		Source:    "test",
	}

	multi.OnEvent(context.Background(), event)
}

func TestMultiObserver_FiltersNilObservers(t *testing.T) {
	obs1 := &captureObserver{}
	obs2 := &captureObserver{}

	multi := observability.NewMultiObserver(obs1, nil, obs2, nil)

	event := observability.Event{
		Type:      observability.EventNodeComplete,
		Timestamp: time.Now(),
		Source:    "test",
	}

	multi.OnEvent(context.Background(), event)

	if len(obs1.getEvents()) != 1 {
		t.Errorf("obs1: got %d events, want 1", len(obs1.getEvents()))
	}
	if len(obs2.getEvents()) != 1 {
		t.Errorf("obs2: got %d events, want 1", len(obs2.getEvents()))
	}
}

func TestMultiObserver_SingleObserver(t *testing.T) {
	obs := &captureObserver{}
	multi := observability.NewMultiObserver(obs)

	event := observability.Event{
		Type:      observability.EventGraphStart,
		Timestamp: time.Now(),
		Source:    "graph",
		Data:      map[string]any{"name": "test-graph"},
	}

	multi.OnEvent(context.Background(), event)

	events := obs.getEvents()
	if len(events) != 1 {
		t.Fatalf("got %d events, want 1", len(events))
	}
	if events[0].Data["name"] != "test-graph" {
		t.Errorf("got name %v, want test-graph", events[0].Data["name"])
	}
}

func TestMultiObserver_PreservesEventData(t *testing.T) {
	obs := &captureObserver{}
	multi := observability.NewMultiObserver(obs)

	originalData := map[string]any{
		"string": "value",
		"number": 42,
		"nested": map[string]any{"inner": "data"},
	}

	event := observability.Event{
		Type:      observability.EventStateSet,
		Timestamp: time.Now(),
		Source:    "state",
		Data:      originalData,
	}

	multi.OnEvent(context.Background(), event)

	events := obs.getEvents()
	if len(events) != 1 {
		t.Fatalf("got %d events, want 1", len(events))
	}

	receivedData := events[0].Data
	if receivedData["string"] != "value" {
		t.Errorf("string: got %v, want value", receivedData["string"])
	}
	if receivedData["number"] != 42 {
		t.Errorf("number: got %v, want 42", receivedData["number"])
	}
}

func TestMultiObserver_PropagatesContext(t *testing.T) {
	type ctxKey string
	key := ctxKey("test-key")

	var receivedCtx context.Context
	obs := &captureObserver{}

	wrapper := &contextCapture{
		inner:      obs,
		capturedFn: func(ctx context.Context) { receivedCtx = ctx },
	}

	multi := observability.NewMultiObserver(wrapper)

	ctx := context.WithValue(context.Background(), key, "test-value")
	event := observability.Event{
		Type:      observability.EventNodeStart,
		Timestamp: time.Now(),
		Source:    "test",
	}

	multi.OnEvent(ctx, event)

	if receivedCtx == nil {
		t.Fatal("context was not propagated")
	}
	if receivedCtx.Value(key) != "test-value" {
		t.Errorf("context value: got %v, want test-value", receivedCtx.Value(key))
	}
}

type contextCapture struct {
	inner      observability.Observer
	capturedFn func(ctx context.Context)
}

func (c *contextCapture) OnEvent(ctx context.Context, event observability.Event) {
	c.capturedFn(ctx)
	c.inner.OnEvent(ctx, event)
}

func TestMultiObserver_ConcurrentEvents(t *testing.T) {
	obs := &captureObserver{}
	multi := observability.NewMultiObserver(obs)

	const numGoroutines = 10
	const eventsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := range numGoroutines {
		go func(id int) {
			defer wg.Done()
			for j := range eventsPerGoroutine {
				event := observability.Event{
					Type:      observability.EventNodeStart,
					Timestamp: time.Now(),
					Source:    "concurrent-test",
					Data:      map[string]any{"goroutine": id, "event": j},
				}
				multi.OnEvent(context.Background(), event)
			}
		}(i)
	}

	wg.Wait()

	events := obs.getEvents()
	expected := numGoroutines * eventsPerGoroutine
	if len(events) != expected {
		t.Errorf("got %d events, want %d", len(events), expected)
	}
}
