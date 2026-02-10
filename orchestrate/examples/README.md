# tau-orchestrate Examples

This directory contains comprehensive examples demonstrating the orchestration capabilities built on top of [tau-core](https://github.com/tailored-agentic-units/tau-core).

## Prerequisites

All examples require:
- **Go 1.23+**
- **Ollama** running locally with models pulled
- **Docker** (for Ollama container)

### Quick Setup

```bash
# Start Ollama with required models
docker-compose up -d

# Verify models are available
docker exec tau-orchestrate-ollama ollama list
```

**Required Models:**
- `llama3.2:3b` - Used by all examples
- `gemma3:4b` - Used by phase-01-hubs only

## Examples Overview

### Phase 1: Hub & Messaging
**Directory:** [`phase-01-hubs/`](./phase-01-hubs/)

Multi-agent coordination using hubs for agent-to-agent communication, broadcasting, pub/sub, and cross-hub messaging.

**Scenario:** ISS Maintenance EVA operation with 4 agents across 2 hubs

**Run:**
```bash
go run examples/phase-01-hubs/main.go
```

**Demonstrates:**
- Agent registration with multiple hubs
- Direct agent-to-agent messaging
- Broadcast communication (one-to-many)
- Pub/sub with topic subscriptions
- Cross-hub coordination (agents in multiple hubs)
- Hub metrics and observability

**Expected Output:**
- 4 agents communicating via 2 hubs
- EVA crew coordination messages
- Cross-hub relay through mission commander
- Execution time: ~15-20 seconds
- Hub metrics showing message counts

**[Full Documentation →](./phase-01-hubs/README.md)**

---

### Phase 2+3: State Graphs
**Directory:** [`phase-02-03-state-graphs/`](./phase-02-03-state-graphs/)

State graph execution with conditional routing, cycle detection, and multiple exit points.

**Scenario:** Software deployment pipeline with automated test/fix/retry cycles

**Run:**
```bash
go run examples/phase-02-03-state-graphs/main.go
```

**Demonstrates:**
- State graph construction (nodes, edges, predicates)
- Linear progression (plan → build → test → deploy)
- Conditional routing (test success vs failure)
- Cycle detection (test → fix → test retry loops)
- Multiple exit points (deploy success, rollback failure)
- Immutable state flow through pipeline
- Complete execution trace via slog observer

**Expected Output:**
- 6 nodes executed with conditional branching
- Cycle detection events when tests fail and retry
- Either deployment success or rollback after 3 retries
- JSON observer events showing graph execution
- Execution time: ~7-10 seconds
- Final deployment status in state

**[Full Documentation →](./phase-02-03-state-graphs/README.md)**

---

### Phase 4: Sequential Chains
**Directory:** [`phase-04-sequential-chains/`](./phase-04-sequential-chains/)

Sequential processing with state accumulation using fold/reduce pattern.

**Scenario:** Research paper analysis processing 5 sections sequentially

**Run:**
```bash
go run examples/phase-04-sequential-chains/main.go
```

**Demonstrates:**
- Sequential chain processing (fold/reduce pattern)
- State accumulation across steps
- Progress tracking (20%, 40%, 60%, 80%, 100%)
- Intermediate state capture (6 states: initial + 5 steps)
- State evolution from initial → final
- Complete step-by-step observability

**Expected Output:**
- 5 paper sections analyzed in order
- Progress updates after each section
- State growing with each analysis
- Final analysis report with all findings
- JSON observer events for each step
- Execution time: ~5-6 seconds
- State evolution summary showing accumulation

**[Full Documentation →](./phase-04-sequential-chains/README.md)**

---

### Phase 5: Parallel Execution
**Directory:** [`phase-05-parallel-execution/`](./phase-05-parallel-execution/)

Concurrent processing with worker pool coordination and order preservation.

**Scenario:** Product review sentiment analysis processing 12 reviews in parallel

**Run:**
```bash
go run examples/phase-05-parallel-execution/main.go
```

**Demonstrates:**
- Parallel processing with worker pool (4 workers)
- Concurrent execution of independent tasks
- Order preservation (results in original sequence)
- Real-time progress tracking across workers
- Worker coordination and load balancing
- Fail-fast vs collect-all-errors modes
- Performance metrics and throughput

**Expected Output:**
- 12 reviews processed concurrently by 4 workers
- Progress updates showing concurrent completion
- Results displayed in original review order
- Sentiment summary (positive/neutral/negative counts)
- JSON observer events showing worker activity
- Execution time: ~2.5 seconds
- Throughput: ~4-5 reviews/second

**[Full Documentation →](./phase-05-parallel-execution/README.md)**

---

### Phase 6: Checkpointing
**Directory:** [`phase-06-checkpointing/`](./phase-06-checkpointing/)

Workflow persistence and recovery through checkpoint save/resume.

**Scenario:** Multi-stage data analysis pipeline with simulated failure and recovery

**Run:**
```bash
go run examples/phase-06-checkpointing/main.go
```

**Demonstrates:**
- Checkpoint save at configurable intervals
- State persistence across execution failures
- Resume execution from saved checkpoints
- Progress preservation (skipping completed work)
- Observer integration (checkpoint events)
- Production fault tolerance patterns

**Expected Output:**
- 4-stage pipeline: ingest → preprocess → analyze → report
- First execution fails at stage 3 (simulated failure)
- Checkpoint saved at stage 2 completion
- Resume skips stages 1-2, continues from stage 3
- Pipeline completes successfully on resume
- JSON observer events showing save/load/resume
- Execution time: ~9.5 seconds (5.9s initial + 3.6s resume)
- Time saved: ~2-3s (skipped completed stages)

**[Full Documentation →](./phase-06-checkpointing/README.md)**

---

### Phase 7: Conditional Routing
**Directory:** [`phase-07-conditional-routing/`](./phase-07-conditional-routing/)

Conditional routing with state management, pattern composition, and revision loops.

**Scenario:** Technical document review workflow with sequential analysis, concurrent review, and conditional approval routing

**Run:**
```bash
go run examples/phase-07-conditional-routing/main.go
```

**Demonstrates:**
- ChainNode integration (sequential analysis by 3 specialists)
- ParallelNode integration (concurrent review by 3 reviewers)
- ConditionalNode routing (approve/revise/reject based on consensus)
- State accumulation across multiple iterations
- Revision loops with termination logic (max 2 revisions)
- Pattern composition within state graphs
- Workflow cycles with cycle detection
- All three integration helpers in one workflow

**Expected Output:**
- 6 agents (3 analysts + 3 reviewers) processing document
- Sequential analysis by technical, security, and business specialists
- Concurrent reviews with consensus calculation (66% threshold)
- Conditional routing based on review consensus
- Revision loop (typically 2 revisions before rejection)
- 9-10 graph iterations with cycle detection events
- Final decision: approved, rejected, or max revisions reached
- JSON observer events from all layers (graph, patterns, agents)
- Execution time: ~40-50 seconds (3 analysis cycles + 3 review cycles)
- Document version incrementing with each revision

**[Full Documentation →](./phase-07-conditional-routing/README.md)**

---

## Execution Patterns

### Basic Execution

Each example can be run directly:

```bash
# From repository root
go run examples/<phase-name>/main.go
```

### Building Examples

To build executables:

```bash
# Build all examples
for example in phase-01-hubs phase-02-03-state-graphs phase-04-sequential-chains phase-05-parallel-execution phase-06-checkpointing phase-07-conditional-routing; do
    go build -o bin/$example examples/$example/main.go
done

# Run built executable
./bin/phase-01-hubs
```

### Filtering Observer Output

All examples use slog observer producing JSON events mixed with human-readable output. Filter events:

```bash
# Show only state events
go run examples/phase-02-03-state-graphs/main.go 2>&1 | grep '"type":"state\.'

# Show only worker events
go run examples/phase-05-parallel-execution/main.go 2>&1 | grep '"type":"worker\.'

# Pretty-print JSON events with jq
go run examples/phase-04-sequential-chains/main.go 2>&1 | jq 'select(.type)'

# Show only human-readable output (filter out JSON)
go run examples/phase-01-hubs/main.go 2>&1 | grep -v '{"time"'
```

## Understanding Example Output

### Console Output Structure

All examples follow this pattern:

```
=== Example Title ===

1. Setup step...
  ✓ Progress indicator

2. Next step...
  ✓ Progress indicator

[Execution with progress updates]

N. Results/Metrics
   Formatted output

=== Example Complete ===
```

### Observer Events

JSON events are logged to stdout alongside human-readable output:

```json
{
  "time": "2025-11-07T15:59:43.599974744-05:00",
  "level": "INFO",
  "msg": "Event",
  "type": "node.start",
  "source": "deployment-pipeline",
  "timestamp": "2025-11-07T15:59:43.59997226-05:00",
  "data": {
    "iteration": 1,
    "node": "plan"
  }
}
```

**Key Event Types by Phase:**
- **Phase 1:** Hub metrics only (no observer integration)
- **Phase 2+3:** `graph.start`, `node.start`, `node.complete`, `edge.evaluate`, `cycle.detected`, `graph.complete`
- **Phase 4:** `chain.start`, `step.start`, `step.complete`, `chain.complete`
- **Phase 5:** `parallel.start`, `worker.start`, `worker.complete`, `parallel.complete`
- **Phase 6:** `checkpoint.save`, `checkpoint.load`, `checkpoint.resume` (plus all Phase 2+3 graph events)
- **Phase 7:** `route.evaluate`, `route.select`, `route.execute` (plus all Phase 2+3 graph and Phase 4-5 pattern events)

All phases emit: `state.create`, `state.clone`, `state.set` for state operations

## Performance Characteristics

### Execution Times (Approximate)

| Example | Duration | Notes |
|---------|----------|-------|
| Phase 1 Hubs | 15-20s | Multiple round-trip agent conversations |
| Phase 2+3 State Graphs | 7-10s | 6-10 node executions depending on test outcomes |
| Phase 4 Sequential Chains | 5-6s | 5 sequential agent calls |
| Phase 5 Parallel Execution | 2.5s | 12 concurrent agent calls with 4 workers |
| Phase 6 Checkpointing | 9.5s | 4 stages total: 5.9s initial (fails) + 3.6s resume |
| Phase 7 Conditional Routing | 40-50s | 9 sequential + 9 concurrent calls across 3 workflow cycles |

**Note:** First execution may be slower due to model loading. Subsequent runs are faster with cached models.

### GPU Performance Impact

**With GPU (NVIDIA):**
- Model load: 1-2s (first request)
- Inference: 100-150ms per call

**Without GPU (CPU only):**
- Model load: 3-5s (first request)
- Inference: 800-1500ms per call

**Speedup:** GPU provides ~6-8x performance improvement

## Configuration

### Agent Configuration

All examples use `config.llama.json` (or `config.gemma.json` for phase-01):

```json
{
  "name": "agent-name",
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

### Observer Configuration

Examples use slog observer for JSON event logging:

```go
// All examples configure observer similarly
slogHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
})
slogObserver := observability.NewSlogObserver(slog.New(slogHandler))
observability.RegisterObserver("slog", slogObserver)
```

Change to `noop` observer for production (zero overhead):

```go
config.Observer = "noop"  // No observability
```

## Common Issues

### Ollama Not Running

**Symptom:** "connection refused" errors

**Solution:**
```bash
docker ps | grep ollama  # Check if running
docker-compose up -d     # Start if needed
```

### Model Not Found

**Symptom:** "model not found" errors

**Solution:**
```bash
docker exec tau-orchestrate-ollama ollama list  # Check available models
docker exec tau-orchestrate-ollama ollama pull llama3.2:3b  # Pull if missing
```

### Slow Performance

**Symptom:** >1 second per agent call

**Possible Causes:**
- CPU-only inference (no GPU)
- Model loading on first request
- Network latency

**Solutions:**
- Use GPU if available
- Expect first few requests to be slower
- Reduce `max_tokens` in config for faster responses

### No JSON Output

**Symptom:** Missing observer events

**Cause:** Phase 1 doesn't use observer pattern (Phase 1 was completed before observer integration)

**Solution:** Run Phase 2+ examples to see observer output

## Next Steps

After exploring the examples:

1. **Review Implementation:** Check session summaries in `_context/sessions/` for implementation details
2. **Combine Patterns:** Try composing state graphs with sequential chains or parallel execution
4. **Custom Scenarios:** Modify examples with your own agents and workflows

## Example Dependencies

```
Phase 1 (Foundation)
    ↓
Phase 2+3 (State Graphs) ──┐
    ↓                      │
Phase 4 (Sequential) ──────┤
    ↓                      ├─→ Can be combined
Phase 5 (Parallel) ────────┤
    ↓                      │
Phase 6 (Checkpointing) ───┤
    ↓                      │
Phase 7 (Conditional) ─────┘
```

**Independence:** Each phase can be explored independently, but understanding earlier phases helps with later concepts.

**Phase 6 Note:** Builds directly on Phase 2+3 state graphs. Understanding state graph execution is recommended before exploring checkpointing.

**Phase 7 Note:** Combines all pattern types (Phases 4-5) within state graphs (Phases 2-3). Understanding both workflow patterns and state graphs is recommended before exploring conditional routing.

## Getting Help

- **Example-specific issues:** See individual README.md in each example directory
- **Bug reports:** Open issue at https://github.com/tailored-agentic-units/tau-orchestrate/issues

## License

All examples are part of the tau-orchestrate project and follow the same license as the main repository.
