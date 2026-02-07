package mock

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/tailored-agentic-units/kernel/core/protocol"
	"github.com/tailored-agentic-units/kernel/agent/providers"
	"github.com/tailored-agentic-units/kernel/core/response"
)

// MockProvider implements providers.Provider interface for testing.
type MockProvider struct {
	name     string
	baseURL  string
	headers  map[string]string
	endpoint string

	// Configurable responses
	marshalResponse       []byte
	marshalError          error
	prepareResponse       *providers.Request
	prepareError          error
	processResponse       any
	processError          error
	streamChunks          []any
	streamError           error
	endpointError         error
	customEndpointMapping map[protocol.Protocol]string
}

// NewMockProvider creates a new MockProvider with default configuration.
func NewMockProvider(opts ...MockProviderOption) *MockProvider {
	m := &MockProvider{
		name:                  "mock-provider",
		baseURL:               "http://mock-provider.local",
		headers:               make(map[string]string),
		endpoint:              "/mock/endpoint",
		customEndpointMapping: make(map[protocol.Protocol]string),
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

// MockProviderOption configures a MockProvider.
type MockProviderOption func(*MockProvider)

// WithProviderName sets the provider name.
func WithProviderName(name string) MockProviderOption {
	return func(m *MockProvider) {
		m.name = name
	}
}

// WithBaseURL sets the base URL.
func WithBaseURL(url string) MockProviderOption {
	return func(m *MockProvider) {
		m.baseURL = url
	}
}

// WithProviderHeaders sets custom headers.
func WithProviderHeaders(headers map[string]string) MockProviderOption {
	return func(m *MockProvider) {
		m.headers = headers
	}
}

// WithEndpoint sets the default endpoint.
func WithEndpoint(endpoint string) MockProviderOption {
	return func(m *MockProvider) {
		m.endpoint = endpoint
	}
}

// WithEndpointMapping sets custom endpoint mapping for protocols.
func WithEndpointMapping(mapping map[protocol.Protocol]string) MockProviderOption {
	return func(m *MockProvider) {
		m.customEndpointMapping = mapping
	}
}

// WithMarshalResponse sets the response for Marshal.
func WithMarshalResponse(body []byte, err error) MockProviderOption {
	return func(m *MockProvider) {
		m.marshalResponse = body
		m.marshalError = err
	}
}

// WithPrepareResponse sets the response for PrepareRequest.
func WithPrepareResponse(response *providers.Request, err error) MockProviderOption {
	return func(m *MockProvider) {
		m.prepareResponse = response
		m.prepareError = err
	}
}

// WithProcessResponse sets the response for ProcessResponse.
func WithProcessResponse(response any, err error) MockProviderOption {
	return func(m *MockProvider) {
		m.processResponse = response
		m.processError = err
	}
}

// WithProviderStreamChunks sets the chunks for ProcessStreamResponse.
func WithProviderStreamChunks(chunks []any, err error) MockProviderOption {
	return func(m *MockProvider) {
		m.streamChunks = chunks
		m.streamError = err
	}
}

// WithEndpointError sets an error for Endpoint.
func WithEndpointError(err error) MockProviderOption {
	return func(m *MockProvider) {
		m.endpointError = err
	}
}

// Name returns the provider name.
func (m *MockProvider) Name() string {
	return m.name
}

// BaseURL returns the base URL.
func (m *MockProvider) BaseURL() string {
	return m.baseURL
}

// Endpoint returns the configured endpoint for a protocol.
func (m *MockProvider) Endpoint(proto protocol.Protocol) (string, error) {
	if m.endpointError != nil {
		return "", m.endpointError
	}

	// Check custom mapping first
	if endpoint, ok := m.customEndpointMapping[proto]; ok {
		return m.baseURL + endpoint, nil
	}

	// Return default endpoint
	return m.baseURL + m.endpoint, nil
}

// SetHeaders sets the configured headers on the request.
func (m *MockProvider) SetHeaders(req *http.Request) {
	for key, value := range m.headers {
		req.Header.Set(key, value)
	}
}

// Marshal returns the predetermined marshaled body.
func (m *MockProvider) Marshal(proto protocol.Protocol, data any) ([]byte, error) {
	if m.marshalError != nil {
		return nil, m.marshalError
	}

	if m.marshalResponse != nil {
		return m.marshalResponse, nil
	}

	// Return empty JSON object as default
	return []byte(`{}`), nil
}

// PrepareRequest returns the predetermined request.
func (m *MockProvider) PrepareRequest(ctx context.Context, proto protocol.Protocol, body []byte, headers map[string]string) (*providers.Request, error) {
	if m.prepareError != nil {
		return nil, m.prepareError
	}

	if m.prepareResponse != nil {
		return m.prepareResponse, nil
	}

	// Return default request
	endpoint, _ := m.Endpoint(proto)
	return &providers.Request{
		URL:     endpoint,
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    body,
	}, nil
}

// PrepareStreamRequest returns a prepared request with streaming headers.
func (m *MockProvider) PrepareStreamRequest(ctx context.Context, proto protocol.Protocol, body []byte, headers map[string]string) (*providers.Request, error) {
	req, err := m.PrepareRequest(ctx, proto, body, headers)
	if err != nil {
		return nil, err
	}

	// Add streaming headers
	req.Headers["Accept"] = "text/event-stream"
	req.Headers["Cache-Control"] = "no-cache"

	return req, nil
}

// ProcessResponse returns the predetermined response.
func (m *MockProvider) ProcessResponse(ctx context.Context, resp *http.Response, proto protocol.Protocol) (any, error) {
	if m.processError != nil {
		return nil, m.processError
	}

	if m.processResponse != nil {
		return m.processResponse, nil
	}

	// Read response body and use protocol parsing
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return response.Parse(proto, body)
}

// ProcessStreamResponse returns a channel with predetermined chunks.
func (m *MockProvider) ProcessStreamResponse(ctx context.Context, resp *http.Response, proto protocol.Protocol) (<-chan any, error) {
	if m.streamError != nil {
		return nil, m.streamError
	}

	ch := make(chan any, len(m.streamChunks))
	for _, chunk := range m.streamChunks {
		ch <- chunk
	}
	close(ch)

	return ch, nil
}

// Verify MockProvider implements providers.Provider interface.
var _ providers.Provider = (*MockProvider)(nil)
