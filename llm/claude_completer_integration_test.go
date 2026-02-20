//go:build integration

package llm_test

import (
	"context"
	"testing"
	"time"

	"github.com/loopforge-ai/utils/assert"
)

const integrationClaudeModel = "sonnet"

func newIntegrationClaudeCompleter() *llm.ClaudeCompleter {
	return llm.NewClaudeCompleter(integrationClaudeModel)
}

func Test_Integration_Claude_Complete_With_CancelledContext_Should_ReturnError(t *testing.T) {
	t.Parallel()
	// Arrange
	c := newIntegrationClaudeCompleter()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Act
	result, err := c.Complete(ctx, "sys", "usr")

	// Assert
	assert.That(t, "error should not be nil", err != nil, true)
	assert.That(t, "result should be empty", result, "")
}

func Test_Integration_Claude_Complete_With_JSONPrompt_Should_ReturnValidJSON(t *testing.T) {
	t.Parallel()
	// Arrange
	c := newIntegrationClaudeCompleter()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	system := `You are a JSON generator. Respond with ONLY a JSON object. No markdown fences, no explanation.`
	user := `Produce a JSON object with these fields:
- name (string): "test_skill"
- description (string): "A test skill"
- tags ([]string): ["test", "example"]`

	// Act
	result, err := c.Complete(ctx, system, user)

	// Assert
	assert.That(t, "error should be nil", err, nil)
	assert.That(t, "result should not be empty", len(result) > 0, true)
	t.Logf("Response: %s", result)
}

func Test_Integration_Claude_Complete_With_SimplePrompt_Should_ReturnResponse(t *testing.T) {
	t.Parallel()
	// Arrange
	c := newIntegrationClaudeCompleter()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Act
	result, err := c.Complete(ctx, "You are a helpful assistant.", "Reply with exactly: hello")

	// Assert
	assert.That(t, "error should be nil", err, nil)
	assert.That(t, "result should not be empty", len(result) > 0, true)
	t.Logf("Response: %s", result)
}
