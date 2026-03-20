package mcp

// OptionalString returns the string value for key from params,
// or an empty string if the key is absent or not a string.
func OptionalString(params ToolsCallParams, key string) string {
	if params.Arguments == nil {
		return ""
	}
	v, exists := params.Arguments[key]
	if !exists {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}
