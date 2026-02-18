// Package kernel implements the single-agent runtime loop that composes
// agent, tools, session, and memory into the observe/think/act/repeat cycle.
//
// The kernel initializes from configuration via New, creating all subsystems
// internally. Functional options allow test overrides of any subsystem.
//
//	k, err := kernel.New(&cfg)
//	result, err := k.Run(ctx, "What's the weather in Boston?")
package kernel

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/tailored-agentic-units/kernel/agent"
	"github.com/tailored-agentic-units/kernel/core/protocol"
	"github.com/tailored-agentic-units/kernel/memory"
	"github.com/tailored-agentic-units/kernel/observability"
	"github.com/tailored-agentic-units/kernel/session"
	"github.com/tailored-agentic-units/kernel/tools"
)

// Result holds the outcome of a kernel Run invocation.
type Result struct {
	Response   string           // Final text response from the agent.
	Iterations int              // Number of loop cycles completed.
	ToolCalls  []ToolCallRecord // Log of all tool invocations.
}

type ToolCallRecord struct {
	protocol.ToolCall
	Iteration int    // Loop cycle in which the call occurred.
	Result    string // Tool execution output.
	IsError   bool   // Whether execution returned an error.
}

// ToolExecutor abstracts tool listing and execution for testability.
// The default implementation delegates to the global tools package.
type ToolExecutor interface {
	List() []protocol.Tool
	Execute(ctx context.Context, name string, args json.RawMessage) (tools.Result, error)
}

type globalToolExecutor struct{}

func (globalToolExecutor) List() []protocol.Tool {
	return tools.List()
}

func (globalToolExecutor) Execute(ctx context.Context, name string, args json.RawMessage) (tools.Result, error) {
	return tools.Execute(ctx, name, args)
}

// Option configures a Kernel after config-driven initialization.
// Applied by New after cold start — overrides replace config-created defaults.
type Option func(*Kernel)

// WithAgent overrides the config-created agent.
func WithAgent(a agent.Agent) Option {
	return func(k *Kernel) { k.agent = a }
}

// WithRegistry overrides the config-created agent registry.
func WithRegistry(r *agent.Registry) Option {
	return func(k *Kernel) { k.registry = r }
}

// WithSession overrides the config-created session.
func WithSession(s session.Session) Option {
	return func(k *Kernel) { k.session = s }
}

// WithToolExecutor overrides the default global tool executor.
func WithToolExecutor(e ToolExecutor) Option {
	return func(k *Kernel) { k.tools = e }
}

// WithMemoryStore overrides the config-created memory store.
func WithMemoryStore(s memory.Store) Option {
	return func(k *Kernel) { k.store = s }
}

// WithObserver overrides the default SlogObserver.
func WithObserver(o observability.Observer) Option {
	return func(k *Kernel) { k.observer = o }
}

// Kernel is the single-agent runtime that executes the agentic loop.
type Kernel struct {
	agent         agent.Agent
	registry      *agent.Registry
	session       session.Session
	store         memory.Store
	tools         ToolExecutor
	observer      observability.Observer
	maxIterations int
	systemPrompt  string
}

// New creates a Kernel from configuration. Subsystems (agent, session, memory)
// are initialized from their respective config sections. Functional options
// applied after initialization can override any subsystem for testing.
func New(cfg *Config, opts ...Option) (*Kernel, error) {
	a, err := agent.New(&cfg.Agent)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	sesh, err := session.New(&cfg.Session)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	store, err := memory.NewStore(&cfg.Memory)
	if err != nil {
		return nil, fmt.Errorf("failed to create memory store: %w", err)
	}

	reg := agent.NewRegistry()
	for name, agentCfg := range cfg.Agents {
		if err := reg.Register(name, agentCfg); err != nil {
			return nil, fmt.Errorf("failed to register agent %q: %w", name, err)
		}
	}

	observer := observability.NewSlogObserver(slog.Default())

	k := &Kernel{
		agent:         a,
		registry:      reg,
		session:       sesh,
		store:         store,
		observer:      observer,
		tools:         globalToolExecutor{},
		maxIterations: cfg.MaxIterations,
		systemPrompt:  cfg.SystemPrompt,
	}

	for _, opt := range opts {
		opt(k)
	}

	return k, nil
}

// Registry returns the kernel's agent registry.
func (k *Kernel) Registry() *agent.Registry {
	return k.registry
}

// Run executes the observe/think/act/repeat agentic loop for the given prompt.
// Returns a Result with the final response, iteration count, and tool call log.
// When maxIterations is 0, the loop runs until the agent produces a final
// response or the context is cancelled. Returns ErrMaxIterations if a non-zero
// iteration budget is exhausted.
func (k *Kernel) Run(ctx context.Context, prompt string) (*Result, error) {
	k.session.AddMessage(
		protocol.NewMessage(protocol.RoleUser, prompt),
	)

	result := &Result{}

	systemContent, err := k.buildSystemContent(ctx)
	if err != nil {
		return result, err
	}

	k.observer.OnEvent(ctx, observability.Event{
		Type:      EventRunStart,
		Level:     observability.LevelInfo,
		Timestamp: time.Now(),
		Source:    "kernel.Run",
		Data: map[string]any{
			"prompt_length":  len(prompt),
			"max_iterations": k.maxIterations,
			"tools":          len(k.tools.List()),
		},
	})

	for iteration := 0; k.maxIterations == 0 || iteration < k.maxIterations; iteration++ {
		if err := ctx.Err(); err != nil {
			return result, err
		}

		k.observer.OnEvent(ctx, observability.Event{
			Type:      EventIterationStart,
			Level:     observability.LevelVerbose,
			Timestamp: time.Now(),
			Source:    "kernel.Run",
			Data:      map[string]any{"iteration": iteration + 1},
		})

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

			k.observer.OnEvent(ctx, observability.Event{
				Type:      EventResponse,
				Level:     observability.LevelInfo,
				Timestamp: time.Now(),
				Source:    "kernel.Run",
				Data: map[string]any{
					"iteration":       iteration + 1,
					"response_length": len(result.Response),
				},
			})

			return result, nil
		}

		k.session.AddMessage(protocol.Message{
			Role:      protocol.RoleAssistant,
			Content:   choice.Message.Content,
			ToolCalls: choice.Message.ToolCalls,
		})

		for _, tc := range choice.Message.ToolCalls {
			k.observer.OnEvent(ctx, observability.Event{
				Type:      EventToolCall,
				Level:     observability.LevelVerbose,
				Timestamp: time.Now(),
				Source:    "kernel.Run",
				Data: map[string]any{
					"iteration": iteration + 1,
					"name":      tc.Function.Name,
				},
			})

			record := ToolCallRecord{
				ToolCall:  tc,
				Iteration: iteration + 1,
			}

			toolResult, toolErr := k.tools.Execute(
				ctx,
				tc.Function.Name,
				json.RawMessage(tc.Function.Arguments),
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

			k.observer.OnEvent(ctx, observability.Event{
				Type:      EventToolComplete,
				Level:     observability.LevelVerbose,
				Timestamp: time.Now(),
				Source:    "kernel.Run",
				Data: map[string]any{
					"iteration": iteration + 1,
					"name":      tc.Function.Name,
					"error":     record.IsError,
				},
			})

			result.ToolCalls = append(result.ToolCalls, record)
		}

		result.Iterations = iteration + 1
	}

	k.observer.OnEvent(ctx, observability.Event{
		Type:      EventError,
		Level:     observability.LevelWarning,
		Timestamp: time.Now(),
		Source:    "kernel.Run",
		Data: map[string]any{
			"error":      "max iterations reached",
			"iterations": k.maxIterations,
		},
	})

	return result, ErrMaxIterations
}

func (k *Kernel) buildMessages(systemContent string) []protocol.Message {
	sessionMsgs := k.session.Messages()

	if systemContent == "" {
		return sessionMsgs
	}

	messages := make([]protocol.Message, 0, len(sessionMsgs)+1)
	messages = append(messages, protocol.NewMessage(protocol.RoleSystem, systemContent))
	messages = append(messages, sessionMsgs...)
	return messages
}

func (k *Kernel) buildSystemContent(ctx context.Context) (string, error) {
	content := k.systemPrompt

	if k.store == nil {
		return content, nil
	}

	keys, err := k.store.List(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list memory keys: %w", err)
	}
	if len(keys) == 0 {
		return content, nil
	}

	entries, err := k.store.Load(ctx, keys...)
	if err != nil {
		return "", fmt.Errorf("failed to load memory entries: %w", err)
	}

	for _, entry := range entries {
		content += "\n\n" + string(entry.Value)
	}

	return content, nil
}
