package llm_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/loopforge-ai/utils/assert"
	"github.com/loopforge-ai/utils/llm"
)

const (
	testAPIKey   = "not-needed-locally"
	testEndpoint = "http://localhost:1234/v1/chat/completions"
	testModel    = "qwen/qwen3-coder-30b"
)

func fakeCompletionResponse(t *testing.T, content string) []byte {
	t.Helper()
	resp := map[string]any{
		"choices": []map[string]any{
			{"message": map[string]string{"content": content}},
		},
	}
	b, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("fakeCompletionResponse: %v", err)
	}
	return b
}

func Test_Complete_With_CancelledContext_Should_ReturnError(t *testing.T) {
	t.Parallel()
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(fakeCompletionResponse(t, "ok"))
	}))
	defer server.Close()
	c := llm.NewOpenAICompleter(server.URL, testModel, testAPIKey, llm.DefaultHTTPTimeout)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Act
	result, err := c.Complete(ctx, "sys", "usr")

	// Assert
	assert.That(t, "error should not be nil", err != nil, true)
	assert.That(t, "result should be empty", result, "")
}

func Test_Complete_With_EmptyAPIKey_Should_OmitAuthHeader(t *testing.T) {
	t.Parallel()
	// Arrange
	var receivedAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(fakeCompletionResponse(t, "ok"))
	}))
	defer server.Close()
	c := llm.NewOpenAICompleter(server.URL, testModel, "", llm.DefaultHTTPTimeout)
	ctx := context.Background()

	// Act
	_, _ = c.Complete(ctx, "sys", "usr")

	// Assert
	assert.That(t, "authorization should be empty", receivedAuth, "")
}

func Test_Complete_With_EmptyChoices_Should_ReturnError(t *testing.T) {
	t.Parallel()
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp, _ := json.Marshal(map[string]any{"choices": []any{}})
		_, _ = w.Write(resp)
	}))
	defer server.Close()
	c := llm.NewOpenAICompleter(server.URL, testModel, testAPIKey, llm.DefaultHTTPTimeout)
	ctx := context.Background()

	// Act
	result, err := c.Complete(ctx, "sys", "usr")

	// Assert
	assert.That(t, "error should not be nil", err != nil, true)
	assert.That(t, "result should be empty", result, "")
}

func Test_Complete_With_EmptyContent_Should_ReturnError(t *testing.T) {
	t.Parallel()
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(fakeCompletionResponse(t, ""))
	}))
	defer server.Close()
	c := llm.NewOpenAICompleter(server.URL, testModel, testAPIKey, llm.DefaultHTTPTimeout)
	ctx := context.Background()

	// Act
	result, err := c.Complete(ctx, "sys", "usr")

	// Assert
	assert.That(t, "error should not be nil", err != nil, true)
	assert.That(t, "result should be empty", result, "")
}

func Test_Complete_With_InvalidJSON_Should_ReturnError(t *testing.T) {
	t.Parallel()
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("not json"))
	}))
	defer server.Close()
	c := llm.NewOpenAICompleter(server.URL, testModel, testAPIKey, llm.DefaultHTTPTimeout)
	ctx := context.Background()

	// Act
	result, err := c.Complete(ctx, "sys", "usr")

	// Assert
	assert.That(t, "error should not be nil", err != nil, true)
	assert.That(t, "result should be empty", result, "")
}

func Test_Complete_With_Non200Status_Should_ReturnError(t *testing.T) {
	t.Parallel()
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}))
	defer server.Close()
	c := llm.NewOpenAICompleter(server.URL, testModel, testAPIKey, llm.DefaultHTTPTimeout)
	ctx := context.Background()

	// Act
	result, err := c.Complete(ctx, "sys", "usr")

	// Assert
	assert.That(t, "error should not be nil", err != nil, true)
	assert.That(t, "result should be empty", result, "")
}

func Test_Complete_With_UnreachableEndpoint_Should_ReturnError(t *testing.T) {
	t.Parallel()
	// Arrange
	c := llm.NewOpenAICompleter("http://127.0.0.1:0/v1/chat/completions", testModel, testAPIKey, llm.DefaultHTTPTimeout)
	ctx := context.Background()

	// Act
	result, err := c.Complete(ctx, "sys", "usr")

	// Assert
	assert.That(t, "error should not be nil", err != nil, true)
	assert.That(t, "result should be empty", result, "")
}

func Test_Complete_With_ValidResponse_Should_ReturnContent(t *testing.T) {
	t.Parallel()
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(fakeCompletionResponse(t, `{"name":"test_skill"}`))
	}))
	defer server.Close()
	c := llm.NewOpenAICompleter(server.URL, testModel, testAPIKey, llm.DefaultHTTPTimeout)
	ctx := context.Background()

	// Act
	result, err := c.Complete(ctx, "system prompt", "user prompt")

	// Assert
	assert.That(t, "error should be nil", err, nil)
	assert.That(t, "content should match", result, `{"name":"test_skill"}`)
}

func Test_Complete_With_ValidResponse_Should_SendCorrectRequest(t *testing.T) {
	t.Parallel()
	// Arrange
	var receivedBody map[string]any
	var receivedHeaders http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header
		_ = json.NewDecoder(r.Body).Decode(&receivedBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(fakeCompletionResponse(t, "ok"))
	}))
	defer server.Close()
	c := llm.NewOpenAICompleter(server.URL, testModel, testAPIKey, llm.DefaultHTTPTimeout)
	ctx := context.Background()

	// Act
	_, _ = c.Complete(ctx, "sys", "usr")

	// Assert
	assert.That(t, "model should match", receivedBody["model"], testModel)
	assert.That(t, "temperature should be 0", receivedBody["temperature"], 0.0)
	assert.That(t, "content-type should be json", receivedHeaders.Get("Content-Type"), "application/json")
	assert.That(t, "authorization should be set", receivedHeaders.Get("Authorization"), "Bearer "+testAPIKey)
	messages := receivedBody["messages"].([]any)
	assert.That(t, "should have 2 messages", len(messages), 2)
}

func Test_NewOpenAICompleter_Should_ReturnCompleter(t *testing.T) {
	t.Parallel()
	// Arrange & Act
	c := llm.NewOpenAICompleter(testEndpoint, testModel, testAPIKey, llm.DefaultHTTPTimeout)

	// Assert
	assert.That(t, "completer should not be nil", c != nil, true)
}
