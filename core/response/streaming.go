package response

import (
	"encoding/json"
	"fmt"
)

// StreamingChunk represents a single chunk from a streaming protocol response.
// Each chunk contains incremental content in the Delta field and metadata.
// The Error field can be set during streaming to indicate processing errors.
type StreamingChunk struct {
	ID      string `json:"id,omitempty"`
	Object  string `json:"object,omitempty"`
	Created int64  `json:"created,omitempty"`
	Model   string `json:"model"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Role    string `json:"role,omitempty"`
			Content string `json:"content,omitempty"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
	Error error `json:"-"`
}

// Content extracts the incremental content from the delta in the first choice.
// Returns empty string if there are no choices or no content in the delta.
func (c *StreamingChunk) Content() string {
	if len(c.Choices) > 0 {
		return c.Choices[0].Delta.Content
	}
	return ""
}

// ParseChatStreamChunk parses a streaming chat chunk from JSON bytes.
func ParseChatStreamChunk(data []byte) (*StreamingChunk, error) {
	var chunk StreamingChunk
	if err := json.Unmarshal(data, &chunk); err != nil {
		return nil, fmt.Errorf("failed to parse streaming chunk: %w", err)
	}
	return &chunk, nil
}

// ParseVisionStreamChunk parses a streaming vision chunk from JSON bytes.
// Vision protocol uses the same streaming format as chat.
func ParseVisionStreamChunk(data []byte) (*StreamingChunk, error) {
	return ParseChatStreamChunk(data)
}

// ParseToolsStreamChunk parses a streaming tools chunk from JSON bytes.
// Tools protocol uses the same streaming format as chat.
func ParseToolsStreamChunk(data []byte) (*StreamingChunk, error) {
	var chunk StreamingChunk
	if err := json.Unmarshal(data, &chunk); err != nil {
		return nil, fmt.Errorf("failed to parse tools streaming chunk: %w", err)
	}
	return &chunk, nil
}
