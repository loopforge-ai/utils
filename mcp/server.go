package mcp

import (
	"bufio"
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"slices"
	"sync"
)

const maxRequestBytes = 10 << 20 // 10 MB

// Server is an MCP server that handles JSON-RPC requests over STDIO.
type Server struct {
	tools       map[string]Tool
	scanner     *bufio.Scanner
	writer      *bufio.Writer
	name        string
	version     string
	mu          sync.RWMutex
	initialized bool
}

// NewServer creates a new MCP server with the given name and version.
func NewServer(name, version string) *Server {
	return newServer(name, version, os.Stdin, os.Stdout)
}

// NewServerWithIO creates a new MCP server with custom IO (for testing).
func NewServerWithIO(name, version string, reader io.Reader, writer io.Writer) *Server {
	return newServer(name, version, reader, writer)
}

// Name returns the server name.
func (s *Server) Name() string {
	return s.name
}

// HandleJSONRPC processes a single JSON-RPC request and returns the response.
// This is the transport-agnostic entry point used by both stdio and HTTP transports.
func (s *Server) HandleJSONRPC(ctx context.Context, req Request) Response {
	return s.routeRequest(ctx, req)
}

// RegisterTool registers a tool with the server.
func (s *Server) RegisterTool(tool Tool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tools[tool.Definition.Name] = tool
}

// Serve starts the server and processes requests until context is canceled.
func (s *Server) Serve(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("serve: %w", ctx.Err())
		default:
			if err := s.handleRequest(ctx); err != nil {
				if errors.Is(err, io.EOF) {
					return nil
				}
				if ctx.Err() != nil {
					return fmt.Errorf("serve: %w", ctx.Err())
				}
				log.Printf("handle request: %v", err)
			}
		}
	}
}

// Tools returns all registered tools.
func (s *Server) Tools() []Tool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tools := make([]Tool, 0, len(s.tools))
	for _, t := range s.tools {
		tools = append(tools, t)
	}
	slices.SortFunc(tools, func(a, b Tool) int {
		return cmp.Compare(a.Definition.Name, b.Definition.Name)
	})
	return tools
}

// Version returns the server version.
func (s *Server) Version() string {
	return s.version
}

// handleInitialize handles the initialize request.
func (s *Server) handleInitialize(req Request) Response {
	s.mu.Lock()
	s.initialized = true
	s.mu.Unlock()

	result := initializeResult{
		ProtocolVersion: protocolVersion,
		Capabilities: capabilities{
			Tools: &toolsCapability{},
		},
		ServerInfo: implementation{
			Name:    s.name,
			Version: s.version,
		},
	}

	return NewResponse(req.ID, result)
}

// handleInitialized handles the initialized notification.
func (s *Server) handleInitialized() Response {
	return Response{}
}

// handleRequest reads a single request from the scanner.
func (s *Server) handleRequest(ctx context.Context) error {
	if !s.scanner.Scan() {
		if err := s.scanner.Err(); err != nil {
			_ = s.writeResponse(NewErrorResponse(nil, ErrorCodeParse, "request too large"))
			return io.EOF
		}
		return io.EOF
	}

	line := s.scanner.Bytes()

	var req Request
	if err := json.Unmarshal(line, &req); err != nil {
		if wErr := s.writeResponse(NewErrorResponse(nil, ErrorCodeParse, "Parse error")); wErr != nil {
			return fmt.Errorf("write parse error response: %w", wErr)
		}
		return nil
	}

	resp := s.routeRequest(ctx, req)
	if err := s.writeResponse(resp); err != nil {
		return fmt.Errorf("write response: %w", err)
	}
	return nil
}

// handleToolsCall handles the tools/call request.
func (s *Server) handleToolsCall(ctx context.Context, req Request) Response {
	var params ToolsCallParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(req.ID, ErrorCodeInvalidParams, "Invalid params")
	}

	s.mu.RLock()
	initialized := s.initialized
	tool, ok := s.tools[params.Name]
	s.mu.RUnlock()

	if !initialized {
		return NewErrorResponse(req.ID, ErrorCodeInternal, errNotInitialized.Error())
	}

	if !ok {
		return NewErrorResponse(req.ID, ErrorCodeInvalidParams, errToolNotFound.Error())
	}

	result, err := tool.Handler(ctx, params)
	if err != nil {
		return NewResponse(req.ID, ToolsCallResult{
			Content: []ContentBlock{NewTextContent(err.Error())},
			IsError: true,
		})
	}

	return NewResponse(req.ID, result)
}

// handleToolsList handles the tools/list request.
func (s *Server) handleToolsList(req Request) Response {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.initialized {
		return NewErrorResponse(req.ID, ErrorCodeInternal, errNotInitialized.Error())
	}

	return NewResponse(req.ID, toolsListResult{Tools: s.sortedToolDefinitions()})
}

func newServer(name, version string, reader io.Reader, writer io.Writer) *Server {
	sc := bufio.NewScanner(reader)
	sc.Buffer(make([]byte, 0, 4096), maxRequestBytes)
	return &Server{
		name:    name,
		version: version,
		tools:   make(map[string]Tool),
		scanner: sc,
		writer:  bufio.NewWriter(writer),
	}
}

// routeRequest routes the request to the appropriate handler.
func (s *Server) routeRequest(ctx context.Context, req Request) (resp Response) { //nolint:nonamedreturns // named return needed for defer+recover
	defer func() {
		if r := recover(); r != nil {
			resp = NewErrorResponse(req.ID, ErrorCodeInternal, fmt.Sprintf("internal error: %v", r))
		}
	}()

	// Validate JSON-RPC version for all methods except notifications.
	if req.Method != "initialized" && req.JSONRPC != "2.0" {
		return NewErrorResponse(req.ID, ErrorCodeInvalidRequest, "invalid jsonrpc version")
	}

	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "initialized":
		return s.handleInitialized()
	case "tools/call":
		return s.handleToolsCall(ctx, req)
	case "tools/list":
		return s.handleToolsList(req)
	default:
		return NewErrorResponse(req.ID, ErrorCodeMethodNotFound, "Method not found")
	}
}

// sortedToolDefinitions returns all tool definitions sorted by name.
// Caller must hold s.mu (at least RLock).
func (s *Server) sortedToolDefinitions() []ToolDefinition {
	defs := make([]ToolDefinition, 0, len(s.tools))
	for _, tool := range s.tools {
		defs = append(defs, tool.Definition)
	}
	slices.SortFunc(defs, func(a, b ToolDefinition) int {
		return cmp.Compare(a.Name, b.Name)
	})
	return defs
}

// writeResponse writes a response to the output writer.
func (s *Server) writeResponse(resp Response) error {
	if resp.JSONRPC == "" {
		return nil
	}

	data, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("marshal response: %w", err)
	}

	if _, err = s.writer.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("write response: %w", err)
	}
	return s.writer.Flush()
}
