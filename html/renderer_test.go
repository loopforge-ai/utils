package html_test

import (
	"bytes"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/loopforge-ai/utils/assert"
	"github.com/loopforge-ai/utils/html"
)

type testPageData struct {
	Title   string
	Version string
	Items   []testItem
	Count   int
}

type testItem struct {
	Name string
}

func newTestConfig() html.RendererConfig {
	return html.RendererConfig{
		CommonFiles: []string{
			"templates/layouts/base.html",
			"templates/partials/footer.html",
			"templates/partials/header.html",
			"templates/partials/item_grid.html",
		},
		Pages:    []string{"index", "detail"},
		Partials: []string{"item_grid"},
	}
}

func newTestFS() fstest.MapFS {
	return fstest.MapFS{
		"templates/layouts/base.html": &fstest.MapFile{
			Data: []byte(`{{define "base"}}<!DOCTYPE html><html><head><title>{{template "title" .}}</title></head><body>{{template "header" .}}{{template "content" .}}{{template "footer" .}}</body></html>{{end}}`),
		},
		"templates/partials/header.html": &fstest.MapFile{
			Data: []byte(`{{define "header"}}<header>nav</header>{{end}}`),
		},
		"templates/partials/footer.html": &fstest.MapFile{
			Data: []byte(`{{define "footer"}}<footer>{{.Version}}</footer>{{end}}`),
		},
		"templates/partials/item_grid.html": &fstest.MapFile{
			Data: []byte(`{{define "item_grid"}}{{range .Items}}<div>{{.Name}}</div>{{end}}{{end}}`),
		},
		"templates/pages/index.html": &fstest.MapFile{
			Data: []byte(`{{define "title"}}Home{{end}}{{define "content"}}<h1>count={{.Count}}</h1>{{end}}`),
		},
		"templates/pages/detail.html": &fstest.MapFile{
			Data: []byte(`{{define "title"}}Detail{{end}}{{define "content"}}<h1>{{.Title}}</h1>{{end}}`),
		},
	}
}

func Test_NewRenderer_With_MissingLayout_Should_ReturnError(t *testing.T) {
	t.Parallel()

	// Arrange
	fsys := fstest.MapFS{
		"templates/partials/header.html":    &fstest.MapFile{Data: []byte(`{{define "header"}}h{{end}}`)},
		"templates/partials/footer.html":    &fstest.MapFile{Data: []byte(`{{define "footer"}}f{{end}}`)},
		"templates/partials/item_grid.html": &fstest.MapFile{Data: []byte(`{{define "item_grid"}}g{{end}}`)},
		"templates/pages/index.html":        &fstest.MapFile{Data: []byte(`{{define "title"}}t{{end}}{{define "content"}}c{{end}}`)},
		"templates/pages/detail.html":       &fstest.MapFile{Data: []byte(`{{define "title"}}t{{end}}{{define "content"}}c{{end}}`)},
	}
	cfg := newTestConfig()

	// Act
	r, err := html.NewRenderer(fsys, cfg)

	// Assert
	assert.That(t, "renderer should be nil", r == nil, true)
	assert.That(t, "error should not be nil", err != nil, true)
}

func Test_NewRenderer_With_ValidFS_Should_ParseTemplates(t *testing.T) {
	t.Parallel()

	// Arrange
	fsys := newTestFS()
	cfg := newTestConfig()

	// Act
	r, err := html.NewRenderer(fsys, cfg)

	// Assert
	assert.That(t, "error should be nil", err, nil)
	assert.That(t, "renderer should not be nil", r != nil, true)
}

func Test_Render_With_IndexPage_Should_WriteHTML(t *testing.T) {
	t.Parallel()

	// Arrange
	r, _ := html.NewRenderer(newTestFS(), newTestConfig())
	var buf bytes.Buffer
	data := testPageData{
		Title:   "Home",
		Count:   5,
		Version: "1.0.0",
	}

	// Act
	err := r.Render(&buf, "index", data)

	// Assert
	assert.That(t, "error should be nil", err, nil)
	assert.That(t, "should contain DOCTYPE", strings.Contains(buf.String(), "<!DOCTYPE html>"), true)
	assert.That(t, "should contain title", strings.Contains(buf.String(), "<title>Home</title>"), true)
	assert.That(t, "should contain count", strings.Contains(buf.String(), "count=5"), true)
	assert.That(t, "should contain version", strings.Contains(buf.String(), "1.0.0"), true)
}

func Test_Render_With_DetailPage_Should_WriteTitle(t *testing.T) {
	t.Parallel()

	// Arrange
	r, _ := html.NewRenderer(newTestFS(), newTestConfig())
	var buf bytes.Buffer
	data := testPageData{
		Title: "my_item",
	}

	// Act
	err := r.Render(&buf, "detail", data)

	// Assert
	assert.That(t, "error should be nil", err, nil)
	assert.That(t, "should contain title", strings.Contains(buf.String(), "my_item"), true)
}

func Test_Render_With_UnknownPage_Should_ReturnError(t *testing.T) {
	t.Parallel()

	// Arrange
	r, _ := html.NewRenderer(newTestFS(), newTestConfig())
	var buf bytes.Buffer

	// Act
	err := r.Render(&buf, "nonexistent", testPageData{})

	// Assert
	assert.That(t, "error should not be nil", err != nil, true)
	assert.That(t, "error should mention page", strings.Contains(err.Error(), "nonexistent"), true)
}

func Test_RenderPartial_With_ItemGrid_Should_WriteItems(t *testing.T) {
	t.Parallel()

	// Arrange
	r, _ := html.NewRenderer(newTestFS(), newTestConfig())
	var buf bytes.Buffer
	data := testPageData{
		Items: []testItem{
			{Name: "alpha"},
			{Name: "beta"},
		},
	}

	// Act
	err := r.RenderPartial(&buf, "item_grid", data)

	// Assert
	assert.That(t, "error should be nil", err, nil)
	assert.That(t, "should contain alpha", strings.Contains(buf.String(), "alpha"), true)
	assert.That(t, "should contain beta", strings.Contains(buf.String(), "beta"), true)
	assert.That(t, "should not contain DOCTYPE", !strings.Contains(buf.String(), "<!DOCTYPE"), true)
}

func Test_RenderPartial_With_UnknownPartial_Should_ReturnError(t *testing.T) {
	t.Parallel()

	// Arrange
	r, _ := html.NewRenderer(newTestFS(), newTestConfig())
	var buf bytes.Buffer

	// Act
	err := r.RenderPartial(&buf, "nonexistent", testPageData{})

	// Assert
	assert.That(t, "error should not be nil", err != nil, true)
	assert.That(t, "error should mention partial", strings.Contains(err.Error(), "nonexistent"), true)
}
