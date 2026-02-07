// Package workflows provides composable workflow patterns for orchestrating multi-step processes.
//
// This package implements generic workflow primitives that work with any item and
// context types. All patterns support direct tau-core usage as the primary approach,
// with optional hub coordination for multi-agent orchestration.
//
// # Sequential Chain Pattern
//
// The sequential chain pattern implements a fold/reduce operation where items are
// processed in order with state accumulation between steps. Each step receives the
// current accumulated state and returns an updated state. Processing stops on first
// error (fail-fast).
//
// Example with direct agent usage:
//
//	questions := []string{"What is AI?", "What is ML?"}
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
//
// # Parallel Execution Pattern
//
// The parallel execution pattern processes items concurrently using a worker pool.
// Results are aggregated and returned in original item order despite concurrent
// execution. Supports both fail-fast and collect-all-errors modes.
//
// Example with fail-fast mode:
//
//	questions := []string{"What is AI?", "What is ML?", "What is DL?"}
//
//	processor := func(ctx context.Context, question string) (string, error) {
//	    response, err := agent.Chat(ctx, question)
//	    if err != nil {
//	        return "", err
//	    }
//	    return response.Content(), nil
//	}
//
//	cfg := config.DefaultParallelConfig() // FailFast() returns true
//	result, err := workflows.ProcessParallel(ctx, cfg, questions, processor, nil)
//	if err != nil {
//	    log.Fatal(err) // First error stops all processing
//	}
//
// Example with collect-all-errors mode:
//
//	failFast := false
//	cfg := config.ParallelConfig{
//	    FailFastNil: &failFast,
//	    Observer:    "slog",
//	}
//	result, err := workflows.ProcessParallel(ctx, cfg, items, processor, nil)
//	if err != nil {
//	    log.Fatal("All items failed")
//	}
//	if len(result.Errors) > 0 {
//	    fmt.Printf("%d succeeded, %d failed\n", len(result.Results), len(result.Errors))
//	    for _, taskErr := range result.Errors {
//	        log.Printf("Item %d failed: %v", taskErr.Index, taskErr.Err)
//	    }
//	}
//
// # Pattern Independence
//
// All workflow patterns are agnostic about processing approach:
//   - Direct tau-core usage (primary pattern)
//   - Hub orchestration (optional for multi-agent coordination)
//   - Pure data transformation (no agents required)
//   - Mixed approaches (some steps with agents, some without)
//
// The processor function signatures intentionally don't constrain implementation,
// enabling maximum flexibility.
//
// # Worker Pool Auto-Detection
//
// Parallel execution automatically sizes the worker pool based on workload and system
// resources when MaxWorkers is 0 (default):
//
//	workers = min(NumCPU * 2, WorkerCap, len(items))
//
// The 2x CPU multiplier is optimal for I/O-bound work like agent API calls. For CPU-bound
// work, set MaxWorkers to runtime.NumCPU(). The WorkerCap (default 16) prevents excessive
// goroutines for large item sets.
//
// # Error Handling Modes
//
// Sequential chains always use fail-fast (stop on first error). Parallel execution
// supports two modes:
//
// Fail-Fast Mode (FailFast=true, default):
//   - Stops on first error
//   - Cancels all workers immediately
//   - Returns ParallelError with partial results
//
// Collect-All-Errors Mode (FailFast=false):
//   - Continues processing all items
//   - Collects all errors in result.Errors
//   - Returns error only if ALL items failed
//   - Check result.Errors for partial failures
//
// # Observer Integration
//
// All patterns emit events at key execution points for observability:
//
// Sequential chains:
//   - EventChainStart, EventChainComplete
//   - EventStepStart, EventStepComplete
//
// Parallel execution:
//   - EventParallelStart, EventParallelComplete
//   - EventWorkerStart, EventWorkerComplete (per item, includes worker ID)
//
// Observer configuration is provided via config package structures, following
// the configuration lifecycle principle (config used only during initialization).
//
// Default observer is "slog" (structured logging). Use "noop" for zero overhead.
//
// # Progress Callbacks
//
// Both patterns support optional progress callbacks for monitoring execution:
//
// Sequential chains:
//
//	progress := func(completed, total int, state Conversation) {
//	    fmt.Printf("Chain: %d/%d steps complete\n", completed, total)
//	}
//
// Parallel execution:
//
//	progress := func(completed, total int, result string) {
//	    fmt.Printf("Parallel: %d/%d items complete\n", completed, total)
//	}
//
// Progress callbacks are called after each successful step/item completion. In parallel
// execution, callbacks use atomic counters for thread-safe progress tracking.
//
// # Error Types
//
// Workflow patterns provide rich error types with complete context:
//
// ChainError (sequential chains):
//   - StepIndex: Where failure occurred
//   - Item: Item being processed
//   - State: Accumulated state at failure
//   - Err: Underlying error (unwrappable)
//
// ParallelError (parallel execution):
//   - Errors: All task failures with context
//   - Error(): Categorized summary with error types and counts
//   - Unwrap(): All underlying errors (Go 1.20+ multiple unwrapping)
//
// TaskError (parallel execution failures):
//   - Index: Original item position
//   - Item: Item that failed
//   - Err: Underlying error
//
// # Integration with State Package
//
// The state package's State type works naturally as TContext for sequential chains,
// enabling composition of workflow patterns with state graph execution (Phase 7):
//
//	items := []string{"task1", "task2"}
//	initial := state.State{"status": "starting"}
//
//	processor := func(ctx context.Context, task string, s state.State) (state.State, error) {
//	    return s.Set("last_task", task), nil
//	}
//
//	result, err := workflows.ProcessChain(ctx, cfg, items, initial, processor, nil)
//
// # Deadlock Prevention
//
// Parallel execution uses three-channel coordination with background result collection
// to prevent deadlocks:
//
//   - Work queue (buffered to len(items))
//   - Result channel (buffered to len(items))
//   - Done signal (unbuffered)
//
// The background collector drains the result channel concurrently with worker execution,
// preventing blocking even when all workers complete simultaneously.
//
// # Context Cancellation
//
// Both patterns support context cancellation for graceful shutdown:
//
// Sequential chains:
//   - Checks context before each step
//   - Returns ChainError with cancellation context
//
// Parallel execution:
//   - Workers select on context before each item
//   - Fail-fast mode creates cancellable child context
//   - First error triggers cancellation in fail-fast mode
//   - User can cancel original context in any mode
package workflows
