package kernel_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tailored-agentic-units/kernel/core/config"
	"github.com/tailored-agentic-units/kernel/kernel"
)

func TestDefaultConfig(t *testing.T) {
	cfg := kernel.DefaultConfig()

	if cfg.MaxIterations != 10 {
		t.Errorf("got MaxIterations %d, want 10", cfg.MaxIterations)
	}
}

func TestConfig_Merge(t *testing.T) {
	cfg := kernel.DefaultConfig()

	source := &kernel.Config{
		MaxIterations: 20,
		SystemPrompt:  "merged prompt",
	}

	cfg.Merge(source)

	if cfg.MaxIterations != 20 {
		t.Errorf("got MaxIterations %d, want 20", cfg.MaxIterations)
	}

	if cfg.SystemPrompt != "merged prompt" {
		t.Errorf("got SystemPrompt %q, want %q", cfg.SystemPrompt, "merged prompt")
	}
}

func TestConfig_Merge_ZeroValuesPreserveDefaults(t *testing.T) {
	cfg := kernel.DefaultConfig()
	original := cfg.MaxIterations

	source := &kernel.Config{} // All zero values

	cfg.Merge(source)

	if cfg.MaxIterations != original {
		t.Errorf("got MaxIterations %d, want %d (preserved default)", cfg.MaxIterations, original)
	}
}

func TestLoadConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")

	content := `{
		"max_iterations": 25,
		"system_prompt": "loaded prompt",
		"memory": {
			"path": "/tmp/mem"
		}
	}`

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	cfg, err := kernel.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.MaxIterations != 25 {
		t.Errorf("got MaxIterations %d, want 25", cfg.MaxIterations)
	}

	if cfg.SystemPrompt != "loaded prompt" {
		t.Errorf("got SystemPrompt %q, want %q", cfg.SystemPrompt, "loaded prompt")
	}

	if cfg.Memory.Path != "/tmp/mem" {
		t.Errorf("got Memory.Path %q, want %q", cfg.Memory.Path, "/tmp/mem")
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	_, err := kernel.LoadConfig("/nonexistent/path/config.json")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "bad.json")

	if err := os.WriteFile(configPath, []byte("{invalid}"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	_, err := kernel.LoadConfig(configPath)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestConfig_MergeAgents(t *testing.T) {
	cfg := kernel.DefaultConfig()

	source := &kernel.Config{
		Agents: map[string]config.AgentConfig{
			"qwen3-8b": {
				Provider: &config.ProviderConfig{Name: "ollama", BaseURL: "http://localhost:11434"},
				Model:    &config.ModelConfig{Name: "qwen3:8b"},
			},
		},
	}

	cfg.Merge(source)

	if len(cfg.Agents) != 1 {
		t.Fatalf("got %d agents, want 1", len(cfg.Agents))
	}
	if _, ok := cfg.Agents["qwen3-8b"]; !ok {
		t.Error("expected qwen3-8b in agents map")
	}
}

func TestConfig_MergeAgents_EmptySourcePreservesTarget(t *testing.T) {
	cfg := kernel.DefaultConfig()
	cfg.Agents = map[string]config.AgentConfig{
		"existing": {
			Provider: &config.ProviderConfig{Name: "ollama"},
		},
	}

	source := &kernel.Config{} // No agents

	cfg.Merge(source)

	if len(cfg.Agents) != 1 {
		t.Errorf("got %d agents, want 1 (preserved)", len(cfg.Agents))
	}
}

func TestConfig_MergeAgents_SourceReplacesTarget(t *testing.T) {
	cfg := kernel.DefaultConfig()
	cfg.Agents = map[string]config.AgentConfig{
		"old": {
			Provider: &config.ProviderConfig{Name: "ollama"},
		},
	}

	source := &kernel.Config{
		Agents: map[string]config.AgentConfig{
			"new": {
				Provider: &config.ProviderConfig{Name: "azure"},
			},
		},
	}

	cfg.Merge(source)

	if len(cfg.Agents) != 1 {
		t.Fatalf("got %d agents, want 1", len(cfg.Agents))
	}
	if _, ok := cfg.Agents["new"]; !ok {
		t.Error("expected new agent, got old")
	}
}

func TestLoadConfig_WithAgents(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")

	content := `{
		"max_iterations": 5,
		"agents": {
			"qwen3-8b": {
				"provider": {"name": "ollama", "base_url": "http://localhost:11434"},
				"model": {"name": "qwen3:8b", "capabilities": {"chat": {}, "tools": {}}}
			}
		}
	}`

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	cfg, err := kernel.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if len(cfg.Agents) != 1 {
		t.Fatalf("got %d agents, want 1", len(cfg.Agents))
	}

	agentCfg, ok := cfg.Agents["qwen3-8b"]
	if !ok {
		t.Fatal("expected qwen3-8b in agents")
	}
	if agentCfg.Model.Name != "qwen3:8b" {
		t.Errorf("got model name %q, want %q", agentCfg.Model.Name, "qwen3:8b")
	}
	if len(agentCfg.Model.Capabilities) != 2 {
		t.Errorf("got %d capabilities, want 2", len(agentCfg.Model.Capabilities))
	}
}
