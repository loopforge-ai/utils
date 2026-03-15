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

func newTestHTTPServer() *httptest.Server {
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

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		body, _ := io.ReadAll(r.Body)
		defer func() { _ = r.Body.Close() }()

		var req mcp.Request
		if err := json.Unmarshal(body, &req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			resp := mcp.NewErrorResponse(nil, mcp.ErrorCodeParse, "Parse error")
			_ = json.NewEncoder(w).Encode(resp)
			return
		}

		resp := server.HandleJSONRPC(r.Context(), req)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
}

func Test_HTTPHandler_With_InvalidJSON_Should_ReturnParseError(t *testing.T) {
	t.Parallel()
	// Arrange
	ts := newTestHTTPServer()
	defer ts.Close()

	// Act
	resp, err := http.Post(ts.URL, "application/json", strings.NewReader(`{invalid`))

	// Assert
	assert.That(t, "error should be nil", err, nil)
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	assert.That(t, "should contain error", strings.Contains(string(body), "error"), true)
}

func Test_HTTPHandler_With_ToolsCall_Should_ReturnResult(t *testing.T) {
	t.Parallel()
	// Arrange
	ts := newTestHTTPServer()
	defer ts.Close()
	reqBody := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"echo","arguments":{"message":"hello"}}}`

	// Act
	resp, err := http.Post(ts.URL, "application/json", strings.NewReader(reqBody))

	// Assert
	assert.That(t, "error should be nil", err, nil)
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	assert.That(t, "should contain hello", strings.Contains(string(body), "hello"), true)
}

func Test_HTTPHandler_With_ToolsList_Should_ReturnTools(t *testing.T) {
	t.Parallel()
	// Arrange
	ts := newTestHTTPServer()
	defer ts.Close()
	reqBody := `{"jsonrpc":"2.0","id":2,"method":"tools/list"}`

	// Act
	resp, err := http.Post(ts.URL, "application/json", strings.NewReader(reqBody))

	// Assert
	assert.That(t, "error should be nil", err, nil)
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	assert.That(t, "should contain echo tool", strings.Contains(string(body), "echo"), true)
}

func Test_HTTPHandler_With_WrongMethod_Should_Return405(t *testing.T) {
	t.Parallel()
	// Arrange
	ts := newTestHTTPServer()
	defer ts.Close()

	// Act
	resp, err := http.Get(ts.URL)

	// Assert
	assert.That(t, "error should be nil", err, nil)
	defer func() { _ = resp.Body.Close() }()
	assert.That(t, "status should be 405", resp.StatusCode, http.StatusMethodNotAllowed)
}
