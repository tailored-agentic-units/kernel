// Package hub provides a central coordination primitive for agent communication and routing.
//
// The hub acts as a message broker and coordination point for agents, implementing
// both point-to-point and publish-subscribe messaging patterns with built-in metrics,
// lifecycle management, and error handling.
//
// # Core Capabilities
//
// The hub provides three primary communication patterns:
//
//   - Point-to-Point: Direct message delivery between agents
//   - Request-Response: Synchronous communication with timeout support
//   - Publish-Subscribe: Topic-based message distribution
//
// # Agent Registration
//
// Agents register with a hub to participate in message routing:
//
//	hub := hub.New(ctx, config.DefaultHubConfig())
//	agent := agent.New(...)
//
//	handler := func(ctx context.Context, msg *messaging.Message, msgCtx *hub.MessageContext) (*messaging.Message, error) {
//	    // Process message
//	    return response, nil
//	}
//
//	err := hub.RegisterAgent(agent, handler)
//
// # Communication Patterns
//
// Point-to-Point Messaging:
//
//	err := hub.Send(ctx, "sender-id", "receiver-id", data)
//
// Request-Response:
//
//	response, err := hub.Request(ctx, "requester-id", "processor-id", request)
//
// Broadcast:
//
//	err := hub.Broadcast(ctx, "sender-id", announcement)
//
// Publish-Subscribe:
//
//	hub.Subscribe("subscriber-id", "events.user.created")
//	hub.Publish(ctx, "publisher-id", "events.user.created", event)
//
// # Message Handlers
//
// Message handlers receive messages and optionally return responses:
//
//	handler := func(ctx context.Context, msg *messaging.Message, msgCtx *hub.MessageContext) (*messaging.Message, error) {
//	    // Access hub context
//	    log.Printf("Hub: %s, Agent: %s", msgCtx.HubName, msgCtx.Agent.ID())
//
//	    // Process message based on type
//	    if msg.IsRequest() {
//	        // Return response
//	        return messaging.NewResponse(msgCtx.Agent.ID(), msg.From, msg.ID, result).Build(), nil
//	    }
//
//	    return nil, nil
//	}
//
// # Lifecycle Management
//
// Hubs support graceful shutdown with timeout:
//
//	err := hub.Shutdown(30 * time.Second)
//
// # Metrics
//
// The hub tracks operational metrics:
//
//	metrics := hub.Metrics()
//	fmt.Printf("Agents: %d, Sent: %d, Received: %d",
//	    metrics.LocalAgents, metrics.MessagesSent, metrics.MessagesRecv)
//
// # Concurrency
//
// The hub is fully concurrent and thread-safe:
//
//   - Message routing runs in a dedicated goroutine
//   - Handlers execute concurrently per message
//   - Agent registration/unregistration is synchronized
//   - Subscription management is thread-safe
//
// # Integration
//
// The hub integrates with tau-core agent.Agent interface and uses the
// messaging package for message primitives. It serves as the foundation
// for higher-level orchestration patterns including sequential chains,
// parallel execution, and stateful workflows.
package hub
