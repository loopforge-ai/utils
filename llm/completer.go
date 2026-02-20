package llm

import "context"

// Completer sends a chat completion request and returns the assistant content.
// Implementations must be safe for concurrent use.
type Completer interface {
	Complete(ctx context.Context, system, user string) (string, error)
}
