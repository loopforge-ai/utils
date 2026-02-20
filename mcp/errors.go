package mcp

import "errors"

// protocolVersion is the MCP protocol version advertised during initialization.
const protocolVersion = "2024-11-05"

// JSON-RPC 2.0 standard error codes.
const (
	ErrorCodeInternal       = -32603
	ErrorCodeInvalidParams  = -32602
	ErrorCodeInvalidRequest = -32600
	ErrorCodeMethodNotFound = -32601
	ErrorCodeParse          = -32700
)

// MCP-specific errors.
var (
	errNotInitialized = errors.New("server not initialized")
	errToolNotFound   = errors.New("tool not found")
)
