package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/CCLJ/playlisthome/internal/auth"
)

type contextKey string

const ClaimsKey contextKey = "claims"

// Authenticate reads the JWT from the Authorization header (Bearer) or
// the auth_token cookie, validates it, and injects the claims into ctx.
func Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := ""

		// Try Authorization header first
		if header := r.Header.Get("Authorization"); strings.HasPrefix(header, "Bearer ") {
			tokenStr = strings.TrimPrefix(header, "Bearer ")
		}

		// Fallback to cookie
		if tokenStr == "" {
			if cookie, err := r.Cookie("auth_token"); err == nil {
				tokenStr = cookie.Value
			}
		}

		if tokenStr == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		claims, err := auth.ParseToken(tokenStr)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), ClaimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
