package llm_test

import (
	"context"
	"testing"

	"github.com/loopforge-ai/utils/assert"
	"github.com/loopforge-ai/utils/llm"
)

func Test_FakeCompleter_With_EmptyResponses_Should_ReturnFallback(t *testing.T) {
	t.Parallel()
	// Arrange
	fc := llm.NewFakeCompleter(map[string]string{}, "default response")

	// Act
	result, err := fc.Complete(context.Background(), "system", "any prompt")

	// Assert
	assert.That(t, "error should be nil", err, nil)
	assert.That(t, "should return fallback", result, "default response")
}

func Test_FakeCompleter_With_MatchingSubstring_Should_ReturnConfiguredResponse(t *testing.T) {
	t.Parallel()
	// Arrange
	responses := map[string]string{
		"define": `{"name": "test_skill"}`,
		"judge":  `{"score": 9}`,
	}
	fc := llm.NewFakeCompleter(responses, "fallback")

	// Act
	result, err := fc.Complete(context.Background(), "system", "please define this skill")

	// Assert
	assert.That(t, "error should be nil", err, nil)
	assert.That(t, "should match define substring", result, `{"name": "test_skill"}`)
}

func Test_FakeCompleter_With_NilResponses_Should_ReturnFallback(t *testing.T) {
	t.Parallel()
	// Arrange
	fc := llm.NewFakeCompleter(nil, "fallback")

	// Act
	result, err := fc.Complete(context.Background(), "system", "any prompt")

	// Assert
	assert.That(t, "error should be nil", err, nil)
	assert.That(t, "should return fallback", result, "fallback")
}

func Test_FakeCompleter_With_MatchingSystemPrompt_Should_ReturnConfiguredResponse(t *testing.T) {
	t.Parallel()
	// Arrange
	responses := map[string]string{
		"expert skill reviewer": `{"score": 9}`,
	}
	fc := llm.NewFakeCompleter(responses, "fallback")

	// Act
	result, err := fc.Complete(context.Background(), "You are an expert skill reviewer", "evaluate this")

	// Assert
	assert.That(t, "error should be nil", err, nil)
	assert.That(t, "should match system prompt substring", result, `{"score": 9}`)
}

func Test_FakeCompleter_With_NoMatch_Should_ReturnFallback(t *testing.T) {
	t.Parallel()
	// Arrange
	responses := map[string]string{
		"define": "defined",
		"judge":  "judged",
	}
	fc := llm.NewFakeCompleter(responses, "no match found")

	// Act
	result, err := fc.Complete(context.Background(), "system", "refine this skill")

	// Assert
	assert.That(t, "error should be nil", err, nil)
	assert.That(t, "should return fallback for non-matching prompt", result, "no match found")
}
