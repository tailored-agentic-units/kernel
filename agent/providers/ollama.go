package providers

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"maps"
	"net/http"
	"strings"

	"github.com/tailored-agentic-units/kernel/core/config"
	"github.com/tailored-agentic-units/kernel/core/protocol"
	"github.com/tailored-agentic-units/kernel/core/response"
)

// OllamaProvider implements Provider for Ollama services with OpenAI-compatible API.
// Supports local and remote Ollama instances with optional authentication.
type OllamaProvider struct {
	*BaseProvider
	options map[string]any
}

// NewOllama creates a new OllamaProvider from configuration.
// Automatically adds /v1 suffix to base URL if not present for OpenAI compatibility.
// Supports optional authentication via "auth_type" and "token" options.
func NewOllama(c *config.ProviderConfig) (Provider, error) {
	baseURL := c.BaseURL
	if !strings.HasSuffix(baseURL, "/v1") {
		baseURL = strings.TrimSuffix(baseURL, "/") + "/v1"
	}

	return &OllamaProvider{
		BaseProvider: NewBaseProvider(c.Name, baseURL),
		options:      c.Options,
	}, nil
}

// Endpoint returns the full Ollama endpoint URL for a protocol.
// Supports chat, vision, tools (all use /chat/completions), and embeddings (/embeddings).
// Returns an error if the protocol is not supported.
func (p *OllamaProvider) Endpoint(proto protocol.Protocol) (string, error) {
	endpoints := map[protocol.Protocol]string{
		protocol.Chat:       "/chat/completions",
		protocol.Vision:     "/chat/completions",
		protocol.Tools:      "/chat/completions",
		protocol.Embeddings: "/embeddings",
		protocol.Audio:      "/audio/transcriptions",
	}

	endpoint, exists := endpoints[proto]
	if !exists {
		return "", fmt.Errorf("protocol %s not supported by Ollama", proto)
	}

	return fmt.Sprintf("%s%s", p.BaseURL(), endpoint), nil
}

// PrepareRequest prepares a standard (non-streaming) Ollama request.
// Returns an error if the endpoint is invalid.
func (p *OllamaProvider) PrepareRequest(ctx context.Context, proto protocol.Protocol, body []byte, headers map[string]string) (*Request, error) {
	endpoint, err := p.Endpoint(proto)
	if err != nil {
		return nil, err
	}

	return &Request{
		URL:     endpoint,
		Headers: headers,
		Body:    body,
	}, nil
}

// PrepareStreamRequest prepares a streaming Ollama request.
// Adds streaming-specific headers (Accept: text/event-stream, Cache-Control: no-cache).
// Returns an error if the endpoint is invalid.
func (p *OllamaProvider) PrepareStreamRequest(ctx context.Context, proto protocol.Protocol, body []byte, headers map[string]string) (*Request, error) {
	endpoint, err := p.Endpoint(proto)
	if err != nil {
		return nil, err
	}

	// Clone headers to avoid mutating the original
	streamHeaders := make(map[string]string)
	maps.Copy(streamHeaders, headers)
	streamHeaders["Accept"] = "text/event-stream"
	streamHeaders["Cache-Control"] = "no-cache"

	return &Request{
		URL:     endpoint,
		Headers: streamHeaders,
		Body:    body,
	}, nil
}

// ProcessResponse processes a standard Ollama HTTP response.
// Returns an error if the HTTP status is not OK.
// Uses response.Parse for protocol-aware parsing.
func (p *OllamaProvider) ProcessResponse(ctx context.Context, resp *http.Response, proto protocol.Protocol) (any, error) {
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return response.Parse(proto, body)
}

// ProcessStreamResponse processes a streaming Ollama HTTP response.
// Ollama uses SSE format with "data: " prefix.
// Returns a channel that emits parsed streaming chunks.
// The channel is closed when the stream completes or context is cancelled.
// Returns an error if the HTTP status is not OK.
func (p *OllamaProvider) ProcessStreamResponse(ctx context.Context, resp *http.Response, proto protocol.Protocol) (<-chan any, error) {
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	output := make(chan any)

	go func() {
		defer close(output)
		defer resp.Body.Close()

		reader := bufio.NewReader(resp.Body)

		for {
			line, err := reader.ReadString('\n')
			if err == io.EOF {
				break
			}
			if err != nil {
				select {
				case output <- &response.StreamingChunk{Error: err}:
				case <-ctx.Done():
				}
				return
			}

			line = strings.TrimSpace(line)

			if line == "" {
				continue
			}

			// Check for completion marker
			if line == "data: [DONE]" {
				return
			}

			// Strip SSE "data: " prefix
			if after, ok := strings.CutPrefix(line, "data: "); ok {
				line = after
			}

			chunk, err := response.ParseStreamChunk(proto, []byte(line))
			if err != nil {
				continue
			}

			select {
			case output <- chunk:
			case <-ctx.Done():
				return
			}
		}
	}()

	return output, nil
}

// SetHeaders sets authentication headers on the HTTP request.
// Supports "bearer" token (Authorization: Bearer <token>) and "api_key" (custom header).
// The "auth_header" option allows customizing the API key header name (default: X-API-Key).
func (p *OllamaProvider) SetHeaders(req *http.Request) {
	if authType, ok := p.options["auth_type"].(string); ok {
		if token, ok := p.options["token"].(string); ok && token != "" {
			switch authType {
			case "bearer":
				req.Header.Set("Authorization", "Bearer "+token)
			case "api_key":
				headerName := "X-API-Key"
				if head, ok := p.options["auth_header"].(string); ok && head != "" {
					headerName = head
				}
				req.Header.Set(headerName, token)
			}
		}
	}
}
