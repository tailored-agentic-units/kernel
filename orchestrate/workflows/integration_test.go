package workflows_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/tailored-agentic-units/kernel/orchestrate/config"
	"github.com/tailored-agentic-units/kernel/orchestrate/observability"
	"github.com/tailored-agentic-units/kernel/orchestrate/state"
	"github.com/tailored-agentic-units/kernel/orchestrate/workflows"
)

func TestChainNode_InStateGraph(t *testing.T) {
	observability.RegisterObserver("noop", &observability.NoOpObserver{})

	ctx := context.Background()

	graphCfg := config.DefaultGraphConfig("test-chain-node")
	graphCfg.Observer = "noop"

	chainCfg := config.DefaultChainConfig()
	chainCfg.Observer = "noop"

	graph, err := state.NewGraph(graphCfg)
	if err != nil {
		t.Fatalf("Failed to create graph: %v", err)
	}

	items := []string{"item1", "item2", "item3"}

	processor := func(ctx context.Context, item string, s state.State) (state.State, error) {
		count, _ := s.Get("count")
		if count == nil {
			count = 0
		}
		newCount := count.(int) + 1
		return s.Set("count", newCount).Set("last_item", item), nil
	}

	chainNode := workflows.ChainNode(chainCfg, items, processor, nil)

	if err := graph.AddNode("chain", chainNode); err != nil {
		t.Fatalf("Failed to add chain node: %v", err)
	}

	if err := graph.SetEntryPoint("chain"); err != nil {
		t.Fatalf("Failed to set entry point: %v", err)
	}

	if err := graph.SetExitPoint("chain"); err != nil {
		t.Fatalf("Failed to set exit point: %v", err)
	}

	initialState := state.New(nil)

	finalState, err := graph.Execute(ctx, initialState)
	if err != nil {
		t.Fatalf("Graph execution failed: %v", err)
	}

	count, ok := finalState.Get("count")
	if !ok {
		t.Fatal("Expected 'count' key in final state")
	}

	if count.(int) != 3 {
		t.Errorf("count = %d, want 3", count.(int))
	}

	lastItem, ok := finalState.Get("last_item")
	if !ok {
		t.Fatal("Expected 'last_item' key in final state")
	}

	if lastItem.(string) != "item3" {
		t.Errorf("last_item = %s, want item3", lastItem.(string))
	}
}

func TestParallelNode_InStateGraph(t *testing.T) {
	observability.RegisterObserver("noop", &observability.NoOpObserver{})

	ctx := context.Background()

	graphCfg := config.DefaultGraphConfig("test-parallel-node")
	graphCfg.Observer = "noop"

	parallelCfg := config.DefaultParallelConfig()
	parallelCfg.Observer = "noop"

	graph, err := state.NewGraph(graphCfg)
	if err != nil {
		t.Fatalf("Failed to create graph: %v", err)
	}

	items := []int{1, 2, 3, 4}

	processor := func(ctx context.Context, item int) (int, error) {
		return item * 2, nil
	}

	aggregator := func(results []int, currentState state.State) state.State {
		sum := 0
		for _, r := range results {
			sum += r
		}
		return currentState.Set("sum", sum).Set("count", len(results))
	}

	parallelNode := workflows.ParallelNode(parallelCfg, items, processor, nil, aggregator)

	if err := graph.AddNode("parallel", parallelNode); err != nil {
		t.Fatalf("Failed to add parallel node: %v", err)
	}

	if err := graph.SetEntryPoint("parallel"); err != nil {
		t.Fatalf("Failed to set entry point: %v", err)
	}

	if err := graph.SetExitPoint("parallel"); err != nil {
		t.Fatalf("Failed to set exit point: %v", err)
	}

	initialState := state.New(nil)

	finalState, err := graph.Execute(ctx, initialState)
	if err != nil {
		t.Fatalf("Graph execution failed: %v", err)
	}

	sum, ok := finalState.Get("sum")
	if !ok {
		t.Fatal("Expected 'sum' key in final state")
	}

	expectedSum := (1 + 2 + 3 + 4) * 2
	if sum.(int) != expectedSum {
		t.Errorf("sum = %d, want %d", sum.(int), expectedSum)
	}

	count, ok := finalState.Get("count")
	if !ok {
		t.Fatal("Expected 'count' key in final state")
	}

	if count.(int) != 4 {
		t.Errorf("count = %d, want 4", count.(int))
	}
}

func TestConditionalNode_InStateGraph(t *testing.T) {
	observability.RegisterObserver("noop", &observability.NoOpObserver{})

	ctx := context.Background()

	graphCfg := config.DefaultGraphConfig("test-conditional-node")
	graphCfg.Observer = "noop"

	conditionalCfg := config.DefaultConditionalConfig()
	conditionalCfg.Observer = "noop"

	graph, err := state.NewGraph(graphCfg)
	if err != nil {
		t.Fatalf("Failed to create graph: %v", err)
	}

	predicate := func(s state.State) (string, error) {
		value, ok := s.Get("value")
		if !ok {
			return "default", nil
		}
		if value.(int) > 50 {
			return "high", nil
		}
		return "low", nil
	}

	routes := workflows.Routes[state.State]{
		Handlers: map[string]workflows.RouteHandler[state.State]{
			"high": func(ctx context.Context, s state.State) (state.State, error) {
				return s.Set("priority", "high"), nil
			},
			"low": func(ctx context.Context, s state.State) (state.State, error) {
				return s.Set("priority", "low"), nil
			},
		},
		Default: func(ctx context.Context, s state.State) (state.State, error) {
			return s.Set("priority", "default"), nil
		},
	}

	conditionalNode := workflows.ConditionalNode(conditionalCfg, predicate, routes)

	if err := graph.AddNode("decision", conditionalNode); err != nil {
		t.Fatalf("Failed to add conditional node: %v", err)
	}

	if err := graph.SetEntryPoint("decision"); err != nil {
		t.Fatalf("Failed to set entry point: %v", err)
	}

	if err := graph.SetExitPoint("decision"); err != nil {
		t.Fatalf("Failed to set exit point: %v", err)
	}

	tests := []struct {
		name          string
		initialValue  int
		wantPriority  string
	}{
		{
			name:         "high_priority",
			initialValue: 75,
			wantPriority: "high",
		},
		{
			name:         "low_priority",
			initialValue: 25,
			wantPriority: "low",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initialState := state.New(nil).Set("value", tt.initialValue)

			finalState, err := graph.Execute(ctx, initialState)
			if err != nil {
				t.Fatalf("Graph execution failed: %v", err)
			}

			priority, ok := finalState.Get("priority")
			if !ok {
				t.Fatal("Expected 'priority' key in final state")
			}

			if priority.(string) != tt.wantPriority {
				t.Errorf("priority = %s, want %s", priority.(string), tt.wantPriority)
			}
		})
	}
}

func TestIntegration_AllHelpers(t *testing.T) {
	observability.RegisterObserver("noop", &observability.NoOpObserver{})

	ctx := context.Background()

	graphCfg := config.DefaultGraphConfig("test-all-helpers")
	graphCfg.Observer = "noop"

	chainCfg := config.DefaultChainConfig()
	chainCfg.Observer = "noop"

	parallelCfg := config.DefaultParallelConfig()
	parallelCfg.Observer = "noop"

	conditionalCfg := config.DefaultConditionalConfig()
	conditionalCfg.Observer = "noop"

	graph, err := state.NewGraph(graphCfg)
	if err != nil {
		t.Fatalf("Failed to create graph: %v", err)
	}

	chainItems := []string{"a", "b", "c"}
	chainProcessor := func(ctx context.Context, item string, s state.State) (state.State, error) {
		count, _ := s.Get("chain_count")
		if count == nil {
			count = 0
		}
		return s.Set("chain_count", count.(int)+1), nil
	}
	chainNode := workflows.ChainNode(chainCfg, chainItems, chainProcessor, nil)

	parallelItems := []int{1, 2, 3}
	parallelProcessor := func(ctx context.Context, item int) (int, error) {
		return item * 2, nil
	}
	parallelAggregator := func(results []int, currentState state.State) state.State {
		sum := 0
		for _, r := range results {
			sum += r
		}
		return currentState.Set("parallel_sum", sum)
	}
	parallelNode := workflows.ParallelNode(parallelCfg, parallelItems, parallelProcessor, nil, parallelAggregator)

	conditionalPredicate := func(s state.State) (string, error) {
		sum, ok := s.Get("parallel_sum")
		if !ok {
			return "unknown", nil
		}
		if sum.(int) > 10 {
			return "high", nil
		}
		return "low", nil
	}
	conditionalRoutes := workflows.Routes[state.State]{
		Handlers: map[string]workflows.RouteHandler[state.State]{
			"high": func(ctx context.Context, s state.State) (state.State, error) {
				return s.Set("result", "success"), nil
			},
			"low": func(ctx context.Context, s state.State) (state.State, error) {
				return s.Set("result", "failure"), nil
			},
		},
	}
	conditionalNode := workflows.ConditionalNode(conditionalCfg, conditionalPredicate, conditionalRoutes)

	if err := graph.AddNode("chain", chainNode); err != nil {
		t.Fatalf("Failed to add chain node: %v", err)
	}

	if err := graph.AddNode("parallel", parallelNode); err != nil {
		t.Fatalf("Failed to add parallel node: %v", err)
	}

	if err := graph.AddNode("decision", conditionalNode); err != nil {
		t.Fatalf("Failed to add conditional node: %v", err)
	}

	if err := graph.AddEdge("chain", "parallel", nil); err != nil {
		t.Fatalf("Failed to add edge: %v", err)
	}

	if err := graph.AddEdge("parallel", "decision", nil); err != nil {
		t.Fatalf("Failed to add edge: %v", err)
	}

	if err := graph.SetEntryPoint("chain"); err != nil {
		t.Fatalf("Failed to set entry point: %v", err)
	}

	if err := graph.SetExitPoint("decision"); err != nil {
		t.Fatalf("Failed to set exit point: %v", err)
	}

	initialState := state.New(nil)

	finalState, err := graph.Execute(ctx, initialState)
	if err != nil {
		t.Fatalf("Graph execution failed: %v", err)
	}

	chainCount, ok := finalState.Get("chain_count")
	if !ok || chainCount.(int) != 3 {
		t.Errorf("chain_count = %v, want 3", chainCount)
	}

	parallelSum, ok := finalState.Get("parallel_sum")
	if !ok || parallelSum.(int) != 12 {
		t.Errorf("parallel_sum = %v, want 12", parallelSum)
	}

	result, ok := finalState.Get("result")
	if !ok || result.(string) != "success" {
		t.Errorf("result = %v, want success", result)
	}
}

func TestIntegration_ErrorPropagation(t *testing.T) {
	observability.RegisterObserver("noop", &observability.NoOpObserver{})

	ctx := context.Background()

	graphCfg := config.DefaultGraphConfig("test-error-propagation")
	graphCfg.Observer = "noop"

	conditionalCfg := config.DefaultConditionalConfig()
	conditionalCfg.Observer = "noop"

	graph, err := state.NewGraph(graphCfg)
	if err != nil {
		t.Fatalf("Failed to create graph: %v", err)
	}

	predicate := func(s state.State) (string, error) {
		return "error_route", nil
	}

	routes := workflows.Routes[state.State]{
		Handlers: map[string]workflows.RouteHandler[state.State]{
			"error_route": func(ctx context.Context, s state.State) (state.State, error) {
				return s, fmt.Errorf("handler error")
			},
		},
	}

	conditionalNode := workflows.ConditionalNode(conditionalCfg, predicate, routes)

	if err := graph.AddNode("decision", conditionalNode); err != nil {
		t.Fatalf("Failed to add conditional node: %v", err)
	}

	if err := graph.SetEntryPoint("decision"); err != nil {
		t.Fatalf("Failed to set entry point: %v", err)
	}

	if err := graph.SetExitPoint("decision"); err != nil {
		t.Fatalf("Failed to set exit point: %v", err)
	}

	initialState := state.New(nil)

	_, err = graph.Execute(ctx, initialState)
	if err == nil {
		t.Fatal("Expected error from graph execution, got nil")
	}
}
