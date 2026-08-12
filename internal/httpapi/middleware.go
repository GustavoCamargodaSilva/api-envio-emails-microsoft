package httpapi

import (
	"crypto/subtle"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

func APIKeyMiddleware(apiKey string) func(http.Handler) http.Handler {
	expected := []byte(apiKey)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			provided := r.Header.Get("X-API-Key")
			if provided == "" {
				auth := r.Header.Get("Authorization")
				if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
					provided = strings.TrimSpace(auth[7:])
				}
			}
			// Comparação em tempo constante para não vazar o tamanho/prefixo
			// da chave via timing attack.
			if subtle.ConstantTimeCompare([]byte(provided), expected) != 1 {
				writeJSON(w, http.StatusUnauthorized, map[string]any{
					"error": "API key inválida ou ausente",
				})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// SecurityHeaders adiciona cabeçalhos de segurança básicos a todas as respostas.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		next.ServeHTTP(w, r)
	})
}

// rateLimiter é um limitador simples de janela deslizante por IP, em memória.
// Protege os endpoints de envio contra abuso (spam) caso a API key vaze.
type rateLimiter struct {
	mu       sync.Mutex
	hits     map[string][]time.Time
	max      int
	interval time.Duration
}

func newRateLimiter(max int, interval time.Duration) *rateLimiter {
	return &rateLimiter{
		hits:     make(map[string][]time.Time),
		max:      max,
		interval: interval,
	}
}

func (rl *rateLimiter) allow(key string) bool {
	now := time.Now()
	rl.mu.Lock()
	defer rl.mu.Unlock()

	janelaInicio := now.Add(-rl.interval)
	recentes := rl.hits[key][:0]
	for _, t := range rl.hits[key] {
		if t.After(janelaInicio) {
			recentes = append(recentes, t)
		}
	}
	if len(recentes) >= rl.max {
		rl.hits[key] = recentes
		return false
	}
	rl.hits[key] = append(recentes, now)
	return true
}

// RateLimitMiddleware limita requisições por IP de origem.
func RateLimitMiddleware(max int, interval time.Duration) func(http.Handler) http.Handler {
	rl := newRateLimiter(max, interval)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !rl.allow(clientIP(r)) {
				w.Header().Set("Retry-After", "60")
				writeJSON(w, http.StatusTooManyRequests, map[string]any{
					"error": "muitas requisições; tente novamente em instantes",
				})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func clientIP(r *http.Request) string {
	// Só confia em X-Forwarded-For quando atrás de proxy que sanitiza o header.
	if strings.EqualFold(os.Getenv("TRUST_PROXY_HEADERS"), "true") {
		if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
			if i := strings.IndexByte(fwd, ','); i > 0 {
				return strings.TrimSpace(fwd[:i])
			}
			return strings.TrimSpace(fwd)
		}
	}
	return r.RemoteAddr
}
