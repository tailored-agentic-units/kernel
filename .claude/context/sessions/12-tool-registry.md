# 12 - Tool Registry Interface and Execution

## Summary

Implemented the tools subsystem registry — a global singleton catalog of tool handlers following the `agent/providers/registry.go` pattern. Resolved the tool type duplication Known Gap by establishing `protocol.Tool` as the canonical tool definition type in `core/protocol`, replacing both `agent.Tool` and `providers.ToolDefinition`.

## Key Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Tool type location | `core/protocol.Tool` | Define once at lowest level; both agent and tools consume directly. Resolves three-way duplication. |
| Registry pattern | Global singleton with package-level functions | Matches providers registry pattern. Catalog of available tools, not the active set — kernel selects subset per turn. |
| Register vs Replace | Separate functions with distinct semantics | Register rejects duplicates; Replace rejects missing tools. Enables future permission boundaries. |
| No Registry interface | Package-level functions only | No scenario requires multiple registry instances. Consistent with providers. |

## Files Modified

- `core/protocol/tool.go` — new canonical Tool type
- `agent/agent.go` — removed Tool struct, updated interface and implementation to use protocol.Tool
- `agent/doc.go` — updated documentation references
- `agent/mock/agent.go` — updated Tools() signature
- `agent/providers/data.go` — removed ToolDefinition, updated ToolsData
- `agent/request/tools.go` — updated to use protocol.Tool
- `agent/agent_test.go` — updated test types
- `agent/client/client_test.go` — updated test types
- `agent/providers/base_test.go` — updated test types
- `cmd/prompt-agent/main.go` — updated to use protocol.Tool
- `tools/errors.go` — new sentinel errors
- `tools/registry.go` — new global tool registry
- `tools/registry_test.go` — new tests (13 tests, 100% coverage)
- `tools/README.md` — updated with usage documentation
- `_project/README.md` — removed tool type duplication Known Gap

## Patterns Established

- Global singleton registry pattern for subsystem catalogs (providers, tools)
- Register/Replace separation for future permission modeling
- Canonical types in `core/protocol` consumed directly by higher-level packages

## Validation Results

- `go vet ./...` — passes
- `go test ./...` — all tests pass
- `go mod tidy` — no changes
- Tools package coverage: 100%
