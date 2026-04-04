# Library Extraction — Concept Development

## Context

The TAU kernel is a single Go module containing agent primitives, orchestration infrastructure, and runtime harness functionality. This extraction decomposes it into five independent libraries plus the kernel harness, so each layer can be optimized independently with a minimal API surface. Smaller repos improve both human cognition and AI context efficiency.

Design philosophy: **build the best possible library architecture, then adapt the kernel to consume it** — not the other way around.

Validated by industry direction (OpenClaw/NemoClaw pattern: primitives → coordination → runtime harness).

## Architecture

```
tau/protocol    (pure types — zero external deps)
      ↑                    ↑
tau/format            tau/provider    (independent peers, both depend on protocol)
      ↑                    ↑          (provider carries cloud SDK weight)
      └────────┬───────────┘
           tau/agent    (composition — client, request, agent interface)
               ↑
tau/orchestrate    (coordination — zero TAU agent deps, has observability)
               ↑
           tau/kernel    (harness — composes agent + orchestrate)
```

Key property: **format and provider are independent of each other**. The provider receives already-marshaled bytes; the format knows nothing about transport. The client (in tau/agent) is the composition point that orchestrates both.

---

## tau/protocol

**Module**: `github.com/tailored-agentic-units/protocol`
**Location**: `~/tau/protocol`

Zero external dependencies. Pure types that all upper layers build on.

### Package Layout

```
protocol (root)    — Protocol constants, Message, Role, ToolCall, ToolFunction
  config/          — AgentConfig, ClientConfig, ModelConfig, ProviderConfig, Duration
  response/        — Response (unified), ContentBlock, TextBlock, ToolUseBlock,
                     StreamingResponse, EmbeddingsResponse, TokenUsage
  model/           — Model runtime type (config → protocol bridge)
  streaming/       — StreamReader interface, StreamLine (interfaces only, no implementations)
```

### Protocol Constants

```go
const (
    Chat       Protocol = "chat"
    Vision     Protocol = "vision"
    Tools      Protocol = "tools"
    Embeddings Protocol = "embeddings"
    Audio      Protocol = "audio"  // reserved, no format support in v0.1.0
)
```

### Message Type (Kernel's Rich Type)

```go
type Role string

const (
    RoleSystem    Role = "system"
    RoleUser      Role = "user"
    RoleAssistant Role = "assistant"
    RoleTool      Role = "tool"
)

type Message struct {
    Role       Role       `json:"role"`
    Content    any        `json:"content"`
    ToolCallID string     `json:"tool_call_id,omitempty"`
    ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
}

// Construction helpers
func NewMessage(role Role, content any) Message
func UserMessage(content string) Message
func SystemMessage(content string) Message
func Messages(role Role, content string) []Message
```

### Response Model (Unified — from go-agents)

Replaces kernel's separate ChatResponse/ToolsResponse/StreamingChunk:

```go
// response/response.go
type Response struct {
    Role       string
    Content    []ContentBlock
    StopReason string
    Usage      *TokenUsage
}

func (r *Response) Text() string
func (r *Response) ToolCalls() []ToolUseBlock

// response/content.go
type ContentBlock interface { blockType() string }
type TextBlock struct { Text string }
type ToolUseBlock struct { ID string; Name string; Input map[string]any }

// response/streaming.go
type StreamingResponse struct {
    Content    []ContentBlock
    StopReason string
    Usage      *TokenUsage
    Error      error
}

func (r *StreamingResponse) Text() string
```

### Streaming Interfaces

```go
// streaming/streaming.go — interfaces only, implementations in tau/provider
type StreamLine struct {
    Data []byte
    Done bool
    Err  error
}

type StreamReader interface {
    ReadStream(ctx context.Context, reader io.Reader) <-chan StreamLine
}
```

### Source Mapping

| Package | Source |
|---------|--------|
| root | kernel/core/protocol/ |
| config/ | kernel/core/config/ (add Format field to AgentConfig) |
| response/ | REWRITE — unified Response + ContentBlock from go-agents |
| model/ | kernel/core/model/ |
| streaming/ | go-agents/pkg/streaming/ (interfaces only) |

### External Dependencies

None. Stdlib only.

---

## tau/format

**Module**: `github.com/tailored-agentic-units/format`
**Location**: `~/tau/format`

Wire format abstraction. Providers handle transport; formats handle serialization.

### Package Layout

```
format (root)    — Format interface, registry (init()-based), data types
                   (ChatData, VisionData, ToolsData, EmbeddingsData, Image, ToolDefinition),
                   OpenAI format, Converse format
```

### Format Interface

```go
type Format interface {
    Name() string
    Marshal(p protocol.Protocol, data any) ([]byte, error)
    Parse(p protocol.Protocol, body []byte) (any, error)
    ParseStreamChunk(p protocol.Protocol, data []byte) (*response.StreamingResponse, error)
}
```

### Data Types

Tool definitions live here (respects layering — request in tau/agent builds on format):

```go
type ToolDefinition struct {
    Name        string         `json:"name"`
    Description string         `json:"description"`
    Parameters  map[string]any `json:"parameters"`
}

type Image struct {
    Data   []byte
    Format string  // "png", "jpeg", etc.
    URL    string  // alternative to Data+Format
}

type ChatData struct {
    Model    string
    Messages []protocol.Message
    Options  map[string]any
}

// VisionData, ToolsData, EmbeddingsData follow same pattern
```

### Registry (init()-based, database/sql pattern)

```go
func Register(name string, factory Factory)
func Create(name string) (Format, error)
func List() []string

// In openai.go
func init() { Register("openai", func() (Format, error) { return &openaiFormat{}, nil }) }
```

### Source Mapping

| Component | Source |
|-----------|--------|
| Format interface | go-agents/pkg/format/format.go |
| Data types | go-agents/pkg/format/data.go |
| Registry | go-agents/pkg/format/registry.go |
| OpenAI format | go-agents/pkg/format/openai.go (adapt for kernel's rich Message) |
| Converse format | go-agents/pkg/format/converse.go (adapt for kernel's rich Message) |

### External Dependencies

- `github.com/tailored-agentic-units/protocol`

---

## tau/provider

**Module**: `github.com/tailored-agentic-units/provider`
**Location**: `~/tau/provider`

Transport, authentication, and streaming implementations. This is where cloud SDK weight lives — consumers who only use Ollama don't pull in AWS/Azure SDKs (they avoid importing provider implementations that require them, or we use build tags).

### Package Layout

```
provider (root)    — Provider interface, BaseProvider, registry (init()-based), Request type
  streaming/       — SSE reader, EventStream reader (implementations of protocol/streaming interfaces)
  identities/      — AWS SigV4, Azure managed identity credential sourcing
```

Provider implementations (Ollama, Azure, Bedrock) live in the root package as separate files.

### Provider Interface

```go
type Provider interface {
    Name() string
    BaseURL() string
    Endpoint(p protocol.Protocol) (string, error)
    Stream() streaming.StreamReader
    SetHeaders(ctx context.Context, req *http.Request) error
    PrepareRequest(ctx context.Context, p protocol.Protocol, body []byte, headers map[string]string) (*Request, error)
    PrepareStreamRequest(ctx context.Context, p protocol.Protocol, body []byte, headers map[string]string) (*Request, error)
}

type Request struct {
    URL     string
    Headers map[string]string
    Body    []byte
}
```

### Registry (init()-based)

```go
func Register(name string, factory Factory)
func Create(cfg *config.ProviderConfig) (Provider, error)
func List() []string

// In ollama.go
func init() { Register("ollama", func(c *config.ProviderConfig) (Provider, error) { return NewOllama(c) }) }
```

### Source Mapping

| Component | Source |
|-----------|--------|
| Provider interface | go-agents/pkg/providers/provider.go |
| BaseProvider | go-agents/pkg/providers/base.go |
| Registry | go-agents/pkg/providers/registry.go |
| Ollama | kernel + go-agents merged |
| Azure | kernel + go-agents merged (ctx-aware SetHeaders) |
| Bedrock | go-agents/pkg/providers/bedrock.go (new to TAU) |
| SSE reader | go-agents/pkg/streaming/sse.go |
| EventStream reader | go-agents/pkg/streaming/eventstream.go |
| AWS identities | go-agents/pkg/identities/aws.go |
| Azure identities | go-agents/pkg/identities/azure.go |

### External Dependencies

- `github.com/tailored-agentic-units/protocol`
- `github.com/aws/aws-sdk-go-v2/*` — Bedrock SigV4, credential chains
- `github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream` — EventStream framing
- `github.com/Azure/azure-sdk-for-go/sdk/*` — Azure managed identity

---

## tau/agent

**Module**: `github.com/tailored-agentic-units/agent`
**Location**: `~/tau/agent`

The composition layer — wires protocol, format, and provider together through client and request infrastructure.

### Package Layout

```
agent (root)     — Agent interface, New(cfg, provider, format), implementation
  request/       — Request interface, Chat/Vision/Tools/Embeddings request types
  client/        — Client interface, HTTP execution, retry, health tracking
  mock/          — MockAgent, MockClient, MockProvider
  registry/      — AgentRegistry (named agents, lazy instantiation)
```

### Agent Interface

Message-array primary — aligns with the raw LLM API contract. The agent is stateless transport: it takes exactly the messages you give it and sends them through the format layer. Callers own context management, enabling explicit strategies (sliding window, summarization, priority-based retention, session branching).

```go
type Agent interface {
    ID() string
    Client() client.Client
    Provider() provider.Provider
    Format() format.Format
    Model() *model.Model

    Chat(ctx context.Context, messages []protocol.Message, opts ...map[string]any) (*response.Response, error)
    ChatStream(ctx context.Context, messages []protocol.Message, opts ...map[string]any) (<-chan *response.StreamingResponse, error)
    Vision(ctx context.Context, messages []protocol.Message, images []format.Image, opts ...map[string]any) (*response.Response, error)
    VisionStream(ctx context.Context, messages []protocol.Message, images []format.Image, opts ...map[string]any) (<-chan *response.StreamingResponse, error)
    Tools(ctx context.Context, messages []protocol.Message, tools []format.ToolDefinition, opts ...map[string]any) (*response.Response, error)
    ToolsStream(ctx context.Context, messages []protocol.Message, tools []format.ToolDefinition, opts ...map[string]any) (<-chan *response.StreamingResponse, error)
    Embed(ctx context.Context, input string, opts ...map[string]any) (*response.EmbeddingsResponse, error)
}
```

### Construction (Explicit Injection)

Agent has no registry awareness. Caller resolves provider and format, then passes them in:

```go
func New(cfg *config.AgentConfig, prov provider.Provider, fmt format.Format) (Agent, error)

// Typical wiring:
fmt, _ := format.Create(cfg.Format)           // from format registry
prov, _ := provider.Create(cfg.Provider)       // from provider registry
ag, _ := agent.New(cfg, prov, fmt)
```

### Agent Registry

Higher-level concern — manages named agent instances with lazy construction:

```go
// registry/registry.go
type AgentRegistry struct { ... }
func New() *AgentRegistry
func (r *AgentRegistry) Register(name string, cfg config.AgentConfig)
func (r *AgentRegistry) Get(name string) (Agent, error)  // lazy instantiation
func (r *AgentRegistry) List() []AgentInfo
```

### Source Mapping

| Package | Source |
|---------|--------|
| agent (root) | REWRITE — []Message interface, unified Response, explicit injection |
| request/ | kernel/agent/request/ adapted for format-based marshaling |
| client/ | kernel/agent/client/ + go-agents streaming integration |
| mock/ | kernel/agent/mock/ updated for new interfaces |
| registry/ | kernel/agent/registry/ adapted |

### External Dependencies

- `github.com/tailored-agentic-units/protocol`
- `github.com/tailored-agentic-units/format`
- `github.com/tailored-agentic-units/provider`
- `github.com/google/uuid` — Agent ID (UUIDv7)

---

## tau/orchestrate

**Module**: `github.com/tailored-agentic-units/orchestrate`
**Location**: `~/tau/orchestrate`

Zero TAU agent dependencies. Coordination infrastructure with integrated observability.

### Package Layout

```
Level 0:  observability/  — Observer, Event, Level (OTel-aligned), SlogObserver, NoOpObserver, MultiObserver, registry
Level 0:  messaging/      — Message, MessageType, Priority, Builder
Level 1:  config/         — HubConfig, GraphConfig, ChainConfig, ParallelConfig, ConditionalConfig, CheckpointConfig
Level 2:  hub/            — Hub interface, Participant interface, MessageHandler, MessageContext, MessageChannel, Metrics
Level 3:  state/          — StateGraph, State, StateNode, FunctionNode, Edge, TransitionPredicate, CheckpointStore
Level 4:  workflows/      — Chain, Parallel, Conditional, Progress
```

### Participant Interface (Decoupling Seam)

```go
// hub/hub.go
type Participant interface {
    ID() string
}

type Hub interface {
    Register(p Participant, handler MessageHandler) error
    Unregister(id string) error
    Send(ctx context.Context, from, to string, data any) error
    Request(ctx context.Context, from, to string, data any) (*messaging.Message, error)
    Broadcast(ctx context.Context, from string, data any) error
    Subscribe(id, topic string) error
    Publish(ctx context.Context, from, topic string, data any) error
    Metrics() MetricsSnapshot
    Shutdown(timeout time.Duration) error
}

type MessageContext struct {
    HubName     string
    Participant Participant
}
```

Kernel bridges: `agent.Agent` satisfies `Participant` (has `ID() string`). Handlers that need full agent capabilities receive them through closure binding at registration time.

### Source Mapping

| Package | Source |
|---------|--------|
| observability/ | kernel/observability/ (port intact) |
| messaging/ | kernel/orchestrate/messaging/ (port intact) |
| config/ | kernel/orchestrate/config/ (port intact) |
| hub/ | kernel/orchestrate/hub/ (Participant decoupling) |
| state/ | kernel/orchestrate/state/ (port intact, imports local observability) |
| workflows/ | kernel/orchestrate/workflows/ (port intact) |

### External Dependencies

None beyond stdlib. Zero TAU dependencies.

---

## tau/kernel (Post-Extraction)

### What Remains

```
kernel/
├── _project/          # Project identity and context
├── kernel/            # Runtime loop (adapts to unified Response)
├── session/           # Session interface (imports protocol for Message)
├── memory/            # Store, FileStore, Cache
├── tools/             # Tool registry (imports format for ToolDefinition)
├── mcp/               # MCP client skeleton
├── cmd/               # Entry points
├── scripts/           # Infrastructure
├── .claude/           # Configuration and skills
└── .github/           # CI
```

### What Gets Removed

- `core/` — migrated to tau/protocol
- `agent/` — migrated to tau/agent
- `observability/` — migrated to tau/orchestrate
- `orchestrate/` — migrated to tau/orchestrate
- `rpc/` — dead ConnectRPC infrastructure

### Key Adaptation Points

**Runtime loop** (`kernel/kernel.go`): Currently navigates `resp.Choices[0].Message.ToolCalls` (OpenAI-shaped). Changes to `resp.ToolCalls()` returning `[]ToolUseBlock`. `ToolUseBlock.Input` is `map[string]any` — needs `json.Marshal()` before passing to `tools.Execute(ctx, name, json.RawMessage)`.

**Session + context management**: Kernel owns the full message array. Builds `[]Message` from session history, passes directly to `agent.Tools(ctx, messages, tools)`. This is already how the kernel works (`k.buildMessages()` at line 196) — the change is that the agent no longer hides this behind internal message construction. Context management strategies (sliding window, summarization) become explicit kernel responsibilities.

**Tools**: Imports `format.ToolDefinition` from tau/format for schema definitions.

**Dependencies**: Drops `connectrpc.com/connect` and `google.golang.org/protobuf`. Adds tau/protocol, tau/format, tau/provider (transitively via tau/agent), tau/agent, and tau/orchestrate.

---

## Key Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Module decomposition | 5 libraries + kernel | Smaller footprint per repo. Cleaner dependency isolation. Better for human cognition and AI context. |
| Agent.Chat signature | `[]Message` (message-array primary) | Aligns with raw LLM API contract. Caller owns context management — enables sliding window, summarization, priority retention, session branching. Agent is stateless transport. |
| Response model | Unified `Response` + `ContentBlock` | Eliminates OpenAI-shaped types leaking into domain. Format-neutral, extensible. |
| Tool type location | `format.ToolDefinition` in tau/format | Respects layering — request/ builds on format/, both below agent root. |
| Format ↔ Provider independence | Separate modules, no cross-dependency | Provider receives marshaled bytes; format knows nothing about transport. Client in tau/agent composes both. |
| Provider SDK isolation | tau/provider is its own module | Consumers using only Ollama don't pull in AWS/Azure SDKs. |
| Registry pattern | Registries in own packages, init()-based globals | Go convention (database/sql pattern). Avoids import cycles. Agent registry separate (higher-level, instance-based). |
| Agent construction | Explicit injection: `New(cfg, provider, format)` | Agent has no registry awareness. Caller resolves and passes in. Cleanest layering. |
| Observability placement | Inside tau/orchestrate | Avoids extra repo for ~200 lines. State, workflows, hub all consume it. Kernel gets it transitively. |
| Participant decoupling | Local interface `ID() string` in hub | Zero TAU deps for orchestrate. Kernel bridges via agent.Agent satisfying Participant. |
| Audio protocol | Constant reserved, no format support | Reserves extension point without blocking v0.1.0. |
| Image type | `format.Image` (structured) in tau/format | Replaces kernel's `[]string`. Supports raw bytes + URL. Cleaner API. |
| Go version | 1.26.1 | Match go-agents. Latest language features. |
| Advanced observability doc | Migrate to tau/orchestrate | Describes orchestrate-level features. Lives with the code it extends. |

---

## Phases

### tau/protocol — Phase 1: Foundation (v0.1.0)

Pure types that all upper layers depend on.

**Objectives**:
1. Protocol & message types — Protocol constants, Message, Role, ToolCall
2. Configuration types — AgentConfig, ClientConfig, ModelConfig, ProviderConfig
3. Response model — unified Response, ContentBlock, StreamingResponse, EmbeddingsResponse
4. Model & streaming interfaces — Model runtime type, StreamReader/StreamLine

### tau/format — Phase 1: Foundation (v0.1.0)

Wire format abstraction with initial format implementations.

**Objectives**:
1. Format interface & data types — Format, registry, ChatData, ToolDefinition, Image
2. OpenAI format — Marshal/Parse/ParseStreamChunk for OpenAI-compatible APIs
3. Converse format — Marshal/Parse/ParseStreamChunk for AWS Bedrock Converse

### tau/provider — Phase 1: Foundation (v0.1.0)

Transport, auth, and streaming implementations.

**Objectives**:
1. Provider interface & streaming — Provider, BaseProvider, registry, SSE reader, EventStream reader
2. Ollama provider — OpenAI-compatible local inference
3. Azure provider — Azure OpenAI with managed identity
4. Bedrock provider — AWS Bedrock with SigV4, Converse API
5. Identity management — AWS credential sourcing, Azure token sourcing

### tau/agent — Phase 1: Foundation (v0.1.0)

Composition layer wiring protocol, format, and provider.

**Objectives**:
1. Request & client — Request interface, HTTP execution, retry, streaming integration
2. Agent interface & implementation — New(cfg, provider, format), []Message contract
3. Mock & registry — Testing doubles, named agent management

### tau/orchestrate — Phase 1: Foundation (v0.1.0)

Coordination library from kernel observability + orchestrate sources.

**Objectives**:
1. Observability — Observer, Event, Level, implementations, registry
2. Messaging & config — Message types, all config types
3. Hub — Participant interface, Hub implementation, MessageHandler
4. State — StateGraph, State, CheckpointStore, predicates
5. Workflows — Chain, Parallel, Conditional patterns

### tau/kernel — Phase 1: Foundation (continued)

Extraction is maintenance/infrastructure for Phase 1, not a new phase. After libraries ship:
1. Rewire — update go.mod, delete extracted packages, update imports
2. Adapt — runtime loop to unified Response, session to external message management, tools to format.ToolDefinition
3. Remove — rpc/ directory, ConnectRPC + protobuf dependencies
4. Re-evaluate — remaining Phase 1 tasks (#26-28) against post-extraction architecture
5. Resume — Phase 1 completion toward v0.1.0

### Build Order

```
1. tau/protocol     (no deps — build first)
2. tau/format       (depends on protocol)
   tau/provider     (depends on protocol — parallel with format)
   tau/orchestrate  (no TAU deps — parallel with format + provider)
3. tau/agent        (depends on protocol + format + provider)
4. tau/kernel       (depends on agent + orchestrate — rewire)
```

---

## Project Board Structure

| Board | Repository | Phases |
|-------|-----------|--------|
| tau/protocol | tailored-agentic-units/protocol | Phase 1 - Foundation (v0.1.0) |
| tau/format | tailored-agentic-units/format | Phase 1 - Foundation (v0.1.0) |
| tau/provider | tailored-agentic-units/provider | Phase 1 - Foundation (v0.1.0) |
| tau/agent | tailored-agentic-units/agent | Phase 1 - Foundation (v0.1.0) |
| tau/orchestrate | tailored-agentic-units/orchestrate | Phase 1 - Foundation (v0.1.0) |
| TAU Kernel (existing) | tailored-agentic-units/kernel | Phase 1 - Foundation (extraction + resume, v0.1.0) |

---

## Risk Areas

1. **Response model migration** — kernel runtime loop deeply assumes OpenAI `Choices[].Message.ToolCalls` shape. `ToolUseBlock.Input` (map[string]any) vs `ToolCall.Function.Arguments` (string) requires marshaling bridge.
2. **Protocol.Message divergence** — kernel's rich Message (typed Role, ToolCalls, ToolCallID) vs go-agents' bare Message. tau/protocol keeps kernel's richer type. Must verify Converse format handles ToolCalls/ToolCallID fields gracefully.
3. **Cloud SDK weight in tau/provider** — identities pulls in significant transitive dependencies. Consider build tags or sub-packages per cloud if zero-cloud usage is a concern.
4. **Format-Message contract** — with `[]Message` as the universal input, each Format must handle all message structures (system, user, assistant, tool-result with ToolCallID). The Converse format separates system messages to a top-level field — verify this works cleanly with the kernel's rich Message type.
5. **Module coordination** — 5 libraries means 5 release cycles. Changes to tau/protocol ripple through all consumers. Mitigated by keeping protocol stable and minimal.

## Open Questions

1. **System prompt handling**: With `[]Message` as the interface, should system prompt still be part of `AgentConfig` (agent prepends it automatically), or should the caller always include it in the message array? Leaning toward caller-managed for consistency with "agent is stateless transport."
2. **Option merging**: go-agents merges model-configured defaults with runtime opts. With explicit injection, where does option merging happen — in the agent, in the caller, or in the format layer?

---

## Verification

After all libraries are built:
1. `go test ./...` passes in each repository independently
2. `go vet ./...` clean in each repository
3. tau/protocol, tau/format, tau/provider, tau/orchestrate have zero mutual TAU imports except declared dependencies
4. tau/agent depends only on tau/protocol + tau/format + tau/provider
5. Kernel CLI (`go run ./cmd/kernel/`) executes against Ollama with the new library architecture
6. Kernel integration tests pass with real provider round-trips
