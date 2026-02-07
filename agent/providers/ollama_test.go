package providers_test

import (
	"context"
	"testing"

	"github.com/tailored-agentic-units/kernel/core/config"
	"github.com/tailored-agentic-units/kernel/core/protocol"
	"github.com/tailored-agentic-units/kernel/agent/providers"
)

func TestNewOllama(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "ollama",
		BaseURL: "http://localhost:11434",
	}

	provider, err := providers.NewOllama(cfg)

	if err != nil {
		t.Fatalf("NewOllama failed: %v", err)
	}

	if provider == nil {
		t.Fatal("NewOllama returned nil provider")
	}

	if provider.Name() != "ollama" {
		t.Errorf("got name %q, want %q", provider.Name(), "ollama")
	}
}

func TestNewOllama_URLSuffixHandling(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		expectedURL string
	}{
		{
			name:        "URL without /v1 suffix",
			baseURL:     "http://localhost:11434",
			expectedURL: "http://localhost:11434/v1/chat/completions",
		},
		{
			name:        "URL with /v1 suffix",
			baseURL:     "http://localhost:11434/v1",
			expectedURL: "http://localhost:11434/v1/chat/completions",
		},
		{
			name:        "URL with trailing slash",
			baseURL:     "http://localhost:11434/",
			expectedURL: "http://localhost:11434/v1/chat/completions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.ProviderConfig{
				Name:    "ollama",
				BaseURL: tt.baseURL,
			}

			provider, err := providers.NewOllama(cfg)
			if err != nil {
				t.Fatalf("NewOllama failed: %v", err)
			}

			endpoint, err := provider.Endpoint(protocol.Chat)
			if err != nil {
				t.Fatalf("Endpoint failed: %v", err)
			}

			if endpoint != tt.expectedURL {
				t.Errorf("got endpoint %q, want %q", endpoint, tt.expectedURL)
			}
		})
	}
}

func TestOllama_Endpoint(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "ollama",
		BaseURL: "http://localhost:11434",
	}

	provider, err := providers.NewOllama(cfg)
	if err != nil {
		t.Fatalf("NewOllama failed: %v", err)
	}

	tests := []struct {
		protocol protocol.Protocol
		expected string
	}{
		{
			protocol.Chat,
			"http://localhost:11434/v1/chat/completions",
		},
		{
			protocol.Vision,
			"http://localhost:11434/v1/chat/completions",
		},
		{
			protocol.Tools,
			"http://localhost:11434/v1/chat/completions",
		},
		{
			protocol.Embeddings,
			"http://localhost:11434/v1/embeddings",
		},
		{
			protocol.Audio,
			"http://localhost:11434/v1/audio/transcriptions",
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

func TestOllama_PrepareRequest(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "ollama",
		BaseURL: "http://localhost:11434",
	}

	provider, err := providers.NewOllama(cfg)
	if err != nil {
		t.Fatalf("NewOllama failed: %v", err)
	}

	// Marshal chat data using the provider
	chatData := &providers.ChatData{
		Model: "llama2",
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

	expectedURL := "http://localhost:11434/v1/chat/completions"
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

func TestOllama_PrepareStreamRequest(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "ollama",
		BaseURL: "http://localhost:11434",
	}

	provider, err := providers.NewOllama(cfg)
	if err != nil {
		t.Fatalf("NewOllama failed: %v", err)
	}

	chatData := &providers.ChatData{
		Model: "llama2",
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
