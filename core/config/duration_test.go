package config_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/tailored-agentic-units/kernel/core/config"
)

func TestDuration_UnmarshalJSON_ParsesStringFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Duration
	}{
		{
			name:     "seconds",
			input:    `"24s"`,
			expected: 24 * time.Second,
		},
		{
			name:     "minutes",
			input:    `"1m"`,
			expected: 1 * time.Minute,
		},
		{
			name:     "hours",
			input:    `"2h"`,
			expected: 2 * time.Hour,
		},
		{
			name:     "composite",
			input:    `"1h30m"`,
			expected: 90 * time.Minute,
		},
		{
			name:     "milliseconds",
			input:    `"500ms"`,
			expected: 500 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d config.Duration
			if err := json.Unmarshal([]byte(tt.input), &d); err != nil {
				t.Fatalf("UnmarshalJSON failed: %v", err)
			}

			if d.ToDuration() != tt.expected {
				t.Errorf("got duration %v, want %v", d.ToDuration(), tt.expected)
			}
		})
	}
}

func TestDuration_UnmarshalJSON_ParsesNumericNanoseconds(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Duration
	}{
		{
			name:     "zero",
			input:    `0`,
			expected: 0,
		},
		{
			name:     "nanoseconds",
			input:    `1000000000`,
			expected: 1 * time.Second,
		},
		{
			name:     "minutes in nanoseconds",
			input:    `60000000000`,
			expected: 1 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d config.Duration
			if err := json.Unmarshal([]byte(tt.input), &d); err != nil {
				t.Fatalf("UnmarshalJSON failed: %v", err)
			}

			if d.ToDuration() != tt.expected {
				t.Errorf("got duration %v, want %v", d.ToDuration(), tt.expected)
			}
		})
	}
}

func TestDuration_UnmarshalJSON_InvalidFormat(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "invalid string",
			input: `"invalid"`,
		},
		{
			name:  "empty string",
			input: `""`,
		},
		{
			name:  "invalid json",
			input: `{invalid}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d config.Duration
			if err := json.Unmarshal([]byte(tt.input), &d); err == nil {
				t.Errorf("expected error for input %s, got nil", tt.input)
			}
		})
	}
}

func TestDuration_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		duration config.Duration
		expected string
	}{
		{
			name:     "seconds",
			duration: config.Duration(24 * time.Second),
			expected: `"24s"`,
		},
		{
			name:     "minutes",
			duration: config.Duration(1 * time.Minute),
			expected: `"1m0s"`,
		},
		{
			name:     "hours",
			duration: config.Duration(2 * time.Hour),
			expected: `"2h0m0s"`,
		},
		{
			name:     "zero",
			duration: config.Duration(0),
			expected: `"0s"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.duration)
			if err != nil {
				t.Fatalf("MarshalJSON failed: %v", err)
			}

			if string(data) != tt.expected {
				t.Errorf("got %s, want %s", string(data), tt.expected)
			}
		})
	}
}

func TestDuration_ToDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration config.Duration
		expected time.Duration
	}{
		{
			name:     "zero",
			duration: config.Duration(0),
			expected: 0,
		},
		{
			name:     "seconds",
			duration: config.Duration(24 * time.Second),
			expected: 24 * time.Second,
		},
		{
			name:     "minutes",
			duration: config.Duration(5 * time.Minute),
			expected: 5 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.duration.ToDuration()
			if result != tt.expected {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDuration_RoundTrip(t *testing.T) {
	original := config.Duration(90 * time.Minute)

	// Marshal to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	// Unmarshal back
	var restored config.Duration
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	if restored.ToDuration() != original.ToDuration() {
		t.Errorf("round trip failed: got %v, want %v", restored, original)
	}
}
