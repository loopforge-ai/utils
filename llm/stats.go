package llm

import "time"

// stats holds timing and token-estimate data for a single completion.
type stats struct {
	Duration     time.Duration
	OutputTokens int
}
