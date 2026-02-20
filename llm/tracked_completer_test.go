package llm_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/loopforge-ai/utils/assert"
	"github.com/loopforge-ai/utils/llm"
)

// fakeCompleter is a test double that returns a fixed response.
type fakeCompleter struct {
	err      error
	response string
}

func (f *fakeCompleter) Complete(_ context.Context, _, _ string) (string, error) {
	time.Sleep(time.Millisecond)
	return f.response, f.err
}

func Test_EstimateTokens_With_KnownInput_Should_EstimateCorrectly(t *testing.T) {
	t.Parallel()
	// Arrange
	// "hello world foo" = 3 words → 3 * 4 / 3 = 4 tokens
	inner := &fakeCompleter{response: "hello world foo"}
	tc := llm.NewTrackedCompleter(inner)

	// Act
	_, _ = tc.Complete(context.Background(), "sys", "usr")

	// Assert
	summary := tc.Summary()
	assert.That(t, "summary should contain ~4 tokens", strings.Contains(summary, "~4 tokens"), true)
}

func Test_NewTrackedCompleter_With_DefaultConfig_Should_ReturnCompleter(t *testing.T) {
	t.Parallel()
	// Arrange & Act
	tc := llm.NewTrackedCompleter(&fakeCompleter{})

	// Assert
	assert.That(t, "completer should not be nil", tc != nil, true)
}

func Test_TrackedCompleter_With_ErrorResponse_Should_NotRecordStats(t *testing.T) {
	t.Parallel()
	// Arrange
	inner := &fakeCompleter{err: errors.New("fail")}
	tc := llm.NewTrackedCompleter(inner)

	// Act
	_, _ = tc.Complete(context.Background(), "sys", "usr")

	// Assert
	assert.That(t, "summary should show 0 calls", tc.Summary(), "0 calls")
}

func Test_TrackedCompleter_With_ErrorResponse_Should_TrackErrorCount(t *testing.T) {
	t.Parallel()
	// Arrange
	inner := &fakeCompleter{err: errors.New("fail")}
	tc := llm.NewTrackedCompleter(inner)

	// Act
	_, _ = tc.Complete(context.Background(), "sys", "usr")
	_, _ = tc.Complete(context.Background(), "sys", "usr")

	// Assert
	assert.That(t, "errors should be 2", tc.Errors(), 2)
}

func Test_TrackedCompleter_With_MultipleCalls_Should_AccumulateStats(t *testing.T) {
	t.Parallel()
	// Arrange
	inner := &fakeCompleter{response: "one two three"}
	tc := llm.NewTrackedCompleter(inner)

	// Act
	_, _ = tc.Complete(context.Background(), "sys", "usr")
	_, _ = tc.Complete(context.Background(), "sys", "usr")

	// Assert
	summary := tc.Summary()
	assert.That(t, "summary should show 2 calls", strings.Contains(summary, "2 calls"), true)
}

func Test_TrackedCompleter_With_NoCalls_Should_ReturnZeroCalls(t *testing.T) {
	t.Parallel()
	// Arrange
	tc := llm.NewTrackedCompleter(&fakeCompleter{})

	// Act
	summary := tc.Summary()

	// Assert
	assert.That(t, "summary should be 0 calls", summary, "0 calls")
}

func Test_TrackedCompleter_With_SuccessfulCall_Should_DelegateToInner(t *testing.T) {
	t.Parallel()
	// Arrange
	inner := &fakeCompleter{response: "hello"}
	tc := llm.NewTrackedCompleter(inner)

	// Act
	result, err := tc.Complete(context.Background(), "sys", "usr")

	// Assert
	assert.That(t, "error should be nil", err == nil, true)
	assert.That(t, "result should match inner", result, "hello")
}

func Test_TrackedCompleter_With_NoCalls_Should_ReturnZeroTokensPerSecond(t *testing.T) {
	t.Parallel()
	// Arrange
	tc := llm.NewTrackedCompleter(&fakeCompleter{})

	// Act
	tps := tc.TokensPerSecond()

	// Assert
	assert.That(t, "tok/s should be zero", tps, float64(0))
}

func Test_TrackedCompleter_With_SuccessfulCall_Should_ReturnPositiveTokensPerSecond(t *testing.T) {
	t.Parallel()
	// Arrange
	inner := &fakeCompleter{response: "hello world foo"}
	tc := llm.NewTrackedCompleter(inner)

	// Act
	_, _ = tc.Complete(context.Background(), "sys", "usr")
	tps := tc.TokensPerSecond()

	// Assert
	assert.That(t, "tok/s should be positive", tps > 0, true)
}

func Test_TrackedCompleter_With_SuccessfulCall_Should_IncludeTokPerSec(t *testing.T) {
	t.Parallel()
	// Arrange
	inner := &fakeCompleter{response: "hello world"}
	tc := llm.NewTrackedCompleter(inner)

	// Act
	_, _ = tc.Complete(context.Background(), "sys", "usr")

	// Assert
	summary := tc.Summary()
	assert.That(t, "summary should contain tok/s", strings.Contains(summary, "tok/s"), true)
	assert.That(t, "summary should contain 1 calls", strings.Contains(summary, "1 calls"), true)
}

func Test_TrackedCompleter_With_ConcurrentCalls_Should_NotRace(t *testing.T) {
	t.Parallel()
	// Arrange
	inner := &fakeCompleter{response: "hello world foo"}
	tc := llm.NewTrackedCompleter(inner)
	const goroutines = 10

	// Act
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for range goroutines {
		go func() {
			defer wg.Done()
			_, _ = tc.Complete(context.Background(), "sys", "usr")
		}()
	}
	wg.Wait()

	// Assert
	summary := tc.Summary()
	assert.That(t, "summary should show 10 calls", strings.Contains(summary, "10 calls"), true)
}
