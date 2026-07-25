package mail

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/app-financeiro/api-envio-emails/internal/auth/msgraph"
)

type Client struct {
	oauth  *msgraph.OAuthClient
	client *http.Client
}

func NewClient(oauth *msgraph.OAuthClient) *Client {
	return &Client{
		oauth:  oauth,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

type SendRequest struct {
	Subject         string
	HTMLBody        string
	To              []string
	CC              []string
	SaveToSentItems bool
}

type SendResult struct {
	ProviderStatus int
	RequestID      string
}

type graphError struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func (c *Client) Send(ctx context.Context, req SendRequest) (SendResult, error) {
	if len(req.To) == 0 {
		return SendResult{}, fmt.Errorf("informe ao menos um destinatário em to")
	}
	if req.Subject == "" {
		return SendResult{}, fmt.Errorf("subject é obrigatório")
	}
	if req.HTMLBody == "" {
		return SendResult{}, fmt.Errorf("body é obrigatório")
	}

	token, err := c.oauth.AccessToken(ctx)
	if err != nil {
		return SendResult{}, err
	}

	payload := map[string]any{
		"message": map[string]any{
			"subject": req.Subject,
			"body": map[string]any{
				"contentType": "HTML",
				"content":     req.HTMLBody,
			},
			"toRecipients": recipients(req.To),
			"ccRecipients": recipients(req.CC),
		},
		"saveToSentItems": req.SaveToSentItems,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return SendResult{}, err
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://graph.microsoft.com/v1.0/me/sendMail",
		bytes.NewReader(body),
	)
	if err != nil {
		return SendResult{}, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return SendResult{}, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	requestID := resp.Header.Get("request-id")

	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfter := resp.Header.Get("Retry-After")
		return SendResult{ProviderStatus: resp.StatusCode, RequestID: requestID},
			fmt.Errorf("throttling Graph (429); Retry-After=%s", retryAfter)
	}

	if resp.StatusCode == http.StatusAccepted || resp.StatusCode == http.StatusOK {
		return SendResult{ProviderStatus: resp.StatusCode, RequestID: requestID}, nil
	}

	var gErr graphError
	_ = json.Unmarshal(respBody, &gErr)
	msg := gErr.Error.Message
	if msg == "" {
		msg = string(respBody)
	}
	return SendResult{ProviderStatus: resp.StatusCode, RequestID: requestID},
		fmt.Errorf("graph sendMail falhou (%d): %s", resp.StatusCode, msg)
}

func (c *Client) SenderProfile(ctx context.Context) (email string, displayName string, err error) {
	token, err := c.oauth.AccessToken(ctx)
	if err != nil {
		return "", "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://graph.microsoft.com/v1.0/me?$select=mail,userPrincipalName,displayName", nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", "", err
	}
	if resp.StatusCode >= 400 {
		return "", "", fmt.Errorf("falha ao ler perfil (%d): %s", resp.StatusCode, string(body))
	}

	var profile struct {
		Mail              string `json:"mail"`
		UserPrincipalName string `json:"userPrincipalName"`
		DisplayName       string `json:"displayName"`
	}
	if err := json.Unmarshal(body, &profile); err != nil {
		return "", "", err
	}

	email = profile.Mail
	if email == "" {
		email = profile.UserPrincipalName
	}
	return email, profile.DisplayName, nil
}

func recipients(addresses []string) []map[string]any {
	out := make([]map[string]any, 0, len(addresses))
	for _, addr := range addresses {
		if addr == "" {
			continue
		}
		out = append(out, map[string]any{
			"emailAddress": map[string]string{"address": addr},
		})
	}
	return out
}

// ParseRetryAfter ajuda em retries futuros.
func ParseRetryAfter(header string) time.Duration {
	if header == "" {
		return 0
	}
	if secs, err := strconv.Atoi(header); err == nil {
		return time.Duration(secs) * time.Second
	}
	return 0
}
