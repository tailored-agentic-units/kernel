// Package mock provides mock implementations of core go-agents interfaces for testing.
//
// This package enables testing of code that depends on go-agents without requiring
// real LLM providers or network connections. Each mock is configurable with
// predetermined responses and behavior.
//
// # Mock Implementations
//
// MockAgent: Implements agent.Agent interface with configurable protocol responses
//
// MockClient: Implements client.Client interface for client layer testing
//
// MockProvider: Implements providers.Provider interface with endpoint mapping
//
// # Usage Example
//
//	// Create a mock agent with predetermined chat response
//	mockAgent := mock.NewMockAgent(
//	    mock.WithChatResponse(&types.ChatResponse{
//	        Choices: []struct{ Message types.Message }{
//	            {Message: types.NewMessage("assistant", "Test response")},
//	        },
//	    }),
//	)
//
//	// Use in tests
//	response, err := mockAgent.Chat(context.Background(), "test prompt")
//	// response contains the predetermined response
//
// # Streaming Support
//
// Streaming methods return pre-populated channels that can be configured
// with test chunks:
//
//	mockAgent := mock.NewMockAgent(
//	    mock.WithStreamChunks([]types.StreamingChunk{
//	        {Content: "chunk1"},
//	        {Content: "chunk2"},
//	    }),
//	)
//
//	chunks, _ := mockAgent.ChatStream(context.Background(), "prompt")
//	for chunk := range chunks {
//	    // Process test chunks
//	}
package mock
