package protocol

// Message represents a single message in a conversation.
// The Role indicates the message sender (user, assistant, system),
// and Content can be either a string for text or a structured object
// for multimodal content (e.g., vision protocol with images).
type Message struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

// NewMessage creates a new Message with the specified role and content.
// Content can be a string for text or a structured object for multimodal inputs.
//
// Example:
//
//	msg := protocol.NewMessage("user", "Hello, world!")
//	visionMsg := protocol.NewMessage("user", []map[string]any{...})
func NewMessage(role string, content any) Message {
	return Message{Role: role, Content: content}
}
