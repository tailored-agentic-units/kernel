package workflows

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tailored-agentic-units/kernel/observability"
	"github.com/tailored-agentic-units/kernel/orchestrate/config"
)

// TaskProcessor processes a single item and returns a result.
//
// This function type implements the parallel processing pattern where each item is
// processed independently and concurrently. The processor is fully generic and can
// implement any processing approach:
//
//   - Direct tau-core calls (primary pattern)
//   - Hub-based multi-agent coordination
//   - Pure data transformation
//   - Mixed approaches
//
// Unlike StepProcessor in sequential chains, TaskProcessor does not receive or return
// accumulated state. Each task executes independently with no dependencies on other tasks.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - item: The item to process
//
// Returns:
//   - Result of processing this item
//   - Error if processing fails
//
// Example with direct agent usage:
//
//	processor := func(ctx context.Context, question string) (string, error) {
//	    response, err := agent.Chat(ctx, question)
//	    if err != nil {
//	        return "", err
//	    }
//	    return response.Content(), nil
//	}
type TaskProcessor[TItem, TResult any] func(
	ctx context.Context,
	item TItem,
) (TResult, error)

type indexedItem[TItem any] struct {
	index int
	item  TItem
}

type indexedResult[TResult any] struct {
	index  int
	result TResult
	err    error
}

// ProcessParallel executes concurrent processing with result aggregation.
//
// ProcessParallel distributes items to a worker pool and processes them concurrently.
// Results are collected and returned in original item order despite concurrent execution.
// The function supports both fail-fast and collect-all-errors modes via configuration.
//
// The pattern is fully generic over both item type (TItem) and result type (TResult),
// enabling usage with any data types. Processing approach is not constrained - use
// direct tau-core calls, hub coordination, or pure data transformation.
//
// Worker Pool Sizing:
//
// Worker count is determined by configuration:
//   - MaxWorkers > 0: Use exact count
//   - MaxWorkers = 0: Auto-detect min(NumCPU*2, WorkerCap, len(items))
//
// Auto-detection balances concurrency with resource usage. The 2x CPU multiplier is
// optimal for I/O-bound work like agent API calls.
//
// Error Handling Modes:
//
// FailFast=true (default):
//   - Stops on first error
//   - Cancels all workers immediately
//   - Returns ParallelError with partial results
//
// FailFast=false:
//   - Continues processing all items
//   - Collects all errors
//   - Returns error only if ALL items failed
//   - Check result.Errors for failures when no error returned
//
// Observer Integration:
//
// Emits events at key execution points:
//   - EventParallelStart: Before processing begins
//   - EventWorkerStart: Before each item processes
//   - EventWorkerComplete: After each item (success or failure)
//   - EventParallelComplete: When execution finishes
//
// Empty Input Behavior:
//
// When items slice is empty, returns immediately with:
//   - Results = empty slice
//   - Errors = empty slice
//   - Emits start/complete events for consistency
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - cfg: Configuration including worker count, fail-fast mode, and observer
//   - items: Slice of items to process concurrently
//   - processor: Function to process each item independently
//   - progress: Optional progress callback (nil to disable)
//
// Returns:
//   - ParallelResult with ordered results and any errors
//   - Error when FailFast=true and any item failed, OR FailFast=false and all failed
//
// Example with fail-fast:
//
//	questions := []string{"What is AI?", "What is ML?", "What is DL?"}
//	processor := func(ctx context.Context, q string) (string, error) {
//	    response, err := agent.Chat(ctx, q)
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
//	fmt.Printf("All %d questions answered\n", len(result.Results))
//
// Example with collect-all-errors:
//
//	failFast := false
//	cfg := config.ParallelConfig{FailFastNil: &failFast, Observer: "slog"}
//	result, err := workflows.ProcessParallel(ctx, cfg, items, processor, nil)
//	if err != nil {
//	    log.Fatal("All items failed") // Only when ALL items failed
//	}
//	if len(result.Errors) > 0 {
//	    fmt.Printf("%d items succeeded, %d failed\n", len(result.Results), len(result.Errors))
//	    for _, taskErr := range result.Errors {
//	        log.Printf("Item %d failed: %v", taskErr.Index, taskErr.Err)
//	    }
//	}
func ProcessParallel[TItem, TResult any](
	ctx context.Context,
	cfg config.ParallelConfig,
	items []TItem,
	processor TaskProcessor[TItem, TResult],
	progress ProgressFunc[TResult],
) (ParallelResult[TItem, TResult], error) {
	observer, err := observability.GetObserver(cfg.Observer)
	if err != nil {
		return ParallelResult[TItem, TResult]{}, fmt.Errorf("failed to resolve observer: %w", err)
	}

	if len(items) == 0 {
		observer.OnEvent(ctx, observability.Event{
			Type:      EventParallelStart,
			Level:     observability.LevelInfo,
			Timestamp: time.Now(),
			Source:    "workflows.ProcessParallel",
			Data: map[string]any{
				"item_count":            0,
				"worker_count":          0,
				"fail_fast":             cfg.FailFast(),
				"has_progress_callback": progress != nil,
			},
		})

		observer.OnEvent(ctx, observability.Event{
			Type:      EventParallelComplete,
			Level:     observability.LevelInfo,
			Timestamp: time.Now(),
			Source:    "workflows.ProcessParallel",
			Data: map[string]any{
				"items_processed": 0,
				"items_failed":    0,
				"error":           false,
			},
		})

		return ParallelResult[TItem, TResult]{
			Results: []TResult{},
			Errors:  []TaskError[TItem]{},
		}, nil
	}

	workerCount := calculateWorkerCount(cfg.MaxWorkers, cfg.WorkerCap, len(items))

	observer.OnEvent(ctx, observability.Event{
		Type:      EventParallelStart,
		Level:     observability.LevelInfo,
		Timestamp: time.Now(),
		Source:    "workflows.ProcessParallel",
		Data: map[string]any{
			"item_count":            len(items),
			"worker_count":          workerCount,
			"fail_fast":             cfg.FailFast(),
			"has_progress_callback": progress != nil,
		},
	})

	workQueue := make(chan indexedItem[TItem], len(items))
	resultChannel := make(chan indexedResult[TResult], len(items))
	done := make(chan struct{})

	var results []TResult
	var errors []TaskError[TItem]
	var collectorErr error

	go func() {
		results, errors, collectorErr = collectResults(resultChannel, len(items), items)
		close(done)
	}()

	var cancelCtx context.Context
	var cancel context.CancelFunc
	if cfg.FailFast() {
		cancelCtx, cancel = context.WithCancel(ctx)
		defer cancel()
	} else {
		cancelCtx = ctx
		cancel = func() {}
	}

	var wg sync.WaitGroup
	var completed atomic.Int32

	for i := range workerCount {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			processWorker(
				cancelCtx,
				workerID,
				workQueue,
				resultChannel,
				processor,
				progress,
				&completed,
				len(items),
				observer,
				cfg.FailFast(),
				cancel,
			)
		}(i)
	}

	for i, item := range items {
		workQueue <- indexedItem[TItem]{index: i, item: item}
	}
	close(workQueue)

	wg.Wait()
	close(resultChannel)
	<-done

	if collectorErr != nil {
		observer.OnEvent(ctx, observability.Event{
			Type:      EventParallelComplete,
			Level:     observability.LevelInfo,
			Timestamp: time.Now(),
			Source:    "workflows.ProcessParallel",
			Data: map[string]any{
				"items_processed": len(results),
				"items_failed":    len(errors),
				"error":           true,
			},
		})
		return ParallelResult[TItem, TResult]{
			Results: results,
			Errors:  errors,
		}, collectorErr
	}

	if ctx.Err() != nil {
		observer.OnEvent(ctx, observability.Event{
			Type:      EventParallelComplete,
			Level:     observability.LevelInfo,
			Timestamp: time.Now(),
			Source:    "workflows.ProcessParallel",
			Data: map[string]any{
				"items_processed": len(results),
				"items_failed":    len(errors),
				"error":           true,
			},
		})
		return ParallelResult[TItem, TResult]{
			Results: results,
			Errors:  errors,
		}, fmt.Errorf("parallel execution cancelled: %w", ctx.Err())
	}

	if len(errors) > 0 {
		if cfg.FailFast() || len(results) == 0 {
			observer.OnEvent(ctx, observability.Event{
				Type:      EventParallelComplete,
				Level:     observability.LevelInfo,
				Timestamp: time.Now(),
				Source:    "workflows.ProcessParallel",
				Data: map[string]any{
					"items_processed": len(results),
					"items_failed":    len(errors),
					"error":           true,
				},
			})
			return ParallelResult[TItem, TResult]{
				Results: results,
				Errors:  errors,
			}, &ParallelError[TItem]{Errors: errors}
		}
	}

	observer.OnEvent(ctx, observability.Event{
		Type:      EventParallelComplete,
		Level:     observability.LevelInfo,
		Timestamp: time.Now(),
		Source:    "workflows.ProcessParallel",
		Data: map[string]any{
			"items_processed": len(results),
			"items_failed":    len(errors),
			"error":           false,
		},
	})

	return ParallelResult[TItem, TResult]{
		Results: results,
		Errors:  errors,
	}, nil
}

// calculateWorkerCount determines optimal worker pool size based on configuration.
//
// The function implements auto-detection logic when MaxWorkers is 0:
//   - Start with NumCPU * 2 (optimal for I/O-bound work)
//   - Cap at WorkerCap to prevent excessive goroutines
//   - Cap at itemCount (no point in more workers than items)
//   - Ensure at least 1 worker
//
// When MaxWorkers > 0, returns that exact count (user override).
func calculateWorkerCount(maxWorkers, workerCap, itemCount int) int {
	if maxWorkers > 0 {
		return maxWorkers
	}

	workers := min(min(runtime.NumCPU()*2, workerCap), itemCount)

	if workers <= 0 {
		workers = 1
	}

	return workers
}

// processWorker implements individual worker goroutine logic.
//
// Each worker runs this function concurrently with other workers. The worker:
//  1. Reads items from workQueue until closed or context cancelled
//  2. Processes each item via processor function
//  3. Sends indexed results to resultChannel
//  4. Calls progress callback on success (thread-safe via atomic counter)
//  5. Cancels context on error if FailFast enabled
//
// Workers exit when:
//   - workQueue is closed (all items distributed)
//   - Context is cancelled (user cancellation or fail-fast triggered)
//
// The select statement ensures responsive cancellation - workers check context
// before pulling each item from the work queue.
func processWorker[TItem, TResult any](
	ctx context.Context,
	workerID int,
	workQueue <-chan indexedItem[TItem],
	resultChannel chan<- indexedResult[TResult],
	processor TaskProcessor[TItem, TResult],
	progress ProgressFunc[TResult],
	completed *atomic.Int32,
	total int,
	observer observability.Observer,
	failFast bool,
	cancel context.CancelFunc,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case work, ok := <-workQueue:
			if !ok {
				return
			}

			observer.OnEvent(ctx, observability.Event{
				Type:      EventWorkerStart,
				Level:     observability.LevelVerbose,
				Timestamp: time.Now(),
				Source:    "workflows.ProcessParallel",
				Data: map[string]any{
					"worker_id":   workerID,
					"item_index":  work.index,
					"total_items": total,
				},
			})

			result, err := processor(ctx, work.item)

			observer.OnEvent(ctx, observability.Event{
				Type:      EventWorkerComplete,
				Level:     observability.LevelVerbose,
				Timestamp: time.Now(),
				Source:    "workflows.ProcessParallel",
				Data: map[string]any{
					"worker_id":   workerID,
					"item_index":  work.index,
					"total_items": total,
					"error":       err != nil,
				},
			})

			if err != nil {
				resultChannel <- indexedResult[TResult]{
					index: work.index,
					err:   err,
				}
				if failFast {
					cancel()
					return
				}
			} else {
				resultChannel <- indexedResult[TResult]{
					index:  work.index,
					result: result,
				}
				if progress != nil {
					count := completed.Add(1)
					progress(int(count), total, result)
				}
			}
		}
	}
}

// collectResults aggregates worker results and preserves original item order.
//
// This function runs in a background goroutine, collecting results concurrently with
// worker execution. Running the collector in the background prevents deadlocks when
// the result channel buffer fills.
//
// The collector:
//  1. Reads all results from resultChannel until closed
//  2. Separates successes into resultMap, failures into errorMap (keyed by index)
//  3. Builds ordered slices by iterating 0 to itemCount
//  4. Returns dense slices (successes-only and failures-only)
//
// Order preservation is achieved through indexed results - even though workers complete
// out of order, the final slices are built by iterating indices sequentially.
func collectResults[TItem, TResult any](
	resultChannel <-chan indexedResult[TResult],
	itemCount int,
	items []TItem,
) ([]TResult, []TaskError[TItem], error) {
	resultMap := make(map[int]TResult)
	errorMap := make(map[int]error)

	for result := range resultChannel {
		if result.err != nil {
			errorMap[result.index] = result.err
		} else {
			resultMap[result.index] = result.result
		}
	}

	results := make([]TResult, 0, len(resultMap))
	errors := make([]TaskError[TItem], 0, len(errorMap))

	for i := range itemCount {
		if result, ok := resultMap[i]; ok {
			results = append(results, result)
		}
		if err, ok := errorMap[i]; ok {
			errors = append(errors, TaskError[TItem]{
				Index: i,
				Item:  items[i],
				Err:   err,
			})
		}
	}

	return results, errors, nil
}
