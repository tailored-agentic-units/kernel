# Issue #15 — Runnable Kernel CLI with Built-in Tools

## Context

The `cmd/kernel/main.go` is a stub printing "kernel: under development". All kernel subsystems are complete (session, memory, tools, kernel loop). This task replaces the stub with a functional CLI that exercises the full agentic loop against a real LLM, providing runtime validation of the entire kernel stack.

## Approach

Two new/modified files in `cmd/kernel/`:

### 1. `cmd/kernel/tools.go` — Built-in tool definitions and registration

A `registerBuiltinTools()` function that registers 3 tools with the global `tools.Register`:

| Tool | Args | Implementation |
|------|------|----------------|
| `datetime` | none | `time.Now().Format(time.RFC3339)` |
| `read_file` | `{"path": "string"}` | `os.ReadFile(path)` |
| `list_directory` | `{"path": "string"}` | `os.ReadDir(path)`, format as newline-separated names |

Each tool defined as `protocol.Tool` with JSON Schema parameters. Handlers are `tools.Handler` functions.

### 2. `cmd/kernel/main.go` — Functional CLI entry point

Following the `cmd/prompt-agent/main.go` pattern:

**Flags:**
- `-config` (required) — path to config JSON
- `-prompt` (required) — user prompt
- `-system-prompt` — override config value
- `-memory` — path to memory directory (override `memory.path`)
- `-max-iterations` — override config value

**Flow:**
1. Parse flags, validate required
2. `kernel.LoadConfig(configFile)`
3. Apply flag overrides to loaded config
4. `registerBuiltinTools()`
5. `kernel.New(&cfg)` → `kernel.Run(ctx, prompt)`
6. Print formatted output: response text, iteration count, tool call log

**Output format:**
```
Response: <text>

Tool Calls:
  [1] datetime() → 2026-02-16T...
  [2] read_file({"path":"go.mod"}) → module github.com/...

Iterations: 3
```

### Files Modified

- `cmd/kernel/main.go` — replace stub (existing)
- `cmd/kernel/tools.go` — new file

### Config File

`cmd/kernel/agent.ollama.qwen3.json` already exists with correct Ollama/Qwen3 config. No changes needed.

### Verification

1. `go vet ./...` passes
2. `go test ./...` passes (no new tests needed for CLI main — runtime validation is the test)
3. Manual: `go run cmd/kernel/main.go -config cmd/kernel/agent.ollama.qwen3.json -prompt "What time is it?"` produces response with tool call log
