package middleware

import (
	"net/http"
	"strings"

	"nutricionist/auth"
)

// SecurityHeaders adds standard security headers to every response.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; script-src 'self' 'unsafe-inline' https://fonts.googleapis.com; "+
				"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com https://fonts.gstatic.com; "+
				"font-src 'self' https://fonts.gstatic.com; img-src 'self' data:; connect-src 'self' https://api.deepseek.com https://api.github.com")
		next.ServeHTTP(w, r)
	})
}

// MaxBodySize rejects requests larger than 1MB.
func MaxBodySize(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		next.ServeHTTP(w, r)
	})
}

// RequireAuth validates the JWT Bearer token and injects userID into the request context.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			http.Error(w, `{"error":"no autorizado"}`, http.StatusUnauthorized)
			return
		}
		token := strings.TrimPrefix(header, "Bearer ")
		userID, err := auth.ValidateToken(token)
		if err != nil {
			http.Error(w, `{"error":"token invalido"}`, http.StatusUnauthorized)
			return
		}
		ctx := auth.WithUserID(r.Context(), userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
