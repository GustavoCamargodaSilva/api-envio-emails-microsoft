package templates

import (
	"strings"
	"testing"
)

func TestRenderInvite(t *testing.T) {
	r, err := NewRenderer()
	if err != nil {
		t.Fatal(err)
	}

	html, err := r.Render("invite", map[string]string{
		"nome":          "Gustavo",
		"linkConvite":   "https://exemplo.com/convite/1",
		"expiraEm":      "2026-07-20",
		"remetenteNome": "Site Teste",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, "Gustavo") || !strings.Contains(html, "https://exemplo.com/convite/1") {
		t.Fatalf("template invite inesperado: %s", html)
	}
}

func TestRenderCampaignRequiresContent(t *testing.T) {
	r, err := NewRenderer()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Render("campaign", map[string]string{"titulo": "Oi"}); err == nil {
		t.Fatal("esperava erro sem conteudo")
	}
}
