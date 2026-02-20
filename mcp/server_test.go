package mcp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/loopforge-ai/utils/assert"
	"github.com/loopforge-ai/utils/mcp"
)

func echoTool() mcp.Tool {
	return mcp.NewTool(
		"echo",
		"echoes input",
		mcp.NewObjectSchema(map[string]mcp.Property{
			"message": mcp.NewStringProperty("message to echo"),
		}, []string{"message"}),
		func(_ context.Context, in mcp.ToolsCallParams) (mcp.ToolsCallResult, error) {
			msg, _ := in.Arguments["message"].(string)
			return mcp.ToolsCallResult{
				Content: []mcp.ContentBlock{mcp.NewTextContent(msg)},
			}, nil
		},
	)
}

func marshalResult(t *testing.T, result any) string {
	t.Helper()
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal result: %v", err)
	}
	return string(data)
}

func parseResponse(t *testing.T, line string) mcp.Response {
	t.Helper()
	var resp mcp.Response
	if err := json.Unmarshal([]byte(line), &resp); err != nil {
		t.Fatalf("unmarshal response %q: %v", line, err)
	}
	return resp
}

func runServer(t *testing.T, input string, tools ...mcp.Tool) string {
	t.Helper()
	reader := strings.NewReader(input)
	var writer bytes.Buffer
	server := mcp.NewServerWithIO("test", "0.0.1", reader, &writer)
	for _, tool := range tools {
		server.RegisterTool(tool)
	}
	_ = server.Serve(context.Background())
	return writer.String()
}

func sendRequest(t *testing.T, method string, id json.RawMessage, params any) string {
	t.Helper()
	req := map[string]any{
		"jsonrpc": "2.0",
		"method":  method,
	}
	if id != nil {
		req["id"] = id
	}
	if params != nil {
		raw, err := json.Marshal(params)
		if err != nil {
			t.Fatalf("marshal params: %v", err)
		}
		req["params"] = json.RawMessage(raw)
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	return string(data) + "\n"
}

func Test_Initialize_Should_ReturnCapabilities(t *testing.T) {
	t.Parallel()
	// Arrange
	id := json.RawMessage(`1`)
	input := sendRequest(t, "initialize", id, nil)

	// Act
	output := runServer(t, input)
	lines := strings.Split(strings.TrimSpace(output), "\n")
	resp := parseResponse(t, lines[0])

	// Assert
	assert.That(t, "error should be nil", resp.Error, (*mcp.ResponseError)(nil))
	assert.That(t, "jsonrpc should be 2.0", resp.JSONRPC, "2.0")
	result := marshalResult(t, resp.Result)
	assert.That(t, "result should contain serverInfo", strings.Contains(result, `"name":"test"`), true)
	assert.That(t, "result should contain version", strings.Contains(result, `"version":"0.0.1"`), true)
}

func Test_Initialized_Should_ProduceNoResponse(t *testing.T) {
	t.Parallel()
	// Arrange
	init := sendRequest(t, "initialize", json.RawMessage(`1`), nil)
	initialized := sendRequest(t, "initialized", nil, nil)
	input := init + initialized

	// Act
	output := runServer(t, input)
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Assert — only the initialize response, nothing for initialized
	assert.That(t, "should have 1 response line", len(lines), 1)
}

func Test_MalformedJSON_Should_ReturnParseError(t *testing.T) {
	t.Parallel()
	// Arrange
	input := "this is not json\n"

	// Act
	output := runServer(t, input)
	resp := parseResponse(t, strings.TrimSpace(output))

	// Assert
	assert.That(t, "should have error", resp.Error != nil, true)
	assert.That(t, "error code should be parse error", resp.Error.Code, mcp.ErrorCodeParse)
}

func Test_NewTool_With_EmptyName_Should_Panic(t *testing.T) {
	t.Parallel()
	// Arrange
	defer func() {
		// Assert
		r := recover()
		assert.That(t, "should panic", r != nil, true)
		assert.That(t, "panic message", r, "mcp: tool name must not be empty")
	}()

	// Act
	mcp.NewTool("", "desc", mcp.NewObjectSchema(nil, nil), func(_ context.Context, _ mcp.ToolsCallParams) (mcp.ToolsCallResult, error) {
		return mcp.ToolsCallResult{}, nil
	})
}

func Test_NewTool_With_NilHandler_Should_Panic(t *testing.T) {
	t.Parallel()
	// Arrange
	defer func() {
		// Assert
		r := recover()
		assert.That(t, "should panic", r != nil, true)
		assert.That(t, "panic message", r, "mcp: tool handler must not be nil")
	}()

	// Act
	mcp.NewTool("test", "desc", mcp.NewObjectSchema(nil, nil), nil)
}

func Test_NewServer_Should_SetNameAndVersion(t *testing.T) {
	t.Parallel()
	// Arrange & Act
	server := mcp.NewServerWithIO("myserver", "1.2.3", strings.NewReader(""), &bytes.Buffer{})

	// Assert
	assert.That(t, "name", server.Name(), "myserver")
	assert.That(t, "version", server.Version(), "1.2.3")
}

func Test_RegisterTool_Should_AppearInTools(t *testing.T) {
	t.Parallel()
	// Arrange
	server := mcp.NewServerWithIO("test", "0.0.1", strings.NewReader(""), &bytes.Buffer{})

	// Act
	server.RegisterTool(echoTool())

	// Assert
	tools := server.Tools()
	assert.That(t, "should have 1 tool", len(tools), 1)
	assert.That(t, "tool name", tools[0].Definition.Name, "echo")
}

func Test_Serve_With_CanceledContext_Should_ReturnContextError(t *testing.T) {
	t.Parallel()
	// Arrange
	reader := strings.NewReader("")
	var writer bytes.Buffer
	server := mcp.NewServerWithIO("test", "0.0.1", reader, &writer)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Act
	err := server.Serve(ctx)

	// Assert
	assert.That(t, "error should wrap context.Canceled", errors.Is(err, context.Canceled), true)
}

func Test_Serve_With_EOF_Should_ReturnNil(t *testing.T) {
	t.Parallel()
	// Arrange
	reader := strings.NewReader("")
	var writer bytes.Buffer
	server := mcp.NewServerWithIO("test", "0.0.1", reader, &writer)

	// Act
	err := server.Serve(context.Background())

	// Assert
	assert.That(t, "error should be nil", err, nil)
}

func Test_ToolsCall_With_HandlerError_Should_ReturnIsError(t *testing.T) {
	t.Parallel()
	// Arrange
	failTool := mcp.NewTool(
		"fail",
		"always fails",
		mcp.NewObjectSchema(nil, nil),
		func(_ context.Context, _ mcp.ToolsCallParams) (mcp.ToolsCallResult, error) {
			return mcp.ToolsCallResult{}, context.DeadlineExceeded
		},
	)
	init := sendRequest(t, "initialize", json.RawMessage(`1`), nil)
	initialized := sendRequest(t, "initialized", nil, nil)
	call := sendRequest(t, "tools/call", json.RawMessage(`2`), map[string]any{"name": "fail"})
	input := init + initialized + call

	// Act
	output := runServer(t, input, failTool)
	lines := strings.Split(strings.TrimSpace(output), "\n")
	resp := parseResponse(t, lines[1])

	// Assert
	assert.That(t, "error should be nil (tool errors are in result)", resp.Error, (*mcp.ResponseError)(nil))
	result := marshalResult(t, resp.Result)
	assert.That(t, "should contain isError", strings.Contains(result, `"isError":true`), true)
}

func Test_ToolsCall_With_NotInitialized_Should_ReturnError(t *testing.T) {
	t.Parallel()
	// Arrange
	id := json.RawMessage(`3`)
	input := sendRequest(t, "tools/call", id, map[string]any{"name": "echo"})

	// Act
	output := runServer(t, input, echoTool())
	resp := parseResponse(t, strings.TrimSpace(output))

	// Assert
	assert.That(t, "should have error", resp.Error != nil, true)
	assert.That(t, "error message", resp.Error.Message, "server not initialized")
}

func Test_ToolsCall_With_UnknownTool_Should_ReturnError(t *testing.T) {
	t.Parallel()
	// Arrange
	init := sendRequest(t, "initialize", json.RawMessage(`1`), nil)
	initialized := sendRequest(t, "initialized", nil, nil)
	call := sendRequest(t, "tools/call", json.RawMessage(`2`), map[string]any{"name": "nope"})
	input := init + initialized + call

	// Act
	output := runServer(t, input)
	lines := strings.Split(strings.TrimSpace(output), "\n")
	resp := parseResponse(t, lines[1])

	// Assert
	assert.That(t, "should have error", resp.Error != nil, true)
	assert.That(t, "error message", resp.Error.Message, "tool not found")
}

func Test_ToolsCall_With_ValidTool_Should_ReturnResult(t *testing.T) {
	t.Parallel()
	// Arrange
	init := sendRequest(t, "initialize", json.RawMessage(`1`), nil)
	initialized := sendRequest(t, "initialized", nil, nil)
	call := sendRequest(t, "tools/call", json.RawMessage(`2`), map[string]any{
		"name":      "echo",
		"arguments": map[string]any{"message": "hello"},
	})
	input := init + initialized + call

	// Act
	output := runServer(t, input, echoTool())
	lines := strings.Split(strings.TrimSpace(output), "\n")
	resp := parseResponse(t, lines[1])

	// Assert
	assert.That(t, "error should be nil", resp.Error, (*mcp.ResponseError)(nil))
	result := marshalResult(t, resp.Result)
	assert.That(t, "should contain hello", strings.Contains(result, "hello"), true)
}

func Test_ToolsList_With_Initialized_Should_ReturnTools(t *testing.T) {
	t.Parallel()
	// Arrange
	init := sendRequest(t, "initialize", json.RawMessage(`1`), nil)
	initialized := sendRequest(t, "initialized", nil, nil)
	list := sendRequest(t, "tools/list", json.RawMessage(`2`), nil)
	input := init + initialized + list

	// Act
	output := runServer(t, input, echoTool())
	lines := strings.Split(strings.TrimSpace(output), "\n")
	// lines[0] = initialize response, lines[1] = tools/list response
	resp := parseResponse(t, lines[1])

	// Assert
	assert.That(t, "error should be nil", resp.Error, (*mcp.ResponseError)(nil))
	result := marshalResult(t, resp.Result)
	assert.That(t, "should contain echo tool", strings.Contains(result, `"name":"echo"`), true)
}

func Test_ToolsList_With_NotInitialized_Should_ReturnError(t *testing.T) {
	t.Parallel()
	// Arrange
	id := json.RawMessage(`2`)
	input := sendRequest(t, "tools/list", id, nil)

	// Act
	output := runServer(t, input)
	resp := parseResponse(t, strings.TrimSpace(output))

	// Assert
	assert.That(t, "should have error", resp.Error != nil, true)
	assert.That(t, "error message", resp.Error.Message, "server not initialized")
}

func Test_UnknownMethod_Should_ReturnMethodNotFound(t *testing.T) {
	t.Parallel()
	// Arrange
	id := json.RawMessage(`99`)
	input := sendRequest(t, "bogus/method", id, nil)

	// Act
	output := runServer(t, input)
	resp := parseResponse(t, strings.TrimSpace(output))

	// Assert
	assert.That(t, "should have error", resp.Error != nil, true)
	assert.That(t, "error code should be method not found", resp.Error.Code, mcp.ErrorCodeMethodNotFound)
}

func Test_RegisterTool_With_ConcurrentRegistration_Should_NotRace(t *testing.T) {
	t.Parallel()
	// Arrange
	server := mcp.NewServer("test", "1.0")
	const goroutines = 10

	// Act
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := range goroutines {
		go func() {
			defer wg.Done()
			tool := mcp.NewTool(
				fmt.Sprintf("tool_%d", i),
				"test tool",
				mcp.NewObjectSchema(nil, nil),
				func(_ context.Context, _ mcp.ToolsCallParams) (mcp.ToolsCallResult, error) {
					return mcp.ToolsCallResult{}, nil
				},
			)
			server.RegisterTool(tool)
		}()
	}
	wg.Wait()

	// Assert
	assert.That(t, "should have 10 tools", len(server.Tools()), goroutines)
}

func Test_ToolsCall_With_PanickingHandler_Should_ReturnInternalError(t *testing.T) {
	t.Parallel()
	// Arrange
	id := json.RawMessage(`3`)
	panicTool := mcp.NewTool(
		"panic_tool",
		"panics on call",
		mcp.NewObjectSchema(nil, nil),
		func(_ context.Context, _ mcp.ToolsCallParams) (mcp.ToolsCallResult, error) {
			panic("unexpected failure")
		},
	)
	init := sendRequest(t, "initialize", json.RawMessage(`1`), nil)
	initialized := sendRequest(t, "initialized", nil, nil)
	call := sendRequest(t, "tools/call", id, map[string]any{"name": "panic_tool"})
	input := init + initialized + call

	// Act
	output := runServer(t, input, panicTool)

	// Assert
	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.That(t, "should have 2 responses", len(lines), 2)
	resp := parseResponse(t, lines[1])
	assert.That(t, "should have error", resp.Error != nil, true)
	assert.That(t, "should be internal error", resp.Error.Code, -32603)
	assert.That(t, "should mention failure", strings.Contains(resp.Error.Message, "unexpected failure"), true)
}

func Test_Initialize_With_MissingJSONRPC_Should_ReturnError(t *testing.T) {
	t.Parallel()
	// Arrange — send initialize without jsonrpc field
	req := `{"method":"initialize","id":1}` + "\n"

	// Act
	output := runServer(t, req)
	resp := parseResponse(t, strings.TrimSpace(output))

	// Assert
	assert.That(t, "should have error", resp.Error != nil, true)
	assert.That(t, "error code should be invalid request", resp.Error.Code, mcp.ErrorCodeInvalidRequest)
	assert.That(t, "should mention jsonrpc", strings.Contains(resp.Error.Message, "jsonrpc"), true)
}

func Test_HandleRequest_With_OversizedInput_Should_ReturnParseError(t *testing.T) {
	t.Parallel()
	// Arrange
	huge := strings.Repeat("x", 11<<20) + "\n" // 11 MB
	var writer bytes.Buffer
	server := mcp.NewServerWithIO("test", "0.0.1", strings.NewReader(huge), &writer)

	// Act
	_ = server.Serve(context.Background())

	// Assert
	output := writer.String()
	assert.That(t, "should have response", output != "", true)
	resp := parseResponse(t, strings.TrimSpace(output))
	assert.That(t, "should have error", resp.Error != nil, true)
	assert.That(t, "should be parse error", resp.Error.Code, -32700)
	assert.That(t, "should mention size", strings.Contains(resp.Error.Message, "too large"), true)
}
