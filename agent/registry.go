package agent

import (
	"fmt"
	"sort"
	"sync"

	"github.com/tailored-agentic-units/kernel/core/config"
	"github.com/tailored-agentic-units/kernel/core/protocol"
)

// AgentInfo describes a registered agent's name and supported protocols.
type AgentInfo struct {
	Name         string
	Capabilities []protocol.Protocol
}

// Registry manages named agent configurations with lazy instantiation.
// Configs are stored at registration time; agents are created on first
// Get call. Thread-safe for concurrent access.
type Registry struct {
	mu      sync.RWMutex
	configs map[string]config.AgentConfig
	agents  map[string]Agent
}

// NewRegistry creates an empty Registry.
func NewRegistry() *Registry {
	return &Registry{
		configs: make(map[string]config.AgentConfig),
		agents:  make(map[string]Agent),
	}
}

// Capabilities returns the protocols supported by a named agent.
// Derived from ModelConfig.Capabilities keys without instantiation.
func (r *Registry) Capabilities(name string) ([]protocol.Protocol, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cfg, exists := r.configs[name]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrAgentNotFound, name)
	}

	return capabilitiesFromConfig(&cfg), nil
}

// Get retrieves a named agent, instantiating it lazily on first access.
func (r *Registry) Get(name string) (Agent, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, registered := r.configs[name]; !registered {
		return nil, fmt.Errorf("%w: %s", ErrAgentNotFound, name)
	}

	if a, exists := r.agents[name]; exists {
		return a, nil
	}

	cfg := r.configs[name]
	a, err := New(&cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent %q: %w", name, err)
	}

	r.agents[name] = a
	return a, nil
}

// List returns information about all registered agents, sorted by name.
func (r *Registry) List() []AgentInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	infos := make([]AgentInfo, 0, len(r.configs))
	for name, cfg := range r.configs {
		infos = append(infos, AgentInfo{
			Name:         name,
			Capabilities: capabilitiesFromConfig(&cfg),
		})
	}

	sort.Slice(infos, func(i, j int) bool {
		return infos[i].Name < infos[j].Name
	})

	return infos
}

// Register adds a named agent configuration to the registry.
// The agent is not instantiated until Get is called.
func (r *Registry) Register(name string, cfg config.AgentConfig) error {
	if name == "" {
		return ErrEmptyAgentName
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.configs[name]; exists {
		return fmt.Errorf("%w: %s", ErrAgentExists, name)
	}

	r.configs[name] = cfg
	return nil
}

// Replace updates the configuration for an existing named agent.
// Any cached agent instance is invalidated; the next Get re-instantiates.
func (r *Registry) Replace(name string, cfg config.AgentConfig) error {
	if name == "" {
		return ErrEmptyAgentName
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.configs[name]; !exists {
		return fmt.Errorf("%w: %s", ErrAgentNotFound, name)
	}

	r.configs[name] = cfg
	delete(r.agents, name)
	return nil
}

// Unregister removes a named agent from the registry.
func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.configs[name]; !exists {
		return fmt.Errorf("%w: %s", ErrAgentNotFound, name)
	}

	delete(r.configs, name)
	delete(r.agents, name)
	return nil
}

func capabilitiesFromConfig(cfg *config.AgentConfig) []protocol.Protocol {
	if cfg.Model == nil || len(cfg.Model.Capabilities) == 0 {
		return nil
	}

	capes := make([]protocol.Protocol, 0, len(cfg.Model.Capabilities))
	for key := range cfg.Model.Capabilities {
		if protocol.IsValid(key) {
			capes = append(capes, protocol.Protocol(key))
		}
	}

	sort.Slice(capes, func(i, j int) bool {
		return string(capes[i]) < string(capes[j])
	})

	return capes
}
