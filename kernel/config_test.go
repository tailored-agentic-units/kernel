package kernel_test

import (
	"os"
	"path/filepath"
	"testing"

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
