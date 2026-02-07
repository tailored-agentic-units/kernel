package request

import (
	"github.com/tailored-agentic-units/kernel/core/model"
	"github.com/tailored-agentic-units/kernel/core/protocol"
	"github.com/tailored-agentic-units/kernel/agent/providers"
)

// AudioRequest represents an audio transcription protocol request.
// Separates audio input (protocol data) from audio-specific options
// and model configuration options.
type AudioRequest struct {
	input        string
	audioOptions map[string]any
	options      map[string]any
	provider     providers.Provider
	model        *model.Model
}

// NewAudio creates a new AudioRequest with the given components.
// Input is the audio source (file path, URL, or base64-encoded data).
// AudioOpts specify transcription options (language, response_format, etc.).
// Opts specify model configuration (temperature, etc.).
func NewAudio(p providers.Provider, m *model.Model, input string, audioOpts, opts map[string]any) *AudioRequest {
	return &AudioRequest{
		input:        input,
		audioOptions: audioOpts,
		options:      opts,
		provider:     p,
		model:        m,
	}
}

func (r *AudioRequest) Protocol() protocol.Protocol {
	return protocol.Audio
}

func (r *AudioRequest) Headers() map[string]string {
	return map[string]string{
		"Content-Type": "application/json",
	}
}

func (r *AudioRequest) Marshal() ([]byte, error) {
	return r.provider.Marshal(protocol.Audio, &providers.AudioData{
		Model:        r.model.Name,
		Input:        r.input,
		AudioOptions: r.audioOptions,
		Options:      r.options,
	})
}

func (r *AudioRequest) Provider() providers.Provider {
	return r.provider
}

func (r *AudioRequest) Model() *model.Model {
	return r.model
}
