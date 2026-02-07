package agent

import (
	"context"
	"fmt"
	"maps"

	"github.com/google/uuid"
	"github.com/tailored-agentic-units/kernel/agent/client"
	"github.com/tailored-agentic-units/kernel/core/config"
	"github.com/tailored-agentic-units/kernel/core/model"
	"github.com/tailored-agentic-units/kernel/core/protocol"
	"github.com/tailored-agentic-units/kernel/agent/providers"
	"github.com/tailored-agentic-units/kernel/agent/request"
	"github.com/tailored-agentic-units/kernel/core/response"
)

// Agent provides a high-level interface for LLM interactions.
// Methods are protocol-specific and handle message initialization,
// system prompt injection, and response type assertions.
//
// Each agent has a unique identifier that remains stable across its lifetime.
// The ID is used for orchestration scenarios including hub registration,
// message routing, lifecycle tracking, and distributed tracing.
// IDs are guaranteed to be unique, stable, and thread-safe.
type Agent interface {
	// ID returns the unique identifier for the agent.
	// The ID is assigned at creation time using UUIDv7 and never changes.
	// Thread-safe for concurrent access and safe to use as map keys.
	ID() string

	// Client returns the underlying HTTP client.
	Client() client.Client

	// Provider returns the provider instance.
	Provider() providers.Provider

	// Model returns the model instance.
	Model() *model.Model

	// Chat executes a chat protocol request with optional system prompt injection.
	// Returns the parsed chat response or an error.
	Chat(ctx context.Context, prompt string, opts ...map[string]any) (*response.ChatResponse, error)

	// ChatStream executes a streaming chat protocol request.
	// Automatically sets stream: true in options.
	// Returns a channel of streaming chunks or an error.
	ChatStream(ctx context.Context, prompt string, opts ...map[string]any) (<-chan *response.StreamingChunk, error)

	// Vision executes a vision protocol request with images.
	// Images can be URLs or base64-encoded data URIs.
	// Returns the parsed chat response or an error.
	Vision(ctx context.Context, prompt string, images []string, opts ...map[string]any) (*response.ChatResponse, error)

	// VisionStream executes a streaming vision protocol request with images.
	// Returns a channel of streaming chunks or an error.
	VisionStream(ctx context.Context, prompt string, images []string, opts ...map[string]any) (<-chan *response.StreamingChunk, error)

	// Tools executes a tools protocol request with function definitions.
	// Returns the parsed tools response with tool calls or an error.
	Tools(ctx context.Context, prompt string, tools []Tool, opts ...map[string]any) (*response.ToolsResponse, error)

	// Embed executes an embeddings protocol request.
	// Returns the parsed embeddings response or an error.
	Embed(ctx context.Context, input string, opts ...map[string]any) (*response.EmbeddingsResponse, error)

	// Audio executes an audio transcription protocol request.
	// Returns the parsed audio response or an error.
	Audio(ctx context.Context, input string, opts ...map[string]any) (*response.AudioResponse, error)
}

// agent implements the Agent interface.
type agent struct {
	id           string
	client       client.Client
	provider     providers.Provider
	model        *model.Model
	systemPrompt string
}

// New creates a new Agent from configuration.
// Creates provider, model, and client from configuration.
// Assigns a unique UUIDv7 identifier for orchestration and tracking.
// Returns an error if provider creation fails.
func New(cfg *config.AgentConfig) (Agent, error) {
	p, err := providers.Create(cfg.Provider)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	m := model.New(cfg.Model)
	c := client.New(cfg.Client)

	return &agent{
		id:           uuid.Must(uuid.NewV7()).String(),
		client:       c,
		provider:     p,
		model:        m,
		systemPrompt: cfg.SystemPrompt,
	}, nil
}

func (a *agent) ID() string {
	return a.id
}

// Client returns the underlying HTTP client.
func (a *agent) Client() client.Client {
	return a.client
}

// Provider returns the provider instance.
func (a *agent) Provider() providers.Provider {
	return a.provider
}

// Model returns the model instance.
func (a *agent) Model() *model.Model {
	return a.model
}

// Chat executes a chat protocol request.
// Initializes messages with system prompt (if configured) and user prompt.
// Merges model's configured chat options with runtime opts.
// Returns parsed ChatResponse or error.
func (a *agent) Chat(ctx context.Context, prompt string, opts ...map[string]any) (*response.ChatResponse, error) {
	messages := a.initMessages(prompt)
	options := a.mergeOptions(protocol.Chat, opts...)

	req := request.NewChat(a.provider, a.model, messages, options)

	result, err := a.client.Execute(ctx, req)
	if err != nil {
		return nil, err
	}

	resp, ok := result.(*response.ChatResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", result)
	}

	return resp, nil
}

// ChatStream executes a streaming chat protocol request.
// Merges model's configured chat options with runtime opts.
// Automatically sets stream: true in options.
// Returns a channel of StreamingChunk or error.
func (a *agent) ChatStream(ctx context.Context, prompt string, opts ...map[string]any) (<-chan *response.StreamingChunk, error) {
	messages := a.initMessages(prompt)
	options := a.mergeOptions(protocol.Chat, opts...)
	options["stream"] = true

	req := request.NewChat(a.provider, a.model, messages, options)

	return a.client.ExecuteStream(ctx, req)
}

// Vision executes a vision protocol request with images.
// Images can be URLs or base64-encoded data URIs.
// Merges model's configured vision options with runtime opts.
// Extracts vision_options from opts if present, separating them from model options.
// Returns parsed ChatResponse or error.
func (a *agent) Vision(ctx context.Context, prompt string, images []string, opts ...map[string]any) (*response.ChatResponse, error) {
	messages := a.initMessages(prompt)
	options := a.mergeOptions(protocol.Vision, opts...)

	// Extract vision_options
	var visionOptions map[string]any
	if vOpts, exists := options["vision_options"]; exists {
		if vOptsMap, ok := vOpts.(map[string]any); ok {
			visionOptions = vOptsMap
			delete(options, "vision_options")
		}
	}

	req := request.NewVision(a.provider, a.model, messages, images, visionOptions, options)

	result, err := a.client.Execute(ctx, req)
	if err != nil {
		return nil, err
	}

	resp, ok := result.(*response.ChatResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", result)
	}

	return resp, nil
}

// VisionStream executes a streaming vision protocol request with images.
// Merges model's configured vision options with runtime opts.
// Extracts vision_options from opts if present, separating them from model options.
// Automatically sets stream: true in options.
// Returns a channel of StreamingChunk or error.
func (a *agent) VisionStream(ctx context.Context, prompt string, images []string, opts ...map[string]any) (<-chan *response.StreamingChunk, error) {
	messages := a.initMessages(prompt)
	options := a.mergeOptions(protocol.Vision, opts...)
	options["stream"] = true

	// Extract vision_options
	var visionOptions map[string]any
	if vOpts, exists := options["vision_options"]; exists {
		if vOptsMap, ok := vOpts.(map[string]any); ok {
			visionOptions = vOptsMap
			delete(options, "vision_options")
		}
	}

	req := request.NewVision(a.provider, a.model, messages, images, visionOptions, options)

	return a.client.ExecuteStream(ctx, req)
}

// Tools executes a tools protocol request with function definitions.
// Converts agent.Tool structs to providers.ToolDefinition format.
// Merges model's configured tools options with runtime opts.
// Returns parsed ToolsResponse with tool calls or error.
func (a *agent) Tools(ctx context.Context, prompt string, tools []Tool, opts ...map[string]any) (*response.ToolsResponse, error) {
	messages := a.initMessages(prompt)
	options := a.mergeOptions(protocol.Tools, opts...)

	// Convert agent.Tool to providers.ToolDefinition
	toolDefs := make([]providers.ToolDefinition, len(tools))
	for i, tool := range tools {
		toolDefs[i] = providers.ToolDefinition{
			Name:        tool.Name,
			Description: tool.Description,
			Parameters:  tool.Parameters,
		}
	}

	req := request.NewTools(a.provider, a.model, messages, toolDefs, options)

	result, err := a.client.Execute(ctx, req)
	if err != nil {
		return nil, err
	}

	resp, ok := result.(*response.ToolsResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", result)
	}

	return resp, nil
}

// Embed executes an embeddings protocol request.
// Merges model's configured embeddings options with runtime opts.
// Returns parsed EmbeddingsResponse or error.
func (a *agent) Embed(ctx context.Context, input string, opts ...map[string]any) (*response.EmbeddingsResponse, error) {
	options := a.mergeOptions(protocol.Embeddings, opts...)

	req := request.NewEmbeddings(a.provider, a.model, input, options)

	result, err := a.client.Execute(ctx, req)
	if err != nil {
		return nil, err
	}

	resp, ok := result.(*response.EmbeddingsResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", result)
	}

	return resp, nil
}

// Audio executes an audio transcription protocol request.
// Merges model's configured audio options with runtime opts.
// Extracts audio_options from opts if present, separating them from model options.
// Returns parsed AudioResponse or error.
func (a *agent) Audio(ctx context.Context, input string, opts ...map[string]any) (*response.AudioResponse, error) {
	options := a.mergeOptions(protocol.Audio, opts...)

	var audioOptions map[string]any
	if aOpts, exists := options["audio_options"]; exists {
		if aOptsMap, ok := aOpts.(map[string]any); ok {
			audioOptions = aOptsMap
			delete(options, "audio_options")
		}
	}

	req := request.NewAudio(a.provider, a.model, input, audioOptions, options)

	result, err := a.client.Execute(ctx, req)
	if err != nil {
		return nil, err
	}

	resp, ok := result.(*response.AudioResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", result)
	}

	return resp, nil
}

// mergeOptions creates options by merging model defaults with runtime options.
func (a *agent) mergeOptions(proto protocol.Protocol, opts ...map[string]any) map[string]any {
	options := make(map[string]any)
	if modelOpts := a.model.Options[proto]; modelOpts != nil {
		maps.Copy(options, modelOpts)
	}
	if len(opts) > 0 && opts[0] != nil {
		maps.Copy(options, opts[0])
	}
	return options
}

// initMessages creates the initial message list with optional system prompt.
// If system prompt is configured, it's added as the first message.
// User prompt is always added after system prompt.
func (a *agent) initMessages(prompt string) []protocol.Message {
	messages := make([]protocol.Message, 0)

	if a.systemPrompt != "" {
		messages = append(messages, protocol.NewMessage("system", a.systemPrompt))
	}

	messages = append(messages, protocol.NewMessage("user", prompt))

	return messages
}

// Tool defines a function that can be called by the LLM.
// Used with the Tools protocol for function calling capabilities.
type Tool struct {
	// Name is the function name that the LLM will call.
	Name string `json:"name"`

	// Description explains what the function does.
	// Should be clear and detailed to help the LLM decide when to use it.
	Description string `json:"description"`

	// Parameters is a JSON Schema defining the function's parameters.
	// Uses the format: {"type": "object", "properties": {...}, "required": [...]}
	Parameters map[string]any `json:"parameters"`
}
