package request_test

import (
	"encoding/json"
	"testing"

	"github.com/tailored-agentic-units/kernel/core/config"
	"github.com/tailored-agentic-units/kernel/core/model"
	"github.com/tailored-agentic-units/kernel/core/protocol"
	"github.com/tailored-agentic-units/kernel/agent/providers"
	"github.com/tailored-agentic-units/kernel/agent/request"
)

func newTestProvider(t *testing.T) providers.Provider {
	t.Helper()
	cfg := &config.ProviderConfig{
		Name:    "ollama",
		BaseURL: "http://localhost:11434",
	}
	p, err := providers.NewOllama(cfg)
	if err != nil {
		t.Fatalf("NewOllama failed: %v", err)
	}
	return p
}

func TestNewAudio(t *testing.T) {
	p := newTestProvider(t)
	m := &model.Model{
		Name:    "whisper-1",
		Options: make(map[protocol.Protocol]map[string]any),
	}

	audioOpts := map[string]any{"language": "en"}
	opts := map[string]any{"temperature": 0.2}

	req := request.NewAudio(p, m, "audio-input", audioOpts, opts)

	if req == nil {
		t.Fatal("NewAudio returned nil")
	}
}

func TestAudioRequest_Protocol(t *testing.T) {
	p := newTestProvider(t)
	m := &model.Model{
		Name:    "whisper-1",
		Options: make(map[protocol.Protocol]map[string]any),
	}

	req := request.NewAudio(p, m, "audio-input", nil, nil)

	if req.Protocol() != protocol.Audio {
		t.Errorf("got protocol %q, want %q", req.Protocol(), protocol.Audio)
	}
}

func TestAudioRequest_Headers(t *testing.T) {
	p := newTestProvider(t)
	m := &model.Model{
		Name:    "whisper-1",
		Options: make(map[protocol.Protocol]map[string]any),
	}

	req := request.NewAudio(p, m, "audio-input", nil, nil)

	headers := req.Headers()
	if headers["Content-Type"] != "application/json" {
		t.Errorf("got Content-Type %q, want %q", headers["Content-Type"], "application/json")
	}
}

func TestAudioRequest_Marshal(t *testing.T) {
	p := newTestProvider(t)
	m := &model.Model{
		Name:    "whisper-1",
		Options: make(map[protocol.Protocol]map[string]any),
	}

	audioOpts := map[string]any{"language": "en", "response_format": "verbose_json"}
	opts := map[string]any{"temperature": 0.2}

	req := request.NewAudio(p, m, "audio-input", audioOpts, opts)

	body, err := req.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	if result["model"] != "whisper-1" {
		t.Errorf("got model %v, want whisper-1", result["model"])
	}

	if result["input"] != "audio-input" {
		t.Errorf("got input %v, want audio-input", result["input"])
	}

	if result["language"] != "en" {
		t.Errorf("got language %v, want en", result["language"])
	}

	if result["response_format"] != "verbose_json" {
		t.Errorf("got response_format %v, want verbose_json", result["response_format"])
	}

	if result["temperature"] != 0.2 {
		t.Errorf("got temperature %v, want 0.2", result["temperature"])
	}
}

func TestAudioRequest_Marshal_NilOptions(t *testing.T) {
	p := newTestProvider(t)
	m := &model.Model{
		Name:    "whisper-1",
		Options: make(map[protocol.Protocol]map[string]any),
	}

	req := request.NewAudio(p, m, "audio-input", nil, nil)

	body, err := req.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	if result["model"] != "whisper-1" {
		t.Errorf("got model %v, want whisper-1", result["model"])
	}

	if result["input"] != "audio-input" {
		t.Errorf("got input %v, want audio-input", result["input"])
	}
}

func TestAudioRequest_Provider(t *testing.T) {
	p := newTestProvider(t)
	m := &model.Model{
		Name:    "whisper-1",
		Options: make(map[protocol.Protocol]map[string]any),
	}

	req := request.NewAudio(p, m, "audio-input", nil, nil)

	if req.Provider().Name() != "ollama" {
		t.Errorf("got provider name %q, want %q", req.Provider().Name(), "ollama")
	}
}

func TestAudioRequest_Model(t *testing.T) {
	p := newTestProvider(t)
	m := &model.Model{
		Name:    "whisper-1",
		Options: make(map[protocol.Protocol]map[string]any),
	}

	req := request.NewAudio(p, m, "audio-input", nil, nil)

	if req.Model() != m {
		t.Error("Model() returned different model than configured")
	}
}
