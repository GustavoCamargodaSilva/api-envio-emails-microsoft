package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var ErrNotConnected = errors.New("caixa Outlook do site não autorizada")

type TokenSet struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
	Scope        string    `json:"scope,omitempty"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type TokenStore struct {
	path string
	mu   sync.Mutex
}

func NewTokenStore(path string) *TokenStore {
	return &TokenStore{path: path}
}

func (s *TokenStore) Save(tokens TokenSet) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tokens.UpdatedAt = time.Now().UTC()
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(tokens, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o600)
}

func (s *TokenStore) Load() (TokenSet, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return TokenSet{}, ErrNotConnected
		}
		return TokenSet{}, err
	}

	var tokens TokenSet
	if err := json.Unmarshal(data, &tokens); err != nil {
		return TokenSet{}, err
	}
	if tokens.RefreshToken == "" {
		return TokenSet{}, ErrNotConnected
	}
	return tokens, nil
}

func (s *TokenStore) Status() (connected bool, updatedAt *time.Time, err error) {
	tokens, err := s.Load()
	if err != nil {
		if errors.Is(err, ErrNotConnected) {
			return false, nil, nil
		}
		return false, nil, err
	}
	t := tokens.UpdatedAt
	return true, &t, nil
}
