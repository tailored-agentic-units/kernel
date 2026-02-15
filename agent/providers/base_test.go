package providers_test

import (
	"encoding/json"
	"testing"

	"github.com/tailored-agentic-units/kernel/agent/providers"
	"github.com/tailored-agentic-units/kernel/core/protocol"
)

func TestNewBaseProvider(t *testing.T) {
	provider := providers.NewBaseProvider("test-provider", "https://api.example.com")

	if provider == nil {
		t.Fatal("NewBaseProvider returned nil")
	}

	if provider.Name() != "test-provider" {
		t.Errorf("got name %q, want %q", provider.Name(), "test-provider")
	}

	if provider.BaseURL() != "https://api.example.com" {
		t.Errorf("got baseURL %q, want %q", provider.BaseURL(), "https://api.example.com")
	}
}

func TestBaseProvider_Name(t *testing.T) {
	provider := providers.NewBaseProvider("my-provider", "https://api.test.com")

	if provider.Name() != "my-provider" {
		t.Errorf("got name %q, want %q", provider.Name(), "my-provider")
	}
}

func TestBaseProvider_BaseURL(t *testing.T) {
	provider := providers.NewBaseProvider("test", "https://custom.api.com/v2")

	if provider.BaseURL() != "https://custom.api.com/v2" {
		t.Errorf("got baseURL %q, want %q", provider.BaseURL(), "https://custom.api.com/v2")
	}
}

func TestBaseProvider_Marshal_Chat(t *testing.T) {
	provider := providers.NewBaseProvider("test", "https://api.test.com")

	chatData := &providers.ChatData{
		Model: "gpt-4",
		Messages: protocol.InitMessages("user", "Hello"),
		Options: map[string]any{
			"temperature": 0.7,
		},
	}

	body, err := provider.Marshal(protocol.Chat, chatData)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	if result["model"] != "gpt-4" {
		t.Errorf("got model %v, want gpt-4", result["model"])
	}

	if result["temperature"] != 0.7 {
		t.Errorf("got temperature %v, want 0.7", result["temperature"])
	}

	messages, ok := result["messages"].([]any)
	if !ok {
		t.Fatal("messages is not an array")
	}
	if len(messages) != 1 {
		t.Errorf("got %d messages, want 1", len(messages))
	}
}

func TestBaseProvider_Marshal_Vision(t *testing.T) {
	provider := providers.NewBaseProvider("test", "https://api.test.com")

	visionData := &providers.VisionData{
		Model: "gpt-4-vision",
		Messages: protocol.InitMessages("user", "What is in this image?"),
		Images: []string{"https://example.com/image.jpg"},
		Options: map[string]any{
			"max_tokens": 1024,
		},
	}

	body, err := provider.Marshal(protocol.Vision, visionData)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	if result["model"] != "gpt-4-vision" {
		t.Errorf("got model %v, want gpt-4-vision", result["model"])
	}

	if result["max_tokens"] != float64(1024) {
		t.Errorf("got max_tokens %v, want 1024", result["max_tokens"])
	}
}

func TestBaseProvider_Marshal_Tools(t *testing.T) {
	provider := providers.NewBaseProvider("test", "https://api.test.com")

	toolsData := &providers.ToolsData{
		Model: "gpt-4",
		Messages: protocol.InitMessages("user", "What's the weather?"),
		Tools: []protocol.Tool{
			{
				Name:        "get_weather",
				Description: "Get weather for a location",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"location": map[string]any{
							"type":        "string",
							"description": "The city name",
						},
					},
				},
			},
		},
		Options: map[string]any{},
	}

	body, err := provider.Marshal(protocol.Tools, toolsData)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	if result["model"] != "gpt-4" {
		t.Errorf("got model %v, want gpt-4", result["model"])
	}

	tools, ok := result["tools"].([]any)
	if !ok {
		t.Fatal("tools is not an array")
	}
	if len(tools) != 1 {
		t.Errorf("got %d tools, want 1", len(tools))
	}
}

func TestBaseProvider_Marshal_Embeddings(t *testing.T) {
	provider := providers.NewBaseProvider("test", "https://api.test.com")

	embeddingsData := &providers.EmbeddingsData{
		Model:   "text-embedding-ada-002",
		Input:   "Hello world",
		Options: map[string]any{},
	}

	body, err := provider.Marshal(protocol.Embeddings, embeddingsData)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	if result["model"] != "text-embedding-ada-002" {
		t.Errorf("got model %v, want text-embedding-ada-002", result["model"])
	}

	if result["input"] != "Hello world" {
		t.Errorf("got input %v, want 'Hello world'", result["input"])
	}
}

func TestBaseProvider_Marshal_Audio(t *testing.T) {
	provider := providers.NewBaseProvider("test", "https://api.test.com")

	audioData := &providers.AudioData{
		Model: "whisper-1",
		Input: "audio-input",
		AudioOptions: map[string]any{
			"language":        "en",
			"response_format": "verbose_json",
		},
		Options: map[string]any{
			"temperature": 0.2,
		},
	}

	body, err := provider.Marshal(protocol.Audio, audioData)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	if result["model"] != "whisper-1" {
		t.Errorf("got model %v, want whisper-1", result["model"])
	}

	if result["input"] != "audio-input" {
		t.Errorf("got input %v, want audio-input", result["input"])
	}

	if result["language"] != "en" {
		t.Errorf("got language %v, want en", result["language"])
	}

	if result["response_format"] != "verbose_json" {
		t.Errorf("got response_format %v, want verbose_json", result["response_format"])
	}

	if result["temperature"] != 0.2 {
		t.Errorf("got temperature %v, want 0.2", result["temperature"])
	}
}

func TestBaseProvider_Marshal_Audio_NilAudioOptions(t *testing.T) {
	provider := providers.NewBaseProvider("test", "https://api.test.com")

	audioData := &providers.AudioData{
		Model:   "whisper-1",
		Input:   "audio-input",
		Options: map[string]any{},
	}

	body, err := provider.Marshal(protocol.Audio, audioData)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	if result["model"] != "whisper-1" {
		t.Errorf("got model %v, want whisper-1", result["model"])
	}

	if result["input"] != "audio-input" {
		t.Errorf("got input %v, want audio-input", result["input"])
	}
}

func TestBaseProvider_Marshal_Audio_InvalidData(t *testing.T) {
	provider := providers.NewBaseProvider("test", "https://api.test.com")

	_, err := provider.Marshal(protocol.Audio, "invalid-data")
	if err == nil {
		t.Error("expected error for invalid data type, got nil")
	}
}

func TestBaseProvider_Marshal_UnsupportedProtocol(t *testing.T) {
	provider := providers.NewBaseProvider("test", "https://api.test.com")

	_, err := provider.Marshal(protocol.Protocol("unsupported"), nil)
	if err == nil {
		t.Error("expected error for unsupported protocol, got nil")
	}
}
