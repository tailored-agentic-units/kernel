# Planned Examples: Phases 2-4 Demonstration

## Overview

This document plans examples demonstrating state graph (phases 2-3) and sequential chain (phase 4) capabilities through progressively complex scenarios. Examples serve dual purposes:
1. **Integration Testing**: Validate API usability, test composition patterns, exercise error paths
2. **Colleague Demonstration**: Show progression from simple to complex, illustrate patterns, provide copy-paste starting points

## Implementation Status Summary

Based on phases 2-4 completion:

**State Graph Capabilities (95.6% coverage)**:
- `State` - Immutable state container with observer integration
- `StateNode` interface - Computation step contract
- `StateGraph` interface - Workflow definition and execution
- `TransitionPredicate` - Conditional routing with helpers (KeyExists, KeyEquals, And, Or, Not)
- `ExecutionError` - Rich error context with node, state, path
- Features: Linear paths, conditional routing, cycle detection, max iterations, multiple exit points, context cancellation

**Sequential Chain Capabilities (97.4% coverage)**:
- `ProcessChain[TItem, TContext]` - Generic fold/reduce pattern
- `StepProcessor[TItem, TContext]` - Item processor with accumulated state
- `ChainResult[TContext]` - Execution result with optional intermediate states
- `ChainError[TItem, TContext]` - Rich error context with step, item, state
- Features: State accumulation, fail-fast, progress callbacks, context cancellation, intermediate capture

**Integration Patterns**:
- State type works naturally as TContext in chains
- Patterns can be used within state graph nodes
- Both support direct tau-core usage (primary) and hub coordination (optional)
- Observer integration throughout both primitives

## Example Directory Structure

```
examples/
├── planned-examples.md          # This document
├── phase-01-hubs/               # Existing - hub coordination (complete)
├── phase-02-03-state-graphs/    # State graph examples
│   ├── 01-linear-workflow/
│   ├── 02-conditional-routing/
│   ├── 03-cyclic-workflow/
│   ├── 04-agent-integration/
│   └── 05-hub-coordination/
├── phase-04-sequential-chains/  # Sequential chain examples
│   ├── 01-document-analysis/
│   ├── 02-conversation-chain/
│   ├── 03-state-context/
│   └── 04-error-handling/
└── phase-02-04-integration/     # Combined pattern examples
    ├── 01-chain-in-graph/
    ├── 02-multi-pattern-workflow/
    └── 03-observability-showcase/
```

## State Graph Examples (Phases 2-3)

### Example 1: Linear Workflow (Basic)

**Directory**: `examples/phase-02-03-state-graphs/01-linear-workflow/`

**Purpose**: Demonstrate simplest state graph usage - foundational example

**Key Concepts Demonstrated**:
- Graph construction API: `NewGraph()`, `AddNode()`, `AddEdge()`, `SetEntryPoint()`, `SetExitPoint()`
- Linear node execution (A → B → C)
- State accumulation across nodes
- `FunctionNode` wrapper for simple functions
- Observer integration (using NoOpObserver)
- Basic error handling

**Scenario**: Document Processing Pipeline
- **Node A: Load** - Simulates loading document, sets `state.Set("content", text)`
- **Node B: Analyze** - Extracts keywords from content, sets `state.Set("keywords", []string)`
- **Node C: Summarize** - Generates summary from content, sets `state.Set("summary", text)`
- **Flow**: Load → Analyze → Summarize (linear, no branching)

**Code Structure**:
```go
// Create graph with config
cfg := config.DefaultGraphConfig()
cfg.Name = "document-pipeline"
graph, _ := state.NewGraph(cfg)

// Define nodes as functions
loadNode := state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
    // Simulate document loading
    return s.Set("content", "Sample document text..."), nil
})

analyzeNode := state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
    content, _ := s.Get("content")
    // Extract keywords
    return s.Set("keywords", []string{"key1", "key2"}), nil
})

summarizeNode := state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
    content, _ := s.Get("content")
    // Generate summary
    return s.Set("summary", "Brief summary..."), nil
})

// Build graph
graph.AddNode("load", loadNode)
graph.AddNode("analyze", analyzeNode)
graph.AddNode("summarize", summarizeNode)

graph.AddEdge("load", "analyze", nil)
graph.AddEdge("analyze", "summarize", nil)

graph.SetEntryPoint("load")
graph.SetExitPoint("summarize")

// Execute
initialState := state.New(observability.NoOpObserver{})
finalState, err := graph.Execute(ctx, initialState)

// Display results
content, _ := finalState.Get("content")
keywords, _ := finalState.Get("keywords")
summary, _ := finalState.Get("summary")
```

**Console Output Example**:
```
Document Processing Pipeline
============================
1. Loading document...
   Content: "Sample document text..." (50 chars)

2. Analyzing content...
   Keywords: [key1, key2, key3]

3. Generating summary...
   Summary: "Brief summary of the document..."

Pipeline Complete
Final state contains: content, keywords, summary
```

**Integration Value**:
- Validates basic graph construction and execution
- Tests linear path traversal
- Confirms state accumulation works correctly
- Provides simplest possible example for new users

---

### Example 2: Conditional Routing (Intermediate)

**Directory**: `examples/phase-02-03-state-graphs/02-conditional-routing/`

**Purpose**: Demonstrate predicate-based branching and multiple execution paths

**Key Concepts Demonstrated**:
- Multiple edges from single node
- `TransitionPredicate` evaluation order (first match wins)
- Predicate helper functions: `KeyExists()`, `KeyEquals()`, `And()`, `Or()`, `Not()`
- Multiple exit points (different terminal nodes)
- Different final states based on path taken
- `ExecutionError` with path tracking

**Scenario**: Document Approval Workflow
- **Node: Validate** - Checks document quality, sets "status" field
- **Branch 1**: If status == "valid" → Approve node → "approved" exit
- **Branch 2**: If status == "invalid" → Reject node → "rejected" exit
- **Branch 3**: Otherwise → Revision node → loops back to Validate
- **Flow**: Validate → (conditional) → Approve|Reject|Revise

**Code Structure**:
```go
// Validation node sets status based on quality check
validateNode := state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
    quality := checkQuality(s)
    if quality > 0.8 {
        return s.Set("status", "valid"), nil
    } else if quality < 0.3 {
        return s.Set("status", "invalid"), nil
    }
    return s.Set("status", "needs_revision"), nil
})

// Multiple edges with predicates (evaluated in order)
graph.AddEdge("validate", "approve", state.KeyEquals("status", "valid"))
graph.AddEdge("validate", "reject", state.KeyEquals("status", "invalid"))
graph.AddEdge("validate", "revise", state.KeyEquals("status", "needs_revision"))
graph.AddEdge("revise", "validate", nil) // Loop back after revision

// Multiple exit points
graph.SetExitPoint("approve")
graph.SetExitPoint("reject")
```

**Console Output Example**:
```
Document Approval Workflow
===========================
Iteration 1: Validate
   Quality Score: 0.65
   Status: needs_revision

Iteration 2: Revise
   Applying revisions...

Iteration 3: Validate
   Quality Score: 0.85
   Status: valid

Iteration 4: Approve
   Document approved!

Workflow Complete
Exit Point: approve
Path: validate → revise → validate → approve
```

**Integration Value**:
- Validates conditional routing logic
- Tests predicate evaluation and ordering
- Confirms multiple exit points work correctly
- Demonstrates cycle prevention with exit conditions
- Shows path tracking in execution

---

### Example 3: Cyclic Workflow (Advanced)

**Directory**: `examples/phase-02-03-state-graphs/03-cyclic-workflow/`

**Purpose**: Demonstrate intentional loops with exit conditions and iteration management

**Key Concepts Demonstrated**:
- Deliberate cycle creation (node pointing back to earlier node)
- Cycle detection events (EventCycleDetected)
- Iteration counting and max iterations protection
- Loop exit via predicate evaluation
- Visit count tracking per node
- Full execution path through multiple cycles

**Scenario**: Iterative Content Refinement
- **Node: Generate** - Creates draft content, increments attempt counter
- **Node: QualityCheck** - Evaluates content, sets "quality_score"
- **Predicate**: If score < threshold (0.8) → loop back to Generate
- **Predicate**: If score >= threshold → exit to Finalize
- **Safety**: Max iterations prevents infinite loops
- **Flow**: Generate → QualityCheck → (conditional) → Generate (loop) or Finalize (exit)

**Code Structure**:
```go
cfg := config.DefaultGraphConfig()
cfg.MaxIterations = 10 // Prevent runaway loops

generateNode := state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
    attempt, _ := s.Get("attempt")
    attemptNum := attempt.(int) + 1

    // Generate content (improves with each attempt)
    content := generateContent(attemptNum)

    return s.Set("content", content).Set("attempt", attemptNum), nil
})

qualityCheckNode := state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
    content, _ := s.Get("content")
    score := evaluateQuality(content)

    return s.Set("quality_score", score), nil
})

// Cycle edge: loop back if quality insufficient
graph.AddEdge("quality_check", "generate", func(s state.State) bool {
    score, _ := s.Get("quality_score")
    return score.(float64) < 0.8 // Continue looping
})

// Exit edge: finalize if quality sufficient
graph.AddEdge("quality_check", "finalize", func(s state.State) bool {
    score, _ := s.Get("quality_score")
    return score.(float64) >= 0.8 // Exit loop
})
```

**Console Output Example**:
```
Iterative Content Refinement
=============================
Iteration 1: Generate (attempt 1)
   Content generated: "Initial draft..."

Iteration 2: QualityCheck
   Quality Score: 0.62 (below threshold 0.8)
   → Cycling back to Generate

Iteration 3: Generate (attempt 2) [CYCLE DETECTED]
   Content generated: "Improved draft..."

Iteration 4: QualityCheck
   Quality Score: 0.75 (below threshold 0.8)
   → Cycling back to Generate

Iteration 5: Generate (attempt 3) [CYCLE DETECTED]
   Content generated: "Refined draft..."

Iteration 6: QualityCheck
   Quality Score: 0.87 (meets threshold!)
   → Proceeding to Finalize

Iteration 7: Finalize
   Content finalized!

Workflow Complete
Total Iterations: 7
Cycles Detected: 2 (at node "generate")
Final Quality Score: 0.87
```

**Integration Value**:
- Validates cycle detection works correctly
- Tests max iterations enforcement
- Confirms iteration counter increments properly
- Demonstrates cycle exit via predicates
- Shows practical use case for intentional loops

---

### Example 4: Direct Agent Integration (Practical)

**Directory**: `examples/phase-02-03-state-graphs/04-agent-integration/`

**Purpose**: Show primary usage pattern with tau-core (no hub required)

**Key Concepts Demonstrated**:
- StateNode calling `agent.Chat()` directly (primary pattern)
- StateNode calling `agent.Vision()` with image data
- No hub infrastructure needed (simplest integration)
- Single agent per node pattern
- Real LLM interaction (with configurable mock mode for CI)
- Error handling from agent calls

**Scenario**: Multi-Modal Document Analysis
- **Node: ExtractText** - Uses agent.Vision() to extract text from document image
- **Node: AnalyzeContent** - Uses agent.Chat() to analyze extracted text
- **Node: GenerateReport** - Uses agent.Chat() to create summary report
- **Flow**: ExtractText → AnalyzeContent → GenerateReport (each node uses agent directly)

**Code Structure**:
```go
// Create tau-core agent (from config)
agentCfg, _ := agentconfig.LoadAgentConfig("agent-config.json")
llmAgent, _ := agent.New(agentCfg)

// Node 1: Vision-based text extraction
extractNode := state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
    imageData, _ := s.Get("image_data").([]byte)
    encoded := encoding.EncodeImageDataURI(imageData, "image/png")

    // Direct agent.Vision call (no hub)
    response, err := llmAgent.Vision(ctx, "Extract all text from this document image", []string{encoded})
    if err != nil {
        return s, fmt.Errorf("vision extraction failed: %w", err)
    }

    return s.Set("extracted_text", response.Content()), nil
})

// Node 2: Text analysis
analyzeNode := state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
    text, _ := s.Get("extracted_text").(string)

    // Direct agent.Chat call (no hub)
    response, err := llmAgent.Chat(ctx, fmt.Sprintf("Analyze this text and identify key themes: %s", text))
    if err != nil {
        return s, fmt.Errorf("analysis failed: %w", err)
    }

    return s.Set("analysis", response.Content()), nil
})

// Node 3: Report generation
reportNode := state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
    analysis, _ := s.Get("analysis").(string)

    // Direct agent.Chat call (no hub)
    response, err := llmAgent.Chat(ctx, fmt.Sprintf("Generate a summary report: %s", analysis))
    if err != nil {
        return s, fmt.Errorf("report generation failed: %w", err)
    }

    return s.Set("report", response.Content()), nil
})
```

**Console Output Example**:
```
Multi-Modal Document Analysis
==============================
1. Extracting text from document image...
   → Vision API call complete
   Extracted: 450 characters

2. Analyzing content...
   → Chat API call complete
   Key themes identified: 3

3. Generating report...
   → Chat API call complete
   Report: 850 characters

Analysis Complete
==================
Extracted Text: "The document discusses..."
Analysis: "Key themes include..."
Report: "Summary: This document..."

Total LLM Calls: 3 (1 vision, 2 chat)
```

**Integration Value**:
- Validates tau-core integration (primary pattern)
- Tests both Chat and Vision capabilities in graph nodes
- Confirms error propagation from agent calls
- Demonstrates simplest pattern (no hub complexity)
- Provides practical starting point for new users

**Mock Mode**: Include flag to use mock agent for CI/testing without API keys

---

### Example 5: Hub Coordination (Optional Advanced)

**Directory**: `examples/phase-02-03-state-graphs/05-hub-coordination/`

**Purpose**: Show optional multi-agent orchestration via hub messaging

**Key Concepts Demonstrated**:
- StateNode using `hub.Request()` for agent routing
- Multiple agents coordinated through hub infrastructure
- Hub registration and message handlers
- When hub coordination provides value over direct calls
- Agent discovery and dynamic routing

**Scenario**: Multi-Reviewer Document Workflow
- **Setup**: 3 reviewer agents registered with hub (technical, business, legal)
- **Node: Submit** - Broadcasts document to all reviewers via hub
- **Node: CollectReviews** - Requests review from each reviewer sequentially
- **Node: AggregateReviews** - Combines reviews from multiple agents
- **Flow**: Submit → CollectReviews → AggregateReviews (hub routes messages)

**Code Structure**:
```go
// Create hub and register reviewer agents
hubCfg := config.DefaultHubConfig()
hubCfg.Name = "review-hub"
reviewHub := hub.New(ctx, hubCfg)

// Register reviewer agents with handlers
reviewHub.RegisterAgent(technicalReviewer, technicalHandler)
reviewHub.RegisterAgent(businessReviewer, businessHandler)
reviewHub.RegisterAgent(legalReviewer, legalHandler)

// Node using hub for broadcasting
submitNode := state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
    doc, _ := s.Get("document")

    // Hub broadcast instead of direct agent call
    err := reviewHub.Broadcast(ctx, "coordinator", doc)
    if err != nil {
        return s, fmt.Errorf("broadcast failed: %w", err)
    }

    return s.Set("submitted", true), nil
})

// Node using hub for sequential requests
collectNode := state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
    doc, _ := s.Get("document")
    reviewers := []string{"technical", "business", "legal"}
    reviews := make([]Review, 0, len(reviewers))

    for _, reviewerID := range reviewers {
        // Hub request instead of direct agent call
        response, err := reviewHub.Request(ctx, "coordinator", reviewerID, doc)
        if err != nil {
            return s, fmt.Errorf("request to %s failed: %w", reviewerID, err)
        }
        reviews = append(reviews, response.Data.(Review))
    }

    return s.Set("reviews", reviews), nil
})
```

**Console Output Example**:
```
Multi-Reviewer Document Workflow
=================================
Hub: review-hub
Registered Reviewers: technical, business, legal

1. Submit Document
   → Broadcasting to all reviewers...
   ✓ Broadcast complete (3 agents notified)

2. Collect Reviews
   → Requesting review from technical...
   ✓ Technical review received

   → Requesting review from business...
   ✓ Business review received

   → Requesting review from legal...
   ✓ Legal review received

3. Aggregate Reviews
   → Combining 3 reviews...
   ✓ Aggregation complete

Workflow Complete
==================
Reviews Collected: 3
- Technical: APPROVED (score: 0.9)
- Business: APPROVED (score: 0.85)
- Legal: NEEDS_REVISION (score: 0.6)

Overall Status: NEEDS_REVISION
Hub Messages: 7 (1 broadcast, 3 requests, 3 responses)
```

**Integration Value**:
- Validates hub integration in state graphs (optional pattern)
- Tests broadcast and request patterns in nodes
- Confirms multiple agent coordination
- Demonstrates when hub adds value (multi-agent steps)
- Shows difference from direct agent pattern

---

## Sequential Chain Examples (Phase 4)

### Example 1: Document Analysis Chain (Basic)

**Directory**: `examples/phase-04-sequential-chains/01-document-analysis/`

**Purpose**: Show pattern extracted from classify-docs - foundational chain example

**Key Concepts Demonstrated**:
- `ProcessChain[TItem, TContext]` generic function
- `StepProcessor[TItem, TContext]` implementation
- State accumulation across steps (fold/reduce pattern)
- `ChainResult[TContext]` with final state
- Progress callback reporting
- Observer integration (using NoOpObserver)

**Scenario**: Sequential Page Processing
- **Items**: Document pages ([]Page - simulated with text chunks)
- **Processor**: Analyzes each page, accumulates findings in prompt
- **Context**: string (growing analysis prompt)
- **Result**: Final comprehensive analysis based on all pages

**Code Structure**:
```go
// Define page type
type Page struct {
    Number  int
    Content string
}

// Create pages
pages := []Page{
    {Number: 1, Content: "Introduction text..."},
    {Number: 2, Content: "Main content..."},
    {Number: 3, Content: "Conclusion..."},
}

// Define processor: accumulates analysis in prompt string
processor := func(ctx context.Context, page Page, prompt string) (string, error) {
    // Simulate page analysis
    analysis := fmt.Sprintf("Page %d analysis: %s\n", page.Number, analyzePage(page))

    // Accumulate into prompt
    return prompt + analysis, nil
}

// Progress callback
progress := func(completed, total int, currentPrompt string) {
    fmt.Printf("Progress: %d/%d pages processed\n", completed, total)
    fmt.Printf("Current prompt length: %d chars\n", len(currentPrompt))
}

// Execute chain
cfg := config.DefaultChainConfig()
result, err := workflows.ProcessChain(
    ctx,
    cfg,
    pages,          // TItem = Page
    "",             // TContext = string (initial prompt)
    processor,
    progress,
)

// result.Final contains accumulated analysis
fmt.Printf("Final Analysis:\n%s\n", result.Final)
fmt.Printf("Processed %d steps\n", result.Steps)
```

**Console Output Example**:
```
Document Analysis Chain
=======================
Processing 3 pages...

Progress: 1/3 pages processed
Current prompt length: 45 chars

Progress: 2/3 pages processed
Current prompt length: 98 chars

Progress: 3/3 pages processed
Current prompt length: 156 chars

Chain Complete
==============
Final Analysis:
Page 1 analysis: Introduction establishes context...
Page 2 analysis: Main content covers key topics...
Page 3 analysis: Conclusion summarizes findings...

Processed 3 steps
Final prompt length: 156 characters
```

**Integration Value**:
- Validates basic chain execution
- Tests generic type parameters (TItem=Page, TContext=string)
- Confirms state accumulation works correctly
- Demonstrates pattern from classify-docs
- Provides simplest chain example for new users

---

### Example 2: Conversation Chain (Intermediate)

**Directory**: `examples/phase-04-sequential-chains/02-conversation-chain/`

**Purpose**: Demonstrate agent dialog with context accumulation

**Key Concepts Demonstrated**:
- `ProcessChain` with TItem=string (questions), TContext=Conversation
- Direct `agent.Chat()` calls in processor (primary pattern)
- Conversation context accumulation
- Progress tracking through dialog
- Real LLM interaction with conversation history

**Scenario**: Sequential Q&A with Accumulated Context
- **Items**: List of questions about a topic
- **Processor**: Calls agent.Chat() with question + conversation history
- **Context**: Conversation struct accumulating exchanges
- **Result**: Complete conversation with context-aware responses

**Code Structure**:
```go
// Conversation type for context
type Conversation struct {
    Topic     string
    Exchanges []Exchange
}

type Exchange struct {
    Question string
    Answer   string
}

func (c *Conversation) AddExchange(q, a string) {
    c.Exchanges = append(c.Exchanges, Exchange{Question: q, Answer: a})
}

func (c *Conversation) FormatHistory() string {
    var history strings.Builder
    for _, ex := range c.Exchanges {
        history.WriteString(fmt.Sprintf("Q: %s\nA: %s\n\n", ex.Question, ex.Answer))
    }
    return history.String()
}

// Create agent
agentCfg, _ := agentconfig.LoadAgentConfig("agent-config.json")
llmAgent, _ := agent.New(agentCfg)

// Questions to ask
questions := []string{
    "What is machine learning?",
    "How does it differ from traditional programming?",
    "What are some practical applications?",
}

// Processor with direct agent calls
processor := func(ctx context.Context, question string, conv Conversation) (Conversation, error) {
    // Build prompt with conversation history
    prompt := fmt.Sprintf("%s\n\nQ: %s", conv.FormatHistory(), question)

    // Direct agent.Chat call (no hub)
    response, err := llmAgent.Chat(ctx, prompt)
    if err != nil {
        return conv, fmt.Errorf("chat failed for question %q: %w", question, err)
    }

    // Accumulate conversation
    conv.AddExchange(question, response.Content())
    return conv, nil
}

// Initial conversation
initial := Conversation{Topic: "Machine Learning"}

// Execute chain
result, err := workflows.ProcessChain(ctx, cfg, questions, initial, processor, progressCallback)
```

**Console Output Example**:
```
Conversation Chain
==================
Topic: Machine Learning
Questions: 3

Q1: What is machine learning?
→ Agent call...
A1: Machine learning is a subset of artificial intelligence...
   (Response: 250 chars)

Q2: How does it differ from traditional programming?
→ Agent call with conversation history (1 prior exchange)...
A2: Unlike traditional programming where rules are explicitly coded...
   (Response: 280 chars)

Q3: What are some practical applications?
→ Agent call with conversation history (2 prior exchanges)...
A3: Building on what we discussed, practical applications include...
   (Response: 320 chars)

Conversation Complete
=====================
Total Exchanges: 3
Total Characters: 850
LLM Calls: 3

Full Conversation:
------------------
Q: What is machine learning?
A: Machine learning is a subset...

Q: How does it differ from traditional programming?
A: Unlike traditional programming...

Q: What are some practical applications?
A: Building on what we discussed...
```

**Integration Value**:
- Validates direct agent integration in chains
- Tests conversation context accumulation
- Confirms context history grows correctly
- Demonstrates practical agent usage pattern
- Shows natural progression of dialog

---

### Example 3: State as TContext (Integration)

**Directory**: `examples/phase-04-sequential-chains/03-state-context/`

**Purpose**: Show state.State as TContext for stateful chains - key integration pattern

**Key Concepts Demonstrated**:
- TContext = `state.State` (natural integration)
- State operations (Get, Set, Merge) in processor
- Immutability maintained through chain
- Foundation for stateful workflows
- Integration between state and workflows packages

**Scenario**: Stateful Task Processing Pipeline
- **Items**: List of tasks to process
- **Processor**: Processes task, updates state.State context with results
- **Context**: state.State accumulating task results and metrics
- **Result**: Final state containing all task results

**Code Structure**:
```go
// Task type
type Task struct {
    ID          string
    Type        string
    Description string
}

// Tasks to process
tasks := []Task{
    {ID: "T1", Type: "analysis", Description: "Analyze data set"},
    {ID: "T2", Type: "transform", Description: "Transform results"},
    {ID: "T3", Type: "report", Description: "Generate report"},
}

// Processor using state.State as context
processor := func(ctx context.Context, task Task, s state.State) (state.State, error) {
    // Get current task count
    count, exists := s.Get("task_count")
    if !exists {
        count = 0
    }

    // Process task (simulated)
    result := processTask(task)

    // Update state with results
    s = s.Set(fmt.Sprintf("task_%s_result", task.ID), result)
    s = s.Set("task_count", count.(int)+1)
    s = s.Set("last_task_type", task.Type)

    // Update metrics
    metrics, _ := s.Get("metrics")
    if metrics == nil {
        metrics = make(map[string]int)
    }
    metricsMap := metrics.(map[string]int)
    metricsMap[task.Type]++
    s = s.Set("metrics", metricsMap)

    return s, nil
}

// Initial state
initial := state.New(observability.NoOpObserver{})
initial = initial.Set("task_count", 0)
initial = initial.Set("metrics", make(map[string]int))

// Execute chain with state.State as TContext
result, err := workflows.ProcessChain(
    ctx,
    cfg,
    tasks,     // TItem = Task
    initial,   // TContext = state.State
    processor,
    nil,
)

// Access final state
finalState := result.Final
taskCount, _ := finalState.Get("task_count")
metrics, _ := finalState.Get("metrics")
```

**Console Output Example**:
```
Stateful Task Processing Pipeline
==================================
Processing 3 tasks with state.State context

Step 1: Process task T1 (analysis)
   → State updated: task_T1_result, task_count=1, last_task_type=analysis
   → Metrics: {analysis: 1}

Step 2: Process task T2 (transform)
   → State updated: task_T2_result, task_count=2, last_task_type=transform
   → Metrics: {analysis: 1, transform: 1}

Step 3: Process task T3 (report)
   → State updated: task_T3_result, task_count=3, last_task_type=report
   → Metrics: {analysis: 1, transform: 1, report: 1}

Chain Complete
==============
Final State Contents:
- task_count: 3
- task_T1_result: [analysis result]
- task_T2_result: [transform result]
- task_T3_result: [report result]
- last_task_type: report
- metrics: {analysis: 1, transform: 1, report: 1}

State Keys: 6
Demonstrates: state.State as accumulated context
```

**Integration Value**:
- Validates state.State works as TContext (key integration)
- Tests state operations within chain processor
- Confirms immutability maintained through accumulation
- Demonstrates foundation for stateful workflows
- Shows packages composing naturally

---

### Example 4: Error Handling (Practical)

**Directory**: `examples/phase-04-sequential-chains/04-error-handling/`

**Purpose**: Demonstrate error recovery strategies and debugging with rich error context

**Key Concepts Demonstrated**:
- `ChainError[TItem, TContext]` with rich context
- Error unwrapping to access underlying error
- Intermediate state capture for debugging (CaptureIntermediateStates=true)
- Failed step identification
- Recovery strategies after failure
- Error context inspection

**Scenario**: Processing with Intentional Failure
- **Items**: Mixed valid and invalid inputs
- **Processor**: Validates and processes, returns error on invalid input
- **Config**: CaptureIntermediateStates=true for debugging
- **Result**: ChainError captures step, item, state at failure

**Code Structure**:
```go
// Input type
type Input struct {
    ID    string
    Value int
    Valid bool // Intentionally invalid inputs
}

// Mixed inputs (some invalid)
inputs := []Input{
    {ID: "I1", Value: 100, Valid: true},
    {ID: "I2", Value: 200, Valid: true},
    {ID: "I3", Value: -1, Valid: false}, // Will cause error
    {ID: "I4", Value: 300, Valid: true},
}

// Result type for context
type Result struct {
    ProcessedCount int
    TotalValue     int
    ProcessedIDs   []string
}

// Processor that validates and errors on invalid
processor := func(ctx context.Context, input Input, result Result) (Result, error) {
    if !input.Valid {
        return result, fmt.Errorf("invalid input: value %d is not acceptable", input.Value)
    }

    // Process valid input
    result.ProcessedCount++
    result.TotalValue += input.Value
    result.ProcessedIDs = append(result.ProcessedIDs, input.ID)

    return result, nil
}

// Config with intermediate state capture
cfg := config.ChainConfig{
    CaptureIntermediateStates: true,
    Observer:                  "noop",
}

// Execute chain (will fail on I3)
result, err := workflows.ProcessChain(
    ctx,
    cfg,
    inputs,
    Result{ProcessedIDs: []string{}},
    processor,
    nil,
)

// Inspect error
if err != nil {
    // Type assert to ChainError for rich context
    var chainErr *workflows.ChainError[Input, Result]
    if errors.As(err, &chainErr) {
        fmt.Printf("Chain failed at step %d\n", chainErr.StepIndex)
        fmt.Printf("Failed item: ID=%s, Value=%d\n", chainErr.Item.ID, chainErr.Item.Value)
        fmt.Printf("State at failure: Processed=%d, Total=%d\n",
            chainErr.State.ProcessedCount, chainErr.State.TotalValue)
        fmt.Printf("Underlying error: %v\n", chainErr.Unwrap())

        // Access intermediate states for debugging
        fmt.Printf("\nIntermediate states captured:\n")
        for i, state := range result.Intermediate {
            fmt.Printf("  After step %d: %+v\n", i, state)
        }
    }
}
```

**Console Output Example**:
```
Error Handling Example
======================
Processing 4 inputs with validation...

Step 1: Process I1 (value=100)
   ✓ Valid input processed
   State: {ProcessedCount: 1, TotalValue: 100, IDs: [I1]}

Step 2: Process I2 (value=200)
   ✓ Valid input processed
   State: {ProcessedCount: 2, TotalValue: 300, IDs: [I1, I2]}

Step 3: Process I3 (value=-1)
   ✗ Invalid input detected!
   Error: invalid input: value -1 is not acceptable

Chain Failed
============
ChainError Details:
- Failed at step: 2 (0-indexed)
- Failed item: {ID: I3, Value: -1, Valid: false}
- State at failure: {ProcessedCount: 2, TotalValue: 300, IDs: [I1, I2]}
- Underlying error: "invalid input: value -1 is not acceptable"

Intermediate States (for debugging):
0. Initial: {ProcessedCount: 0, TotalValue: 0, IDs: []}
1. After step 0: {ProcessedCount: 1, TotalValue: 100, IDs: [I1]}
2. After step 1: {ProcessedCount: 2, TotalValue: 300, IDs: [I1, I2]}

Recovery Options:
1. Skip invalid input and continue with remaining
2. Retry with corrected input
3. Apply default value for invalid input
4. Abort processing (current behavior)

Processed before failure: 2/4 inputs
Accumulated value: 300
```

**Integration Value**:
- Validates ChainError rich context
- Tests intermediate state capture for debugging
- Confirms error unwrapping works correctly
- Demonstrates debugging techniques
- Shows practical error recovery strategies

---

## Integration Examples (Phases 2-4 Combined)

### Example 1: Chain Within Graph Node (Composition)

**Directory**: `examples/phase-02-04-integration/01-chain-in-graph/`

**Purpose**: Demonstrate pattern composition - sequential chain as graph node computation

**Key Concepts Demonstrated**:
- Sequential chain as state graph node implementation
- State graph orchestrating multiple chains
- Pattern composition flexibility
- State.State flowing through both primitives
- Nested execution (graph → node → chain → processor)

**Scenario**: Document Workflow with Sub-Pipelines
- **Graph Node "Analyze"**: Runs chain of page analyses
- **Graph Node "Review"**: Runs chain of review steps
- **Graph Node "Finalize"**: Aggregates all results
- **Flow**: Analyze (chain) → Review (chain) → Finalize (simple)

**Code Structure**:
```go
// Node 1: Analysis chain within graph node
analyzeNode := state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
    pages, _ := s.Get("pages").([]Page)

    // Run sequential chain for page analysis
    processor := func(ctx context.Context, page Page, analysis string) (string, error) {
        return analysis + analyzePage(page) + "\n", nil
    }

    result, err := workflows.ProcessChain(ctx, chainCfg, pages, "", processor, nil)
    if err != nil {
        return s, fmt.Errorf("analysis chain failed: %w", err)
    }

    return s.Set("analysis", result.Final), nil
})

// Node 2: Review chain within graph node
reviewNode := state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
    analysis, _ := s.Get("analysis").(string)
    reviewSteps := []string{"grammar", "content", "style"}

    // Run sequential chain for review steps
    processor := func(ctx context.Context, step string, report string) (string, error) {
        return report + reviewStep(step, analysis) + "\n", nil
    }

    result, err := workflows.ProcessChain(ctx, chainCfg, reviewSteps, "", processor, nil)
    if err != nil {
        return s, fmt.Errorf("review chain failed: %w", err)
    }

    return s.Set("review_report", result.Final), nil
})

// Build graph with chain-based nodes
graph.AddNode("analyze", analyzeNode)
graph.AddNode("review", reviewNode)
graph.AddNode("finalize", finalizeNode)

graph.AddEdge("analyze", "review", nil)
graph.AddEdge("review", "finalize", nil)
```

**Console Output Example**:
```
Document Workflow with Sub-Pipelines
=====================================

Graph Node: Analyze
-------------------
  → Running analysis chain (3 pages)...
  Step 1/3: Analyze page 1
  Step 2/3: Analyze page 2
  Step 3/3: Analyze page 3
  ✓ Analysis chain complete (280 chars accumulated)

Graph Node: Review
------------------
  → Running review chain (3 steps)...
  Step 1/3: Grammar review
  Step 2/3: Content review
  Step 3/3: Style review
  ✓ Review chain complete (420 chars accumulated)

Graph Node: Finalize
--------------------
  → Aggregating analysis and review...
  ✓ Final document ready

Workflow Complete
=================
Graph Execution:
- Nodes executed: 3
- Chains executed: 2
- Total chain steps: 6

Pattern Composition:
- State graph orchestrates workflow
- Sequential chains handle sub-pipelines
- State flows through both primitives

Final state keys: pages, analysis, review_report, final_document
```

**Integration Value**:
- Validates chain as graph node pattern
- Tests nested execution (graph → chain)
- Confirms state flows correctly through composition
- Demonstrates practical pattern composition
- Shows building blocks for complex workflows

---

### Example 2: Multi-Pattern Workflow (Complex)

**Directory**: `examples/phase-02-04-integration/02-multi-pattern-workflow/`

**Purpose**: Show realistic complex orchestration with multiple pattern types

**Key Concepts Demonstrated**:
- State graph with different pattern types per node
- Conditional routing between patterns
- State.State as unified context across patterns
- Both direct agent and simulated hub patterns
- Real-world workflow complexity
- Observer integration across all primitives

**Scenario**: Content Review and Approval Workflow
- **Node "Analyze"**: Sequential chain analyzing content sections
- **Node "Review"**: Simulates parallel review (Phase 5 placeholder)
- **Conditional**: KeyEquals("all_approved", true) → Finalize or Revise
- **Node "Finalize"**: Generates final approved content
- **Node "Revise"**: Loops back for revisions
- **Flow**: Complex with chains, conditionals, and cycles

**Code Structure**:
```go
// Node 1: Analysis chain
analyzeNode := state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
    sections, _ := s.Get("sections").([]Section)

    // Sequential chain for section analysis
    processor := func(ctx context.Context, section Section, s state.State) (state.State, error) {
        analysis := analyzeSection(section)
        return s.Set(fmt.Sprintf("section_%s_analysis", section.ID), analysis), nil
    }

    result, err := workflows.ProcessChain(ctx, chainCfg, sections, s, processor, nil)
    if err != nil {
        return s, err
    }

    return result.Final.Set("analysis_complete", true), nil
})

// Node 2: Review (simulates parallel - Phase 5 will use actual parallel pattern)
reviewNode := state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
    // Simulate review from multiple reviewers
    reviews := simulateParallelReviews(s)

    allApproved := true
    for _, review := range reviews {
        if !review.Approved {
            allApproved = false
            break
        }
    }

    return s.Set("reviews", reviews).Set("all_approved", allApproved), nil
})

// Node 3: Finalize
finalizeNode := state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
    // Generate final content
    final := generateFinalContent(s)
    return s.Set("final_content", final).Set("status", "approved"), nil
})

// Node 4: Revise
reviseNode := state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
    // Apply revisions based on reviews
    revised := applyRevisions(s)
    return s.Set("sections", revised).Set("revision_count", getRevisionCount(s)+1), nil
})

// Build graph with conditional routing
graph.AddNode("analyze", analyzeNode)
graph.AddNode("review", reviewNode)
graph.AddNode("finalize", finalizeNode)
graph.AddNode("revise", reviseNode)

graph.AddEdge("analyze", "review", nil)
graph.AddEdge("review", "finalize", state.KeyEquals("all_approved", true))
graph.AddEdge("review", "revise", state.Not(state.KeyEquals("all_approved", true)))
graph.AddEdge("revise", "analyze", nil) // Loop back

graph.SetEntryPoint("analyze")
graph.SetExitPoint("finalize")
```

**Console Output Example**:
```
Content Review and Approval Workflow
=====================================

Iteration 1: Analyze
--------------------
  → Sequential chain: 4 sections
  Step 1/4: Analyze introduction
  Step 2/4: Analyze methodology
  Step 3/4: Analyze results
  Step 4/4: Analyze conclusion
  ✓ All sections analyzed

Iteration 2: Review
-------------------
  → Simulated parallel review (3 reviewers)
  Reviewer 1 (Technical): APPROVED
  Reviewer 2 (Content): NEEDS_REVISION (score: 0.7)
  Reviewer 3 (Style): APPROVED
  → Status: all_approved = false

Iteration 3: Revise
-------------------
  → Applying revisions based on feedback
  Revision count: 1
  ✓ Sections revised

[Cycle detected: returning to Analyze]

Iteration 4: Analyze
--------------------
  → Sequential chain: 4 sections (revised)
  [Re-analyzing sections...]
  ✓ All sections analyzed

Iteration 5: Review
-------------------
  → Simulated parallel review (3 reviewers)
  Reviewer 1 (Technical): APPROVED
  Reviewer 2 (Content): APPROVED (score: 0.9)
  Reviewer 3 (Style): APPROVED
  → Status: all_approved = true

Iteration 6: Finalize
---------------------
  → Generating final approved content
  ✓ Content finalized

Workflow Complete
=================
Graph Execution:
- Total iterations: 6
- Nodes executed: 6 (analyze:2, review:2, revise:1, finalize:1)
- Chains executed: 2 (8 total steps)
- Cycles detected: 1 (at analyze)
- Revisions: 1

Final Status: APPROVED
Final state keys: 12 (sections, analyses, reviews, final_content, status, etc.)

Pattern Composition Demonstrated:
- State graph: Workflow orchestration
- Sequential chains: Section analysis
- Conditional routing: Approval decision
- Cycles: Revision loop
- State.State: Unified context throughout
```

**Integration Value**:
- Validates complex multi-pattern workflows
- Tests all integration points simultaneously
- Confirms state flows correctly through composition
- Demonstrates realistic production scenario
- Shows full power of combined primitives

---

### Example 3: Observability Showcase (Debugging)

**Directory**: `examples/phase-02-04-integration/03-observability-showcase/`

**Purpose**: Demonstrate observer integration for execution debugging and tracing

**Key Concepts Demonstrated**:
- Custom Observer implementation (simple logger)
- All event types captured (graph, node, edge, chain, step, state)
- Event data inspection and formatting
- Execution tracing through complex workflow
- Performance timing via event timestamps
- Foundation for Phase 8 production observability
- NoOpObserver for zero-overhead alternative

**Scenario**: Traced Workflow with Custom Observer
- Implements simple logging Observer that prints events
- Runs multi-pattern workflow (from Example 2)
- Logs all execution events with structured data
- Shows event types, timing, and context
- Demonstrates observability without performance impact

**Code Structure**:
```go
// Simple logging observer
type LoggingObserver struct {
    startTime time.Time
}

func NewLoggingObserver() *LoggingObserver {
    return &LoggingObserver{startTime: time.Now()}
}

func (o *LoggingObserver) OnEvent(ctx context.Context, event observability.Event) {
    elapsed := event.Timestamp.Sub(o.startTime)

    // Format based on event type
    switch event.Type {
    case observability.EventGraphStart:
        fmt.Printf("[%6dms] GRAPH_START: %s (entry=%s, exits=%d)\n",
            elapsed.Milliseconds(),
            event.Data["graph_name"],
            event.Data["entry_point"],
            event.Data["exit_point_count"])

    case observability.EventNodeStart:
        fmt.Printf("[%6dms]   NODE_START: %s (iteration=%d)\n",
            elapsed.Milliseconds(),
            event.Data["node_name"],
            event.Data["iteration"])

    case observability.EventNodeComplete:
        errFlag := ""
        if event.Data["has_error"].(bool) {
            errFlag = " [ERROR]"
        }
        fmt.Printf("[%6dms]   NODE_COMPLETE: %s%s\n",
            elapsed.Milliseconds(),
            event.Data["node_name"],
            errFlag)

    case observability.EventEdgeTransition:
        fmt.Printf("[%6dms]   EDGE_TRANSITION: %s → %s\n",
            elapsed.Milliseconds(),
            event.Data["from_node"],
            event.Data["to_node"])

    case observability.EventChainStart:
        fmt.Printf("[%6dms]     CHAIN_START: %d items\n",
            elapsed.Milliseconds(),
            event.Data["item_count"])

    case observability.EventStepStart:
        fmt.Printf("[%6dms]       STEP_START: %d/%d\n",
            elapsed.Milliseconds(),
            event.Data["step_index"].(int)+1,
            event.Data["total_steps"])

    case observability.EventCycleDetected:
        fmt.Printf("[%6dms]   CYCLE_DETECTED: node=%s, visits=%d\n",
            elapsed.Milliseconds(),
            event.Data["node_name"],
            event.Data["visit_count"])

    case observability.EventGraphComplete:
        fmt.Printf("[%6dms] GRAPH_COMPLETE: iterations=%d, exit=%s\n",
            elapsed.Milliseconds(),
            event.Data["iterations"],
            event.Data["exit_point"])
    }
}

// Register observer
observability.RegisterObserver("logging", NewLoggingObserver())

// Create graph config with logging observer
graphCfg := config.GraphConfig{
    Name:          "traced-workflow",
    Observer:      "logging", // Use custom observer
    MaxIterations: 20,
}

// Create chain config with logging observer
chainCfg := config.ChainConfig{
    CaptureIntermediateStates: false,
    Observer:                  "logging", // Use same observer
}

// Run workflow (from Example 2) - all events logged
graph, _ := state.NewGraph(graphCfg)
// ... build graph ...
result, err := graph.Execute(ctx, initialState)
```

**Console Output Example**:
```
Observability Showcase
======================
Running traced workflow with LoggingObserver

Execution Trace:
----------------
[     0ms] GRAPH_START: traced-workflow (entry=analyze, exits=1)
[     1ms]   NODE_START: analyze (iteration=1)
[     2ms]     CHAIN_START: 4 items
[     3ms]       STEP_START: 1/4
[     8ms]       STEP_COMPLETE: 1/4
[     9ms]       STEP_START: 2/4
[    14ms]       STEP_COMPLETE: 2/4
[    15ms]       STEP_START: 3/4
[    20ms]       STEP_COMPLETE: 3/4
[    21ms]       STEP_START: 4/4
[    26ms]       STEP_COMPLETE: 4/4
[    27ms]     CHAIN_COMPLETE: 4 steps
[    28ms]   NODE_COMPLETE: analyze
[    29ms]   EDGE_TRANSITION: analyze → review
[    30ms]   NODE_START: review (iteration=2)
[    45ms]   NODE_COMPLETE: review
[    46ms]   EDGE_EVALUATE: from=review, to=finalize, result=false
[    47ms]   EDGE_EVALUATE: from=review, to=revise, result=true
[    48ms]   EDGE_TRANSITION: review → revise
[    49ms]   NODE_START: revise (iteration=3)
[    55ms]   NODE_COMPLETE: revise
[    56ms]   EDGE_TRANSITION: revise → analyze
[    57ms]   NODE_START: analyze (iteration=4)
[    58ms]   CYCLE_DETECTED: node=analyze, visits=2
[    59ms]     CHAIN_START: 4 items
[    60ms]       STEP_START: 1/4
[   ..continues through second analysis...]
[   120ms]     CHAIN_COMPLETE: 4 steps
[   121ms]   NODE_COMPLETE: analyze
[   122ms]   EDGE_TRANSITION: analyze → review
[   123ms]   NODE_START: review (iteration=5)
[   138ms]   NODE_COMPLETE: review
[   139ms]   EDGE_EVALUATE: from=review, to=finalize, result=true
[   140ms]   EDGE_TRANSITION: review → finalize
[   141ms]   NODE_START: finalize (iteration=6)
[   150ms]   NODE_COMPLETE: finalize
[   151ms] GRAPH_COMPLETE: iterations=6, exit=finalize

Trace Summary
=============
Total execution time: 151ms
Events captured: 45
- Graph events: 2 (start, complete)
- Node events: 12 (6 start, 6 complete)
- Edge events: 11 (4 transitions, 7 evaluations)
- Chain events: 4 (2 start, 2 complete)
- Step events: 16 (8 start, 8 complete)
- Cycle events: 1

Performance Insights:
- Average node execution: 15ms
- Average chain step: 5ms
- Longest node: review (15ms)
- Cycle overhead: <1ms

Observer Overhead:
- With LoggingObserver: 151ms
- With NoOpObserver: 148ms (estimate)
- Overhead: ~2% (acceptable for debugging)

Event Data Available:
All events include: Type, Timestamp, Source, Data (metadata)
Example event.Data fields:
- Graph: graph_name, entry_point, exit_point_count
- Node: node_name, iteration, has_error
- Edge: from_node, to_node, edge_index, has_predicate
- Chain: item_count, has_progress, capture_intermediate
- Step: step_index, total_steps, has_error
```

**Integration Value**:
- Validates observer integration across all primitives
- Tests event emission at all key execution points
- Confirms event data structure and content
- Demonstrates custom Observer implementation
- Shows foundation for Phase 8 observability
- Provides debugging technique demonstration

---

## Implementation Priorities

### Priority 1: Foundation (Tomorrow Morning)
**Goal**: Establish basic usage patterns for both primitives

1. **Linear Workflow** (state graphs) - 30-45 min
   - Simplest state graph example
   - Foundation for understanding graph construction

2. **Document Analysis Chain** - 30-45 min
   - Simplest chain example
   - Pattern from classify-docs

3. **State as TContext** - 30-45 min
   - Key integration pattern
   - Shows state/workflows composition

**Total**: ~2-3 hours

---

### Priority 2: Integration (Tomorrow Afternoon)
**Goal**: Demonstrate composition and advanced features

4. **Conditional Routing** (state graphs) - 45-60 min
   - Predicate-based branching
   - Multiple exit points

5. **Chain Within Graph Node** - 45-60 min
   - Pattern composition
   - Practical integration

6. **Conversation Chain** - 30-45 min
   - Direct agent integration
   - Real LLM usage pattern

**Total**: ~2-3 hours

---

### Priority 3: Advanced (Next Session)
**Goal**: Show complex scenarios and debugging

7. **Cyclic Workflow** - 60 min
   - Intentional loops
   - Iteration management

8. **Direct Agent Integration** - 60 min
   - Primary pattern with tau-core
   - Multi-modal example

9. **Multi-Pattern Workflow** - 90 min
   - Complex realistic scenario
   - All features combined

**Total**: ~3-4 hours

---

### Priority 4: Polish (Future)
**Goal**: Optional advanced scenarios

10. **Hub Coordination** - 60 min
    - Optional multi-agent pattern
    - When hub adds value

11. **Error Handling** - 45 min
    - Debugging techniques
    - Recovery strategies

12. **Observability Showcase** - 60 min
    - Custom observer
    - Execution tracing

**Total**: ~3 hours

---

## Example Standards

Each example must meet these standards:

### Structure
- ✅ Self-contained `main.go` in dedicated directory
- ✅ `README.md` explaining purpose and usage
- ✅ `go.mod` if external dependencies needed
- ✅ Runnable with simple `go run main.go`

### Code Quality
- ✅ Clear godoc comments explaining purpose
- ✅ Descriptive variable and function names
- ✅ Error handling demonstrated
- ✅ Console output showing execution progress
- ✅ Comments explaining key concepts inline

### Documentation
- ✅ README.md with:
  - Purpose and key concepts
  - Prerequisites (if any)
  - How to run
  - Expected output
  - What it demonstrates
- ✅ Code comments explaining non-obvious logic
- ✅ Console output showing successful execution

### Consistency
- ✅ Similar structure across all examples
- ✅ Consistent naming conventions
- ✅ Similar README.md format
- ✅ Consistent logging style

---

## Benefits of This Plan

### Integration Testing Benefits
- **API Validation**: Real usage validates design decisions
- **Friction Identification**: Discovers usability issues
- **Composition Testing**: Validates pattern combinations
- **Error Path Exercising**: Tests failure scenarios
- **Performance Baseline**: Establishes timing expectations

### Demonstration Benefits
- **Progressive Complexity**: Simple → Intermediate → Advanced → Expert
- **Pattern Illustration**: Shows both primary (direct agent) and optional (hub) patterns
- **Copy-Paste Ready**: Provides starting points for real projects
- **Best Practices**: Documents idiomatic usage
- **Debugging Techniques**: Shows how to troubleshoot issues

### Development Benefits
- **Future Foundation**: Templates for Phase 5+ examples
- **Convention Establishment**: Sets standard example structure
- **Regression Suite**: Examples serve as integration tests
- **Documentation Support**: Examples referenced in package docs
- **User Onboarding**: Clear learning path for new users

---

## Success Criteria

The examples are successful when:

1. ✅ **Compilation**: All examples compile without errors
2. ✅ **Execution**: All examples run successfully to completion
3. ✅ **Clarity**: Purpose and key concepts clearly demonstrated
4. ✅ **Documentation**: README and comments explain usage
5. ✅ **Coverage**: All phases 2-4 capabilities shown
6. ✅ **Integration**: Composition patterns validated
7. ✅ **Progression**: Clear path from simple to complex
8. ✅ **Usability**: Colleagues can understand and adapt examples

---

## Notes

- Examples use simulation/mock where appropriate to avoid API key requirements
- All examples include both console output and state inspection
- Observer integration demonstrated with NoOpObserver and custom logger
- Direct agent integration (primary pattern) shown in multiple examples
- Hub coordination (optional pattern) shown separately to highlight difference
- State.State as TContext pattern emphasized as key integration point
- Error handling and debugging techniques included throughout

---

## Next Steps

Tomorrow morning (implementation session):
1. Start with Priority 1 examples (foundation)
2. Create directory structure as planned
3. Implement examples following standards
4. Test each example individually
5. Document in READMEs
6. Move to Priority 2 if time permits

This plan provides comprehensive demonstration of phases 2-4 capabilities through 12 well-structured examples covering simple to complex scenarios.
