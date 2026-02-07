package request

import (
	"github.com/tailored-agentic-units/kernel/core/model"
	"github.com/tailored-agentic-units/kernel/core/protocol"
	"github.com/tailored-agentic-units/kernel/agent/providers"
)

// Request defines the interface for protocol requests.
// All request types implement this interface to provide consistent
// access to request components needed for execution.
type Request interface {
	// Protocol returns the protocol identifier for this request.
	Protocol() protocol.Protocol

	// Headers returns the HTTP headers for this request.
	Headers() map[string]string

	// Marshal converts the request to JSON bytes.
	Marshal() ([]byte, error)

	// Provider returns the provider for this request.
	Provider() providers.Provider

	// Model returns the model for this request.
	Model() *model.Model
}
