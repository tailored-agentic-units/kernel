package observability_test

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/tailored-agentic-units/kernel/orchestrate/observability"
)

func TestSlogObserver_OnEvent_LogsEventFields(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	observer := observability.NewSlogObserver(logger)

	ctx := context.Background()
	event := observability.Event{
		Type:      observability.EventChainStart,
		Timestamp: time.Now(),
		Source:    "test.source",
		Data: map[string]any{
			"item_count": 5,
			"test_key":   "test_value",
		},
	}

	observer.OnEvent(ctx, event)

	output := buf.String()
	if !strings.Contains(output, "Event") {
		t.Error("Expected log message to contain 'Event'")
	}
	if !strings.Contains(output, "chain.start") {
		t.Error("Expected log to contain event type 'chain.start'")
	}
	if !strings.Contains(output, "test.source") {
		t.Error("Expected log to contain source 'test.source'")
	}
	if !strings.Contains(output, "item_count") {
		t.Error("Expected log to contain data field 'item_count'")
	}
}

func TestSlogObserver_OnEvent_HandlesAllEventTypes(t *testing.T) {
	tests := []struct {
		name      string
		eventType observability.EventType
	}{
		{"StateCreate", observability.EventStateCreate},
		{"StateClone", observability.EventStateClone},
		{"StateSet", observability.EventStateSet},
		{"StateMerge", observability.EventStateMerge},
		{"GraphStart", observability.EventGraphStart},
		{"GraphComplete", observability.EventGraphComplete},
		{"NodeStart", observability.EventNodeStart},
		{"NodeComplete", observability.EventNodeComplete},
		{"EdgeEvaluate", observability.EventEdgeEvaluate},
		{"EdgeTransition", observability.EventEdgeTransition},
		{"CycleDetected", observability.EventCycleDetected},
		{"ChainStart", observability.EventChainStart},
		{"ChainComplete", observability.EventChainComplete},
		{"StepStart", observability.EventStepStart},
		{"StepComplete", observability.EventStepComplete},
		{"ParallelStart", observability.EventParallelStart},
		{"ParallelComplete", observability.EventParallelComplete},
		{"WorkerStart", observability.EventWorkerStart},
		{"WorkerComplete", observability.EventWorkerComplete},
		{"CheckpointSave", observability.EventCheckpointSave},
		{"CheckpointLoad", observability.EventCheckpointLoad},
		{"CheckpointResume", observability.EventCheckpointResume},
		{"RouteEvaluate", observability.EventRouteEvaluate},
		{"RouteSelect", observability.EventRouteSelect},
		{"RouteExecute", observability.EventRouteExecute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := slog.New(slog.NewTextHandler(&buf, nil))
			observer := observability.NewSlogObserver(logger)

			ctx := context.Background()
			event := observability.Event{
				Type:      tt.eventType,
				Timestamp: time.Now(),
				Source:    "test",
				Data:      map[string]any{},
			}

			observer.OnEvent(ctx, event)

			output := buf.String()
			if !strings.Contains(output, string(tt.eventType)) {
				t.Errorf("Expected log to contain event type %q", tt.eventType)
			}
		})
	}
}

func TestSlogObserver_OnEvent_HandlesEmptyData(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	observer := observability.NewSlogObserver(logger)

	ctx := context.Background()
	event := observability.Event{
		Type:      observability.EventChainStart,
		Timestamp: time.Now(),
		Source:    "test",
		Data:      map[string]any{},
	}

	observer.OnEvent(ctx, event)

	output := buf.String()
	if !strings.Contains(output, "Event") {
		t.Error("Expected log message even with empty data")
	}
}

func TestSlogObserver_OnEvent_HandlesNilData(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	observer := observability.NewSlogObserver(logger)

	ctx := context.Background()
	event := observability.Event{
		Type:      observability.EventChainStart,
		Timestamp: time.Now(),
		Source:    "test",
		Data:      nil,
	}

	observer.OnEvent(ctx, event)

	output := buf.String()
	if !strings.Contains(output, "Event") {
		t.Error("Expected log message even with nil data")
	}
}

func TestSlogObserver_OnEvent_HandlesComplexDataTypes(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	observer := observability.NewSlogObserver(logger)

	ctx := context.Background()
	event := observability.Event{
		Type:      observability.EventParallelStart,
		Timestamp: time.Now(),
		Source:    "workflows.ProcessParallel",
		Data: map[string]any{
			"item_count":   100,
			"worker_count": 8,
			"fail_fast":    true,
			"nested": map[string]any{
				"key1": "value1",
				"key2": 42,
			},
			"slice": []string{"a", "b", "c"},
		},
	}

	observer.OnEvent(ctx, event)

	output := buf.String()
	if !strings.Contains(output, "item_count") {
		t.Error("Expected log to contain 'item_count'")
	}
	if !strings.Contains(output, "worker_count") {
		t.Error("Expected log to contain 'worker_count'")
	}
}

func TestSlogObserver_OnEvent_PropagatesContext(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	observer := observability.NewSlogObserver(logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	event := observability.Event{
		Type:      observability.EventChainStart,
		Timestamp: time.Now(),
		Source:    "test",
		Data:      map[string]any{},
	}

	observer.OnEvent(ctx, event)

	output := buf.String()
	if len(output) == 0 {
		t.Error("Expected log output with valid context")
	}
}

func TestSlogObserver_JSONHandler(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	observer := observability.NewSlogObserver(logger)

	ctx := context.Background()
	event := observability.Event{
		Type:      observability.EventWorkerStart,
		Timestamp: time.Now(),
		Source:    "workflows.ProcessParallel",
		Data: map[string]any{
			"worker_id":  3,
			"item_index": 42,
		},
	}

	observer.OnEvent(ctx, event)

	output := buf.String()
	if !strings.Contains(output, "{") {
		t.Error("Expected JSON output")
	}
	if !strings.Contains(output, "worker.start") {
		t.Error("Expected event type in JSON output")
	}
	if !strings.Contains(output, "worker_id") {
		t.Error("Expected data fields in JSON output")
	}
}

func TestNewSlogObserver_WithDefaultLogger(t *testing.T) {
	observer := observability.NewSlogObserver(slog.Default())

	ctx := context.Background()
	event := observability.Event{
		Type:      observability.EventChainStart,
		Timestamp: time.Now(),
		Source:    "test",
		Data:      map[string]any{},
	}

	observer.OnEvent(ctx, event)
}

func TestSlogObserver_ConcurrentEvents(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	observer := observability.NewSlogObserver(logger)

	ctx := context.Background()
	done := make(chan struct{})
	eventCount := 100

	for i := range eventCount {
		go func(id int) {
			event := observability.Event{
				Type:      observability.EventWorkerStart,
				Timestamp: time.Now(),
				Source:    "test",
				Data: map[string]any{
					"worker_id": id,
				},
			}
			observer.OnEvent(ctx, event)
			done <- struct{}{}
		}(i)
	}

	for range eventCount {
		<-done
	}

	output := buf.String()
	if len(output) == 0 {
		t.Error("Expected log output from concurrent events")
	}
}
