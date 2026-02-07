package mock

import (
	"context"
	"net/http"
	"time"

	"github.com/tailored-agentic-units/kernel/agent/client"
	"github.com/tailored-agentic-units/kernel/agent/request"
	"github.com/tailored-agentic-units/kernel/core/response"
)

// MockClient implements client.Client interface for testing.
type MockClient struct {
	healthy bool

	// Configurable responses
	executeResponse any
	executeError    error
	streamChunks    []*response.StreamingChunk
	streamError     error
	httpClient      *http.Client
}

// NewMockClient creates a new MockClient with default configuration.
func NewMockClient(opts ...MockClientOption) *MockClient {
	m := &MockClient{
		healthy: true,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

// MockClientOption configures a MockClient.
type MockClientOption func(*MockClient)

// WithExecuteResponse sets the response for Execute.
func WithExecuteResponse(response any, err error) MockClientOption {
	return func(m *MockClient) {
		m.executeResponse = response
		m.executeError = err
	}
}

// WithStreamResponse sets the chunks for ExecuteStream.
func WithStreamResponse(chunks []*response.StreamingChunk, err error) MockClientOption {
	return func(m *MockClient) {
		m.streamChunks = chunks
		m.streamError = err
	}
}

// WithHealthy sets the health status.
func WithHealthy(healthy bool) MockClientOption {
	return func(m *MockClient) {
		m.healthy = healthy
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(c *http.Client) MockClientOption {
	return func(m *MockClient) {
		m.httpClient = c
	}
}

// HTTPClient returns the configured HTTP client.
func (m *MockClient) HTTPClient() *http.Client {
	return m.httpClient
}

// Execute returns the predetermined response.
func (m *MockClient) Execute(ctx context.Context, req request.Request) (any, error) {
	return m.executeResponse, m.executeError
}

// ExecuteStream returns a channel with predetermined chunks.
func (m *MockClient) ExecuteStream(ctx context.Context, req request.Request) (<-chan *response.StreamingChunk, error) {
	if m.streamError != nil {
		return nil, m.streamError
	}

	ch := make(chan *response.StreamingChunk, len(m.streamChunks))
	for _, chunk := range m.streamChunks {
		ch <- chunk
	}
	close(ch)

	return ch, nil
}

// IsHealthy returns the mock health status.
func (m *MockClient) IsHealthy() bool {
	return m.healthy
}

// Verify MockClient implements client.Client interface.
var _ client.Client = (*MockClient)(nil)
