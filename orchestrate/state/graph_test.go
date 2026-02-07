package state_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/tailored-agentic-units/kernel/orchestrate/config"
	"github.com/tailored-agentic-units/kernel/orchestrate/observability"
	"github.com/tailored-agentic-units/kernel/orchestrate/state"
)

func newTestNode(key string, value any) state.StateNode {
	return state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
		return s.Set(key, value), nil
	})
}

func newErrorNode(err error) state.StateNode {
	return state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
		return s, err
	})
}

func TestNewGraph(t *testing.T) {
	tests := []struct {
		name        string
		config      config.GraphConfig
		expectError bool
	}{
		{
			name: "valid config with noop observer",
			config: config.GraphConfig{
				Name:          "test-graph",
				Observer:      "noop",
				MaxIterations: 1000,
			},
			expectError: false,
		},
		{
			name: "invalid observer name",
			config: config.GraphConfig{
				Name:          "test-graph",
				Observer:      "invalid",
				MaxIterations: 1000,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			graph, err := state.NewGraph(tt.config)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if graph == nil {
				t.Error("expected graph, got nil")
			}

			if graph.Name() != tt.config.Name {
				t.Errorf("expected name %s, got %s", tt.config.Name, graph.Name())
			}
		})
	}
}

func TestStateGraph_AddNode(t *testing.T) {
	tests := []struct {
		name        string
		nodeName    string
		node        state.StateNode
		expectError bool
	}{
		{
			name:        "valid node",
			nodeName:    "test",
			node:        newTestNode("key", "value"),
			expectError: false,
		},
		{
			name:        "empty name",
			nodeName:    "",
			node:        newTestNode("key", "value"),
			expectError: true,
		},
		{
			name:        "nil node",
			nodeName:    "test",
			node:        nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			graph, err := state.NewGraph(config.DefaultGraphConfig("test"))
			if err != nil {
				t.Fatalf("failed to create graph: %v", err)
			}

			err = graph.AddNode(tt.nodeName, tt.node)

			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestStateGraph_AddNode_Duplicate(t *testing.T) {
	graph, err := state.NewGraph(config.DefaultGraphConfig("test"))
	if err != nil {
		t.Fatalf("failed to create graph: %v", err)
	}

	err = graph.AddNode("test", newTestNode("key", "value"))
	if err != nil {
		t.Fatalf("first AddNode failed: %v", err)
	}

	err = graph.AddNode("test", newTestNode("key", "value"))
	if err == nil {
		t.Error("expected duplicate node error, got nil")
	}
}

func TestStateGraph_AddEdge(t *testing.T) {
	graph, err := state.NewGraph(config.DefaultGraphConfig("test"))
	if err != nil {
		t.Fatalf("failed to create graph: %v", err)
	}

	graph.AddNode("a", newTestNode("step", "a"))
	graph.AddNode("b", newTestNode("step", "b"))

	tests := []struct {
		name        string
		from        string
		to          string
		predicate   state.TransitionPredicate
		expectError bool
	}{
		{
			name:        "valid edge",
			from:        "a",
			to:          "b",
			predicate:   nil,
			expectError: false,
		},
		{
			name:        "empty from",
			from:        "",
			to:          "b",
			predicate:   nil,
			expectError: true,
		},
		{
			name:        "empty to",
			from:        "a",
			to:          "",
			predicate:   nil,
			expectError: true,
		},
		{
			name:        "missing from node",
			from:        "missing",
			to:          "b",
			predicate:   nil,
			expectError: true,
		},
		{
			name:        "missing to node",
			from:        "a",
			to:          "missing",
			predicate:   nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := graph.AddEdge(tt.from, tt.to, tt.predicate)

			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestStateGraph_SetEntryPoint(t *testing.T) {
	graph, err := state.NewGraph(config.DefaultGraphConfig("test"))
	if err != nil {
		t.Fatalf("failed to create graph: %v", err)
	}

	graph.AddNode("start", newTestNode("step", "start"))
	graph.AddNode("other", newTestNode("step", "other"))

	err = graph.SetEntryPoint("start")
	if err != nil {
		t.Errorf("SetEntryPoint failed: %v", err)
	}

	err = graph.SetEntryPoint("other")
	if err == nil {
		t.Error("expected duplicate entry point error, got nil")
	}

	graph2, _ := state.NewGraph(config.DefaultGraphConfig("test2"))
	err = graph2.SetEntryPoint("missing")
	if err == nil {
		t.Error("expected missing node error, got nil")
	}
}

func TestStateGraph_SetExitPoint(t *testing.T) {
	graph, err := state.NewGraph(config.DefaultGraphConfig("test"))
	if err != nil {
		t.Fatalf("failed to create graph: %v", err)
	}

	graph.AddNode("end1", newTestNode("step", "end1"))
	graph.AddNode("end2", newTestNode("step", "end2"))

	err = graph.SetExitPoint("end1")
	if err != nil {
		t.Errorf("SetExitPoint failed: %v", err)
	}

	err = graph.SetExitPoint("end2")
	if err != nil {
		t.Errorf("SetExitPoint (second) failed: %v", err)
	}

	err = graph.SetExitPoint("missing")
	if err == nil {
		t.Error("expected missing node error, got nil")
	}
}

func TestStateGraph_Execute_LinearPath(t *testing.T) {
	graph, err := state.NewGraph(config.DefaultGraphConfig("test"))
	if err != nil {
		t.Fatalf("failed to create graph: %v", err)
	}

	graph.AddNode("a", newTestNode("a", "executed"))
	graph.AddNode("b", newTestNode("b", "executed"))
	graph.AddNode("c", newTestNode("c", "executed"))
	graph.AddEdge("a", "b", nil)
	graph.AddEdge("b", "c", nil)
	graph.SetEntryPoint("a")
	graph.SetExitPoint("c")

	ctx := context.Background()
	initialState := state.New(observability.NoOpObserver{})

	finalState, err := graph.Execute(ctx, initialState)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if val, _ := finalState.Get("a"); val != "executed" {
		t.Error("node a did not execute")
	}
	if val, _ := finalState.Get("b"); val != "executed" {
		t.Error("node b did not execute")
	}
	if val, _ := finalState.Get("c"); val != "executed" {
		t.Error("node c did not execute")
	}
}

func TestStateGraph_Execute_ConditionalRouting(t *testing.T) {
	tests := []struct {
		name         string
		initialValue string
		expectedPath string
	}{
		{
			name:         "route to B",
			initialValue: "go-b",
			expectedPath: "b",
		},
		{
			name:         "route to C",
			initialValue: "go-c",
			expectedPath: "c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			graph, _ := state.NewGraph(config.DefaultGraphConfig("test"))

			graph.AddNode("a", newTestNode("step", "a"))
			graph.AddNode("b", newTestNode("result", "b"))
			graph.AddNode("c", newTestNode("result", "c"))

			graph.AddEdge("a", "b", state.KeyEquals("condition", "go-b"))
			graph.AddEdge("a", "c", state.KeyEquals("condition", "go-c"))

			graph.SetEntryPoint("a")
			graph.SetExitPoint("b")
			graph.SetExitPoint("c")

			ctx := context.Background()
			initialState := state.New(observability.NoOpObserver{})
			initialState = initialState.Set("condition", tt.initialValue)

			finalState, err := graph.Execute(ctx, initialState)
			if err != nil {
				t.Fatalf("execution failed: %v", err)
			}

			result, _ := finalState.Get("result")
			if result != tt.expectedPath {
				t.Errorf("expected path %s, got %v", tt.expectedPath, result)
			}
		})
	}
}

func TestStateGraph_Execute_Cycle(t *testing.T) {
	observer := &captureObserver{}
	observability.RegisterObserver("cycle-capture", observer)

	cfg := config.GraphConfig{
		Name:          "cycle-test",
		Observer:      "cycle-capture",
		MaxIterations: 100,
	}

	graph, err := state.NewGraph(cfg)
	if err != nil {
		t.Fatalf("failed to create graph: %v", err)
	}

	counter := 0
	graph.AddNode("a", newTestNode("step", "a"))
	graph.AddNode("b", state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
		counter++
		return s.Set("b-count", counter), nil
	}))
	graph.AddNode("c", newTestNode("step", "c"))
	graph.AddNode("exit", newTestNode("step", "exit"))

	graph.AddEdge("a", "b", nil)
	graph.AddEdge("b", "c", nil)
	graph.AddEdge("c", "b", func(s state.State) bool {
		count, _ := s.Get("b-count")
		if count == nil {
			return false
		}
		return count.(int) < 2
	})
	graph.AddEdge("c", "exit", func(s state.State) bool {
		count, _ := s.Get("b-count")
		if count == nil {
			return false
		}
		return count.(int) >= 2
	})

	graph.SetEntryPoint("a")
	graph.SetExitPoint("exit")

	ctx := context.Background()
	initialState := state.New(observability.NoOpObserver{})

	finalState, err := graph.Execute(ctx, initialState)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	count, _ := finalState.Get("b-count")
	if count != 2 {
		t.Errorf("expected b-count=2, got %v", count)
	}

	cycleEvents := 0
	for _, event := range observer.events {
		if event.Type == observability.EventCycleDetected {
			cycleEvents++
		}
	}

	if cycleEvents == 0 {
		t.Error("expected EventCycleDetected, got none")
	}
}

func TestStateGraph_Execute_MaxIterations(t *testing.T) {
	cfg := config.GraphConfig{
		Name:          "iteration-test",
		Observer:      "noop",
		MaxIterations: 5,
	}

	graph, err := state.NewGraph(cfg)
	if err != nil {
		t.Fatalf("failed to create graph: %v", err)
	}

	graph.AddNode("a", newTestNode("step", "a"))
	graph.AddNode("b", newTestNode("step", "b"))
	graph.AddNode("never-reached", newTestNode("step", "never-reached"))
	graph.AddEdge("a", "b", nil)
	graph.AddEdge("b", "a", nil)
	graph.SetEntryPoint("a")
	graph.SetExitPoint("never-reached")

	ctx := context.Background()
	initialState := state.New(observability.NoOpObserver{})

	_, err = graph.Execute(ctx, initialState)
	if err == nil {
		t.Fatal("expected max iterations error, got nil")
	}

	var execErr *state.ExecutionError
	if !errors.As(err, &execErr) {
		t.Fatalf("expected ExecutionError, got %T", err)
	}

	if execErr.Err == nil {
		t.Error("ExecutionError.Err is nil")
	}
}

func TestStateGraph_Execute_ContextCancellation(t *testing.T) {
	graph, err := state.NewGraph(config.DefaultGraphConfig("test"))
	if err != nil {
		t.Fatalf("failed to create graph: %v", err)
	}

	graph.AddNode("slow", state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
		time.Sleep(100 * time.Millisecond)
		return s, nil
	}))
	graph.SetEntryPoint("slow")
	graph.SetExitPoint("slow")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	initialState := state.New(observability.NoOpObserver{})

	_, err = graph.Execute(ctx, initialState)
	if err == nil {
		t.Fatal("expected context cancellation error, got nil")
	}

	var execErr *state.ExecutionError
	if !errors.As(err, &execErr) {
		t.Fatalf("expected ExecutionError, got %T", err)
	}
}

func TestStateGraph_Execute_NodeError(t *testing.T) {
	graph, err := state.NewGraph(config.DefaultGraphConfig("test"))
	if err != nil {
		t.Fatalf("failed to create graph: %v", err)
	}

	expectedErr := fmt.Errorf("node failed")

	graph.AddNode("start", newTestNode("step", "start"))
	graph.AddNode("fail", newErrorNode(expectedErr))
	graph.AddEdge("start", "fail", nil)
	graph.SetEntryPoint("start")
	graph.SetExitPoint("fail")

	ctx := context.Background()
	initialState := state.New(observability.NoOpObserver{})

	_, err = graph.Execute(ctx, initialState)
	if err == nil {
		t.Fatal("expected node error, got nil")
	}

	var execErr *state.ExecutionError
	if !errors.As(err, &execErr) {
		t.Fatalf("expected ExecutionError, got %T", err)
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("expected wrapped error, got %v", err)
	}

	if execErr.NodeName != "fail" {
		t.Errorf("expected NodeName='fail', got '%s'", execErr.NodeName)
	}

	if len(execErr.Path) == 0 {
		t.Error("expected non-empty Path")
	}
}

func TestStateGraph_Execute_NoValidTransition(t *testing.T) {
	graph, err := state.NewGraph(config.DefaultGraphConfig("test"))
	if err != nil {
		t.Fatalf("failed to create graph: %v", err)
	}

	graph.AddNode("start", newTestNode("step", "start"))
	graph.AddNode("end", newTestNode("step", "end"))

	graph.AddEdge("start", "end", func(s state.State) bool {
		return false
	})

	graph.SetEntryPoint("start")
	graph.SetExitPoint("end")

	ctx := context.Background()
	initialState := state.New(observability.NoOpObserver{})

	_, err = graph.Execute(ctx, initialState)
	if err == nil {
		t.Fatal("expected no valid transition error, got nil")
	}

	var execErr *state.ExecutionError
	if !errors.As(err, &execErr) {
		t.Fatalf("expected ExecutionError, got %T", err)
	}
}

func TestStateGraph_Execute_MultipleExitPoints(t *testing.T) {
	tests := []struct {
		name         string
		condition    string
		expectedExit string
	}{
		{
			name:         "success exit",
			condition:    "success",
			expectedExit: "success",
		},
		{
			name:         "failure exit",
			condition:    "failure",
			expectedExit: "failure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			graph, _ := state.NewGraph(config.DefaultGraphConfig("test"))

			graph.AddNode("start", newTestNode("step", "start"))
			graph.AddNode("success", newTestNode("exit", "success"))
			graph.AddNode("failure", newTestNode("exit", "failure"))

			graph.AddEdge("start", "success", state.KeyEquals("condition", "success"))
			graph.AddEdge("start", "failure", state.KeyEquals("condition", "failure"))

			graph.SetEntryPoint("start")
			graph.SetExitPoint("success")
			graph.SetExitPoint("failure")

			ctx := context.Background()
			initialState := state.New(observability.NoOpObserver{})
			initialState = initialState.Set("condition", tt.condition)

			finalState, err := graph.Execute(ctx, initialState)
			if err != nil {
				t.Fatalf("execution failed: %v", err)
			}

			exit, _ := finalState.Get("exit")
			if exit != tt.expectedExit {
				t.Errorf("expected exit %s, got %v", tt.expectedExit, exit)
			}
		})
	}
}

func TestStateGraph_Execute_ObserverEvents(t *testing.T) {
	observer := &captureObserver{}
	observability.RegisterObserver("test-capture", observer)

	cfg := config.GraphConfig{
		Name:          "observer-test",
		Observer:      "test-capture",
		MaxIterations: 1000,
	}

	graph, err := state.NewGraph(cfg)
	if err != nil {
		t.Fatalf("failed to create graph: %v", err)
	}

	graph.AddNode("a", newTestNode("step", "a"))
	graph.AddNode("b", newTestNode("step", "b"))
	graph.AddEdge("a", "b", nil)
	graph.SetEntryPoint("a")
	graph.SetExitPoint("b")

	ctx := context.Background()
	initialState := state.New(observability.NoOpObserver{})

	_, err = graph.Execute(ctx, initialState)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	expectedEvents := []observability.EventType{
		observability.EventGraphStart,
		observability.EventNodeStart,
		observability.EventNodeComplete,
		observability.EventEdgeEvaluate,
		observability.EventEdgeTransition,
		observability.EventNodeStart,
		observability.EventNodeComplete,
		observability.EventGraphComplete,
	}

	if len(observer.events) != len(expectedEvents) {
		t.Errorf("expected %d events, got %d", len(expectedEvents), len(observer.events))
	}

	for i, expected := range expectedEvents {
		if i >= len(observer.events) {
			break
		}
		if observer.events[i].Type != expected {
			t.Errorf("event %d: expected %s, got %s", i, expected, observer.events[i].Type)
		}
	}
}

func TestStateGraph_Validate(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() state.StateGraph
		expectError bool
	}{
		{
			name: "valid graph",
			setup: func() state.StateGraph {
				g, _ := state.NewGraph(config.DefaultGraphConfig("test"))
				g.AddNode("start", newTestNode("step", "start"))
				g.AddNode("end", newTestNode("step", "end"))
				g.AddEdge("start", "end", nil)
				g.SetEntryPoint("start")
				g.SetExitPoint("end")
				return g
			},
			expectError: false,
		},
		{
			name: "no nodes",
			setup: func() state.StateGraph {
				g, _ := state.NewGraph(config.DefaultGraphConfig("test"))
				return g
			},
			expectError: true,
		},
		{
			name: "no entry point",
			setup: func() state.StateGraph {
				g, _ := state.NewGraph(config.DefaultGraphConfig("test"))
				g.AddNode("node", newTestNode("step", "node"))
				return g
			},
			expectError: true,
		},
		{
			name: "no exit points",
			setup: func() state.StateGraph {
				g, _ := state.NewGraph(config.DefaultGraphConfig("test"))
				g.AddNode("node", newTestNode("step", "node"))
				g.SetEntryPoint("node")
				return g
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			graph := tt.setup()

			ctx := context.Background()
			initialState := state.New(observability.NoOpObserver{})
			_, err := graph.Execute(ctx, initialState)

			if tt.expectError {
				if err == nil {
					t.Error("expected validation error, got nil")
				}
			}
		})
	}
}

func TestExecutionError_Unwrap(t *testing.T) {
	originalErr := fmt.Errorf("original error")
	execErr := &state.ExecutionError{
		NodeName: "test",
		State:    state.New(observability.NoOpObserver{}),
		Path:     []string{"a", "b"},
		Err:      originalErr,
	}

	if !errors.Is(execErr, originalErr) {
		t.Error("ExecutionError.Unwrap() not working correctly")
	}
}

func TestExecutionError_Error(t *testing.T) {
	originalErr := fmt.Errorf("original error")
	execErr := &state.ExecutionError{
		NodeName: "test-node",
		State:    state.New(observability.NoOpObserver{}),
		Path:     []string{"a", "b"},
		Err:      originalErr,
	}

	errorMsg := execErr.Error()
	if errorMsg == "" {
		t.Error("expected non-empty error message")
	}

	if !contains(errorMsg, "test-node") {
		t.Error("error message should contain node name")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsAtIndex(s, substr))
}

func containsAtIndex(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
