package observability_test

import (
	"bytes"
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/tailored-agentic-units/kernel/observability"
)

func TestLevel_String(t *testing.T) {
	tests := []struct {
		name  string
		level observability.Level
		want  string
	}{
		{name: "trace range", level: 1, want: "TRACE"},
		{name: "verbose maps to DEBUG", level: observability.LevelVerbose, want: "DEBUG"},
		{name: "info maps to INFO", level: observability.LevelInfo, want: "INFO"},
		{name: "warning maps to WARN", level: observability.LevelWarning, want: "WARN"},
		{name: "error maps to ERROR", level: observability.LevelError, want: "ERROR"},
		{name: "fatal range", level: 21, want: "FATAL"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.level.String(); got != tt.want {
				t.Errorf("Level(%d).String() = %q, want %q", tt.level, got, tt.want)
			}
		})
	}
}

func TestLevel_SlogLevel(t *testing.T) {
	tests := []struct {
		name  string
		level observability.Level
		want  slog.Level
	}{
		{name: "verbose maps to Debug", level: observability.LevelVerbose, want: slog.LevelDebug},
		{name: "info maps to Info", level: observability.LevelInfo, want: slog.LevelInfo},
		{name: "warning maps to Warn", level: observability.LevelWarning, want: slog.LevelWarn},
		{name: "error maps to Error", level: observability.LevelError, want: slog.LevelError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.level.SlogLevel(); got != tt.want {
				t.Errorf("Level(%d).SlogLevel() = %v, want %v", tt.level, got, tt.want)
			}
		})
	}
}

func TestLevel_OTelAlignment(t *testing.T) {
	if observability.LevelVerbose != 5 {
		t.Errorf("LevelVerbose = %d, want 5 (OTel DEBUG range)", observability.LevelVerbose)
	}
	if observability.LevelInfo != 9 {
		t.Errorf("LevelInfo = %d, want 9 (OTel INFO range)", observability.LevelInfo)
	}
	if observability.LevelWarning != 13 {
		t.Errorf("LevelWarning = %d, want 13 (OTel WARN range)", observability.LevelWarning)
	}
	if observability.LevelError != 17 {
		t.Errorf("LevelError = %d, want 17 (OTel ERROR range)", observability.LevelError)
	}
}

func TestNoOpObserver(t *testing.T) {
	obs := observability.NoOpObserver{}
	obs.OnEvent(context.Background(), observability.Event{
		Type:      "test.event",
		Level:     observability.LevelInfo,
		Timestamp: time.Now(),
		Source:    "test",
		Data:      map[string]any{"key": "value"},
	})
}

func TestMultiObserver(t *testing.T) {
	var events1, events2 []observability.Event

	obs1 := &captureObserver{events: &events1}
	obs2 := &captureObserver{events: &events2}

	multi := observability.NewMultiObserver(obs1, obs2)

	event := observability.Event{
		Type:      "test.event",
		Level:     observability.LevelInfo,
		Timestamp: time.Now(),
		Source:    "test",
		Data:      map[string]any{"key": "value"},
	}

	multi.OnEvent(context.Background(), event)

	if len(events1) != 1 {
		t.Errorf("observer 1 received %d events, want 1", len(events1))
	}
	if len(events2) != 1 {
		t.Errorf("observer 2 received %d events, want 1", len(events2))
	}
	if events1[0].Type != "test.event" {
		t.Errorf("observer 1 event type = %q, want %q", events1[0].Type, "test.event")
	}
}

func TestMultiObserver_NilFiltering(t *testing.T) {
	var events []observability.Event
	obs := &captureObserver{events: &events}

	multi := observability.NewMultiObserver(nil, obs, nil)

	multi.OnEvent(context.Background(), observability.Event{
		Type:  "test.event",
		Level: observability.LevelInfo,
	})

	if len(events) != 1 {
		t.Errorf("received %d events, want 1 (nil observers should be filtered)", len(events))
	}
}

func TestSlogObserver_LevelMapping(t *testing.T) {
	tests := []struct {
		name      string
		level     observability.Level
		minLevel  slog.Level
		expectLog bool
	}{
		{name: "verbose at debug handler", level: observability.LevelVerbose, minLevel: slog.LevelDebug, expectLog: true},
		{name: "verbose at info handler", level: observability.LevelVerbose, minLevel: slog.LevelInfo, expectLog: false},
		{name: "info at info handler", level: observability.LevelInfo, minLevel: slog.LevelInfo, expectLog: true},
		{name: "info at warn handler", level: observability.LevelInfo, minLevel: slog.LevelWarn, expectLog: false},
		{name: "warning at warn handler", level: observability.LevelWarning, minLevel: slog.LevelWarn, expectLog: true},
		{name: "error at error handler", level: observability.LevelError, minLevel: slog.LevelError, expectLog: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
				Level: tt.minLevel,
			}))

			obs := observability.NewSlogObserver(logger)
			obs.OnEvent(context.Background(), observability.Event{
				Type:      "test.event",
				Level:     tt.level,
				Timestamp: time.Now(),
				Source:    "test",
			})

			hasOutput := buf.Len() > 0
			if hasOutput != tt.expectLog {
				t.Errorf("log output = %v, want %v (buf: %q)", hasOutput, tt.expectLog, buf.String())
			}
		})
	}
}

func TestSlogObserver_EventTypeAsMessage(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	obs := observability.NewSlogObserver(logger)
	obs.OnEvent(context.Background(), observability.Event{
		Type:      "kernel.run.start",
		Level:     observability.LevelInfo,
		Timestamp: time.Now(),
		Source:    "kernel.Run",
		Data: map[string]any{
			"prompt_length": 42,
		},
	})

	output := buf.String()
	if !contains(output, "kernel.run.start") {
		t.Errorf("expected event type as log message, got: %s", output)
	}
	if !contains(output, "source=kernel.Run") {
		t.Errorf("expected source attribute, got: %s", output)
	}
	if !contains(output, "prompt_length=42") {
		t.Errorf("expected data attributes, got: %s", output)
	}
}

func TestRegistry_GetObserver(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{name: "noop exists", key: "noop", wantErr: false},
		{name: "slog exists", key: "slog", wantErr: false},
		{name: "unknown fails", key: "nonexistent", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obs, err := observability.GetObserver(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetObserver(%q) error = %v, wantErr %v", tt.key, err, tt.wantErr)
			}
			if !tt.wantErr && obs == nil {
				t.Errorf("GetObserver(%q) returned nil observer", tt.key)
			}
		})
	}
}

func TestRegistry_RegisterAndGet(t *testing.T) {
	var events []observability.Event
	custom := &captureObserver{events: &events}

	observability.RegisterObserver("test-custom", custom)

	obs, err := observability.GetObserver("test-custom")
	if err != nil {
		t.Fatalf("GetObserver failed: %v", err)
	}

	obs.OnEvent(context.Background(), observability.Event{
		Type:  "test.event",
		Level: observability.LevelInfo,
	})

	if len(events) != 1 {
		t.Errorf("received %d events, want 1", len(events))
	}
}

type captureObserver struct {
	events *[]observability.Event
}

func (c *captureObserver) OnEvent(ctx context.Context, event observability.Event) {
	*c.events = append(*c.events, event)
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
