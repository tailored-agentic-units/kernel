// Package transport provides HTTP client orchestration for LLM protocol execution.
// It integrates providers, models, and capabilities to execute protocol requests
// with proper option management, HTTP configuration, and health tracking.
//
// # Client Interface
//
// The Client interface orchestrates protocol execution:
//
//	type Client interface {
//	    Provider() providers.Provider
//	    Model() models.Model
//	    HTTPClient() *http.Client
//
//	    ExecuteProtocol(ctx context.Context, req *capabilities.CapabilityRequest) (any, error)
//	    ExecuteProtocolStream(ctx context.Context, req *capabilities.CapabilityRequest) (<-chan types.StreamingChunk, error)
//
//	    IsHealthy() bool
//	}
//
// # Creating a Client
//
// Clients are created from transport configuration:
//
//	cfg := &config.TransportConfig{
//	    Provider: &config.ProviderConfig{
//	        Name:    "ollama",
//	        BaseURL: "http://localhost:11434",
//	        Model: &config.ModelConfig{
//	            Name: "llama2",
//	            Capabilities: map[string]config.CapabilityConfig{
//	                "chat": {
//	                    Format: "openai-chat",
//	                    Options: map[string]any{
//	                        "temperature": 0.7,
//	                    },
//	                },
//	            },
//	        },
//	    },
//	    Timeout:            config.Duration(30 * time.Second),
//	    ConnectionTimeout:  config.Duration(10 * time.Second),
//	    ConnectionPoolSize: 10,
//	}
//
//	client, err := transport.New(cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Executing Protocols
//
// Standard protocol execution:
//
//	req := &capabilities.CapabilityRequest{
//	    Protocol: types.Chat,
//	    Messages: []types.Message{
//	        types.NewMessage("user", "What is Go?"),
//	    },
//	    Options: map[string]any{
//	        "temperature": 0.9, // Override model default
//	    },
//	}
//
//	ctx := context.Background()
//	result, err := client.ExecuteProtocol(ctx, req)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Type assert result to expected type
//	if response, ok := result.(*types.ChatResponse); ok {
//	    fmt.Println(response.Content())
//	}
//
// Streaming protocol execution:
//
//	req := &capabilities.CapabilityRequest{
//	    Protocol: types.Chat,
//	    Messages: []types.Message{
//	        types.NewMessage("user", "Tell me a story"),
//	    },
//	    Options: map[string]any{
//	        "stream": true,
//	    },
//	}
//
//	ctx := context.Background()
//	chunks, err := client.ExecuteProtocolStream(ctx, req)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	for chunk := range chunks {
//	    if chunk.Error != nil {
//	        log.Printf("Stream error: %v", chunk.Error)
//	        continue
//	    }
//	    fmt.Print(chunk.Content())
//	}
//
// # Request Flow
//
// The client orchestrates the complete request flow:
//
//  1. Capability Selection: Get capability from model for the requested protocol
//  2. Option Merging: Merge model defaults with request options
//  3. Option Validation: Validate merged options against capability requirements
//  4. Request Creation: Capability creates protocol-specific request
//  5. Request Preparation: Provider prepares HTTP request with endpoint and headers
//  6. HTTP Execution: Execute HTTP request with configured client
//  7. Response Processing: Provider processes response and delegates parsing to capability
//  8. Health Tracking: Update client health status based on success/failure
//
// # Option Management
//
// Options flow through three levels with proper precedence:
//
//  1. Model Defaults: Options configured in model capabilities
//  2. Model Updates: Options updated via Model.UpdateProtocolOptions()
//  3. Request Overrides: Options provided in CapabilityRequest
//
// Request options take precedence over model defaults:
//
//	// Model has temperature: 0.7
//	req := &capabilities.CapabilityRequest{
//	    Protocol: types.Chat,
//	    Messages: messages,
//	    Options: map[string]any{
//	        "temperature": 0.9, // This overrides model default
//	        "max_tokens":  2000, // This is added to model options
//	    },
//	}
//
//	// Final options sent to provider:
//	// {
//	//     "temperature": 0.9,  (from request)
//	//     "max_tokens":  2000, (from request)
//	//     ... (other model defaults)
//	// }
//
// # HTTP Configuration
//
// The client creates HTTP clients with configured timeouts and connection pooling:
//
//	cfg := &config.TransportConfig{
//	    Timeout:            config.Duration(30 * time.Second),  // Request timeout
//	    ConnectionTimeout:  config.Duration(10 * time.Second),  // Idle connection timeout
//	    ConnectionPoolSize: 10,                                  // Max idle connections
//	    Provider:           providerConfig,
//	}
//
// Each protocol execution creates a new HTTP client with these settings.
// Connection pooling is managed by the http.Transport to reuse connections efficiently.
//
// # Health Tracking
//
// The client tracks health status based on request success/failure:
//
//	// Check health before making requests
//	if !client.IsHealthy() {
//	    log.Println("Client is unhealthy, waiting before retry")
//	    time.Sleep(time.Second)
//	}
//
//	// Execute request
//	result, err := client.ExecuteProtocol(ctx, req)
//	if err != nil {
//	    // Client is now unhealthy
//	    if !client.IsHealthy() {
//	        log.Println("Client marked unhealthy after failure")
//	    }
//	}
//
// Health status is updated:
//   - Set to healthy on successful request completion
//   - Set to unhealthy on HTTP errors or response processing failures
//   - Thread-safe for concurrent health checks
//
// # Error Handling
//
// The client returns errors for various failure scenarios:
//
//	result, err := client.ExecuteProtocol(ctx, req)
//	if err != nil {
//	    // Error types:
//	    // - "capability selection failed": Protocol not supported by model
//	    // - "invalid options": Options failed validation
//	    // - "failed to create request": Capability request creation failed
//	    // - "failed to prepare request": Provider request preparation failed
//	    // - "request failed": HTTP request execution failed
//	    // - Response processing errors from provider
//	}
//
// For streaming:
//
//	chunks, err := client.ExecuteProtocolStream(ctx, req)
//	if err != nil {
//	    // Initial errors (before stream starts):
//	    // - "capability selection failed"
//	    // - "capability ... does not support streaming"
//	    // - "invalid options"
//	    // - "failed to create streaming request"
//	    // - "streaming request failed"
//	}
//
//	for chunk := range chunks {
//	    if chunk.Error != nil {
//	        // Streaming errors (during stream):
//	        // - Parsing errors
//	        // - Network errors
//	    }
//	}
//
// # Context Cancellation
//
// Both execution methods respect context cancellation:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//	defer cancel()
//
//	result, err := client.ExecuteProtocol(ctx, req)
//	if err != nil {
//	    if ctx.Err() == context.DeadlineExceeded {
//	        log.Println("Request timed out")
//	    }
//	}
//
// For streaming, cancelling context stops chunk processing:
//
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//
//	chunks, err := client.ExecuteProtocolStream(ctx, req)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Stop streaming after 5 seconds
//	go func() {
//	    time.Sleep(5 * time.Second)
//	    cancel() // Stops chunk processing and closes channel
//	}()
//
//	for chunk := range chunks {
//	    fmt.Print(chunk.Content())
//	}
//
// # Thread Safety
//
// Clients are safe for concurrent use:
//   - Multiple goroutines can call ExecuteProtocol/ExecuteProtocolStream concurrently
//   - Health status tracking uses mutex for thread-safe updates
//   - HTTP client creation is stateless and safe for concurrent calls
//
// # Multi-Protocol Execution
//
// The same client can execute different protocols:
//
//	// Chat protocol
//	chatReq := &capabilities.CapabilityRequest{
//	    Protocol: types.Chat,
//	    Messages: []types.Message{
//	        types.NewMessage("user", "Hello"),
//	    },
//	    Options: map[string]any{},
//	}
//	chatResult, _ := client.ExecuteProtocol(ctx, chatReq)
//
//	// Embeddings protocol
//	embedReq := &capabilities.CapabilityRequest{
//	    Protocol: types.Embeddings,
//	    Messages: []types.Message{},
//	    Options: map[string]any{
//	        "input": "text to embed",
//	    },
//	}
//	embedResult, _ := client.ExecuteProtocol(ctx, embedReq)
//
// The client routes each request to the appropriate capability and handles
// protocol-specific request/response processing.
package client
