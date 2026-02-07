package messaging_test

import (
	"testing"
	"time"

	"github.com/tailored-agentic-units/kernel/orchestrate/messaging"
)

func TestMessage_Builders(t *testing.T) {
	tests := []struct {
		name     string
		builder  func() *messaging.Message
		wantType messaging.MessageType
		wantFrom string
		wantTo   string
	}{
		{
			name: "NewRequest",
			builder: func() *messaging.Message {
				return messaging.NewRequest("agent-a", "agent-b", "test-data").Build()
			},
			wantType: messaging.MessageTypeRequest,
			wantFrom: "agent-a",
			wantTo:   "agent-b",
		},
		{
			name: "NewResponse",
			builder: func() *messaging.Message {
				return messaging.NewResponse("agent-b", "agent-a", "msg-123", "result-data").Build()
			},
			wantType: messaging.MessageTypeResponse,
			wantFrom: "agent-b",
			wantTo:   "agent-a",
		},
		{
			name: "NewNotification",
			builder: func() *messaging.Message {
				return messaging.NewNotification("agent-a", "agent-b", "update-data").Build()
			},
			wantType: messaging.MessageTypeNotification,
			wantFrom: "agent-a",
			wantTo:   "agent-b",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.builder()

			if msg.Type != tt.wantType {
				t.Errorf("Type = %v, want %v", msg.Type, tt.wantType)
			}
			if msg.From != tt.wantFrom {
				t.Errorf("From = %v, want %v", msg.From, tt.wantFrom)
			}
			if msg.To != tt.wantTo {
				t.Errorf("To = %v, want %v", msg.To, tt.wantTo)
			}
			if msg.ID == "" {
				t.Error("ID should not be empty")
			}
			if msg.Timestamp.IsZero() {
				t.Error("Timestamp should not be zero")
			}
			if msg.Priority != messaging.PriorityNormal {
				t.Errorf("Priority = %v, want %v", msg.Priority, messaging.PriorityNormal)
			}
		})
	}
}

func TestMessage_NewResponse_ReplyTo(t *testing.T) {
	replyToID := "original-message-id"
	msg := messaging.NewResponse("agent-b", "agent-a", replyToID, "result").Build()

	if msg.ReplyTo != replyToID {
		t.Errorf("ReplyTo = %v, want %v", msg.ReplyTo, replyToID)
	}
	if msg.Type != messaging.MessageTypeResponse {
		t.Errorf("Type = %v, want %v", msg.Type, messaging.MessageTypeResponse)
	}
}

func TestMessage_FluentAPI(t *testing.T) {
	headers := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	msg := messaging.NewRequest("agent-a", "agent-b", "data").
		Priority(messaging.PriorityHigh).
		Topic("test-topic").
		Headers(headers).
		Build()

	if msg.Priority != messaging.PriorityHigh {
		t.Errorf("Priority = %v, want %v", msg.Priority, messaging.PriorityHigh)
	}
	if msg.Topic != "test-topic" {
		t.Errorf("Topic = %v, want %v", msg.Topic, "test-topic")
	}
	if msg.Headers["key1"] != "value1" {
		t.Errorf("Headers[key1] = %v, want %v", msg.Headers["key1"], "value1")
	}
}

func TestMessage_IsRequest(t *testing.T) {
	tests := []struct {
		name string
		msg  *messaging.Message
		want bool
	}{
		{
			name: "Request message",
			msg:  messaging.NewRequest("a", "b", "data").Build(),
			want: true,
		},
		{
			name: "Response message",
			msg:  messaging.NewResponse("b", "a", "123", "data").Build(),
			want: false,
		},
		{
			name: "Notification message",
			msg:  messaging.NewNotification("a", "b", "data").Build(),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.msg.IsRequest(); got != tt.want {
				t.Errorf("IsRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessage_IsResponse(t *testing.T) {
	tests := []struct {
		name string
		msg  *messaging.Message
		want bool
	}{
		{
			name: "Response message",
			msg:  messaging.NewResponse("b", "a", "123", "data").Build(),
			want: true,
		},
		{
			name: "Request message",
			msg:  messaging.NewRequest("a", "b", "data").Build(),
			want: false,
		},
		{
			name: "Notification message",
			msg:  messaging.NewNotification("a", "b", "data").Build(),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.msg.IsResponse(); got != tt.want {
				t.Errorf("IsResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessage_IsBroadcast(t *testing.T) {
	tests := []struct {
		name string
		msg  *messaging.Message
		want bool
	}{
		{
			name: "Broadcast message",
			msg:  messaging.NewMessage("a", "b", messaging.MessageTypeBroadcast, "data").Build(),
			want: true,
		},
		{
			name: "Request message",
			msg:  messaging.NewRequest("a", "b", "data").Build(),
			want: false,
		},
		{
			name: "Notification message",
			msg:  messaging.NewNotification("a", "b", "data").Build(),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.msg.IsBroadcast(); got != tt.want {
				t.Errorf("IsBroadcast() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessage_Clone(t *testing.T) {
	original := messaging.NewRequest("agent-a", "agent-b", "test-data").
		Priority(messaging.PriorityHigh).
		Topic("test-topic").
		Headers(map[string]string{
			"key1": "value1",
			"key2": "value2",
		}).
		Build()

	clone := original.Clone()

	// Verify clone has same values
	if clone.ID != original.ID {
		t.Errorf("Clone ID = %v, want %v", clone.ID, original.ID)
	}
	if clone.From != original.From {
		t.Errorf("Clone From = %v, want %v", clone.From, original.From)
	}
	if clone.To != original.To {
		t.Errorf("Clone To = %v, want %v", clone.To, original.To)
	}
	if clone.Type != original.Type {
		t.Errorf("Clone Type = %v, want %v", clone.Type, original.Type)
	}
	if clone.Topic != original.Topic {
		t.Errorf("Clone Topic = %v, want %v", clone.Topic, original.Topic)
	}
	if clone.Priority != original.Priority {
		t.Errorf("Clone Priority = %v, want %v", clone.Priority, original.Priority)
	}

	// Verify headers are deep copied
	if clone.Headers["key1"] != original.Headers["key1"] {
		t.Errorf("Clone Headers[key1] = %v, want %v", clone.Headers["key1"], original.Headers["key1"])
	}

	// Modify clone's headers and verify original is unchanged
	clone.Headers["key1"] = "modified"
	if original.Headers["key1"] == "modified" {
		t.Error("Modifying clone headers modified original headers (not deep copied)")
	}
}

func TestMessage_Clone_NilHeaders(t *testing.T) {
	original := messaging.NewRequest("agent-a", "agent-b", "test-data").Build()
	original.Headers = nil

	clone := original.Clone()

	if clone.Headers != nil {
		t.Errorf("Clone Headers = %v, want nil", clone.Headers)
	}
}

func TestMessage_String(t *testing.T) {
	msg := messaging.NewRequest("agent-a", "agent-b", "test-data").
		Topic("test-topic").
		Build()

	str := msg.String()

	// Verify string contains key information
	if str == "" {
		t.Error("String() returned empty string")
	}

	// Check for key components (exact format may vary)
	contains := []string{msg.ID, msg.From, msg.To, string(msg.Type)}
	for _, want := range contains {
		found := false
		for i := 0; i < len(str)-len(want); i++ {
			if str[i:i+len(want)] == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("String() = %v, should contain %v", str, want)
		}
	}
}

func TestMessage_IDUniqueness(t *testing.T) {
	// Generate multiple messages and verify IDs are unique
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		msg := messaging.NewRequest("agent-a", "agent-b", "data").Build()
		if ids[msg.ID] {
			t.Errorf("Duplicate ID generated: %s", msg.ID)
		}
		ids[msg.ID] = true
	}
}

func TestMessage_TimestampSet(t *testing.T) {
	before := time.Now()
	msg := messaging.NewRequest("agent-a", "agent-b", "data").Build()
	after := time.Now()

	if msg.Timestamp.Before(before) || msg.Timestamp.After(after) {
		t.Errorf("Timestamp = %v, should be between %v and %v", msg.Timestamp, before, after)
	}
}

func TestPriority_Values(t *testing.T) {
	tests := []struct {
		name     string
		priority messaging.Priority
		expected int
	}{
		{"PriorityLow", messaging.PriorityLow, 0},
		{"PriorityNormal", messaging.PriorityNormal, 1},
		{"PriorityHigh", messaging.PriorityHigh, 2},
		{"PriorityCritical", messaging.PriorityCritical, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.priority) != tt.expected {
				t.Errorf("Priority value = %d, want %d", int(tt.priority), tt.expected)
			}
		})
	}
}

func TestMessageType_Values(t *testing.T) {
	tests := []struct {
		name     string
		msgType  messaging.MessageType
		expected string
	}{
		{"MessageTypeRequest", messaging.MessageTypeRequest, "request"},
		{"MessageTypeResponse", messaging.MessageTypeResponse, "response"},
		{"MessageTypeNotification", messaging.MessageTypeNotification, "notification"},
		{"MessageTypeBroadcast", messaging.MessageTypeBroadcast, "broadcast"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.msgType) != tt.expected {
				t.Errorf("MessageType value = %s, want %s", string(tt.msgType), tt.expected)
			}
		})
	}
}
