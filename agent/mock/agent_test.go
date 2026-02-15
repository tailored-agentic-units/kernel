package mock_test

import (
	"context"
	"testing"

	"github.com/tailored-agentic-units/kernel/agent/mock"
	"github.com/tailored-agentic-units/kernel/core/protocol"
	"github.com/tailored-agentic-units/kernel/core/response"
)

func TestNewMockAgent(t *testing.T) {
	agent := mock.NewMockAgent(
		mock.WithID("test-id"),
	)

	if agent == nil {
		t.Fatal("NewMockAgent returned nil")
	}

	if agent.ID() != "test-id" {
		t.Errorf("got ID %q, want %q", agent.ID(), "test-id")
	}
}

func TestMockAgent_Chat(t *testing.T) {
	expectedResponse := &response.ChatResponse{
		Model: "test-model",
	}
	expectedResponse.Choices = append(expectedResponse.Choices, struct {
		Index   int              `json:"index"`
		Message protocol.Message `json:"message"`
		Delta   *struct {
			Role    string `json:"role,omitempty"`
			Content string `json:"content,omitempty"`
		} `json:"delta,omitempty"`
		FinishReason string `json:"finish_reason,omitempty"`
	}{
		Index:   0,
		Message: protocol.NewMessage("assistant", "Hello"),
	})

	agent := mock.NewMockAgent(
		mock.WithID("test-id"),
		mock.WithChatResponse(expectedResponse, nil),
	)

	resp, err := agent.Chat(context.Background(), protocol.InitMessages(protocol.RoleUser, "test"))

	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	if resp != expectedResponse {
		t.Error("returned different response than configured")
	}
}

func TestMockAgent_Vision(t *testing.T) {
	expectedResponse := &response.ChatResponse{
		Model: "test-model",
	}
	expectedResponse.Choices = append(expectedResponse.Choices, struct {
		Index   int              `json:"index"`
		Message protocol.Message `json:"message"`
		Delta   *struct {
			Role    string `json:"role,omitempty"`
			Content string `json:"content,omitempty"`
		} `json:"delta,omitempty"`
		FinishReason string `json:"finish_reason,omitempty"`
	}{
		Index:   0,
		Message: protocol.NewMessage("assistant", "I see an image"),
	})

	agent := mock.NewMockAgent(
		mock.WithID("test-id"),
		mock.WithVisionResponse(expectedResponse, nil),
	)

	resp, err := agent.Vision(context.Background(), protocol.InitMessages(protocol.RoleUser, "test"), []string{"image.png"})

	if err != nil {
		t.Fatalf("Vision failed: %v", err)
	}

	if resp != expectedResponse {
		t.Error("returned different response than configured")
	}
}

func TestMockAgent_Tools(t *testing.T) {
	expectedResponse := &response.ToolsResponse{
		Model: "test-model",
	}
	expectedResponse.Choices = append(expectedResponse.Choices, struct {
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
					Name:      "test_func",
					Arguments: `{}`,
				},
			},
		},
	})

	agent := mock.NewMockAgent(
		mock.WithID("test-id"),
		mock.WithToolsResponse(expectedResponse, nil),
	)

	resp, err := agent.Tools(context.Background(), protocol.InitMessages(protocol.RoleUser, "test"), nil)

	if err != nil {
		t.Fatalf("Tools failed: %v", err)
	}

	if resp != expectedResponse {
		t.Error("returned different response than configured")
	}
}

func TestMockAgent_Embed(t *testing.T) {
	expectedResponse := &response.EmbeddingsResponse{
		Object: "list",
		Model:  "test-model",
	}
	expectedResponse.Data = append(expectedResponse.Data, struct {
		Embedding []float64 `json:"embedding"`
		Index     int       `json:"index"`
		Object    string    `json:"object"`
	}{
		Embedding: []float64{0.1, 0.2, 0.3},
		Index:     0,
		Object:    "embedding",
	})

	agent := mock.NewMockAgent(
		mock.WithID("test-id"),
		mock.WithEmbeddingsResponse(expectedResponse, nil),
	)

	resp, err := agent.Embed(context.Background(), "test")

	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}

	if resp != expectedResponse {
		t.Error("returned different response than configured")
	}
}

func TestMockAgent_Audio(t *testing.T) {
	expectedResponse := &response.AudioResponse{
		Task:     "transcribe",
		Language: "english",
		Duration: 3.14,
		Text:     "Hello, world!",
	}

	agent := mock.NewMockAgent(
		mock.WithID("test-id"),
		mock.WithAudioResponse(expectedResponse, nil),
	)

	resp, err := agent.Audio(context.Background(), "audio-input")

	if err != nil {
		t.Fatalf("Audio failed: %v", err)
	}

	if resp != expectedResponse {
		t.Error("returned different response than configured")
	}

	if resp.Content() != "Hello, world!" {
		t.Errorf("got content %q, want %q", resp.Content(), "Hello, world!")
	}
}

func TestNewAudioAgent(t *testing.T) {
	agent := mock.NewAudioAgent("test-id", "Transcribed text")

	resp, err := agent.Audio(context.Background(), "audio-input")

	if err != nil {
		t.Fatalf("Audio failed: %v", err)
	}

	if resp.Content() != "Transcribed text" {
		t.Errorf("got content %q, want %q", resp.Content(), "Transcribed text")
	}
}

func TestNewSimpleChatAgent(t *testing.T) {
	agent := mock.NewSimpleChatAgent("test-id", "Hello, world!")

	resp, err := agent.Chat(context.Background(), protocol.InitMessages(protocol.RoleUser, "test"))

	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	if resp.Content() != "Hello, world!" {
		t.Errorf("got content %q, want %q", resp.Content(), "Hello, world!")
	}
}

func TestNewStreamingChatAgent(t *testing.T) {
	agent := mock.NewStreamingChatAgent("test-id", []string{"Hello", ", ", "world!"})

	stream, err := agent.ChatStream(context.Background(), protocol.InitMessages(protocol.RoleUser, "test"))

	if err != nil {
		t.Fatalf("ChatStream failed: %v", err)
	}

	var content string
	for chunk := range stream {
		if chunk.Error != nil {
			t.Fatalf("Stream error: %v", chunk.Error)
		}
		content += chunk.Content()
	}

	if content != "Hello, world!" {
		t.Errorf("got content %q, want %q", content, "Hello, world!")
	}
}
