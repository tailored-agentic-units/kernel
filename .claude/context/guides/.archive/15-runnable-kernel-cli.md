# 15 — Runnable Kernel CLI with Built-in Tools

## Problem Context

The `cmd/kernel/main.go` is a stub. All kernel subsystems (session, memory, tools, kernel loop) are complete. This task replaces the stub with a functional CLI that exercises the full agentic loop against a real LLM — providing runtime validation of the entire kernel stack.

## Architecture Approach

Two files in `cmd/kernel/`: a tools file that defines and registers built-in tools with the global registry, and a main file that wires config loading, tool registration, kernel creation, and formatted output. Follows the proven `cmd/prompt-agent/main.go` pattern for CLI structure.

Built-in tools live in `cmd/kernel/` (not in the `tools/` library) because they're CLI demo tools, not part of the kernel's core API.

A seed memory directory at `cmd/kernel/memory/` with an identity file exercises the full `buildSystemContent()` path — memory loading, system prompt composition — so every kernel subsystem is validated at runtime.

## Implementation

### Step 1: Support unlimited iterations in `kernel/kernel.go`

Change the `Run` loop from `for iteration := range k.maxIterations` to a traditional for-loop that treats 0 as unlimited. The context cancellation provides the safety net.

In `kernel/kernel.go`, replace the loop header at line 141:

```go
// before
for iteration := range k.maxIterations {

// after
for iteration := 0; k.maxIterations == 0 || iteration < k.maxIterations; iteration++ {
```

Semantics:
- **0** → run until the agent produces a final response (or context cancellation)
- **N > 0** → bounded to N iterations, returns `ErrMaxIterations` if exhausted

### Step 2: Create `cmd/kernel/tools.go`

New file. Registers 3 built-in tools with the global `tools.Register`.

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/tailored-agentic-units/kernel/core/protocol"
	"github.com/tailored-agentic-units/kernel/tools"
)

func registerBuiltinTools() {
	must(tools.Register(protocol.Tool{
		Name:        "datetime",
		Description: "Returns the current date and time in RFC3339 format.",
		Parameters: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}, handleDatetime))

	must(tools.Register(protocol.Tool{
		Name:        "read_file",
		Description: "Reads the contents of a file at the given path.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "Absolute or relative path to the file to read.",
				},
			},
			"required": []string{"path"},
		},
	}, handleReadFile))

	must(tools.Register(protocol.Tool{
		Name:        "list_directory",
		Description: "Lists files and directories at the given path.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "Absolute or relative path to the directory to list.",
				},
			},
			"required": []string{"path"},
		},
	}, handleListDirectory))
}

func must(err error) {
	if err != nil {
		panic(fmt.Sprintf("failed to register tool: %v", err))
	}
}

func handleDatetime(_ context.Context, _ json.RawMessage) (tools.Result, error) {
	return tools.Result{Content: time.Now().Format(time.RFC3339)}, nil
}

func handleReadFile(_ context.Context, raw json.RawMessage) (tools.Result, error) {
	var args struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(raw, &args); err != nil {
		return tools.Result{Content: "invalid arguments: " + err.Error(), IsError: true}, nil
	}
	if args.Path == "" {
		return tools.Result{Content: "path is required", IsError: true}, nil
	}

	data, err := os.ReadFile(args.Path)
	if err != nil {
		return tools.Result{Content: err.Error(), IsError: true}, nil
	}
	return tools.Result{Content: string(data)}, nil
}

func handleListDirectory(_ context.Context, raw json.RawMessage) (tools.Result, error) {
	var args struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(raw, &args); err != nil {
		return tools.Result{Content: "invalid arguments: " + err.Error(), IsError: true}, nil
	}
	if args.Path == "" {
		args.Path = "."
	}

	entries, err := os.ReadDir(args.Path)
	if err != nil {
		return tools.Result{Content: err.Error(), IsError: true}, nil
	}

	var b strings.Builder
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() {
			name += "/"
		}
		b.WriteString(name)
		b.WriteByte('\n')
	}
	return tools.Result{Content: b.String()}, nil
}
```

### Step 3: Replace `cmd/kernel/main.go`

Replace the stub with a functional CLI.

```go
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/tailored-agentic-units/kernel/kernel"
)

func main() {
	var (
		configFile    = flag.String("config", "", "Path to kernel config JSON file (required)")
		prompt        = flag.String("prompt", "", "Prompt to send to the agent (required)")
		systemPrompt  = flag.String("system-prompt", "", "System prompt (overrides config)")
		memoryPath    = flag.String("memory", "", "Path to memory directory (overrides config)")
		maxIterations = flag.Int("max-iterations", -1, "Maximum loop iterations; 0 for unlimited (overrides config)")
	)
	flag.Parse()

	if *configFile == "" || *prompt == "" {
		fmt.Fprintln(os.Stderr, "Usage: kernel -config <file> -prompt <text>")
		flag.PrintDefaults()
		os.Exit(1)
	}

	cfg, err := kernel.LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if *systemPrompt != "" {
		cfg.SystemPrompt = *systemPrompt
	}
	if *memoryPath != "" {
		cfg.Memory.Path = *memoryPath
	}
	if *maxIterations >= 0 {
		cfg.MaxIterations = *maxIterations
	}

	registerBuiltinTools()

	k, err := kernel.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create kernel: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	result, err := k.Run(ctx, *prompt)
	if err != nil {
		log.Fatalf("Kernel run failed: %v", err)
	}

	fmt.Printf("Response: %s\n", result.Response)

	if len(result.ToolCalls) > 0 {
		fmt.Println("\nTool Calls:")
		for i, tc := range result.ToolCalls {
			fmt.Printf("  [%d] %s(%s)\n", i+1, tc.Name, tc.Arguments)
			if tc.IsError {
				fmt.Printf("      error: %s\n", tc.Result)
			} else if len(tc.Result) > 200 {
				fmt.Printf("      → %s...\n", tc.Result[:200])
			} else {
				fmt.Printf("      → %s\n", tc.Result)
			}
		}
	}

	fmt.Printf("\nIterations: %d\n", result.Iterations)
}
```

### Step 4: Create `cmd/kernel/memory/identity.md`

New file. Seed memory content that augments the system prompt via the memory subsystem.

```markdown
# Kernel Agent

You are a TAU kernel agent running locally. You have access to tools for interacting with the filesystem and checking the current time. When asked questions, use your tools to find accurate answers rather than guessing.
```

### Step 5: Update `cmd/kernel/agent.ollama.qwen3.json`

Add the memory path so memory loads by default without requiring the `-memory` flag.

Add to the existing config JSON (sibling of `"agent"`, `"max_iterations"`, `"system_prompt"`):

```json
"memory": {
  "path": "cmd/kernel/memory"
}
```

## Remediation

Steps required to clear blockers discovered during implementation.

### R1: Add `MarshalJSON` to `ToolCall` in `core/protocol/message.go`

The `ToolCall` type has a custom `UnmarshalJSON` that flattens the nested API format (`{function: {name, arguments}}`) into flat fields (`{name, arguments}`). However, there is no corresponding `MarshalJSON` — so when assistant messages containing tool calls are replayed back to the provider, they serialize in the flat format, which Ollama's OpenAI-compatible endpoint rejects as invalid.

Add a `MarshalJSON` method that produces the nested format with `type: "function"`:

```go
func (tc ToolCall) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ID       string `json:"id"`
		Type     string `json:"type"`
		Function struct {
			Name      string `json:"name"`
			Arguments string `json:"arguments"`
		} `json:"function"`
	}{
		ID:   tc.ID,
		Type: "function",
		Function: struct {
			Name      string `json:"name"`
			Arguments string `json:"arguments"`
		}{
			Name:      tc.Name,
			Arguments: tc.Arguments,
		},
	})
}
```

This ensures round-trip fidelity: provider responses decode correctly via `UnmarshalJSON`, and replayed messages serialize correctly via `MarshalJSON`.

### R2: Add `*slog.Logger` to `Kernel` with `WithLogger` option

Add a `*slog.Logger` field to the `Kernel` struct in `kernel/kernel.go` with a discard logger as the default. Add a `WithLogger` functional option.

```go
// in imports
"io"
"log/slog"
```

Add field to `Kernel` struct:

```go
log *slog.Logger
```

Default in `New` (alongside other field initializations):

```go
log: slog.New(slog.NewTextHandler(io.Discard, nil)),
```

Add functional option:

```go
func WithLogger(l *slog.Logger) Option {
	return func(k *Kernel) { k.log = l }
}
```

### R3: Add log points in `buildSystemContent` and `Run`

In `buildSystemContent`, replace the entry iteration loop:

```go
// before
for _, entry := range entries {
	content += "\n\n" + string(entry.Value)
}

// after
for _, entry := range entries {
	k.log.Debug("memory loaded", "key", entry.Key, "bytes", len(entry.Value))
	content += "\n\n" + string(entry.Value)
}
```

In `Run`, the full method with log points integrated:

```go
func (k *Kernel) Run(ctx context.Context, prompt string) (*Result, error) {
	k.session.AddMessage(
		protocol.NewMessage(protocol.RoleUser, prompt),
	)

	result := &Result{}

	systemContent, err := k.buildSystemContent(ctx)
	if err != nil {
		return result, err
	}

	k.log.Info("run started", "prompt_length", len(prompt), "max_iterations", k.maxIterations, "tools", len(k.tools.List()))

	for iteration := 0; k.maxIterations == 0 || iteration < k.maxIterations; iteration++ {
		if err := ctx.Err(); err != nil {
			return result, err
		}

		k.log.Debug("iteration started", "iteration", iteration+1)

		messages := k.buildMessages(systemContent)

		resp, err := k.agent.Tools(ctx, messages, k.tools.List())
		if err != nil {
			return result, fmt.Errorf("agent call failed: %w", err)
		}

		if len(resp.Choices) == 0 {
			return result, fmt.Errorf("agent returned empty response")
		}

		choice := resp.Choices[0]

		if len(choice.Message.ToolCalls) == 0 {
			k.session.AddMessage(protocol.Message{
				Role:    protocol.RoleAssistant,
				Content: choice.Message.Content,
			})
			result.Response = choice.Message.Content
			result.Iterations = iteration + 1
			k.log.Info("run complete", "iterations", iteration+1, "response_length", len(result.Response))
			return result, nil
		}

		k.session.AddMessage(protocol.Message{
			Role:      protocol.RoleAssistant,
			Content:   choice.Message.Content,
			ToolCalls: choice.Message.ToolCalls,
		})

		for _, tc := range choice.Message.ToolCalls {
			k.log.Debug("tool call", "iteration", iteration+1, "name", tc.Name)

			record := ToolCallRecord{
				Iteration: iteration + 1,
				ID:        tc.ID,
				Name:      tc.Name,
				Arguments: tc.Arguments,
			}

			toolResult, toolErr := k.tools.Execute(
				ctx,
				tc.Name,
				json.RawMessage(tc.Arguments),
			)

			if toolErr != nil {
				errContent := fmt.Sprintf("error: %s", toolErr)
				k.session.AddMessage(protocol.Message{
					Role:       protocol.RoleTool,
					Content:    errContent,
					ToolCallID: tc.ID,
				})
				record.Result = errContent
				record.IsError = true
			} else {
				k.session.AddMessage(protocol.Message{
					Role:       protocol.RoleTool,
					Content:    toolResult.Content,
					ToolCallID: tc.ID,
				})
				record.Result = toolResult.Content
				record.IsError = toolResult.IsError
			}

			result.ToolCalls = append(result.ToolCalls, record)
		}

		result.Iterations = iteration + 1
	}

	k.log.Warn("max iterations reached", "iterations", k.maxIterations)
	return result, ErrMaxIterations
}
```

### R4: Add `-verbose` flag to CLI

In `cmd/kernel/main.go`, add a `-verbose` flag that creates an `slog.Logger` writing to stderr at the appropriate level, and pass it via `kernel.WithLogger`.

```go
verbose = flag.Bool("verbose", false, "Enable verbose logging to stderr")
```

After flag parsing, before `kernel.New`:

```go
var logger *slog.Logger
if *verbose {
	logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
} else {
	logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}
```

Pass to kernel:

```go
k, err := kernel.New(cfg, kernel.WithLogger(logger))
```

## Validation Criteria

- [ ] `go vet ./...` passes
- [ ] `go test ./...` passes (existing tests unbroken)
- [ ] `go run ./cmd/kernel/ -config cmd/kernel/agent.ollama.qwen3.json -prompt "What time is it?"` produces a response with tool call log
- [ ] Built-in tools are registered and callable by the LLM
- [ ] Output shows response text, tool call log, and iteration count
- [ ] `-verbose` flag shows memory loading and iteration sequence on stderr
