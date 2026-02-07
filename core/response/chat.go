package response

import (
	"encoding/json"
	"fmt"

	"github.com/tailored-agentic-units/kernel/core/protocol"
)

// ChatResponse represents the response from a non-streaming chat protocol request.
// Contains the model output, metadata, and optional token usage information.
type ChatResponse struct {
	ID      string `json:"id,omitempty"`
	Object  string `json:"object,omitempty"`
	Created int64  `json:"created,omitempty"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int              `json:"index"`
		Message protocol.Message `json:"message"`
		Delta   *struct {
			Role    string `json:"role,omitempty"`
			Content string `json:"content,omitempty"`
		} `json:"delta,omitempty"`
		FinishReason string `json:"finish_reason,omitempty"`
	} `json:"choices"`
	Usage *TokenUsage `json:"usage,omitempty"`
}

// Content extracts the text content from the first choice in the response.
// Handles both string content and structured content (e.g., vision responses).
// Returns empty string if there are no choices.
func (r *ChatResponse) Content() string {
	if len(r.Choices) > 0 {
		switch v := r.Choices[0].Message.Content.(type) {
		case string:
			return v
		default:
			return fmt.Sprintf("%v", v)
		}
	}
	return ""
}

// ParseChat parses a chat response from JSON bytes.
// Returns the parsed ChatResponse or an error if parsing fails.
func ParseChat(body []byte) (*ChatResponse, error) {
	var response ChatResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse chat response: %w", err)
	}
	return &response, nil
}

// ParseVision parses a vision response from JSON bytes.
// Vision protocol uses the same response format as chat.
func ParseVision(body []byte) (*ChatResponse, error) {
	return ParseChat(body)
}
