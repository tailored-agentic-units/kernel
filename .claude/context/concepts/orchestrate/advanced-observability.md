# Advanced Observability

## Summary

Extend the existing observer foundation with production-grade observability features: execution trace correlation across workflows, decision point logging, performance metrics aggregation, and OpenTelemetry integration.

## Current State

The observability infrastructure is complete and integrated across all packages:

- **Observer interface** with `OnEvent(ctx, event)` contract
- **Event model** with Type, Timestamp, Source, and Data fields
- **EventType constants** covering all workflow phases (graph execution, chains, parallel, conditional, checkpointing)
- **NoOpObserver** for zero-overhead operation
- **SlogObserver** for structured logging via Go's `slog` package
- **MultiObserver** for broadcasting events to multiple observers
- **Observer registry** with thread-safe `GetObserver`/`RegisterObserver` for configuration-driven selection
- **Full integration** across hub, state, and workflows packages

All configuration defaults use `"slog"` observer. Users can override to `"noop"` for zero overhead or register custom implementations via the registry.

## Requirements

### Execution Trace Correlation

Correlate events across workflow boundaries so that a single execution can be traced end-to-end, even when workflows compose (e.g., a state graph node that runs a parallel workflow internally).

**Needs addressed:**
- Assign a trace ID at the top-level execution entry point
- Propagate trace ID through nested workflow invocations via context
- Include trace ID in all emitted events
- Enable reconstruction of full execution timeline from collected events

### Decision Point Logging

Capture the reasoning and evaluation context at conditional routing points and edge transitions where the workflow makes a decision.

**Needs addressed:**
- Record which predicate was evaluated and what it returned
- Capture the state snapshot at the moment of decision
- Log which route was selected and why (predicate name, result)
- Support audit trails for compliance-sensitive workflows

### Performance Metrics Aggregation

Collect and aggregate timing and resource metrics across workflow execution.

**Needs addressed:**
- Node execution duration (start/complete timestamps already in events)
- Workflow total duration
- Parallel worker utilization (active workers, queue depth)
- Per-step latency distribution across chain executions
- Error rates and categorization
- Aggregation into summary statistics (min, max, mean, p50, p95, p99)

### OpenTelemetry Integration

Provide an Observer implementation that exports events as OpenTelemetry spans and metrics, enabling integration with standard observability platforms (Jaeger, Grafana, Datadog).

**Needs addressed:**
- Map workflow execution to OTel spans (graph → parent span, nodes → child spans)
- Map EventType to span events/attributes
- Export metrics (counters, histograms) for execution statistics
- Context propagation compatible with OTel trace context
- Optional dependency (separate package or build tag to avoid mandatory OTel dependency)

## Design Constraints

- **Backward compatible**: Existing Observer interface must not change. New capabilities extend through new observer implementations or optional interface methods.
- **Zero-overhead default**: NoOpObserver behavior must remain zero-cost. Advanced features are opt-in.
- **Registry compatible**: New observers must integrate with the existing string-based registry pattern.
- **No core package coupling**: Advanced observers should not add dependencies to the core `pkg/observability` package. Consider a separate package (e.g., `pkg/otel`) or a supplemental module.
- **Event model stability**: The existing Event structure and EventType constants are the contract. New event types can be added; existing ones should not change semantics.

## Open Questions

These questions should be resolved through production usage feedback and design discussion before implementation begins:

1. **Scope boundary**: Should OpenTelemetry integration live in this module or in a separate supplemental module (e.g., `tau-orchestrate-otel`) to avoid adding OTel as a transitive dependency?
2. **Trace context propagation**: Should trace IDs be carried in `context.Context` (Go standard), in `Event.Data`, or in a new field on Event?
3. **Metrics observer vs. trace observer**: Should metrics aggregation and trace correlation be separate observer implementations, or a single combined observer?
4. **Confidence scoring**: The original roadmap mentioned confidence scoring utilities. Is this an observability concern (recording confidence in events) or a workflow concern (routing based on confidence thresholds)? If the latter, it may belong in `pkg/workflows` rather than `pkg/observability`.
5. **Event enrichment**: Should the existing Event.Data map be formalized with typed keys/values, or should it remain `map[string]any` for flexibility?
6. **Sampling**: For high-throughput workflows, should observers support event sampling to reduce overhead?

## Acceptance Criteria

- Trace correlation enables end-to-end reconstruction of composed workflow execution
- Decision points produce audit-quality logs with state snapshots and predicate results
- Performance metrics are aggregatable into summary statistics without external tooling
- OpenTelemetry observer produces valid spans and metrics consumable by standard OTel backends
- All new observer implementations pass through the existing registry pattern
- No performance regression for existing NoOpObserver and SlogObserver usage
- Extensibility demonstrated through at least one custom observer implementation
