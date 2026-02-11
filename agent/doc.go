// Package agent provides a high-level interface for LLM interactions.
// It wraps the transport layer with convenient methods for common operations
// like chat, vision, tools, and embeddings, with automatic system prompt injection
// and simplified error handling.
//
// # Agent Interface
//
// The Agent interface provides protocol-specific methods:
//
//	type Agent interface {
//	    Client() transport.Client
//	    Provider() providers.Provider
//	    Model() models.Model
//
//	    Chat(ctx context.Context, prompt string, opts ...map[string]any) (*types.ChatResponse, error)
//	    ChatStream(ctx context.Context, prompt string, opts ...map[string]any) (<-chan types.StreamingChunk, error)
//
//	    Vision(ctx context.Context, prompt string, images []string, opts ...map[string]any) (*types.ChatResponse, error)
//	    VisionStream(ctx context.Context, prompt string, images []string, opts ...map[string]any) (<-chan types.StreamingChunk, error)
//
//	    Tools(ctx context.Context, prompt string, tools []protocol.Tool, opts ...map[string]any) (*types.ToolsResponse, error)
//
//	    Embed(ctx context.Context, input string, opts ...map[string]any) (*types.EmbeddingsResponse, error)
//	}
//
// # Creating an Agent
//
// Agents are created from configuration that includes transport and optional system prompt:
//
//	cfg := &config.AgentConfig{
//	    SystemPrompt: "You are a helpful AI assistant.",
//	    Transport: &config.TransportConfig{
//	        Provider: &config.ProviderConfig{
//	            Name:    "ollama",
//	            BaseURL: "http://localhost:11434",
//	            Model: &config.ModelConfig{
//	                Name: "llama2",
//	                Capabilities: map[string]config.CapabilityConfig{
//	                    "chat": {
//	                        Format: "openai-chat",
//	                        Options: map[string]any{
//	                            "temperature": 0.7,
//	                        },
//	                    },
//	                },
//	            },
//	        },
//	        Timeout:            config.Duration(30 * time.Second),
//	        ConnectionTimeout:  config.Duration(10 * time.Second),
//	        ConnectionPoolSize: 10,
//	    },
//	}
//
//	agent, err := agent.New(cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Chat Protocol
//
// Simple text-based conversation:
//
//	ctx := context.Background()
//	response, err := agent.Chat(ctx, "What is Go?")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(response.Content())
//
// With options:
//
//	options := map[string]any{
//	    "temperature": 0.9,
//	    "max_tokens":  2000,
//	}
//	response, err := agent.Chat(ctx, "Tell me a story", options)
//
// Streaming:
//
//	chunks, err := agent.ChatStream(ctx, "Tell me a long story")
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
// # Vision Protocol
//
// Image understanding with multimodal inputs:
//
//	images := []string{
//	    "https://example.com/image1.jpg",
//	    "data:image/jpeg;base64,/9j/4AAQ...",
//	}
//
//	response, err := agent.Vision(ctx, "What do you see in these images?", images)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(response.Content())
//
// Streaming vision:
//
//	chunks, err := agent.VisionStream(ctx, "Describe this image in detail", images)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	for chunk := range chunks {
//	    fmt.Print(chunk.Content())
//	}
//
// # Tools Protocol
//
// Function calling with tool definitions:
//
//	tools := []protocol.Tool{
//	    {
//	        Name:        "get_weather",
//	        Description: "Get the current weather for a location",
//	        Parameters: map[string]any{
//	            "type": "object",
//	            "properties": map[string]any{
//	                "location": map[string]any{
//	                    "type":        "string",
//	                    "description": "City name",
//	                },
//	                "unit": map[string]any{
//	                    "type":        "string",
//	                    "enum":        []string{"celsius", "fahrenheit"},
//	                    "description": "Temperature unit",
//	                },
//	            },
//	            "required": []string{"location"},
//	        },
//	    },
//	}
//
//	response, err := agent.Tools(ctx, "What's the weather in San Francisco?", tools)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Process tool calls
//	for _, toolCall := range response.ToolCalls() {
//	    fmt.Printf("Tool: %s\n", toolCall.Name())
//	    fmt.Printf("Arguments: %s\n", toolCall.Arguments())
//	}
//
// # Embeddings Protocol
//
// Text vectorization for semantic search:
//
//	response, err := agent.Embed(ctx, "The quick brown fox")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	embeddings := response.Embeddings()
//	fmt.Printf("Vector dimension: %d\n", len(embeddings[0]))
//
// With options:
//
//	options := map[string]any{
//	    "encoding_format": "float",
//	}
//	response, err := agent.Embed(ctx, "text to embed", options)
//
// # System Prompt Injection
//
// When an agent is created with a system prompt, it's automatically prepended
// to all protocol requests that support messages:
//
//	cfg := &config.AgentConfig{
//	    SystemPrompt: "You are an expert Go programmer.",
//	    Transport:    transportConfig,
//	}
//
//	agent, _ := agent.New(cfg)
//
//	// System prompt is automatically injected before user prompt
//	response, err := agent.Chat(ctx, "How do I use channels?")
//
// The message sequence becomes:
//  1. System: "You are an expert Go programmer."
//  2. User: "How do I use channels?"
//
// Affects: Chat, ChatStream, Vision, VisionStream, Tools
// Does not affect: Embed (embeddings protocol doesn't use messages)
//
// # Options Management
//
// All protocol methods accept optional parameters:
//
//	// No options
//	response, err := agent.Chat(ctx, "Hello")
//
//	// With options
//	options := map[string]any{
//	    "temperature": 0.9,
//	    "max_tokens":  2000,
//	}
//	response, err := agent.Chat(ctx, "Hello", options)
//
// Options are merged with model defaults, with request options taking precedence.
//
// # Tool Definitions
//
// Tool definitions use the canonical protocol.Tool type from core/protocol:
//
//	type Tool struct {
//	    Name        string         `json:"name"`        // Function name
//	    Description string         `json:"description"` // What the function does
//	    Parameters  map[string]any `json:"parameters"`  // JSON Schema for parameters
//	}
//
// The Parameters field uses JSON Schema format:
//
//	Parameters: map[string]any{
//	    "type": "object",
//	    "properties": map[string]any{
//	        "param_name": map[string]any{
//	            "type":        "string",
//	            "description": "Parameter description",
//	        },
//	    },
//	    "required": []string{"param_name"},
//	}
//
// # Error Handling
//
// All methods return standard Go errors:
//
//	response, err := agent.Chat(ctx, "Hello")
//	if err != nil {
//	    // Handle error
//	    log.Printf("Chat failed: %v", err)
//	    return
//	}
//
// For more detailed error information, the package provides AgentError:
//
//	err := agent.NewAgentLLMError(
//	    "Request failed",
//	    agent.WithCode("LLM500"),
//	    agent.WithCause(underlyingError),
//	)
//
// Error types:
//   - ErrorTypeInit: Initialization errors
//   - ErrorTypeLLM: LLM interaction errors
//
// Error options:
//   - WithCode: Error code for categorization
//   - WithCause: Underlying error
//   - WithName: Agent name
//   - WithClient: Client identification
//   - WithID: Unique error ID
//
// # Context Cancellation
//
// All protocol methods respect context cancellation:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	response, err := agent.Chat(ctx, "Hello")
//	if err != nil {
//	    if ctx.Err() == context.DeadlineExceeded {
//	        log.Println("Request timed out")
//	    }
//	}
//
// For streaming:
//
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//
//	chunks, err := agent.ChatStream(ctx, "Tell me a story")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Cancel after 5 seconds
//	go func() {
//	    time.Sleep(5 * time.Second)
//	    cancel()
//	}()
//
//	for chunk := range chunks {
//	    fmt.Print(chunk.Content())
//	}
//
// # Accessing Lower Layers
//
// The agent provides access to underlying components:
//
//	// Transport client
//	client := agent.Client()
//
//	// Provider
//	provider := agent.Provider()
//	fmt.Println("Provider:", provider.Name())
//
//	// Model
//	model := agent.Model()
//	fmt.Println("Model:", model.Name())
//
// This allows advanced usage while maintaining the convenience of agent methods.
//
// # Thread Safety
//
// Agents are safe for concurrent use. Multiple goroutines can call protocol methods
// simultaneously on the same agent instance.
//
// # Complete Example
//
// Comprehensive agent usage:
//
//	package main
//
//	import (
//	    "context"
//	    "fmt"
//	    "log"
//	    "time"
//
//	    "github.com/tailored-agentic-units/kernel/agent"
//	    "github.com/tailored-agentic-units/kernel/core/config"
//	)
//
//	func main() {
//	    cfg := &config.AgentConfig{
//	        SystemPrompt: "You are a helpful assistant.",
//	        Transport: &config.TransportConfig{
//	            Provider: &config.ProviderConfig{
//	                Name:    "ollama",
//	                BaseURL: "http://localhost:11434",
//	                Model: &config.ModelConfig{
//	                    Name: "llama2",
//	                    Capabilities: map[string]config.CapabilityConfig{
//	                        "chat": {
//	                            Format: "openai-chat",
//	                            Options: map[string]any{
//	                                "temperature": 0.7,
//	                            },
//	                        },
//	                    },
//	                },
//	            },
//	            Timeout:            config.Duration(30 * time.Second),
//	            ConnectionTimeout:  config.Duration(10 * time.Second),
//	            ConnectionPoolSize: 10,
//	        },
//	    }
//
//	    a, err := agent.New(cfg)
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//
//	    ctx := context.Background()
//
//	    // Simple chat
//	    response, err := a.Chat(ctx, "What is Go?")
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	    fmt.Println(response.Content())
//
//	    // Streaming chat
//	    chunks, err := a.ChatStream(ctx, "Tell me about channels")
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//
//	    for chunk := range chunks {
//	        if chunk.Error != nil {
//	            log.Printf("Error: %v", chunk.Error)
//	            continue
//	        }
//	        fmt.Print(chunk.Content())
//	    }
//	}
package agent
