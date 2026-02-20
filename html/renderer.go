package html

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
)

// RendererConfig describes the template structure for a Renderer.
type RendererConfig struct {
	CommonFiles []string
	Pages       []string
	Partials    []string
}

// Renderer parses and executes HTML templates from an fs.FS.
type Renderer struct {
	templates map[string]*template.Template
}

// NewRenderer parses all page and partial templates from the given filesystem
// using the provided configuration.
func NewRenderer(fsys fs.FS, cfg RendererConfig) (*Renderer, error) {
	templates := make(map[string]*template.Template, len(cfg.Pages)+len(cfg.Partials))

	for _, page := range cfg.Pages {
		files := make([]string, len(cfg.CommonFiles)+1)
		copy(files, cfg.CommonFiles)
		files[len(cfg.CommonFiles)] = "templates/pages/" + page + ".html"
		tmpl, err := template.New("").ParseFS(fsys, files...)
		if err != nil {
			return nil, fmt.Errorf("parse template %s: %w", page, err)
		}
		templates[page] = tmpl
	}

	for _, partial := range cfg.Partials {
		tmpl, err := template.New("").ParseFS(fsys,
			"templates/partials/"+partial+".html",
		)
		if err != nil {
			return nil, fmt.Errorf("parse partial %s: %w", partial, err)
		}
		templates[partial] = tmpl
	}

	return &Renderer{templates: templates}, nil
}

// Render executes the named page template into w with the given data.
func (r *Renderer) Render(w io.Writer, page string, data any) error {
	tmpl, ok := r.templates[page]
	if !ok {
		return fmt.Errorf("unknown page %q", page)
	}
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		return fmt.Errorf("execute template %s: %w", page, err)
	}
	return nil
}

// RenderPartial executes a named partial template into w with the given data.
func (r *Renderer) RenderPartial(w io.Writer, partial string, data any) error {
	tmpl, ok := r.templates[partial]
	if !ok {
		return fmt.Errorf("unknown partial %q", partial)
	}
	if err := tmpl.ExecuteTemplate(w, partial, data); err != nil {
		return fmt.Errorf("execute partial %s: %w", partial, err)
	}
	return nil
}
