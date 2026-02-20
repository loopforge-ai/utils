package llm

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// maxTrackedCalls is the maximum number of call stats retained.
const maxTrackedCalls = 1000

// TrackedCompleter wraps a Completer and records timing and token-estimate
// statistics for each call. It is safe for concurrent use.
type TrackedCompleter struct {
	inner  Completer
	calls  []stats
	errors int
	mu     sync.Mutex
}

// compile-time interface check.
var _ Completer = (*TrackedCompleter)(nil)

// NewTrackedCompleter returns a TrackedCompleter that delegates to inner.
func NewTrackedCompleter(inner Completer) *TrackedCompleter {
	return &TrackedCompleter{inner: inner}
}

// Complete delegates to the inner Completer and records stats for the call.
func (t *TrackedCompleter) Complete(ctx context.Context, system, user string) (string, error) {
	start := time.Now()
	result, err := t.inner.Complete(ctx, system, user)
	elapsed := time.Since(start)

	t.mu.Lock()
	if err == nil {
		s := stats{
			Duration:     elapsed,
			OutputTokens: estimateTokens(result),
		}
		t.calls = append(t.calls, s)
		if len(t.calls) > maxTrackedCalls {
			t.calls = t.calls[len(t.calls)-maxTrackedCalls:]
		}
	} else {
		t.errors++
	}
	t.mu.Unlock()

	return result, err
}

// Errors returns the total number of failed completion calls.
func (t *TrackedCompleter) Errors() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.errors
}

// TokensPerSecond returns the aggregate tokens per second across all recorded calls.
func (t *TrackedCompleter) TokensPerSecond() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()

	var totalTokens int
	var totalDuration time.Duration
	for _, c := range t.calls {
		totalTokens += c.OutputTokens
		totalDuration += c.Duration
	}

	if secs := totalDuration.Seconds(); secs > 0 {
		return float64(totalTokens) / secs
	}
	return 0
}

// Summary returns a human-readable summary of all recorded calls.
// Format: "6 calls | ~4832 tokens | 38.2s | 126.5 tok/s".
func (t *TrackedCompleter) Summary() string {
	t.mu.Lock()
	calls := make([]stats, len(t.calls))
	copy(calls, t.calls)
	t.mu.Unlock()

	if len(calls) == 0 {
		return "0 calls"
	}

	var totalTokens int
	var totalDuration time.Duration
	for _, c := range calls {
		totalTokens += c.OutputTokens
		totalDuration += c.Duration
	}

	secs := totalDuration.Seconds()
	var tokPerSec float64
	if secs > 0 {
		tokPerSec = float64(totalTokens) / secs
	}

	return fmt.Sprintf("%d calls | ~%d tokens | %.1fs | %.1f tok/s",
		len(calls), totalTokens, secs, tokPerSec)
}

// estimateTokens estimates the number of tokens in s using a word-count heuristic.
func estimateTokens(s string) int {
	words := len(strings.Fields(s))
	return words * 4 / 3
}
