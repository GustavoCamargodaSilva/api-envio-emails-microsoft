package httpapi

import (
	"net/http"
)

func NewRouter(h *Handler, apiKey string) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", h.Health)

	// OAuth da caixa do site (operador). Login/callback abertos; status protegido.
	mux.HandleFunc("GET /v1/oauth/microsoft/login", h.OAuthLogin)
	mux.HandleFunc("GET /v1/oauth/microsoft/callback", h.OAuthCallback)

	protected := http.NewServeMux()
	protected.HandleFunc("GET /v1/oauth/microsoft/status", h.OAuthStatus)
	protected.HandleFunc("POST /v1/emails/send", h.SendGeneric)
	protected.HandleFunc("POST /v1/emails/send-by-tag", h.SendByTag)
	protected.HandleFunc("POST /v1/emails/invites", h.SendInvite)
	protected.HandleFunc("POST /v1/emails/confirmations", h.SendConfirmation)
	protected.HandleFunc("POST /v1/emails/campaigns", h.SendCampaign)

	mux.Handle("/v1/", APIKeyMiddleware(apiKey)(protected))

	return mux
}
