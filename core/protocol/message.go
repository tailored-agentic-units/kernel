package protocol

// Role identifies the sender of a conversation message.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// ToolCall represents a tool invocation in conversation history.
// This is the canonical flat form used across the kernel. Distinct from
// response.ToolCall, which is a JSON deserialization struct for provider responses.
type ToolCall struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
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
