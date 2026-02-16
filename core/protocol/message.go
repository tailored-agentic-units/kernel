package protocol

import "encoding/json"

// Role identifies the sender of a conversation message.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// ToolCall represents a tool invocation in conversation history.
// Fields are flat (ID, Name, Arguments) for direct use across the kernel.
// UnmarshalJSON transparently handles the nested LLM API format
// (function.name, function.arguments) so provider responses decode correctly.
type ToolCall struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// MarshalJSON serializes to the nested LLM API format ({type, function: {name, arguments}})
// ensuring round-trip fidelity with UnmarshalJSON for provider communication.
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

// UnmarshalJSON handles both the nested LLM API format ({function: {name, arguments}})
// and the flat kernel format ({name, arguments}). This allows provider responses to
// decode directly into the canonical ToolCall type.
func (tc *ToolCall) UnmarshalJSON(data []byte) error {
	var nested struct {
		ID       string `json:"id"`
		Function struct {
			Name      string `json:"name"`
			Arguments string `json:"arguments"`
		} `json:"function"`
	}
	if err := json.Unmarshal(data, &nested); err != nil {
		return err
	}

	if nested.Function.Name != "" {
		tc.ID = nested.ID
		tc.Name = nested.Function.Name
		tc.Arguments = nested.Function.Arguments
		return nil
	}

	type plain ToolCall
	return json.Unmarshal(data, (*plain)(tc))
}

// Message represents a single message in a conversation.
// Role indicates the sender, and Content can be a string for text or a
// structured object for multimodal content (e.g., vision arrays).
//
// For tool-calling conversations, assistant messages carry ToolCalls and
// tool result messages carry a ToolCallID that correlates back to the request.
type Message struct {
	Role       Role       `json:"role"`
	Content    any        `json:"content"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
}

// NewMessage creates a Message with the given role and content.
// Use struct literals directly when setting tool call fields.
//
// Example:
//
//	msg := protocol.NewMessage(protocol.RoleUser, "Hello, world!")
func NewMessage(role Role, content any) Message {
	return Message{Role: role, Content: content}
}

// InitMessages creates a single-element message slice from a role and content string.
// Convenience wrapper for the common pattern of initializing a conversation from a prompt.
func InitMessages(role Role, content string) []Message {
	return []Message{NewMessage(role, content)}
}
