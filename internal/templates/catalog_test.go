package templates

import (
	"strings"
	"testing"
)

func TestPrepareVariablesDerivesEmailFromTo(t *testing.T) {
	vars, err := PrepareVariables(TagConviteEdicaoDespesas, map[string]string{
		"nomeUsuarioLogado": "Gustavo",
	}, []string{"convidado@gmail.com"})
	if err != nil {
		t.Fatal(err)
	}
	if vars["emailConvidado"] != "convidado@gmail.com" {
		t.Fatalf("email derivado inesperado: %s", vars["emailConvidado"])
	}
}

func TestPrepareVariablesMissingRequired(t *testing.T) {
	_, err := PrepareVariables(TagConviteEdicaoDespesas, map[string]string{}, []string{"a@b.com"})
	if err == nil || !strings.Contains(err.Error(), "nomeUsuarioLogado") {
		t.Fatalf("esperava erro de nomeUsuarioLogado, got %v", err)
	}
}

func TestPrepareVariablesUnknownTag(t *testing.T) {
	_, err := PrepareVariables("TAG_INEXISTENTE", map[string]string{}, []string{"a@b.com"})
	if err == nil || !strings.Contains(err.Error(), "tag desconhecida") {
		t.Fatalf("esperava tag desconhecida, got %v", err)
	}
}

func TestRenderByTagConviteEdicaoDespesas(t *testing.T) {
	r, err := NewRenderer()
	if err != nil {
		t.Fatal(err)
	}

	subject, html, err := r.RenderByTag(TagConviteEdicaoDespesas, map[string]string{
		"emailConvidado":    "convidado@gmail.com",
		"nomeUsuarioLogado": "Gustavo",
	})
	if err != nil {
		t.Fatal(err)
	}
	if subject == "" {
		t.Fatal("subject vazio")
	}
	if !strings.Contains(html, "convidado@gmail.com") {
		t.Fatalf("html sem email: %s", html)
	}
	if !strings.Contains(html, "Gustavo") {
		t.Fatalf("html sem nomeUsuarioLogado: %s", html)
	}
	if !strings.Contains(html, "despesas financeiras") {
		t.Fatalf("html sem texto do convite: %s", html)
	}
}
