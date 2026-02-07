package response_test

import (
	"encoding/json"
	"testing"

	"github.com/tailored-agentic-units/kernel/core/response"
)

func TestChatResponse_Content_StringContent(t *testing.T) {
	jsonData := `{
		"model": "gpt-4",
		"choices": [{
			"index": 0,
			"message": {
				"role": "assistant",
				"content": "Hello, world!"
			}
		}]
	}`

	var resp response.ChatResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	content := resp.Content()
	if content != "Hello, world!" {
		t.Errorf("got content %q, want %q", content, "Hello, world!")
	}
}

func TestChatResponse_Content_EmptyChoices(t *testing.T) {
	jsonData := `{
		"model": "gpt-4",
		"choices": []
	}`

	var resp response.ChatResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	content := resp.Content()
	if content != "" {
		t.Errorf("got content %q, want empty string", content)
	}
}

func TestChatResponse_Unmarshal(t *testing.T) {
	jsonData := `{
		"id": "chatcmpl-123",
		"object": "chat.completion",
		"created": 1677652288,
		"model": "gpt-4",
		"choices": [{
			"index": 0,
			"message": {
				"role": "assistant",
				"content": "Hello there!"
			},
			"finish_reason": "stop"
		}],
		"usage": {
			"prompt_tokens": 9,
			"completion_tokens": 12,
			"total_tokens": 21
		}
	}`

	var resp response.ChatResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if resp.ID != "chatcmpl-123" {
		t.Errorf("got ID %q, want %q", resp.ID, "chatcmpl-123")
	}

	if resp.Model != "gpt-4" {
		t.Errorf("got model %q, want %q", resp.Model, "gpt-4")
	}

	if len(resp.Choices) != 1 {
		t.Fatalf("got %d choices, want 1", len(resp.Choices))
	}

	if resp.Content() != "Hello there!" {
		t.Errorf("got content %q, want %q", resp.Content(), "Hello there!")
	}

	if resp.Usage == nil {
		t.Fatal("usage is nil")
	}

	if resp.Usage.TotalTokens != 21 {
		t.Errorf("got total tokens %d, want 21", resp.Usage.TotalTokens)
	}
}

func TestStreamingChunk_Content(t *testing.T) {
	jsonData := `{
		"model": "gpt-4",
		"choices": [{
			"index": 0,
			"delta": {
				"content": "Hello"
			}
		}]
	}`

	var chunk response.StreamingChunk
	if err := json.Unmarshal([]byte(jsonData), &chunk); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	content := chunk.Content()
	if content != "Hello" {
		t.Errorf("got content %q, want %q", content, "Hello")
	}
}

func TestStreamingChunk_Content_EmptyChoices(t *testing.T) {
	jsonData := `{
		"model": "gpt-4",
		"choices": []
	}`

	var chunk response.StreamingChunk
	if err := json.Unmarshal([]byte(jsonData), &chunk); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	content := chunk.Content()
	if content != "" {
		t.Errorf("got content %q, want empty string", content)
	}
}

func TestStreamingChunk_Unmarshal(t *testing.T) {
	jsonData := `{
		"id": "chatcmpl-123",
		"object": "chat.completion.chunk",
		"created": 1677652288,
		"model": "gpt-4",
		"choices": [{
			"index": 0,
			"delta": {
				"content": "Hello"
			},
			"finish_reason": null
		}]
	}`

	var chunk response.StreamingChunk
	if err := json.Unmarshal([]byte(jsonData), &chunk); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if chunk.ID != "chatcmpl-123" {
		t.Errorf("got ID %q, want %q", chunk.ID, "chatcmpl-123")
	}

	if chunk.Model != "gpt-4" {
		t.Errorf("got model %q, want %q", chunk.Model, "gpt-4")
	}

	if chunk.Content() != "Hello" {
		t.Errorf("got content %q, want %q", chunk.Content(), "Hello")
	}
}

func TestEmbeddingsResponse_Unmarshal(t *testing.T) {
	jsonData := `{
		"object": "list",
		"data": [{
			"object": "embedding",
			"embedding": [0.1, 0.2, 0.3],
			"index": 0
		}],
		"model": "text-embedding-ada-002",
		"usage": {
			"prompt_tokens": 8,
			"total_tokens": 8
		}
	}`

	var resp response.EmbeddingsResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if resp.Object != "list" {
		t.Errorf("got object %q, want %q", resp.Object, "list")
	}

	if resp.Model != "text-embedding-ada-002" {
		t.Errorf("got model %q, want %q", resp.Model, "text-embedding-ada-002")
	}

	if len(resp.Data) != 1 {
		t.Fatalf("got %d data items, want 1", len(resp.Data))
	}

	if len(resp.Data[0].Embedding) != 3 {
		t.Fatalf("got %d embedding dimensions, want 3", len(resp.Data[0].Embedding))
	}

	if resp.Data[0].Embedding[0] != 0.1 {
		t.Errorf("got embedding[0] %f, want 0.1", resp.Data[0].Embedding[0])
	}
}

func TestToolsResponse_Unmarshal(t *testing.T) {
	jsonData := `{
		"id": "chatcmpl-123",
		"object": "chat.completion",
		"created": 1677652288,
		"model": "gpt-4",
		"choices": [{
			"index": 0,
			"message": {
				"role": "assistant",
				"content": "",
				"tool_calls": [{
					"id": "call_123",
					"type": "function",
					"function": {
						"name": "get_weather",
						"arguments": "{\"location\": \"Boston\"}"
					}
				}]
			},
			"finish_reason": "tool_calls"
		}]
	}`

	var resp response.ToolsResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if resp.ID != "chatcmpl-123" {
		t.Errorf("got ID %q, want %q", resp.ID, "chatcmpl-123")
	}

	if len(resp.Choices) != 1 {
		t.Fatalf("got %d choices, want 1", len(resp.Choices))
	}

	if len(resp.Choices[0].Message.ToolCalls) != 1 {
		t.Fatalf("got %d tool calls, want 1", len(resp.Choices[0].Message.ToolCalls))
	}

	toolCall := resp.Choices[0].Message.ToolCalls[0]

	if toolCall.ID != "call_123" {
		t.Errorf("got tool call ID %q, want %q", toolCall.ID, "call_123")
	}

	if toolCall.Function.Name != "get_weather" {
		t.Errorf("got function name %q, want %q", toolCall.Function.Name, "get_weather")
	}
}

func TestParseChat(t *testing.T) {
	jsonData := []byte(`{
		"model": "gpt-4",
		"choices": [{
			"index": 0,
			"message": {
				"role": "assistant",
				"content": "Hello!"
			}
		}]
	}`)

	resp, err := response.ParseChat(jsonData)
	if err != nil {
		t.Fatalf("ParseChat failed: %v", err)
	}

	if resp.Content() != "Hello!" {
		t.Errorf("got content %q, want %q", resp.Content(), "Hello!")
	}
}

func TestParseChat_InvalidJSON(t *testing.T) {
	jsonData := []byte(`{invalid json}`)

	_, err := response.ParseChat(jsonData)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestAudioResponse_Unmarshal(t *testing.T) {
	jsonData := `{
		"task": "transcribe",
		"language": "english",
		"duration": 5.42,
		"text": "Hello, world!"
	}`

	var resp response.AudioResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if resp.Task != "transcribe" {
		t.Errorf("got task %q, want %q", resp.Task, "transcribe")
	}

	if resp.Language != "english" {
		t.Errorf("got language %q, want %q", resp.Language, "english")
	}

	if resp.Duration != 5.42 {
		t.Errorf("got duration %f, want 5.42", resp.Duration)
	}

	if resp.Text != "Hello, world!" {
		t.Errorf("got text %q, want %q", resp.Text, "Hello, world!")
	}
}

func TestAudioResponse_Unmarshal_Verbose(t *testing.T) {
	jsonData := `{
		"task": "transcribe",
		"language": "english",
		"duration": 8.47,
		"text": "Hello there.",
		"words": [
			{"word": "Hello", "start": 0.0, "end": 0.52},
			{"word": "there.", "start": 0.52, "end": 1.04}
		],
		"segments": [
			{"id": 0, "start": 0.0, "end": 1.04, "text": "Hello there."}
		]
	}`

	var resp response.AudioResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if len(resp.Words) != 2 {
		t.Fatalf("got %d words, want 2", len(resp.Words))
	}

	if resp.Words[0].Word != "Hello" {
		t.Errorf("got word %q, want %q", resp.Words[0].Word, "Hello")
	}

	if resp.Words[0].End != 0.52 {
		t.Errorf("got word end %f, want 0.52", resp.Words[0].End)
	}

	if len(resp.Segments) != 1 {
		t.Fatalf("got %d segments, want 1", len(resp.Segments))
	}

	if resp.Segments[0].Text != "Hello there." {
		t.Errorf("got segment text %q, want %q", resp.Segments[0].Text, "Hello there.")
	}
}

func TestAudioResponse_Content(t *testing.T) {
	resp := response.AudioResponse{
		Text: "Transcribed text content",
	}

	if resp.Content() != "Transcribed text content" {
		t.Errorf("got content %q, want %q", resp.Content(), "Transcribed text content")
	}
}

func TestParseAudio(t *testing.T) {
	jsonData := []byte(`{
		"task": "transcribe",
		"language": "english",
		"duration": 3.14,
		"text": "Test transcription."
	}`)

	resp, err := response.ParseAudio(jsonData)
	if err != nil {
		t.Fatalf("ParseAudio failed: %v", err)
	}

	if resp.Text != "Test transcription." {
		t.Errorf("got text %q, want %q", resp.Text, "Test transcription.")
	}

	if resp.Content() != "Test transcription." {
		t.Errorf("got content %q, want %q", resp.Content(), "Test transcription.")
	}
}

func TestParseAudio_InvalidJSON(t *testing.T) {
	jsonData := []byte(`{invalid json}`)

	_, err := response.ParseAudio(jsonData)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}
