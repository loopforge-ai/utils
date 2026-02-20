package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	// DefaultHTTPTimeout is the default timeout for OpenAI HTTP requests.
	DefaultHTTPTimeout = 5 * time.Minute
	maxErrorSnippet    = 256
	maxResponseBytes   = 10 << 20 // 10 MB
)

// compile-time interface check.
var _ Completer = (*OpenAICompleter)(nil)

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// OpenAICompleter is an implementation of skill.Completer using the
// OpenAI-compatible chat completions API (LM Studio, OpenAI, etc.).
type OpenAICompleter struct {
	client   *http.Client
	endpoint string
	model    string
	apiKey   string
}

// NewOpenAICompleter creates a new completer.
// endpoint is the full URL, e.g. "http://localhost:1234/v1/chat/completions".
// apiKey may be empty for local servers like LM Studio.
// timeout controls the HTTP client timeout; use DefaultHTTPTimeout if unsure.
func NewOpenAICompleter(endpoint, model, apiKey string, timeout time.Duration) *OpenAICompleter {
	return &OpenAICompleter{
		client:   &http.Client{Timeout: timeout},
		endpoint: endpoint,
		model:    model,
		apiKey:   apiKey,
	}
}

// Complete sends system and user prompts to the chat completions endpoint
// and returns the assistant's reply.
func (c *OpenAICompleter) Complete(ctx context.Context, system, user string) (string, error) {
	body, statusCode, err := c.doRequest(ctx, system, user)
	if err != nil {
		return "", err
	}

	if statusCode != http.StatusOK {
		snippet := string(body)
		if len(snippet) > maxErrorSnippet {
			snippet = snippet[:maxErrorSnippet] + "... (truncated)"
		}
		return "", fmt.Errorf("unexpected status %d: %s", statusCode, snippet)
	}

	var chatResp chatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}

	if len(chatResp.Choices) == 0 || chatResp.Choices[0].Message.Content == "" {
		return "", errors.New("openai completer: empty response")
	}

	return chatResp.Choices[0].Message.Content, nil
}

// buildHTTPRequest creates the HTTP request for the chat completions endpoint.
func (c *OpenAICompleter) buildHTTPRequest(ctx context.Context, system, user string) (*http.Request, error) {
	reqBody := chatRequest{
		Model: c.model,
		Messages: []chatMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		Temperature: 0,
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	return req, nil
}

// doRequest builds, sends, and reads the HTTP response.
func (c *OpenAICompleter) doRequest(ctx context.Context, system, user string) ([]byte, int, error) {
	req, err := c.buildHTTPRequest(ctx, system, user)
	if err != nil {
		return nil, 0, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return nil, 0, fmt.Errorf("read response: %w", err)
	}

	return body, resp.StatusCode, nil
}
