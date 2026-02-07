package hub

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/tailored-agentic-units/kernel/agent"
	"github.com/tailored-agentic-units/kernel/orchestrate/config"
	"github.com/tailored-agentic-units/kernel/orchestrate/messaging"
)

type registration struct {
	Agent    agent.Agent
	Handler  MessageHandler
	Channel  *MessageChannel[*messaging.Message]
	LastSeen time.Time
}

type Hub interface {
	RegisterAgent(ag agent.Agent, handler MessageHandler) error
	UnregisterAgent(agentID string) error

	Send(ctx context.Context, from, to string, data any) error
	Request(ctx context.Context, from, to string, data any) (*messaging.Message, error)
	Broadcast(ctx context.Context, from string, data any) error

	Subscribe(agentID, topic string) error
	Publish(ctx context.Context, from, topic string, data any) error

	Metrics() MetricsSnapshot
	Shutdown(timeout time.Duration) error
}

type hub struct {
	name string

	agents      map[string]*registration
	agentsMutex sync.RWMutex

	responseChannels map[string]chan *messaging.Message
	responsesMutex   sync.RWMutex

	subscriptions map[string]map[string]*registration
	subsMutex     sync.RWMutex

	channelBufferSize int
	defaultTimeout    time.Duration

	logger  *slog.Logger
	metrics *Metrics

	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}
}

func New(ctx context.Context, hubConfig config.HubConfig) Hub {
	hubCtx, cancel := context.WithCancel(ctx)

	h := &hub{
		name:              hubConfig.Name,
		agents:            make(map[string]*registration),
		responseChannels:  make(map[string]chan *messaging.Message),
		subscriptions:     make(map[string]map[string]*registration),
		channelBufferSize: hubConfig.ChannelBufferSize,
		defaultTimeout:    hubConfig.DefaultTimeout,
		logger:            hubConfig.Logger,
		metrics:           NewMetrics(),
		ctx:               hubCtx,
		cancel:            cancel,
		done:              make(chan struct{}),
	}

	go h.messageLoop()

	return h
}

func (h *hub) RegisterAgent(ag agent.Agent, handler MessageHandler) error {
	agentID := ag.ID()
	h.agentsMutex.Lock()
	defer h.agentsMutex.Unlock()

	if _, exists := h.agents[agentID]; exists {
		return fmt.Errorf("agent already registered: %s", agentID)
	}

	channel := NewMessageChannel[*messaging.Message](h.ctx, h.channelBufferSize)

	reg := &registration{
		Agent:    ag,
		Handler:  handler,
		Channel:  channel,
		LastSeen: time.Now(),
	}

	h.agents[agentID] = reg
	h.metrics.RecordLocalAgent(1)

	h.logger.DebugContext(
		h.ctx,
		"agent registered",
		slog.String("hub_name", h.name),
		slog.String("agent_id", agentID),
	)

	return nil
}

func (h *hub) UnregisterAgent(agentID string) error {
	h.agentsMutex.Lock()
	reg, exists := h.agents[agentID]
	if exists {
		delete(h.agents, agentID)
		reg.Channel.Close()
	}
	h.agentsMutex.Unlock()

	if !exists {
		return fmt.Errorf("agent not found: %s", agentID)
	}

	h.subsMutex.Lock()
	for topic, subs := range h.subscriptions {
		if _, exists := subs[agentID]; exists {
			delete(subs, agentID)
			if len(subs) == 0 {
				delete(h.subscriptions, topic)
			}
		}
	}
	h.subsMutex.Unlock()

	h.metrics.RecordLocalAgent(-1)
	h.logger.DebugContext(
		h.ctx,
		"agent unregistered",
		slog.String("hub_name", h.name),
		slog.String("agent_id", agentID),
	)

	return nil
}

func (h *hub) Send(ctx context.Context, from, to string, data any) error {
	h.agentsMutex.RLock()
	reg, exists := h.agents[to]
	h.agentsMutex.RUnlock()

	if !exists {
		return fmt.Errorf("destination agent not found: %s", to)
	}

	message := messaging.NewNotification(from, to, data).Build()
	err := reg.Channel.Send(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to deliver message: %w", err)
	}

	h.updateLastSeen(from)
	h.metrics.RecordMessageSent(1)

	return nil
}

func (h *hub) Request(ctx context.Context, from, to string, data any) (*messaging.Message, error) {
	h.agentsMutex.RLock()
	reg, exists := h.agents[to]
	h.agentsMutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("destination agent not found: %s", to)
	}

	message := messaging.NewRequest(from, to, data).Build()
	responseChannel := make(chan *messaging.Message, 1)

	h.responsesMutex.Lock()
	h.responseChannels[message.ID] = responseChannel
	h.responsesMutex.Unlock()

	defer func() {
		h.responsesMutex.Lock()
		delete(h.responseChannels, message.ID)
		h.responsesMutex.Unlock()
		close(responseChannel)
	}()

	err := reg.Channel.Send(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	h.updateLastSeen(from)

	timeout := h.defaultTimeout
	if deadline, ok := ctx.Deadline(); ok {
		timeout = time.Until(deadline)
	}

	select {
	case response := <-responseChannel:
		return response, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("request cancelled: %w", ctx.Err())
	case <-time.After(timeout):
		return nil, fmt.Errorf("request timed out after %v", timeout)
	}
}

func (h *hub) Broadcast(ctx context.Context, from string, data any) error {
	h.agentsMutex.RLock()
	registrations := make([]*registration, 0, len(h.agents))
	for agentID, reg := range h.agents {
		if agentID != from {
			registrations = append(registrations, reg)
		}
	}
	h.agentsMutex.RUnlock()

	delivered := 0
	for _, reg := range registrations {
		message := messaging.NewMessage(
			from,
			reg.Agent.ID(),
			messaging.MessageTypeBroadcast,
			data,
		).Build()

		if err := reg.Channel.Send(ctx, message); err != nil {
			h.logger.WarnContext(
				ctx,
				"failed to deliver broadcast",
				slog.String("hub_name", h.name),
				slog.String("from", from),
				slog.String("to", reg.Agent.ID()),
				slog.String("error", err.Error()),
			)
		} else {
			delivered++
		}
	}

	h.updateLastSeen(from)
	h.logger.DebugContext(
		ctx,
		"broadcast sent",
		slog.String("hub_name", h.name),
		slog.String("from", from),
		slog.Int("recipients", len(registrations)),
		slog.Int("delivered", delivered),
	)

	return nil
}

func (h *hub) Subscribe(agentID, topic string) error {
	h.agentsMutex.RLock()
	reg, exists := h.agents[agentID]
	h.agentsMutex.RUnlock()

	if !exists {
		return fmt.Errorf("agent not found: %s", agentID)
	}

	h.subsMutex.Lock()
	if h.subscriptions[topic] == nil {
		h.subscriptions[topic] = make(map[string]*registration)
	}
	h.subscriptions[topic][agentID] = reg
	h.subsMutex.Unlock()

	h.logger.DebugContext(
		h.ctx,
		"agent subscribed to topic",
		slog.String("hub_name", h.name),
		slog.String("agent_id", agentID),
		slog.String("topic", topic),
	)

	return nil
}

func (h *hub) Publish(ctx context.Context, from, topic string, data any) error {
	h.subsMutex.RLock()
	subscribers, exists := h.subscriptions[topic]
	if !exists {
		h.subsMutex.RUnlock()
		h.logger.DebugContext(
			ctx,
			"no subscribers for topic",
			slog.String("hub_name", h.name),
			slog.String("topic", topic),
		)
		return nil
	}

	subscriberList := make([]*registration, 0, len(subscribers))
	for _, reg := range subscribers {
		subscriberList = append(subscriberList, reg)
	}
	h.subsMutex.RUnlock()

	delivered := 0
	for _, reg := range subscriberList {
		if reg.Agent.ID() == from {
			continue
		}

		message := messaging.NewNotification(from, reg.Agent.ID(), data).Topic(topic).Build()
		if err := reg.Channel.Send(ctx, message); err != nil {
			h.logger.WarnContext(
				ctx,
				"failed to deliver published message",
				slog.String("hub_name", h.name),
				slog.String("topic", topic),
				slog.String("subscriber", reg.Agent.ID()),
				slog.String("error", err.Error()),
			)
		} else {
			delivered++
		}
	}

	h.updateLastSeen(from)
	h.logger.DebugContext(
		ctx,
		"message published",
		slog.String("hub_name", h.name),
		slog.String("topic", topic),
		slog.Int("subscribers", len(subscriberList)),
		slog.Int("delivered", delivered),
	)

	return nil
}

func (h *hub) Metrics() MetricsSnapshot {
	return h.metrics.Snapshot()
}

func (h *hub) Shutdown(timeout time.Duration) error {
	h.logger.DebugContext(
		h.ctx,
		"shutting down hub",
		slog.String("hub_name", h.name),
	)
	h.cancel()

	select {
	case <-h.done:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("hub shutdown timeout after %v", timeout)
	}
}

func (h *hub) messageLoop() {
	defer close(h.done)

	for {
		select {
		case <-h.ctx.Done():
			return
		default:
			h.processAgentMessages()
		}
	}
}

func (h *hub) processAgentMessages() {
	h.agentsMutex.RLock()
	if len(h.agents) == 0 {
		h.agentsMutex.RUnlock()
		return
	}

	registrations := make([]*registration, 0, len(h.agents))
	for _, reg := range h.agents {
		registrations = append(registrations, reg)
	}
	h.agentsMutex.RUnlock()

	for _, reg := range registrations {
		select {
		case <-h.ctx.Done():
			return
		default:
			if message, ok := reg.Channel.TryReceive(); ok && message != nil {
				go h.handleMessage(reg, message)
			}
		}
	}
}

func (h *hub) handleMessage(reg *registration, message *messaging.Message) {
	if reg.Handler == nil {
		return
	}

	h.metrics.RecordMessageRecv(1)

	context := &MessageContext{
		HubName: h.name,
		Agent:   reg.Agent,
	}

	response, err := reg.Handler(h.ctx, message, context)
	if err != nil {
		h.logger.ErrorContext(
			h.ctx,
			"message handler failed",
			slog.String("hub_name", h.name),
			slog.String("agent_id", reg.Agent.ID()),
			slog.String("from", message.From),
			slog.String("error", err.Error()),
		)
		return
	}

	if response != nil {
		if response.Type == messaging.MessageTypeResponse && response.ReplyTo != "" {
			h.responsesMutex.RLock()
			respChan, exists := h.responseChannels[response.ReplyTo]
			h.responsesMutex.RUnlock()

			if exists {
				select {
				case respChan <- response:
				default:
				}
				return
			}
		}

		h.agentsMutex.RLock()
		targetReg, exists := h.agents[response.To]
		h.agentsMutex.RUnlock()

		if exists {
			if err := targetReg.Channel.Send(h.ctx, response); err != nil {
				h.logger.ErrorContext(
					h.ctx,
					"failed to send response",
					slog.String("hub_name", h.name),
					slog.String("from", response.From),
					slog.String("to", response.To),
					slog.String("error", err.Error()),
				)
			}
		}
	}
}

func (h *hub) updateLastSeen(agentID string) {
	h.agentsMutex.Lock()
	if reg, exists := h.agents[agentID]; exists {
		reg.LastSeen = time.Now()
	}
	h.agentsMutex.Unlock()
}
