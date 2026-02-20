package mcp

// capabilities represents MCP capabilities.
type capabilities struct {
	Tools *toolsCapability `json:"tools,omitempty"`
}

// ContentBlock represents a content block in tool results.
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// implementation represents server or client implementation info.
type implementation struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// initializeResult represents the result of the initialize method.
type initializeResult struct {
	ProtocolVersion string         `json:"protocolVersion"`
	Capabilities    capabilities   `json:"capabilities"`
	ServerInfo      implementation `json:"serverInfo"`
}

// ToolsCallParams represents the parameters for tools/call.
type ToolsCallParams struct {
	Arguments map[string]any `json:"arguments,omitempty"`
	Name      string         `json:"name"`
}

// ToolsCallResult represents the result of tools/call.
type ToolsCallResult struct {
	Content []ContentBlock `json:"content"`
	IsError bool           `json:"isError,omitempty"`
}

// toolsCapability represents tools capability.
type toolsCapability struct{}

// toolsListResult represents the result of tools/list.
type toolsListResult struct {
	Tools []ToolDefinition `json:"tools"`
}

// NewTextContent creates a text content block.
func NewTextContent(text string) ContentBlock {
	return ContentBlock{
		Type: "text",
		Text: text,
	}
}
