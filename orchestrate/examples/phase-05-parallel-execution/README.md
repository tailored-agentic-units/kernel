# Phase 5: Parallel Execution Example - Product Review Sentiment Analysis

This example demonstrates concurrent processing with worker pool coordination, order preservation, and real-time progress tracking using a realistic product review sentiment analysis scenario.

## Overview

The example simulates an automated sentiment analysis system where a sentiment-analyst agent processes multiple product reviews concurrently:
- **12 product reviews** analyzed in parallel
- **4 concurrent workers** (auto-detected based on CPU)
- **Independent processing** - each review analyzed separately
- **Order preservation** - results maintain original sequence
- **Real-time progress** - completion tracking across workers

The pipeline demonstrates:
1. **Concurrent execution** - Multiple items processed simultaneously
2. **Worker pool pattern** - Fixed number of workers processing queue
3. **Order preservation** - Results returned in original item order
4. **Progress tracking** - Real-time completion percentage
5. **Error handling** - Fail-fast or collect-all-errors modes
6. **Observer integration** - Worker and parallel events via slog observer

## Architecture

### Parallel Execution Pattern

```
Items: [Review1, Review2, Review3, ..., Review12]
                    ↓
              Work Queue (buffered channel)
                    ↓
        ┌──────────┼──────────┬──────────┐
        ↓          ↓          ↓          ↓
    Worker 1   Worker 2   Worker 3   Worker 4
        ↓          ↓          ↓          ↓
        └──────────┼──────────┴──────────┘
                   ↓
         Result Channel (buffered)
                   ↓
         Background Collector
                   ↓
         Results (ordered by original index)
```

### Three-Channel Coordination

```
Work Queue         Workers             Result Channel      Collector
(buffered)     (N goroutines)          (buffered)         (goroutine)
    ↓                ↓                      ↓                  ↓
{idx: 0}  →→→  Process Item 0  →→→  {idx: 0, result}  →→→  results[0]
{idx: 1}  →→→  Process Item 1  →→→  {idx: 1, result}  →→→  results[1]
{idx: 2}  →→→  Process Item 2  →→→  {idx: 2, result}  →→→  results[2]
   ...           ...                     ...                 ...
```

**Key Point**: Background collector prevents deadlocks when result channel fills before all workers complete.

### Data Structures

**ProductReview** (Item Type):
```go
type ProductReview struct {
    ID      int
    Product string
    Review  string
}
```

**SentimentResult** (Result Type):
```go
type SentimentResult struct {
    ReviewID    int
    Product     string
    Review      string
    Sentiment   string       // "positive", "negative", "neutral"
    Analysis    string       // Full agent response
    ProcessedAt time.Time
}
```

**ParallelResult**:
```go
type ParallelResult[TItem, TResult any] struct {
    Results []TResult           // Dense slice of successful results
    Errors  []TaskError[TItem]  // Failures with complete context
}
```

**TaskError**:
```go
type TaskError[TItem any] struct {
    Index int       // Original position in items slice
    Item  TItem     // The item that failed
    Err   error     // Underlying error
}
```

## Key Concepts Demonstrated

### 1. TaskProcessor Function

The processor executes independently for each item:

```go
taskProcessor := func(ctx context.Context, review ProductReview) (SentimentResult, error) {
    prompt := fmt.Sprintf("Analyze sentiment: \"%s\"", review.Review)

    response, err := agent.Chat(ctx, prompt)
    if err != nil {
        return SentimentResult{}, err
    }

    return SentimentResult{
        ReviewID:  review.ID,
        Sentiment: response.Content(),
        // ...
    }, nil
}
```

**Key Differences from Sequential Chains:**
- **No state accumulation** - each task is independent
- **No dependencies** - tasks can execute in any order
- **Returns result** - not accumulated state
- **Stateless** - processor doesn't receive or update shared state

### 2. Worker Pool Auto-Detection

Worker count is automatically determined:

```go
workers = min(min(runtime.NumCPU()*2, WorkerCap), len(items))
```

**Calculation:**
- `NumCPU()*2` - Optimal for I/O-bound work (agent API calls)
- `WorkerCap` - Default 16, prevents excessive goroutines
- `len(items)` - Never more workers than items

**Example (4-core CPU, 12 items):**
```
NumCPU = 4
NumCPU * 2 = 8
min(8, 16) = 8
min(8, 12) = 8 workers

But in example: WorkerCap set to 4
min(4 * 2, 4) = 4
min(4, 12) = 4 workers
```

### 3. Order Preservation

Despite concurrent execution, results maintain original order:

**Implementation:**
1. Each item tagged with original index in work queue
2. Each result tagged with original index in result channel
3. Maps used for O(1) lookup by index during collection
4. Final slices built by iterating 0 to itemCount sequentially

**Why This Matters:**
- UI displays results in expected order
- Downstream processing can rely on sequence
- Debugging easier with predictable ordering

### 4. Fail-Fast vs Collect-All-Errors

**FailFast=true (default):**
```go
config.FailFast = true
result, err := ProcessParallel(ctx, config, items, processor, progress)
if err != nil {
    // First error stops all processing
    // Partial results available in error
}
```

**FailFast=false:**
```go
config.FailFast = false
result, err := ProcessParallel(ctx, config, items, processor, progress)
if err != nil {
    // Only returned if ALL items failed
}
if len(result.Errors) > 0 {
    // Some items failed, some succeeded
    // Check individual errors
}
```

**Error Return Conditions:**
- FailFast=true: ANY item fails → return error
- FailFast=false: ALL items fail → return error
- FailFast=false with partial success: NO error, check result.Errors

### 5. Context Cancellation Modes

**FailFast=true:**
- ProcessParallel creates cancellable child context
- First error calls cancel()
- All workers receive ctx.Done()
- Workers stop processing remaining items

**FailFast=false:**
- ProcessParallel uses original context (no cancellation on item failure)
- Workers continue processing all items
- All errors collected in result.Errors
- User can still cancel via original context

### 6. Progress Callback

Track execution progress across all workers:

```go
progressCallback := func(completed int, total int, result SentimentResult) {
    percentage := (completed * 100) / total
    fmt.Printf("Progress: %d/%d (%d%%) - Latest: Review #%d\n",
        completed, total, percentage, result.ReviewID)
}
```

**Behavior:**
- Called after each successful task completion
- NOT called before first task or on errors
- Receives latest completed result
- Thread-safe (uses atomic counter internally)
- Order of callbacks may not match item order (concurrent execution)

### 7. Observer Events

The slog observer emits JSON events for all parallel operations:

**Parallel Events:**
- `parallel.start` - Before processing begins
- `parallel.complete` - After all workers finish

**Worker Events:**
- `worker.start` - Before each task processes (includes worker_id)
- `worker.complete` - After each task (includes worker_id and error flag)

**Example Event:**
```json
{
  "type": "worker.complete",
  "source": "workflows.ProcessParallel",
  "data": {
    "worker_id": 2,
    "item_index": 7,
    "total_items": 12,
    "error": false
  }
}
```

**Observing Concurrency:**
Multiple `worker.start` events with same timestamp show true parallelism.

## Prerequisites

### 1. Ollama with Model

The example requires [Ollama](https://ollama.ai) running with:
- `llama3.2:3b` - For sentiment-analyst agent

**Quick Start with Docker Compose:**

```bash
# From repository root
docker-compose up -d

# Verify model is available
docker exec tau-orchestrate-ollama ollama list
```

### 2. Environment

- Go 1.23 or later
- Docker (for Ollama container)
- NVIDIA GPU (optional, significantly improves performance)

**GPU Performance:**
- First 4 requests: ~2s each (model loading)
- Cached requests: ~130ms each (GPU accelerated)
- Without GPU: ~800-1500ms per request

## Running the Example

### Option 1: Using go run

```bash
# From repository root
go run examples/phase-05-parallel-execution/main.go
```

### Option 2: Build and run

```bash
# Build
go build -o bin/phase-05-parallel-execution examples/phase-05-parallel-execution/main.go

# Run
./bin/phase-05-parallel-execution
```

## Expected Output

The example produces two types of output:

1. **Human-readable progress** showing concurrent execution
2. **JSON observer events** showing worker activity

```
=== Product Review Sentiment Analysis - Parallel Execution Example ===

1. Configuring observability...
  ✓ Registered slog observer

2. Loading agent configuration...
  ✓ Created sentiment-analyst agent (llama3.2:3b)

3. Preparing product reviews...
  ✓ Loaded 12 product reviews

4. Configuring parallel processing...
  ✓ Parallel configuration ready
    Worker cap: 4
    Fail-fast: false (collect all errors)

5. Defining sentiment analysis processor...
  ✓ Task processor defined

6. Configuring progress tracking...
  ✓ Progress callback configured

7. Executing parallel sentiment analysis...

  Processing 12 reviews concurrently...

{"type":"parallel.start","data":{"worker_count":4,"item_count":12,...}}
{"type":"worker.start","data":{"worker_id":0,"item_index":0,...}}
{"type":"worker.start","data":{"worker_id":1,"item_index":1,...}}
{"type":"worker.start","data":{"worker_id":2,"item_index":2,...}}
{"type":"worker.start","data":{"worker_id":3,"item_index":3,...}}

  Progress: 1/12 reviews analyzed (8%) - Latest: Review #2 (NEUTRAL)

{"type":"worker.complete","data":{"worker_id":0,"item_index":0,"error":false}}
{"type":"worker.start","data":{"worker_id":0,"item_index":4,...}}

  Progress: 2/12 reviews analyzed (16%) - Latest: Review #3 (NEUTRAL)

[... workers continue processing ...]

  Progress: 12/12 reviews analyzed (100%) - Latest: Review #12 (positive)

{"type":"parallel.complete","data":{"items_processed":12,"items_failed":0}}

  ✓ Parallel processing completed successfully

8. Sentiment Analysis Results

   Analyzed 12/12 reviews successfully

   Individual Results (in original order):

   [1] Wireless Mouse
       Review: Excellent mouse! Great battery life...
       ✓ Sentiment: positive

   [2] USB-C Cable
       Review: Cable stopped working after 2 weeks...
       ✓ Sentiment: negative

   [... all 12 reviews in order ...]

9. Sentiment Summary

   Positive: 6 (50.0%)
   Neutral:  3 (25.0%)
   Negative: 3 (25.0%)

10. Performance Metrics

   Total Duration: 2.5s
   Reviews Processed: 12/12
   Success Rate: 100.0%
   Average Time per Review: 207ms
   Throughput: 4.82 reviews/second

   Concurrency:
     Worker Cap: 4
     Estimated Speedup: 3.8x

=== Sentiment Analysis Complete ===
```

## Configuration

### Agent Configuration

`config.llama.json` contains the base configuration:

```json
{
  "name": "llama-agent",
  "client": {
    "provider": {
      "name": "ollama",
      "base_url": "http://localhost:11434",
      "model": {
        "name": "llama3.2:3b",
        "capabilities": {
          "chat": {
            "max_tokens": 150
          }
        }
      }
    }
  }
}
```

### Parallel Configuration

```go
parallelConfig := config.DefaultParallelConfig()
parallelConfig.Observer = "slog"      // JSON event logging
parallelConfig.FailFast = false       // Collect all errors
parallelConfig.WorkerCap = 4          // Limit concurrent workers
parallelConfig.MaxWorkers = 0         // 0 = auto-detect
```

**Default Values:**
- `FailFast`: true (stop on first error)
- `WorkerCap`: 16 (maximum concurrent workers)
- `MaxWorkers`: 0 (auto-detect based on CPU)
- `Observer`: "noop" (no observability)

### Customization

**Adjust worker count:**

```go
// Explicit worker count (no auto-detection)
parallelConfig.MaxWorkers = 8  // Exactly 8 workers

// Auto-detection with different cap
parallelConfig.WorkerCap = 10   // Max 10 workers even on 8-core CPU
```

**Change error handling mode:**

```go
// Fail-fast: stop on first error
parallelConfig.FailFast = true

// Collect all: continue processing all items
parallelConfig.FailFast = false
```

**Modify reviews:**

Edit the `reviews` slice in `main.go` to analyze different content:

```go
reviews := []ProductReview{
    {ID: 1, Product: "Your Product", Review: "Your review text..."},
    // ...
}
```

**Remove progress callback:**

```go
result, err := workflows.ProcessParallel(ctx, config, reviews, processor, nil)
```

## Performance Analysis

### Execution Timeline

**Example execution (4 workers, 12 items):**

```
Time    Worker 0    Worker 1    Worker 2    Worker 3
0.0s    Item 0      Item 1      Item 2      Item 3    (model loading)
2.1s    Item 4      Item 5      Item 6      Item 7    (cached)
2.3s    Item 8      Item 9      Item 10     Item 11   (cached)
2.5s    DONE        DONE        DONE        DONE
```

**Sequential comparison (same 12 items):**
- Sequential: 12 * 207ms = ~2.5s (but first 4 would be 2s each = ~10s total)
- Parallel (4 workers): ~2.5s
- Actual speedup: ~4x (because first batch loads model once for all)

### GPU vs CPU Performance

**With GPU (NVIDIA):**
- Model load: ~1-2s (first requests)
- Cached inference: ~100-150ms per review
- Total for 12 reviews: ~2.5s

**Without GPU (CPU only):**
- Model load: ~3-5s (first requests)
- CPU inference: ~800-1500ms per review
- Total for 12 reviews: ~15-20s

**Speedup from GPU:** ~6-8x faster

### Memory Usage

**Per Worker:**
- Goroutine stack: ~8KB (initial)
- Task context: minimal
- Result buffer: ~few hundred bytes

**Total Overhead:**
- 4 workers: ~32KB stacks
- Work queue: 12 items * pointer size
- Result channel: buffered for all results
- **Total: < 1MB for coordination**

### Ollama Performance Characteristics

**Model Loading (First Request):**
- Loads model into GPU memory
- Initializes KV cache
- Takes 1-2s for llama3.2:3b

**Cached Requests:**
- Model already in memory
- KV cache warmed up
- Only inference time: ~100-150ms

**Concurrent Requests:**
- Ollama handles concurrency internally
- Queues requests if needed
- 4 concurrent workers work well

## Observer Output Analysis

### Understanding JSON Events

Each worker event includes worker identification:

```json
{
  "type": "worker.start",
  "data": {
    "worker_id": 2,
    "item_index": 7,
    "total_items": 12
  }
}
```

### Tracking Concurrency

**Identify parallel execution:**
Look for multiple `worker.start` events with same/similar timestamps:

```
21:17:50.434 worker.start worker_id=0 item_index=0
21:17:50.434 worker.start worker_id=1 item_index=1
21:17:50.434 worker.start worker_id=2 item_index=2
21:17:50.434 worker.start worker_id=3 item_index=3
```

All started within milliseconds = true parallelism.

**Track worker activity:**
Follow a specific worker through the logs:

```bash
# Show only worker 2's activity
go run main.go 2>&1 | grep '"worker_id":2'
```

### Event Sequence

Complete parallel execution produces:

1. `parallel.start` with worker_count and item_count
2. First wave of `worker.start` (N workers)
3. `worker.complete` + `worker.start` pairs (as workers finish and pick up new items)
4. Final wave of `worker.complete` (last items)
5. `parallel.complete` with totals

### Filtering Events

```bash
# Show only parallel lifecycle
go run main.go 2>&1 | grep '"type":"parallel\.'

# Show worker completion events
go run main.go 2>&1 | grep '"type":"worker.complete"'

# Show errors only
go run main.go 2>&1 | grep '"error":true'

# Pretty-print with jq
go run main.go 2>&1 | jq 'select(.type | startswith("worker"))'
```

## Key Code Patterns

### Independent Task Processing

Unlike sequential chains, parallel tasks don't share state:

```go
// Sequential: receives accumulated state
processor := func(ctx context.Context, item Item, state State) (State, error)

// Parallel: independent execution
processor := func(ctx context.Context, item Item) (Result, error)
```

### Result Aggregation

Collect successful results and errors separately:

```go
result, err := ProcessParallel(ctx, config, items, processor, progress)

// Map results by ID for easy lookup
resultMap := make(map[int]Result)
for _, r := range result.Results {
    resultMap[r.ID] = r
}

// Map errors by item ID
errorMap := make(map[int]error)
for _, taskErr := range result.Errors {
    errorMap[taskErr.Item.ID] = taskErr.Err
}

// Access in original order
for _, item := range items {
    if result, ok := resultMap[item.ID]; ok {
        // Handle success
    } else if err, ok := errorMap[item.ID]; ok {
        // Handle error
    }
}
```

### Partial Success Handling

When FailFast=false, handle partial success:

```go
config.FailFast = false
result, err := ProcessParallel(ctx, config, items, processor, nil)

if err != nil {
    // ALL items failed
    log.Fatal("Complete failure")
}

successCount := len(result.Results)
errorCount := len(result.Errors)

if errorCount > 0 {
    fmt.Printf("Partial success: %d/%d succeeded\n",
        successCount, len(items))

    // Process successful results
    for _, r := range result.Results {
        // ...
    }

    // Handle failures
    for _, taskErr := range result.Errors {
        log.Printf("Item %d failed: %v", taskErr.Index, taskErr.Err)
    }
}
```

### Context Timeout

Set overall timeout for parallel execution:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

result, err := ProcessParallel(ctx, config, items, processor, nil)
if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        fmt.Println("Processing timed out")
        // result.Results contains completed items
        // result.Errors contains cancelled items
    }
}
```

### Empty Input Handling

ProcessParallel gracefully handles empty input:

```go
emptyItems := []ProductReview{}

result, err := ProcessParallel(ctx, config, emptyItems, processor, nil)
// Returns:
// - Results = []
// - Errors = []
// - err = nil
```

## Integration Patterns

### With State Graphs

Parallel processing works naturally as state graph nodes:

```go
// Node that processes a list in parallel
parallelNode := state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
    items, _ := s.Get("items")

    processor := func(ctx context.Context, item Item) (Result, error) {
        // Process each item independently
        return result, nil
    }

    result, err := workflows.ProcessParallel(ctx, config, items.([]Item), processor, nil)
    if err != nil {
        return s, err
    }

    // Store results in state
    return s.Set("results", result.Results), nil
})

graph.AddNode("process_parallel", parallelNode)
```

### With Sequential Chains

Combine parallel and sequential processing:

```go
// Sequential chain where each step processes items in parallel
stepProcessor := func(ctx context.Context, batch []Item, results []Result) ([]Result, error) {
    // Process this batch in parallel
    taskProcessor := func(ctx context.Context, item Item) (Result, error) {
        return processItem(item)
    }

    result, err := workflows.ProcessParallel(ctx, config, batch, taskProcessor, nil)
    if err != nil {
        return results, err
    }

    return append(results, result.Results...), nil
}

batches := [][]Item{batch1, batch2, batch3}
result, _ := workflows.ProcessChain(ctx, chainConfig, batches, []Result{}, stepProcessor, nil)
```

### With Hub Coordination

Processor can use hub for multi-agent workflow:

```go
processor := func(ctx context.Context, review ProductReview) (SentimentResult, error) {
    // Distribute to specialist agents via hub
    responseChannel := make(chan string, 1)

    hub.Send(ctx, coordinatorID, sentimentAgentID, review.Review)

    // Collect response
    sentiment := <-responseChannel

    return SentimentResult{Sentiment: sentiment}, nil
}
```

## Performance Considerations

### Optimal Worker Count

**I/O-Bound Tasks (Agent API Calls):**
```go
// Good: 2x CPU cores
config.MaxWorkers = 0  // Auto: runtime.NumCPU() * 2
```

**CPU-Bound Tasks:**
```go
// Good: 1x CPU cores
config.MaxWorkers = runtime.NumCPU()
```

**Memory-Intensive Tasks:**
```go
// Limit to prevent OOM
config.WorkerCap = 4
```

### When to Use Parallel vs Sequential

**Use Parallel When:**
- Items are independent (no dependencies)
- Order doesn't matter during processing
- Throughput more important than order
- Tasks are I/O bound (API calls, file operations)

**Use Sequential When:**
- Items depend on previous results
- State accumulates across items
- Order matters during processing
- Memory-constrained environment

### Batching Strategy

For large datasets, process in batches:

```go
batchSize := 100
for i := 0; i < len(items); i += batchSize {
    end := min(i+batchSize, len(items))
    batch := items[i:end]

    result, err := ProcessParallel(ctx, config, batch, processor, nil)
    if err != nil {
        return err
    }

    // Process batch results
    storeBatchResults(result.Results)
}
```

### Rate Limiting

Prevent overwhelming downstream services:

```go
import "golang.org/x/time/rate"

limiter := rate.NewLimiter(rate.Every(100*time.Millisecond), 10)

processor := func(ctx context.Context, item Item) (Result, error) {
    // Wait for rate limiter
    if err := limiter.Wait(ctx); err != nil {
        return Result{}, err
    }

    // Process item
    return processItem(item)
}
```

## What's Next

This example demonstrates Phase 5 capabilities (Parallel Execution). Related patterns:

- **Phase 2+3 State Graphs** (`examples/phase-02-03-state-graphs/`) - Conditional workflows
- **Phase 4 Sequential Chains** (`examples/phase-04-sequential-chains/`) - Sequential processing
- **Phase 6 Checkpointing** - Resumable workflows with state snapshots

### Combining Patterns

Parallel execution composes naturally with other patterns:

```go
// Parallel tasks within sequential chain
chainStep := func(ctx context.Context, batch []Item, state State) (State, error) {
    result, _ := ProcessParallel(ctx, config, batch, processor, nil)
    return state.Set("results", result.Results), nil
}

// Parallel processing in graph node
graphNode := state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
    items, _ := s.Get("items")
    result, _ := ProcessParallel(ctx, config, items, processor, nil)
    return s.Set("processed", result.Results), nil
})

// Multiple parallel chains for different datasets
var wg sync.WaitGroup
for _, dataset := range datasets {
    wg.Add(1)
    go func(items []Item) {
        defer wg.Done()
        ProcessParallel(ctx, config, items, processor, nil)
    }(dataset)
}
wg.Wait()
```

## Troubleshooting

**Agent not responding:**
- Verify Ollama is running: `curl http://localhost:11434/api/tags`
- Check model is available: `ollama list | grep llama3.2`
- Ensure config file points to correct Ollama URL

**Slow execution (>1s per item):**
- Check if GPU is being used: `docker logs tau-orchestrate-ollama | grep "offloaded"`
- Verify NVIDIA drivers: `nvidia-smi`
- Reduce max_tokens for faster responses
- First requests load model (expected delay)

**Deadlock or hanging:**
- Should never happen with background collector
- Check if context was cancelled
- Verify no blocking operations in processor
- Increase result channel buffer (not normally needed)

**Out of order results:**
- Results maintain original order in result.Results
- Progress callback order may vary (concurrent execution)
- Check if you're iterating result.Results directly

**High memory usage:**
- Reduce WorkerCap to limit concurrent goroutines
- Process items in batches
- Check for memory leaks in processor function

**No concurrency (sequential execution):**
- Check worker_count in parallel.start event
- Verify MaxWorkers and WorkerCap settings
- With <4 items and WorkerCap=4, may use fewer workers
- Check if Ollama is queuing requests internally

**Progress callback not called:**
- Callback only fires after successful tasks
- Not called before first task or on errors
- Verify callback function is not nil

**No JSON events visible:**
- Events go to stdout (mixed with human output)
- Use `2>&1 | grep '"type":"worker"'` to filter
- Verify observer is "slog" not "noop"

**Context timeout errors:**
- Increase timeout duration
- First requests load model (takes 1-2s)
- Check network latency to Ollama
- Reduce number of concurrent workers

**All items marked as errors:**
- Check FailFast setting
- Verify processor function signature matches
- Look for panic in processor (caught as error)
- Check agent initialization succeeded
