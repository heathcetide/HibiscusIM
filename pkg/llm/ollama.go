package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

// OllamaHandler implements the LLM interface for Ollama
type OllamaHandler struct {
	systemMsg string
	messages  []string
	logger    *logrus.Logger
	ctx       context.Context
	apiKey    string
	ollamaURL string
}

// NewOllamaHandler creates a new Ollama handler
func NewOllamaHandler(ctx context.Context, apiKey, ollamaURL, systemPrompt string, logger *logrus.Logger) *OllamaHandler {
	return &OllamaHandler{
		systemMsg: systemPrompt,
		logger:    logger,
		ctx:       ctx,
		apiKey:    apiKey,
		ollamaURL: ollamaURL,
	}
}

// QueryStream processes the LLM response as a stream for Ollama
func (h *OllamaHandler) QueryStream(model, text string, ttsCallback func(segment string, playID string, autoHangup bool) error) (string, error) {
	// Prepare the request to Ollama's API
	requestBody := map[string]interface{}{
		"model": model,
		"text":  text,
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make the HTTP request to Ollama's API
	req, err := http.NewRequestWithContext(h.ctx, "POST", fmt.Sprintf("%s/query/stream", h.ollamaURL), bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+h.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Process the streaming response (this is a placeholder for real streaming logic)
	// For now, assume a simple response format
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	// Here you would process the stream response in segments
	// For now, we're simulating sending a TTS segment
	segment := "Ollama response stream"
	if err := ttsCallback(segment, "ollama-play-id", false); err != nil {
		return "", fmt.Errorf("failed to send TTS segment: %w", err)
	}

	return segment, nil
}

// Query queries the LLM with text and gets a response for Ollama
func (h *OllamaHandler) Query(model, text string) (string, *HangupTool, error) {
	// Prepare the request to Ollama's API
	requestBody := map[string]interface{}{
		"model": model,
		"text":  text,
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make the HTTP request to Ollama's API
	req, err := http.NewRequestWithContext(h.ctx, "POST", fmt.Sprintf("%s/query", h.ollamaURL), bytes.NewReader(body))
	if err != nil {
		return "", nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+h.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Decode the response from Ollama
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract the response content (this is a placeholder for real response content)
	content := response["text"].(string)

	// Returning the response and simulating no hangup
	return content, nil, nil
}

// Reset clears the conversation history for Ollama
func (h *OllamaHandler) Reset() {
	// Reset logic for Ollama (e.g., clear messages)
	h.messages = []string{h.systemMsg}
}
