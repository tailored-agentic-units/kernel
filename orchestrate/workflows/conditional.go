package workflows

import (
	"context"
	"fmt"
	"time"

	"github.com/tailored-agentic-units/kernel/orchestrate/config"
	"github.com/tailored-agentic-units/kernel/orchestrate/observability"
)

// RoutePredicate evaluates state and returns a route name for conditional routing.
//
// The predicate determines which handler should process the state based on state
// content and business logic. Route names are arbitrary strings that map to handlers
// in the Routes configuration.
//
// Parameters:
//   - state: Current state to evaluate
//
// Returns:
//   - route: Name of the route to execute
//   - err: Error if predicate evaluation fails
//
// Example:
//
//	predicate := func(s State) (string, error) {
//	    score, _ := s.Get("score")
//	    if score.(int) > 80 {
//	        return "high_score", nil
//	    }
//	    return "low_score", nil
//	}
type RoutePredicate[TState any] func(state TState) (route string, err error)

// RouteHandler processes state for a specific route in conditional execution.
//
// Handlers transform state based on the routing decision. Each handler corresponds
// to a specific route name and implements the logic for that execution path.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - state: Current state to process
//
// Returns:
//   - Updated state after handler execution
//   - Error if handler processing fails
//
// Example:
//
//	handler := func(ctx context.Context, s State) (State, error) {
//	    return s.Set("status", "approved"), nil
//	}
type RouteHandler[TState any] func(
	ctx context.Context,
	state TState,
) (TState, error)

// Routes maps route names to handlers for conditional routing with optional default.
//
// The Handlers map defines named routes that correspond to predicate return values.
// The Default handler executes when the predicate returns a route not found in Handlers.
//
// Fields:
//   - Handlers: Map of route names to handler functions
//   - Default: Optional fallback handler for unmatched routes (can be nil)
//
// Example:
//
//	routes := Routes[State]{
//	    Handlers: map[string]RouteHandler[State]{
//	        "approve": approveHandler,
//	        "reject": rejectHandler,
//	    },
//	    Default: defaultHandler,
//	}
type Routes[TState any] struct {
	Handlers map[string]RouteHandler[TState]
	Default  RouteHandler[TState]
}

// ProcessConditional executes conditional routing with predicate-based handler selection.
//
// This function implements the conditional routing pattern where state evaluation determines
// which handler executes. The predicate selects a route, the corresponding handler processes
// the state, and the updated state is returned.
//
// Execution flow:
//  1. Evaluate predicate to select route
//  2. Look up handler in Routes.Handlers map
//  3. Fall back to Default handler if route not found
//  4. Execute selected handler with current state
//  5. Return updated state from handler
//
// Observer events:
//   - EventRouteEvaluate: Before predicate evaluation
//   - EventRouteSelect: After route selection
//   - EventRouteExecute: After handler execution
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - cfg: Configuration with observer settings
//   - state: Initial state to evaluate and process
//   - predicate: Function that evaluates state and returns route name
//   - routes: Route-to-handler mappings with optional default
//
// Returns:
//   - Updated state from handler execution
//   - ConditionalError if predicate, routing, or handler fails
//
// Example:
//
//	predicate := func(s State) (string, error) {
//	    consensus, _ := s.Get("consensus")
//	    if consensus.(bool) {
//	        return "approve", nil
//	    }
//	    return "reject", nil
//	}
//
//	routes := Routes[State]{
//	    Handlers: map[string]RouteHandler[State]{
//	        "approve": func(ctx context.Context, s State) (State, error) {
//	            return s.Set("status", "approved"), nil
//	        },
//	        "reject": func(ctx context.Context, s State) (State, error) {
//	            return s.Set("status", "rejected"), nil
//	        },
//	    },
//	}
//
//	finalState, err := ProcessConditional(ctx, cfg, state, predicate, routes)
func ProcessConditional[TState any](
	ctx context.Context,
	cfg config.ConditionalConfig,
	state TState,
	predicate RoutePredicate[TState],
	routes Routes[TState],
) (TState, error) {
	observer, err := observability.GetObserver(cfg.Observer)
	if err != nil {
		return state, ConditionalError[TState]{
			State: state,
			Err:   fmt.Errorf("failed to get observer: %w", err),
		}
	}

	if err := ctx.Err(); err != nil {
		return state, ConditionalError[TState]{
			State: state,
			Err:   fmt.Errorf("context cancelled before evaluation: %w", err),
		}
	}

	observer.OnEvent(ctx, observability.Event{
		Type:      observability.EventRouteEvaluate,
		Timestamp: time.Now(),
		Source:    "conditional",
		Data: map[string]any{
			"route_count": len(routes.Handlers),
		},
	})

	route, err := predicate(state)
	if err != nil {
		return state, ConditionalError[TState]{
			State: state,
			Err:   fmt.Errorf("predicate evaluation failed: %w", err),
		}
	}

	handler, found := routes.Handlers[route]
	if !found {
		if routes.Default == nil {
			return state, ConditionalError[TState]{
				Route: route,
				State: state,
				Err:   fmt.Errorf("route '%s' not found and no default handler", route),
			}
		}
		handler = routes.Default
		route = "default"
	}

	observer.OnEvent(ctx, observability.Event{
		Type:      observability.EventRouteSelect,
		Timestamp: time.Now(),
		Source:    "conditional",
		Data: map[string]any{
			"route":       route,
			"has_default": routes.Default != nil,
		},
	})

	if err := ctx.Err(); err != nil {
		return state, ConditionalError[TState]{
			Route: route,
			State: state,
			Err:   fmt.Errorf("context cancelled before handler execution: %w", err),
		}
	}

	result, err := handler(ctx, state)
	if err != nil {
		return state, ConditionalError[TState]{
			Route: route,
			State: state,
			Err:   fmt.Errorf("handler execution failed: %w", err),
		}
	}

	observer.OnEvent(ctx, observability.Event{
		Type:      observability.EventRouteExecute,
		Timestamp: time.Now(),
		Source:    "conditional",
		Data: map[string]any{
			"route": route,
			"error": false,
		},
	})

	return result, nil
}
