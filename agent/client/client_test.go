package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tailored-agentic-units/kernel/agent/client"
	"github.com/tailored-agentic-units/kernel/agent/providers"
	"github.com/tailored-agentic-units/kernel/agent/request"
	"github.com/tailored-agentic-units/kernel/core/config"
	"github.com/tailored-agentic-units/kernel/core/model"
	"github.com/tailored-agentic-units/kernel/core/protocol"
	"github.com/tailored-agentic-units/kernel/core/response"
)

func TestNew(t *testing.T) {
	cfg := &config.ClientConfig{
		Timeout:            config.Duration(30 * time.Second),
		ConnectionTimeout:  config.Duration(10 * time.Second),
		ConnectionPoolSize: 10,
	}

	c := client.New(cfg)

	if c == nil {
		t.Fatal("New returned nil client")
	}
}

func TestClient_Execute_Chat(t *testing.T) {
	// Create mock server that returns a valid chat response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		chatResp := response.ChatResponse{
			Model: "test-model",
		}
		chatResp.Choices = append(chatResp.Choices, struct {
			Index   int              `json:"index"`
			Message protocol.Message `json:"message"`
			Delta   *struct {
				Role    string `json:"role,omitempty"`
				Content string `json:"content,omitempty"`
			} `json:"delta,omitempty"`
			FinishReason string `json:"finish_reason,omitempty"`
		}{
			Index:   0,
			Message: protocol.NewMessage("assistant", "Hello, world!"),
		})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(chatResp)
	}))
	defer server.Close()

	// Create provider pointing to mock server
	providerCfg := &config.ProviderConfig{
		Name:    "ollama",
		BaseURL: server.URL,
	}
	provider, err := providers.NewOllama(providerCfg)
	if err != nil {
		t.Fatalf("NewOllama failed: %v", err)
	}

	// Create model
	mdl := model.New(&config.ModelConfig{
		Name: "test-model",
	})

	// Create client
	cfg := &config.ClientConfig{
		Timeout:            config.Duration(30 * time.Second),
		ConnectionTimeout:  config.Duration(10 * time.Second),
		ConnectionPoolSize: 10,
		Retry: config.RetryConfig{
			MaxRetries: 0, // Disable retry for this test
		},
	}
	c := client.New(cfg)

	// Create request
	messages := protocol.InitMessages("user", "Hello")
	req := request.NewChat(provider, mdl, messages, map[string]any{})

	// Execute
	result, err := c.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	chatResp, ok := result.(*response.ChatResponse)
	if !ok {
		t.Fatalf("expected *response.ChatResponse, got %T", result)
	}

	if chatResp.Content() != "Hello, world!" {
		t.Errorf("got content %q, want %q", chatResp.Content(), "Hello, world!")
	}
}

func TestClient_Execute_Tools(t *testing.T) {
	// Create mock server that returns a valid tools response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		toolsResp := response.ToolsResponse{
			Model: "test-model",
		}
		toolsResp.Choices = append(toolsResp.Choices, struct {
			Index   int `json:"index"`
			Message struct {
				Role      string              `json:"role"`
				Content   string              `json:"content"`
				ToolCalls []protocol.ToolCall `json:"tool_calls,omitempty"`
			} `json:"message"`
			FinishReason string `json:"finish_reason,omitempty"`
		}{
			Index: 0,
			Message: struct {
				Role      string              `json:"role"`
				Content   string              `json:"content"`
				ToolCalls []protocol.ToolCall `json:"tool_calls,omitempty"`
			}{
				Role:    "assistant",
				Content: "",
				ToolCalls: []protocol.ToolCall{
					{
						ID:        "call_123",
						Name:      "get_weather",
						Arguments: `{"location":"Boston"}`,
					},
				},
			},
		})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(toolsResp)
	}))
	defer server.Close()

	providerCfg := &config.ProviderConfig{
		Name:    "ollama",
		BaseURL: server.URL,
	}
	provider, err := providers.NewOllama(providerCfg)
	if err != nil {
		t.Fatalf("NewOllama failed: %v", err)
	}

	mdl := model.New(&config.ModelConfig{
		Name: "test-model",
	})

	cfg := &config.ClientConfig{
		Timeout:            config.Duration(30 * time.Second),
		ConnectionTimeout:  config.Duration(10 * time.Second),
		ConnectionPoolSize: 10,
		Retry: config.RetryConfig{
			MaxRetries: 0,
		},
	}
	c := client.New(cfg)

	messages := protocol.InitMessages("user", "What's the weather in Boston?")

	tools := []protocol.Tool{
		{
			Name:        "get_weather",
			Description: "Get current weather for a location",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"location": map[string]any{
						"type":        "string",
						"description": "City name",
					},
				},
			},
		},
	}

	req := request.NewTools(provider, mdl, messages, tools, map[string]any{})

	result, err := c.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	toolsResp, ok := result.(*response.ToolsResponse)
	if !ok {
		t.Fatalf("expected *response.ToolsResponse, got %T", result)
	}

	if len(toolsResp.Choices) == 0 {
		t.Fatal("no choices in response")
	}

	if len(toolsResp.Choices[0].Message.ToolCalls) == 0 {
		t.Fatal("no tool calls in response")
	}

	toolCall := toolsResp.Choices[0].Message.ToolCalls[0]
	if toolCall.Name != "get_weather" {
		t.Errorf("got function name %q, want %q", toolCall.Name, "get_weather")
	}
}

func TestClient_Execute_Embeddings(t *testing.T) {
	// Create mock server that returns a valid embeddings response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		embResp := response.EmbeddingsResponse{
			Object: "list",
			Model:  "test-model",
		}
		embResp.Data = append(embResp.Data, struct {
			Embedding []float64 `json:"embedding"`
			Index     int       `json:"index"`
			Object    string    `json:"object"`
		}{
			Embedding: []float64{0.1, 0.2, 0.3},
			Index:     0,
			Object:    "embedding",
		})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(embResp)
	}))
	defer server.Close()

	providerCfg := &config.ProviderConfig{
		Name:    "ollama",
		BaseURL: server.URL,
	}
	provider, err := providers.NewOllama(providerCfg)
	if err != nil {
		t.Fatalf("NewOllama failed: %v", err)
	}

	mdl := model.New(&config.ModelConfig{
		Name: "test-model",
	})

	cfg := &config.ClientConfig{
		Timeout:            config.Duration(30 * time.Second),
		ConnectionTimeout:  config.Duration(10 * time.Second),
		ConnectionPoolSize: 10,
		Retry: config.RetryConfig{
			MaxRetries: 0,
		},
	}
	c := client.New(cfg)

	req := request.NewEmbeddings(provider, mdl, "Hello, world!", map[string]any{})

	result, err := c.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	embResp, ok := result.(*response.EmbeddingsResponse)
	if !ok {
		t.Fatalf("expected *response.EmbeddingsResponse, got %T", result)
	}

	if len(embResp.Data) == 0 {
		t.Fatal("no embeddings in response")
	}

	if len(embResp.Data[0].Embedding) != 3 {
		t.Errorf("got %d dimensions, want 3", len(embResp.Data[0].Embedding))
	}
}

func TestClient_Execute_HTTPError(t *testing.T) {
	// Create mock server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	providerCfg := &config.ProviderConfig{
		Name:    "ollama",
		BaseURL: server.URL,
	}
	provider, err := providers.NewOllama(providerCfg)
	if err != nil {
		t.Fatalf("NewOllama failed: %v", err)
	}

	mdl := model.New(&config.ModelConfig{
		Name: "test-model",
	})

	cfg := &config.ClientConfig{
		Timeout:            config.Duration(30 * time.Second),
		ConnectionTimeout:  config.Duration(10 * time.Second),
		ConnectionPoolSize: 10,
		Retry: config.RetryConfig{
			MaxRetries: 0, // Disable retry to get immediate error
		},
	}
	c := client.New(cfg)

	messages := protocol.InitMessages("user", "Hello")
	req := request.NewChat(provider, mdl, messages, map[string]any{})

	_, err = c.Execute(context.Background(), req)
	if err == nil {
		t.Error("expected error for HTTP 500, got nil")
	}
}

func TestClient_ExecuteStream_UnsupportedProtocol(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	providerCfg := &config.ProviderConfig{
		Name:    "ollama",
		BaseURL: server.URL,
	}
	provider, err := providers.NewOllama(providerCfg)
	if err != nil {
		t.Fatalf("NewOllama failed: %v", err)
	}

	mdl := model.New(&config.ModelConfig{
		Name: "test-model",
	})

	cfg := &config.ClientConfig{
		Timeout:            config.Duration(30 * time.Second),
		ConnectionTimeout:  config.Duration(10 * time.Second),
		ConnectionPoolSize: 10,
	}
	c := client.New(cfg)

	// Embeddings does not support streaming
	req := request.NewEmbeddings(provider, mdl, "test", map[string]any{})

	_, err = c.ExecuteStream(context.Background(), req)
	if err == nil {
		t.Error("expected error for unsupported streaming protocol, got nil")
	}
}

func TestClient_IsHealthy(t *testing.T) {
	cfg := &config.ClientConfig{
		Timeout:            config.Duration(30 * time.Second),
		ConnectionTimeout:  config.Duration(10 * time.Second),
		ConnectionPoolSize: 10,
	}

	c := client.New(cfg)

	if !c.IsHealthy() {
		t.Error("expected client to be healthy initially")
	}
}

func TestClient_HTTPClient(t *testing.T) {
	cfg := &config.ClientConfig{
		Timeout:            config.Duration(5 * time.Second),
		ConnectionTimeout:  config.Duration(2 * time.Second),
		ConnectionPoolSize: 20,
	}

	c := client.New(cfg)

	httpClient := c.HTTPClient()

	if httpClient == nil {
		t.Fatal("HTTPClient() returned nil")
	}

	if httpClient.Timeout != 5*time.Second {
		t.Errorf("got timeout %v, want %v", httpClient.Timeout, 5*time.Second)
	}
}
