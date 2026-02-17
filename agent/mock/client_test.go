package mock_test

import (
	"context"
	"testing"

	"github.com/tailored-agentic-units/kernel/agent/mock"
	"github.com/tailored-agentic-units/kernel/core/protocol"
	"github.com/tailored-agentic-units/kernel/core/response"
)

func TestNewMockClient(t *testing.T) {
	client := mock.NewMockClient()

	if client == nil {
		t.Fatal("NewMockClient returned nil")
	}
}

func TestMockClient_Execute(t *testing.T) {
	expectedResponse := &response.ChatResponse{
		Model: "test-model",
	}

	client := mock.NewMockClient(
		mock.WithExecuteResponse(expectedResponse, nil),
	)

	// MockClient.Execute takes a request.Request, but we can test
	// with the mock's configured response
	result, err := client.Execute(context.Background(), nil)

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result != expectedResponse {
		t.Error("returned different response than configured")
	}
}

func TestMockClient_ExecuteStream(t *testing.T) {
	// Create properly typed StreamingChunk
	chunk := &response.StreamingChunk{
		Model: "test-model",
	}
	chunk.Choices = make([]struct {
		Index int `json:"index"`
		Delta struct {
			Role      string              `json:"role,omitempty"`
			Content   string              `json:"content,omitempty"`
			ToolCalls []protocol.ToolCall `json:"tool_calls,omitempty"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	}, 1)
	chunk.Choices[0].Delta.Content = "Hello"

	chunks := []*response.StreamingChunk{chunk}

	client := mock.NewMockClient(
		mock.WithStreamResponse(chunks, nil),
	)

	stream, err := client.ExecuteStream(context.Background(), nil)

	if err != nil {
		t.Fatalf("ExecuteStream failed: %v", err)
	}

	count := 0
	for chunk := range stream {
		if chunk.Error != nil {
			t.Fatalf("Stream error: %v", chunk.Error)
		}
		count++
	}

	if count != len(chunks) {
		t.Errorf("got %d chunks, want %d", count, len(chunks))
	}
}

func TestMockClient_IsHealthy(t *testing.T) {
	tests := []struct {
		name     string
		healthy  bool
		expected bool
	}{
		{
			name:     "healthy",
			healthy:  true,
			expected: true,
		},
		{
			name:     "unhealthy",
			healthy:  false,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := mock.NewMockClient(
				mock.WithHealthy(tt.healthy),
			)

			if client.IsHealthy() != tt.expected {
				t.Errorf("got IsHealthy() = %v, want %v", client.IsHealthy(), tt.expected)
			}
		})
	}
}
