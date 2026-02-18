package state_test

import (
	"context"
	"errors"
	"testing"

	"github.com/tailored-agentic-units/kernel/observability"
	"github.com/tailored-agentic-units/kernel/orchestrate/state"
)

func TestFunctionNode_Execute(t *testing.T) {
	tests := []struct {
		name      string
		fn        func(context.Context, state.State) (state.State, error)
		wantErr   bool
		wantKey   string
		wantValue any
	}{
		{
			name: "simple state transformation",
			fn: func(ctx context.Context, s state.State) (state.State, error) {
				return s.Set("result", "success"), nil
			},
			wantErr:   false,
			wantKey:   "result",
			wantValue: "success",
		},
		{
			name: "error propagation",
			fn: func(ctx context.Context, s state.State) (state.State, error) {
				return s, errors.New("test error")
			},
			wantErr: true,
		},
		{
			name: "state accumulation",
			fn: func(ctx context.Context, s state.State) (state.State, error) {
				existingVal, exists := s.Get("counter")
				if !exists {
					return s.Set("counter", 1), nil
				}
				count := existingVal.(int)
				return s.Set("counter", count+1), nil
			},
			wantErr:   false,
			wantKey:   "counter",
			wantValue: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := state.NewFunctionNode(tt.fn)
			initialState := state.New(observability.NoOpObserver{})

			result, err := node.Execute(context.Background(), initialState)

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.wantKey != "" {
				val, exists := result.Get(tt.wantKey)
				if !exists {
					t.Errorf("Execute() did not set key %q", tt.wantKey)
				}
				if val != tt.wantValue {
					t.Errorf("Execute() value = %v, want %v", val, tt.wantValue)
				}
			}
		})
	}
}

func TestFunctionNode_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	node := state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
		if ctx.Err() != nil {
			return s, ctx.Err()
		}
		return s.Set("result", "should not reach"), nil
	})

	initialState := state.New(observability.NoOpObserver{})
	result, err := node.Execute(ctx, initialState)

	if err == nil {
		t.Error("Execute() should return error when context is cancelled")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Execute() error = %v, want context.Canceled", err)
	}

	val, exists := result.Get("result")
	if exists && val != nil {
		t.Error("Execute() should not modify state when context cancelled")
	}
}

func TestFunctionNode_StateImmutability(t *testing.T) {
	node := state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
		return s.Set("modified", true), nil
	})

	initialState := state.New(observability.NoOpObserver{})
	initialState = initialState.Set("original", true)

	result, err := node.Execute(context.Background(), initialState)
	if err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}

	origVal, exists := initialState.Get("original")
	if !exists || origVal != true {
		t.Error("Execute() should not modify original state")
	}

	_, modExists := initialState.Get("modified")
	if modExists {
		t.Error("Execute() should not modify original state")
	}

	resOrigVal, resOrigExists := result.Get("original")
	resModVal, resModExists := result.Get("modified")
	if !resOrigExists || resOrigVal != true {
		t.Error("Execute() result should preserve original keys")
	}
	if !resModExists || resModVal != true {
		t.Error("Execute() result should contain modifications")
	}
}
