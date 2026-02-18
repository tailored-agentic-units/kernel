package state_test

import (
	"testing"

	"github.com/tailored-agentic-units/kernel/observability"
	"github.com/tailored-agentic-units/kernel/orchestrate/state"
)

func TestEdge_Structure(t *testing.T) {
	predicate := state.AlwaysTransition()
	edge := state.Edge{
		From:      "nodeA",
		To:        "nodeB",
		Predicate: predicate,
	}

	if edge.From != "nodeA" {
		t.Errorf("Edge.From = %v, want %v", edge.From, "nodeA")
	}
	if edge.To != "nodeB" {
		t.Errorf("Edge.To = %v, want %v", edge.To, "nodeB")
	}
	if edge.Predicate == nil {
		t.Error("Edge.Predicate should not be nil")
	}
}

func TestPredicate_AlwaysTransition(t *testing.T) {
	predicate := state.AlwaysTransition()
	s := state.New(observability.NoOpObserver{})

	if !predicate(s) {
		t.Error("AlwaysTransition() should always return true")
	}

	s = s.Set("key", "value")
	if !predicate(s) {
		t.Error("AlwaysTransition() should return true regardless of state")
	}
}

func TestPredicate_KeyExists(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		setState func(state.State) state.State
		want     bool
	}{
		{
			name: "key exists",
			key:  "test-key",
			setState: func(s state.State) state.State {
				return s.Set("test-key", "value")
			},
			want: true,
		},
		{
			name: "key does not exist",
			key:  "missing-key",
			setState: func(s state.State) state.State {
				return s
			},
			want: false,
		},
		{
			name: "key exists with nil value",
			key:  "nil-key",
			setState: func(s state.State) state.State {
				return s.Set("nil-key", nil)
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			predicate := state.KeyExists(tt.key)
			s := state.New(observability.NoOpObserver{})
			s = tt.setState(s)

			if got := predicate(s); got != tt.want {
				t.Errorf("KeyExists(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}

func TestPredicate_KeyEquals(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    any
		setState func(state.State) state.State
		want     bool
	}{
		{
			name:  "string value equals",
			key:   "status",
			value: "approved",
			setState: func(s state.State) state.State {
				return s.Set("status", "approved")
			},
			want: true,
		},
		{
			name:  "string value not equals",
			key:   "status",
			value: "approved",
			setState: func(s state.State) state.State {
				return s.Set("status", "pending")
			},
			want: false,
		},
		{
			name:  "int value equals",
			key:   "count",
			value: 42,
			setState: func(s state.State) state.State {
				return s.Set("count", 42)
			},
			want: true,
		},
		{
			name:  "int value not equals",
			key:   "count",
			value: 42,
			setState: func(s state.State) state.State {
				return s.Set("count", 100)
			},
			want: false,
		},
		{
			name:  "key does not exist",
			key:   "missing",
			value: "value",
			setState: func(s state.State) state.State {
				return s
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			predicate := state.KeyEquals(tt.key, tt.value)
			s := state.New(observability.NoOpObserver{})
			s = tt.setState(s)

			if got := predicate(s); got != tt.want {
				t.Errorf("KeyEquals(%q, %v) = %v, want %v", tt.key, tt.value, got, tt.want)
			}
		})
	}
}

func TestPredicate_Not(t *testing.T) {
	s := state.New(observability.NoOpObserver{})
	s = s.Set("key", "value")

	truePredicate := state.KeyExists("key")
	falsePredicate := state.Not(truePredicate)

	if falsePredicate(s) {
		t.Error("Not() should invert true predicate to false")
	}

	missingPredicate := state.KeyExists("missing")
	notMissingPredicate := state.Not(missingPredicate)

	if !notMissingPredicate(s) {
		t.Error("Not() should invert false predicate to true")
	}
}

func TestPredicate_And(t *testing.T) {
	tests := []struct {
		name       string
		predicates []state.TransitionPredicate
		setState   func(state.State) state.State
		want       bool
	}{
		{
			name: "all true",
			predicates: []state.TransitionPredicate{
				state.KeyExists("key1"),
				state.KeyExists("key2"),
			},
			setState: func(s state.State) state.State {
				s = s.Set("key1", "value1")
				s = s.Set("key2", "value2")
				return s
			},
			want: true,
		},
		{
			name: "one false",
			predicates: []state.TransitionPredicate{
				state.KeyExists("key1"),
				state.KeyExists("missing"),
			},
			setState: func(s state.State) state.State {
				return s.Set("key1", "value1")
			},
			want: false,
		},
		{
			name: "all false",
			predicates: []state.TransitionPredicate{
				state.KeyExists("missing1"),
				state.KeyExists("missing2"),
			},
			setState: func(s state.State) state.State {
				return s
			},
			want: false,
		},
		{
			name:       "empty predicates",
			predicates: []state.TransitionPredicate{},
			setState: func(s state.State) state.State {
				return s
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			predicate := state.And(tt.predicates...)
			s := state.New(observability.NoOpObserver{})
			s = tt.setState(s)

			if got := predicate(s); got != tt.want {
				t.Errorf("And() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPredicate_Or(t *testing.T) {
	tests := []struct {
		name       string
		predicates []state.TransitionPredicate
		setState   func(state.State) state.State
		want       bool
	}{
		{
			name: "all true",
			predicates: []state.TransitionPredicate{
				state.KeyExists("key1"),
				state.KeyExists("key2"),
			},
			setState: func(s state.State) state.State {
				s = s.Set("key1", "value1")
				s = s.Set("key2", "value2")
				return s
			},
			want: true,
		},
		{
			name: "one true",
			predicates: []state.TransitionPredicate{
				state.KeyExists("key1"),
				state.KeyExists("missing"),
			},
			setState: func(s state.State) state.State {
				return s.Set("key1", "value1")
			},
			want: true,
		},
		{
			name: "all false",
			predicates: []state.TransitionPredicate{
				state.KeyExists("missing1"),
				state.KeyExists("missing2"),
			},
			setState: func(s state.State) state.State {
				return s
			},
			want: false,
		},
		{
			name:       "empty predicates",
			predicates: []state.TransitionPredicate{},
			setState: func(s state.State) state.State {
				return s
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			predicate := state.Or(tt.predicates...)
			s := state.New(observability.NoOpObserver{})
			s = tt.setState(s)

			if got := predicate(s); got != tt.want {
				t.Errorf("Or() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPredicate_Composition(t *testing.T) {
	s := state.New(observability.NoOpObserver{})
	s = s.Set("status", "approved")
	s = s.Set("count", 5)

	complex := state.And(
		state.KeyEquals("status", "approved"),
		state.Or(
			state.KeyEquals("count", 5),
			state.KeyEquals("count", 10),
		),
	)

	if !complex(s) {
		t.Error("Complex predicate composition should evaluate to true")
	}

	s2 := state.New(observability.NoOpObserver{})
	s2 = s2.Set("status", "pending")
	s2 = s2.Set("count", 5)

	if complex(s2) {
		t.Error("Complex predicate composition should evaluate to false")
	}
}
