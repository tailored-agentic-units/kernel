package protocol_test

import (
	"encoding/json"
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
	msg := protocol.NewMessage(protocol.RoleUser, "Hello, world!")

	if msg.Role != protocol.RoleUser {
		t.Errorf("got role %q, want %q", msg.Role, protocol.RoleUser)
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

	msg := protocol.NewMessage(protocol.RoleAssistant, content)

	if msg.Role != protocol.RoleAssistant {
		t.Errorf("got role %q, want %q", msg.Role, protocol.RoleAssistant)
	}

	if _, ok := msg.Content.(map[string]any); !ok {
		t.Errorf("content is not a map")
	}
}

func TestRole_Constants(t *testing.T) {
	tests := []struct {
		name     string
		role     protocol.Role
		expected string
	}{
		{"system", protocol.RoleSystem, "system"},
		{"user", protocol.RoleUser, "user"},
		{"assistant", protocol.RoleAssistant, "assistant"},
		{"tool", protocol.RoleTool, "tool"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.role) != tt.expected {
				t.Errorf("got %q, want %q", tt.role, tt.expected)
			}
		})
	}
}

func TestMessage_ToolCallFields(t *testing.T) {
	toolCalls := []protocol.ToolCall{
		{ID: "call_1", Name: "get_weather", Arguments: `{"city":"NYC"}`},
		{ID: "call_2", Name: "get_time", Arguments: `{"tz":"UTC"}`},
	}

	msg := protocol.Message{
		Role:      protocol.RoleAssistant,
		ToolCalls: toolCalls,
	}

	if len(msg.ToolCalls) != 2 {
		t.Fatalf("got %d tool calls, want 2", len(msg.ToolCalls))
	}
	if msg.ToolCalls[0].Name != "get_weather" {
		t.Errorf("got name %q, want %q", msg.ToolCalls[0].Name, "get_weather")
	}
	if msg.ToolCalls[1].ID != "call_2" {
		t.Errorf("got id %q, want %q", msg.ToolCalls[1].ID, "call_2")
	}
}

func TestMessage_ToolCallID(t *testing.T) {
	msg := protocol.Message{
		Role:       protocol.RoleTool,
		Content:    `{"temp": 72}`,
		ToolCallID: "call_1",
	}

	if msg.ToolCallID != "call_1" {
		t.Errorf("got tool_call_id %q, want %q", msg.ToolCallID, "call_1")
	}
}

func TestMessage_JSON_OmitsEmptyToolFields(t *testing.T) {
	msg := protocol.NewMessage(protocol.RoleUser, "hello")

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if _, exists := raw["tool_call_id"]; exists {
		t.Error("tool_call_id should be omitted when empty")
	}
	if _, exists := raw["tool_calls"]; exists {
		t.Error("tool_calls should be omitted when empty")
	}
}

func TestMessage_JSON_IncludesToolFields(t *testing.T) {
	msg := protocol.Message{
		Role:      protocol.RoleAssistant,
		ToolCalls: []protocol.ToolCall{{ID: "call_1", Name: "fn", Arguments: "{}"}},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if _, exists := raw["tool_calls"]; !exists {
		t.Error("tool_calls should be present when populated")
	}
}

func TestToolCall_UnmarshalJSON_NestedFormat(t *testing.T) {
	data := `{
		"id": "call_123",
		"type": "function",
		"function": {
			"name": "get_weather",
			"arguments": "{\"location\":\"Boston\"}"
		}
	}`

	var tc protocol.ToolCall
	if err := json.Unmarshal([]byte(data), &tc); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	if tc.ID != "call_123" {
		t.Errorf("got ID %q, want %q", tc.ID, "call_123")
	}
	if tc.Name != "get_weather" {
		t.Errorf("got Name %q, want %q", tc.Name, "get_weather")
	}
	if tc.Arguments != `{"location":"Boston"}` {
		t.Errorf("got Arguments %q, want %q", tc.Arguments, `{"location":"Boston"}`)
	}
}

func TestToolCall_UnmarshalJSON_FlatFormat(t *testing.T) {
	data := `{
		"id": "call_456",
		"name": "search",
		"arguments": "{\"query\":\"test\"}"
	}`

	var tc protocol.ToolCall
	if err := json.Unmarshal([]byte(data), &tc); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	if tc.ID != "call_456" {
		t.Errorf("got ID %q, want %q", tc.ID, "call_456")
	}
	if tc.Name != "search" {
		t.Errorf("got Name %q, want %q", tc.Name, "search")
	}
	if tc.Arguments != `{"query":"test"}` {
		t.Errorf("got Arguments %q, want %q", tc.Arguments, `{"query":"test"}`)
	}
}

func TestToolCall_UnmarshalJSON_InvalidJSON(t *testing.T) {
	var tc protocol.ToolCall
	err := json.Unmarshal([]byte(`{invalid}`), &tc)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestToolCall_UnmarshalJSON_EmptyObject(t *testing.T) {
	var tc protocol.ToolCall
	if err := json.Unmarshal([]byte(`{}`), &tc); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	if tc.ID != "" || tc.Name != "" || tc.Arguments != "" {
		t.Errorf("expected empty ToolCall, got %+v", tc)
	}
}

func TestToolCall_UnmarshalJSON_InArray(t *testing.T) {
	data := `[
		{
			"id": "call_1",
			"type": "function",
			"function": {
				"name": "fn_a",
				"arguments": "{}"
			}
		},
		{
			"id": "call_2",
			"name": "fn_b",
			"arguments": "{\"x\":1}"
		}
	]`

	var calls []protocol.ToolCall
	if err := json.Unmarshal([]byte(data), &calls); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	if len(calls) != 2 {
		t.Fatalf("got %d calls, want 2", len(calls))
	}

	if calls[0].Name != "fn_a" {
		t.Errorf("call[0] Name = %q, want %q", calls[0].Name, "fn_a")
	}
	if calls[1].Name != "fn_b" {
		t.Errorf("call[1] Name = %q, want %q", calls[1].Name, "fn_b")
	}
}

func TestToolCall_MarshalJSON_NestedFormat(t *testing.T) {
	tc := protocol.ToolCall{
		ID:        "call_789",
		Name:      "get_weather",
		Arguments: `{"location":"Boston"}`,
	}

	data, err := json.Marshal(tc)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if raw["id"] != "call_789" {
		t.Errorf("got id %v, want %q", raw["id"], "call_789")
	}
	if raw["type"] != "function" {
		t.Errorf("got type %v, want %q", raw["type"], "function")
	}

	fn, ok := raw["function"].(map[string]any)
	if !ok {
		t.Fatalf("function field is not an object: %T", raw["function"])
	}
	if fn["name"] != "get_weather" {
		t.Errorf("got function.name %v, want %q", fn["name"], "get_weather")
	}
	if fn["arguments"] != `{"location":"Boston"}` {
		t.Errorf("got function.arguments %v, want %q", fn["arguments"], `{"location":"Boston"}`)
	}

	// Verify flat fields are NOT present at top level
	if _, exists := raw["name"]; exists {
		t.Error("name should not be at top level in nested format")
	}
	if _, exists := raw["arguments"]; exists {
		t.Error("arguments should not be at top level in nested format")
	}
}

func TestToolCall_MarshalJSON_RoundTrip(t *testing.T) {
	original := protocol.ToolCall{
		ID:        "call_rt",
		Name:      "search",
		Arguments: `{"query":"test","limit":10}`,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	var restored protocol.ToolCall
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	if restored.ID != original.ID {
		t.Errorf("ID: got %q, want %q", restored.ID, original.ID)
	}
	if restored.Name != original.Name {
		t.Errorf("Name: got %q, want %q", restored.Name, original.Name)
	}
	if restored.Arguments != original.Arguments {
		t.Errorf("Arguments: got %q, want %q", restored.Arguments, original.Arguments)
	}
}

func TestInitMessages(t *testing.T) {
	messages := protocol.InitMessages(protocol.RoleUser, "Hello")

	if len(messages) != 1 {
		t.Fatalf("got %d messages, want 1", len(messages))
	}

	if messages[0].Role != protocol.RoleUser {
		t.Errorf("got role %q, want %q", messages[0].Role, protocol.RoleUser)
	}

	content, ok := messages[0].Content.(string)
	if !ok {
		t.Fatalf("content is not string: %T", messages[0].Content)
	}
	if content != "Hello" {
		t.Errorf("got content %q, want %q", content, "Hello")
	}
}

func TestNewMessage_Roles(t *testing.T) {
	tests := []struct {
		name string
		role protocol.Role
	}{
		{"user", protocol.RoleUser},
		{"assistant", protocol.RoleAssistant},
		{"system", protocol.RoleSystem},
		{"tool", protocol.RoleTool},
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
