package workflows_test

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tailored-agentic-units/kernel/orchestrate/config"
	"github.com/tailored-agentic-units/kernel/orchestrate/workflows"
)

func TestProcessParallel_EmptyInput(t *testing.T) {
	ctx := context.Background()
	cfg := config.DefaultParallelConfig()
	items := []string{}

	processor := func(ctx context.Context, item string) (string, error) {
		return item, nil
	}

	result, err := workflows.ProcessParallel(ctx, cfg, items, processor, nil)

	if err != nil {
		t.Fatalf("Expected no error for empty input, got: %v", err)
	}
	if len(result.Results) != 0 {
		t.Errorf("Expected empty results, got %d results", len(result.Results))
	}
	if len(result.Errors) != 0 {
		t.Errorf("Expected empty errors, got %d errors", len(result.Errors))
	}
}

func TestProcessParallel_SingleItemSuccess(t *testing.T) {
	ctx := context.Background()
	cfg := config.DefaultParallelConfig()
	items := []string{"test"}

	processor := func(ctx context.Context, item string) (string, error) {
		return strings.ToUpper(item), nil
	}

	result, err := workflows.ProcessParallel(ctx, cfg, items, processor, nil)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(result.Results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(result.Results))
	}
	if result.Results[0] != "TEST" {
		t.Errorf("Expected 'TEST', got %q", result.Results[0])
	}
	if len(result.Errors) != 0 {
		t.Errorf("Expected no errors, got %d", len(result.Errors))
	}
}

func TestProcessParallel_MultipleItemsSuccess(t *testing.T) {
	ctx := context.Background()
	cfg := config.DefaultParallelConfig()
	items := []string{"one", "two", "three", "four", "five"}

	processor := func(ctx context.Context, item string) (string, error) {
		return strings.ToUpper(item), nil
	}

	result, err := workflows.ProcessParallel(ctx, cfg, items, processor, nil)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(result.Results) != 5 {
		t.Fatalf("Expected 5 results, got %d", len(result.Results))
	}

	expected := []string{"ONE", "TWO", "THREE", "FOUR", "FIVE"}
	for i, want := range expected {
		if result.Results[i] != want {
			t.Errorf("Result[%d]: expected %q, got %q", i, want, result.Results[i])
		}
	}

	if len(result.Errors) != 0 {
		t.Errorf("Expected no errors, got %d", len(result.Errors))
	}
}

func TestProcessParallel_OrderPreservation(t *testing.T) {
	ctx := context.Background()
	cfg := config.DefaultParallelConfig()
	items := make([]int, 100)
	for i := range items {
		items[i] = i
	}

	processor := func(ctx context.Context, item int) (int, error) {
		time.Sleep(time.Millisecond * time.Duration(100-item))
		return item * 2, nil
	}

	result, err := workflows.ProcessParallel(ctx, cfg, items, processor, nil)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(result.Results) != 100 {
		t.Fatalf("Expected 100 results, got %d", len(result.Results))
	}

	for i, item := range result.Results {
		expected := i * 2
		if item != expected {
			t.Errorf("Order not preserved: result[%d] = %d, expected %d", i, item, expected)
		}
	}
}

func TestProcessParallel_FailFastMode_SingleError(t *testing.T) {
	ctx := context.Background()
	failFast := true
	cfg := config.ParallelConfig{
		MaxWorkers:  4,
		WorkerCap:   16,
		FailFastNil: &failFast,
		Observer:    "noop",
	}
	items := make([]int, 100)
	for i := range items {
		items[i] = i
	}

	testErr := errors.New("processing failed")
	processor := func(ctx context.Context, item int) (int, error) {
		if item == 50 {
			return 0, testErr
		}
		time.Sleep(time.Millisecond * 10)
		return item * 2, nil
	}

	result, err := workflows.ProcessParallel(ctx, cfg, items, processor, nil)

	if err == nil {
		t.Fatal("Expected error in fail-fast mode, got nil")
	}

	var pErr *workflows.ParallelError[int]
	if !errors.As(err, &pErr) {
		t.Fatalf("Expected ParallelError, got %T", err)
	}

	if len(pErr.Errors) == 0 {
		t.Error("Expected at least one error in ParallelError")
	}

	if len(result.Results) >= 100 {
		t.Error("Expected fail-fast to prevent processing all items")
	}
}

func TestProcessParallel_CollectAllErrorsMode_AllSuccess(t *testing.T) {
	ctx := context.Background()
	failFast := false
	cfg := config.ParallelConfig{
		MaxWorkers:  4,
		WorkerCap:   16,
		FailFastNil: &failFast,
		Observer:    "noop",
	}
	items := []int{1, 2, 3, 4, 5}

	processor := func(ctx context.Context, item int) (int, error) {
		return item * 2, nil
	}

	result, err := workflows.ProcessParallel(ctx, cfg, items, processor, nil)

	if err != nil {
		t.Fatalf("Expected no error when all succeed, got: %v", err)
	}
	if len(result.Results) != 5 {
		t.Errorf("Expected 5 results, got %d", len(result.Results))
	}
	if len(result.Errors) != 0 {
		t.Errorf("Expected no errors, got %d", len(result.Errors))
	}
}

func TestProcessParallel_CollectAllErrorsMode_PartialFailure(t *testing.T) {
	ctx := context.Background()
	failFast := false
	cfg := config.ParallelConfig{
		MaxWorkers:  4,
		WorkerCap:   16,
		FailFastNil: &failFast,
		Observer:    "noop",
	}
	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	testErr := errors.New("processing failed")
	processor := func(ctx context.Context, item int) (int, error) {
		if item%2 == 0 {
			return 0, testErr
		}
		return item * 2, nil
	}

	result, err := workflows.ProcessParallel(ctx, cfg, items, processor, nil)

	if err != nil {
		t.Fatalf("Expected no error for partial failure in collect-all mode, got: %v", err)
	}

	if len(result.Results) != 5 {
		t.Errorf("Expected 5 successful results, got %d", len(result.Results))
	}
	if len(result.Errors) != 5 {
		t.Errorf("Expected 5 errors, got %d", len(result.Errors))
	}

	for _, taskErr := range result.Errors {
		if taskErr.Item%2 != 0 {
			t.Errorf("Expected only even numbers to fail, got failure for %d", taskErr.Item)
		}
		if !errors.Is(taskErr.Err, testErr) {
			t.Errorf("Expected testErr, got %v", taskErr.Err)
		}
	}

	expected := []int{2, 6, 10, 14, 18}
	for i, want := range expected {
		if result.Results[i] != want {
			t.Errorf("Result[%d]: expected %d, got %d", i, want, result.Results[i])
		}
	}
}

func TestProcessParallel_CollectAllErrorsMode_AllFailures(t *testing.T) {
	ctx := context.Background()
	failFast := false
	cfg := config.ParallelConfig{
		MaxWorkers:  4,
		WorkerCap:   16,
		FailFastNil: &failFast,
		Observer:    "noop",
	}
	items := []int{1, 2, 3, 4, 5}

	testErr := errors.New("all items fail")
	processor := func(ctx context.Context, item int) (int, error) {
		return 0, testErr
	}

	result, err := workflows.ProcessParallel(ctx, cfg, items, processor, nil)

	if err == nil {
		t.Fatal("Expected error when all items fail, got nil")
	}

	var pErr *workflows.ParallelError[int]
	if !errors.As(err, &pErr) {
		t.Fatalf("Expected ParallelError, got %T", err)
	}

	if len(result.Results) != 0 {
		t.Errorf("Expected no successful results, got %d", len(result.Results))
	}
	if len(result.Errors) != 5 {
		t.Errorf("Expected 5 errors, got %d", len(result.Errors))
	}
}

func TestProcessParallel_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cfg := config.DefaultParallelConfig()
	items := make([]int, 100)
	for i := range items {
		items[i] = i
	}

	var processed atomic.Int32
	processor := func(ctx context.Context, item int) (int, error) {
		if item == 10 {
			cancel()
		}
		time.Sleep(time.Millisecond * 10)
		processed.Add(1)
		return item * 2, nil
	}

	result, err := workflows.ProcessParallel(ctx, cfg, items, processor, nil)

	if err == nil {
		t.Error("Expected error due to context cancellation")
	}

	if !strings.Contains(err.Error(), "cancelled") && !errors.Is(err, context.Canceled) {
		t.Errorf("Expected cancellation error, got: %v", err)
	}

	total := int(processed.Load())
	if total >= 100 {
		t.Errorf("Expected cancellation to stop processing, but processed %d items", total)
	}

	totalItems := len(result.Results) + len(result.Errors)
	if totalItems >= 100 {
		t.Errorf("Expected partial results due to cancellation, got %d total items", totalItems)
	}
}

func TestProcessParallel_WorkerPoolSizing_MaxWorkers(t *testing.T) {
	ctx := context.Background()
	failFast := true
	cfg := config.ParallelConfig{
		MaxWorkers:  2,
		WorkerCap:   16,
		FailFastNil: &failFast,
		Observer:    "noop",
	}
	items := make([]int, 20)
	for i := range items {
		items[i] = i
	}

	var concurrent atomic.Int32
	var maxConcurrent atomic.Int32

	processor := func(ctx context.Context, item int) (int, error) {
		current := concurrent.Add(1)
		defer concurrent.Add(-1)

		for {
			max := maxConcurrent.Load()
			if current <= max || maxConcurrent.CompareAndSwap(max, current) {
				break
			}
		}

		time.Sleep(time.Millisecond * 10)
		return item * 2, nil
	}

	result, err := workflows.ProcessParallel(ctx, cfg, items, processor, nil)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(result.Results) != 20 {
		t.Errorf("Expected 20 results, got %d", len(result.Results))
	}

	max := maxConcurrent.Load()
	if max > 2 {
		t.Errorf("Expected max 2 concurrent workers, observed %d", max)
	}
}

func TestProcessParallel_ProgressCallback(t *testing.T) {
	ctx := context.Background()
	cfg := config.DefaultParallelConfig()
	items := []int{1, 2, 3, 4, 5}

	var progressCalls atomic.Int32
	var lastCompleted atomic.Int32
	progress := func(completed, total int, result int) {
		progressCalls.Add(1)
		lastCompleted.Store(int32(completed))
		if total != 5 {
			t.Errorf("Progress: expected total=5, got %d", total)
		}
	}

	processor := func(ctx context.Context, item int) (int, error) {
		return item * 2, nil
	}

	result, err := workflows.ProcessParallel(ctx, cfg, items, processor, progress)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(result.Results) != 5 {
		t.Errorf("Expected 5 results, got %d", len(result.Results))
	}

	calls := progressCalls.Load()
	if calls != 5 {
		t.Errorf("Expected 5 progress callbacks, got %d", calls)
	}

	final := lastCompleted.Load()
	if final != 5 {
		t.Errorf("Expected final completed count of 5, got %d", final)
	}
}

func TestProcessParallel_ProgressCallback_OnlyCalledOnSuccess(t *testing.T) {
	ctx := context.Background()
	failFast := false
	cfg := config.ParallelConfig{
		MaxWorkers:  4,
		WorkerCap:   16,
		FailFastNil: &failFast,
		Observer:    "noop",
	}
	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	var progressCalls atomic.Int32
	progress := func(completed, total int, result int) {
		progressCalls.Add(1)
	}

	testErr := errors.New("fail")
	processor := func(ctx context.Context, item int) (int, error) {
		if item%2 == 0 {
			return 0, testErr
		}
		return item * 2, nil
	}

	result, err := workflows.ProcessParallel(ctx, cfg, items, processor, progress)

	if err != nil {
		t.Fatalf("Expected no error in collect-all mode with partial success, got: %v", err)
	}

	calls := progressCalls.Load()
	if calls != 5 {
		t.Errorf("Expected 5 progress callbacks (only successes), got %d", calls)
	}

	if len(result.Results) != 5 {
		t.Errorf("Expected 5 successful results, got %d", len(result.Results))
	}
	if len(result.Errors) != 5 {
		t.Errorf("Expected 5 errors, got %d", len(result.Errors))
	}
}

func TestProcessParallel_TaskError_PreservesContext(t *testing.T) {
	ctx := context.Background()
	failFast := false
	cfg := config.ParallelConfig{
		MaxWorkers:  4,
		WorkerCap:   16,
		FailFastNil: &failFast,
		Observer:    "noop",
	}
	items := []string{"a", "b", "c", "d", "e"}

	testErr := errors.New("specific error")
	processor := func(ctx context.Context, item string) (string, error) {
		if item == "b" || item == "d" {
			return "", testErr
		}
		return strings.ToUpper(item), nil
	}

	result, err := workflows.ProcessParallel(ctx, cfg, items, processor, nil)

	if err != nil {
		t.Fatalf("Expected no error with partial success, got: %v", err)
	}

	if len(result.Errors) != 2 {
		t.Fatalf("Expected 2 errors, got %d", len(result.Errors))
	}

	for _, taskErr := range result.Errors {
		if taskErr.Item != "b" && taskErr.Item != "d" {
			t.Errorf("Unexpected failed item: %q", taskErr.Item)
		}
		if !errors.Is(taskErr.Err, testErr) {
			t.Errorf("Expected testErr, got %v", taskErr.Err)
		}
		if (taskErr.Item == "b" && taskErr.Index != 1) || (taskErr.Item == "d" && taskErr.Index != 3) {
			t.Errorf("Index mismatch for item %q: got %d", taskErr.Item, taskErr.Index)
		}
	}
}

func TestProcessParallel_StressTest_ManyItems(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	ctx := context.Background()
	cfg := config.DefaultParallelConfig()
	items := make([]int, 1000)
	for i := range items {
		items[i] = i
	}

	processor := func(ctx context.Context, item int) (int, error) {
		return item * 2, nil
	}

	result, err := workflows.ProcessParallel(ctx, cfg, items, processor, nil)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(result.Results) != 1000 {
		t.Errorf("Expected 1000 results, got %d", len(result.Results))
	}
	if len(result.Errors) != 0 {
		t.Errorf("Expected no errors, got %d", len(result.Errors))
	}

	for i, item := range result.Results {
		expected := i * 2
		if item != expected {
			t.Errorf("Result[%d]: expected %d, got %d", i, expected, item)
		}
	}
}

func TestParallelError_ErrorMessage_Single(t *testing.T) {
	taskErr := workflows.TaskError[string]{
		Index: 5,
		Item:  "test-item",
		Err:   errors.New("connection refused"),
	}

	pErr := &workflows.ParallelError[string]{
		Errors: []workflows.TaskError[string]{taskErr},
	}

	msg := pErr.Error()
	if !strings.Contains(msg, "item 5") {
		t.Errorf("Expected error message to contain 'item 5', got: %s", msg)
	}
	if !strings.Contains(msg, "connection refused") {
		t.Errorf("Expected error message to contain underlying error, got: %s", msg)
	}
}

func TestParallelError_ErrorMessage_Multiple(t *testing.T) {
	errors := []workflows.TaskError[int]{
		{Index: 1, Item: 1, Err: errors.New("connection refused")},
		{Index: 2, Item: 2, Err: errors.New("connection refused")},
		{Index: 3, Item: 3, Err: errors.New("timeout")},
		{Index: 4, Item: 4, Err: errors.New("connection refused")},
	}

	pErr := &workflows.ParallelError[int]{Errors: errors}
	msg := pErr.Error()

	if !strings.Contains(msg, "4 items failed") {
		t.Errorf("Expected message to mention '4 items failed', got: %s", msg)
	}
	if !strings.Contains(msg, "2 error types") {
		t.Errorf("Expected message to mention '2 error types', got: %s", msg)
	}
	if !strings.Contains(msg, "connection refused") {
		t.Errorf("Expected message to contain 'connection refused', got: %s", msg)
	}
	if !strings.Contains(msg, "(3 items)") {
		t.Errorf("Expected message to show count for 'connection refused', got: %s", msg)
	}
}

func TestParallelError_Unwrap(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")
	err3 := errors.New("error 3")

	taskErrors := []workflows.TaskError[int]{
		{Index: 0, Item: 0, Err: err1},
		{Index: 1, Item: 1, Err: err2},
		{Index: 2, Item: 2, Err: err3},
	}

	pErr := &workflows.ParallelError[int]{Errors: taskErrors}
	unwrapped := pErr.Unwrap()

	if len(unwrapped) != 3 {
		t.Fatalf("Expected 3 unwrapped errors, got %d", len(unwrapped))
	}

	if !errors.Is(unwrapped[0], err1) {
		t.Error("Expected first error to be err1")
	}
	if !errors.Is(unwrapped[1], err2) {
		t.Error("Expected second error to be err2")
	}
	if !errors.Is(unwrapped[2], err3) {
		t.Error("Expected third error to be err3")
	}
}

func TestProcessParallel_InvalidObserver(t *testing.T) {
	ctx := context.Background()
	failFast := true
	cfg := config.ParallelConfig{
		MaxWorkers:  4,
		WorkerCap:   16,
		FailFastNil: &failFast,
		Observer:    "invalid-observer-name",
	}
	items := []int{1, 2, 3}

	processor := func(ctx context.Context, item int) (int, error) {
		return item * 2, nil
	}

	_, err := workflows.ProcessParallel(ctx, cfg, items, processor, nil)

	if err == nil {
		t.Fatal("Expected error for invalid observer, got nil")
	}
	if !strings.Contains(err.Error(), "observer") {
		t.Errorf("Expected error about observer, got: %v", err)
	}
}

func TestProcessParallel_Timeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	cfg := config.DefaultParallelConfig()
	items := make([]int, 100)
	for i := range items {
		items[i] = i
	}

	processor := func(ctx context.Context, item int) (int, error) {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-time.After(100 * time.Millisecond):
			return item * 2, nil
		}
	}

	result, err := workflows.ProcessParallel(ctx, cfg, items, processor, nil)

	if err == nil {
		t.Error("Expected timeout error, got nil")
	}

	totalItems := len(result.Results) + len(result.Errors)
	if totalItems >= 100 {
		t.Errorf("Expected timeout to prevent processing all items, got %d", totalItems)
	}
}

func BenchmarkProcessParallel_SmallBatch(b *testing.B) {
	ctx := context.Background()
	cfg := config.DefaultParallelConfig()
	items := make([]int, 10)
	for i := range items {
		items[i] = i
	}

	processor := func(ctx context.Context, item int) (int, error) {
		return item * 2, nil
	}

	b.ResetTimer()
	for range b.N {
		_, _ = workflows.ProcessParallel(ctx, cfg, items, processor, nil)
	}
}

func BenchmarkProcessParallel_LargeBatch(b *testing.B) {
	ctx := context.Background()
	cfg := config.DefaultParallelConfig()
	items := make([]int, 1000)
	for i := range items {
		items[i] = i
	}

	processor := func(ctx context.Context, item int) (int, error) {
		return item * 2, nil
	}

	b.ResetTimer()
	for range b.N {
		_, _ = workflows.ProcessParallel(ctx, cfg, items, processor, nil)
	}
}
