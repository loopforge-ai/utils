package mcp_test

import (
	"strings"
	"testing"

	"github.com/loopforge-ai/utils/assert"
	"github.com/loopforge-ai/utils/mcp"
)

func Test_ErrResult_With_FormatArgs_Should_ReturnFormattedError(t *testing.T) {
	t.Parallel()
	// Arrange & Act
	res := mcp.ErrResult("failed: %s (%d)", "disk", 42)

	// Assert
	assert.That(t, "should be error", res.IsError, true)
	assert.That(t, "should have content", len(res.Content) > 0, true)
	text := res.Content[0].Text
	assert.That(t, "should contain formatted message", strings.Contains(text, "failed: disk (42)"), true)
}

func Test_JSONResult_With_UnmarshalableValue_Should_ReturnError(t *testing.T) {
	t.Parallel()
	// Arrange
	v := make(chan int)

	// Act
	res := mcp.JSONResult(v)

	// Assert
	assert.That(t, "should be error", res.IsError, true)
	assert.That(t, "should have content", len(res.Content) > 0, true)
	assert.That(t, "should mention marshal", strings.Contains(res.Content[0].Text, "marshal"), true)
}

func Test_JSONResult_With_ValidStruct_Should_ReturnJSON(t *testing.T) {
	t.Parallel()
	// Arrange
	v := struct {
		Name string `json:"name"`
	}{Name: "test"}

	// Act
	res := mcp.JSONResult(v)

	// Assert
	assert.That(t, "should not be error", res.IsError, false)
	assert.That(t, "should have content", len(res.Content) > 0, true)
	assert.That(t, "should contain name", strings.Contains(res.Content[0].Text, `"name": "test"`), true)
}

func Test_OptionalJSONMap_With_InvalidJSON_Should_ReturnError(t *testing.T) {
	t.Parallel()
	// Arrange
	params := mcp.ToolsCallParams{
		Arguments: map[string]any{"args": "not json"},
	}

	// Act
	_, errRes := mcp.OptionalJSONMap(params, "args")

	// Assert
	assert.That(t, "error result should not be nil", errRes != nil, true)
	assert.That(t, "should be error", errRes.IsError, true)
}

func Test_OptionalJSONMap_With_MissingKey_Should_ReturnNil(t *testing.T) {
	t.Parallel()
	// Arrange
	params := mcp.ToolsCallParams{
		Arguments: map[string]any{},
	}

	// Act
	m, errRes := mcp.OptionalJSONMap(params, "args")

	// Assert
	assert.That(t, "error result should be nil", errRes == nil, true)
	assert.That(t, "map should be nil", m == nil, true)
}

func Test_OptionalJSONMap_With_NonStringValue_Should_ReturnError(t *testing.T) {
	t.Parallel()
	// Arrange
	params := mcp.ToolsCallParams{
		Arguments: map[string]any{"args": 42},
	}

	// Act
	m, errRes := mcp.OptionalJSONMap(params, "args")

	// Assert
	assert.That(t, "map should be nil", m == nil, true)
	assert.That(t, "error result should not be nil", errRes != nil, true)
	assert.That(t, "should be error", errRes.IsError, true)
	assert.That(t, "should mention string", strings.Contains(errRes.Content[0].Text, "string"), true)
}

func Test_OptionalJSONMap_With_ValidJSON_Should_ReturnMap(t *testing.T) {
	t.Parallel()
	// Arrange
	params := mcp.ToolsCallParams{
		Arguments: map[string]any{"args": `{"key":"val"}`},
	}

	// Act
	m, errRes := mcp.OptionalJSONMap(params, "args")

	// Assert
	assert.That(t, "error result should be nil", errRes == nil, true)
	assert.That(t, "map should have entry", m["key"], "val")
}

func Test_SuccessWithSummary_With_EmptySummary_Should_ReturnMessageOnly(t *testing.T) {
	t.Parallel()
	// Arrange & Act
	res := mcp.SuccessWithSummary("done", "")

	// Assert
	assert.That(t, "should not be error", res.IsError, false)
	assert.That(t, "should have one block", len(res.Content), 1)
	assert.That(t, "should contain message", res.Content[0].Text, "done")
}

func Test_SuccessWithSummary_With_Summary_Should_AppendBlock(t *testing.T) {
	t.Parallel()
	// Arrange & Act
	res := mcp.SuccessWithSummary("done", "stats: 42 tok/s")

	// Assert
	assert.That(t, "should not be error", res.IsError, false)
	assert.That(t, "should have two blocks", len(res.Content), 2)
	assert.That(t, "first block should be message", res.Content[0].Text, "done")
	assert.That(t, "second block should be summary", res.Content[1].Text, "stats: 42 tok/s")
}

func Test_RequireStrings_With_AllPresent_Should_ReturnMap(t *testing.T) {
	t.Parallel()
	// Arrange
	params := mcp.ToolsCallParams{
		Arguments: map[string]any{"a": "one", "b": "two"},
	}

	// Act
	vals, errRes := mcp.RequireStrings(params, "a", "b")

	// Assert
	assert.That(t, "error result should be nil", errRes == nil, true)
	assert.That(t, "a should match", vals["a"], "one")
	assert.That(t, "b should match", vals["b"], "two")
}

func Test_RequireStrings_With_MissingKey_Should_ReturnError(t *testing.T) {
	t.Parallel()
	// Arrange
	params := mcp.ToolsCallParams{
		Arguments: map[string]any{"a": "one"},
	}

	// Act
	vals, errRes := mcp.RequireStrings(params, "a", "b")

	// Assert
	assert.That(t, "vals should be nil", vals == nil, true)
	assert.That(t, "error result should not be nil", errRes != nil, true)
	assert.That(t, "should be error", errRes.IsError, true)
	assert.That(t, "should mention missing", strings.Contains(errRes.Content[0].Text, "missing"), true)
}

func Test_OptionalJSONMap_With_EmptyString_Should_ReturnNil(t *testing.T) {
	t.Parallel()
	// Arrange
	params := mcp.ToolsCallParams{
		Arguments: map[string]any{"args": ""},
	}
	// Act
	m, errRes := mcp.OptionalJSONMap(params, "args")
	// Assert
	assert.That(t, "error result should be nil", errRes == nil, true)
	assert.That(t, "map should be nil", m == nil, true)
}

func Test_OptionalJSONMap_With_NilArguments_Should_ReturnNil(t *testing.T) {
	t.Parallel()
	// Arrange
	params := mcp.ToolsCallParams{Arguments: nil}
	// Act
	m, errRes := mcp.OptionalJSONMap(params, "args")
	// Assert
	assert.That(t, "error result should be nil", errRes == nil, true)
	assert.That(t, "map should be nil", m == nil, true)
}

func Test_TextResult_With_Message_Should_ReturnTextContent(t *testing.T) {
	t.Parallel()
	// Arrange
	message := "operation completed"
	// Act
	res := mcp.TextResult(message)
	// Assert
	assert.That(t, "should not be error", res.IsError, false)
	assert.That(t, "should have one content block", len(res.Content), 1)
	assert.That(t, "content type should be text", res.Content[0].Type, "text")
	assert.That(t, "content text should match", res.Content[0].Text, "operation completed")
}

func Test_RequireString_With_MissingParam_Should_ReturnError(t *testing.T) {
	t.Parallel()
	// Arrange
	params := mcp.ToolsCallParams{
		Arguments: map[string]any{},
	}

	// Act
	val, errRes := mcp.RequireString(params, "name")

	// Assert
	assert.That(t, "value should be empty", val, "")
	assert.That(t, "error result should not be nil", errRes != nil, true)
	assert.That(t, "should be error", errRes.IsError, true)
	assert.That(t, "should mention missing", strings.Contains(errRes.Content[0].Text, "missing"), true)
}

func Test_RequireString_With_NonStringValue_Should_ReturnTypeError(t *testing.T) {
	t.Parallel()
	// Arrange
	params := mcp.ToolsCallParams{
		Arguments: map[string]any{"name": 42},
	}

	// Act
	val, errRes := mcp.RequireString(params, "name")

	// Assert
	assert.That(t, "value should be empty", val, "")
	assert.That(t, "error result should not be nil", errRes != nil, true)
	assert.That(t, "should be error", errRes.IsError, true)
	assert.That(t, "should mention string", strings.Contains(errRes.Content[0].Text, "must be a string"), true)
}

func Test_RequireString_With_ValidParam_Should_ReturnValue(t *testing.T) {
	t.Parallel()
	// Arrange
	params := mcp.ToolsCallParams{
		Arguments: map[string]any{"name": "hello"},
	}

	// Act
	val, errRes := mcp.RequireString(params, "name")

	// Assert
	assert.That(t, "value should match", val, "hello")
	assert.That(t, "error result should be nil", errRes == nil, true)
}

func Test_RequireString_With_WhitespaceOnly_Should_ReturnError(t *testing.T) {
	t.Parallel()
	// Arrange
	params := mcp.ToolsCallParams{
		Arguments: map[string]any{"name": "   "},
	}

	// Act
	val, errRes := mcp.RequireString(params, "name")

	// Assert
	assert.That(t, "value should be empty", val, "")
	assert.That(t, "error result should not be nil", errRes != nil, true)
	assert.That(t, "should be error", errRes.IsError, true)
}
