package providers

import "github.com/tailored-agentic-units/kernel/core/protocol"

// ChatData contains the data needed to marshal a chat request.
type ChatData struct {
	Model    string
	Messages []protocol.Message
	Options  map[string]any
}

// VisionData contains the data needed to marshal a vision request.
type VisionData struct {
	Model         string
	Messages      []protocol.Message
	Images        []string
	VisionOptions map[string]any
	Options       map[string]any
}

// ToolsData contains the data needed to marshal a tools request.
type ToolsData struct {
	Model    string
	Messages []protocol.Message
	Tools    []ToolDefinition
	Options  map[string]any
}

// ToolDefinition represents a provider-agnostic tool (function) definition.
// Providers transform this generic format to their specific API format
// (OpenAI, Anthropic, Google, etc.).
type ToolDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"` // JSON Schema
}

// EmbeddingsData contains the data needed to marshal an embeddings request.
type EmbeddingsData struct {
	Model   string
	Input   any // string or []string for batch embeddings
	Options map[string]any
}

// AudioData contains the data needed to marshal an audio transcription request.
type AudioData struct {
	Model        string
	Input        string
	AudioOptions map[string]any
	Options      map[string]any
}
