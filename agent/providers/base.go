package providers

import (
	"encoding/json"
	"fmt"
	"maps"

	"github.com/tailored-agentic-units/kernel/core/protocol"
)

// BaseProvider provides common functionality for provider implementations.
// It stores the provider name and base URL, and provides default OpenAI-compatible
// marshaling for all protocols.
// Provider implementations typically embed BaseProvider to inherit this functionality.
type BaseProvider struct {
	name    string
	baseURL string
}

// NewBaseProvider creates a new BaseProvider with the given name and base URL.
// This is typically called by provider constructors to initialize common fields.
func NewBaseProvider(name, baseURL string) *BaseProvider {
	return &BaseProvider{
		name:    name,
		baseURL: baseURL,
	}
}

// Name returns the provider's identifier.
func (p *BaseProvider) Name() string {
	return p.name
}

// BaseURL returns the provider's base URL.
// Provider implementations use this to construct full endpoint URLs.
func (p *BaseProvider) BaseURL() string {
	return p.baseURL
}

// Marshal converts request data to OpenAI-compatible JSON format.
// This default implementation works for OpenAI, Azure, and Ollama providers.
// Providers with different wire formats (Anthropic, Google) should override this method.
func (p *BaseProvider) Marshal(proto protocol.Protocol, data any) ([]byte, error) {
	switch proto {
	case protocol.Chat:
		return p.marshalChat(data)
	case protocol.Vision:
		return p.marshalVision(data)
	case protocol.Tools:
		return p.marshalTools(data)
	case protocol.Embeddings:
		return p.marshalEmbeddings(data)
	case protocol.Audio:
		return p.marshalAudio(data)
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", proto)
	}
}

func (p *BaseProvider) marshalChat(data any) ([]byte, error) {
	d, ok := data.(*ChatData)
	if !ok {
		return nil, fmt.Errorf("expected *ChatData, got %T", data)
	}

	combined := make(map[string]any)
	combined["model"] = d.Model
	combined["messages"] = d.Messages
	maps.Copy(combined, d.Options)
	return json.Marshal(combined)
}

func (p *BaseProvider) marshalVision(data any) ([]byte, error) {
	d, ok := data.(*VisionData)
	if !ok {
		return nil, fmt.Errorf("expected *VisionData, got %T", data)
	}

	if len(d.Messages) == 0 {
		return nil, fmt.Errorf("messages cannot be empty for vision requests")
	}

	if len(d.Images) == 0 {
		return nil, fmt.Errorf("images cannot be empty for vision requests")
	}

	// Transform the last message to embed images
	lastIdx := len(d.Messages) - 1
	message := d.Messages[lastIdx]

	var textContent string
	switch v := message.Content.(type) {
	case string:
		textContent = v
	default:
		return nil, fmt.Errorf("message content must be a string for vision transformation")
	}

	// Build structured content starting with text
	content := []map[string]any{
		{"type": "text", "text": textContent},
	}

	// Add each image with embedded options
	for _, imgURL := range d.Images {
		imageURL := map[string]any{
			"url": imgURL,
		}

		// Embed vision_options into image_url map
		if d.VisionOptions != nil {
			maps.Copy(imageURL, d.VisionOptions)
		}

		content = append(content, map[string]any{
			"type":      "image_url",
			"image_url": imageURL,
		})
	}

	// Create transformed messages
	transformedMessages := make([]protocol.Message, len(d.Messages))
	copy(transformedMessages, d.Messages)
	transformedMessages[lastIdx] = protocol.Message{
		Role:    message.Role,
		Content: content,
	}

	// Combine model, messages, and options at root level
	combined := make(map[string]any)
	combined["model"] = d.Model
	combined["messages"] = transformedMessages
	maps.Copy(combined, d.Options)

	return json.Marshal(combined)
}

func (p *BaseProvider) marshalTools(data any) ([]byte, error) {
	d, ok := data.(*ToolsData)
	if !ok {
		return nil, fmt.Errorf("expected *ToolsData, got %T", data)
	}

	combined := make(map[string]any)
	combined["model"] = d.Model
	combined["messages"] = d.Messages

	// Transform tools to OpenAI format: {"type": "function", "function": {...}}
	openAITools := make([]map[string]any, len(d.Tools))
	for i, tool := range d.Tools {
		openAITools[i] = map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        tool.Name,
				"description": tool.Description,
				"parameters":  tool.Parameters,
			},
		}
	}
	combined["tools"] = openAITools

	maps.Copy(combined, d.Options)
	return json.Marshal(combined)
}

func (p *BaseProvider) marshalEmbeddings(data any) ([]byte, error) {
	d, ok := data.(*EmbeddingsData)
	if !ok {
		return nil, fmt.Errorf("expected *EmbeddingsData, got %T", data)
	}

	combined := make(map[string]any)
	combined["model"] = d.Model
	combined["input"] = d.Input
	maps.Copy(combined, d.Options)
	return json.Marshal(combined)
}

func (p *BaseProvider) marshalAudio(data any) ([]byte, error) {
	d, ok := data.(*AudioData)
	if !ok {
		return nil, fmt.Errorf("expected *AudioData, got %T", data)
	}

	combined := make(map[string]any)
	combined["model"] = d.Model
	combined["input"] = d.Input
	if d.AudioOptions != nil {
		maps.Copy(combined, d.AudioOptions)
	}
	maps.Copy(combined, d.Options)
	return json.Marshal(combined)
}
