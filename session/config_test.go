package session_test

import (
	"testing"

	"github.com/tailored-agentic-units/kernel/session"
)

func TestDefaultConfig(t *testing.T) {
	cfg := session.DefaultConfig()

	// Currently an empty struct; verify it doesn't panic.
	_ = cfg
}

func TestConfig_Merge(t *testing.T) {
	cfg := session.DefaultConfig()
	source := session.DefaultConfig()

	// Merge should not panic on empty configs.
	cfg.Merge(&source)
}

func TestNew_FromConfig(t *testing.T) {
	cfg := session.DefaultConfig()

	s, err := session.New(&cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	if s == nil {
		t.Fatal("New returned nil session")
	}

	if s.ID() == "" {
		t.Error("session ID is empty")
	}
}
