package kernel_test

import (
	"context"
	"encoding/json"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/tailored-agentic-units/kernel/agent/mock"
	"github.com/tailored-agentic-units/kernel/core/protocol"
	"github.com/tailored-agentic-units/kernel/core/response"
	"github.com/tailored-agentic-units/kernel/kernel"
	"github.com/tailored-agentic-units/kernel/memory"
	"github.com/tailored-agentic-units/kernel/tools"
)

// --- Test helpers ---

// sequentialAgent returns different responses on successive Tools calls.
type sequentialAgent struct {
	*mock.MockAgent
	responses []*response.ToolsResponse
	errors    []error
	callCount atomic.Int32
}

func newSequentialAgent(responses []*response.ToolsResponse, errs []error) *sequentialAgent {
	return &sequentialAgent{
		MockAgent: mock.NewMockAgent(mock.WithID("sequential-agent")),
		responses: responses,
		errors:    errs,
	}
}

func (a *sequentialAgent) Tools(ctx context.Context, prompt []protocol.Message, t []protocol.Tool, opts ...map[string]any) (*response.ToolsResponse, error) {
	i := int(a.callCount.Add(1)) - 1
	if i < len(a.responses) {
		var err error
		if i < len(a.errors) {
			err = a.errors[i]
		}
		return a.responses[i], err
	}
	return nil, errors.New("no more responses configured")
}

// mockToolExecutor implements kernel.ToolExecutor for testing.
type mockToolExecutor struct {
	tools   []protocol.Tool
	handler func(ctx context.Context, name string, args json.RawMessage) (tools.Result, error)
}

func (e *mockToolExecutor) List() []protocol.Tool {
	return e.tools
}

func (e *mockToolExecutor) Execute(ctx context.Context, name string, args json.RawMessage) (tools.Result, error) {
	return e.handler(ctx, name, args)
}

// mockMemoryStore implements memory.Store for testing.
type mockMemoryStore struct {
	keys    []string
	entries []memory.Entry
	listErr error
	loadErr error
}

func (s *mockMemoryStore) List(ctx context.Context) ([]string, error) {
	return s.keys, s.listErr
}

func (s *mockMemoryStore) Load(ctx context.Context, keys ...string) ([]memory.Entry, error) {
	return s.entries, s.loadErr
}

func (s *mockMemoryStore) Save(ctx context.Context, entries ...memory.Entry) error {
	return nil
}

func (s *mockMemoryStore) Delete(ctx context.Context, keys ...string) error {
	return nil
}

// makeToolsResponse builds a ToolsResponse with tool calls.
func makeToolsResponse(toolCalls []protocol.ToolCall) *response.ToolsResponse {
	resp := &response.ToolsResponse{Model: "mock"}
	resp.Choices = append(resp.Choices, struct {
		Index   int `json:"index"`
		Message struct {
			Role      string              `json:"role"`
			Content   string              `json:"content"`
			ToolCalls []protocol.ToolCall `json:"tool_calls,omitempty"`
		} `json:"message"`
		FinishReason string `json:"finish_reason,omitempty"`
	}{
		Index: 0,
		Message: struct {
			Role      string              `json:"role"`
			Content   string              `json:"content"`
			ToolCalls []protocol.ToolCall `json:"tool_calls,omitempty"`
		}{
			Role:      "assistant",
			ToolCalls: toolCalls,
		},
	})
	return resp
}

// makeFinalResponse builds a ToolsResponse with text content (no tool calls).
func makeFinalResponse(content string) *response.ToolsResponse {
	resp := &response.ToolsResponse{Model: "mock"}
	resp.Choices = append(resp.Choices, struct {
		Index   int `json:"index"`
		Message struct {
			Role      string              `json:"role"`
			Content   string              `json:"content"`
			ToolCalls []protocol.ToolCall `json:"tool_calls,omitempty"`
		} `json:"message"`
		FinishReason string `json:"finish_reason,omitempty"`
	}{
		Index: 0,
		Message: struct {
			Role      string              `json:"role"`
			Content   string              `json:"content"`
			ToolCalls []protocol.ToolCall `json:"tool_calls,omitempty"`
		}{
			Role:    "assistant",
			Content: content,
		},
	})
	return resp
}

// minimalConfig returns a Config suitable for tests using functional options.
// Uses DefaultConfig so the cold start (agent, session, memory creation) succeeds
// before options override subsystems with test mocks.
func minimalConfig() *kernel.Config {
	cfg := kernel.DefaultConfig()
	return &cfg
}

// --- Tests ---

func TestRun_DirectResponse(t *testing.T) {
	agent := newSequentialAgent(
		[]*response.ToolsResponse{makeFinalResponse("Hello!")},
		nil,
	)

	k, err := kernel.New(minimalConfig(),
		kernel.WithAgent(agent),
		kernel.WithSession(newTestSession()),
		kernel.WithToolExecutor(&mockToolExecutor{}),
	)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	result, err := k.Run(context.Background(), "Hi")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if result.Response != "Hello!" {
		t.Errorf("got response %q, want %q", result.Response, "Hello!")
	}

	if result.Iterations != 1 {
		t.Errorf("got %d iterations, want 1", result.Iterations)
	}

	if len(result.ToolCalls) != 0 {
		t.Errorf("got %d tool calls, want 0", len(result.ToolCalls))
	}
}

func TestRun_SingleToolCall(t *testing.T) {
	agent := newSequentialAgent(
		[]*response.ToolsResponse{
			makeToolsResponse([]protocol.ToolCall{
				{ID: "call_1", Name: "greet", Arguments: `{"name":"world"}`},
			}),
			makeFinalResponse("Done: hello world"),
		},
		nil,
	)

	executor := &mockToolExecutor{
		tools: []protocol.Tool{{Name: "greet", Description: "Greet someone"}},
		handler: func(ctx context.Context, name string, args json.RawMessage) (tools.Result, error) {
			return tools.Result{Content: "hello world"}, nil
		},
	}

	k, err := kernel.New(minimalConfig(),
		kernel.WithAgent(agent),
		kernel.WithSession(newTestSession()),
		kernel.WithToolExecutor(executor),
	)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	result, err := k.Run(context.Background(), "Greet the world")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if result.Response != "Done: hello world" {
		t.Errorf("got response %q, want %q", result.Response, "Done: hello world")
	}

	if result.Iterations != 2 {
		t.Errorf("got %d iterations, want 2", result.Iterations)
	}

	if len(result.ToolCalls) != 1 {
		t.Fatalf("got %d tool calls, want 1", len(result.ToolCalls))
	}

	tc := result.ToolCalls[0]
	if tc.Name != "greet" {
		t.Errorf("got tool name %q, want %q", tc.Name, "greet")
	}
	if tc.Result != "hello world" {
		t.Errorf("got tool result %q, want %q", tc.Result, "hello world")
	}
	if tc.IsError {
		t.Error("tool call marked as error, want success")
	}
}

func TestRun_MultipleToolCalls(t *testing.T) {
	agent := newSequentialAgent(
		[]*response.ToolsResponse{
			makeToolsResponse([]protocol.ToolCall{
				{ID: "call_1", Name: "add", Arguments: `{"a":1,"b":2}`},
				{ID: "call_2", Name: "add", Arguments: `{"a":3,"b":4}`},
			}),
			makeFinalResponse("3 and 7"),
		},
		nil,
	)

	executor := &mockToolExecutor{
		handler: func(ctx context.Context, name string, args json.RawMessage) (tools.Result, error) {
			var params struct{ A, B int }
			json.Unmarshal(args, &params)
			return tools.Result{Content: json.Number(json.Number(string(rune('0' + params.A + params.B)))).String()}, nil
		},
	}

	k, err := kernel.New(minimalConfig(),
		kernel.WithAgent(agent),
		kernel.WithSession(newTestSession()),
		kernel.WithToolExecutor(executor),
	)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	result, err := k.Run(context.Background(), "Add these")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if len(result.ToolCalls) != 2 {
		t.Fatalf("got %d tool calls, want 2", len(result.ToolCalls))
	}
}

func TestRun_ToolExecutionError(t *testing.T) {
	agent := newSequentialAgent(
		[]*response.ToolsResponse{
			makeToolsResponse([]protocol.ToolCall{
				{ID: "call_1", Name: "fail", Arguments: `{}`},
			}),
			makeFinalResponse("I handled the error"),
		},
		nil,
	)

	executor := &mockToolExecutor{
		handler: func(ctx context.Context, name string, args json.RawMessage) (tools.Result, error) {
			return tools.Result{}, errors.New("tool broke")
		},
	}

	k, err := kernel.New(minimalConfig(),
		kernel.WithAgent(agent),
		kernel.WithSession(newTestSession()),
		kernel.WithToolExecutor(executor),
	)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	result, err := k.Run(context.Background(), "Try the failing tool")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if result.Response != "I handled the error" {
		t.Errorf("got response %q, want %q", result.Response, "I handled the error")
	}

	if len(result.ToolCalls) != 1 {
		t.Fatalf("got %d tool calls, want 1", len(result.ToolCalls))
	}

	tc := result.ToolCalls[0]
	if !tc.IsError {
		t.Error("tool call not marked as error")
	}
	if tc.Result != "error: tool broke" {
		t.Errorf("got error result %q, want %q", tc.Result, "error: tool broke")
	}
}

func TestRun_MaxIterations(t *testing.T) {
	// Agent always returns tool calls, never a final response
	infiniteToolCall := makeToolsResponse([]protocol.ToolCall{
		{ID: "call_loop", Name: "loop", Arguments: `{}`},
	})

	responses := make([]*response.ToolsResponse, 5)
	for i := range responses {
		responses[i] = infiniteToolCall
	}

	agent := newSequentialAgent(responses, nil)

	executor := &mockToolExecutor{
		handler: func(ctx context.Context, name string, args json.RawMessage) (tools.Result, error) {
			return tools.Result{Content: "looping"}, nil
		},
	}

	cfg := minimalConfig()
	cfg.MaxIterations = 3

	k, err := kernel.New(cfg,
		kernel.WithAgent(agent),
		kernel.WithSession(newTestSession()),
		kernel.WithToolExecutor(executor),
	)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	result, err := k.Run(context.Background(), "Loop forever")
	if !errors.Is(err, kernel.ErrMaxIterations) {
		t.Fatalf("got error %v, want ErrMaxIterations", err)
	}

	if result.Iterations != 3 {
		t.Errorf("got %d iterations, want 3", result.Iterations)
	}

	if len(result.ToolCalls) != 3 {
		t.Errorf("got %d tool calls, want 3", len(result.ToolCalls))
	}
}

func TestRun_ContextCancellation(t *testing.T) {
	agent := newSequentialAgent(
		[]*response.ToolsResponse{
			makeToolsResponse([]protocol.ToolCall{
				{ID: "call_1", Name: "slow", Arguments: `{}`},
			}),
		},
		nil,
	)

	ctx, cancel := context.WithCancel(context.Background())

	executor := &mockToolExecutor{
		handler: func(ctx context.Context, name string, args json.RawMessage) (tools.Result, error) {
			cancel() // Cancel after first tool execution
			return tools.Result{Content: "done"}, nil
		},
	}

	k, err := kernel.New(minimalConfig(),
		kernel.WithAgent(agent),
		kernel.WithSession(newTestSession()),
		kernel.WithToolExecutor(executor),
	)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	_, err = k.Run(ctx, "Do something")
	if !errors.Is(err, context.Canceled) {
		t.Errorf("got error %v, want context.Canceled", err)
	}
}

func TestRun_AgentError(t *testing.T) {
	agent := newSequentialAgent(nil, []error{errors.New("agent exploded")})
	agent.responses = []*response.ToolsResponse{nil}

	k, err := kernel.New(minimalConfig(),
		kernel.WithAgent(agent),
		kernel.WithSession(newTestSession()),
		kernel.WithToolExecutor(&mockToolExecutor{}),
	)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	_, err = k.Run(context.Background(), "Boom")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, errors.New("")) && err.Error() != "agent call failed: agent exploded" {
		// Just check it wraps the agent error
		if err.Error() != "agent call failed: agent exploded" {
			t.Errorf("got error %q, want wrapped agent error", err)
		}
	}
}

func TestRun_EmptyResponse(t *testing.T) {
	// Response with no choices
	emptyResp := &response.ToolsResponse{Model: "mock"}

	agent := newSequentialAgent([]*response.ToolsResponse{emptyResp}, nil)

	k, err := kernel.New(minimalConfig(),
		kernel.WithAgent(agent),
		kernel.WithSession(newTestSession()),
		kernel.WithToolExecutor(&mockToolExecutor{}),
	)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	_, err = k.Run(context.Background(), "Hello")
	if err == nil {
		t.Fatal("expected error for empty response, got nil")
	}
}

func TestRun_SystemPrompt(t *testing.T) {
	var capturedMessages []protocol.Message

	agent := newSequentialAgent(
		[]*response.ToolsResponse{makeFinalResponse("ok")},
		nil,
	)
	// Wrap to capture messages
	wrapper := &messageCapturingAgent{
		sequentialAgent: agent,
		captured:        &capturedMessages,
	}

	cfg := minimalConfig()
	cfg.SystemPrompt = "You are a test assistant."

	k, err := kernel.New(cfg,
		kernel.WithAgent(wrapper),
		kernel.WithSession(newTestSession()),
		kernel.WithToolExecutor(&mockToolExecutor{}),
	)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	_, err = k.Run(context.Background(), "Hello")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if len(capturedMessages) < 2 {
		t.Fatalf("expected at least 2 messages (system + user), got %d", len(capturedMessages))
	}

	if capturedMessages[0].Role != protocol.RoleSystem {
		t.Errorf("first message role = %q, want %q", capturedMessages[0].Role, protocol.RoleSystem)
	}
	if capturedMessages[0].Content != "You are a test assistant." {
		t.Errorf("system content = %q, want %q", capturedMessages[0].Content, "You are a test assistant.")
	}
}

func TestRun_MemoryInjection(t *testing.T) {
	var capturedMessages []protocol.Message

	agent := newSequentialAgent(
		[]*response.ToolsResponse{makeFinalResponse("ok")},
		nil,
	)
	wrapper := &messageCapturingAgent{
		sequentialAgent: agent,
		captured:        &capturedMessages,
	}

	store := &mockMemoryStore{
		keys: []string{"key1"},
		entries: []memory.Entry{
			{Key: "key1", Value: []byte("remembered context")},
		},
	}

	cfg := minimalConfig()
	cfg.SystemPrompt = "Base prompt."

	k, err := kernel.New(cfg,
		kernel.WithAgent(wrapper),
		kernel.WithSession(newTestSession()),
		kernel.WithToolExecutor(&mockToolExecutor{}),
		kernel.WithMemoryStore(store),
	)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	_, err = k.Run(context.Background(), "Hello")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if len(capturedMessages) == 0 {
		t.Fatal("no messages captured")
	}

	systemContent, ok := capturedMessages[0].Content.(string)
	if !ok {
		t.Fatalf("system content is not string: %T", capturedMessages[0].Content)
	}

	if systemContent != "Base prompt.\n\nremembered context" {
		t.Errorf("got system content %q, want %q", systemContent, "Base prompt.\n\nremembered context")
	}
}

func TestRun_MemoryListError(t *testing.T) {
	agent := newSequentialAgent(
		[]*response.ToolsResponse{makeFinalResponse("ok")},
		nil,
	)

	store := &mockMemoryStore{
		listErr: errors.New("disk failure"),
	}

	k, err := kernel.New(minimalConfig(),
		kernel.WithAgent(agent),
		kernel.WithSession(newTestSession()),
		kernel.WithToolExecutor(&mockToolExecutor{}),
		kernel.WithMemoryStore(store),
	)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	_, err = k.Run(context.Background(), "Hello")
	if err == nil {
		t.Fatal("expected error from memory list, got nil")
	}
}

func TestRun_MemoryLoadError(t *testing.T) {
	agent := newSequentialAgent(
		[]*response.ToolsResponse{makeFinalResponse("ok")},
		nil,
	)

	store := &mockMemoryStore{
		keys:    []string{"key1"},
		loadErr: errors.New("corrupt data"),
	}

	k, err := kernel.New(minimalConfig(),
		kernel.WithAgent(agent),
		kernel.WithSession(newTestSession()),
		kernel.WithToolExecutor(&mockToolExecutor{}),
		kernel.WithMemoryStore(store),
	)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	_, err = k.Run(context.Background(), "Hello")
	if err == nil {
		t.Fatal("expected error from memory load, got nil")
	}
}

func TestRun_NoMemoryStore(t *testing.T) {
	var capturedMessages []protocol.Message

	agent := newSequentialAgent(
		[]*response.ToolsResponse{makeFinalResponse("ok")},
		nil,
	)
	wrapper := &messageCapturingAgent{
		sequentialAgent: agent,
		captured:        &capturedMessages,
	}

	cfg := minimalConfig()
	cfg.SystemPrompt = "Just the prompt."

	k, err := kernel.New(cfg,
		kernel.WithAgent(wrapper),
		kernel.WithSession(newTestSession()),
		kernel.WithToolExecutor(&mockToolExecutor{}),
	)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	_, err = k.Run(context.Background(), "Hello")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	systemContent, ok := capturedMessages[0].Content.(string)
	if !ok {
		t.Fatalf("system content is not string: %T", capturedMessages[0].Content)
	}

	if systemContent != "Just the prompt." {
		t.Errorf("got %q, want %q", systemContent, "Just the prompt.")
	}
}

func TestRun_ToolCallRecordFields(t *testing.T) {
	agent := newSequentialAgent(
		[]*response.ToolsResponse{
			makeToolsResponse([]protocol.ToolCall{
				{ID: "call_abc", Name: "mytool", Arguments: `{"x":1}`},
			}),
			makeFinalResponse("done"),
		},
		nil,
	)

	executor := &mockToolExecutor{
		handler: func(ctx context.Context, name string, args json.RawMessage) (tools.Result, error) {
			return tools.Result{Content: "result_value"}, nil
		},
	}

	k, err := kernel.New(minimalConfig(),
		kernel.WithAgent(agent),
		kernel.WithSession(newTestSession()),
		kernel.WithToolExecutor(executor),
	)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	result, err := k.Run(context.Background(), "test")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if len(result.ToolCalls) != 1 {
		t.Fatalf("got %d tool calls, want 1", len(result.ToolCalls))
	}

	tc := result.ToolCalls[0]
	if tc.Iteration != 1 {
		t.Errorf("got iteration %d, want 1", tc.Iteration)
	}
	if tc.ID != "call_abc" {
		t.Errorf("got ID %q, want %q", tc.ID, "call_abc")
	}
	if tc.Name != "mytool" {
		t.Errorf("got name %q, want %q", tc.Name, "mytool")
	}
	if tc.Arguments != `{"x":1}` {
		t.Errorf("got arguments %q, want %q", tc.Arguments, `{"x":1}`)
	}
	if tc.Result != "result_value" {
		t.Errorf("got result %q, want %q", tc.Result, "result_value")
	}
	if tc.IsError {
		t.Error("expected IsError false")
	}
}

// --- Helper types ---

// messageCapturingAgent wraps sequentialAgent to capture the messages passed to Tools.
type messageCapturingAgent struct {
	*sequentialAgent
	captured *[]protocol.Message
}

func (a *messageCapturingAgent) Tools(ctx context.Context, prompt []protocol.Message, t []protocol.Tool, opts ...map[string]any) (*response.ToolsResponse, error) {
	*a.captured = make([]protocol.Message, len(prompt))
	copy(*a.captured, prompt)
	return a.sequentialAgent.Tools(ctx, prompt, t, opts...)
}

func newTestSession() *testSession {
	return &testSession{}
}

// testSession is a minimal Session implementation for kernel tests.
type testSession struct {
	messages []protocol.Message
}

func (s *testSession) ID() string                        { return "test-session" }
func (s *testSession) AddMessage(msg protocol.Message)   { s.messages = append(s.messages, msg) }
func (s *testSession) Messages() []protocol.Message      { return append([]protocol.Message{}, s.messages...) }
func (s *testSession) Clear()                            { s.messages = nil }
