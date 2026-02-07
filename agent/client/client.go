package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/tailored-agentic-units/kernel/core/config"
	"github.com/tailored-agentic-units/kernel/agent/request"
	"github.com/tailored-agentic-units/kernel/core/response"
)

// Client provides the interface for executing LLM protocol requests.
// It orchestrates HTTP execution with retry logic and health tracking.
// Provider and model come from requests, enabling flexible request composition.
type Client interface {
	// HTTPClient returns a configured HTTP client.
	// Creates a new client on each call with timeout and connection pool settings.
	HTTPClient() *http.Client

	// Execute executes a protocol request and returns the parsed response.
	// Provider and model are obtained from the request.
	// Automatically retries on transient failures (HTTP 429/502/503/504, network errors).
	// Returns an error if request fails.
	Execute(ctx context.Context, req request.Request) (any, error)

	// ExecuteStream executes a streaming protocol request and returns a channel of chunks.
	// Provider and model are obtained from the request.
	// The channel is closed when streaming completes or context is cancelled.
	// Returns an error if protocol doesn't support streaming or request fails.
	ExecuteStream(ctx context.Context, req request.Request) (<-chan *response.StreamingChunk, error)

	// IsHealthy returns the current health status of the client.
	// Set to false after request failures, true after successful requests.
	// Thread-safe for concurrent access.
	IsHealthy() bool
}

// client implements the Client interface with HTTP orchestration.
type client struct {
	config *config.ClientConfig

	mutex      sync.RWMutex
	healthy    bool
	lastHealth time.Time
}

// New creates a new Client from configuration.
// Initializes HTTP settings and health tracking.
func New(cfg *config.ClientConfig) Client {
	return &client{
		config:     cfg,
		healthy:    true,
		lastHealth: time.Now(),
	}
}

// HTTPClient creates and returns a configured HTTP client.
// Each call creates a new client with timeout and connection pool settings from configuration.
func (c *client) HTTPClient() *http.Client {
	return &http.Client{
		Timeout: c.config.Timeout.ToDuration(),
		Transport: &http.Transport{
			MaxIdleConns:        c.config.ConnectionPoolSize,
			MaxIdleConnsPerHost: c.config.ConnectionPoolSize,
			IdleConnTimeout:     c.config.ConnectionTimeout.ToDuration(),
		},
	}
}

// Execute executes a standard (non-streaming) protocol request.
// Provider and model are obtained from the request.
// Executes with retry on transient failures.
func (c *client) Execute(ctx context.Context, req request.Request) (any, error) {
	return doWithRetry(ctx, c.config.Retry, func(ctx context.Context) (any, error) {
		return c.execute(ctx, req)
	})
}

// execute performs a single HTTP request attempt without retry logic.
// Returns HTTPStatusError for bad status codes, which retry logic evaluates.
func (c *client) execute(ctx context.Context, req request.Request) (any, error) {
	provider := req.Provider()
	proto := req.Protocol()

	// Marshal request body through provider
	body, err := req.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Prepare provider request
	providerRequest, err := provider.PrepareRequest(ctx, proto, body, req.Headers())
	if err != nil {
		return nil, fmt.Errorf("failed to prepare request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(
		ctx,
		"POST",
		providerRequest.URL,
		bytes.NewBuffer(providerRequest.Body),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	for key, value := range providerRequest.Headers {
		httpReq.Header.Set(key, value)
	}
	provider.SetHeaders(httpReq)

	// Execute HTTP request
	httpClient := c.HTTPClient()
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		c.setHealthy(false)
		return nil, err // Network error - retry logic will evaluate
	}
	defer resp.Body.Close()

	// Check for non-OK status - return HTTPStatusError for retry evaluation
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		c.setHealthy(false)
		return nil, &HTTPStatusError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Body:       bodyBytes,
		}
	}

	// Process response through provider
	result, err := provider.ProcessResponse(ctx, resp, proto)
	if err != nil {
		c.setHealthy(false)
		return nil, err
	}

	c.setHealthy(true)
	return result, nil
}

// ExecuteStream executes a streaming protocol request.
// Provider and model are obtained from the request.
// Verifies protocol supports streaming and executes streaming flow.
func (c *client) ExecuteStream(ctx context.Context, req request.Request) (<-chan *response.StreamingChunk, error) {
	proto := req.Protocol()

	// Verify protocol supports streaming
	if !proto.SupportsStreaming() {
		return nil, fmt.Errorf("protocol %s does not support streaming", proto)
	}

	return c.executeStream(ctx, req)
}

// executeStream performs the streaming HTTP request.
// Streaming requests are not retried - they fail immediately on error.
func (c *client) executeStream(ctx context.Context, req request.Request) (<-chan *response.StreamingChunk, error) {
	provider := req.Provider()
	proto := req.Protocol()

	// Marshal request body through provider
	body, err := req.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Prepare streaming request
	providerRequest, err := provider.PrepareStreamRequest(ctx, proto, body, req.Headers())
	if err != nil {
		return nil, fmt.Errorf("failed to prepare streaming request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(
		ctx,
		"POST",
		providerRequest.URL,
		bytes.NewBuffer(providerRequest.Body),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	for key, value := range providerRequest.Headers {
		httpReq.Header.Set(key, value)
	}
	provider.SetHeaders(httpReq)

	// Execute HTTP request
	httpClient := c.HTTPClient()
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		c.setHealthy(false)
		return nil, fmt.Errorf("streaming request failed: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		c.setHealthy(false)
		return nil, fmt.Errorf("streaming request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Process stream through provider
	stream, err := provider.ProcessStreamResponse(ctx, resp, proto)
	if err != nil {
		c.setHealthy(false)
		resp.Body.Close()
		return nil, err
	}

	// Convert provider stream to typed chunk stream
	output := make(chan *response.StreamingChunk)
	go func() {
		defer close(output)
		defer resp.Body.Close()

		for data := range stream {
			if chunk, ok := data.(*response.StreamingChunk); ok {
				select {
				case output <- chunk:
				case <-ctx.Done():
					return
				}
			}
		}
		c.setHealthy(true)
	}()

	return output, nil
}

// IsHealthy returns the current health status.
// Thread-safe for concurrent access via read mutex.
func (c *client) IsHealthy() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.healthy
}

// setHealthy updates the health status with timestamp.
// Thread-safe via write mutex.
func (c *client) setHealthy(healthy bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.healthy = healthy
	c.lastHealth = time.Now()
}
