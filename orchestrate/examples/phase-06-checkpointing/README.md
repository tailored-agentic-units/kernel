# Phase 6: Multi-Stage Data Analysis with Checkpoint Recovery

This example demonstrates **Phase 6 checkpointing capabilities** through a multi-stage data analysis pipeline that simulates a real-world scenario where long-running workflows can fail partway through execution and need to resume from checkpoints without losing completed work.

## Overview

Scientific research often involves multi-stage data processing pipelines where each stage can take significant time and computational resources. When failures occur (system crashes, network interruptions, resource exhaustion), restarting from the beginning wastes time and money. Checkpointing provides fault tolerance by saving progress at key stages and enabling resume from the last successful checkpoint.

This example shows:
- **Checkpoint save** at configurable intervals during graph execution
- **Simulated failure** during the analysis stage
- **Resume from checkpoint** skipping completed stages
- **Time and cost savings** from not repeating work
- **Observer integration** capturing checkpoint lifecycle events
- **Production patterns** for fault-tolerant workflows

## Scenario

A climate research team processes large datasets through a 4-stage pipeline:

```
┌─────────────┐    ┌──────────────┐    ┌──────────┐    ┌────────────────┐
│   Stage 1   │───▶│   Stage 2    │───▶│ Stage 3  │───▶│    Stage 4     │
│  Ingestion  │    │ Preprocessing│    │ Analysis │    │ Report Gen     │
│  (~2-3s)    │    │   (~2-3s)    │    │  (~2-3s) │    │    (~2-3s)     │
└─────────────┘    └──────────────┘    └──────────┘    └────────────────┘
       ↓                  ↓                   ↓                 ↓
  Checkpoint 1       Checkpoint 2        Checkpoint 3     Checkpoint 4
```

**First Execution**: Pipeline runs through stages 1-2, then **fails** during stage 3 (simulated system failure). Checkpoint saved at stage 2.

**Resume Execution**: Pipeline loads checkpoint from stage 2, **skips** stages 1-2 (already completed), continues from stage 3, and completes successfully.

**Result**: Time saved by not repeating expensive stages 1-2.

## Architecture

### State Graph with Checkpointing

```go
// Configure checkpointing
graphConfig := config.DefaultGraphConfig("data-pipeline")
graphConfig.Checkpoint.Store = "memory"      // Checkpoint store implementation
graphConfig.Checkpoint.Interval = 1          // Checkpoint every 1 node
graphConfig.Checkpoint.Preserve = true       // Keep checkpoints after success

graph, _ := state.NewGraph(graphConfig)

// Build pipeline
graph.AddNode("ingest", ingestNode)
graph.AddNode("preprocess", preprocessNode)
graph.AddNode("analyze", analyzeNode)
graph.AddNode("report", reportNode)

graph.AddEdge("ingest", "preprocess", nil)
graph.AddEdge("preprocess", "analyze", nil)
graph.AddEdge("analyze", "report", nil)

graph.SetEntryPoint("ingest")
graph.SetExitPoint("report")
```

### Checkpoint Lifecycle

```go
// First execution (fails partway)
initialState := state.New(observer)
initialState = initialState.Set("dataset", "climate-research-2024")

runID := initialState.RunID  // Capture for resume

state, err := graph.Execute(ctx, initialState)
// err: "analysis interrupted: simulated system failure"
// state.CheckpointNode: "preprocess" (last successful stage)

// Resume execution
state, err = graph.Resume(ctx, runID)
// Loads checkpoint from memory store
// Skips completed stages (ingest, preprocess)
// Continues from "analyze" node
// Completes successfully
```

### Checkpoint Metadata

State carries checkpoint provenance through execution:

```go
type State struct {
    Data           map[string]any         `json:"data"`
    Observer       observability.Observer `json:"-"`
    RunID          string                 `json:"run_id"`
    CheckpointNode string                 `json:"checkpoint_node"`
    Timestamp      time.Time              `json:"timestamp"`
}

// Direct field access
state.RunID           // "8f9b50d4-385c-4b46-90b8-785c5e453254"
state.CheckpointNode  // "preprocess"
state.Timestamp       // 2025-11-12 10:56:58
```

## Key Concepts

### Checkpoint Interval

Controls how frequently checkpoints are saved:

```go
Interval: 0  // Checkpointing disabled
Interval: 1  // Checkpoint after every node
Interval: 5  // Checkpoint every 5 nodes
```

**Trade-off**: Frequent checkpoints provide fine-grained recovery but increase overhead. Infrequent checkpoints reduce overhead but lose more progress on failure.

**This Example**: Uses `Interval: 1` to demonstrate checkpoint at every stage.

### Checkpoint Store

Abstraction for persistence:

```go
type CheckpointStore interface {
    Save(state State) error
    Load(runID string) (State, error)
    Delete(runID string) error
    List() ([]string, error)
}
```

**Implementations**:
- `memory` - In-memory storage (development/testing, this example)
- Custom stores via registry (disk, database for production)

**This Example**: Uses `Store: "memory"` - checkpoints lost when process terminates.

### Resume Semantics

Checkpoints are saved **after** node execution completes:

```
Execute Node → ✓ Success → Save Checkpoint → Move to Next Node
```

When resuming:
1. Load checkpoint State (includes all data from completed node)
2. Checkpoint represents **completed work**
3. Resume **skips to next node** after checkpoint
4. Continue execution forward

**This Example**:
- Checkpoint at "preprocess" means stage 2 completed successfully
- Resume continues from "analyze" (stage 3)
- Stages 1-2 are skipped (already done)

### Preserve Flag

Controls checkpoint cleanup:

```go
Preserve: false  // Auto-delete checkpoints on successful completion
Preserve: true   // Keep checkpoints after success (audit/debugging)
```

**This Example**: Uses `Preserve: true` to demonstrate checkpoint persistence for the resume operation.

### Observer Events

Checkpointing emits three new event types:

```json
{"type":"checkpoint.save","data":{"node":"preprocess","run_id":"8f9b50d4..."}}
{"type":"checkpoint.load","data":{"node":"preprocess","run_id":"8f9b50d4..."}}
{"type":"checkpoint.resume","data":{"checkpoint_node":"preprocess","resume_node":"analyze","run_id":"8f9b50d4..."}}
```

## Prerequisites

1. **Go 1.23+** installed
2. **Ollama** running locally with **llama3.2:3b** model:
   ```bash
   ollama pull llama3.2:3b
   ollama serve
   ```
3. **tau-core** and **tau-orchestrate** packages

## Running the Example

From the repository root:

```bash
go run examples/phase-06-checkpointing/main.go
```

### What Happens

**Phase 1: Configuration**
- Registers slog observer for JSON event logging
- Creates data-analyst agent with llama3.2:3b
- Builds state graph with checkpointing enabled

**Phase 2: First Execution (Will Fail)**
- Executes stage 1 (ingestion) → Checkpoint saved
- Executes stage 2 (preprocessing) → Checkpoint saved
- Attempts stage 3 (analysis) → **FAILS** (simulated)
- Returns ExecutionError with checkpoint at stage 2

**Phase 3: Resume from Checkpoint**
- Loads checkpoint using RunID
- Skips stages 1-2 (already completed)
- Executes stage 3 (analysis) → Success, checkpoint saved
- Executes stage 4 (report) → Success, checkpoint saved
- Pipeline completes successfully

## Expected Output

### Execution 1: Initial Run (Fails at Stage 3)

```
=== Multi-Stage Data Analysis with Checkpoint Recovery ===

1. Configuring observability...
  ✓ Registered slog observer

2. Loading agent configuration...
  ✓ Created data-analyst agent (llama3.2:3b)

3. Creating data analysis pipeline with checkpointing...
  ✓ Created state graph with checkpointing enabled
     - Checkpoint interval: Every 1 node
     - Checkpoint store: memory
     - Preserve on success: true

4. Defining pipeline stages...
  ✓ Defined 4 pipeline stages
     - ingest → preprocess → analyze → report

5. Building pipeline graph...
  ✓ Pipeline graph constructed

=                                                            =
EXECUTION 1: Initial Run (Will Fail)
=                                                            =

Pipeline RunID: 8f9b50d4-385c-4b46-90b8-785c5e453254

  → STAGE 1: Data Ingestion
     Loading research dataset...
     Characteristics: The 'climate-research-2024' dataset contains approximately 1.5 million entries,
                      with an average record length of 10 variables and 12 time points (yearly from
                      1970 to 2100). The most prominent variable trends reveal rising CO2 levels,
                      increasing global temperatures, and notable regional climate variability.
     ✓ Stage 1 complete

  → STAGE 2: Preprocessing
     Cleaning and normalizing data...
     Steps: Based on the dataset characteristics, preliminary preprocessing steps may include handling
            missing values for all variables using techniques such as imputation or interpolation, and
            then scaling/normalizing the data using Standardization Min-Max Scaler (SMASS) or Robust
            Scaler to account for variable differences in ranges and distributions.
     ✓ Stage 2 complete

  → STAGE 3: Analysis
     Running statistical analysis...
     ✗ SIMULATED FAILURE: Analysis process interrupted

❌ EXECUTION FAILED after 5.90s
   Error: execution failed at node analyze: node execution failed: analysis interrupted: simulated system failure
   Checkpoint saved at: preprocess
```

### Execution 2: Resume from Checkpoint (Succeeds)

```
=                                                            =
EXECUTION 2: Resume from Checkpoint
=                                                            =

Resuming pipeline from RunID: 8f9b50d4-385c-4b46-90b8-785c5e453254
Last completed stage: preprocess

Note: Stages 1-2 will be skipped (already completed)
      Execution resumes from Stage 3

  → STAGE 3: Analysis
     Running statistical analysis...
     Insights: Analyzing the 'climate-research-2024' dataset reveals correlations between temperature
               increase and greenhouse gas emissions, with a significant positive linear relationship
               found across 75% of the dataset. Additionally, time-series analysis detects anomalies
               in CO2 levels from 2018 to 2020, suggesting potential disruptions in global climate patterns.
     ✓ Stage 3 complete

  → STAGE 4: Report Generation
     Generating final report...
     Summary: Conclusion: Our analysis of the 'climate-research-2024' dataset reveals a strong positive
              correlation between temperature increase and greenhouse gas emissions, with anomalous events
              detected in CO2 levels during 2018-2020, indicating potential perturbations in global climate
              patterns. These findings highlight the need for further investigation into the underlying
              causes and implications of these disruptions for our understanding of climate dynamics.
     ✓ Stage 4 complete

✓ Pipeline completed successfully after resume!
   Resume execution time: 3.65s
   Total time (initial + resume): 9.55s
   Time saved by checkpointing: ~2-3s (skipped stages 1-2)

=                                                            =
FINAL RESULTS
=                                                            =

Report Summary:
Conclusion: Our analysis of the 'climate-research-2024' dataset reveals a strong positive correlation
between temperature increase and greenhouse gas emissions, with anomalous events detected in CO2 levels
during 2018-2020, indicating potential perturbations in global climate patterns.

Checkpoint Demonstration Summary:
  ✓ Initial execution failed at Stage 3
  ✓ Checkpoint preserved progress through Stage 2
  ✓ Resume skipped completed stages (1-2)
  ✓ Execution continued from Stage 3
  ✓ Pipeline completed successfully
  ✓ Time and cost savings demonstrated
```

### Observer Events (JSON Logs)

**Checkpoint Save Events** (after each node):
```json
{"time":"2025-11-12T10:56:55.966425-05:00","level":"INFO","msg":"Event",
 "type":"checkpoint.save","source":"data-pipeline","timestamp":"2025-11-12T10:56:55.966424-05:00",
 "data":{"node":"ingest","run_id":"8f9b50d4-385c-4b46-90b8-785c5e453254"}}

{"time":"2025-11-12T10:56:58.038587-05:00","level":"INFO","msg":"Event",
 "type":"checkpoint.save","source":"data-pipeline","timestamp":"2025-11-12T10:56:58.038587-05:00",
 "data":{"node":"preprocess","run_id":"8f9b50d4-385c-4b46-90b8-785c5e453254"}}
```

**Checkpoint Load/Resume Events** (when resuming):
```json
{"time":"2025-11-12T10:57:00.038781-05:00","level":"INFO","msg":"Event",
 "type":"checkpoint.load","source":"data-pipeline","timestamp":"2025-11-12T10:57:00.038781-05:00",
 "data":{"node":"preprocess","run_id":"8f9b50d4-385c-4b46-90b8-785c5e453254"}}

{"time":"2025-11-12T10:57:00.038819-05:00","level":"INFO","msg":"Event",
 "type":"checkpoint.resume","source":"data-pipeline","timestamp":"2025-11-12T10:57:00.038819-05:00",
 "data":{"checkpoint_node":"preprocess","resume_node":"analyze","run_id":"8f9b50d4-385c-4b46-90b8-785c5e453254"}}
```

## Configuration Details

### GraphConfig with Checkpointing

```go
type GraphConfig struct {
    Name          string           // "data-pipeline"
    Observer      string           // "slog"
    MaxIterations int              // 10
    Checkpoint    CheckpointConfig // Checkpoint configuration
}

type CheckpointConfig struct {
    Store    string  // "memory" - CheckpointStore implementation name
    Interval int     // 1 - Checkpoint every 1 node
    Preserve bool    // true - Keep checkpoints after success
}
```

### Agent Configuration

Located in `config.llama.json`:

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

System prompt configures agent as scientific data analyst with concise responses (1-2 sentences per stage).

## Integration Patterns

### With Other Phases

Checkpointing integrates seamlessly with other orchestration patterns:

**State Graphs (Phase 2-3)**:
```go
// Checkpointing works with conditional routing, cycles, multiple exit points
graph.AddEdge("validate", "process", state.KeyEquals("valid", true))
graph.AddEdge("validate", "error", state.KeyEquals("valid", false))
graphConfig.Checkpoint.Interval = 3  // Checkpoint every 3 nodes
```

**Sequential Chains (Phase 4)**:
```go
// Wrap chain execution in checkpointed graph node
chainNode := state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
    items, _ := s.Get("items")
    result := workflows.ProcessChain(ctx, items, processor, chainConfig)
    return s.Set("chain_result", result), nil
})
graph.AddNode("chain", chainNode)
```

**Parallel Execution (Phase 5)**:
```go
// Checkpoints preserve parallel processing results
parallelNode := state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
    items, _ := s.Get("items")
    result := workflows.ProcessParallel(ctx, items, processor, parallelConfig)
    return s.Set("parallel_result", result), nil
})
graph.AddNode("parallel", parallelNode)
```

### Production Patterns

**Long-Running Workflows**:
```go
// Checkpoint after expensive stages
graphConfig.Checkpoint.Interval = 1  // After each major stage
graphConfig.Checkpoint.Store = "postgres"  // Persist to database
graphConfig.Checkpoint.Preserve = false  // Cleanup on success
```

**Multi-Day Processing**:
```go
// Daily batch processing with resume on failure
graphConfig.Checkpoint.Interval = 10  // Every 10 records processed
graphConfig.Checkpoint.Store = "s3"  // Cloud storage
graphConfig.Checkpoint.Preserve = true  // Audit trail
```

**Human-in-the-Loop**:
```go
// Checkpoint before human review stages
reviewNode := state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
    // Save checkpoint before waiting for human input
    s.Checkpoint(store)
    // Wait for human review (could take hours/days)
    // Resume continues after review
})
```

## Troubleshooting

### Checkpoint Not Found Error

**Error**: `failed to load checkpoint: checkpoint not found: <runID>`

**Cause**: Memory store is in-process only. Checkpoints lost when program terminates.

**Solutions**:
1. Use persistent store (file/database) for production
2. Keep process running between executions for testing
3. Verify RunID matches between save and load

### Resume Skips Too Much Work

**Issue**: Resume skips more stages than expected

**Cause**: Checkpoint saved at wrong node

**Debug**:
```go
fmt.Printf("Checkpoint at: %s\n", state.CheckpointNode)
fmt.Printf("Resume will skip to next node after checkpoint\n")
```

**Solution**: Adjust checkpoint interval or verify checkpoint placement

### Ollama Connection Errors

**Error**: `failed to connect to ollama`

**Cause**: Ollama not running or wrong base URL

**Solutions**:
```bash
# Start Ollama
ollama serve

# Verify model available
ollama list | grep llama3.2

# Pull if needed
ollama pull llama3.2:3b
```

### High Memory Usage with Frequent Checkpoints

**Issue**: Memory grows with many checkpoints

**Cause**: Preserve=true keeps all checkpoints in memory

**Solutions**:
1. Set `Preserve: false` to enable auto-cleanup
2. Increase checkpoint interval
3. Implement cleanup logic:
```go
store.Delete(oldRunID)  // Manual cleanup
```

## Code Structure

### Main Components

```go
main.go                              // ~350 lines
├── Configuration (lines 1-70)       // Observer, agent, graph setup
├── Node Definitions (lines 72-180)  // 4 pipeline stage nodes
│   ├── ingestNode                   // Data ingestion with LLM
│   ├── preprocessNode               // Preprocessing with LLM
│   ├── analyzeNode                  // Analysis with simulated failure
│   └── reportNode                   // Report generation with LLM
├── Graph Construction (lines 182-220) // Nodes, edges, entry/exit
├── Execution 1 (lines 222-260)      // Initial run (fails)
├── Execution 2 (lines 262-300)      // Resume (succeeds)
└── Results Display (lines 302-350)  // Final output and summary
```

### Key Features Demonstrated

**Lines 67-70**: Checkpoint configuration
```go
graphConfig.Checkpoint.Store = "memory"
graphConfig.Checkpoint.Interval = 1
graphConfig.Checkpoint.Preserve = true
```

**Lines 122-130**: Simulated failure logic
```go
if !firstExecutionFailed && analysisAttempts == 1 {
    firstExecutionFailed = true
    return s, fmt.Errorf("analysis interrupted: simulated system failure")
}
```

**Lines 242-245**: Capturing RunID for resume
```go
initialState := state.New(observer)
initialState = initialState.Set("dataset", "climate-research-2024")
runID := initialState.RunID  // Needed for Resume
```

**Lines 272-275**: Resume execution
```go
resumedState, err := graph.Resume(ctx, runID)
// Loads checkpoint, skips completed stages, continues execution
```

## Learning Points

### When to Use Checkpointing

✅ **Good Use Cases**:
- Long-running workflows (minutes to hours)
- Expensive operations (LLM calls, data processing, API calls)
- Multi-stage pipelines where stages are independent
- Production workflows requiring fault tolerance
- Workflows with potential for intermittent failures

❌ **Not Needed For**:
- Fast workflows (< 10 seconds total)
- Simple linear processing with no failure risk
- Workflows that can cheaply restart from beginning
- Development/testing unless testing checkpoint feature

### Checkpoint Interval Strategy

**Fine-grained** (Interval: 1):
- ✅ Minimal work lost on failure
- ✅ Resume from precise point
- ❌ Higher storage overhead
- ❌ More checkpoint save operations

**Coarse-grained** (Interval: 10+):
- ✅ Lower storage overhead
- ✅ Fewer checkpoint operations
- ❌ More work lost on failure
- ❌ Resume from further back

**Optimal**: Checkpoint after natural stage boundaries (data loaded, preprocessing complete, analysis done).

### State Design for Checkpointing

**Good State Design**:
```go
// Self-contained, serializable data
state.Set("dataset_path", "/path/to/data")
state.Set("processing_params", params)
state.Set("results", analysisResults)
```

**Avoid**:
```go
// External resources that don't serialize
state.Set("file_handle", openFile)     // ❌ Won't survive checkpoint
state.Set("database_conn", conn)       // ❌ Won't survive checkpoint
state.Set("channel", make(chan int))   // ❌ Won't serialize
```

**Best Practice**: Store resource locators (paths, IDs, URLs) not the resources themselves.

## Performance Characteristics

**Example Timings** (actual run):
- First execution: 5.90s (stages 1-2 + failure)
- Resume execution: 3.65s (stages 3-4 only)
- Total time: 9.55s
- Time saved: ~2-3s (40% of work skipped)

**Checkpoint Overhead**:
- Memory store: < 1ms per checkpoint
- Checkpoint interval = 1: 4 checkpoints total
- Total checkpoint overhead: < 10ms

**Scaling**:
- 10 stages: Resume saves 4-5s if failing at stage 6
- 100 stages: Resume saves 40-50s if failing at stage 60
- Cost savings scale linearly with stage count and LLM call expense

## Next Steps

After understanding checkpointing basics:

1. **Explore Phase 7**: Conditional routing with checkpoint integration
2. **Production Deployment**: Implement persistent checkpoint stores (PostgreSQL, Redis, S3)
3. **Multi-Workflow**: Coordinate checkpoints across multiple related workflows
4. **Monitoring**: Track checkpoint frequency, storage usage, resume success rates
5. **Testing**: Inject failures at various stages to validate recovery

## Related Examples

- **Phase 2+3** (`examples/phase-02-03-state-graphs`): State graph execution without checkpointing
- **Phase 4** (`examples/phase-04-sequential-chains`): Sequential chains that could benefit from checkpointing
- **Phase 5** (`examples/phase-05-parallel-execution`): Parallel processing with checkpoint integration potential

## References

- **Phase 6 Session Summary**: `_context/sessions/phase-06-checkpointing.md`
- **ARCHITECTURE.md**: Checkpointing design and interfaces
- **PROJECT.md**: Phase 6 requirements and success criteria
- **Source Code**: `pkg/state/checkpoint.go` - CheckpointStore interface and implementations
