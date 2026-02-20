package html

import (
	"bytes"
	"log/slog"
	"net/http"
)

// RenderPage renders a full page template into the response writer.
// On render error it writes a 500 response.
func RenderPage(w http.ResponseWriter, renderer *Renderer, page string, data any) {
	var buf bytes.Buffer
	if err := renderer.Render(&buf, page, data); err != nil {
		slog.Error("render "+page, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if _, err := buf.WriteTo(w); err != nil {
		slog.Error("write "+page+" response", "error", err)
	}
}

// RenderPartial renders a partial template into the response writer.
// On render error it writes a 500 response.
func RenderPartial(w http.ResponseWriter, renderer *Renderer, partial string, data any) {
	var buf bytes.Buffer
	if err := renderer.RenderPartial(&buf, partial, data); err != nil {
		slog.Error("render "+partial+" partial", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if _, err := buf.WriteTo(w); err != nil {
		slog.Error("write "+partial+" response", "error", err)
	}
}
