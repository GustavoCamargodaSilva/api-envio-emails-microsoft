# Exemplos cURL

Substitua `<your-api-key>` e os e-mails por valores seus.  
**Não** cole API keys reais, secrets ou e-mails pessoais em issues/PRs públicos.

## Health

```bash
curl -s http://localhost:8081/health
```

## Status da caixa do site

```bash
curl -s http://localhost:8081/v1/oauth/ms/status \
  -H "X-API-Key: <your-api-key>"
```

## Envio por tag (recomendado)

```bash
curl -s -X POST http://localhost:8081/v1/emails/send-by-tag \
  -H "Content-Type: application/json" \
  -H "X-API-Key: <your-api-key>" \
  -d "{\"tag\":\"CONVITE_EDICAO_DESPESAS\",\"to\":[\"recipient@example.com\"],\"variables\":{\"nomeUsuarioLogado\":\"Ana Silva\",\"linkAceite\":\"https://your-app.example.com/convites/aceitar?token=REPLACE_ME\"}}"
```

Com `emailConvidado` explícito:

```bash
curl -s -X POST http://localhost:8081/v1/emails/send-by-tag \
  -H "Content-Type: application/json" \
  -H "X-API-Key: <your-api-key>" \
  -d "{\"tag\":\"CONVITE_EDICAO_DESPESAS\",\"to\":[\"recipient@example.com\"],\"variables\":{\"emailConvidado\":\"recipient@example.com\",\"nomeUsuarioLogado\":\"Ana Silva\"}}"
```

## Envio genérico (com template)

```bash
curl -s -X POST http://localhost:8081/v1/emails/send \
  -H "Content-Type: application/json" \
  -H "X-API-Key: <your-api-key>" \
  -d "{\"to\":[\"recipient@example.com\"],\"template\":\"confirmation\",\"variables\":{\"nome\":\"João\",\"acao\":\"Cadastro\",\"detalhes\":\"Ok\",\"remetenteNome\":\"Meu Site\"}}"
```

## Convite

```bash
curl -s -X POST http://localhost:8081/v1/emails/invites \
  -H "Content-Type: application/json" \
  -H "X-API-Key: <your-api-key>" \
  -d "{\"to\":[\"recipient@example.com\"],\"variables\":{\"nome\":\"Ana\",\"linkConvite\":\"https://your-app.example.com/convite/abc\",\"expiraEm\":\"2026-12-31\",\"remetenteNome\":\"Meu Site\"}}"
```

## Confirmação

```bash
curl -s -X POST http://localhost:8081/v1/emails/confirmations \
  -H "Content-Type: application/json" \
  -H "X-API-Key: <your-api-key>" \
  -d "{\"to\":[\"recipient@example.com\"],\"variables\":{\"nome\":\"João\",\"acao\":\"Cadastro\",\"detalhes\":\"Sua conta foi criada.\"}}"
```

## Campanha

```bash
curl -s -X POST http://localhost:8081/v1/emails/campaigns \
  -H "Content-Type: application/json" \
  -H "X-API-Key: <your-api-key>" \
  -d "{\"to\":[\"recipient@example.com\"],\"variables\":{\"nome\":\"Maria\",\"titulo\":\"Novidade\",\"conteudo\":\"Temos uma promoção esta semana.\",\"linkCTA\":\"https://your-app.example.com\",\"textoCTA\":\"Ver agora\"}}"
```

> No Windows PowerShell, prefira aspas simples no `-d` ou um arquivo JSON (`curl --data @payload.json`) para evitar problemas de escape.
