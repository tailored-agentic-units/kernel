package response

// TokenUsage tracks token consumption for a request/response cycle.
// Provides counts for prompt tokens, completion tokens, and total tokens used.
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
