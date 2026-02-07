package providers

import (
	"context"
	"net/http"

	"github.com/tailored-agentic-units/kernel/core/protocol"
)

// Provider defines the interface for LLM service provider implementations.
// Providers handle endpoint routing, authentication, request marshaling,
// and response processing for their specific service.
//
// The Marshal method enables provider-specific wire formats. OpenAI-compatible
// providers can use the default implementation from BaseProvider, while providers
// with different formats (Anthropic, Google) override with their own marshaling.
type Provider interface {
	// Name returns the provider identifier.
	Name() string

	// BaseURL returns the provider's base URL.
	BaseURL() string

	// Endpoint returns the full endpoint URL for a protocol.
	// Returns an error if the protocol is not supported by this provider.
	Endpoint(p protocol.Protocol) (string, error)

	// SetHeaders sets provider-specific authentication and custom headers on an HTTP request.
	// This is called after the request is created but before it is executed.
	SetHeaders(req *http.Request)

	// Marshal converts request data to provider-specific JSON format.
	// The data parameter should be *ChatData, *VisionData, *ToolsData, or *EmbeddingsData
	// or *AudioData based on the protocol. Providers implement this to support their wire format.
	// BaseProvider provides a default OpenAI-compatible implementation.
	Marshal(p protocol.Protocol, data any) ([]byte, error)

	// PrepareRequest creates a Request for standard (non-streaming) protocol execution.
	// Accepts pre-marshaled request body and headers from the request structure.
	PrepareRequest(ctx context.Context, p protocol.Protocol, body []byte, headers map[string]string) (*Request, error)

	// PrepareStreamRequest creates a Request for streaming protocol execution.
	// Accepts pre-marshaled request body and headers, adds streaming-specific headers.
	PrepareStreamRequest(ctx context.Context, p protocol.Protocol, body []byte, headers map[string]string) (*Request, error)

	// ProcessResponse processes a standard HTTP response and returns the parsed result.
	// Uses response.Parse for protocol-aware parsing.
	// Returns an error if the HTTP status is not OK or parsing fails.
	ProcessResponse(ctx context.Context, resp *http.Response, p protocol.Protocol) (any, error)

	// ProcessStreamResponse processes a streaming HTTP response and returns a channel of chunks.
	// The channel is closed when the stream completes or an error occurs.
	// Context cancellation stops processing and closes the channel.
	ProcessStreamResponse(ctx context.Context, resp *http.Response, p protocol.Protocol) (<-chan any, error)
}

// Request represents a prepared provider request with all necessary components for HTTP execution.
// This structure decouples request preparation from HTTP client execution.
type Request struct {
	// URL is the complete endpoint URL including query parameters.
	URL string

	// Headers contains protocol-specific and provider-specific headers.
	Headers map[string]string

	// Body is the marshaled request body ready for HTTP transmission.
	Body []byte
}
