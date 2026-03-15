package mcp

import (
	"net/http"
	"strings"
)

// BearerAuth returns middleware that validates the Authorization: Bearer <token> header.
// Requests without a valid token receive 401 Unauthorized.
func BearerAuth(token string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			http.Error(w, "Missing Authorization header. Include 'Authorization: Bearer <token>' in your request.", http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] != token {
			http.Error(w, "Invalid authorization. Check your bearer token and try again.", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
