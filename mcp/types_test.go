package mcp_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/loopforge-ai/utils/assert"
	"github.com/loopforge-ai/utils/mcp"
)

func Test_ErrorResponse_JSON_RoundTrip_Should_PreserveError(t *testing.T) {
	t.Parallel()
	// Arrange
	resp := mcp.NewErrorResponse(json.RawMessage(`1`), mcp.ErrorCodeMethodNotFound, "Method not found")

	// Act
	data, err := json.Marshal(resp)
	assert.That(t, "marshal error", err, nil)

	var decoded mcp.Response
	err = json.Unmarshal(data, &decoded)

	// Assert
	assert.That(t, "unmarshal error", err, nil)
	assert.That(t, "error should not be nil", decoded.Error != nil, true)
	assert.That(t, "error code", decoded.Error.Code, mcp.ErrorCodeMethodNotFound)
	assert.That(t, "error message", decoded.Error.Message, "Method not found")
}

func Test_NewErrorResponse_Should_SetErrorFields(t *testing.T) {
	t.Parallel()
	// Arrange
	id := json.RawMessage(`2`)

	// Act
	resp := mcp.NewErrorResponse(id, mcp.ErrorCodeParse, "parse error")

	// Assert
	assert.That(t, "jsonrpc", resp.JSONRPC, "2.0")
	assert.That(t, "error should not be nil", resp.Error != nil, true)
	assert.That(t, "error code", resp.Error.Code, mcp.ErrorCodeParse)
	assert.That(t, "error message", resp.Error.Message, "parse error")
	assert.That(t, "result should be nil", resp.Result, nil)
}

func Test_NewErrorResponse_With_NilID_Should_SetNilID(t *testing.T) {
	t.Parallel()
	// Arrange & Act
	resp := mcp.NewErrorResponse(nil, mcp.ErrorCodeInternal, "internal")

	// Assert
	assert.That(t, "id should be null", string(resp.ID), "null")
}

func Test_NewObjectSchema_Should_SetTypeObject(t *testing.T) {
	t.Parallel()
	// Arrange
	props := map[string]mcp.Property{
		"name": mcp.NewStringProperty("the name"),
	}

	// Act
	schema := mcp.NewObjectSchema(props, []string{"name"})

	// Assert
	assert.That(t, "type", schema.Type, "object")
	assert.That(t, "required", len(schema.Required), 1)
	assert.That(t, "required[0]", schema.Required[0], "name")
	assert.That(t, "properties count", len(schema.Properties), 1)
}

func Test_NewResponse_Should_SetFields(t *testing.T) {
	t.Parallel()
	// Arrange
	id := json.RawMessage(`1`)

	// Act
	resp := mcp.NewResponse(id, "ok")

	// Assert
	assert.That(t, "jsonrpc", resp.JSONRPC, "2.0")
	assert.That(t, "id", string(resp.ID), "1")
	assert.That(t, "result", resp.Result, "ok")
	assert.That(t, "error should be nil", resp.Error, (*mcp.ResponseError)(nil))
}

func Test_NewStringProperty_Should_SetTypeString(t *testing.T) {
	t.Parallel()
	// Arrange & Act
	prop := mcp.NewStringProperty("a description")

	// Assert
	assert.That(t, "type", prop.Type, "string")
	assert.That(t, "description", prop.Description, "a description")
}

func Test_NewTextContent_Should_SetTypeAndText(t *testing.T) {
	t.Parallel()
	// Arrange & Act
	block := mcp.NewTextContent("hello")

	// Assert
	assert.That(t, "type", block.Type, "text")
	assert.That(t, "text", block.Text, "hello")
}

func Test_NewTool_Should_SetDefinitionAndHandler(t *testing.T) {
	t.Parallel()
	// Arrange & Act
	tool := mcp.NewTool("test_tool", "desc", mcp.NewObjectSchema(nil, nil), func(_ context.Context, _ mcp.ToolsCallParams) (mcp.ToolsCallResult, error) {
		return mcp.ToolsCallResult{}, nil
	})

	// Assert
	assert.That(t, "name", tool.Definition.Name, "test_tool")
	assert.That(t, "description", tool.Definition.Description, "desc")
	assert.That(t, "schema type", tool.Definition.InputSchema.Type, "object")
}

func Test_Response_JSON_RoundTrip_Should_PreserveFields(t *testing.T) {
	t.Parallel()
	// Arrange
	resp := mcp.NewResponse(json.RawMessage(`42`), map[string]string{"key": "value"})

	// Act
	data, err := json.Marshal(resp)
	assert.That(t, "marshal error", err, nil)

	var decoded mcp.Response
	err = json.Unmarshal(data, &decoded)

	// Assert
	assert.That(t, "unmarshal error", err, nil)
	assert.That(t, "jsonrpc", decoded.JSONRPC, "2.0")
	assert.That(t, "id", string(decoded.ID), "42")
}
