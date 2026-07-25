package msgraph

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/app-financeiro/api-envio-emails/internal/config"
	"github.com/app-financeiro/api-envio-emails/internal/store"
)

var ErrNeedsReauth = errors.New("refresh token inválido; reautorize a caixa do site")

type OAuthClient struct {
	cfg    config.Config
	store  *store.TokenStore
	client *http.Client

	mu     sync.Mutex
	states map[string]time.Time
}

func NewOAuthClient(cfg config.Config, tokenStore *store.TokenStore) *OAuthClient {
	return &OAuthClient{
		cfg:    cfg,
		store:  tokenStore,
		client: &http.Client{Timeout: 30 * time.Second},
		states: make(map[string]time.Time),
	}
}

func (o *OAuthClient) AuthorizeURL() (string, error) {
	if err := o.cfg.ValidateMicrosoft(); err != nil {
		return "", err
	}

	state, err := randomState(16)
	if err != nil {
		return "", err
	}

	o.mu.Lock()
	o.states[state] = time.Now().Add(15 * time.Minute)
	o.cleanupExpiredStatesLocked()
	o.mu.Unlock()

	q := url.Values{}
	q.Set("client_id", o.cfg.MSClientID)
	q.Set("response_type", "code")
	q.Set("redirect_uri", o.cfg.MSRedirectURI)
	q.Set("response_mode", "query")
	q.Set("scope", strings.Join(o.cfg.ScopeList(), " "))
	q.Set("state", state)

	return fmt.Sprintf(
		"https://login.microsoftonline.com/%s/oauth2/v2.0/authorize?%s",
		url.PathEscape(o.cfg.MSTenant),
		q.Encode(),
	), nil
}

func (o *OAuthClient) ExchangeCode(ctx context.Context, code, state string) error {
	if err := o.cfg.ValidateMicrosoft(); err != nil {
		return err
	}
	if !o.consumeState(state) {
		return fmt.Errorf("state OAuth inválido ou expirado")
	}

	form := url.Values{}
	form.Set("client_id", o.cfg.MSClientID)
	form.Set("client_secret", o.cfg.MSClientSecret)
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", o.cfg.MSRedirectURI)
	form.Set("scope", strings.Join(o.cfg.ScopeList(), " "))

	tokens, err := o.requestToken(ctx, form)
	if err != nil {
		return err
	}
	return o.store.Save(tokens)
}

func (o *OAuthClient) AccessToken(ctx context.Context) (string, error) {
	if err := o.cfg.ValidateMicrosoft(); err != nil {
		return "", err
	}

	tokens, err := o.store.Load()
	if err != nil {
		return "", err
	}

	if tokens.AccessToken != "" && time.Now().Before(tokens.ExpiresAt.Add(-2*time.Minute)) {
		return tokens.AccessToken, nil
	}

	form := url.Values{}
	form.Set("client_id", o.cfg.MSClientID)
	form.Set("client_secret", o.cfg.MSClientSecret)
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", tokens.RefreshToken)
	form.Set("scope", strings.Join(o.cfg.ScopeList(), " "))

	refreshed, err := o.requestToken(ctx, form)
	if err != nil {
		if errors.Is(err, ErrNeedsReauth) {
			return "", err
		}
		return "", err
	}

	// Microsoft pode rotacionar o refresh token; preservar o anterior se não vier novo.
	if refreshed.RefreshToken == "" {
		refreshed.RefreshToken = tokens.RefreshToken
	}
	if err := o.store.Save(refreshed); err != nil {
		return "", err
	}
	return refreshed.AccessToken, nil
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	Error        string `json:"error"`
	ErrorDesc    string `json:"error_description"`
}

func (o *OAuthClient) requestToken(ctx context.Context, form url.Values) (store.TokenSet, error) {
	endpoint := fmt.Sprintf(
		"https://login.microsoftonline.com/%s/oauth2/v2.0/token",
		url.PathEscape(o.cfg.MSTenant),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return store.TokenSet{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := o.client.Do(req)
	if err != nil {
		return store.TokenSet{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return store.TokenSet{}, err
	}

	var parsed tokenResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return store.TokenSet{}, fmt.Errorf("resposta de token inválida: %w", err)
	}

	if resp.StatusCode >= 400 || parsed.Error != "" {
		if parsed.Error == "invalid_grant" {
			return store.TokenSet{}, ErrNeedsReauth
		}
		msg := parsed.ErrorDesc
		if msg == "" {
			msg = string(body)
		}
		return store.TokenSet{}, fmt.Errorf("falha ao obter token (%d): %s", resp.StatusCode, msg)
	}

	if parsed.AccessToken == "" {
		return store.TokenSet{}, fmt.Errorf("access_token ausente na resposta")
	}

	expiresIn := parsed.ExpiresIn
	if expiresIn <= 0 {
		expiresIn = 3600
	}

	return store.TokenSet{
		AccessToken:  parsed.AccessToken,
		RefreshToken: parsed.RefreshToken,
		ExpiresAt:    time.Now().UTC().Add(time.Duration(expiresIn) * time.Second),
		TokenType:    parsed.TokenType,
		Scope:        parsed.Scope,
	}, nil
}

func (o *OAuthClient) consumeState(state string) bool {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.cleanupExpiredStatesLocked()

	exp, ok := o.states[state]
	if !ok {
		return false
	}
	delete(o.states, state)
	return time.Now().Before(exp)
}

func (o *OAuthClient) cleanupExpiredStatesLocked() {
	now := time.Now()
	for k, exp := range o.states {
		if now.After(exp) {
			delete(o.states, k)
		}
	}
}

func randomState(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
