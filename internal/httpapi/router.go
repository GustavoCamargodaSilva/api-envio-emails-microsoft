package httpapi

import (
	"net/http"
	"time"
)

func NewRouter(h *Handler, apiKey string) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", h.Health)

	// OAuth da caixa do site (operador). Mantém-se sem API key para o redirect Entra,
	// mas com rate limit agressivo. Em produção, exponha só via VPN/localhost/túnel.
	oauthLimited := RateLimitMiddleware(10, time.Minute)
	mux.Handle("GET /v1/oauth/ms/login", oauthLimited(http.HandlerFunc(h.OAuthLogin)))
	mux.Handle("GET /v1/oauth/ms/callback", oauthLimited(http.HandlerFunc(h.OAuthCallback)))

	protected := http.NewServeMux()
	protected.HandleFunc("GET /v1/oauth/ms/status", h.OAuthStatus)
	protected.HandleFunc("POST /v1/emails/send", h.SendGeneric)
	protected.HandleFunc("POST /v1/emails/send-by-tag", h.SendByTag)
	protected.HandleFunc("POST /v1/emails/invites", h.SendInvite)
	protected.HandleFunc("POST /v1/emails/confirmations", h.SendConfirmation)
	protected.HandleFunc("POST /v1/emails/campaigns", h.SendCampaign)

	// Rate limiting por IP como defesa em profundidade contra abuso caso a
	// API key seja comprometida (60 req/min por origem).
	limited := RateLimitMiddleware(60, time.Minute)(protected)
	mux.Handle("/v1/", APIKeyMiddleware(apiKey)(limited))

	return mux
}
