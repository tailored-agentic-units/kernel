package tools_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/tailored-agentic-units/kernel/core/protocol"
	"github.com/tailored-agentic-units/kernel/tools"
)

func testTool(name string) protocol.Tool {
	return protocol.Tool{
		Name:        name,
		Description: "test tool: " + name,
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"input": map[string]any{"type": "string"},
			},
		},
	}
}

func echoHandler(_ context.Context, args json.RawMessage) (tools.Result, error) {
	return tools.Result{Content: string(args)}, nil
}

func TestRegister(t *testing.T) {
	tests := []struct {
		name    string
		tool    protocol.Tool
		wantErr error
	}{
		{
			name: "valid tool",
			tool: testTool("register_valid"),
		},
		{
			name:    "empty name",
			tool:    protocol.Tool{Name: ""},
			wantErr: tools.ErrEmptyName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tools.Register(tt.tool, echoHandler)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Register() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("Register() unexpected error: %v", err)
			}
		})
	}
}

func TestRegister_Duplicate(t *testing.T) {
	tool := testTool("register_duplicate")

	if err := tools.Register(tool, echoHandler); err != nil {
		t.Fatalf("first Register() failed: %v", err)
	}

	err := tools.Register(tool, echoHandler)
	if !errors.Is(err, tools.ErrAlreadyExists) {
		t.Errorf("second Register() error = %v, want %v", err, tools.ErrAlreadyExists)
	}
}

func TestReplace(t *testing.T) {
	tool := testTool("replace_existing")

	if err := tools.Register(tool, echoHandler); err != nil {
		t.Fatalf("Register() failed: %v", err)
	}

	replacementHandler := func(_ context.Context, _ json.RawMessage) (tools.Result, error) {
		return tools.Result{Content: "replaced"}, nil
	}

	if err := tools.Replace(tool, replacementHandler); err != nil {
		t.Fatalf("Replace() failed: %v", err)
	}

	result, err := tools.Execute(context.Background(), "replace_existing", json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("Execute() after Replace() failed: %v", err)
	}
	if result.Content != "replaced" {
		t.Errorf("Execute() content = %q, want %q", result.Content, "replaced")
	}
}

func TestReplace_NotFound(t *testing.T) {
	tool := testTool("replace_nonexistent")

	err := tools.Replace(tool, echoHandler)
	if !errors.Is(err, tools.ErrNotFound) {
		t.Errorf("Replace() error = %v, want %v", err, tools.ErrNotFound)
	}
}

func TestReplace_EmptyName(t *testing.T) {
	err := tools.Replace(protocol.Tool{Name: ""}, echoHandler)
	if !errors.Is(err, tools.ErrEmptyName) {
		t.Errorf("Replace() error = %v, want %v", err, tools.ErrEmptyName)
	}
}

func TestGet(t *testing.T) {
	tool := testTool("get_existing")

	if err := tools.Register(tool, echoHandler); err != nil {
		t.Fatalf("Register() failed: %v", err)
	}

	handler, exists := tools.Get("get_existing")
	if !exists {
		t.Fatal("Get() returned exists=false, want true")
	}
	if handler == nil {
		t.Fatal("Get() returned nil handler")
	}
}

func TestGet_NotFound(t *testing.T) {
	_, exists := tools.Get("get_nonexistent")
	if exists {
		t.Error("Get() returned exists=true for nonexistent tool")
	}
}

func TestList(t *testing.T) {
	tool1 := testTool("list_tool_1")
	tool2 := testTool("list_tool_2")

	tools.Register(tool1, echoHandler)
	tools.Register(tool2, echoHandler)

	list := tools.List()

	found1, found2 := false, false
	for _, tool := range list {
		if tool.Name == "list_tool_1" {
			found1 = true
		}
		if tool.Name == "list_tool_2" {
			found2 = true
		}
	}

	if !found1 {
		t.Error("List() missing list_tool_1")
	}
	if !found2 {
		t.Error("List() missing list_tool_2")
	}
}

func TestExecute(t *testing.T) {
	tool := testTool("execute_valid")
	handler := func(_ context.Context, args json.RawMessage) (tools.Result, error) {
		var params struct {
			Input string `json:"input"`
		}
		if err := json.Unmarshal(args, &params); err != nil {
			return tools.Result{}, err
		}
		return tools.Result{Content: "echo: " + params.Input}, nil
	}

	if err := tools.Register(tool, handler); err != nil {
		t.Fatalf("Register() failed: %v", err)
	}

	result, err := tools.Execute(
		context.Background(),
		"execute_valid",
		json.RawMessage(`{"input":"hello"}`),
	)
	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}
	if result.Content != "echo: hello" {
		t.Errorf("Execute() content = %q, want %q", result.Content, "echo: hello")
	}
	if result.IsError {
		t.Error("Execute() IsError = true, want false")
	}
}

func TestExecute_NotFound(t *testing.T) {
	_, err := tools.Execute(context.Background(), "execute_nonexistent", nil)
	if !errors.Is(err, tools.ErrNotFound) {
		t.Errorf("Execute() error = %v, want %v", err, tools.ErrNotFound)
	}
}

func TestExecute_HandlerError(t *testing.T) {
	tool := testTool("execute_error")
	handlerErr := errors.New("handler failed")
	handler := func(_ context.Context, _ json.RawMessage) (tools.Result, error) {
		return tools.Result{}, handlerErr
	}

	if err := tools.Register(tool, handler); err != nil {
		t.Fatalf("Register() failed: %v", err)
	}

	_, err := tools.Execute(context.Background(), "execute_error", nil)
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}
	if !errors.Is(err, handlerErr) {
		t.Errorf("Execute() error chain does not contain handler error: %v", err)
	}
}

func TestExecute_RespectsContext(t *testing.T) {
	tool := testTool("execute_ctx")
	handler := func(ctx context.Context, _ json.RawMessage) (tools.Result, error) {
		if err := ctx.Err(); err != nil {
			return tools.Result{}, err
		}
		return tools.Result{Content: "ok"}, nil
	}

	if err := tools.Register(tool, handler); err != nil {
		t.Fatalf("Register() failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := tools.Execute(ctx, "execute_ctx", nil)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Execute() error = %v, want context.Canceled", err)
	}
}
