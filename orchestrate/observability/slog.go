package observability

import (
	"context"
	"log/slog"
)

// SlogObserver provides structured logging observability using Go's slog package.
//
// SlogObserver writes all orchestration events to a structured logger at Info level,
// capturing event type, source, timestamp, and associated metadata. This enables
// debugging and monitoring of workflow execution through standard log aggregation tools.
//
// The observer uses slog's context-aware logging (InfoContext) to propagate cancellation
// signals and tracing context from the workflow execution context.
//
// Example:
//
//	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
//	observer := observability.NewSlogObserver(logger)
//	observability.RegisterObserver("production", observer)
//
//	cfg := config.ChainConfig{Observer: "production"}
//	result, err := workflows.ProcessChain(ctx, cfg, items, initial, processor, nil)
type SlogObserver struct {
	logger *slog.Logger
}

// NewSlogObserver creates a new SlogObserver with the specified logger.
//
// The logger parameter allows customization of the slog handler, output destination,
// and log level filtering. Pass slog.Default() for the default logger configuration.
//
// Example:
//
//	observer := observability.NewSlogObserver(slog.Default())
//
// Example with custom handler:
//
//	handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
//	    Level: slog.LevelDebug,
//	})
//	logger := slog.New(handler)
//	observer := observability.NewSlogObserver(logger)
func NewSlogObserver(logger *slog.Logger) *SlogObserver {
	return &SlogObserver{
		logger: logger,
	}
}

// OnEvent logs the event at Info level with structured fields.
//
// The event is logged with the following slog attributes:
//   - type: The EventType constant (e.g., "chain.start")
//   - source: The component that emitted the event (e.g., "workflows.ProcessChain")
//   - timestamp: When the event occurred
//   - data: Event-specific metadata map
//
// The context is propagated to InfoContext for cancellation and tracing integration.
func (o *SlogObserver) OnEvent(ctx context.Context, event Event) {
	o.logger.InfoContext(
		ctx,
		"Event",
		"type", event.Type,
		"source", event.Source,
		"timestamp", event.Timestamp,
		"data", event.Data,
	)
}
