package llm

import (
	"context"
	"strings"
)

// Compile-time interface check.
var _ Completer = (*FakeCompleter)(nil)

// FakeCompleter is a test double for the Completer interface that returns
// configurable responses based on substring matching against both the system
// and user prompts. This allows routing responses by LLM operation (system prompt)
// or by input content (user prompt).
type FakeCompleter struct {
	responses map[string]string
	fallback  string
}

// NewFakeCompleter returns a FakeCompleter that matches prompts by substring.
// When either the system or user prompt contains a key from responses, the corresponding
// value is returned. First match wins when multiple substrings match. If no key matches,
// fallback is returned.
func NewFakeCompleter(responses map[string]string, fallback string) *FakeCompleter {
	return &FakeCompleter{
		fallback:  fallback,
		responses: responses,
	}
}

// Complete returns the first matching response or the fallback.
// It checks both system and user prompts for substring matches.
func (f *FakeCompleter) Complete(_ context.Context, system, user string) (string, error) {
	combined := system + "\n" + user
	for k, v := range f.responses {
		if strings.Contains(combined, k) {
			return v, nil
		}
	}
	return f.fallback, nil
}
