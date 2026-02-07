# Phase 2+3: State Graphs Example - Software Deployment Pipeline

This example demonstrates state graph execution with conditional routing, cycle detection, and observability using a realistic software deployment pipeline scenario.

## Overview

The example simulates an automated deployment pipeline where a deployment-manager agent evaluates each stage:
- **Plan** → **Build** → **Test** → **Deploy** (success path)
- **Test** → **Fix** → **Test** (retry cycle, up to 3 attempts)
- **Test** → **Rollback** (failure path after max retries)

The pipeline demonstrates:
1. **Linear workflow execution** - Sequential progression through nodes
2. **Conditional routing** - Test results determine next transition
3. **Cycle detection** - Retry loops are detected and logged
4. **Multiple exit points** - Success (deploy) or failure (rollback) termination
5. **State accumulation** - Deployment metadata flows through the pipeline
6. **Observer integration** - Complete execution trace via slog observer

## Architecture

### State Graph Structure

```
┌─────────┐
│  PLAN   │  Entry Point
└────┬────┘
     │ always
     ▼
┌─────────┐
│  BUILD  │
└────┬────┘
     │ always
     ▼
┌─────────┐                  ┌──────────┐
│  TEST   │─────────────────→│  DEPLOY  │  Exit Point (success)
└────┬────┘   tests passed   └──────────┘
     │
     │ tests failed
     │ (retries < 3)
     ▼
┌─────────┐
│   FIX   │
└────┬────┘
     │ always
     └──────────┐
                │ (cycle: test → fix → test)
                ▼
           ┌─────────┐
           │  TEST   │ (revisited)
           └────┬────┘
                │
                │ max retries
                │ exceeded (3)
                ▼
           ┌──────────┐
           │ ROLLBACK │  Exit Point (failure)
           └──────────┘
```

### Node Descriptions

| Node | Purpose | State Updates | Agent Query |
|------|---------|---------------|-------------|
| `plan` | Analyze deployment requirements | `plan`, `status="planned"` | "Analyze deployment plan for {app} to {env}" |
| `build` | Create deployment artifacts | `artifacts`, `status="built"` | "What artifacts should be built for {app}?" |
| `test` | Run automated test suite | `test_result`, `status="tested"` | "Evaluate test results (attempt N)" |
| `fix` | Address test failures | `fix_details`, `retry_count++`, `status="fixed"` | "Test failed: {result}. What fix to apply?" |
| `deploy` | Deploy to target environment | `deployment_result`, `status="deployed"` | "Confirm deployment to {env} with {artifacts}" |
| `rollback` | Revert after max retries | `rollback_details`, `status="rolled_back"` | "Deployment failed after N attempts. Describe rollback" |

### Edge Predicates

| From | To | Condition |
|------|-----|-----------|
| plan | build | Always |
| build | test | Always |
| test | deploy | Tests passed (response starts with 'y', 'Y', 'p', or 'P') |
| test | fix | Tests failed AND retry_count < 3 |
| test | rollback | retry_count >= 3 |
| fix | test | Always (creates cycle) |

### State Flow

The state accumulates deployment metadata as it flows through the pipeline:

**Initial State:**
```go
{
  "app_name": "cloud-api-service",
  "target_env": "production",
  "retry_count": 0
}
```

**Final State (Deployment Success):**
```go
{
  "app_name": "cloud-api-service",
  "target_env": "production",
  "retry_count": 0,
  "plan": "...",
  "artifacts": "...",
  "test_result": "...",
  "deployment_result": "...",
  "status": "deployed"
}
```

**Final State (Rollback):**
```go
{
  "app_name": "cloud-api-service",
  "target_env": "production",
  "retry_count": 3,
  "plan": "...",
  "artifacts": "...",
  "test_result": "...",
  "fix_details": "...",
  "rollback_details": "...",
  "status": "rolled_back"
}
```

## Key Concepts Demonstrated

### 1. State Graph Construction

```go
// Create graph with observer configuration
graphConfig := config.DefaultGraphConfig("deployment-pipeline")
graphConfig.Observer = "slog"
graphConfig.MaxIterations = 10

graph, err := state.NewGraph(graphConfig)

// Add nodes
graph.AddNode("plan", planNode)
graph.AddNode("build", buildNode)
// ...

// Add edges with predicates
graph.AddEdge("test", "deploy", testsPassed)
graph.AddEdge("test", "fix", testsFailedWithRetriesLeft)

// Configure entry and exit points
graph.SetEntryPoint("plan")
graph.SetExitPoint("deploy")
graph.SetExitPoint("rollback")
```

### 2. Function Nodes

Nodes are created using `state.NewFunctionNode` which wraps a function:

```go
planNode := state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
    appName, _ := s.Get("app_name")

    // Call agent for processing
    response, err := deploymentAgent.Chat(ctx, prompt)
    if err != nil {
        return s, fmt.Errorf("plan failed: %w", err)
    }

    // Update state immutably
    return s.Set("plan", response.Content()).Set("status", "planned"), nil
})
```

### 3. Conditional Edge Predicates

Predicates are functions that evaluate state to determine transitions:

```go
testsPassed := func(s state.State) bool {
    result, exists := s.Get("test_result")
    if !exists {
        return false
    }
    testStr := fmt.Sprintf("%v", result)
    return len(testStr) > 0 && (testStr[0] == 'y' || testStr[0] == 'Y')
}

graph.AddEdge("test", "deploy", testsPassed)
```

### 4. Cycle Detection

The graph automatically detects when a node is revisited:

```json
{
  "type": "cycle.detected",
  "source": "deployment-pipeline",
  "data": {
    "node": "test",
    "visit_count": 2,
    "iteration": 5,
    "path_length": 5
  }
}
```

### 5. Observer Events

The slog observer emits JSON events for every state and graph operation:

**State Operations:**
- `state.create` - New state created
- `state.clone` - State cloned for immutability
- `state.set` - Key set in state

**Graph Operations:**
- `graph.start` - Execution begins
- `node.start` - Node begins processing
- `node.complete` - Node finishes (with error flag)
- `edge.evaluate` - Predicate evaluation
- `edge.transition` - Transition selected
- `cycle.detected` - Node revisited
- `graph.complete` - Execution finishes

## Prerequisites

### 1. Ollama with Model

The example requires [Ollama](https://ollama.ai) running with:
- `llama3.2:3b` - For deployment-manager agent

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
go run examples/phase-02-03-state-graphs/main.go
```

### Option 2: Build and run

```bash
# Build
go build -o bin/phase-02-03-state-graphs examples/phase-02-03-state-graphs/main.go

# Run
./bin/phase-02-03-state-graphs
```

## Expected Output

The example produces two types of output:

1. **Human-readable progress** showing node execution
2. **JSON observer events** showing complete execution trace

```
=== Software Deployment Pipeline - State Graph Example ===

1. Configuring observability...
  ✓ Registered slog observer

2. Loading agent configuration...
  ✓ Created deployment-manager agent (llama3.2:3b)

3. Creating deployment pipeline state graph...
  ✓ Created state graph with observer

4. Defining pipeline nodes...
  ✓ Added 6 nodes (plan, build, test, fix, deploy, rollback)

5. Defining pipeline transitions...
  ✓ Added 6 edges with conditional routing

6. Configuring entry and exit points...
  ✓ Entry point: plan
  ✓ Exit points: deploy, rollback

7. Executing deployment pipeline...

  Initial deployment request:
    Application: cloud-api-service
    Environment: production

{"time":"...","level":"INFO","msg":"Event","type":"graph.start","source":"deployment-pipeline",...}

  → PLAN: Analyzing deployment requirements...
     Plan: For 'cloud-api-service', a canary deployment to production would be recommended...

  → BUILD: Compiling and creating artifacts...
     Artifacts: Docker image, Configuration package, Component manifests...

  → TEST: Running automated test suite...
     Test Result: Tests indicate 87% pass rate, failures from unit tests...

  → FIX: Addressing test failures...
     Fix Applied: Re-run tests with increased timeout values...

{"time":"...","level":"INFO","msg":"Event","type":"cycle.detected","data":{"node":"test","visit_count":2,...}}

  → TEST: Running automated test suite...
     Test Result: 100% pass rate, all tests passed...

  → DEPLOY: Deploying to target environment...
     Deployment: Deployment confirmed to production...

  ✓ Pipeline execution completed

8. Deployment Results

   Final Status: deployed

   ✓ DEPLOYMENT SUCCESSFUL
   Details: Deployment confirmed to production...

9. Execution Metrics
   Duration: 7.5s
   Max Iterations Allowed: 10

=== Deployment Pipeline Complete ===
```

## Execution Outcomes

Due to the probabilistic nature of LLM responses, the pipeline may follow different paths:

### Success Path (Deploy)

When tests pass quickly:
```
plan → build → test → deploy (4 iterations)
```

### Retry Path (Fix and Deploy)

When tests fail initially but pass after fixes:
```
plan → build → test → fix → test → deploy (6 iterations)
plan → build → test → fix → test → fix → test → deploy (8 iterations)
```

### Failure Path (Rollback)

When tests continue failing after 3 retry attempts:
```
plan → build → test → fix → test → fix → test → fix → test → rollback (10 iterations)
```

The example demonstrates cycle detection in all retry scenarios, with `cycle.detected` events emitted when `test` or `fix` nodes are revisited.

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

### Graph Configuration

The graph is configured with:

```go
graphConfig := config.DefaultGraphConfig("deployment-pipeline")
graphConfig.Observer = "slog"           // Use slog observer for JSON events
graphConfig.MaxIterations = 10          // Prevent infinite loops
```

### Customization

**Adjust retry limit:**

Modify the predicate logic in `main.go`:

```go
testsFailedWithRetriesLeft := func(s state.State) bool {
    retryCount, _ := s.Get("retry_count")
    return retryCount.(int) < 5  // Increase from 3 to 5 retries
}

maxRetriesExceeded := func(s state.State) bool {
    retryCount, _ := s.Get("retry_count")
    return retryCount.(int) >= 5  // Update threshold
}
```

**Adjust max iterations:**

```go
graphConfig.MaxIterations = 20  // Increase for more complex workflows
```

**Change application/environment:**

```go
initialState = initialState.Set("app_name", "web-service")
initialState = initialState.Set("target_env", "staging")
```

## Observer Output Analysis

### Understanding JSON Events

Each JSON event follows this structure:

```json
{
  "time": "2025-11-07T15:59:43.600034317-05:00",
  "level": "INFO",
  "msg": "Event",
  "type": "graph.start",
  "source": "deployment-pipeline",
  "timestamp": "2025-11-07T15:59:43.600032984-05:00",
  "data": {
    "entry_point": "plan",
    "exit_points": 2
  }
}
```

**Key Fields:**
- `type` - Event type (state.*, node.*, edge.*, graph.*, cycle.*)
- `source` - Event origin (graph name or "state")
- `data` - Event-specific metadata

### Filtering Events

To focus on specific event types:

```bash
# Show only node execution
go run examples/phase-02-03-state-graphs/main.go 2>&1 | grep '"type":"node\.'

# Show cycle detection
go run examples/phase-02-03-state-graphs/main.go 2>&1 | grep '"type":"cycle.detected"'

# Show edge transitions
go run examples/phase-02-03-state-graphs/main.go 2>&1 | grep '"type":"edge.transition"'
```

### Execution Path Reconstruction

Follow the execution path by tracking `node.start` and `edge.transition` events:

1. `graph.start` with entry_point
2. `node.start` for each node execution
3. `edge.transition` showing path taken
4. `cycle.detected` when nodes repeat
5. `graph.complete` with exit_point and iterations

## Key Code Patterns

### Immutable State Updates

State updates return new state instances:

```go
// Immutable chaining
newState := s.Set("key1", value1).Set("key2", value2)

// Original state unchanged
oldValue, _ := s.Get("key1")  // Returns original value
newValue, _ := newState.Get("key1")  // Returns updated value
```

### Error Handling in Nodes

Nodes return errors which stop execution:

```go
planNode := state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
    response, err := agent.Chat(ctx, prompt)
    if err != nil {
        return s, fmt.Errorf("plan failed: %w", err)
    }
    return s.Set("plan", response.Content()), nil
})
```

### Predicate Composition

Predicates can be composed using And, Or, Not:

```go
import "github.com/tailored-agentic-units/tau-orchestrate/pkg/state"

complexPredicate := state.And(
    state.KeyExists("test_result"),
    state.Not(testsPassed),
)

graph.AddEdge("test", "fix", complexPredicate)
```

## What's Next

This example demonstrates Phase 2+3 capabilities (State Management and Graph Execution). Future phases will add:

- **Phase 4**: Sequential chains pattern for iterative processing
- **Phase 5**: Parallel execution pattern for concurrent workflows
- **Phase 6**: Checkpointing for resumable workflows
- **Phase 7**: Advanced routing patterns

### Related Examples

- **Phase 1 Hubs** (`examples/phase-01-hubs/`) - Hub and messaging primitives
- **Phase 4 Sequential Chains** - Coming soon
- **Phase 5 Parallel Execution** - Coming soon

## Troubleshooting

**Agent not responding:**
- Verify Ollama is running: `curl http://localhost:11434/api/tags`
- Check model is available: `ollama list | grep llama3.2`
- Ensure config file points to correct Ollama URL

**Pipeline always rolls back:**
- This is expected behavior sometimes due to LLM response variability
- The agent's responses don't always start with "yes" or "pass"
- Run multiple times to see different paths (success, partial retry, rollback)
- Demonstrates realistic deployment challenges and retry logic

**No JSON output visible:**
- JSON events go to stdout along with human-readable output
- Use `2>&1` to capture both streams
- Filter with `jq` for structured viewing: `go run main.go 2>&1 | jq 'select(.type)'`

**Timeout errors:**
- Increase agent timeout in config
- Reduce max_tokens for faster responses
- Use GPU-accelerated Ollama for better performance

**Connection refused:**
- Verify Ollama container is running: `docker ps`
- Check port 11434 is accessible
- Restart Ollama: `docker-compose restart`
