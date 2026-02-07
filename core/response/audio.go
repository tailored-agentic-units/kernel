package response

import (
	"encoding/json"
	"fmt"
)

// AudioResponse represents a parsed audio transcription response.
// Fields beyond Text are populated when the response_format is "verbose_json".
type AudioResponse struct {
	Task     string         `json:"task,omitempty"`
	Language string         `json:"language,omitempty"`
	Duration float64        `json:"duration,omitempty"`
	Text     string         `json:"text"`
	Words    []AudioWord    `json:"words,omitempty"`
	Segments []AudioSegment `json:"segments,omitempty"`
}

// AudioWord represents a word-level timestamp from verbose transcription output.
type AudioWord struct {
	Word  string  `json:"word"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

// AudioSegment represents a segment-level timestamp from verbose transcription output.
type AudioSegment struct {
	ID    int     `json:"id"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Text  string  `json:"text"`
}

// Content returns the transcribed text.
func (r *AudioResponse) Content() string {
	return r.Text
}

// ParseAudio parses a JSON audio transcription response body into an AudioResponse.
func ParseAudio(body []byte) (*AudioResponse, error) {
	var response AudioResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse audio response: %w", err)
	}
	return &response, nil
}
