package response

import (
	"encoding/json"
	"fmt"
)

// EmbeddingsResponse represents the response from an embeddings protocol request.
// Contains vector embeddings for the input text along with metadata and token usage.
type EmbeddingsResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Embedding []float64 `json:"embedding"`
		Index     int       `json:"index"`
		Object    string    `json:"object"`
	}
	Model string      `json:"model"`
	Usage *TokenUsage `json:"usage,omitempty"`
}

// ParseEmbeddings parses an embeddings response from JSON bytes.
// Returns the parsed EmbeddingsResponse or an error if parsing fails.
func ParseEmbeddings(body []byte) (*EmbeddingsResponse, error) {
	var response EmbeddingsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse embeddings response: %w", err)
	}
	return &response, nil
}
