package httpapi

import (
	"net/http"
	"strings"
)

func APIKeyMiddleware(apiKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			provided := r.Header.Get("X-API-Key")
			if provided == "" {
				auth := r.Header.Get("Authorization")
				if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
					provided = strings.TrimSpace(auth[7:])
				}
			}
			if provided == "" || provided != apiKey {
				writeJSON(w, http.StatusUnauthorized, map[string]any{
					"error": "API key inválida ou ausente",
				})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
