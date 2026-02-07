// Package providers implements LLM service provider integrations.
// It provides a unified Provider interface for interacting with different LLM services
// (Ollama, Azure OpenAI) while handling provider-specific authentication, endpoints,
// and response formats.
//
// # Provider System
//
// The provider system follows a factory pattern with a global registry:
//
//	// Register a provider factory
//	providers.Register("custom", func(c *config.ProviderConfig) (Provider, error) {
//	    // Create and configure provider
//	    return customProvider, nil
//	})
//
//	// Create provider from configuration
//	provider, err := providers.Create(&config.ProviderConfig{
//	    Name:    "ollama",
//	    BaseURL: "http://localhost:11434",
//	    Model:   modelConfig,
//	})
//
// # Provider Interface
//
// All providers implement the Provider interface:
//
//	type Provider interface {
//	    Name() string
//	    Model() models.Model
//
//	    GetEndpoint(protocol types.Protocol) (string, error)
//	    SetHeaders(req *http.Request)
//
//	    PrepareRequest(ctx context.Context, protocol types.Protocol, request *types.Request) (*Request, error)
//	    PrepareStreamRequest(ctx context.Context, protocol types.Protocol, request *types.Request) (*Request, error)
//	    ProcessResponse(response *http.Response, capability capabilities.Capability) (any, error)
//	    ProcessStreamResponse(ctx context.Context, response *http.Response, capability capabilities.StreamingCapability) (<-chan any, error)
//	}
//
// # Built-in Providers
//
// ## Ollama Provider
//
// Ollama provider connects to local or remote Ollama instances with OpenAI-compatible API:
//
//	cfg := &config.ProviderConfig{
//	    Name:    "ollama",
//	    BaseURL: "http://localhost:11434",
//	    Model: &config.ModelConfig{
//	        Name: "llama2",
//	        Capabilities: map[string]config.CapabilityConfig{
//	            "chat": {Format: "openai-chat"},
//	        },
//	    },
//	    Options: map[string]any{
//	        "auth_type": "bearer",      // Optional: "bearer" or "api_key"
//	        "token":     "your-token",  // Optional: authentication token
//	    },
//	}
//
//	provider, err := providers.NewOllama(cfg)
//
// Features:
//   - Automatic /v1 suffix handling for OpenAI compatibility
//   - Optional bearer or API key authentication
//   - Custom authentication header support
//   - Streaming and non-streaming responses
//
// ## Azure OpenAI Provider
//
// Azure provider integrates with Azure OpenAI Service with deployment-based routing:
//
//	cfg := &config.ProviderConfig{
//	    Name:    "azure",
//	    BaseURL: "https://your-resource.openai.azure.com",
//	    Model: &config.ModelConfig{
//	        Name: "gpt-4",
//	        Capabilities: map[string]config.CapabilityConfig{
//	            "chat": {Format: "openai-chat"},
//	        },
//	    },
//	    Options: map[string]any{
//	        "deployment":  "gpt-4-deployment",  // Required: deployment name
//	        "auth_type":   "api_key",           // Required: "api_key" or "bearer"
//	        "token":       "your-api-key",      // Required: API key or bearer token
//	        "api_version": "2024-02-01",        // Required: API version
//	    },
//	}
//
//	provider, err := providers.NewAzure(cfg)
//
// Features:
//   - Deployment-based endpoint routing
//   - API key or Entra ID (bearer token) authentication
//   - API version management
//   - Server-sent events with "data: " prefix for streaming
//
// # Base Provider
//
// BaseProvider provides common functionality that provider implementations can embed:
//
//	type CustomProvider struct {
//	    *providers.BaseProvider
//	    // Custom fields
//	}
//
//	func NewCustomProvider(cfg *config.ProviderConfig) (Provider, error) {
//	    model, err := models.New(cfg.Model)
//	    if err != nil {
//	        return nil, err
//	    }
//
//	    return &CustomProvider{
//	        BaseProvider: providers.NewBaseProvider(cfg.Name, cfg.BaseURL, model),
//	    }, nil
//	}
//
// BaseProvider handles:
//   - Provider name management
//   - Base URL storage
//   - Model instance management
//
// # Request and Response Flow
//
// Standard request flow:
//
//	// 1. Get endpoint for protocol
//	endpoint, err := provider.GetEndpoint(types.Chat)
//
//	// 2. Prepare request
//	request, err := provider.PrepareRequest(ctx, types.Chat, protocolRequest)
//
//	// 3. Create HTTP request
//	httpReq, err := http.NewRequestWithContext(ctx, "POST", request.URL, bytes.NewReader(request.Body))
//	for key, value := range request.Headers {
//	    httpReq.Header.Set(key, value)
//	}
//	provider.SetHeaders(httpReq)
//
//	// 4. Execute request
//	resp, err := httpClient.Do(httpReq)
//
//	// 5. Process response
//	result, err := provider.ProcessResponse(resp, capability)
//
// Streaming request flow:
//
//	// 1-4. Same as standard flow, but use PrepareStreamRequest
//	request, err := provider.PrepareStreamRequest(ctx, types.Chat, protocolRequest)
//
//	// 5. Process streaming response
//	chunks, err := provider.ProcessStreamResponse(ctx, resp, capability)
//
//	// 6. Read streaming chunks
//	for chunk := range chunks {
//	    // Handle chunk
//	}
//
// # Request Structure
//
// The Request type packages provider-specific request details:
//
//	type Request struct {
//	    URL     string            // Full endpoint URL
//	    Headers map[string]string // Request headers
//	    Body    []byte            // Marshaled request body
//	}
//
// This structure decouples request preparation from HTTP execution.
//
// # Authentication
//
// Providers handle authentication through the SetHeaders method:
//
//	// Ollama with bearer token
//	Options: map[string]any{
//	    "auth_type": "bearer",
//	    "token":     "your-token",
//	}
//
//	// Ollama with API key
//	Options: map[string]any{
//	    "auth_type":   "api_key",
//	    "token":       "your-key",
//	    "auth_header": "X-Custom-Auth", // Optional, defaults to "X-API-Key"
//	}
//
//	// Azure with API key
//	Options: map[string]any{
//	    "auth_type": "api_key",
//	    "token":     "your-api-key",
//	}
//
//	// Azure with Entra ID token
//	Options: map[string]any{
//	    "auth_type": "bearer",
//	    "token":     "your-bearer-token",
//	}
//
// # Error Handling
//
// Providers return errors for:
//   - Unsupported protocols: GetEndpoint returns error
//   - Invalid configuration: NewProvider constructors return error
//   - HTTP failures: ProcessResponse/ProcessStreamResponse return error with status
//   - Response parsing failures: delegated to capability.ParseResponse
//
// # Thread Safety
//
// The provider registry is thread-safe for concurrent registration and creation.
// Individual provider instances are safe for concurrent use after creation.
//
// # Extending with Custom Providers
//
// To implement a custom provider:
//
//  1. Define provider struct (optionally embedding BaseProvider)
//  2. Implement Provider interface methods
//  3. Create factory function: func(c *config.ProviderConfig) (Provider, error)
//  4. Register factory: providers.Register("custom", NewCustomProvider)
//
// Example:
//
//	type CustomProvider struct {
//	    *providers.BaseProvider
//	    apiKey string
//	}
//
//	func NewCustomProvider(cfg *config.ProviderConfig) (providers.Provider, error) {
//	    apiKey, ok := cfg.Options["api_key"].(string)
//	    if !ok || apiKey == "" {
//	        return nil, fmt.Errorf("api_key is required")
//	    }
//
//	    model, err := models.New(cfg.Model)
//	    if err != nil {
//	        return nil, err
//	    }
//
//	    return &CustomProvider{
//	        BaseProvider: providers.NewBaseProvider(cfg.Name, cfg.BaseURL, model),
//	        apiKey:       apiKey,
//	    }, nil
//	}
//
//	func (p *CustomProvider) GetEndpoint(protocol types.Protocol) (string, error) {
//	    // Implement endpoint logic
//	}
//
//	// Implement remaining Provider interface methods...
package providers
