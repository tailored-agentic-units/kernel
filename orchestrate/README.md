# orchestrate

Go-native agent coordination primitives for the TAU kernel — hub coordination, messaging, state management, workflow patterns, and observability.

## Packages

### config

Configuration structures for all orchestration primitives (hubs, state graphs, chains, parallel, conditional).

### hub

Multi-hub agent coordination with message routing and cross-hub communication.

- `Hub` - Central coordinator for agent registration and message dispatch
- `RegisterAgent` / `DeregisterAgent` for agent lifecycle
- Cross-hub agent registration for multi-hub topologies

### messaging

Structured inter-agent messaging with builders.

- Send (fire-and-forget), Request/Response, Broadcast, Pub/Sub patterns
- `Message` type with routing metadata
- Builder API: `NewRequest`, `NewResponse`, `NewBroadcast`

### observability

Execution tracing with configurable observers.

- `Observer` interface for execution events
- `NoOpObserver` - Silent (production default)
- `SlogObserver` - Structured logging via slog
- `MultiObserver` - Fan-out to multiple observers

### state

LangGraph-inspired state graph execution with checkpointing and persistence.

- `State` - Immutable state container with typed get/set
- `Graph` - Directed graph with nodes, edges, transition predicates
- `Checkpoint` / `CheckpointStore` for workflow persistence and recovery
- State secrets for sensitive data excluded from serialization

### workflows

Composable workflow patterns with state graph integration.

- `ProcessChain` - Sequential execution with state accumulation
- `ProcessParallel` - Concurrent execution with worker pools and order preservation
- `ProcessConditional` - Predicate-based routing with handler maps
- Integration helpers: `ChainNode`, `ParallelNode`, `ConditionalNode`

## Examples

See `orchestrate/examples/` for working demonstrations:

- **phase-01-hubs** - ISS Maintenance EVA: hub communication patterns
- **phase-02-03-state-graphs** - State graph execution
- **phase-04-sequential-chains** - Sequential chain processing
- **phase-05-parallel-execution** - Parallel execution with worker pools
- **phase-06-checkpointing** - Checkpoint persistence and recovery
- **phase-07-conditional-routing** - Document review workflow with conditional routing
- **darpa-procurement** - Multi-agent procurement analysis workflow
