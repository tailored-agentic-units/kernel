package hub_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/tailored-agentic-units/kernel/agent/mock"
	"github.com/tailored-agentic-units/kernel/orchestrate/config"
	"github.com/tailored-agentic-units/kernel/orchestrate/hub"
	"github.com/tailored-agentic-units/kernel/orchestrate/messaging"
)

// Helper function to create a test hub
func createTestHub(t *testing.T) hub.Hub {
	ctx := context.Background()
	cfg := config.DefaultHubConfig()
	cfg.Name = "test-hub"
	return hub.New(ctx, cfg)
}

func TestHub_RegisterAgent(t *testing.T) {
	h := createTestHub(t)
	defer h.Shutdown(5 * time.Second)

	agent := mock.NewSimpleChatAgent("test-agent", "response")
	handler := func(ctx context.Context, msg *messaging.Message, msgCtx *hub.MessageContext) (*messaging.Message, error) {
		return nil, nil
	}

	err := h.RegisterAgent(agent, handler)
	if err != nil {
		t.Fatalf("RegisterAgent() error = %v", err)
	}

	// Verify metrics updated
	metrics := h.Metrics()
	if metrics.LocalAgents != 1 {
		t.Errorf("LocalAgents = %d, want 1", metrics.LocalAgents)
	}
}

func TestHub_RegisterAgent_Duplicate(t *testing.T) {
	h := createTestHub(t)
	defer h.Shutdown(5 * time.Second)

	agent := mock.NewSimpleChatAgent("test-agent", "response")
	handler := func(ctx context.Context, msg *messaging.Message, msgCtx *hub.MessageContext) (*messaging.Message, error) {
		return nil, nil
	}

	// First registration should succeed
	err := h.RegisterAgent(agent, handler)
	if err != nil {
		t.Fatalf("First RegisterAgent() error = %v", err)
	}

	// Duplicate registration should fail
	err = h.RegisterAgent(agent, handler)
	if err == nil {
		t.Error("RegisterAgent() should fail for duplicate registration")
	}
}

func TestHub_UnregisterAgent(t *testing.T) {
	h := createTestHub(t)
	defer h.Shutdown(5 * time.Second)

	agent := mock.NewSimpleChatAgent("test-agent", "response")
	handler := func(ctx context.Context, msg *messaging.Message, msgCtx *hub.MessageContext) (*messaging.Message, error) {
		return nil, nil
	}

	// Register agent
	err := h.RegisterAgent(agent, handler)
	if err != nil {
		t.Fatalf("RegisterAgent() error = %v", err)
	}

	// Unregister agent
	err = h.UnregisterAgent("test-agent")
	if err != nil {
		t.Fatalf("UnregisterAgent() error = %v", err)
	}

	// Verify metrics updated
	metrics := h.Metrics()
	if metrics.LocalAgents != 0 {
		t.Errorf("LocalAgents = %d, want 0", metrics.LocalAgents)
	}
}

func TestHub_UnregisterAgent_NotFound(t *testing.T) {
	h := createTestHub(t)
	defer h.Shutdown(5 * time.Second)

	err := h.UnregisterAgent("nonexistent-agent")
	if err == nil {
		t.Error("UnregisterAgent() should fail for nonexistent agent")
	}
}

func TestHub_Send(t *testing.T) {
	h := createTestHub(t)
	defer h.Shutdown(5 * time.Second)

	received := make(chan string, 1)

	agentA := mock.NewSimpleChatAgent("agent-a", "response-a")
	agentB := mock.NewSimpleChatAgent("agent-b", "response-b")

	handlerA := func(ctx context.Context, msg *messaging.Message, msgCtx *hub.MessageContext) (*messaging.Message, error) {
		return nil, nil
	}

	handlerB := func(ctx context.Context, msg *messaging.Message, msgCtx *hub.MessageContext) (*messaging.Message, error) {
		if data, ok := msg.Data.(string); ok {
			received <- data
		}
		return nil, nil
	}

	h.RegisterAgent(agentA, handlerA)
	h.RegisterAgent(agentB, handlerB)

	// Send message
	ctx := context.Background()
	err := h.Send(ctx, "agent-a", "agent-b", "test-message")
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	// Wait a bit for message processing
	time.Sleep(100 * time.Millisecond)

	// Verify received
	select {
	case data := <-received:
		if data != "test-message" {
			t.Errorf("Received data = %v, want %v", data, "test-message")
		}
	case <-time.After(time.Second):
		t.Error("Timeout waiting for message")
	}

	// Verify metrics
	metrics := h.Metrics()
	if metrics.MessagesSent != 1 {
		t.Errorf("MessagesSent = %d, want 1", metrics.MessagesSent)
	}
	if metrics.MessagesRecv != 1 {
		t.Errorf("MessagesRecv = %d, want 1", metrics.MessagesRecv)
	}
}

func TestHub_Send_AgentNotFound(t *testing.T) {
	h := createTestHub(t)
	defer h.Shutdown(5 * time.Second)

	agent := mock.NewSimpleChatAgent("agent-a", "response")
	handler := func(ctx context.Context, msg *messaging.Message, msgCtx *hub.MessageContext) (*messaging.Message, error) {
		return nil, nil
	}

	h.RegisterAgent(agent, handler)

	ctx := context.Background()
	err := h.Send(ctx, "agent-a", "nonexistent-agent", "test")
	if err == nil {
		t.Error("Send() should fail when destination agent not found")
	}
}

func TestHub_Request(t *testing.T) {
	h := createTestHub(t)
	defer h.Shutdown(5 * time.Second)

	agentA := mock.NewSimpleChatAgent("agent-a", "response-a")
	agentB := mock.NewSimpleChatAgent("agent-b", "response-b")

	handlerA := func(ctx context.Context, msg *messaging.Message, msgCtx *hub.MessageContext) (*messaging.Message, error) {
		return nil, nil
	}

	handlerB := func(ctx context.Context, msg *messaging.Message, msgCtx *hub.MessageContext) (*messaging.Message, error) {
		// Echo back with modification
		if data, ok := msg.Data.(string); ok {
			return messaging.NewResponse("agent-b", msg.From, msg.ID, "processed: "+data).Build(), nil
		}
		return nil, nil
	}

	h.RegisterAgent(agentA, handlerA)
	h.RegisterAgent(agentB, handlerB)

	// Send request
	ctx := context.Background()
	response, err := h.Request(ctx, "agent-a", "agent-b", "task")
	if err != nil {
		t.Fatalf("Request() error = %v", err)
	}

	if data, ok := response.Data.(string); !ok || data != "processed: task" {
		t.Errorf("Response data = %v, want %v", response.Data, "processed: task")
	}
}

func TestHub_Request_Timeout(t *testing.T) {
	h := createTestHub(t)
	defer h.Shutdown(5 * time.Second)

	agentA := mock.NewSimpleChatAgent("agent-a", "response-a")
	agentB := mock.NewSimpleChatAgent("agent-b", "response-b")

	handlerA := func(ctx context.Context, msg *messaging.Message, msgCtx *hub.MessageContext) (*messaging.Message, error) {
		return nil, nil
	}

	// Handler B never responds
	handlerB := func(ctx context.Context, msg *messaging.Message, msgCtx *hub.MessageContext) (*messaging.Message, error) {
		return nil, nil
	}

	h.RegisterAgent(agentA, handlerA)
	h.RegisterAgent(agentB, handlerB)

	// Send request with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := h.Request(ctx, "agent-a", "agent-b", "task")
	if err == nil {
		t.Error("Request() should timeout when no response received")
	}
}

func TestHub_Request_AgentNotFound(t *testing.T) {
	h := createTestHub(t)
	defer h.Shutdown(5 * time.Second)

	agent := mock.NewSimpleChatAgent("agent-a", "response")
	handler := func(ctx context.Context, msg *messaging.Message, msgCtx *hub.MessageContext) (*messaging.Message, error) {
		return nil, nil
	}

	h.RegisterAgent(agent, handler)

	ctx := context.Background()
	_, err := h.Request(ctx, "agent-a", "nonexistent-agent", "task")
	if err == nil {
		t.Error("Request() should fail when destination agent not found")
	}
}

func TestHub_Broadcast(t *testing.T) {
	h := createTestHub(t)
	defer h.Shutdown(5 * time.Second)

	received := make(chan string, 3)

	agentA := mock.NewSimpleChatAgent("agent-a", "response-a")
	agentB := mock.NewSimpleChatAgent("agent-b", "response-b")
	agentC := mock.NewSimpleChatAgent("agent-c", "response-c")

	handlerA := func(ctx context.Context, msg *messaging.Message, msgCtx *hub.MessageContext) (*messaging.Message, error) {
		return nil, nil
	}

	makeReceiver := func() hub.MessageHandler {
		return func(ctx context.Context, msg *messaging.Message, msgCtx *hub.MessageContext) (*messaging.Message, error) {
			if data, ok := msg.Data.(string); ok {
				received <- data
			}
			return nil, nil
		}
	}

	h.RegisterAgent(agentA, handlerA)
	h.RegisterAgent(agentB, makeReceiver())
	h.RegisterAgent(agentC, makeReceiver())

	// Broadcast message
	ctx := context.Background()
	err := h.Broadcast(ctx, "agent-a", "broadcast-message")
	if err != nil {
		t.Fatalf("Broadcast() error = %v", err)
	}

	// Wait for messages
	time.Sleep(100 * time.Millisecond)

	// Should receive 2 messages (agent-b and agent-c, not agent-a)
	receivedCount := 0
	timeout := time.After(time.Second)
	for receivedCount < 2 {
		select {
		case data := <-received:
			if data != "broadcast-message" {
				t.Errorf("Received data = %v, want %v", data, "broadcast-message")
			}
			receivedCount++
		case <-timeout:
			t.Errorf("Received %d messages, want 2", receivedCount)
			return
		}
	}
}

func TestHub_Subscribe_Publish(t *testing.T) {
	h := createTestHub(t)
	defer h.Shutdown(5 * time.Second)

	received := make(chan string, 2)

	agentA := mock.NewSimpleChatAgent("agent-a", "response-a")
	agentB := mock.NewSimpleChatAgent("agent-b", "response-b")
	agentC := mock.NewSimpleChatAgent("agent-c", "response-c")

	handlerA := func(ctx context.Context, msg *messaging.Message, msgCtx *hub.MessageContext) (*messaging.Message, error) {
		return nil, nil
	}

	makeSubscriber := func() hub.MessageHandler {
		return func(ctx context.Context, msg *messaging.Message, msgCtx *hub.MessageContext) (*messaging.Message, error) {
			if data, ok := msg.Data.(string); ok {
				received <- data
			}
			return nil, nil
		}
	}

	h.RegisterAgent(agentA, handlerA)
	h.RegisterAgent(agentB, makeSubscriber())
	h.RegisterAgent(agentC, makeSubscriber())

	// Subscribe agents B and C to topic
	err := h.Subscribe("agent-b", "test-topic")
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	err = h.Subscribe("agent-c", "test-topic")
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	// Publish message
	ctx := context.Background()
	err = h.Publish(ctx, "agent-a", "test-topic", "topic-message")
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	// Wait for messages
	time.Sleep(100 * time.Millisecond)

	// Should receive 2 messages
	receivedCount := 0
	timeout := time.After(time.Second)
	for receivedCount < 2 {
		select {
		case data := <-received:
			if data != "topic-message" {
				t.Errorf("Received data = %v, want %v", data, "topic-message")
			}
			receivedCount++
		case <-timeout:
			t.Errorf("Received %d messages, want 2", receivedCount)
			return
		}
	}
}

func TestHub_Subscribe_AgentNotFound(t *testing.T) {
	h := createTestHub(t)
	defer h.Shutdown(5 * time.Second)

	err := h.Subscribe("nonexistent-agent", "test-topic")
	if err == nil {
		t.Error("Subscribe() should fail for nonexistent agent")
	}
}

func TestHub_Publish_NoSubscribers(t *testing.T) {
	h := createTestHub(t)
	defer h.Shutdown(5 * time.Second)

	agent := mock.NewSimpleChatAgent("agent-a", "response")
	handler := func(ctx context.Context, msg *messaging.Message, msgCtx *hub.MessageContext) (*messaging.Message, error) {
		return nil, nil
	}

	h.RegisterAgent(agent, handler)

	// Publish to topic with no subscribers (should not error)
	ctx := context.Background()
	err := h.Publish(ctx, "agent-a", "empty-topic", "message")
	if err != nil {
		t.Errorf("Publish() error = %v, should succeed with no subscribers", err)
	}
}

func TestHub_UnregisterAgent_CleansUpSubscriptions(t *testing.T) {
	h := createTestHub(t)
	defer h.Shutdown(5 * time.Second)

	received := make(chan string, 1)

	agentA := mock.NewSimpleChatAgent("agent-a", "response-a")
	agentB := mock.NewSimpleChatAgent("agent-b", "response-b")

	handlerA := func(ctx context.Context, msg *messaging.Message, msgCtx *hub.MessageContext) (*messaging.Message, error) {
		return nil, nil
	}

	handlerB := func(ctx context.Context, msg *messaging.Message, msgCtx *hub.MessageContext) (*messaging.Message, error) {
		if data, ok := msg.Data.(string); ok {
			received <- data
		}
		return nil, nil
	}

	h.RegisterAgent(agentA, handlerA)
	h.RegisterAgent(agentB, handlerB)

	// Subscribe agent-b
	err := h.Subscribe("agent-b", "test-topic")
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	// Unregister agent-b
	err = h.UnregisterAgent("agent-b")
	if err != nil {
		t.Fatalf("UnregisterAgent() error = %v", err)
	}

	// Publish to topic (agent-b should not receive)
	ctx := context.Background()
	err = h.Publish(ctx, "agent-a", "test-topic", "message")
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	// Wait and verify no message received
	time.Sleep(100 * time.Millisecond)
	select {
	case <-received:
		t.Error("Unregistered agent should not receive messages")
	default:
		// Expected - no message received
	}
}

func TestHub_MessageContext(t *testing.T) {
	h := createTestHub(t)
	defer h.Shutdown(5 * time.Second)

	contextReceived := make(chan *hub.MessageContext, 1)

	agentA := mock.NewSimpleChatAgent("agent-a", "response-a")
	agentB := mock.NewSimpleChatAgent("agent-b", "response-b")

	handlerA := func(ctx context.Context, msg *messaging.Message, msgCtx *hub.MessageContext) (*messaging.Message, error) {
		return nil, nil
	}

	handlerB := func(ctx context.Context, msg *messaging.Message, msgCtx *hub.MessageContext) (*messaging.Message, error) {
		contextReceived <- msgCtx
		return nil, nil
	}

	h.RegisterAgent(agentA, handlerA)
	h.RegisterAgent(agentB, handlerB)

	// Send message
	ctx := context.Background()
	err := h.Send(ctx, "agent-a", "agent-b", "test")
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	// Wait for context
	time.Sleep(100 * time.Millisecond)

	select {
	case msgCtx := <-contextReceived:
		if msgCtx.HubName != "test-hub" {
			t.Errorf("MessageContext.HubName = %v, want %v", msgCtx.HubName, "test-hub")
		}
		if msgCtx.Agent.ID() != "agent-b" {
			t.Errorf("MessageContext.Agent.ID() = %v, want %v", msgCtx.Agent.ID(), "agent-b")
		}
	case <-time.After(time.Second):
		t.Error("Timeout waiting for message context")
	}
}

func TestHub_HandlerError(t *testing.T) {
	h := createTestHub(t)
	defer h.Shutdown(5 * time.Second)

	agentA := mock.NewSimpleChatAgent("agent-a", "response-a")
	agentB := mock.NewSimpleChatAgent("agent-b", "response-b")

	handlerA := func(ctx context.Context, msg *messaging.Message, msgCtx *hub.MessageContext) (*messaging.Message, error) {
		return nil, nil
	}

	handlerB := func(ctx context.Context, msg *messaging.Message, msgCtx *hub.MessageContext) (*messaging.Message, error) {
		return nil, errors.New("handler error")
	}

	h.RegisterAgent(agentA, handlerA)
	h.RegisterAgent(agentB, handlerB)

	// Send message (should not panic even if handler errors)
	ctx := context.Background()
	err := h.Send(ctx, "agent-a", "agent-b", "test")
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	// Wait for message processing
	time.Sleep(100 * time.Millisecond)

	// Hub should still be operational
	metrics := h.Metrics()
	if metrics.MessagesSent != 1 {
		t.Errorf("MessagesSent = %d, want 1", metrics.MessagesSent)
	}
}

func TestHub_Metrics(t *testing.T) {
	h := createTestHub(t)
	defer h.Shutdown(5 * time.Second)

	// Initial metrics should be zero
	metrics := h.Metrics()
	if metrics.LocalAgents != 0 {
		t.Errorf("Initial LocalAgents = %d, want 0", metrics.LocalAgents)
	}
	if metrics.MessagesSent != 0 {
		t.Errorf("Initial MessagesSent = %d, want 0", metrics.MessagesSent)
	}
	if metrics.MessagesRecv != 0 {
		t.Errorf("Initial MessagesRecv = %d, want 0", metrics.MessagesRecv)
	}

	// Register agents
	agentA := mock.NewSimpleChatAgent("agent-a", "response-a")
	agentB := mock.NewSimpleChatAgent("agent-b", "response-b")

	handlerA := func(ctx context.Context, msg *messaging.Message, msgCtx *hub.MessageContext) (*messaging.Message, error) {
		return nil, nil
	}

	handlerB := func(ctx context.Context, msg *messaging.Message, msgCtx *hub.MessageContext) (*messaging.Message, error) {
		return nil, nil
	}

	h.RegisterAgent(agentA, handlerA)
	h.RegisterAgent(agentB, handlerB)

	metrics = h.Metrics()
	if metrics.LocalAgents != 2 {
		t.Errorf("LocalAgents after registration = %d, want 2", metrics.LocalAgents)
	}

	// Send message
	ctx := context.Background()
	h.Send(ctx, "agent-a", "agent-b", "test")

	time.Sleep(100 * time.Millisecond)

	metrics = h.Metrics()
	if metrics.MessagesSent != 1 {
		t.Errorf("MessagesSent = %d, want 1", metrics.MessagesSent)
	}
	if metrics.MessagesRecv != 1 {
		t.Errorf("MessagesRecv = %d, want 1", metrics.MessagesRecv)
	}
}

func TestHub_Shutdown(t *testing.T) {
	h := createTestHub(t)

	agent := mock.NewSimpleChatAgent("test-agent", "response")
	handler := func(ctx context.Context, msg *messaging.Message, msgCtx *hub.MessageContext) (*messaging.Message, error) {
		return nil, nil
	}

	h.RegisterAgent(agent, handler)

	// Shutdown should complete successfully
	err := h.Shutdown(5 * time.Second)
	if err != nil {
		t.Errorf("Shutdown() error = %v", err)
	}
}

func TestHub_Shutdown_Timeout(t *testing.T) {
	h := createTestHub(t)

	agent := mock.NewSimpleChatAgent("test-agent", "response")
	handler := func(ctx context.Context, msg *messaging.Message, msgCtx *hub.MessageContext) (*messaging.Message, error) {
		// Block forever to test timeout
		<-make(chan struct{})
		return nil, nil
	}

	h.RegisterAgent(agent, handler)

	// This test would timeout if the shutdown doesn't handle it properly
	// We'll use a very short timeout to test the timeout path
	err := h.Shutdown(1 * time.Nanosecond)
	if err == nil {
		t.Error("Shutdown() should timeout with very short duration")
	}
}
