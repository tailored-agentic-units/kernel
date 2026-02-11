package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/tailored-agentic-units/kernel/core/protocol"
)

// Handler is the function signature for tool implementations.
// Handlers receive the request context and JSON-encoded arguments from the LLM.
type Handler func(ctx context.Context, args json.RawMessage) (Result, error)

// Result is the tool execution output that feeds back into the next LLM turn.
// IsError signals to the LLM that the tool invocation failed.
type Result struct {
	Content string
	IsError bool
}

type entry struct {
	tool    protocol.Tool
	handler Handler
}

type registry struct {
	entries map[string]entry
	mu      sync.RWMutex
}

var register = &registry{
	entries: make(map[string]entry),
}

// Register adds a new tool to the global registry.
// Returns ErrAlreadyExists if a tool with the same name is already registered.
// Use Replace to update an existing tool's handler.
// Thread-safe for concurrent registration.
func Register(tool protocol.Tool, handler Handler) error {
	if tool.Name == "" {
		return ErrEmptyName
	}

	register.mu.Lock()
	defer register.mu.Unlock()

	if _, exists := register.entries[tool.Name]; exists {
		return fmt.Errorf("%w: %s", ErrAlreadyExists, tool.Name)
	}

	register.entries[tool.Name] = entry{tool: tool, handler: handler}
	return nil
}

// Replace updates an existing tool's definition and handler.
// Returns ErrNotFound if no tool with the given name is registered.
// Thread-safe for concurrent access.
func Replace(tool protocol.Tool, handler Handler) error {
	if tool.Name == "" {
		return ErrEmptyName
	}

	register.mu.Lock()
	defer register.mu.Unlock()

	if _, exists := register.entries[tool.Name]; !exists {
		return fmt.Errorf("%w: %s", ErrNotFound, tool.Name)
	}

	register.entries[tool.Name] = entry{tool: tool, handler: handler}
	return nil
}

// Get retrieves a handler by tool name.
// Returns the handler and true if found, nil and false otherwise.
// Thread-safe for concurrent access.
func Get(name string) (Handler, bool) {
	register.mu.RLock()
	defer register.mu.RUnlock()

	e, exists := register.entries[name]
	if !exists {
		return nil, false
	}
	return e.handler, true
}

// List returns the definitions of all registered tools.
// Thread-safe for concurrent access.
func List() []protocol.Tool {
	register.mu.RLock()
	defer register.mu.RUnlock()

	tools := make([]protocol.Tool, 0, len(register.entries))
	for _, e := range register.entries {
		tools = append(tools, e.tool)
	}
	return tools
}

// Execute dispatches a tool call to the registered handler by name.
// Returns ErrNotFound if the tool is not registered.
// Handler errors are wrapped with the tool name for context.
// Thread-safe for concurrent execution.
func Execute(ctx context.Context, name string, args json.RawMessage) (Result, error) {
	register.mu.RLock()
	e, exists := register.entries[name]
	register.mu.RUnlock()

	if !exists {
		return Result{}, fmt.Errorf("%w: %s", ErrNotFound, name)
	}

	result, err := e.handler(ctx, args)
	if err != nil {
		return Result{}, fmt.Errorf("tool %s execution failed: %w", name, err)
	}

	return result, nil
}
