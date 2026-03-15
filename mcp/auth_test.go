package mcp_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/loopforge-ai/utils/assert"
	"github.com/loopforge-ai/utils/mcp"
)

const authTestRequest = `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`

func newAuthTestServer() *httptest.Server {
	server := mcp.NewServer("test", "1.0")
	server.RegisterTool(mcp.NewTool(
		"echo",
		"Echo the input back.",
		mcp.NewObjectSchema(
			map[string]mcp.Property{
				"message": mcp.NewStringProperty("Message to echo."),
			},
			[]string{"message"},
		),
		func(_ context.Context, params mcp.ToolsCallParams) (mcp.ToolsCallResult, error) {
			msg := params.Arguments["message"].(string)
			return mcp.TextResult(msg), nil
		},
	))

	// Auto-initialize for HTTP.
	server.HandleJSONRPC(context.Background(), mcp.Request{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`1`),
		Method:  "initialize",
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		body, _ := io.ReadAll(r.Body)
		defer func() { _ = r.Body.Close() }()

		var req mcp.Request
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, "Parse error", http.StatusBadRequest)
			return
		}

		resp := server.HandleJSONRPC(r.Context(), req)
		w.Header().Set("Content-Type", "application/json")
		data, mErr := json.Marshal(resp)
		if mErr != nil {
			http.Error(w, "marshal error", http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(data)
	})

	return httptest.NewServer(mcp.BearerAuth("secret-token", handler))
}

func Test_BearerAuth_With_InvalidToken_Should_Return401(t *testing.T) {
	t.Parallel()
	// Arrange
	ts := newAuthTestServer()
	defer ts.Close()
	req, _ := http.NewRequest(http.MethodPost, ts.URL, strings.NewReader(authTestRequest))
	req.Header.Set("Authorization", "Bearer wrong-token")

	// Act
	resp, err := http.DefaultClient.Do(req)

	// Assert
	assert.That(t, "error should be nil", err, nil)
	defer func() { _ = resp.Body.Close() }()
	assert.That(t, "status should be 401", resp.StatusCode, http.StatusUnauthorized)
}

func Test_BearerAuth_With_MissingHeader_Should_Return401(t *testing.T) {
	t.Parallel()
	// Arrange
	ts := newAuthTestServer()
	defer ts.Close()

	// Act
	resp, err := http.Post(ts.URL, "application/json", strings.NewReader(authTestRequest))

	// Assert
	assert.That(t, "error should be nil", err, nil)
	defer func() { _ = resp.Body.Close() }()
	assert.That(t, "status should be 401", resp.StatusCode, http.StatusUnauthorized)
}

func Test_BearerAuth_With_ValidToken_Should_PassThrough(t *testing.T) {
	t.Parallel()
	// Arrange
	ts := newAuthTestServer()
	defer ts.Close()
	req, _ := http.NewRequest(http.MethodPost, ts.URL, strings.NewReader(authTestRequest))
	req.Header.Set("Authorization", "Bearer secret-token")

	// Act
	resp, err := http.DefaultClient.Do(req)

	// Assert
	assert.That(t, "error should be nil", err, nil)
	defer func() { _ = resp.Body.Close() }()
	assert.That(t, "status should be 200", resp.StatusCode, http.StatusOK)
	body, _ := io.ReadAll(resp.Body)
	assert.That(t, "should contain echo tool", strings.Contains(string(body), "echo"), true)
}

func Test_BearerAuth_With_WrongScheme_Should_Return401(t *testing.T) {
	t.Parallel()
	// Arrange
	ts := newAuthTestServer()
	defer ts.Close()
	req, _ := http.NewRequest(http.MethodPost, ts.URL, strings.NewReader(authTestRequest))
	req.Header.Set("Authorization", "Basic secret-token")

	// Act
	resp, err := http.DefaultClient.Do(req)

	// Assert
	assert.That(t, "error should be nil", err, nil)
	defer func() { _ = resp.Body.Close() }()
	assert.That(t, "status should be 401", resp.StatusCode, http.StatusUnauthorized)
}
