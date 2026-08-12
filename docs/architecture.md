# Arquitetura — api-envio-emails

## Visão geral

Serviço HTTP em Go (somente stdlib) que envia e-mails pela caixa Outlook do site via Microsoft Graph. OAuth é feito **uma vez pelo operador**; backends autenticam com API key.

```mermaid
flowchart TB
    subgraph Clients
        Backend[Backend / API interna]
        Operator[Operador no navegador]
    end

    subgraph API["api-envio-emails"]
        Router[HTTP Router]
        AuthMW[API Key + Rate Limit]
        OAuth[OAuth Microsoft]
        Mailer[Graph sendMail]
        Tpl[html/template]
        Store[(tokens.json local)]
    end

    subgraph External
        Entra[Microsoft Entra / login]
        Graph[Microsoft Graph]
    end

    Backend -->|X-API-Key| AuthMW --> Router
    Operator -->|login/callback com rate limit| OAuth
    OAuth --> Entra
    OAuth --> Store
    Router --> Tpl
    Router --> Mailer
    Mailer --> OAuth
    Mailer --> Graph
```

**Deploy OAuth:** exponha `/v1/oauth/ms/login` e `/callback` apenas em localhost, VPN ou túnel autenticado. Em produção pública, o risco é takeover da caixa do site. Rate limit: **10 req/min** por IP nessas rotas.
## Camadas

| Pacote | Papel |
|--------|--------|
| `cmd/server` | Bootstrap, timeouts, security headers, logging |
| `internal/config` | Carrega env / `.env` opcional; valida `API_KEY` |
| `internal/httpapi` | Rotas, handlers, middleware |
| `internal/auth/msgraph` | Authorize, exchange, refresh, state CSRF |
| `internal/mail` | `sendMail` + leitura de perfil do remetente |
| `internal/templates` | Catálogo de tags + renderização HTML |
| `internal/store` | Persistência JSON dos tokens OAuth |

## Fluxo OAuth (operador)

```mermaid
sequenceDiagram
    participant Op as Operador
    participant API as api-envio-emails
    participant MS as Microsoft Entra
    participant Disk as tokens.json

    Op->>API: GET /v1/oauth/ms/login
    API->>API: Gera state (TTL 15 min)
    API-->>Op: 302 Authorize URL
    Op->>MS: Login + consent (caixa do site)
    MS-->>API: GET /callback?code&state
    API->>API: Valida/consome state
    API->>MS: Troca code por tokens
    API->>Disk: Salva access + refresh
    API-->>Op: 200 status connected
```

- Scopes padrão: `Mail.Send`, `User.Read`, `offline_access`, `openid`, `profile`.
- Access token é renovado via refresh quando próximo do vencimento.
- `invalid_grant` → status `needs_reauth` (operador precisa logar de novo).

## Envio de e-mail

```mermaid
sequenceDiagram
    participant B as Backend
    participant API as api-envio-emails
    participant G as Graph

    B->>API: POST /v1/emails/send-by-tag + X-API-Key
    API->>API: Valida key + rate limit + tag/vars
    API->>API: Renderiza HTML
    API->>G: POST /v1.0/me/sendMail
    G-->>API: 202 Accepted
    API-->>B: 202 accepted + requestId
```

`202` = aceito pelo Graph para envio; não garante entrega na caixa do destinatário.

## Microsoft Entra

1. [Entra admin center](https://entra.microsoft.com) → **App registrations** → **New registration**.
2. Contas: **Personal Microsoft accounts only** (ou misto, se necessário).
3. Anote o **Application (client) ID**.
4. **Authentication** → plataforma **Web** → Redirect URI igual a `MS_REDIRECT_URI`.
5. **Certificates & secrets** → novo client secret (copie o Value uma vez).
6. **API permissions** → Microsoft Graph → **Delegated**: `Mail.Send`, `User.Read`.
7. Preencha `.env` a partir de `.env.example`.

Em produção, cadastre a Redirect URI pública HTTPS correspondente ao serviço (sem publicar URLs internas de rede privada).

## Segurança

| Controle | Comportamento |
|----------|----------------|
| API key | `X-API-Key` ou `Authorization: Bearer`; comparação em tempo constante |
| Validação no boot | Falha se `API_KEY` vazia, curta (&lt; 24) ou valor fraco conhecido |
| Rate limit | 60 req/min por IP nos endpoints com API key; **10 req/min** em OAuth login/callback (`429` + `Retry-After`) |
| `TRUST_PROXY_HEADERS` | Se `true`, usa `X-Forwarded-For` no rate limit (só atrás de proxy que sanitiza o header) |
| Headers | `nosniff`, `X-Frame-Options: DENY`, `Referrer-Policy: no-referrer`, HSTS |
| OAuth state | Aleatório, TTL 15 min, uso único |
| Token file | Escrito com permissão `0600` |
| JSON | `DisallowUnknownFields` + limite **1 MiB** (`413` se exceder) |
| Destinatários | `net/mail.ParseAddress` em `to`/`cc`; máx. 50 por campo |
| HTML | `htmlBody` cru **rejeitado**; exige `template` ou rotas por tag/tipo |

### Dados sensíveis (não publicar)

- `API_KEY`, `MS_CLIENT_SECRET`
- Conteúdo de `TOKEN_STORE_PATH` (refresh/access tokens)
- Segredos de deploy / variáveis de ambiente de produção
- E-mails pessoais reais em docs/issues

## Configuração

Ver tabela no [README](../README.md#configuração) e o template [`.env.example`](../.env.example).

## Referência de API

Ver [api.md](./api.md).
