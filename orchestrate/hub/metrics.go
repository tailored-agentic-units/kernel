package hub

import "sync/atomic"

type MetricsSnapshot struct {
	LocalAgents  int64
	MessagesSent int64
	MessagesRecv int64
}

type Metrics struct {
	localAgents  atomic.Int64
	messagesSent atomic.Int64
	messagesRecv atomic.Int64
}

func NewMetrics() *Metrics {
	return &Metrics{}
}

func (m *Metrics) RecordLocalAgent(delta int) {
	m.localAgents.Add(int64(delta))
}

func (m *Metrics) RecordMessageSent(delta int) {
	m.messagesSent.Add(int64(delta))
}

func (m *Metrics) RecordMessageRecv(delta int) {
	m.messagesRecv.Add(int64(delta))
}

func (m *Metrics) Snapshot() MetricsSnapshot {
	return MetricsSnapshot{
		LocalAgents:  m.localAgents.Load(),
		MessagesSent: m.messagesSent.Load(),
		MessagesRecv: m.messagesRecv.Load(),
	}
}
