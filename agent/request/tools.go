package request

import (
	"github.com/tailored-agentic-units/kernel/core/model"
	"github.com/tailored-agentic-units/kernel/core/protocol"
	"github.com/tailored-agentic-units/kernel/agent/providers"
)

// ToolsRequest represents a tools (function calling) protocol request.
// Separates tool definitions (protocol input data) from model configuration options.
type ToolsRequest struct {
	messages []protocol.Message
	tools    []providers.ToolDefinition
	options  map[string]any
	provider providers.Provider
	model    *model.Model
}

// NewTools creates a new ToolsRequest with the given components.
// Messages contain the conversation history.
// Tools define the available functions the model can call.
// Options specify model configuration (temperature, max_tokens, etc.).
func NewTools(p providers.Provider, m *model.Model, messages []protocol.Message, tools []providers.ToolDefinition, opts map[string]any) *ToolsRequest {
	return &ToolsRequest{
		messages: messages,
		tools:    tools,
		options:  opts,
		provider: p,
		model:    m,
	}
}

// Protocol returns the Tools protocol identifier.
func (r *ToolsRequest) Protocol() protocol.Protocol {
	return protocol.Tools
}

// Headers returns the HTTP headers for a tools request.
func (r *ToolsRequest) Headers() map[string]string {
	return map[string]string{
		"Content-Type": "application/json",
	}
}

// Marshal delegates to the provider for provider-specific JSON formatting.
// Different providers use different tool formats (OpenAI, Anthropic, Google).
func (r *ToolsRequest) Marshal() ([]byte, error) {
	return r.provider.Marshal(protocol.Tools, &providers.ToolsData{
		Model:    r.model.Name,
		Messages: r.messages,
		Tools:    r.tools,
		Options:  r.options,
	})
}

// Provider returns the provider for this request.
func (r *ToolsRequest) Provider() providers.Provider {
	return r.provider
}

// Model returns the model for this request.
func (r *ToolsRequest) Model() *model.Model {
	return r.model
}
