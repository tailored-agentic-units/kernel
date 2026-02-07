# Tools Implementation Guide

This document outlines the architecture and implementation approach for adding actual tool execution capabilities to the agentic-toolkit. Currently, the system supports tool calling (structured output generation) but does not execute the tools.

## Current State vs Future State

### Current Flow (Tool Calling Only)
```
Agent → Model → Tool Call JSON (displayed, not executed)
```

The model generates structured JSON indicating which tools it would call and with what parameters, but no actual execution occurs.

### Desired Flow (Tool Execution)
```
Agent → Model → Tool Call JSON → Tool Executor → Real Function → Results → Model → Final Response
```

The model generates tool calls, the system executes them, and results are fed back to the model for a final response incorporating the tool outputs.

## Architecture Overview

### Core Components

1. **Tool Registry**: Maps tool names to executable functions
2. **Tool Executor**: Orchestrates the execution flow
3. **Result Integration**: Sends tool outputs back to the model
4. **Error Handling**: Manages tool execution failures gracefully
5. **Security Layer**: Controls tool permissions and sandboxing
6. **Configuration Management**: Handles API keys and tool-specific settings

## Implementation Design

### 1. Tool Registry System

```go
// pkg/tools/registry.go
package tools

import (
    "fmt"
    "sync"
)

type ToolRegistry struct {
    tools map[string]ToolFunc
    mu    sync.RWMutex
}

type ToolFunc func(args map[string]any) (any, error)

type ToolResult struct {
    Success bool        `json:"success"`
    Data    any         `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
}

func NewToolRegistry() *ToolRegistry {
    registry := &ToolRegistry{
        tools: make(map[string]ToolFunc),
    }

    // Register built-in tools
    registry.Register("get_weather", GetWeather)
    registry.Register("calculate", Calculate)
    registry.Register("web_search", WebSearch)
    registry.Register("file_read", FileRead)
    registry.Register("datetime", GetDateTime)

    return registry
}

func (r *ToolRegistry) Register(name string, fn ToolFunc) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.tools[name] = fn
}

func (r *ToolRegistry) Execute(name string, args map[string]any) (*ToolResult, error) {
    r.mu.RLock()
    fn, exists := r.tools[name]
    r.mu.RUnlock()

    if !exists {
        return &ToolResult{
            Success: false,
            Error:   fmt.Sprintf("tool %s not found", name),
        }, fmt.Errorf("tool %s not found", name)
    }

    result, err := fn(args)
    if err != nil {
        return &ToolResult{
            Success: false,
            Error:   err.Error(),
        }, err
    }

    return &ToolResult{
        Success: true,
        Data:    result,
    }, nil
}

func (r *ToolRegistry) ListTools() []string {
    r.mu.RLock()
    defer r.mu.RUnlock()

    tools := make([]string, 0, len(r.tools))
    for name := range r.tools {
        tools = append(tools, name)
    }
    return tools
}
```

### 2. Built-in Tool Implementations

```go
// pkg/tools/builtin.go
package tools

import (
    "encoding/json"
    "fmt"
    "math"
    "net/http"
    "os"
    "strconv"
    "strings"
    "time"
)

// Weather tool using OpenWeatherMap API
func GetWeather(args map[string]any) (any, error) {
    location, ok := args["location"].(string)
    if !ok {
        return nil, fmt.Errorf("location parameter required")
    }

    apiKey := os.Getenv("WEATHER_API_KEY")
    if apiKey == "" {
        return nil, fmt.Errorf("WEATHER_API_KEY environment variable not set")
    }

    url := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s&units=metric",
        location, apiKey)

    resp, err := http.Get(url)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch weather: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("weather API returned status %d", resp.StatusCode)
    }

    var weather WeatherResponse
    if err := json.NewDecoder(resp.Body).Decode(&weather); err != nil {
        return nil, fmt.Errorf("failed to parse weather response: %w", err)
    }

    return map[string]any{
        "location":    weather.Name,
        "temperature": weather.Main.Temp,
        "feels_like":  weather.Main.FeelsLike,
        "humidity":    weather.Main.Humidity,
        "description": weather.Weather[0].Description,
        "wind_speed":  weather.Wind.Speed,
    }, nil
}

// Mathematical calculator using expression evaluation
func Calculate(args map[string]any) (any, error) {
    expression, ok := args["expression"].(string)
    if !ok {
        return nil, fmt.Errorf("expression parameter required")
    }

    // Basic expression evaluator (would use library like govaluate in real implementation)
    result, err := evaluateExpression(expression)
    if err != nil {
        return nil, fmt.Errorf("calculation error: %w", err)
    }

    return map[string]any{
        "expression": expression,
        "result":     result,
    }, nil
}

// Web search using search API
func WebSearch(args map[string]any) (any, error) {
    query, ok := args["query"].(string)
    if !ok {
        return nil, fmt.Errorf("query parameter required")
    }

    limit := 5
    if l, ok := args["limit"].(float64); ok {
        limit = int(l)
    }

    // Implementation would use actual search API
    // This is a placeholder structure
    return map[string]any{
        "query":   query,
        "results": []map[string]any{
            {
                "title": "Example Result",
                "url":   "https://example.com",
                "snippet": "Example snippet for " + query,
            },
        },
        "total": limit,
    }, nil
}

// File reading with security constraints
func FileRead(args map[string]any) (any, error) {
    path, ok := args["path"].(string)
    if !ok {
        return nil, fmt.Errorf("path parameter required")
    }

    // Security: restrict to specific directories
    allowedPaths := []string{"/tmp/", "./data/", "./uploads/"}
    allowed := false
    for _, allowedPath := range allowedPaths {
        if strings.HasPrefix(path, allowedPath) {
            allowed = true
            break
        }
    }

    if !allowed {
        return nil, fmt.Errorf("access denied: path not in allowed directories")
    }

    content, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read file: %w", err)
    }

    return map[string]any{
        "path":    path,
        "content": string(content),
        "size":    len(content),
    }, nil
}

// Date and time operations
func GetDateTime(args map[string]any) (any, error) {
    format := "2006-01-02 15:04:05"
    if f, ok := args["format"].(string); ok {
        format = f
    }

    timezone := "UTC"
    if tz, ok := args["timezone"].(string); ok {
        timezone = tz
    }

    loc, err := time.LoadLocation(timezone)
    if err != nil {
        loc = time.UTC
    }

    now := time.Now().In(loc)

    return map[string]any{
        "timestamp": now.Unix(),
        "formatted": now.Format(format),
        "timezone":  timezone,
        "weekday":   now.Weekday().String(),
    }, nil
}

// Supporting types
type WeatherResponse struct {
    Name string `json:"name"`
    Main struct {
        Temp      float64 `json:"temp"`
        FeelsLike float64 `json:"feels_like"`
        Humidity  int     `json:"humidity"`
    } `json:"main"`
    Weather []struct {
        Description string `json:"description"`
    } `json:"weather"`
    Wind struct {
        Speed float64 `json:"speed"`
    } `json:"wind"`
}

// Basic expression evaluator (placeholder - use proper library)
func evaluateExpression(expr string) (float64, error) {
    // This would use a proper math expression library like govaluate
    // For now, handle simple cases
    expr = strings.ReplaceAll(expr, " ", "")

    // Handle basic arithmetic (very simplified)
    if strings.Contains(expr, "+") {
        parts := strings.Split(expr, "+")
        if len(parts) == 2 {
            a, err1 := strconv.ParseFloat(parts[0], 64)
            b, err2 := strconv.ParseFloat(parts[1], 64)
            if err1 == nil && err2 == nil {
                return a + b, nil
            }
        }
    }

    // Add more operations as needed
    return 0, fmt.Errorf("unsupported expression: %s", expr)
}
```

### 3. Agent Integration

```go
// pkg/agent/tools.go - Add tool execution to agent

type Agent interface {
    // ... existing methods ...

    // New method for tool execution
    ToolsWithExecution(ctx context.Context, prompt string, tools []Tool) (*protocols.ChatResponse, error)
    ToolsStreamWithExecution(ctx context.Context, prompt string, tools []Tool) (<-chan *protocols.StreamingChunk, error)
}

type agent struct {
    client       transport.Client
    systemPrompt string
    toolRegistry *tools.ToolRegistry  // Add tool registry
}

func New(config *config.AgentConfig) (Agent, error) {
    client, err := transport.New(config.Transport)
    if err != nil {
        return nil, fmt.Errorf("failed to create transport client: %w", err)
    }

    // Initialize tool registry
    toolRegistry := tools.NewToolRegistry()

    return &agent{
        client:       client,
        systemPrompt: config.SystemPrompt,
        toolRegistry: toolRegistry,
    }, nil
}

func (a *agent) ToolsWithExecution(ctx context.Context, prompt string, tools []Tool) (*protocols.ChatResponse, error) {
    // First, get tool calls from the model
    response, err := a.Tools(ctx, prompt, tools)
    if err != nil {
        return nil, err
    }

    // Check if the model wants to call tools
    toolCalls := extractToolCalls(response)
    if len(toolCalls) == 0 {
        return response, nil
    }

    // Execute tools and get results
    toolResults, err := a.executeTools(toolCalls)
    if err != nil {
        return nil, fmt.Errorf("tool execution failed: %w", err)
    }

    // Send tool results back to model for final response
    return a.continueWithToolResults(ctx, prompt, tools, toolResults)
}

func (a *agent) executeTools(toolCalls []ToolCall) ([]ToolResult, error) {
    results := make([]ToolResult, len(toolCalls))

    for i, call := range toolCalls {
        result, err := a.toolRegistry.Execute(call.Function.Name, call.Function.Arguments)
        if err != nil {
            // Log error but continue with other tools
            results[i] = ToolResult{
                ID:      call.ID,
                Success: false,
                Error:   err.Error(),
            }
            continue
        }

        results[i] = ToolResult{
            ID:      call.ID,
            Success: true,
            Data:    result.Data,
        }
    }

    return results, nil
}

func (a *agent) continueWithToolResults(ctx context.Context, prompt string, tools []Tool, results []ToolResult) (*protocols.ChatResponse, error) {
    // Create continuation messages with tool results
    messages := a.initMessages(prompt)

    // Add tool results as system messages
    for _, result := range results {
        resultContent, _ := json.Marshal(result)
        messages = append(messages, protocols.Message{
            Role:    "system",
            Content: fmt.Sprintf("Tool execution result: %s", string(resultContent)),
        })
    }

    // Get final response from model
    options := a.Model().Options()
    req := &capabilities.CapabilityRequest{
        Protocol: protocols.Chat,
        Messages: messages,
        Options:  options,
    }

    result, err := a.client.ExecuteProtocol(ctx, req)
    if err != nil {
        return nil, err
    }

    response, ok := result.(*protocols.ChatResponse)
    if !ok {
        return nil, fmt.Errorf("unexpected response type")
    }

    return response, nil
}

// Helper types
type ToolCall struct {
    ID       string `json:"id"`
    Type     string `json:"type"`
    Function struct {
        Name      string         `json:"name"`
        Arguments map[string]any `json:"arguments"`
    } `json:"function"`
}

type ToolResult struct {
    ID      string `json:"id"`
    Success bool   `json:"success"`
    Data    any    `json:"data,omitempty"`
    Error   string `json:"error,omitempty"`
}
```

### 4. CLI Integration

```go
// tools/prompt-agent/main.go - Enhanced CLI with execution option

func main() {
    var (
        // ... existing flags ...
        executeTools = flag.Bool("execute-tools", false, "Actually execute tool calls (default: false)")
        toolTimeout  = flag.Duration("tool-timeout", 30*time.Second, "Timeout for tool execution")
    )
    flag.Parse()

    // ... existing setup ...

    switch *protocol {
    case "tools":
        if *prompt == "" {
            log.Fatal("Error: -prompt flag is required for tools protocol")
        }
        if *toolsFile == "" {
            log.Fatal("Error: -tools-file flag is required for tools protocol")
        }

        toolList := loadTools(*toolsFile)

        if *executeTools {
            if *stream {
                handleToolsStreamWithExecution(ctx, a, *prompt, toolList)
            } else {
                handleToolsWithExecution(ctx, a, *prompt, toolList)
            }
        } else {
            // Current behavior - just show tool calls
            if *stream {
                executeToolsStream(ctx, a, *prompt, toolList)
            } else {
                executeTools(ctx, a, *prompt, toolList)
            }
        }
    }
}

func handleToolsWithExecution(ctx context.Context, agent agent.Agent, prompt string, tools []agent.Tool) {
    response, err := agent.ToolsWithExecution(ctx, prompt, tools)
    if err != nil {
        log.Fatalf("Tools execution failed: %v", err)
    }

    fmt.Printf("Final response: %s\n", response.Content())
    if response.Usage != nil {
        fmt.Printf("Tokens: %d prompt + %d completion = %d total\n",
            response.Usage.PromptTokens,
            response.Usage.CompletionTokens,
            response.Usage.TotalTokens)
    }
}

func handleToolsStreamWithExecution(ctx context.Context, agent agent.Agent, prompt string, tools []agent.Tool) {
    stream, err := agent.ToolsStreamWithExecution(ctx, prompt, tools)
    if err != nil {
        log.Fatalf("Tools stream execution failed: %v", err)
    }

    fmt.Println("Streaming response with tool execution:")
    for chunk := range stream {
        if chunk.Error != nil {
            log.Fatalf("Stream error: %v", chunk.Error)
        }
        fmt.Print(chunk.Content())
    }
    fmt.Println()
}
```

### 5. Configuration Enhancement

```go
// pkg/config/tools.go
package config

type ToolsConfig struct {
    ExecutionEnabled bool              `json:"execution_enabled"`
    Timeout          time.Duration     `json:"timeout"`
    AllowedTools     []string          `json:"allowed_tools,omitempty"`
    BlockedTools     []string          `json:"blocked_tools,omitempty"`
    APIKeys          map[string]string `json:"api_keys,omitempty"`
    SecurityPolicy   SecurityPolicy    `json:"security_policy"`
}

type SecurityPolicy struct {
    AllowFileAccess    bool     `json:"allow_file_access"`
    AllowNetworkAccess bool     `json:"allow_network_access"`
    AllowedPaths       []string `json:"allowed_paths,omitempty"`
    AllowedDomains     []string `json:"allowed_domains,omitempty"`
}

// Update AgentConfig to include tools
type AgentConfig struct {
    Name         string           `json:"name"`
    SystemPrompt string           `json:"system_prompt,omitempty"`
    Transport    *TransportConfig `json:"transport,omitempty"`
    Tools        *ToolsConfig     `json:"tools,omitempty"`
}
```

### 6. Enhanced Configuration Example

```json
{
  "name": "tools-enabled-agent",
  "system_prompt": "You are an AI assistant with access to various tools.",
  "transport": {
    "provider": {
      "name": "ollama",
      "base_url": "http://localhost:11434",
      "model": {
        "name": "llama3.2:3b",
        "format": "openai-standard",
        "options": {
          "max_tokens": 4096,
          "temperature": 0.7
        }
      }
    },
    "timeout": 60000000000
  },
  "tools": {
    "execution_enabled": true,
    "timeout": 30000000000,
    "allowed_tools": ["get_weather", "calculate", "datetime"],
    "blocked_tools": ["file_read"],
    "api_keys": {
      "weather": "${WEATHER_API_KEY}",
      "search": "${SEARCH_API_KEY}"
    },
    "security_policy": {
      "allow_file_access": false,
      "allow_network_access": true,
      "allowed_domains": ["api.openweathermap.org", "api.duckduckgo.com"]
    }
  }
}
```

## Usage Examples

### Basic Tool Execution
```bash
# Show tool calls only (current behavior)
./prompt-agent --protocol tools --tools-file tools.json --prompt "What's the weather in NYC?"

# Actually execute tools
./prompt-agent --protocol tools --tools-file tools.json --execute-tools --prompt "What's the weather in NYC?"
```

### Complex Multi-Tool Workflow
```bash
./prompt-agent --protocol tools --tools-file tools.json --execute-tools \
  --prompt "Get the weather in San Francisco, calculate a 20% tip on $85.50, and tell me the current time in Pacific timezone"
```

### Streaming with Tool Execution
```bash
./prompt-agent --protocol tools --tools-file tools.json --execute-tools --stream \
  --prompt "Search for 'Go programming tutorials' and summarize the first 3 results"
```

## Security Considerations

### Tool Permission System
- **Allowlist/Blocklist**: Control which tools can be executed
- **Path Restrictions**: Limit file system access to specific directories
- **Network Restrictions**: Control which domains tools can access
- **Timeout Controls**: Prevent long-running tool executions
- **Resource Limits**: CPU and memory constraints for tool execution

### Error Handling
- **Graceful Degradation**: Continue with partial results if some tools fail
- **Error Reporting**: Clear feedback when tools fail to execute
- **Fallback Behavior**: Option to continue without tool results
- **Logging**: Comprehensive logs for debugging and security auditing

## Implementation Phases

### Phase 1: Core Infrastructure
1. Implement ToolRegistry system
2. Add basic built-in tools (calculator, datetime)
3. Integrate with agent layer
4. Add CLI execution flags

### Phase 2: Network Tools
1. Implement weather API integration
2. Add web search capabilities
3. Implement security policies
4. Add configuration management

### Phase 3: Advanced Features
1. Tool composition and chaining
2. Async tool execution
3. Tool result caching
4. Custom tool loading from external packages

### Phase 4: Production Features
1. Comprehensive security hardening
2. Performance optimization
3. Monitoring and observability
4. Tool marketplace/registry

## Benefits

1. **Enhanced Capabilities**: AI agents can interact with real systems
2. **Flexibility**: Easy to add new tools without changing core logic
3. **Security**: Granular control over tool permissions and access
4. **Extensibility**: Third-party tools can be registered dynamically
5. **Debugging**: Clear separation between tool calling and execution
6. **Configuration**: Runtime control over tool availability and behavior

## Future Considerations

- **Tool Composition**: Combining multiple tools in complex workflows
- **Async Execution**: Non-blocking tool execution for better performance
- **Tool Marketplace**: Registry for community-contributed tools
- **Visual Tool Builder**: GUI for creating custom tools
- **Tool Versioning**: Managing different versions of tools
- **Distributed Tools**: Tools running on remote systems or cloud functions