package templates

import (
	"fmt"
	"strings"
)

const TagConviteEdicaoDespesas = "CONVITE_EDICAO_DESPESAS"

type Definition struct {
	Tag          string
	Subject      string
	TemplateFile string
	RequiredVars []string
	OptionalVars []string
}

var catalog = map[string]Definition{
	TagConviteEdicaoDespesas: {
		Tag:          TagConviteEdicaoDespesas,
		Subject:      "Convite para editar despesas financeiras",
		TemplateFile: "files/convite_edicao_despesas.html",
		RequiredVars: []string{"nomeUsuarioLogado"},
		OptionalVars: []string{"emailConvidado", "linkAceite"},
	},
}

func GetDefinition(tag string) (Definition, error) {
	def, ok := catalog[strings.TrimSpace(tag)]
	if !ok {
		return Definition{}, fmt.Errorf("tag desconhecida: %s", tag)
	}
	return def, nil
}

func ListTags() []string {
	tags := make([]string, 0, len(catalog))
	for tag := range catalog {
		tags = append(tags, tag)
	}
	return tags
}

// PrepareVariables valida e normaliza variáveis da tag.
// Se emailConvidado estiver vazio, deriva de to[0].
func PrepareVariables(tag string, vars map[string]string, to []string) (map[string]string, error) {
	def, err := GetDefinition(tag)
	if err != nil {
		return nil, err
	}

	prepared := map[string]string{}
	if vars != nil {
		for k, v := range vars {
			prepared[k] = strings.TrimSpace(v)
		}
	}

	if strings.TrimSpace(prepared["emailConvidado"]) == "" {
		if len(to) == 0 || strings.TrimSpace(to[0]) == "" {
			return nil, fmt.Errorf("variables.emailConvidado é obrigatório quando to está vazio")
		}
		prepared["emailConvidado"] = strings.TrimSpace(to[0])
	}

	missing := make([]string, 0)
	for _, key := range def.RequiredVars {
		if strings.TrimSpace(prepared[key]) == "" {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("variáveis obrigatórias ausentes: %s", strings.Join(missing, ", "))
	}

	return prepared, nil
}
