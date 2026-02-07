package config_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/tailored-agentic-units/kernel/orchestrate/config"
)

func TestGraphConfig_DefaultGraphConfig(t *testing.T) {
	cfg := config.DefaultGraphConfig("test-graph")

	if cfg.Name != "test-graph" {
		t.Errorf("DefaultGraphConfig().Name = %v, want %v", cfg.Name, "test-graph")
	}
	if cfg.Observer != "slog" {
		t.Errorf("DefaultGraphConfig().Observer = %v, want %v", cfg.Observer, "slog")
	}
	if cfg.MaxIterations != 1000 {
		t.Errorf("DefaultGraphConfig().MaxIterations = %v, want %v", cfg.MaxIterations, 1000)
	}
}

func TestGraphConfig_JSONMarshaling(t *testing.T) {
	original := config.GraphConfig{
		Name:          "my-graph",
		Observer:      "slog",
		MaxIterations: 500,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled config.GraphConfig
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.Name != original.Name {
		t.Errorf("Unmarshaled Name = %v, want %v", unmarshaled.Name, original.Name)
	}
	if unmarshaled.Observer != original.Observer {
		t.Errorf("Unmarshaled Observer = %v, want %v", unmarshaled.Observer, original.Observer)
	}
	if unmarshaled.MaxIterations != original.MaxIterations {
		t.Errorf("Unmarshaled MaxIterations = %v, want %v",
			unmarshaled.MaxIterations, original.MaxIterations)
	}
}

func TestGraphConfig_JSONUnmarshalFromString(t *testing.T) {
	tests := []struct {
		name     string
		jsonStr  string
		wantName string
		wantObs  string
		wantIter int
	}{
		{
			name:     "complete config",
			jsonStr:  `{"name":"classify-doc-14-graph","observer":"open-telemetry","max_iterations":3}`,
			wantName: "classify-doc-14-graph",
			wantObs:  "open-telemetry",
			wantIter: 3,
		},
		{
			name:     "noop observer",
			jsonStr:  `{"name":"simple-graph","observer":"noop","max_iterations":1000}`,
			wantName: "simple-graph",
			wantObs:  "noop",
			wantIter: 1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg config.GraphConfig
			err := json.Unmarshal([]byte(tt.jsonStr), &cfg)
			if err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}

			if cfg.Name != tt.wantName {
				t.Errorf("Name = %v, want %v", cfg.Name, tt.wantName)
			}
			if cfg.Observer != tt.wantObs {
				t.Errorf("Observer = %v, want %v", cfg.Observer, tt.wantObs)
			}
			if cfg.MaxIterations != tt.wantIter {
				t.Errorf("MaxIterations = %v, want %v", cfg.MaxIterations, tt.wantIter)
			}
		})
	}
}

func TestGraphConfig_ObserverAsString(t *testing.T) {
	cfg := config.GraphConfig{
		Name:          "test",
		Observer:      "custom-observer",
		MaxIterations: 100,
	}

	if cfg.Observer != "custom-observer" {
		t.Errorf("Observer field should store string, got %v", cfg.Observer)
	}
}

func TestHubConfig_DefaultHubConfig(t *testing.T) {
	cfg := config.DefaultHubConfig()

	if cfg.Name != "default" {
		t.Errorf("DefaultHubConfig().Name = %v, want %v", cfg.Name, "default")
	}
	if cfg.ChannelBufferSize != 100 {
		t.Errorf("DefaultHubConfig().ChannelBufferSize = %v, want %v",
			cfg.ChannelBufferSize, 100)
	}
	if cfg.DefaultTimeout != 30*time.Second {
		t.Errorf("DefaultHubConfig().DefaultTimeout = %v, want %v",
			cfg.DefaultTimeout, 30*time.Second)
	}
	if cfg.Logger == nil {
		t.Error("DefaultHubConfig().Logger should not be nil")
	}
}
