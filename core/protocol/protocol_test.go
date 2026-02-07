package protocol_test

import (
	"testing"

	"github.com/tailored-agentic-units/kernel/core/protocol"
)

func TestProtocol_Constants(t *testing.T) {
	tests := []struct {
		name     string
		protocol protocol.Protocol
		expected string
	}{
		{"Chat", protocol.Chat, "chat"},
		{"Vision", protocol.Vision, "vision"},
		{"Tools", protocol.Tools, "tools"},
		{"Embeddings", protocol.Embeddings, "embeddings"},
		{"Audio", protocol.Audio, "audio"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.protocol) != tt.expected {
				t.Errorf("got %s, want %s", string(tt.protocol), tt.expected)
			}
		})
	}
}

func TestIsValid(t *testing.T) {
	tests := []struct {
		name     string
		protocol string
		expected bool
	}{
		{"chat valid", "chat", true},
		{"vision valid", "vision", true},
		{"tools valid", "tools", true},
		{"embeddings valid", "embeddings", true},
		{"audio valid", "audio", true},
		{"invalid", "invalid", false},
		{"empty string", "", false},
		{"uppercase", "CHAT", false},
		{"mixed case", "Chat", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := protocol.IsValid(tt.protocol)
			if result != tt.expected {
				t.Errorf("IsValid(%q) = %v, want %v", tt.protocol, result, tt.expected)
			}
		})
	}
}

func TestValidProtocols(t *testing.T) {
	result := protocol.ValidProtocols()

	expected := []protocol.Protocol{
		protocol.Chat,
		protocol.Vision,
		protocol.Tools,
		protocol.Embeddings,
		protocol.Audio,
	}

	if len(result) != len(expected) {
		t.Fatalf("got %d protocols, want %d", len(result), len(expected))
	}

	for i, p := range expected {
		if result[i] != p {
			t.Errorf("index %d: got %s, want %s", i, result[i], p)
		}
	}
}

func TestProtocolStrings(t *testing.T) {
	result := protocol.ProtocolStrings()
	expected := "chat, vision, tools, embeddings, audio"

	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestProtocol_SupportsStreaming(t *testing.T) {
	tests := []struct {
		name     string
		protocol protocol.Protocol
		expected bool
	}{
		{"Chat supports streaming", protocol.Chat, true},
		{"Vision supports streaming", protocol.Vision, true},
		{"Tools supports streaming", protocol.Tools, true},
		{"Embeddings does not support streaming", protocol.Embeddings, false},
		{"Audio does not support streaming", protocol.Audio, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.protocol.SupportsStreaming(); got != tt.expected {
				t.Errorf("SupportsStreaming() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNewMessage_StringContent(t *testing.T) {
	msg := protocol.NewMessage("user", "Hello, world!")

	if msg.Role != "user" {
		t.Errorf("got role %q, want %q", msg.Role, "user")
	}

	content, ok := msg.Content.(string)
	if !ok {
		t.Errorf("content is not a string")
	} else if content != "Hello, world!" {
		t.Errorf("got content %q, want %q", content, "Hello, world!")
	}
}

func TestNewMessage_StructuredContent(t *testing.T) {
	content := map[string]any{
		"type": "text",
		"text": "Hello",
	}

	msg := protocol.NewMessage("assistant", content)

	if msg.Role != "assistant" {
		t.Errorf("got role %q, want %q", msg.Role, "assistant")
	}

	if _, ok := msg.Content.(map[string]any); !ok {
		t.Errorf("content is not a map")
	}
}

func TestNewMessage_Roles(t *testing.T) {
	tests := []struct {
		name string
		role string
	}{
		{"user", "user"},
		{"assistant", "assistant"},
		{"system", "system"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := protocol.NewMessage(tt.role, "content")
			if msg.Role != tt.role {
				t.Errorf("got role %q, want %q", msg.Role, tt.role)
			}
		})
	}
}
