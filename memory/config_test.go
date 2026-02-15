package memory_test

import (
	"testing"

	"github.com/tailored-agentic-units/kernel/memory"
)

func TestDefaultConfig(t *testing.T) {
	cfg := memory.DefaultConfig()

	if cfg.Path != "" {
		t.Errorf("got Path %q, want empty string", cfg.Path)
	}
}

func TestConfig_Merge(t *testing.T) {
	cfg := memory.DefaultConfig()

	source := &memory.Config{Path: "/data/memory"}
	cfg.Merge(source)

	if cfg.Path != "/data/memory" {
		t.Errorf("got Path %q, want %q", cfg.Path, "/data/memory")
	}
}

func TestConfig_Merge_EmptyPreservesDefault(t *testing.T) {
	cfg := memory.Config{Path: "/original"}

	source := &memory.Config{} // Empty path
	cfg.Merge(source)

	if cfg.Path != "/original" {
		t.Errorf("got Path %q, want %q (preserved)", cfg.Path, "/original")
	}
}

func TestNewStore_EmptyPath(t *testing.T) {
	cfg := &memory.Config{}

	store, err := memory.NewStore(cfg)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	if store != nil {
		t.Error("expected nil store for empty path")
	}
}

func TestNewStore_WithPath(t *testing.T) {
	dir := t.TempDir()

	cfg := &memory.Config{Path: dir}

	store, err := memory.NewStore(cfg)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	if store == nil {
		t.Fatal("expected non-nil store for valid path")
	}
}
