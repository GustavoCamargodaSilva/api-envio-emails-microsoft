package httpapi

import (
	"testing"
)

func TestValidateRecipientList(t *testing.T) {
	if err := validateRecipientList("to", []string{"ok@example.com"}, true); err != nil {
		t.Fatalf("esperado ok, got %v", err)
	}
	if err := validateRecipientList("to", []string{"nao-email"}, true); err == nil {
		t.Fatal("esperado erro para e-mail inválido")
	}
	if err := validateRecipientList("to", nil, true); err == nil {
		t.Fatal("esperado erro para to vazio")
	}
	if err := validateRecipientList("cc", nil, false); err != nil {
		t.Fatalf("cc opcional vazio deve ser ok: %v", err)
	}
}
