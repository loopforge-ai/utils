package llm

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

const maxStderrSnippet = 512

// compile-time interface check.
var _ Completer = (*ClaudeCompleter)(nil)

// ClaudeCompleter is an implementation of skill.Completer using the Claude CLI.
type ClaudeCompleter struct {
	model string
}

// NewClaudeCompleter creates a new completer that shells out to the claude CLI.
func NewClaudeCompleter(model string) *ClaudeCompleter {
	return &ClaudeCompleter{model: model}
}

// Complete sends system and user prompts to the Claude CLI and returns the response.
func (c *ClaudeCompleter) Complete(ctx context.Context, system, user string) (string, error) {
	args := []string{"--print", "--system-prompt", system}
	if c.model != "" {
		args = append([]string{"--model", c.model}, args...)
	}
	cmd := exec.CommandContext(ctx, "claude", args...) //nolint:gosec // claude CLI is a trusted tool
	cmd.Stdin = strings.NewReader(user)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		snippet := stderr.String()
		if len(snippet) > maxStderrSnippet {
			snippet = snippet[:maxStderrSnippet] + "... (truncated)"
		}
		return "", fmt.Errorf("claude cli: %w: %s", err, snippet)
	}

	result := strings.TrimSpace(stdout.String())
	if result == "" {
		return "", errors.New("claude completer: empty response")
	}

	return result, nil
}
