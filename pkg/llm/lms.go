package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

// LMStudioHandler implements the LLM interface for LM Studio
type LMStudioHandler struct {
	systemMsg   string
	messages    []string
	logger      *logrus.Logger
	ctx         context.Context
	apiKey      string
	lmStudioURL string
}

// NewLMStudioHandler creates a new LM Studio handler
func NewLMStudioHandler(ctx context.Context, apiKey, lmStudioURL, systemPrompt string, logger *logrus.Logger) *LMStudioHandler {
	return &LMStudioHandler{
		systemMsg:   systemPrompt,
		logger:      logger,
		ctx:         ctx,
		apiKey:      apiKey,
		lmStudioURL: lmStudioURL,
	}
}

// QueryStream processes the LLM response as a stream for LM Studio
func (h *LMStudioHandler) QueryStream(model, text string, ttsCallback func(segment string, playID string, autoHangup bool) error) (string, error) {
	// Prepare the request to LM Studio's API
	requestBody := map[string]interface{}{
		"model": model,
		"text":  text,
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make the HTTP request to LM Studio's API
	req, err := http.NewRequestWithContext(h.ctx, "POST", fmt.Sprintf("%s/query/stream", h.lmStudioURL), bytes.NewReader(body))
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
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	// Here you would process the stream response in segments
	// For now, we're simulating sending a TTS segment
	segment := "LM Studio response stream"
	if err := ttsCallback(segment, "lmstudio-play-id", false); err != nil {
		return "", fmt.Errorf("failed to send TTS segment: %w", err)
	}

	return segment, nil
}

// Query queries the LLM with text and gets a response for LM Studio
func (h *LMStudioHandler) Query(model, text string) (string, *HangupTool, error) {
	// Prepare the request to LM Studio's API
	requestBody := map[string]interface{}{
		"model": model,
		"text":  text,
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make the HTTP request to LM Studio's API
	req, err := http.NewRequestWithContext(h.ctx, "POST", fmt.Sprintf("%s/query", h.lmStudioURL), bytes.NewReader(body))
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

	// Decode the response from LM Studio
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract the response content (this is a placeholder for real response content)
	content := response["text"].(string)

	// Returning the response and simulating no hangup
	return content, nil, nil
}

// Reset clears the conversation history for LM Studio
func (h *LMStudioHandler) Reset() {
	// Reset logic for LM Studio (e.g., clear messages)
	h.messages = []string{h.systemMsg}
}
