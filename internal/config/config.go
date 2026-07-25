package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Port           string
	APIKey         string
	MSClientID     string
	MSClientSecret string
	MSTenant       string
	MSRedirectURI  string
	MSScopes       string
	TokenStorePath string
}

func Load() (Config, error) {
	_ = loadDotEnv(".env")

	cfg := Config{
		Port:           getEnv("PORT", "8081"),
		APIKey:         os.Getenv("API_KEY"),
		MSClientID:     os.Getenv("MS_CLIENT_ID"),
		MSClientSecret: os.Getenv("MS_CLIENT_SECRET"),
		MSTenant:       getEnv("MS_TENANT", "consumers"),
		MSRedirectURI:  getEnv("MS_REDIRECT_URI", "http://localhost:8081/v1/oauth/microsoft/callback"),
		MSScopes: getEnv(
			"MS_SCOPES",
			"https://graph.microsoft.com/Mail.Send https://graph.microsoft.com/User.Read offline_access openid profile",
		),
		TokenStorePath: getEnv("TOKEN_STORE_PATH", "./data/tokens.json"),
	}

	if cfg.APIKey == "" {
		return Config{}, fmt.Errorf("API_KEY é obrigatória")
	}

	return cfg, nil
}

func (c Config) ValidateMicrosoft() error {
	if c.MSClientID == "" || c.MSClientSecret == "" {
		return fmt.Errorf("MS_CLIENT_ID e MS_CLIENT_SECRET são obrigatórios para OAuth/envio")
	}
	return nil
}

func (c Config) ScopeList() []string {
	parts := strings.Fields(c.MSScopes)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func loadDotEnv(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, value)
		}
	}
	return scanner.Err()
}
