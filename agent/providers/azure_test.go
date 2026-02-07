package providers_test

import (
	"context"
	"testing"

	"github.com/tailored-agentic-units/kernel/core/config"
	"github.com/tailored-agentic-units/kernel/core/protocol"
	"github.com/tailored-agentic-units/kernel/agent/providers"
)

func TestNewAzure(t *testing.T) {
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

	provider, err := providers.NewAzure(cfg)

	if err != nil {
		t.Fatalf("NewAzure failed: %v", err)
	}

	if provider == nil {
		t.Fatal("NewAzure returned nil provider")
	}

	if provider.Name() != "azure" {
		t.Errorf("got name %q, want %q", provider.Name(), "azure")
	}
}

func TestNewAzure_MissingDeployment(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "azure",
		BaseURL: "https://my-resource.openai.azure.com",
		Options: map[string]any{
			"auth_type":   "api_key",
			"token":       "test-key",
			"api_version": "2024-02-01",
		},
	}

	_, err := providers.NewAzure(cfg)

	if err == nil {
		t.Error("expected error for missing deployment, got nil")
	}
}

func TestNewAzure_MissingAuthType(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "azure",
		BaseURL: "https://my-resource.openai.azure.com",
		Options: map[string]any{
			"deployment":  "gpt-4-deployment",
			"token":       "test-key",
			"api_version": "2024-02-01",
		},
	}

	_, err := providers.NewAzure(cfg)

	if err == nil {
		t.Error("expected error for missing auth_type, got nil")
	}
}

func TestNewAzure_MissingToken(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "azure",
		BaseURL: "https://my-resource.openai.azure.com",
		Options: map[string]any{
			"deployment":  "gpt-4-deployment",
			"auth_type":   "api_key",
			"api_version": "2024-02-01",
		},
	}

	_, err := providers.NewAzure(cfg)

	if err == nil {
		t.Error("expected error for missing token, got nil")
	}
}

func TestNewAzure_MissingAPIVersion(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "azure",
		BaseURL: "https://my-resource.openai.azure.com",
		Options: map[string]any{
			"deployment": "gpt-4-deployment",
			"auth_type":  "api_key",
			"token":      "test-key",
		},
	}

	_, err := providers.NewAzure(cfg)

	if err == nil {
		t.Error("expected error for missing api_version, got nil")
	}
}

func TestAzure_Endpoint(t *testing.T) {
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

	provider, err := providers.NewAzure(cfg)
	if err != nil {
		t.Fatalf("NewAzure failed: %v", err)
	}

	tests := []struct {
		protocol protocol.Protocol
		expected string
	}{
		{
			protocol.Chat,
			"https://my-resource.openai.azure.com/deployments/gpt-4-deployment/chat/completions?api-version=2024-02-01",
		},
		{
			protocol.Vision,
			"https://my-resource.openai.azure.com/deployments/gpt-4-deployment/chat/completions?api-version=2024-02-01",
		},
		{
			protocol.Tools,
			"https://my-resource.openai.azure.com/deployments/gpt-4-deployment/chat/completions?api-version=2024-02-01",
		},
		{
			protocol.Embeddings,
			"https://my-resource.openai.azure.com/deployments/gpt-4-deployment/embeddings?api-version=2024-02-01",
		},
		{
			protocol.Audio,
			"https://my-resource.openai.azure.com/deployments/gpt-4-deployment/audio/transcriptions?api-version=2024-02-01",
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.protocol), func(t *testing.T) {
			endpoint, err := provider.Endpoint(tt.protocol)

			if err != nil {
				t.Fatalf("Endpoint failed: %v", err)
			}

			if endpoint != tt.expected {
				t.Errorf("got endpoint %q, want %q", endpoint, tt.expected)
			}
		})
	}
}

func TestAzure_PrepareRequest(t *testing.T) {
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

	provider, err := providers.NewAzure(cfg)
	if err != nil {
		t.Fatalf("NewAzure failed: %v", err)
	}

	chatData := &providers.ChatData{
		Model: "gpt-4",
		Messages: []protocol.Message{
			protocol.NewMessage("user", "Hello"),
		},
		Options: map[string]any{},
	}

	body, err := provider.Marshal(protocol.Chat, chatData)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	request, err := provider.PrepareRequest(context.Background(), protocol.Chat, body, headers)

	if err != nil {
		t.Fatalf("PrepareRequest failed: %v", err)
	}

	if request == nil {
		t.Fatal("PrepareRequest returned nil request")
	}

	expectedURL := "https://my-resource.openai.azure.com/deployments/gpt-4-deployment/chat/completions?api-version=2024-02-01"
	if request.URL != expectedURL {
		t.Errorf("got URL %q, want %q", request.URL, expectedURL)
	}

	if len(request.Body) == 0 {
		t.Error("request body is empty")
	}

	if request.Headers == nil {
		t.Error("request headers is nil")
	}
}

func TestAzure_PrepareStreamRequest(t *testing.T) {
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

	provider, err := providers.NewAzure(cfg)
	if err != nil {
		t.Fatalf("NewAzure failed: %v", err)
	}

	chatData := &providers.ChatData{
		Model: "gpt-4",
		Messages: []protocol.Message{
			protocol.NewMessage("user", "Hello"),
		},
		Options: map[string]any{"stream": true},
	}

	body, err := provider.Marshal(protocol.Chat, chatData)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	request, err := provider.PrepareStreamRequest(context.Background(), protocol.Chat, body, headers)

	if err != nil {
		t.Fatalf("PrepareStreamRequest failed: %v", err)
	}

	if request == nil {
		t.Fatal("PrepareStreamRequest returned nil request")
	}

	if request.Headers["Accept"] != "text/event-stream" {
		t.Errorf("got Accept header %q, want %q", request.Headers["Accept"], "text/event-stream")
	}

	if request.Headers["Cache-Control"] != "no-cache" {
		t.Errorf("got Cache-Control header %q, want %q", request.Headers["Cache-Control"], "no-cache")
	}
}
