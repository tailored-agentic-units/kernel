package workflows

import (
	"context"
	"fmt"

	"github.com/tailored-agentic-units/kernel/orchestrate/config"
	"github.com/tailored-agentic-units/kernel/orchestrate/state"
)

// ChainNode creates a StateNode from sequential chain execution for state graph integration.
//
// This helper wraps ProcessChain in a StateNode, enabling sequential processing as a node
// within state graphs. The chain processes items sequentially with state accumulation,
// and returns the final state to the graph for further execution.
//
// Parameters:
//   - cfg: Configuration for chain execution (observer, intermediate state capture)
//   - items: Items to process sequentially
//   - processor: Function that processes each item with accumulated state
//   - progress: Optional callback for progress tracking (can be nil)
//
// Returns:
//   - StateNode that executes the chain and integrates with state graphs
//
// Example:
//
//	items := []string{"analyze", "validate", "transform"}
//	processor := func(ctx context.Context, item string, s state.State) (state.State, error) {
//	    count, _ := s.Get("count")
//	    return s.Set("count", count.(int)+1), nil
//	}
//
//	node := ChainNode(cfg, items, processor, nil)
//	graph.AddNode("analysis", node)
func ChainNode[TItem any](
	cfg config.ChainConfig,
	items []TItem,
	processor StepProcessor[TItem, state.State],
	progress ProgressFunc[state.State],
) state.StateNode {
	return state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
		result, err := ProcessChain(ctx, cfg, items, s, processor, progress)
		if err != nil {
			return s, fmt.Errorf("chain node failed: %w", err)
		}
		return result.Final, nil
	})
}

// ParallelNode creates a StateNode from parallel execution with result aggregation.
//
// This helper wraps ProcessParallel in a StateNode, enabling concurrent processing as a node
// within state graphs. Items are processed concurrently by worker pool, results are aggregated
// into state, and the aggregated state is returned to the graph.
//
// The aggregator function transforms parallel execution results into state updates. It receives
// both the results array and current state, allowing conditional aggregation based on state.
//
// Parameters:
//   - cfg: Configuration for parallel execution (workers, fail-fast, observer)
//   - items: Items to process concurrently
//   - processor: Function that processes each item independently
//   - progress: Optional callback for progress tracking (can be nil)
//   - aggregator: Function that merges results into state
//
// Returns:
//   - StateNode that executes parallel processing and integrates with state graphs
//
// Example:
//
//	items := []int{1, 2, 3, 4}
//	processor := func(ctx context.Context, item int) (int, error) {
//	    return item * 2, nil
//	}
//	aggregator := func(results []int, s state.State) state.State {
//	    sum := 0
//	    for _, r := range results {
//	        sum += r
//	    }
//	    return s.Set("sum", sum)
//	}
//
//	node := ParallelNode(cfg, items, processor, nil, aggregator)
//	graph.AddNode("parallel-task", node)
func ParallelNode[TItem, TResult any](
	cfg config.ParallelConfig,
	items []TItem,
	processor TaskProcessor[TItem, TResult],
	progress ProgressFunc[TResult],
	aggregator func(results []TResult, currentState state.State) state.State,
) state.StateNode {
	return state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
		result, err := ProcessParallel(ctx, cfg, items, processor, progress)
		if err != nil {
			return s, fmt.Errorf("parallel node failed: %w", err)
		}

		aggregated := aggregator(result.Results, s)
		return aggregated, nil
	})
}

// ConditionalNode creates a StateNode from conditional routing with predicate-based handler selection.
//
// This helper wraps ProcessConditional in a StateNode, enabling conditional logic as a node
// within state graphs. The predicate evaluates state to select a route, the corresponding
// handler processes the state, and the updated state is returned to the graph.
//
// The predicate function determines execution flow by returning a route name based on state.
// Route names map to handlers in the Routes configuration, with optional default fallback.
//
// Parameters:
//   - cfg: Configuration for conditional execution (observer settings)
//   - predicate: Function that evaluates state and returns route name
//   - routes: Route-to-handler mappings with optional default handler
//
// Returns:
//   - StateNode that executes conditional routing and integrates with state graphs
//
// Example:
//
//	predicate := func(s state.State) (string, error) {
//	    consensus, _ := s.Get("consensus")
//	    if consensus.(bool) {
//	        return "approve", nil
//	    }
//	    return "reject", nil
//	}
//
//	routes := workflows.Routes[state.State]{
//	    Handlers: map[string]workflows.RouteHandler[state.State]{
//	        "approve": func(ctx context.Context, s state.State) (state.State, error) {
//	            return s.Set("status", "approved"), nil
//	        },
//	        "reject": func(ctx context.Context, s state.State) (state.State, error) {
//	            return s.Set("status", "rejected"), nil
//	        },
//	    },
//	}
//
//	node := ConditionalNode(cfg, predicate, routes)
//	graph.AddNode("approval", node)
func ConditionalNode(
	cfg config.ConditionalConfig,
	predicate RoutePredicate[state.State],
	routes Routes[state.State],
) state.StateNode {
	return state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
		result, err := ProcessConditional(ctx, cfg, s, predicate, routes)
		if err != nil {
			return s, fmt.Errorf("conditional node failed: %w", err)
		}
		return result, nil
	})
}
