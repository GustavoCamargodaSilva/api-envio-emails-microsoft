# Referência da API — api-envio-emails

Base URL local de exemplo: `http://localhost:8081`

Substitua `<your-api-key>` e e-mails por valores seus. **Não** use credenciais reais em documentação pública.

## Autenticação

Endpoints protegidos exigem um dos headers:

```http
X-API-Key: <your-api-key>
```

```http
Authorization: Bearer <your-api-key>
```

| Rota | Auth |
|------|------|
| `GET /health` | Público |
| `GET /v1/oauth/ms/login` | Público |
| `GET /v1/oauth/ms/callback` | Público |
| Demais `/v1/**` | API key |

### Erros comuns

| Status | Situação |
|--------|----------|
| 400 | JSON inválido, campos ausentes, tag/template inválido |
| 401 | API key ausente ou inválida |
| 409 | Caixa Outlook não autorizada / precisa reauth |
| 429 | Rate limit (60/min por IP) |
| 502 | Falha no Microsoft Graph |

---

## Health

```http
GET /health
```

**Response `200`**

```json
{
  "status": "ok",
  "time": "2026-07-26T12:00:00Z"
}
```

---

## OAuth Microsoft

### Iniciar autorização (operador)

```http
GET /v1/oauth/ms/login
```

**Response:** `302` para a URL de authorize da Microsoft.

### Callback

```http
GET /v1/oauth/ms/callback?code=...&state=...
```

**Response `200` (sucesso)**

```json
{
  "status": "connected",
  "message": "Caixa Outlook do site autorizada com sucesso.",
  "senderEmail": "sender@example.com",
  "senderName": "Site Name"
}
```

### Status da caixa

```http
GET /v1/oauth/ms/status
X-API-Key: <your-api-key>
```

**Response `200`**

```json
{
  "connected": true,
  "status": "connected",
  "updatedAt": "2026-07-26T12:00:00Z",
  "senderEmail": "sender@example.com",
  "senderName": "Site Name"
}
```

`status` também pode ser `"needs_reauth"` quando o refresh token for inválido.

---

## Envio por tag (recomendado)

```http
POST /v1/emails/send-by-tag
X-API-Key: <your-api-key>
Content-Type: application/json
```

| Campo | Tipo | Obrigatório | Descrição |
|-------|------|-------------|-----------|
| tag | string | Sim | Tag do catálogo |
| to | string[] | Sim | Destinatários (≥ 1) |
| cc | string[] | Não | Cópia |
| variables | object | Condicional | Conforme a tag |
| saveToSentItems | boolean | Não | Default `false` |

### Tag `CONVITE_EDICAO_DESPESAS`

| Variável | Obrigatória | Descrição |
|----------|-------------|-----------|
| nomeUsuarioLogado | Sim | Nome de quem convida |
| emailConvidado | Não | Default = `to[0]` |
| linkAceite | Não | Link de aceite do convite |

```json
{
  "tag": "CONVITE_EDICAO_DESPESAS",
  "to": ["recipient@example.com"],
  "variables": {
    "nomeUsuarioLogado": "Ana Silva",
    "linkAceite": "https://your-app.example.com/convites/aceitar?token=REPLACE_ME"
  }
}
```

**Response `202`**

```json
{
  "status": "accepted",
  "provider": "microsoft-graph",
  "providerStatus": 202,
  "type": "CONVITE_EDICAO_DESPESAS",
  "tag": "CONVITE_EDICAO_DESPESAS",
  "to": ["recipient@example.com"],
  "requestId": "a1b2c3d4e5f67890",
  "providerReqId": "..."
}
```

---

## Envio genérico

```http
POST /v1/emails/send
```

| Campo | Tipo | Obrigatório | Descrição |
|-------|------|-------------|-----------|
| to | string[] | Sim* | Destinatários |
| cc | string[] | Não | Cópia |
| subject | string | Condicional | Obrigatório se não houver subject default do template |
| htmlBody | string | Condicional | HTML cru (se sem template) |
| template | string | Não | `invite`, `confirmation`, `campaign` (e aliases) |
| variables | object | Condicional | Variáveis do template |
| saveToSentItems | boolean | Não | Default `false` |

```json
{
  "to": ["recipient@example.com"],
  "subject": "Teste",
  "htmlBody": "<p>Olá</p>"
}
```

Com template:

```json
{
  "to": ["recipient@example.com"],
  "template": "confirmation",
  "variables": {
    "nome": "João",
    "acao": "Cadastro",
    "detalhes": "Sua conta foi criada com sucesso.",
    "remetenteNome": "Meu Site"
  }
}
```

---

## Endpoints tipados (legado / atalho)

Usam o mesmo body de `send`, com template fixo:

| Endpoint | Template | Variáveis principais |
|----------|----------|----------------------|
| `POST /v1/emails/invites` | invite | `nome`, `linkConvite`, `expiraEm`, `remetenteNome` |
| `POST /v1/emails/confirmations` | confirmation | `nome`, `acao`, `detalhes`, `remetenteNome` |
| `POST /v1/emails/campaigns` | campaign | `nome`, `titulo`, **`conteudo` (obrig.)**, `linkCTA`, `textoCTA`, `remetenteNome` |

Campanha sem `variables.conteudo` → `400`.

---

## Exemplos cURL

Ver [examples/curl-examples.md](../examples/curl-examples.md).

## Arquitetura

Ver [architecture.md](./architecture.md).
