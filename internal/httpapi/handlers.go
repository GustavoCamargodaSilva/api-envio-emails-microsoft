package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/app-financeiro/api-envio-emails/internal/auth/msgraph"
	"github.com/app-financeiro/api-envio-emails/internal/mail"
	"github.com/app-financeiro/api-envio-emails/internal/store"
	"github.com/app-financeiro/api-envio-emails/internal/templates"
)

type Handler struct {
	oauth    *msgraph.OAuthClient
	mailer   *mail.Client
	tokens   *store.TokenStore
	renderer *templates.Renderer
}

func NewHandler(
	oauth *msgraph.OAuthClient,
	mailer *mail.Client,
	tokens *store.TokenStore,
	renderer *templates.Renderer,
) *Handler {
	return &Handler{
		oauth:    oauth,
		mailer:   mailer,
		tokens:   tokens,
		renderer: renderer,
	}
}

func (h *Handler) Health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *Handler) OAuthLogin(w http.ResponseWriter, r *http.Request) {
	url, err := h.oauth.AuthorizeURL()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	http.Redirect(w, r, url, http.StatusFound)
}

func (h *Handler) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	if errMsg := r.URL.Query().Get("error"); errMsg != "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error":             errMsg,
			"error_description": r.URL.Query().Get("error_description"),
		})
		return
	}

	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	if code == "" || state == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "code e state são obrigatórios",
		})
		return
	}

	if err := h.oauth.ExchangeCode(r.Context(), code, state); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	email, name, err := h.mailer.SenderProfile(r.Context())
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"status":  "connected",
			"message": "Caixa do site autorizada. Não foi possível ler o perfil agora.",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status":      "connected",
		"message":     "Caixa Outlook do site autorizada com sucesso.",
		"senderEmail": email,
		"senderName":  name,
	})
}

func (h *Handler) OAuthStatus(w http.ResponseWriter, r *http.Request) {
	connected, updatedAt, err := h.tokens.Status()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	payload := map[string]any{
		"connected": connected,
		"status":    "needs_reauth",
	}
	if connected {
		payload["status"] = "connected"
		if updatedAt != nil {
			payload["updatedAt"] = updatedAt.Format(time.RFC3339)
		}

		email, name, err := h.mailer.SenderProfile(r.Context())
		if err == nil {
			payload["senderEmail"] = email
			payload["senderName"] = name
		} else if errors.Is(err, msgraph.ErrNeedsReauth) {
			payload["connected"] = false
			payload["status"] = "needs_reauth"
			payload["error"] = err.Error()
		}
	}

	writeJSON(w, http.StatusOK, payload)
}

type sendBody struct {
	To              []string          `json:"to"`
	CC              []string          `json:"cc"`
	Subject         string            `json:"subject"`
	HTMLBody        string            `json:"htmlBody"`
	Template        string            `json:"template"`
	Variables       map[string]string `json:"variables"`
	SaveToSentItems *bool             `json:"saveToSentItems"`
}

type sendByTagBody struct {
	Tag             string            `json:"tag"`
	To              []string          `json:"to"`
	CC              []string          `json:"cc"`
	Variables       map[string]string `json:"variables"`
	SaveToSentItems *bool             `json:"saveToSentItems"`
}

func (h *Handler) SendGeneric(w http.ResponseWriter, r *http.Request) {
	var body sendBody
	if err := decodeJSON(w, r, &body); err != nil {
		writeDecodeError(w, err)
		return
	}

	if strings.TrimSpace(body.Template) == "" {
		if strings.TrimSpace(body.HTMLBody) != "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{
				"error": "htmlBody cru não é permitido; use template ou POST /v1/emails/send-by-tag",
			})
			return
		}
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "template é obrigatório",
		})
		return
	}

	rendered, err := h.renderer.Render(body.Template, body.Variables)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	if body.Subject == "" {
		body.Subject = defaultSubject(body.Template, body.Variables)
	}

	h.send(w, r, body.Template, body.Subject, rendered, body.To, body.CC, body.SaveToSentItems)
}

func (h *Handler) SendByTag(w http.ResponseWriter, r *http.Request) {
	var body sendByTagBody
	if err := decodeJSON(w, r, &body); err != nil {
		writeDecodeError(w, err)
		return
	}
	if strings.TrimSpace(body.Tag) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "tag é obrigatória"})
		return
	}

	vars, err := templates.PrepareVariables(body.Tag, body.Variables, body.To)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	subject, html, err := h.renderer.RenderByTag(body.Tag, vars)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	h.send(w, r, body.Tag, subject, html, body.To, body.CC, body.SaveToSentItems)
}

func (h *Handler) SendInvite(w http.ResponseWriter, r *http.Request) {
	h.sendTyped(w, r, "invite")
}

func (h *Handler) SendConfirmation(w http.ResponseWriter, r *http.Request) {
	h.sendTyped(w, r, "confirmation")
}

func (h *Handler) SendCampaign(w http.ResponseWriter, r *http.Request) {
	h.sendTyped(w, r, "campaign")
}

func (h *Handler) sendTyped(w http.ResponseWriter, r *http.Request, kind string) {
	var body sendBody
	if err := decodeJSON(w, r, &body); err != nil {
		writeDecodeError(w, err)
		return
	}

	html, err := h.renderer.Render(kind, body.Variables)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	subject := body.Subject
	if subject == "" {
		subject = defaultSubject(kind, body.Variables)
	}

	h.send(w, r, kind, subject, html, body.To, body.CC, body.SaveToSentItems)
}

func (h *Handler) send(
	w http.ResponseWriter,
	r *http.Request,
	mailType, subject, html string,
	to, cc []string,
	saveToSentItems *bool,
) {
	if err := validateRecipientList("to", to, true); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	if err := validateRecipientList("cc", cc, false); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	requestID := newRequestID()
	save := false
	if saveToSentItems != nil {
		save = *saveToSentItems
	}

	result, err := h.mailer.Send(r.Context(), mail.SendRequest{
		Subject:         subject,
		HTMLBody:        html,
		To:              to,
		CC:              cc,
		SaveToSentItems: save,
	})
	if err != nil {
		status := http.StatusBadGateway
		if errors.Is(err, store.ErrNotConnected) || errors.Is(err, msgraph.ErrNeedsReauth) {
			status = http.StatusConflict
		} else if strings.Contains(err.Error(), "obrigatório") || strings.Contains(err.Error(), "destinatário") {
			status = http.StatusBadRequest
		} else if result.ProviderStatus == http.StatusTooManyRequests {
			status = http.StatusTooManyRequests
		}

		payload := map[string]any{
			"status":         "error",
			"error":          err.Error(),
			"type":           mailType,
			"requestId":      requestID,
			"providerStatus": result.ProviderStatus,
		}
		if looksLikeTag(mailType) {
			payload["tag"] = mailType
		}
		writeJSON(w, status, payload)
		return
	}

	payload := map[string]any{
		"status":         "accepted",
		"provider":       "microsoft-graph",
		"providerStatus": result.ProviderStatus,
		"type":           mailType,
		"to":             to,
		"requestId":      requestID,
		"providerReqId":  result.RequestID,
	}
	if looksLikeTag(mailType) {
		payload["tag"] = mailType
	}
	writeJSON(w, http.StatusAccepted, payload)
}

func looksLikeTag(value string) bool {
	return value != "" && value == strings.ToUpper(value) && strings.Contains(value, "_")
}

func defaultSubject(kind string, vars map[string]string) string {
	switch strings.ToLower(kind) {
	case "invite", "invites", "convite":
		return "Você recebeu um convite"
	case "confirmation", "confirmations", "confirmacao", "confirmação":
		if vars["acao"] != "" {
			return "Confirmação: " + vars["acao"]
		}
		return "Confirmação"
	case "campaign", "campaigns", "campanha":
		if vars["titulo"] != "" {
			return vars["titulo"]
		}
		return "Novidade"
	default:
		return "Mensagem"
	}
}

func newRequestID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
