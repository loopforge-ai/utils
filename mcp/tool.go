package mcp

import "context"

// InputSchema represents the JSON Schema for tool input.
type InputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties,omitempty"`
	Required   []string            `json:"required,omitempty"`
}

// Property represents a JSON Schema property.
type Property struct {
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

// Tool represents a registered tool with its definition and handler.
type Tool struct {
	Handler    ToolHandler
	Definition ToolDefinition
}

// ToolDefinition describes a tool's metadata for tools/list.
type ToolDefinition struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	InputSchema InputSchema `json:"inputSchema"`
}

// ToolHandler is a function that handles tool calls.
type ToolHandler func(ctx context.Context, in ToolsCallParams) (out ToolsCallResult, err error)

// NewObjectSchema creates an object input schema.
func NewObjectSchema(properties map[string]Property, required []string) InputSchema {
	return InputSchema{
		Type:       "object",
		Properties: properties,
		Required:   required,
	}
}

// NewStringProperty creates a string property.
func NewStringProperty(description string) Property {
	return Property{
		Type:        "string",
		Description: description,
	}
}

// NewTool creates a new tool with the given definition and handler.
// Panics if name is empty or handler is nil.
func NewTool(name, description string, schema InputSchema, handler ToolHandler) Tool {
	if name == "" {
		panic("mcp: tool name must not be empty")
	}
	if handler == nil {
		panic("mcp: tool handler must not be nil")
	}
	return Tool{
		Definition: ToolDefinition{
			Name:        name,
			Description: description,
			InputSchema: schema,
		},
		Handler: handler,
	}
}
