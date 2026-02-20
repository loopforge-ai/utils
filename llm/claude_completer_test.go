package llm_test

import (
	"context"
	"testing"

	"github.com/loopforge-ai/utils/assert"
	"github.com/loopforge-ai/utils/llm"
)

func Test_ClaudeComplete_With_CancelledContext_Should_ReturnError(t *testing.T) {
	t.Parallel()
	// Arrange
	c := llm.NewClaudeCompleter("sonnet")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Act
	result, err := c.Complete(ctx, "sys", "usr")

	// Assert
	assert.That(t, "error should not be nil", err != nil, true)
	assert.That(t, "result should be empty", result, "")
}

func Test_NewClaudeCompleter_With_DefaultConfig_Should_ReturnCompleter(t *testing.T) {
	t.Parallel()
	// Arrange & Act
	c := llm.NewClaudeCompleter("sonnet")

	// Assert
	assert.That(t, "completer should not be nil", c != nil, true)
}
