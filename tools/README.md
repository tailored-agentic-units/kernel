# tools

Tool registry and execution for the TAU kernel. Provides a global catalog of tool handlers that maps LLM tool calls to their execution.

## Usage

```go
import (
    "github.com/tailored-agentic-units/kernel/tools"
    "github.com/tailored-agentic-units/kernel/core/protocol"
)

// Register a tool
tools.Register(protocol.Tool{
    Name:        "get_weather",
    Description: "Get weather for a location",
    Parameters: map[string]any{
        "type": "object",
        "properties": map[string]any{
            "location": map[string]any{"type": "string"},
        },
        "required": []string{"location"},
    },
}, weatherHandler)

// List all registered tools
available := tools.List()

// Execute a tool by name
result, err := tools.Execute(ctx, "get_weather", argsJSON)
```

Built-in tools register via `init()` in sub-packages. External libraries extend the catalog by calling `tools.Register()`.
