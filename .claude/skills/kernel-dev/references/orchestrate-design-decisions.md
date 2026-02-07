# Design Decisions Reference

Key architectural decisions made during planning and development, organized by phase.

## Planning Phase Decisions

### Hub as Primary Coordination Primitive

**Context**: Need a coordination mechanism for multi-agent systems.

**Options**: Direct agent-to-agent communication, hub/broker pattern, event bus, actor model.

**Decision**: Hub pattern as persistent networking fabric.

**Rationale**: Clear coordination boundary, multi-hub networking through shared agents (fractal growth), natural fit for Go concurrency model (channels + goroutines), proven in research prototype.

**Trade-offs**: Agents must communicate through hub rather than directly.

### Minimal Agent Interface

**Context**: Hub needs to register agents from the tau-core library.

**Decision**: Single `ID()` method contract.

**Rationale**: Loose coupling, users compose agents with hub participation through wrapper pattern, follows contract interface pattern, easy to mock for testing.

### Message Structure Design

**Context**: Need structured messages for inter-agent communication.

**Decision**: Minimal core (ID, From, To, Type, Data, Timestamp) + optional metadata (ReplyTo, Topic, Priority, Headers).

**Rationale**: Start simple, avoid over-engineering. Dropped complex fields from research prototype (TTL, DeliveryMode, Ack, Retry, Sequence) for MVP.

### Hub and State Graph Relationship

**Context**: Need to clarify relationship between hub (coordination) and state graph (workflow).

**Decision**: Dual purpose — state graphs are independent but can leverage hub.

**Rationale**: Hub is persistent networking fabric; state graph is transient workflow execution. These are complementary, not mutually exclusive. State graph nodes CAN be hub agents, or can execute without hub.

### Configuration Strategy

**Decision**: Code-first with config structures (tau-core pattern). Configuration transforms to domain objects at boundaries. No persistence beyond initialization.

### Package Dependency Hierarchy

**Decision**: Hierarchical with strict bottom-up dependencies. Lower layers cannot import higher layers.

**Rationale**: Prevents circular dependencies, enables independent layer validation, matches Go package conventions.

### MessageHandler Callback Pattern

**Decision**: Function-based callbacks over interface methods or channel-based pull.

**Rationale**: Simple, flexible, proven in research. Easy to create inline handlers, supports request/response correlation, aligns with Go functional patterns.

### Multi-Hub Coordination via Shared Agents

**Decision**: Agents register with multiple hubs, acting as bridges (fractal growth). No central coordinator.

**Rationale**: Decentralized coordination, scalable, no hub hierarchy needed.

### Avoid Premature Abstraction

**Decision**: Build foundation first, let abstractions emerge organically. Skip biological metaphors (Cell, Organelle) and role abstractions (Orchestrator, Processor, Actor) until patterns emerge from real usage.

## Phase 1: Hub & Messaging

- **Hub Message Loop**: Polling-based over blocking select (processes messages from multiple agents without complex select cases)
- **Sender Filtering**: Both broadcast and pub/sub exclude sender from receiving own messages
- **MessageContext**: Handlers receive hub context, enabling context-aware responses
- **Request/Response Correlation**: Per-message response channels mapped by message ID

## Phase 2: State Management Core

- **State Not Generic**: Uses `map[string]any` instead of `State[T]` to avoid cascading generics and maintain pattern flexibility. Loses compile-time type safety but enables flexible workflow patterns.
- **Observer from Start**: Minimal Observer interface with NoOpObserver for zero-overhead default, prevents retrofit friction
- **Event.Data as Metadata**: Events contain execution metadata (keys_changed count, node name) not application data (privacy, performance)
- **Observer Registry Pattern**: String-based resolution enables JSON configuration without circular dependencies

## Phase 3: State Graph Execution Engine

- **Multiple Exit Points**: `map[string]bool` supports diverse workflow terminations (success/failure paths)
- **Cycle Detection on Every Revisit**: Emits EventCycleDetected for every `visit_count > 1`; observer decides what's concerning (no arbitrary thresholds)
- **Full Path Tracking**: Maintains complete execution path in `[]string` for debugging (~8KB for 1000 iterations)
- **Rich ExecutionError**: Captures NodeName, State, Path, and underlying Err for complete debugging context

## Phase 4: Sequential Chains

- **Package Naming**: `workflows/` over `patterns/` — more specific, better communicates multi-step orchestration intent
- **Rich ChainError**: Captures StepIndex, Item, State, and underlying error
- **Progress Callback After Success**: Called after step completion (represents completed work), aligns with fold/reduce semantics
- **Intermediate State Capture Includes Initial**: When enabled, index 0 captures initial state for complete evolution

## Phase 5: Parallel Execution

- **SlogObserver Added Early**: Practical observability needed during development; validates observer pattern before full Phase 8
- **Three-Channel Coordination**: Work queue → workers → result channel + background collector (prevents deadlocks when result buffer fills)
- **Error Handling Semantics**: Return error when (FailFast=true AND any failed) OR (FailFast=false AND ALL failed). No error on partial success — check `result.Errors`
- **Worker Pool Auto-Detection**: `min(NumCPU*2, WorkerCap, itemCount)` — 2x CPU optimal for I/O-bound agent API calls
- **Order Preservation via Indexing**: Tag each item/result with original index; build final slices by sequential iteration

## Phase 6: Checkpointing

- **State IS Checkpoint**: Eliminated separate Checkpoint wrapper. Checkpoint metadata (runID, checkpointNode, timestamp) lives in State directly — checkpoint is execution provenance.
- **Resume After Node**: Checkpoint saved AFTER node execution; resume continues to next node (checkpoint represents completed work)
- **Checkpoint Save Errors = Fail-Fast**: Execution halts if checkpoint cannot be saved (production reliability)
- **Single ID per Run (Overwrite Model)**: RunID serves as identifier; each save overwrites previous (simpler, future-proof)
- **CheckpointStore Registry**: Follows Observer pattern — string-based resolution for extensibility

## Phase 7: Conditional Routing + Integration

- **No Hub Parameter in Integration Helpers**: Hub captured in closures if needed (eliminates parameter clutter; direct tau-core is primary pattern)
- **Aggregator Function for ParallelNode**: Required function transforms `[]TResult` to `State` (bridges pattern output and graph expectations)
- **No Progress Callback for Conditional**: Single decision point, not iterative — progress callbacks only for patterns processing multiple items
- **Routes Structure with Default**: Map with optional Default handler as fallback when predicate returns unexpected route

## Cross-Cutting Principles

**Pattern Independence**: All workflow patterns work with direct tau-core agent calls (primary), hub coordination (optional, captured in closures), or pure computation (no agents required).

**Error Context Preservation**: Rich error types (ChainError, ParallelError, ConditionalError, ExecutionError) capture execution context with `Unwrap()` support.

**Immutability Through Composition**: State transformations create new state. State flows through patterns and graphs unchanged in identity, transformed in content. Metadata preserved through Clone/Set/Merge.

**Observer Integration Strategy**: Minimal hooks from Phase 2 (NoOpObserver default), implementations added incrementally (SlogObserver Phase 5), full production observability planned for Phase 8.
