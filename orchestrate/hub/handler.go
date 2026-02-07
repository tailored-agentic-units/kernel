package hub

import (
	"context"

	"github.com/tailored-agentic-units/kernel/agent"
	"github.com/tailored-agentic-units/kernel/orchestrate/messaging"
)

type MessageContext struct {
	HubName string
	Agent   agent.Agent
}

type MessageHandler func(
	ctx context.Context,
	message *messaging.Message,
	context *MessageContext,
) (*messaging.Message, error)
