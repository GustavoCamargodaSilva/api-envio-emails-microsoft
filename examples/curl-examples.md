# Exemplos cURL

Substitua `SUA_API_KEY` e os e-mails de teste.

## Health

```bash
curl -s http://localhost:8081/health
```

## Status da caixa do site

```bash
curl -s http://localhost:8081/v1/oauth/microsoft/status ^
  -H "X-API-Key: SUA_API_KEY"
```

## Envio por tag (recomendado)

```bash
curl -s -X POST http://localhost:8081/v1/emails/send-by-tag ^
  -H "Content-Type: application/json" ^
  -H "X-API-Key: SUA_API_KEY" ^
  -d "{\"tag\":\"CONVITE_EDICAO_DESPESAS\",\"to\":[\"convidado@gmail.com\"],\"variables\":{\"nomeUsuarioLogado\":\"Gustavo\"}}"
```

Com `emailConvidado` explícito:

```bash
curl -s -X POST http://localhost:8081/v1/emails/send-by-tag ^
  -H "Content-Type: application/json" ^
  -H "X-API-Key: SUA_API_KEY" ^
  -d "{\"tag\":\"CONVITE_EDICAO_DESPESAS\",\"to\":[\"convidado@gmail.com\"],\"variables\":{\"emailConvidado\":\"convidado@gmail.com\",\"nomeUsuarioLogado\":\"Gustavo\"}}"
```

## Envio genérico

```bash
curl -s -X POST http://localhost:8081/v1/emails/send ^
  -H "Content-Type: application/json" ^
  -H "X-API-Key: SUA_API_KEY" ^
  -d "{\"to\":[\"seu-email@gmail.com\"],\"subject\":\"Teste API\",\"htmlBody\":\"<p>Olá do site</p>\"}"
```

## Convite

```bash
curl -s -X POST http://localhost:8081/v1/emails/invites ^
  -H "Content-Type: application/json" ^
  -H "X-API-Key: SUA_API_KEY" ^
  -d "{\"to\":[\"amigo@gmail.com\"],\"variables\":{\"nome\":\"Ana\",\"linkConvite\":\"https://seusite.com/convite/abc\",\"expiraEm\":\"2026-07-20\",\"remetenteNome\":\"Site Teste\"}}"
```

## Confirmação

```bash
curl -s -X POST http://localhost:8081/v1/emails/confirmations ^
  -H "Content-Type: application/json" ^
  -H "X-API-Key: SUA_API_KEY" ^
  -d "{\"to\":[\"usuario@gmail.com\"],\"variables\":{\"nome\":\"João\",\"acao\":\"Cadastro\",\"detalhes\":\"Sua conta foi criada.\"}}"
```

## Campanha

```bash
curl -s -X POST http://localhost:8081/v1/emails/campaigns ^
  -H "Content-Type: application/json" ^
  -H "X-API-Key: SUA_API_KEY" ^
  -d "{\"to\":[\"usuario@gmail.com\"],\"variables\":{\"nome\":\"Maria\",\"titulo\":\"Novidade\",\"conteudo\":\"Temos uma promoção esta semana.\",\"linkCTA\":\"https://seusite.com\",\"textoCTA\":\"Ver agora\"}}"
```
