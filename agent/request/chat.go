package request

import (
	"github.com/tailored-agentic-units/kernel/core/model"
	"github.com/tailored-agentic-units/kernel/core/protocol"
	"github.com/tailored-agentic-units/kernel/agent/providers"
)

// ChatRequest represents a chat protocol request.
// Encapsulates conversation messages, model configuration options,
// and the provider/model needed for execution.
type ChatRequest struct {
	messages []protocol.Message
	options  map[string]any
	provider providers.Provider
	model    *model.Model
}

// NewChat creates a new ChatRequest with the given components.
// Messages contain the conversation history.
// Options specify model configuration (temperature, max_tokens, etc.).
func NewChat(p providers.Provider, m *model.Model, messages []protocol.Message, opts map[string]any) *ChatRequest {
	return &ChatRequest{
		messages: messages,
		options:  opts,
		provider: p,
		model:    m,
	}
}

// Protocol returns the Chat protocol identifier.
func (r *ChatRequest) Protocol() protocol.Protocol {
	return protocol.Chat
}

// Headers returns the HTTP headers for a chat request.
func (r *ChatRequest) Headers() map[string]string {
	return map[string]string{
		"Content-Type": "application/json",
	}
}

// Marshal delegates to the provider for provider-specific JSON formatting.
func (r *ChatRequest) Marshal() ([]byte, error) {
	return r.provider.Marshal(protocol.Chat, &providers.ChatData{
		Model:    r.model.Name,
		Messages: r.messages,
		Options:  r.options,
	})
}

// Provider returns the provider for this request.
func (r *ChatRequest) Provider() providers.Provider {
	return r.provider
}

// Model returns the model for this request.
func (r *ChatRequest) Model() *model.Model {
	return r.model
}
