package messaging

import "time"

type MessageBuilder struct {
	message *Message
}

func NewMessage(from, to string, messageType MessageType, data any) *MessageBuilder {
	return &MessageBuilder{
		message: &Message{
			ID:        generateID(),
			From:      from,
			To:        to,
			Type:      messageType,
			Data:      data,
			Timestamp: time.Now(),
			Priority:  PriorityNormal,
		},
	}
}

func NewRequest(from, to string, data any) *MessageBuilder {
	return NewMessage(from, to, MessageTypeRequest, data)
}

func NewResponse(from, to, replyTo string, data any) *MessageBuilder {
	return NewMessage(from, to, MessageTypeResponse, data).ReplyTo(replyTo)
}

func NewNotification(from, to string, data any) *MessageBuilder {
	return NewMessage(from, to, MessageTypeNotification, data)
}

func (mb *MessageBuilder) ReplyTo(replyTo string) *MessageBuilder {
	mb.message.ReplyTo = replyTo
	return mb
}

func (mb *MessageBuilder) Topic(topic string) *MessageBuilder {
	mb.message.Topic = topic
	return mb
}

func (mb *MessageBuilder) Priority(priority Priority) *MessageBuilder {
	mb.message.Priority = priority
	return mb
}

func (mb *MessageBuilder) Headers(headers map[string]string) *MessageBuilder {
	mb.message.Headers = headers
	return mb
}

func (mb *MessageBuilder) Build() *Message {
	return mb.message
}
