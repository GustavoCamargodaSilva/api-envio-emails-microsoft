package templates

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"strings"
)

//go:embed files/*.html
var files embed.FS

type Renderer struct {
	invite       *template.Template
	confirmation *template.Template
	campaign     *template.Template
	byTag        map[string]*template.Template
}

func NewRenderer() (*Renderer, error) {
	invite, err := template.ParseFS(files, "files/invite.html")
	if err != nil {
		return nil, err
	}
	confirmation, err := template.ParseFS(files, "files/confirmation.html")
	if err != nil {
		return nil, err
	}
	campaign, err := template.ParseFS(files, "files/campaign.html")
	if err != nil {
		return nil, err
	}

	byTag := make(map[string]*template.Template, len(catalog))
	for tag, def := range catalog {
		tpl, err := template.ParseFS(files, def.TemplateFile)
		if err != nil {
			return nil, fmt.Errorf("template da tag %s: %w", tag, err)
		}
		byTag[tag] = tpl
	}

	return &Renderer{
		invite:       invite,
		confirmation: confirmation,
		campaign:     campaign,
		byTag:        byTag,
	}, nil
}

func (r *Renderer) Render(kind string, vars map[string]string) (string, error) {
	data := map[string]string{
		"Nome":          vars["nome"],
		"LinkConvite":   vars["linkConvite"],
		"ExpiraEm":      vars["expiraEm"],
		"Acao":          vars["acao"],
		"Detalhes":      vars["detalhes"],
		"Titulo":        vars["titulo"],
		"Conteudo":      vars["conteudo"],
		"LinkCTA":       vars["linkCTA"],
		"TextoCTA":      vars["textoCTA"],
		"RemetenteNome": firstNonEmpty(vars["remetenteNome"], "Equipe"),
	}

	var buf bytes.Buffer
	var err error
	switch strings.ToLower(kind) {
	case "invite", "invites", "convite":
		err = r.invite.Execute(&buf, data)
	case "confirmation", "confirmations", "confirmacao", "confirmação":
		err = r.confirmation.Execute(&buf, data)
	case "campaign", "campaigns", "campanha":
		if data["Titulo"] == "" {
			data["Titulo"] = "Novidade"
		}
		if data["Conteudo"] == "" {
			return "", fmt.Errorf("variables.conteudo é obrigatório para campanha")
		}
		err = r.campaign.Execute(&buf, data)
	default:
		return "", fmt.Errorf("template desconhecido: %s", kind)
	}
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (r *Renderer) RenderByTag(tag string, vars map[string]string) (subject string, html string, err error) {
	def, err := GetDefinition(tag)
	if err != nil {
		return "", "", err
	}

	tpl, ok := r.byTag[def.Tag]
	if !ok {
		return "", "", fmt.Errorf("template não carregado para tag: %s", tag)
	}

	data := map[string]string{
		"EmailConvidado":    vars["emailConvidado"],
		"NomeUsuarioLogado": vars["nomeUsuarioLogado"],
		"LinkAceite":        vars["linkAceite"],
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return "", "", err
	}
	return def.Subject, buf.String(), nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
