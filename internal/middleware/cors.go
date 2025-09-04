package middleware

import (
	"net/http"
	"strings"
)

// CORS returns middleware that sets common CORS headers.
// If allowedOrigins is empty, no Access-Control-Allow-Origin header is set.
// Use ["*"] to allow any origin.
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" {
				if len(allowedOrigins) == 0 || contains(allowedOrigins, "*") || contains(allowedOrigins, origin) {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Vary", "Origin")
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
					w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Requested-With")
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}
			}

			// Preflight
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if strings.EqualFold(v, s) {
			return true
		}
	}
	return false
}