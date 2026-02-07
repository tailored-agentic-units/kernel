package workflows

import (
	"context"
	"fmt"
	"time"

	"github.com/tailored-agentic-units/kernel/orchestrate/config"
	"github.com/tailored-agentic-units/kernel/orchestrate/observability"
)

// StepProcessor processes a single item and updates the accumulated context.
//
// This function type implements the fold/reduce pattern where each step receives
// the current accumulated state and returns an updated state. The processor is
// fully generic and can implement any processing approach:
//
//   - Direct tau-core calls (primary pattern)
//   - Hub-based multi-agent coordination
//   - Pure data transformation
//   - Mixed approaches
//
// Parameters:
//
//   - ctx: Context for cancellation and timeout control
//   - item: The current item to process
//   - state: The accumulated state from all previous steps
//
// Returns:
//
//   - Updated state after processing this item
//   - Error if processing fails (stops the chain)
//
// Example with direct agent usage:
//
//	processor := func(ctx context.Context, question string, conversation Conversation) (Conversation, error) {
//	    response, err := agent.Chat(ctx, question)
//	    if err != nil {
//	        return conversation, err
//	    }
//	    conversation.AddExchange(question, response.Content())
//	    return conversation, nil
//	}
type StepProcessor[TItem, TContext any] func(
	ctx context.Context,
	item TItem,
	state TContext,
) (TContext, error)

// ChainResult contains the results of chain execution.
//
// The Final field always contains the result (either final state on success
// or initial state on immediate failure). Intermediate states are only populated
// when ChainConfig.CaptureIntermediateStates is true.
type ChainResult[TContext any] struct {
	// Final is the accumulated state after all steps completed
	Final TContext

	// Intermediate contains state after each step when captured.
	// Index 0 is the initial state, index N is state after step N.
	// Only populated when ChainConfig.CaptureIntermediateStates is true.
	Intermediate []TContext

	// Steps is the number of steps successfully completed
	Steps int
}

// ProcessChain executes a sequential chain with state accumulation.
//
// Implements a fold/reduce pattern where each item is processed in order, with
// accumulated state passed from step to step. Processing stops on first error
// (fail-fast). Context cancellation is checked at the start of each step.
//
// The chain is fully generic over both item type (TItem) and context type (TContext),
// enabling usage with any data types. The state package's State type works naturally
// as TContext for stateful workflows.
//
// Observer Integration:
//
// Emits events at all key execution points for observability:
//   - EventChainStart: Before processing begins
//   - EventStepStart: Before each step processes
//   - EventStepComplete: After each step (success or failure)
//   - EventChainComplete: When chain finishes
//
// Error Handling:
//
// Errors are wrapped in ChainError with complete context including:
//   - Step index where failure occurred
//   - Item being processed
//   - State at time of failure
//   - Underlying error
//
// Empty Chain Behavior:
//
// When items slice is empty, returns immediately with:
//   - Final = initial state
//   - Steps = 0
//   - Emits start/complete events for consistency
//
// Parameters:
//
//   - ctx: Context for cancellation and timeout control
//   - cfg: Configuration including observer and intermediate state capture
//   - items: Slice of items to process sequentially
//   - initial: Initial state for accumulation
//   - processor: Function to process each item with current state
//   - progress: Optional progress callback (nil to disable)
//
// Returns:
//
//   - ChainResult with final state, optional intermediate states, and step count
//   - Error wrapped in ChainError on failure, nil on success
//
// Example with direct agent usage:
//
//	questions := []string{"What is AI?", "What is ML?", "What is DL?"}
//	initial := Conversation{}
//
//	processor := func(ctx context.Context, question string, conv Conversation) (Conversation, error) {
//	    response, err := agent.Chat(ctx, question)
//	    if err != nil {
//	        return conv, err
//	    }
//	    conv.AddExchange(question, response.Content())
//	    return conv, nil
//	}
//
//	result, err := workflows.ProcessChain(ctx, config.DefaultChainConfig(), questions, initial, processor, nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Final conversation: %v\n", result.Final)
//
// Example with state package integration:
//
//	items := []string{"task1", "task2", "task3"}
//	initial := state.State{"status": "starting"}
//
//	processor := func(ctx context.Context, task string, s state.State) (state.State, error) {
//	    return s.Set("last_task", task), nil
//	}
//
//	result, err := workflows.ProcessChain(ctx, cfg, items, initial, processor, nil)
func ProcessChain[TItem, TContext any](
	ctx context.Context,
	cfg config.ChainConfig,
	items []TItem,
	initial TContext,
	processor StepProcessor[TItem, TContext],
	progress ProgressFunc[TContext],
) (ChainResult[TContext], error) {
	observer, err := observability.GetObserver(cfg.Observer)
	if err != nil {
		return ChainResult[TContext]{}, fmt.Errorf("failed to resolve observer: %w", err)
	}

	result := ChainResult[TContext]{
		Final: initial,
		Steps: 0,
	}

	observer.OnEvent(ctx, observability.Event{
		Type:      observability.EventChainStart,
		Timestamp: time.Now(),
		Source:    "workflows.ProcessChain",
		Data: map[string]any{
			"item_count":            len(items),
			"has_progress_callback": progress != nil,
			"capture_intermediate":  cfg.CaptureIntermediateStates,
		},
	})

	if len(items) == 0 {
		observer.OnEvent(ctx, observability.Event{
			Type:      observability.EventChainComplete,
			Timestamp: time.Now(),
			Source:    "workflows.ProcessChain",
			Data: map[string]any{
				"steps_completed": 0,
				"error":           false,
			},
		})
		return result, nil
	}

	var intermediate []TContext
	if cfg.CaptureIntermediateStates {
		intermediate = make([]TContext, 0, len(items)+1)
		intermediate = append(intermediate, initial)
	}

	state := initial

	for i, item := range items {
		if err := ctx.Err(); err != nil {
			chainErr := &ChainError[TItem, TContext]{
				StepIndex: i,
				Item:      item,
				State:     state,
				Err:       fmt.Errorf("processing cancelled: %w", err),
			}
			observer.OnEvent(ctx, observability.Event{
				Type:      observability.EventChainComplete,
				Timestamp: time.Now(),
				Source:    "workflows.ProcessChain",
				Data: map[string]any{
					"steps_completed": i,
					"error":           true,
					"error_type":      "cancellation",
				},
			})
			return result, chainErr
		}

		observer.OnEvent(ctx, observability.Event{
			Type:      observability.EventStepStart,
			Timestamp: time.Now(),
			Source:    "workflows.ProcessChain",
			Data: map[string]any{
				"step_index":  i,
				"total_steps": len(items),
			},
		})

		updated, err := processor(ctx, item, state)
		if err != nil {
			chainErr := &ChainError[TItem, TContext]{
				StepIndex: i,
				Item:      item,
				State:     state,
				Err:       err,
			}
			observer.OnEvent(ctx, observability.Event{
				Type:      observability.EventStepComplete,
				Timestamp: time.Now(),
				Source:    "workflows.ProcessChain",
				Data: map[string]any{
					"step_index":  i,
					"total_steps": len(items),
					"error":       true,
				},
			})
			observer.OnEvent(ctx, observability.Event{
				Type:      observability.EventChainComplete,
				Timestamp: time.Now(),
				Source:    "workflows.ProcessChain",
				Data: map[string]any{
					"steps_completed": i,
					"error":           true,
					"error_type":      "processor",
				},
			})
			return result, chainErr
		}

		state = updated

		if cfg.CaptureIntermediateStates {
			intermediate = append(intermediate, state)
		}

		observer.OnEvent(ctx, observability.Event{
			Type:      observability.EventStepComplete,
			Timestamp: time.Now(),
			Source:    "workflows.ProcessChain",
			Data: map[string]any{
				"step_index":  i,
				"total_steps": len(items),
				"error":       false,
			},
		})

		if progress != nil {
			progress(i+1, len(items), state)
		}
	}

	result.Final = state
	result.Intermediate = intermediate
	result.Steps = len(items)

	observer.OnEvent(ctx, observability.Event{
		Type:      observability.EventChainComplete,
		Timestamp: time.Now(),
		Source:    "workflows.ProcessChain",
		Data: map[string]any{
			"steps_completed": len(items),
			"error":           false,
		},
	})

	return result, nil
}
