package observability_test

import (
	"context"
	"testing"
	"time"

	"github.com/tailored-agentic-units/kernel/orchestrate/observability"
)

func TestObserver_NoOpObserver(t *testing.T) {
	observer := observability.NoOpObserver{}
	event := observability.Event{
		Type:      observability.EventStateCreate,
		Timestamp: time.Now(),
		Source:    "test",
		Data:      map[string]any{"key": "value"},
	}

	observer.OnEvent(context.Background(), event)
}

func TestObserverRegistry_GetObserver(t *testing.T) {
	tests := []struct {
		name        string
		observerKey string
		wantErr     bool
	}{
		{
			name:        "noop observer exists",
			observerKey: "noop",
			wantErr:     false,
		},
		{
			name:        "unknown observer returns error",
			observerKey: "unknown",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			observer, err := observability.GetObserver(tt.observerKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetObserver() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && observer == nil {
				t.Error("GetObserver() returned nil observer for valid key")
			}
		})
	}
}

type testObserver struct{}

func (testObserver) OnEvent(ctx context.Context, event observability.Event) {}

func TestObserverRegistry_RegisterObserver(t *testing.T) {
	observability.RegisterObserver("test-observer", testObserver{})

	observer, err := observability.GetObserver("test-observer")
	if err != nil {
		t.Errorf("GetObserver() after registration failed: %v", err)
	}
	if observer == nil {
		t.Error("GetObserver() returned nil for registered observer")
	}
}

func TestEvent_Structure(t *testing.T) {
	now := time.Now()
	event := observability.Event{
		Type:      observability.EventStateSet,
		Timestamp: now,
		Source:    "test-source",
		Data:      map[string]any{"key": "test-key"},
	}

	if event.Type != observability.EventStateSet {
		t.Errorf("Event.Type = %v, want %v", event.Type, observability.EventStateSet)
	}
	if event.Source != "test-source" {
		t.Errorf("Event.Source = %v, want %v", event.Source, "test-source")
	}
	if event.Data["key"] != "test-key" {
		t.Errorf("Event.Data[key] = %v, want %v", event.Data["key"], "test-key")
	}
}

func TestEventType_Constants(t *testing.T) {
	eventTypes := []observability.EventType{
		observability.EventStateCreate,
		observability.EventStateClone,
		observability.EventStateSet,
		observability.EventStateMerge,
		observability.EventGraphStart,
		observability.EventGraphComplete,
		observability.EventNodeStart,
		observability.EventNodeComplete,
		observability.EventEdgeEvaluate,
		observability.EventEdgeTransition,
		observability.EventCycleDetected,
		observability.EventChainStart,
		observability.EventChainComplete,
		observability.EventStepStart,
		observability.EventStepComplete,
		observability.EventParallelStart,
		observability.EventParallelComplete,
		observability.EventWorkerStart,
		observability.EventWorkerComplete,
		observability.EventCheckpointSave,
		observability.EventCheckpointLoad,
		observability.EventCheckpointResume,
		observability.EventRouteEvaluate,
		observability.EventRouteSelect,
		observability.EventRouteExecute,
	}

	for _, et := range eventTypes {
		if et == "" {
			t.Errorf("EventType constant is empty string")
		}
	}
}
