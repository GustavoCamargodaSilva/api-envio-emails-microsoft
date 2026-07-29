# api-envio-emails

API HTTP em Go para envio de e-mails (convites, confirmações, campanhas e envio por tag) usando a **caixa Outlook do site** via Microsoft Graph.

- Usuários finais **não** fazem login na Microsoft.
- Um operador autoriza **uma vez** a caixa Outlook do site.
- Backends autenticam com API key e disparam o envio.

## Features

- Envio via Microsoft Graph (`Mail.Send` delegated)
- OAuth do operador com refresh token persistido em disco
- Templates HTML tipados + envio por **tag**
- Autenticação por API key (`X-API-Key` ou Bearer)
- Rate limiting, headers de segurança e validação de chave fraca

## Stack

| Item | Tecnologia |
|------|------------|
| Linguagem | Go 1.23 |
| Dependências | Somente stdlib |
| E-mail | Microsoft Graph |
| Templates | `html/template` (embed) |

## Quick Start

### Pré-requisitos

- Go 1.23+
- App Registration no Microsoft Entra (contas pessoais Outlook)
- Uma conta Outlook.com que será o **remetente**

### Configuração

```bash
cp .env.example .env
# edite .env com seus valores (nunca commite .env)
```

| Variável | Obrigatória | Descrição |
|----------|-------------|-----------|
| `API_KEY` | Sim | Chave dos backends (≥ 24 chars aleatórios) |
| `MS_CLIENT_ID` | Para OAuth/envio | Application (client) ID |
| `MS_CLIENT_SECRET` | Para OAuth/envio | Client secret (**segredo**) |
| `MS_TENANT` | Não | Default `consumers` |
| `MS_REDIRECT_URI` | Não | Default local: `http://localhost:8081/v1/oauth/ms/callback` |
| `TOKEN_STORE_PATH` | Não | Default `./data/tokens.json` (**não versionar**) |
| `PORT` | Não | Default `8081` |

### Subir localmente

```bash
go test ./...
go run ./cmd/server
```

1. Abra `http://localhost:8081/v1/oauth/ms/login` e autorize com o **Outlook do site**.
2. Chame os endpoints de envio com `X-API-Key: <sua-api-key>`.

## Documentação

| Documento | Conteúdo |
|-----------|----------|
| [docs/architecture.md](./docs/architecture.md) | Arquitetura, OAuth, segurança |
| [docs/api.md](./docs/api.md) | Referência dos endpoints |
| [examples/curl-examples.md](./examples/curl-examples.md) | Exemplos cURL |

## Microsoft Entra (resumo)

1. Crie um **App registration** (Personal Microsoft accounts, ou misto).
2. Cadastre Redirect URI Web: `http://localhost:8081/v1/oauth/ms/callback` (e a URI pública em produção).
3. Crie um **client secret**.
4. Permissões **Delegated** do Microsoft Graph: `Mail.Send`, `User.Read`.
5. Use `MS_TENANT=consumers` para Outlook.com pessoal.

Passo a passo detalhado: [docs/architecture.md](./docs/architecture.md#microsoft-entra).

## Endpoints (visão rápida)

| Método | Rota | Auth |
|--------|------|------|
| `GET` | `/health` | Público |
| `GET` | `/v1/oauth/ms/login` | Público (operador) |
| `GET` | `/v1/oauth/ms/callback` | Público (Microsoft) |
| `GET` | `/v1/oauth/ms/status` | API key |
| `POST` | `/v1/emails/send-by-tag` | API key |
| `POST` | `/v1/emails/send` | API key |
| `POST` | `/v1/emails/invites` | API key |
| `POST` | `/v1/emails/confirmations` | API key |
| `POST` | `/v1/emails/campaigns` | API key |

## Segurança — o que NÃO publicar

Este repositório é público. **Nunca** versione ou cole em issues/PRs:

- Arquivo `.env` / valores reais de `API_KEY`, `MS_CLIENT_SECRET`
- Arquivo `data/tokens.json` (access/refresh tokens)
- Client secrets, connection strings ou URLs internas de infraestrutura
- E-mails pessoais reais em exemplos (use `recipient@example.com`)

O `.gitignore` já ignora `.env` e `data/*.json`.

## Estrutura

```text
api-envio-emails/
  cmd/server/              # bootstrap HTTP
  internal/config/         # env + validação
  internal/auth/msgraph/   # OAuth + refresh
  internal/mail/           # Graph sendMail
  internal/templates/      # HTML + catálogo de tags
  internal/httpapi/        # rotas, handlers, middleware
  internal/store/          # persistência local de tokens
  docs/                    # documentação
  examples/                # cURL
  .env.example             # template sem segredos
```

## Limitações (MVP)

- Contas Outlook.com gratuitas têm limites diários/anti-spam.
- Se o refresh token for revogado, o operador precisa reabrir `/v1/oauth/ms/login`.
- `202` do Graph significa “aceito para envio”, não entrega garantida.
- Rate limit é por instância (memória local): 60 req/min por IP nos endpoints protegidos.

## License

Uso conforme definido pelo autor do repositório.
