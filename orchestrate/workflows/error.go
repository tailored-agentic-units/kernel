package workflows

import (
	"fmt"
	"sort"
	"strings"
)

// ChainError provides rich error context for chain execution failures.
//
// Generic over both TItem and TContext to preserve complete error state including
// the item being processed and the accumulated state at failure point. This enables
// detailed debugging and recovery strategies.
//
// The error implements standard error unwrapping via Unwrap(), enabling errors.Is
// and errors.As for error chain inspection.
//
// Example usage:
//
//	result, err := workflows.ProcessChain(ctx, cfg, items, initial, processor, nil)
//	if err != nil {
//	    if chainErr, ok := err.(*ChainError[Item, State]); ok {
//	        fmt.Printf("Failed at step %d\n", chainErr.StepIndex)
//	        fmt.Printf("Failed item: %v\n", chainErr.Item)
//	        fmt.Printf("State at failure: %v\n", chainErr.State)
//	    }
//	}
type ChainError[TItem, TContext any] struct {
	// StepIndex is the 0-based index of the step that failed
	StepIndex int

	// Item is the item being processed when the error occurred
	Item TItem

	// State is the accumulated context at the time of failure
	State TContext

	// Err is the underlying error that caused the failure
	Err error
}

// Error returns a formatted error message with step index context.
// Implements the standard error interface.
func (e *ChainError[TItem, TContext]) Error() string {
	return fmt.Sprintf("chain failed at step %d: %v", e.StepIndex, e.Err)
}

// Unwrap returns the underlying error, enabling errors.Is and errors.As.
// This supports standard Go error unwrapping patterns.
func (e *ChainError[TItem, TContext]) Unwrap() error {
	return e.Err
}

// TaskError captures failure context for a single parallel task.
//
// TaskError preserves the original item index, the item itself, and the underlying
// error. This enables detailed error reporting and selective retry strategies.
//
// The Index field corresponds to the original position in the items slice passed to
// ProcessParallel, enabling correlation between failures and input data.
//
// Example:
//
//	result, err := workflows.ProcessParallel(ctx, cfg, items, processor, nil)
//	for _, taskErr := range result.Errors {
//	    fmt.Printf("Item %d failed: %v\n", taskErr.Index, taskErr.Err)
//	    fmt.Printf("Failed item: %v\n", taskErr.Item)
//	}
type TaskError[TItem any] struct {
	// Index is the 0-based position of the item in the original items slice
	Index int

	// Item is the actual item that failed processing
	Item TItem

	// Err is the underlying error returned by the processor function
	Err error
}

// ParallelResult contains the results of parallel execution.
//
// ParallelResult separates successful results from failures using dense slices.
// The Results slice contains only successes (no zero values or gaps), and the
// Errors slice contains only failures with complete context.
//
// Result interpretation:
//   - Total items processed: len(Results) + len(Errors)
//   - Success count: len(Results)
//   - Failure count: len(Errors)
//
// When ProcessParallel returns an error, the ParallelResult still contains partial
// results from items that completed before the error condition. This enables
// recovery and partial success handling even when the function returns an error.
//
// Example:
//
//	result, err := workflows.ProcessParallel(ctx, cfg, items, processor, nil)
//	if err != nil {
//	    fmt.Printf("Some failures occurred, but %d items succeeded\n", len(result.Results))
//	}
//	fmt.Printf("Successfully processed: %d items\n", len(result.Results))
//	fmt.Printf("Failed to process: %d items\n", len(result.Errors))
type ParallelResult[TItem, TResult any] struct {
	// Results contains all successfully processed items (dense slice, no gaps)
	Results []TResult

	// Errors contains all failed items with context (index, item, error)
	Errors []TaskError[TItem]
}

// ParallelError wraps task processing failures from parallel execution.
//
// ParallelError is returned when ProcessParallel encounters item processing failures
// that meet the error return criteria (FailFast=true and any failure, or FailFast=false
// and all items failed). The error provides categorized summary messages and supports
// Go 1.20+ multiple error unwrapping.
//
// Error message formats:
//   - Single failure: "parallel execution failed: item 5: connection refused"
//   - Multiple failures: "parallel execution failed: 15 items failed with 2 error types: 'connection refused' (12 items), 'timeout exceeded' (3 items)"
//
// The Unwrap() method returns all underlying errors as a slice, enabling errors.Is
// and errors.As to search across all task failures.
//
// Example:
//
//	result, err := workflows.ProcessParallel(ctx, cfg, items, processor, nil)
//	if err != nil {
//	    var pErr *ParallelError[string]
//	    if errors.As(err, &pErr) {
//	        for _, taskErr := range pErr.Errors {
//	            log.Printf("Task %d failed: %v", taskErr.Index, taskErr.Err)
//	        }
//	    }
//	}
type ParallelError[TItem any] struct {
	// Errors contains all task failures that contributed to this error
	Errors []TaskError[TItem]
}

// Error returns a categorized summary of parallel execution failures.
//
// The error message format depends on the number of failures:
//   - 0 errors: Generic message (should not occur in practice)
//   - 1 error: Full detail with item index and error message
//   - Multiple errors: Categorized by error type with counts, sorted by frequency
//
// Example messages:
//   - "parallel execution failed: item 5: connection refused"
//   - "parallel execution failed: 8 items failed with 3 error types: 'invalid input' (5 items), 'context canceled' (2 items), 'not found' (1 item)"
func (e *ParallelError[TItem]) Error() string {
	if len(e.Errors) == 0 {
		return "parallel execution failed"
	}
	if len(e.Errors) == 1 {
		return fmt.Sprintf("parallel execution failed: item %d: %v",
			e.Errors[0].Index, e.Errors[0].Err,
		)
	}

	errorCounts := make(map[string]int)
	for _, taskErr := range e.Errors {
		errorCounts[taskErr.Err.Error()]++
	}

	type errorSummary struct {
		msg   string
		count int
	}
	var summaries []errorSummary
	for msg, count := range errorCounts {
		summaries = append(summaries, errorSummary{msg, count})
	}
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].count > summaries[j].count
	})

	var parts []string
	for _, s := range summaries {
		if s.count == 1 {
			parts = append(parts, fmt.Sprintf("'%s' (1 item)", s.msg))
		} else {
			parts = append(parts, fmt.Sprintf("'%s' (%d items)", s.msg, s.count))
		}
	}

	return fmt.Sprintf(
		"parallel execution failed: %d items failed with %d error types: %s",
		len(e.Errors), len(errorCounts), strings.Join(parts, ", "),
	)
}

// Unwrap returns all underlying task errors for Go 1.20+ error unwrapping.
//
// This method enables errors.Is and errors.As to search across all task failures
// in the parallel execution. The returned slice contains only the underlying errors,
// not the TaskError wrappers.
//
// Example:
//
//	result, err := workflows.ProcessParallel(ctx, cfg, items, processor, nil)
//	if errors.Is(err, context.DeadlineExceeded) {
//	    // At least one task failed due to timeout
//	}
func (e *ParallelError[TItem]) Unwrap() []error {
	errs := make([]error, len(e.Errors))
	for i, taskErr := range e.Errors {
		errs[i] = taskErr.Err
	}
	return errs
}

type ConditionalError[TState any] struct {
	Route string
	State TState
	Err   error
}

func (e ConditionalError[TState]) Error() string {
	if e.Route == "" {
		return fmt.Sprintf("conditional routing failed: %v", e.Err)
	}
	return fmt.Sprintf("conditional routing failed for route '%s': %v", e.Route, e.Err)
}

func (e ConditionalError[TState]) Unwrap() error {
	return e.Err
}
