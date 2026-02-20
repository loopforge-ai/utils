package mcp

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ErrResult builds an error ToolsCallResult with a formatted message.
func ErrResult(msg string, args ...any) ToolsCallResult {
	return ToolsCallResult{
		Content: []ContentBlock{NewTextContent(fmt.Sprintf(msg, args...))},
		IsError: true,
	}
}

// JSONResult marshals v to indented JSON and wraps it in a ToolsCallResult.
// On marshal error it returns an error result rather than a Go error.
func JSONResult(v any) ToolsCallResult {
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return ErrResult("marshal result: %v", err)
	}
	return ToolsCallResult{
		Content: []ContentBlock{NewTextContent(string(out))},
	}
}

// OptionalJSONMap extracts an optional JSON string argument and unmarshals it
// into a map[string]string. Returns nil map if the key is absent or empty.
func OptionalJSONMap(params ToolsCallParams, key string) (map[string]string, *ToolsCallResult) {
	if params.Arguments == nil {
		return nil, nil
	}
	v, exists := params.Arguments[key]
	if !exists {
		return nil, nil
	}
	raw, ok := v.(string)
	if !ok {
		r := ErrResult("%s must be a string", key)
		return nil, &r
	}
	if raw == "" {
		return nil, nil
	}
	var m map[string]string
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		r := ErrResult("invalid %s JSON: %v", key, err)
		return nil, &r
	}
	return m, nil
}

// RequireString extracts a required string argument from the tool call params.
// If the argument is missing or empty, it returns an error result.
func RequireString(params ToolsCallParams, key string) (string, *ToolsCallResult) {
	raw, exists := params.Arguments[key]
	if !exists {
		r := ErrResult("missing required parameter %q", key)
		return "", &r
	}
	v, ok := raw.(string)
	if !ok {
		r := ErrResult("parameter %q must be a string", key)
		return "", &r
	}
	if strings.TrimSpace(v) == "" {
		r := ErrResult("parameter %q must not be empty", key)
		return "", &r
	}
	return v, nil
}

// RequireStrings extracts multiple required string arguments from the tool call params.
// Returns a map of key→value on success, or an error result for the first missing key.
func RequireStrings(params ToolsCallParams, keys ...string) (map[string]string, *ToolsCallResult) {
	result := make(map[string]string, len(keys))
	for _, key := range keys {
		v, errRes := RequireString(params, key)
		if errRes != nil {
			return nil, errRes
		}
		result[key] = v
	}
	return result, nil
}

// SuccessWithSummary builds a success result with an optional appended summary block.
func SuccessWithSummary(message, summary string) ToolsCallResult {
	content := []ContentBlock{NewTextContent(message)}
	if summary != "" {
		content = append(content, NewTextContent(summary))
	}
	return ToolsCallResult{Content: content}
}

// TextResult builds a success ToolsCallResult with a single text content block.
func TextResult(text string) ToolsCallResult {
	return ToolsCallResult{
		Content: []ContentBlock{NewTextContent(text)},
	}
}
