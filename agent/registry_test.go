package agent_test

import (
	"errors"
	"sync"
	"testing"

	"github.com/tailored-agentic-units/kernel/agent"
	"github.com/tailored-agentic-units/kernel/core/config"
	"github.com/tailored-agentic-units/kernel/core/protocol"
)

func ollamaConfig(modelName string, caps ...string) config.AgentConfig {
	capabilities := make(map[string]map[string]any, len(caps))
	for _, c := range caps {
		capabilities[c] = map[string]any{}
	}

	return config.AgentConfig{
		Provider: &config.ProviderConfig{
			Name:    "ollama",
			BaseURL: "http://localhost:11434",
		},
		Model: &config.ModelConfig{
			Name:         modelName,
			Capabilities: capabilities,
		},
	}
}

func TestRegistry_RegisterAndGet(t *testing.T) {
	r := agent.NewRegistry()

	cfg := ollamaConfig("qwen3:8b", "chat", "tools")
	if err := r.Register("qwen3-8b", cfg); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	a, err := r.Get("qwen3-8b")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if a == nil {
		t.Fatal("Get returned nil agent")
	}
	if a.ID() == "" {
		t.Error("agent has empty ID")
	}

	// Second Get returns same cached instance
	a2, err := r.Get("qwen3-8b")
	if err != nil {
		t.Fatalf("second Get failed: %v", err)
	}
	if a.ID() != a2.ID() {
		t.Errorf("cached agent ID mismatch: got %q and %q", a.ID(), a2.ID())
	}
}

func TestRegistry_RegisterEmptyName(t *testing.T) {
	r := agent.NewRegistry()

	err := r.Register("", config.AgentConfig{})
	if !errors.Is(err, agent.ErrEmptyAgentName) {
		t.Errorf("got %v, want ErrEmptyAgentName", err)
	}
}

func TestRegistry_RegisterDuplicate(t *testing.T) {
	r := agent.NewRegistry()

	cfg := ollamaConfig("qwen3:8b", "chat")
	if err := r.Register("qwen3-8b", cfg); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	err := r.Register("qwen3-8b", cfg)
	if !errors.Is(err, agent.ErrAgentExists) {
		t.Errorf("got %v, want ErrAgentExists", err)
	}
}

func TestRegistry_GetNotFound(t *testing.T) {
	r := agent.NewRegistry()

	_, err := r.Get("nonexistent")
	if !errors.Is(err, agent.ErrAgentNotFound) {
		t.Errorf("got %v, want ErrAgentNotFound", err)
	}
}

func TestRegistry_Replace(t *testing.T) {
	r := agent.NewRegistry()

	cfg := ollamaConfig("qwen3:8b", "chat")
	if err := r.Register("qwen3-8b", cfg); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Get to populate cache
	a1, err := r.Get("qwen3-8b")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	// Replace with new config
	newCfg := ollamaConfig("qwen3:8b", "chat", "tools")
	if err := r.Replace("qwen3-8b", newCfg); err != nil {
		t.Fatalf("Replace failed: %v", err)
	}

	// Get should re-instantiate (different agent ID)
	a2, err := r.Get("qwen3-8b")
	if err != nil {
		t.Fatalf("Get after Replace failed: %v", err)
	}
	if a1.ID() == a2.ID() {
		t.Error("expected new agent instance after Replace, got same ID")
	}

	// Capabilities should reflect new config
	caps, err := r.Capabilities("qwen3-8b")
	if err != nil {
		t.Fatalf("Capabilities failed: %v", err)
	}
	if len(caps) != 2 {
		t.Errorf("got %d capabilities, want 2", len(caps))
	}
}

func TestRegistry_ReplaceEmptyName(t *testing.T) {
	r := agent.NewRegistry()

	err := r.Replace("", config.AgentConfig{})
	if !errors.Is(err, agent.ErrEmptyAgentName) {
		t.Errorf("got %v, want ErrEmptyAgentName", err)
	}
}

func TestRegistry_ReplaceNotFound(t *testing.T) {
	r := agent.NewRegistry()

	err := r.Replace("nonexistent", config.AgentConfig{})
	if !errors.Is(err, agent.ErrAgentNotFound) {
		t.Errorf("got %v, want ErrAgentNotFound", err)
	}
}

func TestRegistry_List(t *testing.T) {
	r := agent.NewRegistry()

	r.Register("llava-13b", ollamaConfig("llava:13b", "chat", "vision"))
	r.Register("qwen3-8b", ollamaConfig("qwen3:8b", "chat", "tools"))

	infos := r.List()
	if len(infos) != 2 {
		t.Fatalf("got %d entries, want 2", len(infos))
	}

	// Sorted by name
	if infos[0].Name != "llava-13b" {
		t.Errorf("got first name %q, want %q", infos[0].Name, "llava-13b")
	}
	if infos[1].Name != "qwen3-8b" {
		t.Errorf("got second name %q, want %q", infos[1].Name, "qwen3-8b")
	}

	// Capabilities sorted
	if len(infos[0].Capabilities) != 2 {
		t.Fatalf("got %d capabilities for llava-13b, want 2", len(infos[0].Capabilities))
	}
	if infos[0].Capabilities[0] != protocol.Chat {
		t.Errorf("got first capability %q, want %q", infos[0].Capabilities[0], protocol.Chat)
	}
	if infos[0].Capabilities[1] != protocol.Vision {
		t.Errorf("got second capability %q, want %q", infos[0].Capabilities[1], protocol.Vision)
	}
}

func TestRegistry_ListEmpty(t *testing.T) {
	r := agent.NewRegistry()

	infos := r.List()
	if len(infos) != 0 {
		t.Errorf("got %d entries, want 0", len(infos))
	}
}

func TestRegistry_Unregister(t *testing.T) {
	r := agent.NewRegistry()

	cfg := ollamaConfig("qwen3:8b", "chat")
	r.Register("qwen3-8b", cfg)

	// Populate cache
	r.Get("qwen3-8b")

	if err := r.Unregister("qwen3-8b"); err != nil {
		t.Fatalf("Unregister failed: %v", err)
	}

	// Get should fail
	_, err := r.Get("qwen3-8b")
	if !errors.Is(err, agent.ErrAgentNotFound) {
		t.Errorf("got %v, want ErrAgentNotFound after Unregister", err)
	}

	// List should be empty
	if infos := r.List(); len(infos) != 0 {
		t.Errorf("got %d entries after Unregister, want 0", len(infos))
	}
}

func TestRegistry_UnregisterNotFound(t *testing.T) {
	r := agent.NewRegistry()

	err := r.Unregister("nonexistent")
	if !errors.Is(err, agent.ErrAgentNotFound) {
		t.Errorf("got %v, want ErrAgentNotFound", err)
	}
}

func TestRegistry_Capabilities(t *testing.T) {
	r := agent.NewRegistry()

	cfg := ollamaConfig("qwen3:8b", "chat", "tools", "embeddings")
	r.Register("qwen3-8b", cfg)

	caps, err := r.Capabilities("qwen3-8b")
	if err != nil {
		t.Fatalf("Capabilities failed: %v", err)
	}

	expected := []protocol.Protocol{protocol.Chat, protocol.Embeddings, protocol.Tools}
	if len(caps) != len(expected) {
		t.Fatalf("got %d capabilities, want %d", len(caps), len(expected))
	}
	for i, want := range expected {
		if caps[i] != want {
			t.Errorf("capability[%d] = %q, want %q", i, caps[i], want)
		}
	}
}

func TestRegistry_CapabilitiesNotFound(t *testing.T) {
	r := agent.NewRegistry()

	_, err := r.Capabilities("nonexistent")
	if !errors.Is(err, agent.ErrAgentNotFound) {
		t.Errorf("got %v, want ErrAgentNotFound", err)
	}
}

func TestRegistry_CapabilitiesNilModel(t *testing.T) {
	r := agent.NewRegistry()

	cfg := config.AgentConfig{
		Provider: &config.ProviderConfig{
			Name:    "ollama",
			BaseURL: "http://localhost:11434",
		},
	}
	r.Register("no-model", cfg)

	caps, err := r.Capabilities("no-model")
	if err != nil {
		t.Fatalf("Capabilities failed: %v", err)
	}
	if caps != nil {
		t.Errorf("got %v, want nil for nil model", caps)
	}
}

func TestRegistry_CapabilitiesInvalidKeysFiltered(t *testing.T) {
	r := agent.NewRegistry()

	cfg := config.AgentConfig{
		Provider: &config.ProviderConfig{
			Name:    "ollama",
			BaseURL: "http://localhost:11434",
		},
		Model: &config.ModelConfig{
			Name: "test",
			Capabilities: map[string]map[string]any{
				"chat":    {},
				"invalid": {},
				"tools":   {},
			},
		},
	}
	r.Register("mixed", cfg)

	caps, err := r.Capabilities("mixed")
	if err != nil {
		t.Fatalf("Capabilities failed: %v", err)
	}
	if len(caps) != 2 {
		t.Fatalf("got %d capabilities, want 2 (invalid filtered)", len(caps))
	}
	if caps[0] != protocol.Chat || caps[1] != protocol.Tools {
		t.Errorf("got %v, want [chat tools]", caps)
	}
}

func TestRegistry_ConcurrentAccess(t *testing.T) {
	r := agent.NewRegistry()

	for i := range 10 {
		name := string(rune('a' + i))
		r.Register(name, ollamaConfig("model-"+name, "chat"))
	}

	var wg sync.WaitGroup
	for range 50 {
		wg.Go(func() {
			r.List()
		})
		wg.Go(func() {
			r.Capabilities("a")
		})
		wg.Go(func() {
			r.Get("b")
		})
	}
	wg.Wait()
}
