package httpapi

import (
	"fmt"
	netmail "net/mail"
	"strings"
)

const maxRecipients = 50

func validateRecipientList(field string, addresses []string, required bool) error {
	trimmed := make([]string, 0, len(addresses))
	for _, raw := range addresses {
		addr := strings.TrimSpace(raw)
		if addr == "" {
			continue
		}
		trimmed = append(trimmed, addr)
	}
	if required && len(trimmed) == 0 {
		return fmt.Errorf("informe ao menos um destinatário em %s", field)
	}
	if len(trimmed) > maxRecipients {
		return fmt.Errorf("máximo de %d destinatários em %s", maxRecipients, field)
	}
	for _, addr := range trimmed {
		if _, err := netmail.ParseAddress(addr); err != nil {
			return fmt.Errorf("%s inválido: %s", field, addr)
		}
	}
	return nil
}
