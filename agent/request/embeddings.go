package request

import (
	"github.com/tailored-agentic-units/kernel/core/model"
	"github.com/tailored-agentic-units/kernel/core/protocol"
	"github.com/tailored-agentic-units/kernel/agent/providers"
)

// EmbeddingsRequest represents an embeddings protocol request.
// Separates input text (protocol data) from model configuration options.
// Does not use messages array - input is the primary data field.
type EmbeddingsRequest struct {
	input    any // string or []string for batch embeddings
	options  map[string]any
	provider providers.Provider
	model    *model.Model
}

// NewEmbeddings creates a new EmbeddingsRequest with the given components.
// Input is the text to embed (string or []string for batch).
// Options specify model configuration (encoding_format, dimensions, etc.).
func NewEmbeddings(p providers.Provider, m *model.Model, input any, opts map[string]any) *EmbeddingsRequest {
	return &EmbeddingsRequest{
		input:    input,
		options:  opts,
		provider: p,
		model:    m,
	}
}

// Protocol returns the Embeddings protocol identifier.
func (r *EmbeddingsRequest) Protocol() protocol.Protocol {
	return protocol.Embeddings
}

// Headers returns the HTTP headers for an embeddings request.
func (r *EmbeddingsRequest) Headers() map[string]string {
	return map[string]string{
		"Content-Type": "application/json",
	}
}

// Marshal delegates to the provider for provider-specific JSON formatting.
func (r *EmbeddingsRequest) Marshal() ([]byte, error) {
	return r.provider.Marshal(protocol.Embeddings, &providers.EmbeddingsData{
		Model:   r.model.Name,
		Input:   r.input,
		Options: r.options,
	})
}

// Provider returns the provider for this request.
func (r *EmbeddingsRequest) Provider() providers.Provider {
	return r.provider
}

// Model returns the model for this request.
func (r *EmbeddingsRequest) Model() *model.Model {
	return r.model
}
