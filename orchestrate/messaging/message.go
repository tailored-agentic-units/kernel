package messaging

import (
	"fmt"
	"maps"
	"time"

	"github.com/google/uuid"
)

type MessageType string

const (
	MessageTypeRequest      MessageType = "request"
	MessageTypeResponse     MessageType = "response"
	MessageTypeNotification MessageType = "notification"
	MessageTypeBroadcast    MessageType = "broadcast"
)

type Priority int

const (
	PriorityLow Priority = iota
	PriorityNormal
	PriorityHigh
	PriorityCritical
)

type Message struct {
	ID        string            `json:"id"`
	From      string            `json:"from"`
	To        string            `json:"to"`
	Type      MessageType       `json:"type"`
	Data      any               `json:"data"`
	ReplyTo   string            `json:"reply_to,omitempty"`
	Topic     string            `json:"topic,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
	Priority  Priority          `json:"priority,omitempty"`
	Headers   map[string]string `json:"headers,omitempty"`
}

func (msg *Message) IsRequest() bool {
	return msg.Type == MessageTypeRequest
}

func (msg *Message) IsResponse() bool {
	return msg.Type == MessageTypeResponse
}

func (msg *Message) IsBroadcast() bool {
	return msg.Type == MessageTypeBroadcast
}

func (msg *Message) Clone() *Message {
	clone := *msg
	clone.Headers = maps.Clone(msg.Headers)
	return &clone
}

func (msg *Message) String() string {
	return fmt.Sprintf(
		"Message{ID: %s, From: %s, To: %s, Type: %s, Topic: %s}",
		msg.ID,
		msg.From,
		msg.To,
		msg.Type,
		msg.Topic,
	)
}

func generateID() string {
	return uuid.Must(uuid.NewV7()).String()
}
