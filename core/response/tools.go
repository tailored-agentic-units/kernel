package response

import (
	"encoding/json"
	"fmt"

	"github.com/tailored-agentic-units/kernel/core/protocol"
)

// ToolsResponse represents the response from a tools (function calling) protocol request.
// Contains function calls requested by the model along with metadata and token usage.
type ToolsResponse struct {
	ID      string `json:"id,omitempty"`
	Object  string `json:"object,omitempty"`
	Created int64  `json:"created,omitempty"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role      string              `json:"role"`
			Content   string              `json:"content"`
			ToolCalls []protocol.ToolCall `json:"tool_calls,omitempty"`
		} `json:"message"`
		FinishReason string `json:"finish_reason,omitempty"`
	} `json:"choices"`
	Usage *TokenUsage `json:"usage,omitempty"`
}

// ParseTools parses a tools response from JSON bytes.
// Returns the parsed ToolsResponse or an error if parsing fails.
func ParseTools(body []byte) (*ToolsResponse, error) {
	var response ToolsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse tools response: %w", err)
	}
	return &response, nil
}
