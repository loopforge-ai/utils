package html_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/loopforge-ai/utils/assert"
	"github.com/loopforge-ai/utils/html"
)

func newBrokenConfig() html.RendererConfig {
	return html.RendererConfig{
		CommonFiles: []string{
			"templates/layouts/base.html",
			"templates/partials/footer.html",
			"templates/partials/header.html",
			"templates/partials/item_grid.html",
		},
		Pages:    []string{"index"},
		Partials: []string{"item_grid"},
	}
}

func newBrokenFS() fstest.MapFS {
	return fstest.MapFS{
		"templates/layouts/base.html": &fstest.MapFile{
			Data: []byte(`{{define "base"}}{{call .Title}}{{end}}`),
		},
		"templates/partials/header.html": &fstest.MapFile{
			Data: []byte(`{{define "header"}}h{{end}}`),
		},
		"templates/partials/footer.html": &fstest.MapFile{
			Data: []byte(`{{define "footer"}}f{{end}}`),
		},
		"templates/partials/item_grid.html": &fstest.MapFile{
			Data: []byte(`{{define "item_grid"}}{{call .Title}}{{end}}`),
		},
		"templates/pages/index.html": &fstest.MapFile{
			Data: []byte(`{{define "title"}}t{{end}}{{define "content"}}c{{end}}`),
		},
	}
}

func Test_RenderPage_With_ValidRenderer_Should_WriteHTML(t *testing.T) {
	t.Parallel()

	// Arrange
	r, _ := html.NewRenderer(newTestFS(), newTestConfig())
	rec := httptest.NewRecorder()
	data := testPageData{
		Title:   "Home",
		Count:   3,
		Version: "1.0.0",
	}

	// Act
	html.RenderPage(rec, r, "index", data)

	// Assert
	assert.That(t, "status code", rec.Code, http.StatusOK)
	assert.That(t, "should contain DOCTYPE", strings.Contains(rec.Body.String(), "<!DOCTYPE html>"), true)
	assert.That(t, "should contain count", strings.Contains(rec.Body.String(), "count=3"), true)
}

func Test_RenderPage_With_BrokenRenderer_Should_Return500(t *testing.T) {
	t.Parallel()

	// Arrange
	r, _ := html.NewRenderer(newBrokenFS(), newBrokenConfig())
	rec := httptest.NewRecorder()
	data := testPageData{Title: "test"}

	// Act
	html.RenderPage(rec, r, "index", data)

	// Assert
	assert.That(t, "status code", rec.Code, http.StatusInternalServerError)
	assert.That(t, "body contains error", strings.Contains(rec.Body.String(), "Internal Server Error"), true)
}

func Test_RenderPartial_With_ValidRenderer_Should_WriteHTML(t *testing.T) {
	t.Parallel()

	// Arrange
	r, _ := html.NewRenderer(newTestFS(), newTestConfig())
	rec := httptest.NewRecorder()
	data := testPageData{
		Items: []testItem{
			{Name: "alpha"},
		},
	}

	// Act
	html.RenderPartial(rec, r, "item_grid", data)

	// Assert
	assert.That(t, "status code", rec.Code, http.StatusOK)
	assert.That(t, "should contain alpha", strings.Contains(rec.Body.String(), "alpha"), true)
	assert.That(t, "should not contain DOCTYPE", !strings.Contains(rec.Body.String(), "<!DOCTYPE"), true)
}

func Test_RenderPartial_With_BrokenRenderer_Should_Return500(t *testing.T) {
	t.Parallel()

	// Arrange
	r, _ := html.NewRenderer(newBrokenFS(), newBrokenConfig())
	rec := httptest.NewRecorder()
	data := testPageData{Title: "test"}

	// Act
	html.RenderPartial(rec, r, "item_grid", data)

	// Assert
	assert.That(t, "status code", rec.Code, http.StatusInternalServerError)
	assert.That(t, "body contains error", strings.Contains(rec.Body.String(), "Internal Server Error"), true)
}
