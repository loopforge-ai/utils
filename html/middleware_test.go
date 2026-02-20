package html_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/loopforge-ai/utils/assert"
	"github.com/loopforge-ai/utils/html"
)

func Test_CacheControl_With_Request_Should_SetHeader(t *testing.T) {
	t.Parallel()

	// Arrange
	inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := html.CacheControl(inner)
	req := httptest.NewRequest(http.MethodGet, "/static/css/style.css", nil)
	rec := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	assert.That(t, "status code", rec.Code, http.StatusOK)
	assert.That(t, "cache-control header", strings.Contains(rec.Header().Get("Cache-Control"), "public"), true)
}

func Test_ContentType_With_Request_Should_SetHeader(t *testing.T) {
	t.Parallel()

	// Arrange
	inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := html.ContentType(inner)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	assert.That(t, "status code", rec.Code, http.StatusOK)
	assert.That(t, "content-type header", rec.Header().Get("Content-Type"), "text/html; charset=utf-8")
}

func Test_Log_With_Request_Should_CallNext(t *testing.T) {
	t.Parallel()

	// Arrange
	called := false
	inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})
	handler := html.Log(inner)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	assert.That(t, "inner handler called", called, true)
	assert.That(t, "status code", rec.Code, http.StatusOK)
}

func Test_Recover_With_NoPanic_Should_PassThrough(t *testing.T) {
	t.Parallel()

	// Arrange
	inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := html.Recover(inner)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	assert.That(t, "status code", rec.Code, http.StatusOK)
}

func Test_Recover_With_Panic_Should_Return500(t *testing.T) {
	t.Parallel()

	// Arrange
	inner := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		panic("test panic")
	})
	handler := html.Recover(inner)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	assert.That(t, "status code", rec.Code, http.StatusInternalServerError)
	assert.That(t, "body contains error", strings.Contains(rec.Body.String(), "Internal Server Error"), true)
}

func Test_SecurityHeaders_With_Request_Should_SetAllHeaders(t *testing.T) {
	t.Parallel()

	// Arrange
	inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := html.SecurityHeaders(inner)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	assert.That(t, "status code", rec.Code, http.StatusOK)
	assert.That(t, "csp header set", strings.Contains(rec.Header().Get("Content-Security-Policy"), "default-src 'self'"), true)
	assert.That(t, "permissions-policy header set", strings.Contains(rec.Header().Get("Permissions-Policy"), "camera=()"), true)
	assert.That(t, "referrer-policy header", rec.Header().Get("Referrer-Policy"), "strict-origin-when-cross-origin")
	assert.That(t, "strict-transport-security header set", strings.Contains(rec.Header().Get("Strict-Transport-Security"), "max-age="), true)
	assert.That(t, "x-content-type-options header", rec.Header().Get("X-Content-Type-Options"), "nosniff")
	assert.That(t, "x-frame-options header", rec.Header().Get("X-Frame-Options"), "DENY")
}
