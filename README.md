# kernel

The TAU (Tailored Agentic Units) kernel — a single Go module containing the integrated subsystems that power the TAU agent runtime.

## Module

```
github.com/tailored-agentic-units/kernel
```

## Subsystems

| Package | Description |
|---------|-------------|
| `core/` | Foundational type vocabulary: protocol constants, response types, configuration, model |
| `agent/` | LLM communication: agent interface, HTTP client, providers (Ollama, Azure), request construction |
| `orchestrate/` | Multi-agent coordination: hubs, messaging, state graphs, workflow patterns, observability |
| `memory/` | Context composition: Store interface, FileStore, session-scoped Cache |
| `tools/` | Tool execution: global registry with Register, Execute, List |
| `session/` | Conversation management: Session interface, in-memory implementation |
| `mcp/` | Model Context Protocol client (under development) |
| `kernel/` | Agent runtime loop with config-driven initialization |

## ConnectRPC Interface

The kernel exposes a single ConnectRPC service (`tau.kernel.v1.KernelService`) as the boundary between the kernel and external extensions.

```protobuf
service KernelService {
  rpc CreateSession(CreateSessionRequest) returns (CreateSessionResponse);
  rpc Run(RunRequest) returns (stream RunResponse);
  rpc InjectContext(InjectContextRequest) returns (InjectContextResponse);
  rpc GetSession(GetSessionRequest) returns (GetSessionResponse);
}
```

Proto definitions live in `rpc/proto/`, generated code in `rpc/gen/`.

## Prerequisites

- Go 1.25.7 or later
- For Ollama: Docker (optionally with `nvidia-container-toolkit` for GPU acceleration)
- For Azure: Azure CLI authenticated to a tenant with deployed Azure OpenAI models
- For proto codegen: `buf`, `protoc-gen-go`, `protoc-gen-connect-go`

## Quick Start

```bash
# Start Ollama (local LLM)
docker compose up -d

# Run the prompt-agent testing utility
go run cmd/prompt-agent/main.go \
  -config cmd/prompt-agent/config.ollama.json \
  -prompt "Describe the Go programming language" \
  -stream
```

## Development

```bash
# Run all tests
go test ./...

# Vet all packages
go vet ./...

# Lint proto definitions
cd rpc && buf lint

# Regenerate proto code
cd rpc && buf generate
```

## Related

- [tau-platform](https://github.com/tailored-agentic-units/tau-platform) — Ecosystem coordination hub
