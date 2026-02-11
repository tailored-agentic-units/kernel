package mock

import (
	"context"

	"github.com/tailored-agentic-units/kernel/agent"
	"github.com/tailored-agentic-units/kernel/agent/client"
	"github.com/tailored-agentic-units/kernel/agent/providers"
	"github.com/tailored-agentic-units/kernel/core/model"
	"github.com/tailored-agentic-units/kernel/core/protocol"
	"github.com/tailored-agentic-units/kernel/core/response"
)

// MockAgent implements agent.Agent interface for testing.
// All methods return predetermined responses configured during construction.
type MockAgent struct {
	id string

	// Protocol responses
	chatResponse       *response.ChatResponse
	chatError          error
	visionResponse     *response.ChatResponse
	visionError        error
	toolsResponse      *response.ToolsResponse
	toolsError         error
	embeddingsResponse *response.EmbeddingsResponse
	embeddingsError    error
	audioResponse      *response.AudioResponse
	audioError         error

	// Streaming responses
	streamChunks []response.StreamingChunk
	streamError  error

	// Dependencies
	mockClient   client.Client
	mockProvider providers.Provider
	mockModel    *model.Model
}

// NewMockAgent creates a new MockAgent with default configuration.
// Use option functions to configure specific behaviors.
func NewMockAgent(opts ...MockAgentOption) *MockAgent {
	m := &MockAgent{
		id:           "mock-agent-id",
		mockClient:   NewMockClient(),
		mockProvider: NewMockProvider(),
		mockModel: &model.Model{
			Name:    "mock-model",
			Options: make(map[protocol.Protocol]map[string]any),
		},
		streamChunks: []response.StreamingChunk{},
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

// MockAgentOption configures a MockAgent.
type MockAgentOption func(*MockAgent)

// WithID sets the agent ID.
func WithID(id string) MockAgentOption {
	return func(m *MockAgent) {
		m.id = id
	}
}

// WithChatResponse sets the chat response and error.
func WithChatResponse(resp *response.ChatResponse, err error) MockAgentOption {
	return func(m *MockAgent) {
		m.chatResponse = resp
		m.chatError = err
	}
}

// WithVisionResponse sets the vision response and error.
func WithVisionResponse(resp *response.ChatResponse, err error) MockAgentOption {
	return func(m *MockAgent) {
		m.visionResponse = resp
		m.visionError = err
	}
}

// WithToolsResponse sets the tools response and error.
func WithToolsResponse(resp *response.ToolsResponse, err error) MockAgentOption {
	return func(m *MockAgent) {
		m.toolsResponse = resp
		m.toolsError = err
	}
}

// WithEmbeddingsResponse sets the embeddings response and error.
func WithEmbeddingsResponse(resp *response.EmbeddingsResponse, err error) MockAgentOption {
	return func(m *MockAgent) {
		m.embeddingsResponse = resp
		m.embeddingsError = err
	}
}

// WithAudioResponse sets the audio response and error.
func WithAudioResponse(resp *response.AudioResponse, err error) MockAgentOption {
	return func(m *MockAgent) {
		m.audioResponse = resp
		m.audioError = err
	}
}

// WithStreamChunks sets the streaming chunks for stream methods.
func WithStreamChunks(chunks []response.StreamingChunk, err error) MockAgentOption {
	return func(m *MockAgent) {
		m.streamChunks = chunks
		m.streamError = err
	}
}

// WithClient sets a custom client.
func WithClient(c client.Client) MockAgentOption {
	return func(m *MockAgent) {
		m.mockClient = c
	}
}

// WithProvider sets a custom provider.
func WithProvider(p providers.Provider) MockAgentOption {
	return func(m *MockAgent) {
		m.mockProvider = p
	}
}

// WithModel sets a custom model.
func WithModel(mdl *model.Model) MockAgentOption {
	return func(m *MockAgent) {
		m.mockModel = mdl
	}
}

// ID returns the mock agent's unique identifier.
func (m *MockAgent) ID() string {
	return m.id
}

// Client returns the mock client.
func (m *MockAgent) Client() client.Client {
	return m.mockClient
}

// Provider returns the mock provider.
func (m *MockAgent) Provider() providers.Provider {
	return m.mockProvider
}

// Model returns the mock model.
func (m *MockAgent) Model() *model.Model {
	return m.mockModel
}

// Chat returns the predetermined chat response.
func (m *MockAgent) Chat(ctx context.Context, prompt string, opts ...map[string]any) (*response.ChatResponse, error) {
	return m.chatResponse, m.chatError
}

// ChatStream returns a channel with predetermined streaming chunks.
func (m *MockAgent) ChatStream(ctx context.Context, prompt string, opts ...map[string]any) (<-chan *response.StreamingChunk, error) {
	if m.streamError != nil {
		return nil, m.streamError
	}

	ch := make(chan *response.StreamingChunk, len(m.streamChunks))
	for i := range m.streamChunks {
		ch <- &m.streamChunks[i]
	}
	close(ch)

	return ch, nil
}

// Vision returns the predetermined vision response.
func (m *MockAgent) Vision(ctx context.Context, prompt string, images []string, opts ...map[string]any) (*response.ChatResponse, error) {
	return m.visionResponse, m.visionError
}

// VisionStream returns a channel with predetermined streaming chunks.
func (m *MockAgent) VisionStream(ctx context.Context, prompt string, images []string, opts ...map[string]any) (<-chan *response.StreamingChunk, error) {
	if m.streamError != nil {
		return nil, m.streamError
	}

	ch := make(chan *response.StreamingChunk, len(m.streamChunks))
	for i := range m.streamChunks {
		ch <- &m.streamChunks[i]
	}
	close(ch)

	return ch, nil
}

// Tools returns the predetermined tools response.
func (m *MockAgent) Tools(ctx context.Context, prompt string, tools []protocol.Tool, opts ...map[string]any) (*response.ToolsResponse, error) {
	return m.toolsResponse, m.toolsError
}

// Embed returns the predetermined embeddings response.
func (m *MockAgent) Embed(ctx context.Context, input string, opts ...map[string]any) (*response.EmbeddingsResponse, error) {
	return m.embeddingsResponse, m.embeddingsError
}

// Audio returns the predetermined audio response.
func (m *MockAgent) Audio(ctx context.Context, input string, opts ...map[string]any) (*response.AudioResponse, error) {
	return m.audioResponse, m.audioError
}

// Verify MockAgent implements agent.Agent interface.
var _ agent.Agent = (*MockAgent)(nil)
