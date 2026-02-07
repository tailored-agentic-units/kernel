// Package messaging provides structured message primitives for inter-agent communication.
//
// This package implements the foundational messaging layer for agent orchestration,
// providing type-safe message construction, routing metadata, and protocol patterns
// for request-response and publish-subscribe communication.
//
// # Message Types
//
// The package defines four core message types:
//
//   - Request: Expects a response from the recipient
//   - Response: Reply to a previous request
//   - Notification: One-way message requiring no response
//   - Broadcast: Message sent to multiple recipients
//
// # Message Construction
//
// Messages are constructed using a fluent builder API:
//
//	msg := messaging.NewRequest("agent-a", "agent-b", taskData).
//	    Priority(messaging.PriorityHigh).
//	    Topic("task-queue").
//	    Headers(map[string]string{"correlation-id": "123"}).
//	    Build()
//
// # Message Metadata
//
// Each message includes:
//
//   - ID: UUIDv7 providing time-sortable unique identification
//   - Timestamp: Creation time for ordering and expiration
//   - Priority: Four levels (Low, Normal, High, Critical)
//   - Topic: Optional routing key for pub/sub patterns
//   - Headers: Extensible key-value metadata
//   - ReplyTo: Reference to original request for responses
//
// # Usage Example
//
//	// Create a request
//	request := messaging.NewRequest("worker", "processor", data).Build()
//
//	// Create a response
//	response := messaging.NewResponse("processor", "worker", request.ID, result).Build()
//
//	// Create a notification
//	notification := messaging.NewNotification("monitor", "logger", event).Build()
//
// # Integration
//
// This package is used by the hub package for agent-to-agent communication
// and serves as the foundation for higher-level orchestration patterns.
package messaging
