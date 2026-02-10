package session

import (
	"slices"
	"sync"

	"github.com/google/uuid"
	"github.com/tailored-agentic-units/kernel/core/protocol"
)

type memorySession struct {
	id       string
	messages []protocol.Message
	mu       sync.RWMutex
}

// NewMemorySession creates a Session backed by an in-memory slice.
// The session is assigned a unique UUIDv7 identifier.
func NewMemorySession() Session {
	return &memorySession{
		id: uuid.Must(uuid.NewV7()).String(),
	}
}

func (s *memorySession) ID() string {
	return s.id
}

func (s *memorySession) AddMessage(msg protocol.Message) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages = append(s.messages, msg)
}

func (s *memorySession) Messages() []protocol.Message {
	s.mu.RLock()
	defer s.mu.RUnlock()

	copied := make([]protocol.Message, len(s.messages))
	for i, msg := range s.messages {
		copied[i] = msg
		copied[i].ToolCalls = slices.Clone(msg.ToolCalls)
	}
	return copied
}

func (s *memorySession) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages = nil
}
