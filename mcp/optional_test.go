package mcp_test

import (
	"testing"

	"github.com/loopforge-ai/utils/assert"
	"github.com/loopforge-ai/utils/mcp"
)

func Test_OptionalString_With_AbsentKey_Should_ReturnEmpty(t *testing.T) {
	t.Parallel()
	// Arrange
	params := mcp.ToolsCallParams{
		Arguments: map[string]any{"other": "value"},
	}

	// Act
	result := mcp.OptionalString(params, "template_repo")

	// Assert
	assert.That(t, "should return empty string", result, "")
}

func Test_OptionalString_With_EmptyString_Should_ReturnEmpty(t *testing.T) {
	t.Parallel()
	// Arrange
	params := mcp.ToolsCallParams{
		Arguments: map[string]any{"template_repo": ""},
	}

	// Act
	result := mcp.OptionalString(params, "template_repo")

	// Assert
	assert.That(t, "should return empty string", result, "")
}

func Test_OptionalString_With_NilArguments_Should_ReturnEmpty(t *testing.T) {
	t.Parallel()
	// Arrange
	params := mcp.ToolsCallParams{Arguments: nil}

	// Act
	result := mcp.OptionalString(params, "template_repo")

	// Assert
	assert.That(t, "should return empty string", result, "")
}

func Test_OptionalString_With_NonStringType_Should_ReturnEmpty(t *testing.T) {
	t.Parallel()
	// Arrange
	params := mcp.ToolsCallParams{
		Arguments: map[string]any{"template_repo": 42},
	}

	// Act
	result := mcp.OptionalString(params, "template_repo")

	// Assert
	assert.That(t, "should return empty string for non-string type", result, "")
}

func Test_OptionalString_With_StringValue_Should_ReturnValue(t *testing.T) {
	t.Parallel()
	// Arrange
	params := mcp.ToolsCallParams{
		Arguments: map[string]any{"template_repo": "https://github.com/org/repo"},
	}

	// Act
	result := mcp.OptionalString(params, "template_repo")

	// Assert
	assert.That(t, "should return the string value", result, "https://github.com/org/repo")
}
