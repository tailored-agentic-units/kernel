package session_test

import (
	"sync"
	"testing"

	"github.com/tailored-agentic-units/kernel/core/protocol"
	"github.com/tailored-agentic-units/kernel/session"
)

func TestNew(t *testing.T) {
	s := session.NewMemorySession()

	if s.ID() == "" {
		t.Error("session ID should not be empty")
	}
	if len(s.Messages()) != 0 {
		t.Errorf("new session should have 0 messages, got %d", len(s.Messages()))
	}
}

func TestSession_ID_Unique(t *testing.T) {
	s1 := session.NewMemorySession()
	s2 := session.NewMemorySession()

	if s1.ID() == s2.ID() {
		t.Errorf("two sessions should have different IDs, both got %q", s1.ID())
	}
}

func TestSession_ID_Stable(t *testing.T) {
	s := session.NewMemorySession()

	id1 := s.ID()
	id2 := s.ID()

	if id1 != id2 {
		t.Errorf("same session returned different IDs: %q and %q", id1, id2)
	}
}

func TestSession_AddMessage_And_Messages(t *testing.T) {
	s := session.NewMemorySession()

	toolCalls := []protocol.ToolCall{
		{ID: "call_1", Name: "get_weather", Arguments: `{"city":"NYC"}`},
	}

	msg := protocol.Message{
		Role:      protocol.RoleAssistant,
		Content:   "Let me check the weather.",
		ToolCalls: toolCalls,
	}

	s.AddMessage(msg)
	msgs := s.Messages()

	if len(msgs) != 1 {
		t.Fatalf("got %d messages, want 1", len(msgs))
	}

	got := msgs[0]
	if got.Role != protocol.RoleAssistant {
		t.Errorf("got role %q, want %q", got.Role, protocol.RoleAssistant)
	}
	if got.Content != "Let me check the weather." {
		t.Errorf("got content %q, want %q", got.Content, "Let me check the weather.")
	}
	if len(got.ToolCalls) != 1 {
		t.Fatalf("got %d tool calls, want 1", len(got.ToolCalls))
	}
	if got.ToolCalls[0].Name != "get_weather" {
		t.Errorf("got tool call name %q, want %q", got.ToolCalls[0].Name, "get_weather")
	}
}

func TestSession_Messages_Order(t *testing.T) {
	s := session.NewMemorySession()

	roles := []protocol.Role{
		protocol.RoleSystem,
		protocol.RoleUser,
		protocol.RoleAssistant,
		protocol.RoleTool,
	}

	for _, role := range roles {
		s.AddMessage(protocol.NewMessage(role, string(role)))
	}

	msgs := s.Messages()
	if len(msgs) != len(roles) {
		t.Fatalf("got %d messages, want %d", len(msgs), len(roles))
	}

	for i, msg := range msgs {
		if msg.Role != roles[i] {
			t.Errorf("message %d: got role %q, want %q", i, msg.Role, roles[i])
		}
	}
}

func TestSession_Messages_Roles(t *testing.T) {
	tests := []struct {
		name string
		role protocol.Role
	}{
		{"system", protocol.RoleSystem},
		{"user", protocol.RoleUser},
		{"assistant", protocol.RoleAssistant},
		{"tool", protocol.RoleTool},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := session.NewMemorySession()
			s.AddMessage(protocol.NewMessage(tt.role, "content"))

			msgs := s.Messages()
			if len(msgs) != 1 {
				t.Fatalf("got %d messages, want 1", len(msgs))
			}
			if msgs[0].Role != tt.role {
				t.Errorf("got role %q, want %q", msgs[0].Role, tt.role)
			}
		})
	}
}

func TestSession_Messages_ToolCalls(t *testing.T) {
	s := session.NewMemorySession()

	// Assistant message with tool calls
	s.AddMessage(protocol.Message{
		Role: protocol.RoleAssistant,
		ToolCalls: []protocol.ToolCall{
			{ID: "call_1", Name: "get_weather", Arguments: `{"city":"NYC"}`},
		},
	})

	// Tool result correlated back
	s.AddMessage(protocol.Message{
		Role:       protocol.RoleTool,
		Content:    `{"temp": 72}`,
		ToolCallID: "call_1",
	})

	msgs := s.Messages()
	if len(msgs) != 2 {
		t.Fatalf("got %d messages, want 2", len(msgs))
	}

	if len(msgs[0].ToolCalls) != 1 {
		t.Fatalf("assistant message: got %d tool calls, want 1", len(msgs[0].ToolCalls))
	}
	if msgs[0].ToolCalls[0].ID != "call_1" {
		t.Errorf("got tool call ID %q, want %q", msgs[0].ToolCalls[0].ID, "call_1")
	}
	if msgs[1].ToolCallID != "call_1" {
		t.Errorf("got tool_call_id %q, want %q", msgs[1].ToolCallID, "call_1")
	}
}

func TestSession_Messages_DefensiveCopy(t *testing.T) {
	s := session.NewMemorySession()
	s.AddMessage(protocol.NewMessage(protocol.RoleUser, "hello"))
	s.AddMessage(protocol.NewMessage(protocol.RoleAssistant, "hi"))

	msgs := s.Messages()
	msgs[0] = protocol.NewMessage(protocol.RoleSystem, "tampered")
	msgs = append(msgs, protocol.NewMessage(protocol.RoleUser, "extra"))

	original := s.Messages()
	if len(original) != 2 {
		t.Fatalf("got %d messages, want 2", len(original))
	}
	if original[0].Role != protocol.RoleUser {
		t.Errorf("first message role was mutated: got %q, want %q", original[0].Role, protocol.RoleUser)
	}
}

func TestSession_Messages_ToolCalls_DefensiveCopy(t *testing.T) {
	s := session.NewMemorySession()
	s.AddMessage(protocol.Message{
		Role: protocol.RoleAssistant,
		ToolCalls: []protocol.ToolCall{
			{ID: "call_1", Name: "original", Arguments: "{}"},
		},
	})

	msgs := s.Messages()
	msgs[0].ToolCalls[0].Name = "tampered"
	msgs[0].ToolCalls = append(msgs[0].ToolCalls, protocol.ToolCall{ID: "call_2", Name: "extra"})

	original := s.Messages()
	if len(original[0].ToolCalls) != 1 {
		t.Fatalf("got %d tool calls, want 1", len(original[0].ToolCalls))
	}
	if original[0].ToolCalls[0].Name != "original" {
		t.Errorf("tool call name was mutated: got %q, want %q", original[0].ToolCalls[0].Name, "original")
	}
}

func TestSession_Clear(t *testing.T) {
	s := session.NewMemorySession()
	s.AddMessage(protocol.NewMessage(protocol.RoleUser, "hello"))
	s.AddMessage(protocol.NewMessage(protocol.RoleAssistant, "hi"))

	s.Clear()

	if len(s.Messages()) != 0 {
		t.Errorf("got %d messages after Clear, want 0", len(s.Messages()))
	}
}

func TestSession_Clear_ThenAdd(t *testing.T) {
	s := session.NewMemorySession()
	s.AddMessage(protocol.NewMessage(protocol.RoleUser, "first"))
	s.Clear()
	s.AddMessage(protocol.NewMessage(protocol.RoleUser, "second"))

	msgs := s.Messages()
	if len(msgs) != 1 {
		t.Fatalf("got %d messages, want 1", len(msgs))
	}
	if msgs[0].Content != "second" {
		t.Errorf("got content %q, want %q", msgs[0].Content, "second")
	}
}

func TestSession_Concurrent_AddMessage(t *testing.T) {
	s := session.NewMemorySession()
	const n = 100

	var wg sync.WaitGroup
	wg.Add(n)
	for range n {
		go func() {
			defer wg.Done()
			s.AddMessage(protocol.NewMessage(protocol.RoleUser, "msg"))
		}()
	}
	wg.Wait()

	msgs := s.Messages()
	if len(msgs) != n {
		t.Errorf("got %d messages, want %d", len(msgs), n)
	}
}

func TestSession_Concurrent_AddAndRead(t *testing.T) {
	s := session.NewMemorySession()
	const n = 100

	var wg sync.WaitGroup
	wg.Add(2 * n)

	for range n {
		go func() {
			defer wg.Done()
			s.AddMessage(protocol.NewMessage(protocol.RoleUser, "msg"))
		}()
		go func() {
			defer wg.Done()
			_ = s.Messages()
		}()
	}
	wg.Wait()
}

func TestSession_Concurrent_AddAndClear(t *testing.T) {
	s := session.NewMemorySession()
	const n = 100

	var wg sync.WaitGroup
	wg.Add(2 * n)

	for range n {
		go func() {
			defer wg.Done()
			s.AddMessage(protocol.NewMessage(protocol.RoleUser, "msg"))
		}()
		go func() {
			defer wg.Done()
			s.Clear()
		}()
	}
	wg.Wait()
}
