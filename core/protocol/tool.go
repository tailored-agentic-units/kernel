package protocol

// Tool defines a function that can be called by the LLM.
// This is the canonical tool definition type used across the kernel.
// Parameters uses JSON Schema format to describe the function's input.
type Tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}
