package llm

// LLM represents a generic interface for interacting with LLMs
type LLM interface {
	// QueryStream processes the LLM response as a stream and sends segments to TTS as they arrive
	QueryStream(model, text string, ttsCallback func(segment string, playID string, autoHangup bool) error) (string, error)

	// Query queries the LLM with text and gets a response
	Query(model, text string) (string, *HangupTool, error)

	// Reset clears the conversation history
	Reset()
}
