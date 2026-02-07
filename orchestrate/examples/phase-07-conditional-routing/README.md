# Phase 7: Conditional Routing - Technical Document Review Workflow

This example demonstrates Phase 7 capabilities: conditional routing with state management, pattern composition, and revision loops.

## Scenario

A technical document review workflow with sequential analysis, concurrent reviews, and conditional approval routing.

**Document**: API Authentication System Design specification
**Workflow**: Analyze → Review → Decide (with revision loop)
**Agents**: 6 agents (3 analysts + 3 reviewers)
**Outcome**: Approval, revision request, or rejection based on consensus

## Architecture

### Stateful Workflow Graph

```
┌─────────┐
│ analyze │ (ChainNode - Sequential)
└────┬────┘
     │
     ↓
┌─────────┐
│ review  │ (ParallelNode - Concurrent)
└────┬────┘
     │ [if consensus calculated]
     ↓
┌──────────┐
│ decision │ (ConditionalNode - Routing)
└─────┬────┘
      │
      ├─→ [approve] → finalize (exit)
      ├─→ [revise]  → analyze (loop)
      └─→ [reject]  → finalize (exit)
```

### Pattern Integration

**ChainNode** - Sequential Analysis
- Technical analyst → Security analyst → Business analyst
- Each analysis appended to state
- State flows through all 3 analysts sequentially

**ParallelNode** - Concurrent Review
- 3 reviewers process document concurrently
- Results aggregated with consensus calculation (66% threshold)
- Average score computed across all reviewers

**ConditionalNode** - Decision Routing
- **approve** route: Consensus ≥ 66% → workflow complete
- **revise** route: Consensus < 66% and revisions < 2 → loop to analyze
- **reject** route: Max revisions (2) reached → workflow complete

### State Management

**State Keys:**
- `document`: Document struct with ID, title, content, version, status
- `analyses`: Accumulated analyses from all iterations
- `reviews`: Review results from parallel execution
- `consensus`: Boolean indicating if 66% threshold met
- `average_score`: Average review score across reviewers
- `approved_count`: Number of reviewers who approved
- `decision`: Final decision with approval status and reasoning
- `revision_count`: Number of revisions requested
- `workflow_complete`: Boolean controlling loop termination

**State Flow:**
1. Initial state contains document
2. Each analysis iteration appends to analyses array
3. Parallel reviews aggregated into state
4. Decision updates document status and revision count
5. Loop continues until workflow_complete = true

## Agents

### Analysts (Sequential via ChainNode)

1. **technical-analyst** (llama)
   - Analyzes technical accuracy and implementation details
   - Identifies technical errors and unclear explanations

2. **security-analyst** (gemma)
   - Analyzes security implications and best practices
   - Identifies vulnerabilities and missing security warnings

3. **business-analyst** (llama)
   - Analyzes business value and clarity for non-technical readers
   - Identifies unclear justification and missing user perspective

### Reviewers (Concurrent via ParallelNode)

4. **reviewer-alpha** (gemma)
   - Experienced technical reviewer
   - Thorough but fair assessment

5. **reviewer-beta** (llama)
   - Senior reviewer focused on quality
   - Emphasizes completeness and clarity

6. **reviewer-gamma** (gemma)
   - Principal engineer reviewer
   - Focuses on technical depth and accuracy

## Execution Flow

### Happy Path (Approval)

1. **Analyze** (ChainNode): Sequential analysis by 3 specialists
2. **Review** (ParallelNode): Concurrent review by 3 reviewers
3. **Decision** (ConditionalNode): Consensus ≥ 66% → **approve** route
4. **Finalize**: Document status = "approved", workflow complete

### Revision Loop (Example Run)

1. **Analyze** (iteration 1): 3 analyses completed, document v1
2. **Review** (iteration 1): Reviews show insufficient consensus
3. **Decision** (iteration 1): **revise** route selected, revision_count = 1
4. **Analyze** (iteration 2): 3 more analyses, document v2
5. **Review** (iteration 2): Reviews still insufficient
6. **Decision** (iteration 2): **revise** route selected, revision_count = 2
7. **Analyze** (iteration 3): 3 more analyses, document v3
8. **Review** (iteration 3): Reviews still insufficient
9. **Decision** (iteration 3): Max revisions reached → **reject** route
10. **Finalize**: Document status = "rejected", workflow complete

### Observer Events

**Graph Events:**
- `graph.start`, `graph.complete`
- `node.start`, `node.complete` (10 iterations in example)
- `edge.evaluate`, `edge.transition`
- `cycle.detected` (9 cycles in example)
- `checkpoint.save` (after each node)

**Pattern Events:**
- **Chain**: `chain.start`, `step.start`, `step.complete`, `chain.complete`
- **Parallel**: `parallel.start`, `worker.start`, `worker.complete`, `parallel.complete`
- **Conditional**: `route.evaluate`, `route.select`, `route.execute`

## Key Features Demonstrated

### 1. Pattern Composition
All three integration helpers (ChainNode, ParallelNode, ConditionalNode) used as state graph nodes.

### 2. State Accumulation
- Analyses accumulated across multiple iterations
- Reviews aggregated from parallel execution
- Decision state updated with each routing

### 3. Conditional Routing
Three distinct routes with different outcomes based on state evaluation.

### 4. Revision Loop
Workflow loops back to analyze node until approval or max revisions reached.

### 5. Loop Termination
`revision_count` prevents infinite loops (max 2 revisions).

### 6. Checkpointing
State persisted after each node (checkpoint interval = 1).

### 7. Cycle Detection
State graph detects and logs cycles during execution.

### 8. Observer Integration
Complete observability across all layers (graph, patterns, agents).

## Running the Example

### Prerequisites

1. **Ollama running locally** with models:
   ```bash
   ollama run llama3.2:latest
   ollama run gemma2:2b
   ```

2. **Agent configurations** in place:
   - `config.llama.json`
   - `config.gemma.json`

### Execute

```bash
go run examples/phase-07-conditional-routing/main.go
```

### Expected Output

```
=== Technical Document Review Workflow ===
Demonstrating: Chain → Parallel → Conditional routing with state management

1. Loading agent configurations...
   ✓ Agents created: 3 analysts + 3 reviewers

2. Configuring stateful workflow...
   ✓ Graph configured with conditional routing + revision loop

3. Executing stateful workflow...

[Observer events showing workflow execution...]

=== Workflow Complete ===

Document: DOC-2025-001 (v3)
  Title: API Authentication System Design
  Status: rejected

Analyses Completed: 9
  [Technical] technical-analyst
  [Security] security-analyst
  [Business] business-analyst
  [... 6 more from revision iterations ...]

Reviews Completed: 3
  Approved: 0 of 3 (avg score: 50)
  [✗ REJECTED] reviewer-alpha
  [✗ REJECTED] reviewer-beta
  [✗ REJECTED] reviewer-gamma

Final Decision: REJECTED
  Reason: Maximum revisions (2) reached without consensus

Revisions: 2

Workflow Features Demonstrated:
  ✓ ChainNode - Sequential analysis by 3 specialists
  ✓ ParallelNode - Concurrent review by 3 reviewers
  ✓ ConditionalNode - Decision routing (approve/revise/reject)
  ✓ State Management - Document, analyses, reviews, decisions
  ✓ Conditional Edges - Workflow loops based on state
  ✓ Checkpointing - State persisted after each node
```

## Code Highlights

### ChainNode Integration

```go
analyzeNode := workflows.ChainNode(
    chainCfg,
    analysisAgents,
    analyzeProcessor,  // Sequential analysis processor
    nil,               // No progress callback
)
```

The chain processor analyzes the document with each specialist and appends results to state.

### ParallelNode Integration

```go
reviewNode := workflows.ParallelNode(
    parallelCfg,
    reviewAgents,
    reviewProcessor,   // Concurrent review processor
    nil,               // No progress callback
    reviewAggregator,  // Aggregates results to state
)
```

The aggregator calculates consensus (66% approval threshold) and merges review results into state.

### ConditionalNode Integration

```go
decisionNode := workflows.ConditionalNode(
    conditionalCfg,
    decisionPredicate,  // Routes based on consensus
    decisionRoutes,     // approve/revise/reject handlers
)
```

The predicate evaluates consensus and revision count to select the appropriate route.

### Revision Loop Control

```go
// Decision predicate
decisionPredicate := func(s state.State) (string, error) {
    consensus, _ := s.Get("consensus")
    if consensus.(bool) {
        return "approve", nil  // Consensus reached
    }

    revisionCount, _ := s.Get("revision_count")
    if revisionCount == nil || revisionCount.(int) < 2 {
        return "revise", nil   // Request revision
    }

    return "reject", nil       // Max revisions reached
}

// Conditional edge for loop
workflowComplete := func(s state.State) bool {
    complete, _ := s.Get("workflow_complete")
    return complete != nil && complete.(bool)
}

graph.AddEdge("decision", "finalize", workflowComplete)
graph.AddEdge("decision", "analyze", state.Not(workflowComplete))
```

## What Makes This Example Unique

**vs. Phase 1 (Hubs)**: Adds stateful workflow composition and conditional routing
**vs. Phase 4 (Chains)**: Composes chain as node within larger graph
**vs. Phase 5 (Parallel)**: Composes parallel as node with result aggregation
**vs. Phase 6 (Checkpointing)**: Adds conditional routing with state-based loops

This example is the first to demonstrate:
- All three integration helpers in one workflow
- Conditional routing with multiple routes
- Revision loops with termination logic
- State accumulation across multiple iterations
- Pattern composition within state graphs

## Learn More

- **Implementation Guide**: `_context/phase-07-conditional-routing-integration.md`
- **Package Documentation**: `pkg/workflows/doc.go`
- **Architecture Details**: `ARCHITECTURE.md` (Phase 7 section)
