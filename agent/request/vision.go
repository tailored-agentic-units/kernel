package request

import (
	"github.com/tailored-agentic-units/kernel/core/model"
	"github.com/tailored-agentic-units/kernel/core/protocol"
	"github.com/tailored-agentic-units/kernel/agent/providers"
)

// VisionRequest represents a vision protocol request with image inputs.
// Separates images and vision-specific options from model configuration options.
type VisionRequest struct {
	messages      []protocol.Message
	images        []string       // URLs or base64 data URIs
	visionOptions map[string]any // Vision-specific options (e.g., detail: "high")
	options       map[string]any // Model configuration options
	provider      providers.Provider
	model         *model.Model
}

// NewVision creates a new VisionRequest with the given components.
// Messages contain the conversation history.
// Images are URLs or base64 data URIs to analyze.
// VisionOptions are vision-specific settings (e.g., detail level).
// Options specify model configuration (temperature, max_tokens, etc.).
func NewVision(p providers.Provider, m *model.Model, messages []protocol.Message, images []string, visionOpts, opts map[string]any) *VisionRequest {
	return &VisionRequest{
		messages:      messages,
		images:        images,
		visionOptions: visionOpts,
		options:       opts,
		provider:      p,
		model:         m,
	}
}

// Protocol returns the Vision protocol identifier.
func (r *VisionRequest) Protocol() protocol.Protocol {
	return protocol.Vision
}

// Headers returns the HTTP headers for a vision request.
func (r *VisionRequest) Headers() map[string]string {
	return map[string]string{
		"Content-Type": "application/json",
	}
}

// Marshal delegates to the provider for provider-specific JSON formatting.
func (r *VisionRequest) Marshal() ([]byte, error) {
	return r.provider.Marshal(protocol.Vision, &providers.VisionData{
		Model:         r.model.Name,
		Messages:      r.messages,
		Images:        r.images,
		VisionOptions: r.visionOptions,
		Options:       r.options,
	})
}

// Provider returns the provider for this request.
func (r *VisionRequest) Provider() providers.Provider {
	return r.provider
}

// Model returns the model for this request.
func (r *VisionRequest) Model() *model.Model {
	return r.model
}
