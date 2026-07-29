package main

import (
	"log"
	"net/http"
	"time"

	"github.com/app-financeiro/api-envio-emails/internal/auth/msgraph"
	"github.com/app-financeiro/api-envio-emails/internal/config"
	"github.com/app-financeiro/api-envio-emails/internal/httpapi"
	"github.com/app-financeiro/api-envio-emails/internal/mail"
	"github.com/app-financeiro/api-envio-emails/internal/store"
	"github.com/app-financeiro/api-envio-emails/internal/templates"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	renderer, err := templates.NewRenderer()
	if err != nil {
		log.Fatalf("templates: %v", err)
	}

	tokenStore := store.NewTokenStore(cfg.TokenStorePath)
	oauth := msgraph.NewOAuthClient(cfg, tokenStore)
	mailer := mail.NewClient(oauth)
	handler := httpapi.NewHandler(oauth, mailer, tokenStore, renderer)
	router := httpapi.NewRouter(handler, cfg.APIKey)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           httpapi.SecurityHeaders(loggingMiddleware(router)),
		ReadHeaderTimeout: 10 * time.Second,
	}

	log.Printf("api-envio-emails ouvindo em http://localhost:%s", cfg.Port)
	log.Printf("autorize a caixa do site em http://localhost:%s/v1/oauth/ms/login", cfg.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server: %v", err)
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}
