package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/mcpfleet/registry/internal/db"
)

type contextKey string

const tokenIDKey contextKey = "token_id"

// BearerAuth returns a chi-compatible middleware that validates
// Authorization: Bearer <token> against the token store.
// Requests to public paths (prefix match) bypass auth entirely.
func BearerAuth(store *db.Store, publicPrefixes ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Allow public paths through without auth
			for _, prefix := range publicPrefixes {
				if strings.HasPrefix(r.URL.Path, prefix) {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Extract Bearer token
			authorization := r.Header.Get("Authorization")
			if authorization == "" {
				unauthorized(w, "missing Authorization header")
				return
			}

			parts := strings.SplitN(authorization, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				unauthorized(w, "Authorization header must be: Bearer <token>")
				return
			}

			raw := strings.TrimSpace(parts[1])
			if raw == "" {
				unauthorized(w, "empty token")
				return
			}

			// Validate against store
			valid, err := store.ValidateToken(r.Context(), raw)
			if err != nil {
				http.Error(w, `{"title":"Internal Server Error","status":500}`, http.StatusInternalServerError)
				w.Header().Set("Content-Type", "application/problem+json")
				return
			}
			if !valid {
				unauthorized(w, "invalid or expired token")
				return
			}

			// Stash raw token in context for potential downstream use
			ctx := context.WithValue(r.Context(), tokenIDKey, raw)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func unauthorized(w http.ResponseWriter, detail string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.Header().Set("WWW-Authenticate", `Bearer realm="mcpfleet-registry"`)
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"title":"Unauthorized","status":401,"detail":"` + detail + `"}`))
}
