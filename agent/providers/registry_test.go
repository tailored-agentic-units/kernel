package providers_test

import (
	"testing"

	"github.com/tailored-agentic-units/kernel/core/config"
	"github.com/tailored-agentic-units/kernel/agent/providers"
)

func TestCreate_Ollama(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "ollama",
		BaseURL: "http://localhost:11434",
	}

	provider, err := providers.Create(cfg)

	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if provider == nil {
		t.Fatal("Create returned nil provider")
	}

	if provider.Name() != "ollama" {
		t.Errorf("got name %q, want %q", provider.Name(), "ollama")
	}
}

func TestCreate_Azure(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "azure",
		BaseURL: "https://my-resource.openai.azure.com",
		Options: map[string]any{
			"deployment":  "gpt-4-deployment",
			"auth_type":   "api_key",
			"token":       "test-key",
			"api_version": "2024-02-01",
		},
	}

	provider, err := providers.Create(cfg)

	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if provider == nil {
		t.Fatal("Create returned nil provider")
	}

	if provider.Name() != "azure" {
		t.Errorf("got name %q, want %q", provider.Name(), "azure")
	}
}

func TestCreate_UnknownProvider(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "unknown-provider",
		BaseURL: "http://localhost",
	}

	_, err := providers.Create(cfg)

	if err == nil {
		t.Error("expected error for unknown provider, got nil")
	}
}

func TestListProviders(t *testing.T) {
	names := providers.ListProviders()

	if len(names) == 0 {
		t.Error("ListProviders returned empty list")
	}

	// Check for expected providers
	found := make(map[string]bool)
	for _, name := range names {
		found[name] = true
	}

	if !found["ollama"] {
		t.Error("ollama provider not registered")
	}

	if !found["azure"] {
		t.Error("azure provider not registered")
	}
}
