package kernel

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/tailored-agentic-units/kernel/core/config"
	"github.com/tailored-agentic-units/kernel/memory"
	"github.com/tailored-agentic-units/kernel/session"
)

const defaultMaxIterations = 10

// Config holds initialization parameters for all kernel subsystems.
// Each subsystem section delegates to that subsystem's config-driven constructor.
type Config struct {
	Agent         config.AgentConfig            `json:"agent"`
	Agents        map[string]config.AgentConfig `json:"agents,omitempty"`
	Session       session.Config                `json:"session"`
	Memory        memory.Config                 `json:"memory"`
	MaxIterations int                           `json:"max_iterations,omitempty"`
	SystemPrompt  string                        `json:"system_prompt,omitempty"`
}

// DefaultConfig returns a Config with sensible defaults for all subsystems.
func DefaultConfig() Config {
	return Config{
		Agent:         config.DefaultAgentConfig(),
		Session:       session.DefaultConfig(),
		Memory:        memory.DefaultConfig(),
		MaxIterations: defaultMaxIterations,
	}
}

// Merge applies non-zero values from source into c, delegating to each
// subsystem's Merge method.
func (c *Config) Merge(source *Config) {
	c.Agent.Merge(&source.Agent)
	c.Session.Merge(&source.Session)
	c.Memory.Merge(&source.Memory)

	if source.MaxIterations > 0 {
		c.MaxIterations = source.MaxIterations
	}
	if source.SystemPrompt != "" {
		c.SystemPrompt = source.SystemPrompt
	}

	if len(source.Agents) > 0 {
		c.Agents = source.Agents
	}
}

// LoadConfig reads a JSON config file, merges it with defaults, and returns
// the resulting Config.
func LoadConfig(filename string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var loaded Config
	if err := json.Unmarshal(data, &loaded); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	cfg.Merge(&loaded)
	return &cfg, nil
}
