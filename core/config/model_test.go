package config_test

import (
	"encoding/json"
	"testing"

	"github.com/tailored-agentic-units/kernel/core/config"
)

func TestModelConfig_Unmarshal(t *testing.T) {
	jsonData := `{
		"name": "gpt-4",
		"capabilities": {
			"chat": {
				"temperature": 0.7,
				"max_tokens": 4096
			},
			"vision": {
				"detail": "auto",
				"max_tokens": 2048
			}
		}
	}`

	var cfg config.ModelConfig
	if err := json.Unmarshal([]byte(jsonData), &cfg); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if cfg.Name != "gpt-4" {
		t.Errorf("got name %s, want gpt-4", cfg.Name)
	}

	if len(cfg.Capabilities) != 2 {
		t.Errorf("got %d capabilities, want 2", len(cfg.Capabilities))
	}

	chatCap, exists := cfg.Capabilities["chat"]
	if !exists {
		t.Fatal("chat capability not found")
	}

	temp, exists := chatCap["temperature"]
	if !exists {
		t.Fatal("temperature option not found")
	}
	if temp != 0.7 {
		t.Errorf("got temperature %v, want 0.7", temp)
	}

	maxTokens, exists := chatCap["max_tokens"]
	if !exists {
		t.Fatal("max_tokens option not found")
	}
	// JSON numbers unmarshal as float64
	if maxTokens != float64(4096) {
		t.Errorf("got max_tokens %v, want 4096", maxTokens)
	}
}

func TestModelConfig_Capabilities(t *testing.T) {
	cfg := &config.ModelConfig{
		Name: "test-model",
		Capabilities: map[string]map[string]any{
			"chat": {
				"temperature": 0.7,
				"max_tokens":  4096,
			},
		},
	}

	if len(cfg.Capabilities) != 1 {
		t.Errorf("got %d capabilities, want 1", len(cfg.Capabilities))
	}

	chatCap, exists := cfg.Capabilities["chat"]
	if !exists {
		t.Fatal("chat capability not found")
	}

	temp, exists := chatCap["temperature"]
	if !exists {
		t.Fatal("temperature option not found")
	}
	if temp != 0.7 {
		t.Errorf("got temperature %v, want 0.7", temp)
	}
}

func TestDefaultModelConfig(t *testing.T) {
	cfg := config.DefaultModelConfig()

	if cfg == nil {
		t.Fatal("DefaultModelConfig returned nil")
	}

	if cfg.Capabilities == nil {
		t.Fatal("Capabilities map is nil")
	}

	if len(cfg.Capabilities) != 0 {
		t.Errorf("expected empty capabilities, got %d", len(cfg.Capabilities))
	}
}

func TestModelConfig_Merge(t *testing.T) {
	tests := []struct {
		name     string
		base     *config.ModelConfig
		source   *config.ModelConfig
		expected *config.ModelConfig
	}{
		{
			name: "merge name",
			base: &config.ModelConfig{
				Name: "base-model",
			},
			source: &config.ModelConfig{
				Name: "source-model",
			},
			expected: &config.ModelConfig{
				Name: "source-model",
			},
		},
		{
			name: "merge capabilities",
			base: &config.ModelConfig{
				Name: "test-model",
				Capabilities: map[string]map[string]any{
					"chat": {
						"temperature": 0.7,
					},
				},
			},
			source: &config.ModelConfig{
				Capabilities: map[string]map[string]any{
					"vision": {
						"detail": "auto",
					},
				},
			},
			expected: &config.ModelConfig{
				Name: "test-model",
				Capabilities: map[string]map[string]any{
					"chat": {
						"temperature": 0.7,
					},
					"vision": {
						"detail": "auto",
					},
				},
			},
		},
		{
			name: "source empty name preserves base",
			base: &config.ModelConfig{
				Name: "base-model",
			},
			source: &config.ModelConfig{
				Name: "",
			},
			expected: &config.ModelConfig{
				Name: "base-model",
			},
		},
		{
			name: "nil capabilities initialized",
			base: &config.ModelConfig{
				Name: "test-model",
			},
			source: &config.ModelConfig{
				Capabilities: map[string]map[string]any{
					"chat": {
						"temperature": 0.7,
					},
				},
			},
			expected: &config.ModelConfig{
				Name: "test-model",
				Capabilities: map[string]map[string]any{
					"chat": {
						"temperature": 0.7,
					},
				},
			},
		},
		{
			name: "merge overlapping protocol options",
			base: &config.ModelConfig{
				Name: "test-model",
				Capabilities: map[string]map[string]any{
					"chat": {
						"temperature": 0.7,
						"max_tokens":  4096,
					},
				},
			},
			source: &config.ModelConfig{
				Capabilities: map[string]map[string]any{
					"chat": {
						"temperature": 0.9,
					},
				},
			},
			expected: &config.ModelConfig{
				Name: "test-model",
				Capabilities: map[string]map[string]any{
					"chat": {
						"temperature": 0.9,
						"max_tokens":  4096,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.base.Merge(tt.source)

			if tt.base.Name != tt.expected.Name {
				t.Errorf("got name %s, want %s", tt.base.Name, tt.expected.Name)
			}

			if len(tt.base.Capabilities) != len(tt.expected.Capabilities) {
				t.Errorf("got %d capabilities, want %d", len(tt.base.Capabilities), len(tt.expected.Capabilities))
			}

			for protocol, expectedOpts := range tt.expected.Capabilities {
				baseOpts, exists := tt.base.Capabilities[protocol]
				if !exists {
					t.Errorf("protocol %s missing from result", protocol)
					continue
				}

				for key, expectedVal := range expectedOpts {
					baseVal, exists := baseOpts[key]
					if !exists {
						t.Errorf("protocol %s: option %s missing", protocol, key)
						continue
					}
					if baseVal != expectedVal {
						t.Errorf("protocol %s: option %s got %v, want %v", protocol, key, baseVal, expectedVal)
					}
				}
			}
		})
	}
}
