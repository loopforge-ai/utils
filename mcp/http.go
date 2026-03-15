package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Option configures the HTTP MCP server.
type Option func(*httpConfig)

type httpConfig struct {
	authToken string
}

// WithBearerAuth requires a valid Authorization: Bearer <token> header on all requests.
func WithBearerAuth(token string) Option {
	return func(cfg *httpConfig) {
		cfg.authToken = token
	}
}

// ServeHTTP starts an HTTP server that serves MCP tools over JSON-RPC.
// POST requests to the root path are treated as JSON-RPC requests.
// The server auto-initializes (no separate initialize handshake required).
func ServeHTTP(ctx context.Context, server *Server, addr string, opts ...Option) error {
	cfg := &httpConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	// Auto-initialize the server for HTTP transport (no handshake needed).
	server.mu.Lock()
	server.initialized = true
	server.mu.Unlock()

	handler := newHTTPHandler(server)

	var h http.Handler = handler
	if cfg.authToken != "" {
		h = BearerAuth(cfg.authToken, h)
	}

	httpServer := &http.Server{
		Addr:              addr,
		Handler:           h,
		ReadHeaderTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("listen: %w", err)
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		if err := httpServer.Close(); err != nil {
			return fmt.Errorf("close: %w", err)
		}
		return nil
	case err := <-errCh:
		return err
	}
}

// httpHandler handles JSON-RPC requests over HTTP.
type httpHandler struct {
	server *Server
}

func newHTTPHandler(server *Server) *httpHandler {
	return &httpHandler{server: server}
}

func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed. Use POST with a JSON-RPC 2.0 request body.", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, maxRequestBytes))
	if err != nil {
		writeJSONError(w, nil, ErrorCodeParse, "Failed to read request body.")
		return
	}
	defer func() { _ = r.Body.Close() }()

	var req Request
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSONError(w, nil, ErrorCodeParse, "Invalid JSON. Send a valid JSON-RPC 2.0 request.")
		return
	}

	resp := h.server.HandleJSONRPC(r.Context(), req)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {		http.Error(w, "Internal error during response encoding.", http.StatusInternalServerError)
	}
}

func writeJSONError(w http.ResponseWriter, id json.RawMessage, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // JSON-RPC errors use 200 with error in body
	resp := NewErrorResponse(id, code, message)
	if err := json.NewEncoder(w).Encode(resp); err != nil {		http.Error(w, "Internal error during response encoding.", http.StatusInternalServerError)
	}
}
