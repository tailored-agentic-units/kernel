# Phase 4: Sequential Chains Example - Research Paper Analysis Pipeline

This example demonstrates sequential chain processing with state accumulation, progress tracking, and intermediate state capture using a realistic research paper analysis scenario.

## Overview

The example simulates an automated research paper analysis pipeline where a research-analyst agent processes each section of a paper sequentially:
- **Abstract** → Extract main contribution
- **Introduction** → Identify key problem
- **Methodology** → Summarize research method
- **Results** → List top quantitative results
- **Conclusion** → Extract future work directions

The pipeline demonstrates:
1. **Sequential processing** - Items processed in order with state accumulation
2. **Fold/reduce pattern** - Each step receives and updates accumulated state
3. **Progress tracking** - Real-time completion percentage via callback
4. **Intermediate state capture** - Complete state evolution history
5. **Observer integration** - Chain and step events via slog observer
6. **Direct agent usage** - Primary pattern for LLM integration

## Architecture

### Sequential Chain Pattern

```
Items: [Abstract, Introduction, Methodology, Results, Conclusion]

Initial State  →  Step 1  →  Step 2  →  Step 3  →  Step 4  →  Step 5  →  Final State
   {title,         {main_      {problem_    {method_     {key_        {future_
    timestamp}      contrib}     statement}   ology}       results}     work}

                   ↓           ↓            ↓            ↓            ↓
Progress:          20%         40%          60%          80%          100%
```

### Processing Flow

Each step follows this pattern:

1. **Input**: Current section + accumulated state from previous steps
2. **Process**: Agent analyzes section and extracts specific information
3. **Update**: State updated with new findings (immutably)
4. **Capture**: Intermediate state saved (if enabled)
5. **Progress**: Callback invoked with completion percentage
6. **Continue**: Updated state passed to next step

### Data Structures

**PaperSection** (Item Type):
```go
type PaperSection struct {
    Name    string  // Section name (Abstract, Introduction, etc.)
    Content string  // Section text content
}
```

**state.State** (Context Type):
```go
// Immutable key-value state that accumulates findings
initialState := state.New(observer)
    .Set("paper_title", "Adaptive Sharding for Blockchain Scalability")
    .Set("analysis_start", timestamp)

// After all steps:
finalState contains:
    - paper_title
    - analysis_start
    - main_contribution
    - problem_statement
    - methodology
    - key_results
    - future_work
```

**ChainResult**:
```go
type ChainResult[TContext any] struct {
    Final        TContext      // Final accumulated state
    Intermediate []TContext    // State after each step (when captured)
    Steps        int           // Number of steps completed
}
```

## Key Concepts Demonstrated

### 1. StepProcessor Function

The processor implements the fold/reduce pattern:

```go
stepProcessor := func(ctx context.Context, section PaperSection, s state.State) (state.State, error) {
    // Extract relevant information based on section
    prompt := formatPromptForSection(section)

    // Call agent for analysis
    response, err := agent.Chat(ctx, prompt)
    if err != nil {
        return s, err
    }

    // Update state immutably
    return s.Set(stateKey, response.Content()), nil
}
```

**Key Points:**
- Generic over `TItem` (PaperSection) and `TContext` (state.State)
- Receives current item and accumulated state
- Returns updated state (immutably)
- Errors stop the chain (fail-fast)

### 2. Progress Callback

Track execution progress in real-time:

```go
progressCallback := func(completed int, total int, s state.State) {
    percentage := (completed * 100) / total
    fmt.Printf("Progress: Step %d/%d complete (%d%%)\n", completed, total, percentage)
}
```

**Behavior:**
- Called after each successful step
- NOT called before first step or on errors
- Receives current accumulated state snapshot
- Thread-safe (chain processes sequentially)

### 3. Intermediate State Capture

When `CaptureIntermediateStates` is enabled:

```go
chainConfig := config.DefaultChainConfig()
chainConfig.CaptureIntermediateStates = true

result, err := workflows.ProcessChain(ctx, chainConfig, sections, initialState, processor, progress)

// result.Intermediate contains:
// [0] Initial state (paper metadata)
// [1] After step 1 (+ main_contribution)
// [2] After step 2 (+ problem_statement)
// [3] After step 3 (+ methodology)
// [4] After step 4 (+ key_results)
// [5] After step 5 (+ future_work)
```

**Use Cases:**
- Debugging state transformations
- Visualizing data accumulation
- Implementing undo/rollback
- Performance analysis

### 4. Observer Events

The slog observer emits JSON events for every operation:

**Chain Events:**
- `chain.start` - Before processing begins
- `chain.complete` - After all steps finish

**Step Events:**
- `step.start` - Before each step processes
- `step.complete` - After each step (includes error flag)

**State Events:**
- `state.create` - New state created
- `state.clone` - State cloned for immutability
- `state.set` - Key updated in state

**Example Event:**
```json
{
  "type": "step.complete",
  "source": "workflows.ProcessChain",
  "data": {
    "step_index": 2,
    "total_steps": 5,
    "error": false
  }
}
```

### 5. Generic Type Flexibility

ProcessChain is fully generic and works with any types:

```go
// Research paper analysis (this example)
ProcessChain[PaperSection, state.State](...)

// Document classification
ProcessChain[Document, state.State](...)

// Conversation chain
ProcessChain[string, Conversation](...)

// Data transformation
ProcessChain[RawData, ProcessedData](...)
```

## Prerequisites

### 1. Ollama with Model

The example requires [Ollama](https://ollama.ai) running with:
- `llama3.2:3b` - For research-analyst agent

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
- NVIDIA GPU (optional, improves performance)

## Running the Example

### Option 1: Using go run

```bash
# From repository root
go run examples/phase-04-sequential-chains/main.go
```

### Option 2: Build and run

```bash
# Build
go build -o bin/phase-04-sequential-chains examples/phase-04-sequential-chains/main.go

# Run
./bin/phase-04-sequential-chains
```

## Expected Output

The example produces two types of output:

1. **Human-readable progress** showing step execution
2. **JSON observer events** showing complete execution trace

```
=== Research Paper Analysis Pipeline - Sequential Chains Example ===

1. Configuring observability...
  ✓ Registered slog observer

2. Loading agent configuration...
  ✓ Created research-analyst agent (llama3.2:3b)

3. Preparing research paper sections...
  ✓ Loaded 5 paper sections

4. Configuring sequential analysis chain...
  ✓ Chain configuration ready
    Intermediate state capture: enabled

5. Defining analysis step processor...
  ✓ Step processor defined

6. Configuring progress tracking...
  ✓ Progress callback configured

7. Executing sequential analysis pipeline...

  Starting analysis of 5 paper sections...

{"type":"chain.start","source":"workflows.ProcessChain","data":{"item_count":5,...}}
{"type":"step.start","data":{"step_index":0,"total_steps":5}}

  Progress: Step 1/5 complete (20%)

{"type":"step.complete","data":{"step_index":0,"total_steps":5,"error":false}}
{"type":"step.start","data":{"step_index":1,"total_steps":5}}

  Progress: Step 2/5 complete (40%)

[... steps 3-5 ...]

  Progress: Step 5/5 complete (100%)

{"type":"chain.complete","data":{"steps_completed":5,"error":false}}

  ✓ Analysis pipeline completed

8. Analysis Results

   Paper: Adaptive Sharding for Blockchain Scalability

   Key Findings:

   Main Contribution:
     The main research contribution is the development of an adaptive sharding
     technique that improves transaction throughput by 3x...

   Problem Statement:
     The key problem is the scalability challenge of traditional blockchain
     systems, specifically the trade-off between high transaction throughput
     and decentralization...

   Methodology:
     The research method employed an adaptive sharding protocol that uses a
     reputation-based validator selection mechanism...

   Key Results:
     1. A 3.2x improvement in transaction throughput compared to baseline systems.
     2. Average latency reduced from 12 seconds to 4 seconds...

   Future Work:
     The main future work directions are the exploration of integrating adaptive
     sharding with zero-knowledge proofs...

9. State Evolution Analysis

   Total states captured: 6 (initial + 5 processing steps)

   State progression:
     [0] Initial state (paper metadata)
     [1] After processing: Abstract
     [2] After processing: Introduction
     [3] After processing: Methodology
     [4] After processing: Results
     [5] After processing: Conclusion

10. Execution Metrics
    Duration: 4.8s
    Steps Completed: 5/5
    Intermediate States Captured: 6
    Average Time per Step: 960ms

=== Research Paper Analysis Complete ===
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

### Chain Configuration

```go
chainConfig := config.DefaultChainConfig()
chainConfig.Observer = "slog"                    // JSON event logging
chainConfig.CaptureIntermediateStates = true     // Save state evolution
```

### Customization

**Adjust paper content:**

Modify the `sections` slice in `main.go` to analyze different content:

```go
sections := []PaperSection{
    {Name: "Abstract", Content: "Your abstract text..."},
    {Name: "Introduction", Content: "Your introduction text..."},
    // ...
}
```

**Modify analysis steps:**

Change the `stepProcessor` logic to extract different information:

```go
switch sectionName {
case "Abstract":
    prompt = "Extract key hypotheses from: " + section.Content
    stateKey = "hypotheses"
// ...
}
```

**Disable intermediate capture:**

```go
chainConfig.CaptureIntermediateStates = false  // Reduce memory usage
```

**Remove progress callback:**

```go
result, err := workflows.ProcessChain(ctx, chainConfig, sections, initialState, processor, nil)
```

## Observer Output Analysis

### Understanding JSON Events

Each JSON event has this structure:

```json
{
  "time": "2025-11-07T16:07:02.971558609-05:00",
  "level": "INFO",
  "msg": "Event",
  "type": "step.complete",
  "source": "workflows.ProcessChain",
  "timestamp": "2025-11-07T16:07:02.971555674-05:00",
  "data": {
    "step_index": 2,
    "total_steps": 5,
    "error": false
  }
}
```

**Key Fields:**
- `type` - Event type (chain.*, step.*, state.*)
- `source` - Event origin ("workflows.ProcessChain" or "state")
- `data` - Event-specific metadata

### Event Sequence

A complete chain execution produces this event sequence:

1. **Initialization**: `state.create` → `state.set` (initial state setup)
2. **Chain Start**: `chain.start` with item count and configuration
3. **Per Step**:
   - `step.start` with step index
   - `state.clone` → `state.set` (state update)
   - `step.complete` with error flag
4. **Chain End**: `chain.complete` with final step count

### Filtering Events

Focus on specific event types:

```bash
# Show only step events
go run examples/phase-04-sequential-chains/main.go 2>&1 | grep '"type":"step\.'

# Show only state mutations
go run examples/phase-04-sequential-chains/main.go 2>&1 | grep '"type":"state\.'

# Show chain lifecycle
go run examples/phase-04-sequential-chains/main.go 2>&1 | grep '"type":"chain\.'

# Pretty-print with jq
go run examples/phase-04-sequential-chains/main.go 2>&1 | jq 'select(.type)'
```

## Key Code Patterns

### Immutable State Updates

State updates return new state instances:

```go
// Immutable chaining
newState := s.Set("key1", value1).Set("key2", value2)

// Original state unchanged
s.Get("key1")      // Returns original value (or not exists)
newState.Get("key1")  // Returns new value
```

### Error Handling

Processor errors stop the chain immediately:

```go
processor := func(ctx context.Context, item Item, s State) (State, error) {
    result, err := agent.Chat(ctx, prompt)
    if err != nil {
        return s, fmt.Errorf("processing %s failed: %w", item.Name, err)
    }
    return s.Set("result", result.Content()), nil
}

// On error, ProcessChain returns ChainError with:
// - StepIndex: which step failed
// - Item: what was being processed
// - State: state at time of failure
// - Err: underlying error
```

### Context Cancellation

ProcessChain checks context at the start of each step:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

result, err := workflows.ProcessChain(ctx, config, items, initial, processor, progress)
if err != nil {
    // Check if cancellation caused error
    if errors.Is(err, context.DeadlineExceeded) {
        fmt.Println("Processing timed out")
    }
}
```

### Empty Chain Handling

ProcessChain gracefully handles empty input:

```go
emptyItems := []PaperSection{}

result, err := workflows.ProcessChain(ctx, config, emptyItems, initialState, processor, progress)
// Returns:
// - Final = initialState
// - Steps = 0
// - Intermediate = [initialState] (if capture enabled)
// - err = nil
```

## Integration Patterns

### With State Graphs

Sequential chains work naturally as state graph nodes:

```go
// Node that processes a list sequentially
chainNode := state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
    items, _ := s.Get("items")

    result, err := workflows.ProcessChain(ctx, config, items.([]Item), s, processor, nil)
    if err != nil {
        return s, err
    }

    return result.Final, nil
})

graph.AddNode("process_items", chainNode)
```

### With Hub Coordination

Processor can use hub for multi-agent coordination:

```go
processor := func(ctx context.Context, section PaperSection, s state.State) (state.State, error) {
    // Send section to specialist agent via hub
    response := make(chan string, 1)
    hub.Send(ctx, coordinatorID, specialistID, section.Content)

    // Collect response
    analysis := <-response

    return s.Set(section.Name, analysis), nil
}
```

### Custom Context Types

Use any type for state accumulation:

```go
type AnalysisReport struct {
    Title    string
    Sections map[string]string
    Summary  string
}

processor := func(ctx context.Context, section PaperSection, report AnalysisReport) (AnalysisReport, error) {
    analysis, _ := agent.Chat(ctx, section.Content)
    report.Sections[section.Name] = analysis.Content()
    return report, nil
}

result, _ := workflows.ProcessChain(ctx, config, sections, AnalysisReport{...}, processor, nil)
fmt.Printf("Final report: %+v\n", result.Final)
```

## Performance Considerations

### Memory Usage

Intermediate state capture increases memory:

```go
// High memory: Captures every state
config.CaptureIntermediateStates = true  // ~N * stateSize

// Low memory: Only final state
config.CaptureIntermediateStates = false  // ~1 * stateSize
```

For large chains (1000+ items) or large states, disable capture.

### Processing Time

Sequential chains process one item at a time:

```go
// 5 items × 1 second each = 5 seconds total
// No parallelization within the chain
```

For parallel processing, use `workflows.ProcessParallel` instead (Phase 5).

### Observer Overhead

JSON logging has minimal overhead:

```go
// NoOp observer: ~0% overhead
config.Observer = "noop"

// Slog observer: ~1-2% overhead
config.Observer = "slog"
```

Use noop observer in production if events not needed.

## What's Next

This example demonstrates Phase 4 capabilities (Sequential Chains). Related patterns:

- **Phase 2+3 State Graphs** (`examples/phase-02-03-state-graphs/`) - Conditional workflows
- **Phase 5 Parallel Execution** - Concurrent processing with worker pools
- **Phase 6 Checkpointing** - Resumable workflows with state snapshots

### Combining Patterns

Sequential chains compose naturally with other patterns:

```go
// Chain within a graph node
graph.AddNode("analyze", chainNode)

// Graph within a chain step
processor := func(...) {
    subGraph.Execute(ctx, s)
}

// Parallel chains for different datasets
for dataset := range datasets {
    go workflows.ProcessChain(ctx, config, dataset, ...)
}
```

## Troubleshooting

**Agent not responding:**
- Verify Ollama is running: `curl http://localhost:11434/api/tags`
- Check model is available: `ollama list | grep llama3.2`
- Ensure config file points to correct Ollama URL

**Chain stops mid-execution:**
- Check for processor errors in output
- Verify agent responses are valid
- Increase timeout in context if needed

**No intermediate states captured:**
- Verify `CaptureIntermediateStates = true` in config
- Check `result.Intermediate` length matches `result.Steps + 1`

**Progress callback not called:**
- Callback only fires after successful steps
- Not called before first step
- Not called if step errors occur

**High memory usage:**
- Disable intermediate state capture
- Process items in smaller batches
- Use custom context type with minimal state

**Slow execution:**
- Each step is sequential (not parallel)
- Agent API calls are the bottleneck
- Consider Phase 5 parallel execution for independent items

**No JSON events visible:**
- Events go to stdout (mixed with human output)
- Use `2>&1 | grep '"type"'` to filter
- Verify observer is "slog" not "noop"

**Context timeout errors:**
- Increase timeout duration
- Reduce max_tokens for faster agent responses
- Check network latency to Ollama
