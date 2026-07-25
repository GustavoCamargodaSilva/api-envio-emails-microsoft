# API de Envio de E-mails (Go + Microsoft Graph)

Serviço HTTP em Go que envia e-mails (convites, confirmações, campanhas) usando a **caixa Outlook pessoal do site** via Microsoft Graph.

- Usuários do site **não** fazem login na Microsoft.
- Você (operador) autoriza **uma vez** o Outlook padrão do site.
- A API usa o refresh token salvo para enviar como essa caixa.

Levantamento: [`../docs/levantamento-api-envio-emails-go.md`](../docs/levantamento-api-envio-emails-go.md)

---

## 1. O que configurar (Microsoft / Entra)

### 1.1 Conta remetente

1. Tenha (ou crie) um e-mail Outlook.com do site, ex.: `meu-site-teste@outlook.com`.
2. Esse será o **From** de todos os e-mails.

### 1.2 App Registration

1. Acesse [Microsoft Entra admin center](https://entra.microsoft.com) → **Applications** → **App registrations** → **New registration**.
2. Preencha:
   - **Name:** `api-envio-emails` (ou similar)
   - **Supported account types:** **Personal Microsoft accounts only**  
     (ou “Accounts in any organizational directory and personal Microsoft accounts”)
3. Após criar, anote:
   - **Application (client) ID** → `MS_CLIENT_ID`
4. **Authentication** → **Add a platform** → **Web**:
   - Redirect URI: `http://localhost:8081/v1/oauth/microsoft/callback`
5. **Certificates & secrets** → **New client secret**:
   - Copie o **Value** imediatamente → `MS_CLIENT_SECRET`
6. **API permissions** → **Add a permission** → **Microsoft Graph** → **Delegated**:
   - `Mail.Send`
   - `User.Read` (opcional, para mostrar o e-mail autorizado no `/status`)
   - Não use Application permissions neste cenário.

> Tenant OAuth: use `consumers` no `.env` (`MS_TENANT=consumers`) para contas pessoais.

### 1.3 Arquivo `.env`

```bash
cd api-envio-emails
copy .env.example .env
```

Edite `.env`:

| Variável | Exemplo | Descrição |
|----------|---------|-----------|
| `PORT` | `8081` | Porta HTTP |
| `API_KEY` | string longa | Chave que backends usam para chamar a API |
| `MS_CLIENT_ID` | GUID | Client ID do App Registration |
| `MS_CLIENT_SECRET` | secret | Client secret |
| `MS_TENANT` | `consumers` | Contas pessoais Outlook |
| `MS_REDIRECT_URI` | `http://localhost:8081/v1/oauth/microsoft/callback` | Deve bater com o Entra |
| `TOKEN_STORE_PATH` | `./data/tokens.json` | Onde fica o refresh token da caixa |

---

## 2. Como subir a API

```bash
cd api-envio-emails
go test ./...
go run ./cmd/server
```

Logs esperados:

```text
api-envio-emails ouvindo em http://localhost:8081
autorize a caixa do site em http://localhost:8081/v1/oauth/microsoft/login
```

---

## 3. Autorizar a caixa do site (só operador, uma vez)

1. Com a API rodando, abra no navegador:

   `http://localhost:8081/v1/oauth/microsoft/login`

2. Faça login com o **Outlook do site** (não com o e-mail de um usuário final).
3. Aceite a permissão de envio de e-mail.
4. Você volta para o callback e deve ver JSON com `"status":"connected"` e `senderEmail`.
5. O token fica em `data/tokens.json` (não versionar).

Conferir status (precisa da API key):

```bash
curl -s http://localhost:8081/v1/oauth/microsoft/status -H "X-API-Key: SUA_API_KEY"
```

Se aparecer `"needs_reauth"`, repita o login do passo 1.

---

## 4. Como chamar a API (do seu backend / testes)

### Autenticação da API

Todos os endpoints `/v1/emails/*` e `/v1/oauth/microsoft/status` exigem:

```http
X-API-Key: SUA_API_KEY
```

(ou `Authorization: Bearer SUA_API_KEY`)

Os endpoints de OAuth login/callback **não** pedem API key (fluxo do navegador do operador).

### Endpoints

| Método | Rota | Quem usa | Função |
|--------|------|----------|--------|
| `GET` | `/health` | qualquer | Health check |
| `GET` | `/v1/oauth/microsoft/login` | operador | Inicia autorização Outlook |
| `GET` | `/v1/oauth/microsoft/callback` | Microsoft | Retorno OAuth |
| `GET` | `/v1/oauth/microsoft/status` | operador/ops | Status da caixa |
| `POST` | `/v1/emails/send-by-tag` | backend | **Envio por tag** (recomendado) |
| `POST` | `/v1/emails/send` | backend | Envio genérico ou com template |
| `POST` | `/v1/emails/invites` | backend | Convite (legado) |
| `POST` | `/v1/emails/confirmations` | backend | Confirmação (legado) |
| `POST` | `/v1/emails/campaigns` | backend | Campanha (legado) |

### Envio por tag

```http
POST http://localhost:8081/v1/emails/send-by-tag
X-API-Key: SUA_API_KEY
Content-Type: application/json
```

```json
{
  "tag": "CONVITE_EDICAO_DESPESAS",
  "to": ["convidado@gmail.com"],
  "variables": {
    "nomeUsuarioLogado": "Gustavo"
  }
}
```

| Tag | Variáveis obrigatórias | Opcionais |
|-----|------------------------|-----------|
| `CONVITE_EDICAO_DESPESAS` | `nomeUsuarioLogado` | `emailConvidado` (se omitido, usa `to[0]`) |

Texto gerado:

> Olá **convidado@gmail.com**, você está sendo convidado para participar da edição das despesas financeiras pelo **Gustavo**.  
> Seja muito bem-vindo(a) à plataforma de finanças pessoais.

### Exemplo: usuário do site dispara cadastro

No seu backend (futuro), após criar o usuário:

```http
POST http://localhost:8081/v1/emails/confirmations
X-API-Key: SUA_API_KEY
Content-Type: application/json

{
  "to": ["usuario@gmail.com"],
  "variables": {
    "nome": "João",
    "acao": "Cadastro",
    "detalhes": "Sua conta foi criada com sucesso.",
    "remetenteNome": "Site Teste"
  }
}
```

Resposta de sucesso (`202`):

```json
{
  "status": "accepted",
  "provider": "microsoft-graph",
  "providerStatus": 202,
  "type": "confirmation",
  "to": ["usuario@gmail.com"],
  "requestId": "..."
}
```

Mais exemplos: [`examples/curl-examples.md`](examples/curl-examples.md)

---

## 5. Cenários de teste

### Cenário A — Smoke da API (sem Microsoft)

| Passo | Ação | Esperado |
|-------|------|----------|
| A1 | `GET /health` | `200` + `"status":"ok"` |
| A2 | `POST /v1/emails/send` sem `X-API-Key` | `401` |
| A3 | `GET /v1/oauth/microsoft/status` com API key, sem autorização Outlook | `200` + `"connected": false` |

### Cenário B — Autorização da caixa do site

| Passo | Ação | Esperado |
|-------|------|----------|
| B1 | Abrir `/v1/oauth/microsoft/login` | Redirect para login.microsoftonline.com |
| B2 | Logar com Outlook do site e consentir | Callback com `"status":"connected"` |
| B3 | `GET /status` com API key | `"connected": true` e `senderEmail` preenchido |
| B4 | Conferir `data/tokens.json` | Arquivo criado (não commitar) |

### Cenário C — Envio para você mesmo

| Passo | Ação | Esperado |
|-------|------|----------|
| C1 | `POST /v1/emails/send` com `to` = seu Gmail/Outlook pessoal | `202 accepted` |
| C2 | Abrir a caixa do destinatário | E-mail chegou (pode ir para spam na 1ª vez) |
| C3 | Abrir **Itens Enviados** do Outlook do site | Mensagem aparece se `saveToSentItems` = true |

Body sugerido:

```json
{
  "to": ["seu-email@gmail.com"],
  "subject": "Teste api-envio-emails",
  "htmlBody": "<p>Funcionou.</p>"
}
```

### Cenário D — Templates tipados

| Passo | Endpoint | Validar |
|-------|----------|---------|
| D1 | `POST /v1/emails/invites` | HTML com nome + link |
| D2 | `POST /v1/emails/confirmations` | HTML com ação confirmada |
| D3 | `POST /v1/emails/campaigns` | HTML com título/conteúdo |
| D4 | Campanha sem `variables.conteudo` | `400` |

### Cenário E — Fluxo “usuário do site” (simulado)

Simula o que seu produto fará depois:

1. API já autorizada (Cenário B).
2. Não abrir login Microsoft de novo.
3. Chamar só `POST /v1/emails/confirmations` com o e-mail do “usuário”.
4. Destinatário recebe; remetente continua sendo o Outlook do site.

### Cenário F — Falhas esperadas

| Situação | Como reproduzir | Esperado |
|----------|-----------------|----------|
| Caixa não autorizada | Apagar `data/tokens.json` e enviar | `409` + mensagem de não conectado / reauth |
| API key errada | Header inválido | `401` |
| Destinatário vazio | `"to": []` | `400` |
| Secret Entra errado | `.env` com secret inválido + reauth | Erro no callback/token |
| Redirect URI diferente | URI no Entra ≠ `.env` | Erro OAuth no login |

### Cenário G — Graph Explorer (opcional, antes da API)

No [Graph Explorer](https://developer.microsoft.com/graph/graph-explorer), logado com o Outlook do site:

```http
POST https://graph.microsoft.com/v1.0/me/sendMail
```

Se isso falhar, o problema é permissão/conta — não a API Go.

---

## 6. Checklist rápido de primeiro uso

1. [ ] App Registration criado (Personal accounts)
2. [ ] Redirect URI cadastrado
3. [ ] `Mail.Send` (Delegated) adicionado
4. [ ] `.env` preenchido
5. [ ] `go run ./cmd/server`
6. [ ] Login em `/v1/oauth/microsoft/login` com Outlook do site
7. [ ] `POST /v1/emails/send` para o seu e-mail pessoal
8. [ ] Confirmar recebimento

---

## 7. Estrutura do projeto

```text
api-envio-emails/
  cmd/server/           # bootstrap HTTP
  internal/config/      # env
  internal/auth/msgraph/# OAuth + refresh
  internal/mail/        # Graph sendMail
  internal/templates/   # HTML tipados
  internal/httpapi/     # rotas e handlers
  internal/store/       # tokens.json
  examples/             # cURL
  data/                 # tokens (local)
```

---

## 8. Limitações conhecidas (MVP)

- Conta Outlook.com gratuita: limites diários/anti-spam mais rígidos.
- Se o refresh token expirar/for revogado, o operador precisa reabrir `/login`.
- `202` do Graph significa “aceito para envio”, não entrega garantida.
- Sem fila/retry avançado ainda (retorna erro em `429`).
