package workflows_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/tailored-agentic-units/kernel/orchestrate/config"
	"github.com/tailored-agentic-units/kernel/observability"
	"github.com/tailored-agentic-units/kernel/orchestrate/workflows"
)

type captureObserver struct {
	events []observability.Event
}

func (o *captureObserver) OnEvent(ctx context.Context, event observability.Event) {
	o.events = append(o.events, event)
}

func newCaptureObserver() *captureObserver {
	return &captureObserver{events: []observability.Event{}}
}

func TestProcessChain_BasicExecution(t *testing.T) {
	ctx := context.Background()
	observer := newCaptureObserver()

	observability.RegisterObserver("capture", observer)

	cfg := config.ChainConfig{
		CaptureIntermediateStates: false,
		Observer:                  "capture",
	}

	items := []string{"a", "b", "c"}
	initial := "start"

	processor := func(ctx context.Context, item string, current string) (string, error) {
		return current + "->" + item, nil
	}

	result, err := workflows.ProcessChain(ctx, cfg, items, initial, processor, nil)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	expected := "start->a->b->c"
	if result.Final != expected {
		t.Errorf("Expected final state %q, got %q", expected, result.Final)
	}

	if result.Steps != 3 {
		t.Errorf("Expected 3 steps, got %d", result.Steps)
	}

	if len(result.Intermediate) != 0 {
		t.Errorf("Expected no intermediate states, got %d", len(result.Intermediate))
	}
}

func TestProcessChain_EmptyChain(t *testing.T) {
	ctx := context.Background()
	cfg := config.DefaultChainConfig()

	items := []string{}
	initial := "initial"

	processor := func(ctx context.Context, item string, current string) (string, error) {
		return current + "->" + item, nil
	}

	result, err := workflows.ProcessChain(ctx, cfg, items, initial, processor, nil)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.Final != initial {
		t.Errorf("Expected final state %q, got %q", initial, result.Final)
	}

	if result.Steps != 0 {
		t.Errorf("Expected 0 steps, got %d", result.Steps)
	}
}

func TestProcessChain_SingleItem(t *testing.T) {
	ctx := context.Background()
	cfg := config.DefaultChainConfig()

	items := []string{"only"}
	initial := "start"

	processor := func(ctx context.Context, item string, current string) (string, error) {
		return current + "->" + item, nil
	}

	result, err := workflows.ProcessChain(ctx, cfg, items, initial, processor, nil)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	expected := "start->only"
	if result.Final != expected {
		t.Errorf("Expected final state %q, got %q", expected, result.Final)
	}

	if result.Steps != 1 {
		t.Errorf("Expected 1 step, got %d", result.Steps)
	}
}

func TestProcessChain_ErrorInMiddle(t *testing.T) {
	ctx := context.Background()
	cfg := config.DefaultChainConfig()

	items := []string{"a", "b", "c", "d"}
	initial := "start"
	expectedError := errors.New("processing failed")

	processor := func(ctx context.Context, item string, current string) (string, error) {
		if item == "c" {
			return current, expectedError
		}
		return current + "->" + item, nil
	}

	result, err := workflows.ProcessChain(ctx, cfg, items, initial, processor, nil)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	chainErr, ok := err.(*workflows.ChainError[string, string])
	if !ok {
		t.Fatalf("Expected ChainError, got %T", err)
	}

	if chainErr.StepIndex != 2 {
		t.Errorf("Expected error at step 2, got %d", chainErr.StepIndex)
	}

	if chainErr.Item != "c" {
		t.Errorf("Expected error item 'c', got %q", chainErr.Item)
	}

	expectedState := "start->a->b"
	if chainErr.State != expectedState {
		t.Errorf("Expected state %q, got %q", expectedState, chainErr.State)
	}

	if !errors.Is(err, expectedError) {
		t.Error("Expected error to unwrap to expectedError")
	}

	if result.Steps != 0 {
		t.Errorf("Expected 0 steps on error, got %d", result.Steps)
	}
}

func TestProcessChain_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cfg := config.DefaultChainConfig()

	items := []string{"a", "b", "c", "d"}
	initial := "start"
	processedCount := 0

	processor := func(ctx context.Context, item string, current string) (string, error) {
		processedCount++
		if item == "b" {
			cancel()
		}
		return current + "->" + item, nil
	}

	result, err := workflows.ProcessChain(ctx, cfg, items, initial, processor, nil)
	if err == nil {
		t.Fatal("Expected cancellation error, got nil")
	}

	chainErr, ok := err.(*workflows.ChainError[string, string])
	if !ok {
		t.Fatalf("Expected ChainError, got %T", err)
	}

	if !errors.Is(chainErr.Err, context.Canceled) {
		t.Error("Expected context.Canceled in error chain")
	}

	if processedCount > 3 {
		t.Errorf("Expected at most 3 items processed before cancellation, got %d", processedCount)
	}

	if result.Steps != 0 {
		t.Errorf("Expected 0 steps on error, got %d", result.Steps)
	}
}

func TestProcessChain_ProgressCallback(t *testing.T) {
	ctx := context.Background()
	cfg := config.DefaultChainConfig()

	items := []string{"a", "b", "c"}
	initial := "start"

	progressCalls := []struct {
		completed int
		total     int
		state     string
	}{}

	progress := func(completed, total int, current string) {
		progressCalls = append(progressCalls, struct {
			completed int
			total     int
			state     string
		}{completed, total, current})
	}

	processor := func(ctx context.Context, item string, current string) (string, error) {
		return current + "->" + item, nil
	}

	_, err := workflows.ProcessChain(ctx, cfg, items, initial, processor, progress)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(progressCalls) != 3 {
		t.Fatalf("Expected 3 progress calls, got %d", len(progressCalls))
	}

	expected := []struct {
		completed int
		total     int
		state     string
	}{
		{1, 3, "start->a"},
		{2, 3, "start->a->b"},
		{3, 3, "start->a->b->c"},
	}

	for i, call := range progressCalls {
		if call.completed != expected[i].completed {
			t.Errorf("Call %d: expected completed=%d, got %d", i, expected[i].completed, call.completed)
		}
		if call.total != expected[i].total {
			t.Errorf("Call %d: expected total=%d, got %d", i, expected[i].total, call.total)
		}
		if call.state != expected[i].state {
			t.Errorf("Call %d: expected state=%q, got %q", i, expected[i].state, call.state)
		}
	}
}

func TestProcessChain_IntermediateStateCapture(t *testing.T) {
	ctx := context.Background()
	cfg := config.ChainConfig{
		CaptureIntermediateStates: true,
		Observer:                  "noop",
	}

	items := []string{"a", "b", "c"}
	initial := "start"

	processor := func(ctx context.Context, item string, current string) (string, error) {
		return current + "->" + item, nil
	}

	result, err := workflows.ProcessChain(ctx, cfg, items, initial, processor, nil)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(result.Intermediate) != 4 {
		t.Fatalf("Expected 4 intermediate states (initial + 3 steps), got %d", len(result.Intermediate))
	}

	expected := []string{
		"start",
		"start->a",
		"start->a->b",
		"start->a->b->c",
	}

	for i, state := range result.Intermediate {
		if state != expected[i] {
			t.Errorf("Intermediate[%d]: expected %q, got %q", i, expected[i], state)
		}
	}
}

func TestProcessChain_ObserverIntegration(t *testing.T) {
	ctx := context.Background()
	observer := newCaptureObserver()

	observability.RegisterObserver("test-observer", observer)

	cfg := config.ChainConfig{
		CaptureIntermediateStates: false,
		Observer:                  "test-observer",
	}

	items := []string{"a", "b"}
	initial := "start"

	processor := func(ctx context.Context, item string, current string) (string, error) {
		return current + "->" + item, nil
	}

	_, err := workflows.ProcessChain(ctx, cfg, items, initial, processor, nil)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	expectedEvents := []observability.EventType{
		workflows.EventChainStart,
		workflows.EventStepStart,
		workflows.EventStepComplete,
		workflows.EventStepStart,
		workflows.EventStepComplete,
		workflows.EventChainComplete,
	}

	if len(observer.events) != len(expectedEvents) {
		t.Fatalf("Expected %d events, got %d", len(expectedEvents), len(observer.events))
	}

	for i, event := range observer.events {
		if event.Type != expectedEvents[i] {
			t.Errorf("Event %d: expected type %v, got %v", i, expectedEvents[i], event.Type)
		}
		if event.Source != "workflows.ProcessChain" {
			t.Errorf("Event %d: expected source 'workflows.ProcessChain', got %q", i, event.Source)
		}
	}
}

func TestProcessChain_ObserverOnError(t *testing.T) {
	ctx := context.Background()
	observer := newCaptureObserver()

	observability.RegisterObserver("error-observer", observer)

	cfg := config.ChainConfig{
		CaptureIntermediateStates: false,
		Observer:                  "error-observer",
	}

	items := []string{"a", "b"}
	initial := "start"

	processor := func(ctx context.Context, item string, current string) (string, error) {
		if item == "b" {
			return current, errors.New("fail")
		}
		return current + "->" + item, nil
	}

	_, err := workflows.ProcessChain(ctx, cfg, items, initial, processor, nil)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	expectedEvents := []observability.EventType{
		workflows.EventChainStart,
		workflows.EventStepStart,
		workflows.EventStepComplete,
		workflows.EventStepStart,
		workflows.EventStepComplete,
		workflows.EventChainComplete,
	}

	if len(observer.events) != len(expectedEvents) {
		t.Fatalf("Expected %d events, got %d", len(expectedEvents), len(observer.events))
	}

	stepCompleteEvent := observer.events[4]
	if errorFlag, ok := stepCompleteEvent.Data["error"].(bool); !ok || !errorFlag {
		t.Error("Expected EventStepComplete to have error=true in Data")
	}

	chainCompleteEvent := observer.events[5]
	if errorFlag, ok := chainCompleteEvent.Data["error"].(bool); !ok || !errorFlag {
		t.Error("Expected EventChainComplete to have error=true in Data")
	}
}

func TestProcessChain_ChainErrorUnwrapping(t *testing.T) {
	ctx := context.Background()
	cfg := config.DefaultChainConfig()

	items := []string{"a"}
	initial := "start"
	baseErr := errors.New("base error")

	processor := func(ctx context.Context, item string, current string) (string, error) {
		return current, fmt.Errorf("wrapped: %w", baseErr)
	}

	_, err := workflows.ProcessChain(ctx, cfg, items, initial, processor, nil)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !errors.Is(err, baseErr) {
		t.Error("Expected error chain to contain baseErr")
	}

	var chainErr *workflows.ChainError[string, string]
	if !errors.As(err, &chainErr) {
		t.Fatal("Expected error to be ChainError")
	}

	if chainErr.StepIndex != 0 {
		t.Errorf("Expected step index 0, got %d", chainErr.StepIndex)
	}
}

func TestProcessChain_GenericTypes(t *testing.T) {
	ctx := context.Background()
	cfg := config.DefaultChainConfig()

	type CustomItem struct {
		Value int
	}

	type CustomContext struct {
		Sum int
	}

	items := []CustomItem{{Value: 1}, {Value: 2}, {Value: 3}}
	initial := CustomContext{Sum: 0}

	processor := func(ctx context.Context, item CustomItem, current CustomContext) (CustomContext, error) {
		return CustomContext{Sum: current.Sum + item.Value}, nil
	}

	result, err := workflows.ProcessChain(ctx, cfg, items, initial, processor, nil)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.Final.Sum != 6 {
		t.Errorf("Expected sum 6, got %d", result.Final.Sum)
	}

	if result.Steps != 3 {
		t.Errorf("Expected 3 steps, got %d", result.Steps)
	}
}

func TestProcessChain_LargeChain(t *testing.T) {
	ctx := context.Background()
	cfg := config.DefaultChainConfig()

	itemCount := 1000
	items := make([]int, itemCount)
	for i := range itemCount {
		items[i] = i
	}

	initial := 0

	processor := func(ctx context.Context, item int, current int) (int, error) {
		return current + item, nil
	}

	start := time.Now()
	result, err := workflows.ProcessChain(ctx, cfg, items, initial, processor, nil)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	expectedSum := (itemCount - 1) * itemCount / 2
	if result.Final != expectedSum {
		t.Errorf("Expected sum %d, got %d", expectedSum, result.Final)
	}

	if result.Steps != itemCount {
		t.Errorf("Expected %d steps, got %d", itemCount, result.Steps)
	}

	if elapsed > 5*time.Second {
		t.Errorf("Processing %d items took too long: %v", itemCount, elapsed)
	}
}

func TestProcessChain_InvalidObserver(t *testing.T) {
	ctx := context.Background()
	cfg := config.ChainConfig{
		CaptureIntermediateStates: false,
		Observer:                  "nonexistent",
	}

	items := []string{"a"}
	initial := "start"

	processor := func(ctx context.Context, item string, current string) (string, error) {
		return current + "->" + item, nil
	}

	_, err := workflows.ProcessChain(ctx, cfg, items, initial, processor, nil)
	if err == nil {
		t.Fatal("Expected observer resolution error, got nil")
	}

	errMsg := err.Error()
	expectedPrefix := "failed to resolve observer:"
	if len(errMsg) < len(expectedPrefix) || errMsg[:len(expectedPrefix)] != expectedPrefix {
		t.Errorf("Expected error to start with %q, got: %v", expectedPrefix, err)
	}
}

func TestProcessChain_NoProgressCallback(t *testing.T) {
	ctx := context.Background()
	cfg := config.DefaultChainConfig()

	items := []string{"a", "b"}
	initial := "start"

	processor := func(ctx context.Context, item string, current string) (string, error) {
		return current + "->" + item, nil
	}

	result, err := workflows.ProcessChain(ctx, cfg, items, initial, processor, nil)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	expected := "start->a->b"
	if result.Final != expected {
		t.Errorf("Expected final state %q, got %q", expected, result.Final)
	}
}
