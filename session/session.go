// Package session manages conversation history for the kernel runtime loop.
package session

import (
	"github.com/tailored-agentic-units/kernel/core/protocol"
)

// Session holds an ordered sequence of conversation messages. Implementations
// must be safe for concurrent use.
type Session interface {
	// ID returns the unique session identifier.
	ID() string
	// AddMessage appends a message to the conversation history.
	AddMessage(msg protocol.Message)
	// Messages returns a defensive copy of the conversation history.
	Messages() []protocol.Message
	// Clear resets the conversation history.
	Clear()
}
