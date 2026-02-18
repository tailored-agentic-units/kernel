package workflows_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/tailored-agentic-units/kernel/orchestrate/config"
	"github.com/tailored-agentic-units/kernel/observability"
	"github.com/tailored-agentic-units/kernel/orchestrate/workflows"
)

type testState struct {
	value string
	count int
}

func TestProcessConditional_BasicRouting(t *testing.T) {
	observability.RegisterObserver("noop", &observability.NoOpObserver{})

	tests := []struct {
		name          string
		state         testState
		predicateFunc func(testState) (string, error)
		routes        workflows.Routes[testState]
		want          testState
		wantErr       bool
	}{
		{
			name:  "route_to_first_handler",
			state: testState{value: "initial", count: 0},
			predicateFunc: func(s testState) (string, error) {
				return "route1", nil
			},
			routes: workflows.Routes[testState]{
				Handlers: map[string]workflows.RouteHandler[testState]{
					"route1": func(ctx context.Context, s testState) (testState, error) {
						s.value = "route1_handled"
						s.count++
						return s, nil
					},
					"route2": func(ctx context.Context, s testState) (testState, error) {
						s.value = "route2_handled"
						return s, nil
					},
				},
			},
			want:    testState{value: "route1_handled", count: 1},
			wantErr: false,
		},
		{
			name:  "route_to_second_handler",
			state: testState{value: "initial", count: 5},
			predicateFunc: func(s testState) (string, error) {
				return "route2", nil
			},
			routes: workflows.Routes[testState]{
				Handlers: map[string]workflows.RouteHandler[testState]{
					"route1": func(ctx context.Context, s testState) (testState, error) {
						s.value = "route1_handled"
						return s, nil
					},
					"route2": func(ctx context.Context, s testState) (testState, error) {
						s.value = "route2_handled"
						s.count += 10
						return s, nil
					},
				},
			},
			want:    testState{value: "route2_handled", count: 15},
			wantErr: false,
		},
		{
			name:  "state_based_routing",
			state: testState{value: "initial", count: 42},
			predicateFunc: func(s testState) (string, error) {
				if s.count > 40 {
					return "high", nil
				}
				return "low", nil
			},
			routes: workflows.Routes[testState]{
				Handlers: map[string]workflows.RouteHandler[testState]{
					"high": func(ctx context.Context, s testState) (testState, error) {
						s.value = "high_priority"
						return s, nil
					},
					"low": func(ctx context.Context, s testState) (testState, error) {
						s.value = "low_priority"
						return s, nil
					},
				},
			},
			want:    testState{value: "high_priority", count: 42},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			cfg := config.DefaultConditionalConfig()
			cfg.Observer = "noop"

			got, err := workflows.ProcessConditional(
				ctx,
				cfg,
				tt.state,
				tt.predicateFunc,
				tt.routes,
			)

			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessConditional() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.want {
				t.Errorf("ProcessConditional() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestProcessConditional_DefaultHandler(t *testing.T) {
	observability.RegisterObserver("noop", &observability.NoOpObserver{})

	ctx := context.Background()
	cfg := config.DefaultConditionalConfig()
	cfg.Observer = "noop"

	state := testState{value: "initial", count: 0}

	predicate := func(s testState) (string, error) {
		return "unknown_route", nil
	}

	routes := workflows.Routes[testState]{
		Handlers: map[string]workflows.RouteHandler[testState]{
			"known_route": func(ctx context.Context, s testState) (testState, error) {
				s.value = "known"
				return s, nil
			},
		},
		Default: func(ctx context.Context, s testState) (testState, error) {
			s.value = "default_handled"
			s.count = 99
			return s, nil
		},
	}

	got, err := workflows.ProcessConditional(ctx, cfg, state, predicate, routes)
	if err != nil {
		t.Fatalf("ProcessConditional() unexpected error: %v", err)
	}

	want := testState{value: "default_handled", count: 99}
	if got != want {
		t.Errorf("ProcessConditional() = %+v, want %+v", got, want)
	}
}

func TestProcessConditional_MissingRouteWithoutDefault(t *testing.T) {
	observability.RegisterObserver("noop", &observability.NoOpObserver{})

	ctx := context.Background()
	cfg := config.DefaultConditionalConfig()
	cfg.Observer = "noop"

	state := testState{value: "initial", count: 0}

	predicate := func(s testState) (string, error) {
		return "nonexistent_route", nil
	}

	routes := workflows.Routes[testState]{
		Handlers: map[string]workflows.RouteHandler[testState]{
			"existing_route": func(ctx context.Context, s testState) (testState, error) {
				s.value = "handled"
				return s, nil
			},
		},
	}

	_, err := workflows.ProcessConditional(ctx, cfg, state, predicate, routes)
	if err == nil {
		t.Fatal("ProcessConditional() expected error for missing route, got nil")
	}

	var condErr workflows.ConditionalError[testState]
	if !errors.As(err, &condErr) {
		t.Errorf("ProcessConditional() error type = %T, want ConditionalError", err)
	}
}

func TestProcessConditional_PredicateError(t *testing.T) {
	observability.RegisterObserver("noop", &observability.NoOpObserver{})

	ctx := context.Background()
	cfg := config.DefaultConditionalConfig()
	cfg.Observer = "noop"

	state := testState{value: "initial", count: 0}

	predicate := func(s testState) (string, error) {
		return "", fmt.Errorf("predicate evaluation failed")
	}

	routes := workflows.Routes[testState]{
		Handlers: map[string]workflows.RouteHandler[testState]{
			"route1": func(ctx context.Context, s testState) (testState, error) {
				s.value = "handled"
				return s, nil
			},
		},
	}

	_, err := workflows.ProcessConditional(ctx, cfg, state, predicate, routes)
	if err == nil {
		t.Fatal("ProcessConditional() expected error from predicate, got nil")
	}

	var condErr workflows.ConditionalError[testState]
	if !errors.As(err, &condErr) {
		t.Errorf("ProcessConditional() error type = %T, want ConditionalError", err)
	}
}

func TestProcessConditional_HandlerError(t *testing.T) {
	observability.RegisterObserver("noop", &observability.NoOpObserver{})

	ctx := context.Background()
	cfg := config.DefaultConditionalConfig()
	cfg.Observer = "noop"

	state := testState{value: "initial", count: 0}

	predicate := func(s testState) (string, error) {
		return "failing_route", nil
	}

	routes := workflows.Routes[testState]{
		Handlers: map[string]workflows.RouteHandler[testState]{
			"failing_route": func(ctx context.Context, s testState) (testState, error) {
				return s, fmt.Errorf("handler execution failed")
			},
		},
	}

	_, err := workflows.ProcessConditional(ctx, cfg, state, predicate, routes)
	if err == nil {
		t.Fatal("ProcessConditional() expected error from handler, got nil")
	}

	var condErr workflows.ConditionalError[testState]
	if !errors.As(err, &condErr) {
		t.Errorf("ProcessConditional() error type = %T, want ConditionalError", err)
	}
}

func TestProcessConditional_ContextCancellation(t *testing.T) {
	observability.RegisterObserver("noop", &observability.NoOpObserver{})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cfg := config.DefaultConditionalConfig()
	cfg.Observer = "noop"

	state := testState{value: "initial", count: 0}

	predicate := func(s testState) (string, error) {
		return "route1", nil
	}

	routes := workflows.Routes[testState]{
		Handlers: map[string]workflows.RouteHandler[testState]{
			"route1": func(ctx context.Context, s testState) (testState, error) {
				s.value = "handled"
				return s, nil
			},
		},
	}

	_, err := workflows.ProcessConditional(ctx, cfg, state, predicate, routes)
	if err == nil {
		t.Fatal("ProcessConditional() expected error for cancelled context, got nil")
	}

	var condErr workflows.ConditionalError[testState]
	if !errors.As(err, &condErr) {
		t.Errorf("ProcessConditional() error type = %T, want ConditionalError", err)
	}
}

func TestProcessConditional_ObserverEvents(t *testing.T) {
	capture := newCaptureObserver()
	observability.RegisterObserver("capture", capture)

	ctx := context.Background()
	cfg := config.DefaultConditionalConfig()
	cfg.Observer = "capture"

	state := testState{value: "initial", count: 0}

	predicate := func(s testState) (string, error) {
		return "test_route", nil
	}

	routes := workflows.Routes[testState]{
		Handlers: map[string]workflows.RouteHandler[testState]{
			"test_route": func(ctx context.Context, s testState) (testState, error) {
				s.value = "handled"
				return s, nil
			},
		},
	}

	_, err := workflows.ProcessConditional(ctx, cfg, state, predicate, routes)
	if err != nil {
		t.Fatalf("ProcessConditional() unexpected error: %v", err)
	}

	expectedEvents := []observability.EventType{
		workflows.EventRouteEvaluate,
		workflows.EventRouteSelect,
		workflows.EventRouteExecute,
	}

	if len(capture.events) != len(expectedEvents) {
		t.Errorf("Expected %d events, got %d", len(expectedEvents), len(capture.events))
	}

	for i, expectedType := range expectedEvents {
		if i >= len(capture.events) {
			t.Errorf("Missing event %d: %s", i, expectedType)
			continue
		}
		if capture.events[i].Type != expectedType {
			t.Errorf("Event %d type = %s, want %s", i, capture.events[i].Type, expectedType)
		}
	}
}

func TestConditionalError_ErrorMessage(t *testing.T) {
	tests := []struct {
		name        string
		err         workflows.ConditionalError[testState]
		wantContain string
	}{
		{
			name: "with_route",
			err: workflows.ConditionalError[testState]{
				Route: "test_route",
				State: testState{value: "test", count: 1},
				Err:   fmt.Errorf("underlying error"),
			},
			wantContain: "test_route",
		},
		{
			name: "without_route",
			err: workflows.ConditionalError[testState]{
				Route: "",
				State: testState{value: "test", count: 1},
				Err:   fmt.Errorf("underlying error"),
			},
			wantContain: "conditional routing failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.err.Error()
			if msg == "" {
				t.Error("Error() returned empty string")
			}
			if tt.wantContain != "" && !contains(msg, tt.wantContain) {
				t.Errorf("Error() = %q, want to contain %q", msg, tt.wantContain)
			}
		})
	}
}

func TestConditionalError_Unwrap(t *testing.T) {
	underlying := fmt.Errorf("underlying error")
	condErr := workflows.ConditionalError[testState]{
		Route: "test",
		State: testState{},
		Err:   underlying,
	}

	unwrapped := condErr.Unwrap()
	if unwrapped != underlying {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, underlying)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
